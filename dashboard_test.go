package dashboard

import (
	"errors"
	"testing"
)

const (
	fixtureJenkinsUrl = "http://ci.jenkins.io"
)

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
		t.Fatal("Status should be set to 'not available' in case of errors.")
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
func createDashboardInstanceWithSingleJenkinsInstance() *Dashboard {
	jenkinsInstances := []*JenkinsInstance{
		{Name: "Jenkins Public", Url: fixtureJenkinsUrl, BrokenJobsUrl: fixtureJenkinsUrl},
	}

	return createDashboardInstanceWithMocks(jenkinsInstances, true, nil)
}

func createDashboardInstanceWithMocks(jenkinsInstances []*JenkinsInstance, basicAuth bool, err error) *Dashboard {
	for _, jenkinsInstance := range jenkinsInstances {
		jenkinsClient := &TestJenkinsClient{
			Name:          jenkinsInstance.Name,
			Url:           jenkinsInstance.Url,
			BasicAuth:     basicAuth,
			error:         err,
			busyExecutors: 0,
			executors:     &Executors{},
			jobs:          make([]Job, 0),
			overallLoad:   &OverallLoad{},
			queue:         &Queue{},
		}
		jenkinsInstance.Client = jenkinsClient
	}

	return &Dashboard{
		jenkinsInstances: jenkinsInstances,
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
 * Test implementation of JenkinsClient
 */
type TestJenkinsClient struct {
	Name      string
	Url       string
	BasicAuth bool

	queue         *Queue
	jobs          []Job
	overallLoad   *OverallLoad
	executors     *Executors
	busyExecutors int

	error error
}

func (t *TestJenkinsClient) GetQueue() (*Queue, error) {
	if t.error != nil {
		return nil, t.error
	}
	return t.queue, nil
}

func (t *TestJenkinsClient) GetJobsFromView(viewName string) ([]Job, error) {
	if t.error != nil {
		return nil, t.error
	}
	return t.jobs, nil
}

func (t *TestJenkinsClient) GetJobsFromViewWithTree(viewName string, tree string) ([]Job, error) {
	if t.error != nil {
		return nil, t.error
	}
	return t.jobs, nil
}

func (t *TestJenkinsClient) GetJobsFromViewByPath(path string) ([]Job, error) {
	if t.error != nil {
		return nil, t.error
	}
	return t.jobs, nil
}

func (t *TestJenkinsClient) GetJobsFromViewWithTreeByPath(path string, tree string) ([]Job, error) {
	if t.error != nil {
		return nil, t.error
	}
	return t.jobs, nil
}

func (t *TestJenkinsClient) GetOverallLoad() (*OverallLoad, error) {
	if t.error != nil {
		return nil, t.error
	}
	return t.overallLoad, nil
}

func (t *TestJenkinsClient) GetExecutors() (*Executors, error) {
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
