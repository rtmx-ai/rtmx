// Package testutil provides test utilities for adapter testing.
package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
)

// MockResponse defines a response for a mock endpoint.
type MockResponse struct {
	StatusCode int
	Body       interface{} // Will be JSON encoded if not a string
	Headers    map[string]string
}

// RecordedRequest captures details about a request made to the mock server.
type RecordedRequest struct {
	Method  string
	Path    string
	Headers http.Header
	Body    []byte
}

// MockServer provides a configurable HTTP test server for adapter testing.
// It records all requests for verification and returns configured responses.
type MockServer struct {
	*httptest.Server
	mu        sync.Mutex
	responses map[string]map[string]MockResponse // method -> path -> response
	Requests  []RecordedRequest
}

// NewMockServer creates a new mock HTTP server.
func NewMockServer() *MockServer {
	m := &MockServer{
		responses: make(map[string]map[string]MockResponse),
		Requests:  make([]RecordedRequest, 0),
	}

	m.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.handleRequest(w, r)
	}))

	return m
}

// ExpectRequest configures a response for a specific method and path.
func (m *MockServer) ExpectRequest(method, path string, resp MockResponse) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.responses[method] == nil {
		m.responses[method] = make(map[string]MockResponse)
	}
	m.responses[method][path] = resp
}

// ExpectGET is a convenience method for GET requests.
func (m *MockServer) ExpectGET(path string, resp MockResponse) {
	m.ExpectRequest("GET", path, resp)
}

// ExpectPOST is a convenience method for POST requests.
func (m *MockServer) ExpectPOST(path string, resp MockResponse) {
	m.ExpectRequest("POST", path, resp)
}

// ExpectPATCH is a convenience method for PATCH requests.
func (m *MockServer) ExpectPATCH(path string, resp MockResponse) {
	m.ExpectRequest("PATCH", path, resp)
}

// ExpectPUT is a convenience method for PUT requests.
func (m *MockServer) ExpectPUT(path string, resp MockResponse) {
	m.ExpectRequest("PUT", path, resp)
}

// ExpectDELETE is a convenience method for DELETE requests.
func (m *MockServer) ExpectDELETE(path string, resp MockResponse) {
	m.ExpectRequest("DELETE", path, resp)
}

// handleRequest processes an incoming request.
func (m *MockServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Read body for recording
	var body []byte
	if r.Body != nil {
		body, _ = readBody(r)
	}

	// Record the request
	m.Requests = append(m.Requests, RecordedRequest{
		Method:  r.Method,
		Path:    r.URL.Path,
		Headers: r.Header.Clone(),
		Body:    body,
	})

	// Find matching response
	if methodResponses, ok := m.responses[r.Method]; ok {
		if resp, ok := methodResponses[r.URL.Path]; ok {
			m.writeResponse(w, resp)
			return
		}
	}

	// Default 404 response
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte(`{"error":"no mock response configured"}`))
}

// writeResponse writes the configured response.
func (m *MockServer) writeResponse(w http.ResponseWriter, resp MockResponse) {
	// Set headers
	for k, v := range resp.Headers {
		w.Header().Set(k, v)
	}

	// Set default content type if not specified
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}

	// Set status code
	if resp.StatusCode == 0 {
		resp.StatusCode = http.StatusOK
	}
	w.WriteHeader(resp.StatusCode)

	// Write body
	if resp.Body != nil {
		switch body := resp.Body.(type) {
		case string:
			_, _ = w.Write([]byte(body))
		case []byte:
			_, _ = w.Write(body)
		default:
			// JSON encode
			_ = json.NewEncoder(w).Encode(body)
		}
	}
}

// RequestCount returns the number of requests received.
func (m *MockServer) RequestCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.Requests)
}

// LastRequest returns the most recent request, or nil if none.
func (m *MockServer) LastRequest() *RecordedRequest {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.Requests) == 0 {
		return nil
	}
	return &m.Requests[len(m.Requests)-1]
}

// RequestsForPath returns all requests to a specific path.
func (m *MockServer) RequestsForPath(path string) []RecordedRequest {
	m.mu.Lock()
	defer m.mu.Unlock()

	var result []RecordedRequest
	for _, req := range m.Requests {
		if req.Path == path {
			result = append(result, req)
		}
	}
	return result
}

// Reset clears all recorded requests and configured responses.
func (m *MockServer) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Requests = make([]RecordedRequest, 0)
	m.responses = make(map[string]map[string]MockResponse)
}

// readBody reads the request body.
func readBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}

	body := make([]byte, r.ContentLength)
	if r.ContentLength > 0 {
		_, err := r.Body.Read(body)
		return body, err
	}
	return nil, nil
}
