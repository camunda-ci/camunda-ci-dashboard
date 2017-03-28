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
	_ "net/http/pprof"
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

	//viper.WatchConfig()
	//viper.OnConfigChange(func(e fsnotify.Event) {
	//	fmt.Println("Config file changed:", e.Name)
	//})

	jenkinsInstances := []*dashboard.JenkinsInstance{}

	jenkins := viper.Get("jenkins")
	if jenkins != nil {
		jenkinsMap := jenkins.(map[string]interface{})

		for k, v := range jenkinsMap {
			jenkinsInstance := new(dashboard.JenkinsInstance)
			jenkinsInstance.Name = k
			url := v.(map[string]interface{})["url"].(string)
			jenkinsInstance.Url = url

			jenkinsInstances = append(jenkinsInstances, jenkinsInstance)
		}
	}

	config = &Config{
		Debug:       viper.GetBool("debug"),
		BindAddress: viper.GetString("bindAddress"),
		Username:    viper.GetString("username"),
		Password:    viper.GetString("password"),
		Jenkins:     jenkinsInstances,
	}

	if config.Debug {
		log.Printf("Config: %+v", *config)
		viper.Debug()
	}

	dashboard.Debug = config.Debug
}

func main() {
	readConfig()
	brokenBoard = dashboard.Init(config.Jenkins, config.Username, config.Password)
	initServer(config.BindAddress)
}

func initServer(bindAddress string) {
	router := mux.NewRouter()
	//attachProfiler(router)

	router.HandleFunc(dashboardEndpoint, brokenBoardDataHandler).Methods(http.MethodGet)
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

func brokenBoardDataHandler(w http.ResponseWriter, r *http.Request) {
	brokenJenkinsBuilds := brokenBoard.GetBrokenJenkinsBuilds()

	w.Header().Set("Content-Type", contentTypeJSON)
	json.NewEncoder(w).Encode(brokenJenkinsBuilds)
}
