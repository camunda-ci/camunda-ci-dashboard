package dashboard

import (
	"sync"
)

const (
	notAvailable Status = "n/a"
	ok           Status = "ok"
)

// Dashboard is a container for all configured JenkinsInstance's.
type Dashboard struct {
	jenkinsInstances []*JenkinsInstance
}

// JenkinsInstance holds basic informations about a Jenkins instance and the client connected to it.
type JenkinsInstance struct {
	Name   string
	Url    string
	Client Jenkins
}

// JenkinsAggregations is a container for all retrieved JenkinsAggregation
type JenkinsAggregations struct {
	jenkinsAggregation []JenkinsAggregation
}

// Status describes the current state of the Jenkins instance.
// Either 'ok', when the client was able to connect to it, or 'not available', when the connection was unsuccessful.
type Status string

// JenkinsAggregation holds all dashboard relevant informations for a Jenkins instance
type JenkinsAggregation struct {
	Name           string `json:"name"`
	Url            string `json:"url"`
	Status         Status `json:"status"`
	BusyExecutors  int    `json:"busyExecutors"`
	BuildQueueSize int    `json:"buildQueueSize"`
	Jobs           []Job  `json:"jobs"`
}

// Init initializes the Dashboard with the given JenkinsInstance's and how to access them.
func Init(jenkinsInstances []*JenkinsInstance, jenkinsUsername string, jenkinsPassword string) *Dashboard {
	// initialize Jenkins clients
	for _, jenkinsInstance := range jenkinsInstances {
		jenkinsInstance.Client = NewJenkinsClient(jenkinsInstance.Url, jenkinsUsername, jenkinsPassword)
	}

	return &Dashboard{
		jenkinsInstances: jenkinsInstances,
	}
}

// GetBrokenJenkinsBuilds retrieves the failed builds displayed on the Broken page from all configured JenkinsInstance's.
func (d *Dashboard) GetBrokenJenkinsBuilds() []*JenkinsAggregation {
	jenkinsAggregations := make([]*JenkinsAggregation, len(d.jenkinsInstances))

	var wg sync.WaitGroup
	wg.Add(len(d.jenkinsInstances))

	for index, jenkinsInstance := range d.jenkinsInstances {
		go func(instance *JenkinsInstance, index int) {
			defer wg.Done()
			jenkinsAggregations[index] = getBrokenBuildsForJenkinsInstance(instance)
		}(jenkinsInstance, index)
	}

	wg.Wait()

	return jenkinsAggregations
}

func getBrokenBuildsForJenkinsInstance(instance *JenkinsInstance) *JenkinsAggregation {
	jenkinsAggregation := &JenkinsAggregation{
		Name: instance.Name,
		Url:  instance.Url,
	}

	var wg sync.WaitGroup
	wg.Add(3)

	go func(instance *JenkinsInstance, aggregation *JenkinsAggregation) {
		defer wg.Done()

		queue, error := instance.Client.GetQueue()
		if error != nil {

		}
		jenkinsAggregation.BuildQueueSize = len(queue.Items)

	}(instance, jenkinsAggregation)

	go func(instance *JenkinsInstance, aggregation *JenkinsAggregation) {
		defer wg.Done()

		jenkinsAggregation.BusyExecutors, _ = instance.Client.GetBusyExecutors()

	}(instance, jenkinsAggregation)

	go func(instance *JenkinsInstance, aggregation *JenkinsAggregation) {
		defer wg.Done()

		tree := "jobs[name,fullDisplayName,color,url,lastBuild[actions[foundFailureCauses[categories,description],failCount,skipCount,totalCount]]]"
		jenkinsAggregation.Jobs, _ = instance.Client.GetJobsFromViewWithTree("Broken", tree)

	}(instance, jenkinsAggregation)

	wg.Wait()

	return jenkinsAggregation
}
