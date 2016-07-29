package db

import (
	"errors"
	"fmt"
	"time"

	r "github.com/dancannon/gorethink"
	"github.com/rethinkdb/horizon-cloud/internal/hzlog"
	"github.com/rethinkdb/horizon-cloud/internal/types"
	"github.com/rethinkdb/horizon-cloud/internal/util"
)

type DBConnection struct {
	session *r.Session
}

var (
	ErrCanceled = errors.New("canceled")

	projects = r.DB("web_backend").Table("projects")
	domains  = r.DB("web_backend").Table("domains")
	users    = r.DB("web_backend_internal").Table("users")
)

type hzUser struct {
	Name string `gorethink:"id"`
	Data struct {
		PublicSSHKeys []string `gorethink:"keys"`
	} `gorethink:"data"`
}

func New(addr string) (*DBConnection, error) {
	session, err := r.Connect(r.ConnectOpts{
		Address:           addr,
		MaxIdle:           10,
		MaxOpen:           10,
		HostDecayDuration: time.Second * 10,
	})
	if err != nil {
		return nil, err
	}
	return &DBConnection{session}, nil
}

func (d *DBConnection) WithLogger(l *hzlog.Logger) *DB {
	return &DB{d, l}
}

type DB struct {
	*DBConnection
	log *hzlog.Logger
}

func setBasicError(typeName string, objectName string) error {
	return fmt.Errorf("internal error: unable to set %v `%v`", typeName, objectName)
}

func getBasicError(typeName string, name string) error {
	return fmt.Errorf("internal error: unable to get %v `%s`", typeName, name)
}

type projectWriteResp struct {
	Errors     int    `gorethink:"errors"`
	Unchanged  int    `gorethink:"unchanged"`
	FirstError string `gorethink:"first_error"`
	Changes    []struct {
		NewVal *types.Project `gorethink:"new_val"`
	} `gorethink:"changes"`
}

func (d *DB) runProjectWriteDetailed(
	q r.Term) (*types.Project, *projectWriteResp, error) {
	var resp projectWriteResp
	err := runOne(q, d.session, &resp)

	if err != nil {
		return nil, nil, err
	}
	if resp.Errors != 0 {
		return nil, nil, errors.New(resp.FirstError)
	}
	if len(resp.Changes) != 1 {
		return nil, nil, fmt.Errorf("Unexpected number of changes in response: %v", resp)
	}
	return resp.Changes[0].NewVal, &resp, nil
}

func (d *DB) runProjectWrite(q r.Term) (*types.Project, error) {
	p, _, err := d.runProjectWriteDetailed(q)
	return p, err
}

func (d *DB) SetProjectKubeConfig(
	project string, kc types.KubeConfig) (*types.Project, error) {
	q := projects.Insert(types.Project{
		ID:    util.TrueName(project),
		Name:  project,
		Users: []string{},

		KubeConfig:        kc,
		KubeConfigVersion: types.ConfigVersion{Desired: 1},

		HorizonConfig: []byte{},
	}, r.InsertOpts{Conflict: func(id r.Term, oldVal r.Term, newVal r.Term) r.Term {
		return oldVal.Merge(map[string]r.Term{
			"KubeConfig": newVal.Field("KubeConfig"),
			"KubeConfigVersion": r.Expr(map[string]r.Term{
				"Desired": oldVal.Field("KubeConfigVersion").Field("Desired").Add(1),
			}),
		})
	}, ReturnChanges: "always"})
	return d.runProjectWrite(q)
}

func (d *DB) UpdateProject(projectID string, projectPatch types.Project) (bool, error) {
	q := projects.Get(projectID).Update(projectPatch)
	res, err := q.RunWrite(d.session)
	if err != nil {
		return false, err
	}
	return res.Replaced == 1, nil
}

func (d *DB) DeleteProject(projectID string) error {
	q := projects.Get(projectID).Delete()
	_, err := q.RunWrite(d.session)
	return err
}

func (d *DB) AddProjectUsers(project string, users []string) (*types.Project, error) {
	q := projects.Get(util.TrueName(project)).Update(func(oldVal r.Term) r.Term {
		return r.Expr(map[string]r.Term{
			"Users": oldVal.Field("Users").Default([]string{}).SetUnion(users),
		})
	}, r.UpdateOpts{ReturnChanges: "always"})
	return d.runProjectWrite(q)
}

func (d *DB) DelProjectUsers(project string, users []string) (*types.Project, error) {
	q := projects.Get(util.TrueName(project)).Update(func(oldVal r.Term) r.Term {
		return r.Expr(map[string]r.Term{
			"Users": oldVal.Field("Users").Default([]string{}).SetDifference(users),
		})
	}, r.UpdateOpts{ReturnChanges: "always"})
	return d.runProjectWrite(q)
}

func (d *DB) MaybeUpdateHorizonConfig(
	projectName string, hzConf types.HorizonConfig) (int64, string, error) {
	hcv := "HorizonConfigVersion"
	q := r.Expr(hzConf).Do(func(hzc r.Term) r.Term {
		return projects.Get(util.TrueName(projectName)).Update(func(config r.Term) r.Term {
			return r.Branch(
				config.Field("HorizonConfig").Eq(hzc).Default(false),
				nil,
				map[string]interface{}{
					"HorizonConfig": r.Expr(hzc),
					hcv: map[string]interface{}{
						"Desired": config.Field(hcv).Field("Desired").Default(0).Add(1),
					},
				})
		}, r.UpdateOpts{ReturnChanges: "always"})
	})
	project, resp, err := d.runProjectWriteDetailed(q)
	if err != nil {
		return 0, "", err
	}
	if resp.Unchanged != 0 {
		if project.HorizonConfigVersion.Desired == project.HorizonConfigVersion.Applied {
			// We return a special value if nothing changed so that we can skip waiting.
			return 0, "", nil
		}
		if project.HorizonConfigVersion.Desired == project.HorizonConfigVersion.Error {
			return 0, project.HorizonConfigVersion.LastError, nil
		}
	}
	return project.HorizonConfigVersion.Desired, "", nil
}

type HZStateType int

const (
	HZError      HZStateType = 0
	HZApplied    HZStateType = 1
	HZSuperseded HZStateType = 2
	HZDeleted    HZStateType = 3
)

type HZState struct {
	Typ       HZStateType
	LastError string
}

func (d *DB) WaitForHorizonConfigVersion(
	project string, version int64) (HZState, error) {
	q := projects.Get(util.TrueName(project)).Changes(r.ChangesOpts{IncludeInitial: true})
	cursor, err := q.Run(d.session)
	if err != nil {
		d.log.Error("WaitForHorizonConfigVersion(%s, %d): %v", project, version, err)
		return HZState{}, err
	}
	defer cursor.Close()
	var c ProjectChange
	for cursor.Next(&c) {
		if c.NewVal == nil {
			return HZState{Typ: HZDeleted}, nil
		}
		if c.NewVal.HorizonConfigVersion.Applied == version {
			return HZState{Typ: HZApplied}, nil
		}
		if c.NewVal.HorizonConfigVersion.Error == version {
			return HZState{
				Typ:       HZError,
				LastError: c.NewVal.HorizonConfigVersion.LastError,
			}, nil
		}
		if c.NewVal.HorizonConfigVersion.Applied > version ||
			c.NewVal.HorizonConfigVersion.Error > version {
			return HZState{Typ: HZSuperseded}, nil
		}
	}

	err = fmt.Errorf("Changefeed aborted unexpectedly.")
	d.log.Error("WaitForHorizonConfigVersion(%s, %d): %v", project, version, err)
	return HZState{}, err
}

func (d *DB) GetUsersByKey(publicKey string) ([]string, error) {
	q := users.GetAllByIndex("PublicSSHKeys", publicKey)
	cursor, err := q.Run(d.session)
	if err != nil {
		d.log.Error("Couldn't get users by key: %v", err)
		return nil, err
	}
	defer cursor.Close()
	var users []string
	var u hzUser
	for cursor.Next(&u) {
		users = append(users, u.Name)
	}
	return users, nil
}

func (d *DB) GetProjectNamesByKey(publicKey string) ([]string, error) {
	q := projects.GetAllByIndex("Users",
		r.Args(users.GetAllByIndex("PublicSSHKeys", publicKey).
			Field("id").CoerceTo("array")))
	cursor, err := q.Run(d.session)
	if err != nil {
		d.log.Error("Couldn't get projectAddrs by key: %v", err)
		return nil, err
	}
	defer cursor.Close()
	var names []string
	var p types.Project
	for cursor.Next(&p) {
		names = append(names, p.Name)
	}
	return names, nil
}

func (d *DB) GetProjectNamesByUsers(users []string) ([]string, error) {
	q := projects.GetAllByIndex("Users", r.Args(users))
	cursor, err := q.Run(d.session)
	if err != nil {
		d.log.Error("Couldn't get projects by key: %v", err)
		return nil, err
	}
	defer cursor.Close()
	var names []string
	var p types.Project
	for cursor.Next(&p) {
		names = append(names, p.Name)
	}
	return names, nil
}

func (d *DB) GetProjectNameByDomain(domainName string) (string, error) {
	var domain types.Domain
	err := runOne(domains.Get(domainName), d.session, &domain)
	if err != nil {
		if err != r.ErrEmptyResult {
			return "", err
		}
		return "", nil
	}
	return domain.Project, nil
}

func (d *DB) GetProject(name string) (*types.Project, error) {
	var p types.Project
	err := d.getBasicType(projects, "project", util.TrueName(name), &p)
	return &p, err
}

func (d *DB) SetDomain(domain types.Domain) error {
	res, err := domains.Insert(domain).RunWrite(d.session)
	if err != nil {
		d.log.Error("Couldn't set %v %#v: %v", "domain", domain.Domain, err)
		return setBasicError("domain", domain.Domain)
	}
	if res.Inserted+res.Unchanged+res.Replaced != 1 {
		d.log.Error("Couldn't set %v %#v: unexpected result counts", "domain", domain.Domain)
		return setBasicError("domain", domain.Domain)
	}
	return nil
}

func (d *DB) DelDomain(domain types.Domain) error {
	res, err := domains.Get(domain.Domain).Replace(func(row r.Term) r.Term {
		return r.Branch(
			row.Field("Project").Eq(domain.Project),
			nil,
			r.Error("Project name didn't match domain."))
	}).RunWrite(d.session)
	if err != nil {
		return err
	}
	if res.Deleted != 1 {
		return fmt.Errorf("unable to delete domain `%s`", domain.Domain)
	}
	return nil
}

func (d *DB) GetDomainsByProject(project string) ([]string, error) {
	cursor, err := domains.GetAllByIndex("Project", project).Run(d.session)
	if err != nil {
		return nil, err
	}
	defer cursor.Close()
	domains := []string{}
	var dom types.Domain
	for cursor.Next(&dom) {
		domains = append(domains, dom.Domain)
	}
	return domains, nil
}

type ProjectChange struct {
	OldVal *types.Project `gorethink:"old_val"`
	NewVal *types.Project `gorethink:"new_val"`
}

func (d *DB) projectChangesLoop(out chan<- ProjectChange) {
	query := projects.Changes(r.ChangesOpts{IncludeInitial: true})
	for {
		ch := make(chan ProjectChange)
		cur, err := query.Run(d.session)
		if err != nil {
			d.log.Error("Couldn't query for config changes: %s", err)
			time.Sleep(time.Second * 5)
			continue
		}
		cur.Listen(ch)
		for el := range ch {
			out <- el
		}
		err = cur.Err()
		d.log.Error("Config changes loop ended: %v", err)
	}
}

func (d *DB) ProjectChanges(out chan<- ProjectChange) {
	go d.projectChangesLoop(out)
}

func runOne(query r.Term, session *r.Session, out interface{}) error {
	cur, err := query.Run(session)
	if err != nil {
		return err
	}
	return cur.One(out)
}

func (d *DB) getBasicType(
	table r.Term, typeName string, id string, i interface{}) error {

	err := runOne(table.Get(id), d.session, i)
	if err == r.ErrEmptyResult {
		return fmt.Errorf("%s `%s` does not exist", typeName, id)
	} else if err != nil {
		return getBasicError(typeName, id)
	}
	return nil
}
