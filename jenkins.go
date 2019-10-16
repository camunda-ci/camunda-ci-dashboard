package dashboard

import (
	"encoding/json"
	"fmt"
	client "github.com/camunda-ci/camunda-ci-dashboard/http"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
)

const (
	jsonAPI       = "/api/json"
	queue         = "queue" + jsonAPI
	computer      = "computer" + jsonAPI
	busyExecutors = "computer" + jsonAPI + "?tree=busyExecutors"
	overallLoad   = "overallLoad" + jsonAPI
)

var (
	// Debug set to true enable request debugging
	Debug = false
)

// JenkinsInstance holds basic informations about a Jenkins instance and the client connected to it.
type JenkinsInstance struct {
	Name          string
	Url           string
	BrokenJobsUrl string
	Client        Jenkins
}

// JenkinsAggregations is a container for all retrieved JenkinsAggregation
type JenkinsAggregations struct {
	jenkinsAggregation []JenkinsAggregation
}

// Holds all dashboard relevant informations for a Jenkins instance
type JenkinsAggregation struct {
	Aggregation
	BrokenJobsUrl  string       `json:"brokenJobsUrl"`
	BusyExecutors  int          `json:"busyExecutors"`
	BuildQueueSize int          `json:"buildQueueSize"`
	Jobs           []JenkinsJob `json:"jobs"`
}

// Jenkins is high-level API for accessing the underlying Jenkins instance.
type Jenkins interface {
	GetQueue() (*JenkinsQueue, error)
	GetJobsFromView(viewName string) ([]JenkinsJob, error)
	GetJobsFromViewWithTree(viewName string, tree string) ([]JenkinsJob, error)
	GetJobsFromViewByPath(path string) ([]JenkinsJob, error)
	GetJobsFromViewWithTreeByPath(path string, tree string) ([]JenkinsJob, error)
	GetOverallLoad() (*JenkinsOverallLoad, error)
	GetExecutors() (*JenkinsExecutors, error)
	GetBusyExecutors() (int, error)
}

// JenkinsClient implements the Jenkins interface and holds the client connected to the underlying Jenkins instance.
type JenkinsClient struct {
	client *client.HTTPClient
}

// JenkinsQueue represents the Jenkins Build queue.
type JenkinsQueue struct {
	Items []struct {
		Actions []struct {
			Causes []struct {
				ShortDescription string `json:"shortDescription"`
				UpstreamBuild    int    `json:"upstreamBuild"`
				UpstreamProject  string `json:"upstreamProject"`
				UpstreamURL      string `json:"upstreamUrl"`
			} `json:"causes"`
		} `json:"actions"`
		Blocked      bool   `json:"blocked"`
		Buildable    bool   `json:"buildable"`
		ID           int    `json:"id"`
		InQueueSince int64  `json:"inQueueSince"`
		Params       string `json:"params"`
		Stuck        bool   `json:"stuck"`
		Task         struct {
			Name  string `json:"name"`
			URL   string `json:"url"`
			Color string `json:"color"`
		} `json:"task"`
		URL                        string `json:"url"`
		Why                        string `json:"why"`
		BuildableStartMilliseconds int64  `json:"buildableStartMilliseconds"`
		Pending                    bool   `json:"pending"`
	} `json:"items"`
}

func (q *JenkinsQueue) String() string {
	return fmt.Sprintf("%#v", q)
}

type JenkinsJob struct {
	Name      string `json:"name"`
	URL       string `json:"url"`
	Color     string `json:"color"`
	LastBuild struct {
		Actions []struct {
			FailCount          int           `json:"failCount,omitempty"`
			SkipCount          int           `json:"skipCount,omitempty"`
			TotalCount         int           `json:"totalCount,omitempty"`
			FoundFailureCauses []interface{} `json:"foundFailureCauses,omitempty"`
		} `json:"actions"`
	} `json:"lastBuild"`
}

func (j *JenkinsJob) String() string {
	return fmt.Sprintf("%#v", j)
}

// JenkinsView represents a view inside Jenkins including all jobs on it.
type JenkinsView struct {
	Jobs []JenkinsJob `json:"jobs"`
}

func (v *JenkinsView) String() string {
	return fmt.Sprintf("%#v", v)
}

// represents the configured executors of the underlying Jenkins instance.
type JenkinsExecutors struct {
	BusyExecutors int `json:"busyExecutors"`
	Computer      []struct {
		Actions []struct {
		} `json:"actions"`
		DisplayName string `json:"displayName"`
		Executors   []struct {
		} `json:"executors"`
		Icon            string `json:"icon"`
		IconClassName   string `json:"iconClassName"`
		Idle            bool   `json:"idle"`
		JnlpAgent       bool   `json:"jnlpAgent"`
		LaunchSupported bool   `json:"launchSupported"`
		LoadStatistics  struct {
		} `json:"loadStatistics"`
		ManualLaunchAllowed bool `json:"manualLaunchAllowed"`
		MonitorData         struct {
			HudsonNodeMonitorsSwapSpaceMonitor struct {
				AvailablePhysicalMemory int64 `json:"availablePhysicalMemory"`
				AvailableSwapSpace      int64 `json:"availableSwapSpace"`
				TotalPhysicalMemory     int64 `json:"totalPhysicalMemory"`
				TotalSwapSpace          int64 `json:"totalSwapSpace"`
			} `json:"hudson.node_monitors.SwapSpaceMonitor"`
			HudsonNodeMonitorsArchitectureMonitor string `json:"hudson.node_monitors.ArchitectureMonitor"`
			HudsonNodeMonitorsResponseTimeMonitor struct {
				Average int `json:"average"`
			} `json:"hudson.node_monitors.ResponseTimeMonitor"`
			HudsonNodeMonitorsTemporarySpaceMonitor struct {
				Path string `json:"path"`
				Size int64  `json:"size"`
			} `json:"hudson.node_monitors.TemporarySpaceMonitor"`
			HudsonNodeMonitorsDiskSpaceMonitor struct {
				Path string `json:"path"`
				Size int64  `json:"size"`
			} `json:"hudson.node_monitors.DiskSpaceMonitor"`
			HudsonNodeMonitorsClockMonitor struct {
				Diff int `json:"diff"`
			} `json:"hudson.node_monitors.ClockMonitor"`
		} `json:"monitorData"`
		NumExecutors       int           `json:"numExecutors"`
		Offline            bool          `json:"offline"`
		OfflineCause       interface{}   `json:"offlineCause"`
		OfflineCauseReason string        `json:"offlineCauseReason"`
		OneOffExecutors    []interface{} `json:"oneOffExecutors"`
		TemporarilyOffline bool          `json:"temporarilyOffline"`
	} `json:"computer"`
	DisplayName    string `json:"displayName"`
	TotalExecutors int    `json:"totalExecutors"`
}

func (e *JenkinsExecutors) String() string {
	return fmt.Sprintf("%#v", e)
}

// represents the overall load of the underlying Jenkins instance.
type JenkinsOverallLoad struct {
	AvailableExecutors struct {
		Hour struct {
			History []float64 `json:"history"`
			Latest  float64   `json:"latest"`
		} `json:"hour"`
		Min struct {
			History []float64 `json:"history"`
			Latest  float64   `json:"latest"`
		} `json:"min"`
		Sec10 struct {
			History []float64 `json:"history"`
			Latest  float64   `json:"latest"`
		} `json:"sec10"`
	} `json:"availableExecutors"`
	BusyExecutors struct {
		Hour struct {
			History []float64 `json:"history"`
			Latest  float64   `json:"latest"`
		} `json:"hour"`
		Min struct {
			History []float64 `json:"history"`
			Latest  float64   `json:"latest"`
		} `json:"min"`
		Sec10 struct {
			History []float64 `json:"history"`
			Latest  float64   `json:"latest"`
		} `json:"sec10"`
	} `json:"busyExecutors"`
	ConnectingExecutors struct {
		Hour struct {
			History []float64 `json:"history"`
			Latest  float64   `json:"latest"`
		} `json:"hour"`
		Min struct {
			History []float64 `json:"history"`
			Latest  float64   `json:"latest"`
		} `json:"min"`
		Sec10 struct {
			History []float64 `json:"history"`
			Latest  float64   `json:"latest"`
		} `json:"sec10"`
	} `json:"connectingExecutors"`
	DefinedExecutors struct {
		Hour struct {
			History []float64 `json:"history"`
			Latest  float64   `json:"latest"`
		} `json:"hour"`
		Min struct {
			History []float64 `json:"history"`
			Latest  float64   `json:"latest"`
		} `json:"min"`
		Sec10 struct {
			History []float64 `json:"history"`
			Latest  float64   `json:"latest"`
		} `json:"sec10"`
	} `json:"definedExecutors"`
	IdleExecutors struct {
		Hour struct {
			History []float64 `json:"history"`
			Latest  float64   `json:"latest"`
		} `json:"hour"`
		Min struct {
			History []float64 `json:"history"`
			Latest  float64   `json:"latest"`
		} `json:"min"`
		Sec10 struct {
			History []float64 `json:"history"`
			Latest  float64   `json:"latest"`
		} `json:"sec10"`
	} `json:"idleExecutors"`
	OnlineExecutors struct {
		Hour struct {
			History []float64 `json:"history"`
			Latest  float64   `json:"latest"`
		} `json:"hour"`
		Min struct {
			History []float64 `json:"history"`
			Latest  float64   `json:"latest"`
		} `json:"min"`
		Sec10 struct {
			History []float64 `json:"history"`
			Latest  float64   `json:"latest"`
		} `json:"sec10"`
	} `json:"onlineExecutors"`
	QueueLength struct {
		Hour struct {
			History []float64 `json:"history"`
			Latest  float64   `json:"latest"`
		} `json:"hour"`
		Min struct {
			History []float64 `json:"history"`
			Latest  float64   `json:"latest"`
		} `json:"min"`
		Sec10 struct {
			History []float64 `json:"history"`
			Latest  float64   `json:"latest"`
		} `json:"sec10"`
	} `json:"queueLength"`
	TotalExecutors struct {
		Hour struct {
			History []float64 `json:"history"`
			Latest  float64   `json:"latest"`
		} `json:"hour"`
		Min struct {
			History []float64 `json:"history"`
			Latest  float64   `json:"latest"`
		} `json:"min"`
		Sec10 struct {
			History []float64 `json:"history"`
			Latest  float64   `json:"latest"`
		} `json:"sec10"`
	} `json:"totalExecutors"`
	TotalQueueLength struct {
		Hour struct {
			History []float64 `json:"history"`
			Latest  float64   `json:"latest"`
		} `json:"hour"`
		Min struct {
			History []float64 `json:"history"`
			Latest  float64   `json:"latest"`
		} `json:"min"`
		Sec10 struct {
			History []float64 `json:"history"`
			Latest  float64   `json:"latest"`
		} `json:"sec10"`
	} `json:"totalQueueLength"`
}

func (o *JenkinsOverallLoad) String() string {
	return fmt.Sprintf("%#v", o)
}

// GetQueue retrieves the JenkinsQueue of the underlying Jenkins instance.
// It will return an error, if the connection or the JSON un-marshalling breaks.
func (j *JenkinsClient) GetQueue() (*JenkinsQueue, error) {
	response, err := j.client.GetFrom(queue)
	if err != nil {
		return nil, err
	}

	queue := &JenkinsQueue{}
	if error := j.processQueueResponse(response, queue); error != nil {
		return nil, error
	}

	return queue, nil
}

func (j *JenkinsClient) processQueueResponse(resp *http.Response, queue *JenkinsQueue) error {
	return processResponse(resp, queue, "JenkinsQueue")
}

// GetJobsFromView returns a slice with all Jobs from a given JenkinsView name from the underlying Jenkins instance.
// It will return an error, if the connection or the JSON un-marshalling breaks.
func (j *JenkinsClient) GetJobsFromView(viewName string) ([]JenkinsJob, error) {
	response, err := j.client.GetFrom("view/" + viewName + jsonAPI)
	if err != nil {
		return nil, err
	}

	view := &JenkinsView{}
	if error := j.processViewResponse(response, view); error != nil {
		return nil, error
	}

	return view.Jobs, nil
}

// GetJobsFromViewWithTree returns a slice with all Jobs from a given JenkinsView name, restricting the returned attributes by the given tree string.
// It will return an error, if the connection or the JSON un-marshalling breaks.
func (j *JenkinsClient) GetJobsFromViewWithTree(viewName string, tree string) ([]JenkinsJob, error) {
	response, err := j.client.GetFrom("view/" + viewName + jsonAPI + "?tree=" + tree)
	if err != nil {
		return nil, err
	}

	view := &JenkinsView{}
	if error := j.processViewResponse(response, view); error != nil {
		return nil, error
	}

	return view.Jobs, nil
}

// GetJobsFromViewByPath returns a slice with all Jobs from a given JenkinsView name from the underlying Jenkins instance.
// It will return an error, if the connection or the JSON un-marshalling breaks.
func (j *JenkinsClient) GetJobsFromViewByPath(path string) ([]JenkinsJob, error) {
	response, err := j.client.GetFrom(path + jsonAPI)
	if err != nil {
		return nil, err
	}

	view := &JenkinsView{}
	if error := j.processViewResponse(response, view); error != nil {
		return nil, error
	}

	return view.Jobs, nil
}

// GetJobsFromViewWithTreeByPath returns a slice with all Jobs from a given JenkinsView name, restricting the returned attributes by the given tree string.
// It will return an error, if the connection or the JSON un-marshalling breaks.
func (j *JenkinsClient) GetJobsFromViewWithTreeByPath(path string, tree string) ([]JenkinsJob, error) {
	response, err := j.client.GetFrom(path + jsonAPI + "?tree=" + tree)
	if err != nil {
		return nil, err
	}

	view := &JenkinsView{}
	if error := j.processViewResponse(response, view); error != nil {
		return nil, error
	}

	return view.Jobs, nil
}

func (j *JenkinsClient) processViewResponse(resp *http.Response, view *JenkinsView) error {
	return processResponse(resp, view, "JenkinsView")
}

// GetOverallLoad returns the JenkinsOverallLoad of the underlying Jenkins instance.
// It will return an error, if the connection or the JSON un-marshalling breaks.
func (j *JenkinsClient) GetOverallLoad() (*JenkinsOverallLoad, error) {
	response, err := j.client.GetFrom(overallLoad)
	if err != nil {
		return nil, err
	}

	overallLoad := &JenkinsOverallLoad{}
	if error := j.processOverallLoadResponse(response, overallLoad); error != nil {
		return nil, error
	}
	return overallLoad, nil
}

func (j *JenkinsClient) processOverallLoadResponse(resp *http.Response, overallLoad *JenkinsOverallLoad) error {
	return processResponse(resp, overallLoad, "JenkinsOverallLoad")
}

// GetExecutors returns the currently configured JenkinsExecutors of the underlying Jenkins instance.
// It will return an error, if the connection or the JSON un-marshalling breaks.
func (j *JenkinsClient) GetExecutors() (*JenkinsExecutors, error) {
	response, err := j.client.GetFrom(computer)
	if err != nil {
		return nil, err
	}

	executors := &JenkinsExecutors{}
	if error := j.processExecutorsResponse(response, executors); error != nil {
		return nil, error
	}

	return executors, nil
}

// GetBusyExecutors returns the number of currently occupied JenkinsExecutors of the underlying Jenkins instance.
// It will return an error, if the connection or the JSON un-marshalling breaks.
func (j *JenkinsClient) GetBusyExecutors() (int, error) {
	response, err := j.client.GetFrom(busyExecutors)
	if err != nil {
		return -1, err
	}

	executors := &JenkinsExecutors{}
	if error := j.processExecutorsResponse(response, executors); error != nil {
		return -1, error
	}

	return executors.BusyExecutors, nil
}

func (j *JenkinsClient) processExecutorsResponse(resp *http.Response, executors *JenkinsExecutors) error {
	return processResponse(resp, executors, "JenkinsExecutors")
}

// NewJenkinsClient returns a new Jenkins instance with the given url, username and password.
func NewJenkinsClient(url string, username string, password string) Jenkins {
	config := client.NewHTTPConfig(url, username, password, "application/json")
	client := client.NewHTTPClient(config)

	return &JenkinsClient{client: client}
}

// Process given resp and un-marshall it to the given v.
// Throws either an error if the resp.Body couldn't be read or the un-marshalling failed.
func processResponse(resp *http.Response, v interface{}, component string) error {
	if Debug {
		debugResponse(resp, component)
	}

	if resp.Body != nil {
		defer resp.Body.Close()

		body, error := ioutil.ReadAll(resp.Body)
		if error != nil {
			return fmt.Errorf("Unable to read response body: %s", error)
		}

		error = json.Unmarshal(body, v)
		if error != nil {
			return fmt.Errorf("Error while unmarshalling body: %s", error)
		}
	}

	return nil
}

func debugResponse(resp *http.Response, component string) {
	dumpResponse, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Printf("[DEBUG] %s: %s", component, err)
	}
	log.Printf("[DEBUG][REQ]: %s: %s", component, resp.Request.URL)
	log.Printf("[DEBUG][RESP]: %s: %s", component, string(dumpResponse))
}
