package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/rethinkdb/horizon-cloud/internal/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type config struct {
	APIClient *api.Client
}

var cfgFile string

// This represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "horizon-cloud-http-proxy",
	Short: "horizon-cloud-http-proxy",
	Long:  `horizon-cloud-http-proxy`,
	Run: func(cmd *cobra.Command, args []string) {
		conf := &config{}

		secretPath := viper.GetString("secret_path")
		secret, err := ioutil.ReadFile(secretPath)
		if err != nil {
			log.Fatalf("Couldn't read api server secret from %v: %v", secretPath, err)
		}

		conf.APIClient, err = api.NewClient(viper.GetString("api_server"), string(secret))
		if err != nil {
			log.Fatalf("Couldn't create API client: %v", err)
		}

		handler := NewHandler(conf)

		if tlsAddr := viper.GetString("listen_tls"); tlsAddr != "" {
			go func() {
				certFile := viper.GetString("tls_cert")
				keyFile := viper.GetString("tls_key")
				log.Fatal(http.ListenAndServeTLS(tlsAddr, certFile, keyFile, handler))
			}()
		}

		log.Fatal(http.ListenAndServe(viper.GetString("listen"), handler))
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	pf := RootCmd.PersistentFlags()

	pf.StringP("listen", "l", ":80", "Address to listen for HTTP connections on.")
	viper.BindPFlag("listen", pf.Lookup("listen"))

	pf.StringP("listen_tls", "", "", "Address to listen for HTTPS connections on.")
	viper.BindPFlag("listen_tls", pf.Lookup("listen_tls"))

	pf.StringP("tls_cert", "", "", "Path to TLS certificate chain")
	viper.BindPFlag("tls_cert", pf.Lookup("tls_cert"))

	pf.StringP("tls_key", "", "", "Path to TLS key")
	viper.BindPFlag("tls_key", pf.Lookup("tls_key"))

	pf.StringP("api_server", "a", "http://api-server:8000", "API server base URL.")
	viper.BindPFlag("api_server", pf.Lookup("api_server"))

	pf.StringP("secret_path", "s",
		"/secrets/api-shared-secret/api-shared-secret",
		"Path to API server shared secret")
	viper.BindPFlag("secret_path", pf.Lookup("secret_path"))
}

// initConfig reads in ENV variables
func initConfig() {
	viper.AutomaticEnv()
}
