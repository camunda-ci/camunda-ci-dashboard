package dashboard

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	client "github.com/camunda-ci/camunda-ci-dashboard/http"
)

func TestJenkinsClient_GetQueue_Request(t *testing.T) {
	testRequest := func(r *http.Request) {
		assertPathIs(r, queue, t)
	}

	server := mockServerForRequestTest(testRequest)
	defer server.Close()

	createTestJenkinsClient(server).
		GetQueue()
}

func TestJenkinsClient_GetQueue_Response(t *testing.T) {
	server := mockSuccesfulResponseWithBodyFromFile("testdata/queue.json", t)
	defer server.Close()

	jenkins := createTestJenkinsClient(server)

	queue, err := jenkins.GetQueue()
	assertNoError(err, t, "queue")

	if len(queue.Items) != 28 {
		t.Fatal("Queue size is wrong.")
	}
}

func TestJenkinsClient_GetJobsFromView_Request(t *testing.T) {
	testRequest := func(r *http.Request) {
		assertPathIs(r, "view/broken"+jsonAPI, t)
	}

	server := mockServerForRequestTest(testRequest)
	defer server.Close()

	createTestJenkinsClient(server).
		GetJobsFromView("broken")
}

func TestJenkinsClient_GetJobsFromView_Response(t *testing.T) {
	server := mockSuccesfulResponseWithBodyFromFile("testdata/view_broken_depth3.json", t)
	defer server.Close()

	jobs, err := createTestJenkinsClient(server).
		GetJobsFromView("broken")
	assertNoError(err, t, "view 'broken'")

	assertSizeOf(jobs, 2, t)
}

func TestJenkinsClient_GetJobsFromNonExistingView(t *testing.T) {
	server := mockFailureResponseWithBodyFromFile(http.StatusInternalServerError, "text/html", "testdata/view_does_not_exist.html", t)
	defer server.Close()

	jobs, err := createTestJenkinsClient(server).
		GetJobsFromView("broken")
	assertHttpStatusError(err, http.StatusInternalServerError, t)

	assertSizeOf(jobs, 0, t)
}

func TestJenkinsClient_GetJobsFromViewWithTree_Request(t *testing.T) {
	tree := "jobs[name,fullDisplayName,color,url,lastBuild[actions[foundFailureCauses[categories,description],failCount,skipCount,totalCount]]]"

	testRequest := func(r *http.Request) {
		assertPathIs(r, "view/broken"+jsonAPI+"?tree="+tree, t)
	}

	server := mockServerForRequestTest(testRequest)
	defer server.Close()

	createTestJenkinsClient(server).
		GetJobsFromViewWithTree("broken", tree)
}

func TestJenkinsClient_GetJobsFromViewWithTree_Response(t *testing.T) {
	// Test makes not so much sense because it relies on the Jenkins doing its thing
	server := mockSuccesfulResponseWithBodyFromFile("testdata/view_broken_depth3.json", t)
	defer server.Close()

	jobs, err := createTestJenkinsClient(server).
		GetJobsFromView("broken")
	assertNoError(err, t, "view 'broken'")

	assertSizeOf(jobs, 2, t)
}

func TestJenkinsClient_GetOverallLoad_Request(t *testing.T) {
	testRequest := func(r *http.Request) {
		assertPathIs(r, "overallLoad"+jsonAPI, t)
	}

	server := mockServerForRequestTest(testRequest)
	defer server.Close()

	createTestJenkinsClient(server).
		GetOverallLoad()
}

func TestJenkinsClient_GetOverallLoad_Response(t *testing.T) {
	server := mockSuccesfulResponseWithBodyFromFile("testdata/overallLoad_depth3.json", t)
	defer server.Close()

	overallLoad, err := createTestJenkinsClient(server).
		GetOverallLoad()
	assertNoError(err, t, "overallLoad")

	assertOverallLoadValues(overallLoad, t)
}

func TestJenkinsClient_GetExecutors_Request(t *testing.T) {
	testRequest := func(r *http.Request) {
		assertPathIs(r, "computer"+jsonAPI, t)
	}

	server := mockServerForRequestTest(testRequest)
	defer server.Close()

	createTestJenkinsClient(server).
		GetExecutors()
}

func TestJenkinsClient_GetExecutors_Response(t *testing.T) {
	server := mockSuccesfulResponseWithBodyFromFile("testdata/computer.json", t)
	defer server.Close()

	executors, err := createTestJenkinsClient(server).
		GetExecutors()
	assertNoError(err, t, "executors")
	if executors.BusyExecutors != 5 {
		t.Fatalf("Wrong number of busy executors")
	}
	if executors.TotalExecutors != 6 {
		t.Fatalf("Wrong number of total executors")
	}
	if executors.DisplayName != "nodes" {
		t.Fatalf("Wrong number of nodes")
	}
}

func TestJenkinsClient_GetBusyExecutors_Request(t *testing.T) {
	testRequest := func(r *http.Request) {
		assertPathIs(r, "computer"+jsonAPI+"?tree=busyExecutors", t)
	}

	server := mockServerForRequestTest(testRequest)
	defer server.Close()

	createTestJenkinsClient(server).
		GetBusyExecutors()
}

func TestJenkinsClient_GetBusyExecutors_Response(t *testing.T) {
	server := mockSuccesfulResponseWithBodyFromFile("testdata/computer.json", t)
	defer server.Close()

	busyExecutors, err := createTestJenkinsClient(server).
		GetBusyExecutors()
	assertNoError(err, t, "busyExecutors")
	if busyExecutors != 5 {
		t.Fatalf("Wrong number of busy executors")
	}

}

func TestJenkinsClientTypesUnmarshalling(t *testing.T) {
	// map with
	// - type
	// - sourcefile
	// - result
}

//
// Test Helpers
//
type testRequestFunction func(r *http.Request)

// mock server and allow to test the receiving request
func mockServerForRequestTest(testRequest testRequestFunction) *httptest.Server {
	return mockServer(http.StatusOK, "application/json", "", testRequest)
}

// mock server and return successful response with body specified by fileName
func mockSuccesfulResponseWithBodyFromFile(fileName string, t *testing.T) *httptest.Server {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Fatalf("Unable to read file: %s. Error: %s", fileName, err)
	}

	return mockServer(http.StatusOK, "application/json", string(content), nil)
}

// mock server and return failure response with body specified by fileName
func mockFailureResponseWithBodyFromFile(statusCode int, contentType string, fileName string, t *testing.T) *httptest.Server {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Fatalf("Unable to read file: %s. Error: %s", fileName, err)
	}

	return mockServer(statusCode, contentType, string(content), nil)
}

func mockServer(statusCode int, contentType string, body string, testRequest testRequestFunction) *httptest.Server {
	f := func(w http.ResponseWriter, r *http.Request) {

		if testRequest != nil {
			testRequest(r)
		}

		w.WriteHeader(statusCode)
		w.Header().Set("Content-Type", contentType)
		fmt.Fprint(w, body)
	}

	return httptest.NewServer(http.HandlerFunc(f))
}

func assertPathIs(req *http.Request, expectedPath string, t *testing.T) {
	normalizedPath := strings.TrimPrefix(req.URL.Path, "/")
	if req.URL.RawQuery != "" {
		// if url has query params, add them
		normalizedPath += "?" + req.URL.RawQuery
	}
	if normalizedPath != expectedPath {
		t.Errorf("Expected path '%s', got '%s'.", expectedPath, normalizedPath)
	}
}

func assertNoError(err error, t *testing.T, msg string) {
	if err != nil {
		t.Errorf("Error while retrieving %s. %s", msg, err)
	}
}

func assertHttpStatusError(err error, expectedStatusCode int, t *testing.T) {
	if err != nil {
		if tmp, ok := err.(*client.HttpError); ok {
			if tmp.StatusCode != expectedStatusCode {
				t.Errorf("Expected status code does not match. Expected %d, got %d", expectedStatusCode, tmp.StatusCode)
			}
		}
	}
}

func assertSizeOf(v []Job, expectedSize int, t *testing.T) {
	if len(v) != expectedSize {
		t.Errorf("Expected size of %d, got %d.\n%+v", expectedSize, len(v), v)
	}
}

func assertOverallLoadValues(overallLoad *OverallLoad, t *testing.T) {
	if overallLoad.AvailableExecutors.Hour.Latest != 0.0 {
		t.Errorf("Wrong number of available executors")
	}
	if overallLoad.BusyExecutors.Hour.Latest != 2.8016105 {
		t.Errorf("Wrong number of busy executors")
	}
	if overallLoad.ConnectingExecutors.Hour.Latest != 0.0 {
		t.Errorf("Wrong number of connecting executors")
	}
	if overallLoad.DefinedExecutors.Hour.Latest != 2.8016105 {
		t.Errorf("Wrong number of defined executors")
	}
	if overallLoad.IdleExecutors.Hour.Latest != 0.0 {
		t.Errorf("Wrong number of idle executors")
	}
	if overallLoad.OnlineExecutors.Hour.Latest != 2.8016105 {
		t.Errorf("Wrong number of online executors")
	}
	if overallLoad.TotalExecutors.Hour.Latest != 2.8016105 {
		t.Errorf("Wrong number of total executors")
	}
	if overallLoad.QueueLength.Hour.Latest != 9.712103 {
		t.Errorf("Wrong queue length")
	}
	if overallLoad.TotalQueueLength.Hour.Latest != 9.712103 {
		t.Errorf("Wrong total queue length")
	}
}

func createTestJenkinsClient(server *httptest.Server) Jenkins {
	return NewJenkinsClient(server.URL, "", "")
}
