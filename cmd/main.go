package main

import (
	"encoding/json"
	"fmt"
	dashboard "github.com/camunda-ci/camunda-ci-dashboard"
	"github.com/gorilla/mux"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"os/user"
	"time"
)

const (
	contentTypeJSON = "application/json"
	cfgFileName     = ".camunda-ci-dashboard"
)

type Config struct {
	Jenkins     []*dashboard.JenkinsInstance
	Travis      []*dashboard.TravisInstance
	Username    string
	Password    string
	Debug       bool
	BindAddress string
}

func (c *Config) String() string {
	return fmt.Sprintf(
		"Config{username:%s, password:%s, bindAddress:%s, debug:%t, jenkins:%+v}",
		c.Username, c.Password, c.BindAddress, c.Debug, c.Jenkins)
}

var (
	Build string

	Timeout = 30 * time.Second

	dashboardEndpoint = "/dashboard"
	jenkinsEndpoint   = dashboardEndpoint + "/jenkins"
	travisEndpoint    = dashboardEndpoint + "/travis"
	brokenBoard       *dashboard.Dashboard
	config            *Config
)

func homeDir() string {
	usr, err := user.Current()
	var homeDir string

	if err == nil {
		homeDir = usr.HomeDir
	} else {
		// Maybe it's cross compilation without cgo support. (darwin, unix)
		homeDir = os.Getenv("HOME")
	}

	return homeDir
}

func readConfig() {
	viper.SetConfigName(cfgFileName)
	viper.SetConfigType("json")
	viper.AddConfigPath(".")
	viper.AddConfigPath(homeDir())

	// default values
	viper.SetDefault("username", "")
	viper.SetDefault("password", "")
	viper.SetDefault("bindAddress", "127.0.0.1:8000")
	viper.SetDefault("debug", false)

	// cmd line flags
	pflag.String("bindAddress", "127.0.0.1:8000", "")
	pflag.String("username", "", "")
	pflag.String("password", "", "")
	pflag.Bool("debug", false, "")
	viper.BindPFlag("bindAddress", pflag.Lookup("bindAddress"))
	viper.BindPFlag("username", pflag.Lookup("username"))
	viper.BindPFlag("password", pflag.Lookup("password"))
	viper.BindPFlag("debug", pflag.Lookup("debug"))
	pflag.Parse()

	// ENV vars
	viper.SetEnvPrefix("ccd")
	viper.BindEnv("username")
	viper.BindEnv("password")
	viper.BindEnv("bindAddress")
	viper.BindEnv("debug")
	viper.AutomaticEnv()

	// evaluate
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		log.Printf("[INFO] No config file found. Using defaults.\n")
	}

	config = &Config{
		Debug:       viper.GetBool("debug"),
		BindAddress: viper.GetString("bindAddress"),
		Username:    viper.GetString("username"),
		Password:    viper.GetString("password"),
		Jenkins:     parseJenkinsInstanceConfig(),
		Travis:      parseTravisInstanceConfig(),
	}

	if config.Debug {
		log.Printf("Config: %+v", *config)
		viper.Debug()
	}

	dashboard.Debug = config.Debug
}

func parseTravisInstanceConfig() []*dashboard.TravisInstance {
	var travisInstances []*dashboard.TravisInstance

	type config struct {
		AccessToken   string
		Organizations []struct {
			Name  string
			Repos []struct {
				Name   string
				Branch string
			}
		}
	}

	var cfg config
	err := viper.UnmarshalKey("travis", &cfg)
	if err != nil {
		log.Fatalln("Error while parsing Travis config:", err)
	}

	for _, org := range cfg.Organizations {
		if org.Name == "" {
			continue
		}

		client := dashboard.NewTravisClient(dashboard.TravisApiUrl, cfg.AccessToken)
		travisInstance := &dashboard.TravisInstance{Client: client, Name: org.Name}

		for _, r := range org.Repos {
			if r.Name == "" {
				continue
			}

			branch := r.Branch
			if branch == "" {
				branch = "master"
			}

			travisInstance.Repos = append(travisInstance.Repos,
				dashboard.TravisRepository{Organization: org.Name, Name: r.Name, Branch: branch})

		}
		travisInstances = append(travisInstances, travisInstance)
	}

	return travisInstances
}

func parseJenkinsInstanceConfig() []*dashboard.JenkinsInstance {
	var jenkinsInstances []*dashboard.JenkinsInstance

	jenkins := viper.Get("jenkins")
	if jenkins != nil {
		jenkinsMap := jenkins.(map[string]interface{})

		for k, v := range jenkinsMap {
			jenkinsInstance := new(dashboard.JenkinsInstance)
			jenkinsInstance.Name = k

			url := v.(map[string]interface{})["url"].(string)
			jenkinsInstance.Url = url

			publicUrl := v.(map[string]interface{})["publicurl"].(string)
			jenkinsInstance.PublicUrl = publicUrl

			if brokenJobsUrl, ok := v.(map[string]interface{})["brokenjobsurl"]; ok {
				jenkinsInstance.BrokenJobsUrl = brokenJobsUrl.(string)
			}

			jenkinsInstances = append(jenkinsInstances, jenkinsInstance)
		}
	}

	return jenkinsInstances
}

func main() {
	readConfig()
	brokenBoard = dashboard.Init(config.Jenkins, config.Travis, config.Username, config.Password)
	initServer(config.BindAddress)
}

func initServer(bindAddress string) {
	router := mux.NewRouter()

	router.HandleFunc(jenkinsEndpoint, jenkinsBoardHandler).Methods(http.MethodGet)
	router.HandleFunc(travisEndpoint, travisBoardHandler).Methods(http.MethodGet)
	router.PathPrefix("/static").Handler(http.StripPrefix("/static", http.FileServer(assetFS())))
	router.Path("/").Handler(http.StripPrefix("/", http.FileServer(assetFS())))

	srv := &http.Server{
		Handler:      router,
		Addr:         bindAddress,
		WriteTimeout: Timeout,
		ReadTimeout:  Timeout,
	}

	log.Printf("[INFO] Dashboard (v%s) can be access using your browser at '%s'", Build, bindAddress)

	log.Fatal(srv.ListenAndServe())
}

func attachProfiler(router *mux.Router) {
	router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
}

func travisBoardHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", contentTypeJSON)
	_ = json.NewEncoder(w).Encode(brokenBoard.GetBrokenTravisBuilds())
}

func jenkinsBoardHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", contentTypeJSON)
	_ = json.NewEncoder(w).Encode(brokenBoard.GetBrokenJenkinsBuilds())
}
