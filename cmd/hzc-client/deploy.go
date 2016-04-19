package main

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rethinkdb/horizon-cloud/internal/api"
	"github.com/rethinkdb/horizon-cloud/internal/types"
	"github.com/rethinkdb/horizon-cloud/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	RootCmd.AddCommand(deployCmd)
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "deploy a project",
	Long:  `Deploy the specified project.`,
	Run: func(cmd *cobra.Command, args []string) {
		name := viper.GetString("name")
		if name == "" {
			log.Fatalf("no project name specified (use `-n` or `%s`)", cfgFile)
		}

		/*
			sshServer := viper.GetString("ssh_server")
			identityFile := viper.GetString("identity_file")

			kh, err := ssh.NewKnownHosts([]string{viper.GetString("fingerprint")})
			if err != nil {
				log.Fatalf("failed to deploy: %s", err)
			}
			defer kh.Close()

			sshClient := ssh.New(ssh.Options{
				Host:         sshServer,
				User:         "horizon",
				Environment:  map[string]string{api.ProjectEnvVarName: name},
				KnownHosts:   kh,
				IdentityFile: identityFile,
			})
		*/

		apiClient, err := api.NewClient(viper.GetString("api_server"), "")
		if err != nil {
			log.Fatalf("Couldn't create API client: %v", err)
		}

		log.Printf("Deploying %s...", name)

		// RSI: check whether dist exists.

		files, err := createFileList("dist")
		if err != nil {
			log.Fatal(err)
		}

		for {
			resp, err := apiClient.UpdateProjectManifest(api.UpdateProjectManifestReq{
				Project: name,
				Files:   files,
			})
			if err != nil {
				log.Fatal(err)
			}

			if len(resp.NeededRequests) == 0 {
				break
			}

			err = sendRequests("dist", resp.NeededRequests)
			if err != nil {
				log.Fatal(err)
			}
		}

		log.Printf("Deploy complete!\n")
	},
}

var skipNames = map[string]struct{}{
	"thumbs.db": struct{}{},
}

func createFileList(basePath string) ([]types.FileDescription, error) {
	files, err := walk(basePath)
	if err != nil {
		return nil, err
	}

	trim := basePath + string(filepath.Separator)
	for i := range files {
		files[i].Path = filepath.ToSlash(strings.TrimPrefix(files[i].Path, trim))
		if strings.HasPrefix(files[i].Path, "horizon/") {
			return nil, errors.New("the directory name \"horizon\" is reserved and must not exist")
		}
	}

	return files, nil
}

func walk(basePath string) ([]types.FileDescription, error) {
	f, err := os.Open(basePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if !st.IsDir() {
		var desc types.FileDescription
		desc.Path = basePath

		var buf [16384]byte
		h := md5.New()
		for {
			n, err := f.Read(buf[:])
			_, _ = h.Write(buf[:n])
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}

			if desc.ContentType == "" {
				desc.ContentType = http.DetectContentType(buf[:n])
			}
		}

		desc.MD5 = h.Sum(nil)

		return []types.FileDescription{desc}, nil
	}

	names, err := f.Readdirnames(0)
	if err != nil {
		return nil, err
	}

	sort.Strings(names)

	ret := make([]types.FileDescription, 0, 8)
	for _, name := range names {
		if strings.HasPrefix(name, ".") {
			continue
		}
		if _, ok := skipNames[strings.ToLower(name)]; ok {
			continue
		}

		path := filepath.Join(basePath, name)

		inner, err := walk(path)
		if err != nil {
			return nil, err
		}

		ret = append(ret, inner...)
	}

	return ret, nil
}

func sendRequests(baseDir string, uploads []types.FileUploadRequest) error {
	for _, upload := range uploads {
		err := sendRequest(baseDir, upload)
		if err != nil {
			return err
		}
	}
	return nil
}

func sendRequest(baseDir string, upload types.FileUploadRequest) error {
	var body io.Reader
	if !util.IsSafeRelPath(upload.SourcePath) {
		return fmt.Errorf("%#v is not a safe path", upload.SourcePath)
	}
	if upload.SourcePath != "" {
		fh, err := os.Open(filepath.Join(baseDir, filepath.FromSlash(upload.SourcePath)))
		if err != nil {
			return err
		}
		defer fh.Close()
		body = fh
	}

	log.Printf("Uploading %v", upload.SourcePath)

	r, err := http.NewRequest(upload.Method, upload.URL, body)
	if err != nil {
		return err
	}

	r.Header.Set("Content-Type", "")
	for k, v := range upload.Headers {
		r.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Got unexpected status code %v from %v %v: %#v", resp.StatusCode, upload.Method, upload.URL, string(body))
	}

	return nil
}
