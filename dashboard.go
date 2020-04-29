package dashboard

import (
	"fmt"
	"log"
	"strings"
	"sync"
)

const (
	failed Status = false
	ok     Status = true
)

// Dashboard is a container for all configured JenkinsInstance's.
type Dashboard struct {
	jenkinsInstances []*JenkinsInstance
	travisInstances  []*TravisInstance
}

// status describes the current state of the instance.
// 'ok', when the client was able to connect to it, or 'not available', when the connection was unsuccessful.
type Status bool

type Aggregation struct {
	Name   string `json:"name"`
	Url    string `json:"url"`
	Type   string `json:"type"`
	Status Status `json:"status"`
}

// Init initializes the Dashboard with the given JenkinsInstance's and how to access them.
func Init(jenkinsInstances []*JenkinsInstance, travisInstances []*TravisInstance,
	jenkinsUsername string, jenkinsPassword string) *Dashboard {
	// initialize Jenkins clients
	for _, jenkinsInstance := range jenkinsInstances {
		jenkinsInstance.Client = NewJenkinsClient(jenkinsInstance.Url, jenkinsUsername, jenkinsPassword)
	}

	return &Dashboard{
		jenkinsInstances: jenkinsInstances,
		travisInstances:  travisInstances,
	}
}

func getBrokenBuildsForTravisInstance(instance *TravisInstance) *TravisAggregation {
	count := len(instance.Repos)
	aggregation := &TravisAggregation{
		Aggregation: Aggregation{
			Type:   "travis",
			Name:   instance.Name,
			Url:    instance.Url(),
			Status: ok,
		},
		Jobs: make([]TravisJob, count),
	}

	var wg sync.WaitGroup
	wg.Add(count)

	for i, r := range instance.Repos {
		go func(index int, repo TravisRepository) {
			defer wg.Done()
			job, _ := instance.Client.Job(repo)
			aggregation.Jobs[index] = job
		}(i, r)
	}

	wg.Wait()

	failedJobs := make([]TravisJob, 0)
	for _, job := range aggregation.Jobs {
		if !job.IsSuccessful() {
			failedJobs = append(failedJobs, job)
		}
	}
	aggregation.Jobs = failedJobs

	return aggregation
}

func (d *Dashboard) GetBrokenTravisBuilds() []*TravisAggregation {
	count := len(d.travisInstances)
	aggregations := make([]*TravisAggregation, count)

	var wg sync.WaitGroup
	wg.Add(count)

	for index, instance := range d.travisInstances {
		go func(ix int, i *TravisInstance) {
			defer wg.Done()
			aggregations[ix] = getBrokenBuildsForTravisInstance(i)
		}(index, instance)
	}

	wg.Wait()
	return aggregations
}

// GetBrokenJenkinsBuilds retrieves the failed builds displayed on the Broken page from all configured JenkinsInstance's.
func (d *Dashboard) GetBrokenJenkinsBuilds() []*JenkinsAggregation {
	count := len(d.jenkinsInstances)
	jenkinsAggregations := make([]*JenkinsAggregation, count)

	var wg sync.WaitGroup
	wg.Add(count)

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
	if instance.BrokenJobsUrl == "" {
		instance.BrokenJobsUrl = instance.Url
	}

	jenkinsAggregation := &JenkinsAggregation{
		Aggregation: Aggregation{
			Type:   "jenkins",
			Name:   instance.Name,
			Url:    instance.Url,
			Status: ok,
		},
		BrokenJobsUrl: instance.BrokenJobsUrl,
		PublicUrl:     instance.PublicUrl,
	}

	var wg sync.WaitGroup
	wg.Add(3)

	go func(instance *JenkinsInstance, aggregation *JenkinsAggregation) {
		defer wg.Done()

		queue, err := instance.Client.GetQueue()
		if err != nil {
			log.Printf("[WARN] %s", err)
			aggregation.BuildQueueSize = 0
			aggregation.Status = failed
			return
		}
		aggregation.BuildQueueSize = len(queue.Items)

	}(instance, jenkinsAggregation)

	go func(instance *JenkinsInstance, aggregation *JenkinsAggregation) {
		defer wg.Done()

		currentBusyExecutors, err := instance.Client.GetBusyExecutors()
		if err != nil {
			log.Printf("[WARN] %s", err)
			aggregation.BusyExecutors = 0
			aggregation.Status = failed
			return
		}
		aggregation.BusyExecutors = currentBusyExecutors

	}(instance, jenkinsAggregation)

	go func(instance *JenkinsInstance, aggregation *JenkinsAggregation) {
		defer wg.Done()

		tree := "jobs[name,fullDisplayName,color,url,lastBuild[actions[foundFailureCauses[categories,description],failCount,skipCount,totalCount]]]"

		path := getBrokenJobsPath(instance)
		jobs, err := instance.Client.GetJobsFromViewWithTreeByPath(path+"/view/Broken", tree)
		if err != nil {
			log.Printf("[WARN] %s", err)
			aggregation.Jobs = make([]JenkinsJob, 0)
			aggregation.Status = failed
			return
		}
		aggregation.Jobs = jobs

	}(instance, jenkinsAggregation)

	wg.Wait()

	return jenkinsAggregation
}

func getBrokenJobsPath(instance *JenkinsInstance) string {
	if strings.HasPrefix(instance.BrokenJobsUrl, instance.Url) {
		return strings.TrimPrefix(instance.BrokenJobsUrl, instance.Url)
	}

	panic(fmt.Sprintf("Instance URL '%s' must be part of broken jobs URL '%s'.", instance.Url, instance.BrokenJobsUrl))
}
