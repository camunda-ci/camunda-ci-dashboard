package dashboard

import (
	"errors"
	"log"
	"reflect"
	"testing"
)

const (
	fixtureJenkinsUrl = "http://ci.jenkins.io"
)

func TestDashboard_GetBrokenTravisBuilds(t *testing.T) {
	i := createDashboardInstanceWithSingleTravisInstance()
	res := i.GetBrokenTravisBuilds()
	expAggrs := 1

	if len(res) != expAggrs {
		log.Fatalf("Wrong number of Travis aggregations returned. Expected %d, got %d",
			expAggrs, len(res))
	}

	expAggr := TravisAggregation{
		Aggregation: Aggregation{
			Name:   "camunda",
			Url:    "https://travis-ci.org/camunda",
			Type:   "travis",
			Status: true},
		Jobs: []TravisJob{
			{Name: "repo1", URL: "https://github.com/org/repo1", Color: "red"},
		}, // only broken jobs returned
	}

	if !reflect.DeepEqual(*res[0], expAggr) {
		log.Fatalf("Wrong aggregation returned. Expected %v, got %v",
			expAggr, *res[0])
	}
}

func TestDashboard_GetBrokenJenkinsBuilds_Happy(t *testing.T) {
	instance := createDashboardInstanceWithSingleJenkinsInstance()

	brokenJenkinsBuilds := instance.GetBrokenJenkinsBuilds()
	if len(brokenJenkinsBuilds) != 1 {
		t.Fatalf("Wrong number of jenkins aggregations returned. Expected 1, but got %d", len(brokenJenkinsBuilds))
	}
	for _, brokenJenkinsBuild := range brokenJenkinsBuilds {
		t.Logf("%+v", *brokenJenkinsBuild)
		if brokenJenkinsBuild.Url != fixtureJenkinsUrl {
			t.FailNow()
		}
	}
}

func TestDashboard_GetBrokenJenkinsBuilds_JenkinsClientReturnsError(t *testing.T) {
	jenkinsInstance := &JenkinsInstance{Name: "Jenkins Public", Url: fixtureJenkinsUrl}
	client := &TestJenkinsClient{
		Name:  jenkinsInstance.Name,
		Url:   jenkinsInstance.Url,
		error: errors.New("timeout"),
		queue: nil,
	}

	instance := createDashboardInstanceWithCustomJenkinsClient([]*JenkinsInstance{jenkinsInstance}, client)

	brokenJenkinsBuilds := instance.GetBrokenJenkinsBuilds()

	if len(brokenJenkinsBuilds) != 1 {
		t.Fatalf("Wrong number of jenkins aggregations returned. Expected 1, but got %d", len(brokenJenkinsBuilds))
	}
	if brokenJenkinsBuilds[0].Status != failed {
		t.Fatal("status should be set to 'not available' in case of errors.")
	}

	for _, brokenJenkinsBuild := range brokenJenkinsBuilds {
		t.Logf("%+v", *brokenJenkinsBuild)
		if brokenJenkinsBuild.Url != fixtureJenkinsUrl {
			t.FailNow()
		}
	}
}

/**
 * Helpers
 */

func createDashboardInstanceWithSingleTravisInstance() *Dashboard {
	travisInstances := []*TravisInstance{
		{
			Name: "camunda",
		},
	}
	return createDashboardInstanceWithMocks(nil, travisInstances, true, nil)
}

func createDashboardInstanceWithSingleJenkinsInstance() *Dashboard {
	jenkinsInstances := []*JenkinsInstance{
		{Name: "Jenkins Public", Url: fixtureJenkinsUrl, BrokenJobsUrl: fixtureJenkinsUrl},
	}

	return createDashboardInstanceWithMocks(jenkinsInstances, nil, true, nil)
}

func createDashboardInstanceWithMocks(jenkinsInstances []*JenkinsInstance, travisInstances []*TravisInstance,
	basicAuth bool, err error) *Dashboard {
	for _, jenkinsInstance := range jenkinsInstances {
		jenkinsClient := &TestJenkinsClient{
			Name:          jenkinsInstance.Name,
			Url:           jenkinsInstance.Url,
			BasicAuth:     basicAuth,
			error:         err,
			busyExecutors: 0,
			executors:     &JenkinsExecutors{},
			jobs:          make([]JenkinsJob, 0),
			overallLoad:   &JenkinsOverallLoad{},
			queue:         &JenkinsQueue{},
		}
		jenkinsInstance.Client = jenkinsClient
	}

	for _, travisInstance := range travisInstances {
		tj1 := TravisJob{Name: "repo1", URL: "https://github.com/org/repo1", Color: "red"}
		tj2 := TravisJob{Name: "repo2", URL: "https://github.com/org/repo2", Color: "green"}
		r1 := TravisRepository{Organization: "org", Name: "repo1", Branch: "master"}
		r2 := TravisRepository{Organization: "org", Name: "repo2", Branch: "feature"}

		travisClient := &TestTravisClient{
			jobs: map[TravisRepository]TravisJob{
				r1: tj1,
				r2: tj2,
			},
			error: nil,
		}
		travisInstance.Client = travisClient
		travisInstance.Repos = []TravisRepository{r1, r2}
	}

	return &Dashboard{
		jenkinsInstances: jenkinsInstances,
		travisInstances:  travisInstances,
	}
}

func createDashboardInstanceWithCustomJenkinsClient(jenkinsInstances []*JenkinsInstance, jenkinsClient *TestJenkinsClient) *Dashboard {
	for _, jenkinsInstance := range jenkinsInstances {
		jenkinsInstance.Client = jenkinsClient
	}

	return &Dashboard{
		jenkinsInstances: jenkinsInstances,
	}
}

/**
 * Test implementation of TravisClient
 */

type TestTravisClient struct {
	jobs  map[TravisRepository]TravisJob
	error error
}

func (t *TestTravisClient) Job(r TravisRepository) (TravisJob, error) {
	if t.error != nil {
		return TravisJob{}, t.error
	}
	return t.jobs[r], nil
}

/**
 * Test implementation of JenkinsClient
 */
type TestJenkinsClient struct {
	Name      string
	Url       string
	BasicAuth bool

	queue         *JenkinsQueue
	jobs          []JenkinsJob
	overallLoad   *JenkinsOverallLoad
	executors     *JenkinsExecutors
	busyExecutors int

	error error
}

func (t *TestJenkinsClient) GetQueue() (*JenkinsQueue, error) {
	if t.error != nil {
		return nil, t.error
	}
	return t.queue, nil
}

func (t *TestJenkinsClient) GetJobsFromView(viewName string) ([]JenkinsJob, error) {
	if t.error != nil {
		return nil, t.error
	}
	return t.jobs, nil
}

func (t *TestJenkinsClient) GetJobsFromViewWithTree(viewName string, tree string) ([]JenkinsJob, error) {
	if t.error != nil {
		return nil, t.error
	}
	return t.jobs, nil
}

func (t *TestJenkinsClient) GetJobsFromViewByPath(path string) ([]JenkinsJob, error) {
	if t.error != nil {
		return nil, t.error
	}
	return t.jobs, nil
}

func (t *TestJenkinsClient) GetJobsFromViewWithTreeByPath(path string, tree string) ([]JenkinsJob, error) {
	if t.error != nil {
		return nil, t.error
	}
	return t.jobs, nil
}

func (t *TestJenkinsClient) GetOverallLoad() (*JenkinsOverallLoad, error) {
	if t.error != nil {
		return nil, t.error
	}
	return t.overallLoad, nil
}

func (t *TestJenkinsClient) GetExecutors() (*JenkinsExecutors, error) {
	if t.error != nil {
		return nil, t.error
	}
	return t.executors, nil
}

func (t *TestJenkinsClient) GetBusyExecutors() (int, error) {
	if t.error != nil {
		return 0, t.error
	}
	return t.busyExecutors, nil
}

func newTestJenkinsClient(name string, url string, basicAuth bool) *TestJenkinsClient {
	var _ Jenkins = (*TestJenkinsClient)(nil)

	return &TestJenkinsClient{
		Name:      name,
		Url:       url,
		BasicAuth: basicAuth,
	}
}
