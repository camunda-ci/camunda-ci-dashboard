package http

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	jsonType = "application/json"
)

// Client provides a high-level API for working with HTTP requests and constructing them.
type Client interface {
	GetFrom(path string) (*http.Response, error)
	PostTo(path string, body io.Reader) (*http.Response, error)
	PutTo(path string, body io.Reader) (*http.Response, error)
	DeleteFrom(path string) (*http.Response, error)

	GetRequest(path string) (*http.Request, error)
	PostRequest(path string, body io.Reader) (*http.Request, error)
	PutRequest(path string, body io.Reader) (*http.Request, error)
	DeleteRequest(path string) (*http.Request, error)
}

// HTTPConfig holds the base configuration for the HTTPClient.
type HTTPConfig struct {
	baseURL  string
	username string
	password string
	accept   string
}

// HTTPClient wraps the underlying http.Client and it's HTTPConfig.
type HTTPClient struct {
	client *http.Client
	config *HTTPConfig
}

type NotFoundError struct {
	Message string
	Url     string
}

func (e NotFoundError) Error() string {
	return e.Message
}

type UnauthorizedError struct {
	Message string
	Url     string
	Status  int
}

func (e UnauthorizedError) Error() string {
	return e.Message
}

type RemoteError struct {
	Host string
	err  error
}

func (e RemoteError) Error() string {
	return e.err.Error()
}

func NewHTTPConfig(baseURL string, username string, password string, accept string) *HTTPConfig {
	config := &HTTPConfig{
		baseURL:  baseURL,
		username: username,
		password: password,
		accept:   jsonType,
	}

	if accept != "" {
		config.accept = accept
	}

	return config
}

func DefaultHTTPConfig(baseURL string) *HTTPConfig {
	return NewHTTPConfig(baseURL, "", "", jsonType)
}

/**
 * Create a new HTTPClient with a custom transport for clean resource usage
 */
func NewHTTPClient(config *HTTPConfig) *HTTPClient {
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		Timeout: 30 * time.Second,
	}

	return &HTTPClient{
		client: client,
		config: config,
	}
}

func NewDefaultHTTPClient(baseURL string) *HTTPClient {
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		Timeout: 30 * time.Second,
	}

	config := DefaultHTTPConfig(baseURL)

	return &HTTPClient{
		client: client,
		config: config,
	}
}

func (h *HTTPClient) GetFrom(path string) (*http.Response, error) {
	request, error := createRequest(h.config.baseURL, path, http.MethodGet, nil, h.config.username, h.config.password)
	if error != nil {
		return nil, error
	}
	return h.executeRequest(request)
}

func (h *HTTPClient) PostTo(path string, body io.Reader) (*http.Response, error) {
	request, error := createRequest(h.config.baseURL, path, http.MethodPost, body, h.config.username, h.config.password)
	if error != nil {
		return nil, error
	}
	return h.executeRequest(request)
}

func (h *HTTPClient) PutTo(path string, body io.Reader) (*http.Response, error) {
	request, error := createRequest(h.config.baseURL, path, http.MethodPut, body, h.config.username, h.config.password)
	if error != nil {
		return nil, error
	}
	return h.executeRequest(request)
}

func (h *HTTPClient) DeleteFrom(path string) (*http.Response, error) {
	request, error := createRequest(h.config.baseURL, path, http.MethodDelete, nil, h.config.username, h.config.password)
	if error != nil {
		return nil, error
	}
	return h.executeRequest(request)
}

func (h *HTTPClient) GetRequest(path string) (*http.Request, error) {
	return createRequest(h.config.baseURL, path, http.MethodGet, nil, h.config.username, h.config.password)
}

func (h *HTTPClient) PostRequest(path string, body io.Reader) (*http.Request, error) {
	return createRequest(h.config.baseURL, path, http.MethodPost, body, h.config.username, h.config.password)
}

func (h *HTTPClient) PutRequest(path string, body io.Reader) (*http.Request, error) {
	return createRequest(h.config.baseURL, path, http.MethodPut, body, h.config.username, h.config.password)
}

func (h *HTTPClient) DeleteRequest(path string) (*http.Request, error) {
	return createRequest(h.config.baseURL, path, http.MethodDelete, nil, h.config.username, h.config.password)
}

func createRequest(baseURL string, endpoint string, method string, body io.Reader, username string, password string) (*http.Request, error) {
	// construct url by appending endpoint to base url
	baseURL = strings.TrimSuffix(baseURL, "/")

	request, err := http.NewRequest(method, baseURL+"/"+endpoint, body)
	if err != nil {
		return request, err
	}

	request.Header.Set("Content-Type", jsonType)
	request.Header.Set("Accept", jsonType)

	if username != "" && password != "" {
		request.SetBasicAuth(username, password)
	}

	return request, nil
}

func (h *HTTPClient) executeRequest(r *http.Request) (*http.Response, error) {
	resp, error := h.client.Do(r)

	if error != nil {
		return handleError(resp, error)
	}

	return resp, nil
}

func handleError(resp *http.Response, error error) (*http.Response, error) {
	log.Fatal(error)

	if resp.StatusCode == http.StatusUnauthorized {
		return resp, &UnauthorizedError{Message: "Authentication required.", Url: resp.Request.URL.String()}
	}

	if resp.StatusCode == http.StatusNotFound {
		return resp, &NotFoundError{Message: "Resource not found.", Url: resp.Request.URL.String()}
	}

	return resp, &RemoteError{resp.Request.URL.Host, fmt.Errorf("%d: (%s)", resp.StatusCode, resp.Request.URL.String())}
}
