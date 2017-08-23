package http

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

const (
	contentTypeJSON = "application/json"

	fixtureBaseURL   = "https://github.com/camunda"
	fixtureBasicJSON = "{ \"id\": 1 }"

	fixtureHTMLErrorPage = `
<!DOCTYPE html>
<html>
	<head>
		<title>Error</title>
	</head>
	<body>
		<h1>An error occurred.</h1>
		<p>Sorry, the page you are looking for is currently unavailable.<br/>
		Please try again later.</p>
		<p>If you are the system administrator of this resource then you should check
		the <a href="http://nginx.org/r/error_log">error log</a> for details.</p>
		<p><em>Faithfully yours, nginx.</em></p>
	</body>
</html>`
)

func TestDefaultHttpConfig(t *testing.T) {
	defaultHTTPConfig := DefaultHTTPConfig(fixtureBaseURL)

	if defaultHTTPConfig.baseURL != fixtureBaseURL {
		t.Errorf("Expected %s but got %s", fixtureBaseURL, defaultHTTPConfig.baseURL)
	}
	if defaultHTTPConfig.username != "" {
		t.Errorf("Expected \"\" but got %s", defaultHTTPConfig.username)
	}
	if defaultHTTPConfig.password != "" {
		t.Errorf("Expected \"\" but got %s", defaultHTTPConfig.password)
	}
	if defaultHTTPConfig.accept != contentTypeJSON {
		t.Errorf("Expected '%s' but got %s", contentTypeJSON, defaultHTTPConfig.accept)
	}

}

func TestNewDefaultHttpClient(t *testing.T) {
	client := NewDefaultHTTPClient(fixtureBaseURL)

	httpConfig := client.config

	if httpConfig.baseURL != fixtureBaseURL {
		t.Errorf("Expected %s but got %s", fixtureBaseURL, httpConfig.baseURL)
	}
	if httpConfig.username != "" {
		t.Errorf("Expected \"\" but got %s", httpConfig.username)
	}
	if httpConfig.password != "" {
		t.Errorf("Expected \"\" but got %s", httpConfig.password)
	}
	if httpConfig.accept != contentTypeJSON {
		t.Errorf("Expected '%s' but got %s", contentTypeJSON, httpConfig.accept)
	}
}

func TestNewHttpClient(t *testing.T) {
	customHTTPConfig := NewHTTPConfig(fixtureBaseURL, "", "", contentTypeJSON)
	client := NewHTTPClient(customHTTPConfig)

	httpConfig := client.config

	if httpConfig.baseURL != fixtureBaseURL {
		t.Errorf("Expected %s but got %s", fixtureBaseURL, httpConfig.baseURL)
	}
	if httpConfig.username != "" {
		t.Errorf("Expected \"\", got %s", httpConfig.username)
	}
	if httpConfig.password != "" {
		t.Errorf("Expected \"\", got %s", httpConfig.password)
	}
	if httpConfig.accept != contentTypeJSON {
		t.Errorf("Expected '%s' but got %s", contentTypeJSON, httpConfig.accept)
	}
}

func TestHttpClient_GetFrom(t *testing.T) {
	server := mockServer(http.StatusOK, contentTypeJSON, fixtureBasicJSON)
	defer server.Close()

	client := createTestHTTPClient(server.URL)
	resp, _ := client.GetFrom("")

	assertResponseHasStatus(resp, http.StatusOK, t)
	assertResponseBodyIs(resp, fixtureBasicJSON, t)
}

func TestHttpClient_PostTo(t *testing.T) {
	server := mockEchoServer(http.StatusOK)
	defer server.Close()

	client := createTestHTTPClient(server.URL)
	resp, _ := client.PostTo("", strings.NewReader(fixtureBasicJSON))

	assertResponseHasStatus(resp, http.StatusOK, t)
	assertResponseBodyIs(resp, fixtureBasicJSON, t)
}

func TestHttpClient_PutTo(t *testing.T) {
	server := mockEchoServer(http.StatusOK)
	defer server.Close()

	client := createTestHTTPClient(server.URL)
	resp, _ := client.PutTo("", strings.NewReader(fixtureBasicJSON))

	assertResponseHasStatus(resp, http.StatusOK, t)
	assertResponseBodyIs(resp, fixtureBasicJSON, t)
}

func TestHttpClient_DeleteFrom(t *testing.T) {
	f := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", contentTypeJSON)
	}
	server := mockServerWith(http.HandlerFunc(f))
	defer server.Close()

	client := createTestHTTPClient(server.URL)
	resp, _ := client.DeleteFrom("")

	assertResponseHasStatus(resp, http.StatusOK, t)
	assertResponseBodyIs(resp, "", t)
}

func TestHttpClient_GetRequest(t *testing.T) {
	client := createTestHTTPClient(fixtureBaseURL)
	req, _ := client.GetRequest("path")

	assertURLIs(req.URL, fixtureBaseURL+"/path", t)
}

func TestHttpClient_PostRequest(t *testing.T) {

}

func TestHttpClient_PutRequest(t *testing.T) {

}

func TestHttpClient_DeleteRequest(t *testing.T) {

}

func TestHttpClient_StatusCodeErrorHandling(t *testing.T) {
	server := mockServer(http.StatusServiceUnavailable, contentTypeJSON, fixtureHTMLErrorPage)
	defer server.Close()

	client := createTestHTTPClient(server.URL)
	resp, _ := client.GetFrom("503please")

	assertResponseHasStatus(resp, http.StatusServiceUnavailable, t)
	assertResponseBodyIs(resp, fixtureHTMLErrorPage, t)
}

func mockEchoServer(statusCode int) *httptest.Server {
	f := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Header().Set("Content-Type", contentTypeJSON)

		body := ""
		if r.Body != nil {
			defer r.Body.Close()
			bodyAsBytes, _ := ioutil.ReadAll(r.Body)
			body = string(bodyAsBytes)
		}
		fmt.Fprint(w, body)
	}

	return httptest.NewServer(http.HandlerFunc(f))
}

// mockServer returns a pointer to a server to handle the get call.
func mockServer(statusCode int, contentType string, body string) *httptest.Server {
	f := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Header().Set("Content-Type", contentType)
		fmt.Fprint(w, body)
	}

	return httptest.NewServer(http.HandlerFunc(f))
}

func mockServerWith(handlerFunc http.HandlerFunc) *httptest.Server {
	if handlerFunc == nil {
		handlerFunc = func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", contentTypeJSON)
			fmt.Fprint(w, "{}")
		}
	}

	return httptest.NewServer(handlerFunc)
}

func createTestHTTPClient(baseURL string) *HTTPClient {
	config := NewHTTPConfig(baseURL, "", "", contentTypeJSON)
	return NewHTTPClient(config)
}

func assertResponseHasStatus(resp *http.Response, statusCode int, t *testing.T) {
	if resp.StatusCode != statusCode {
		t.Errorf("Expected statuscode %d, got %d.", statusCode, resp.StatusCode)
	}
}

func assertResponseBodyIs(resp *http.Response, expectedValue string, t *testing.T) {
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error("Unexpected error")
	}
	expectedValueAsBytes := []byte(expectedValue)
	if bytes.Compare(body, expectedValueAsBytes) != 0 {
		t.Errorf("Expected body %s, got %s.", expectedValueAsBytes, body)
	}
}

func assertURLIs(actual *url.URL, expectedV interface{}, t *testing.T) {
	switch expected := expectedV.(type) {
	case string:
		if expected != actual.String() {
			t.Fail()
		}
	case *url.URL:
		if expected != actual {
			t.Fail()
		}
	default:
		t.Errorf("Type not supported %v", expected)
	}
}
