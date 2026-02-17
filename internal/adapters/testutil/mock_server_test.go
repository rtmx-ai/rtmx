package testutil

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

// TestMockServer validates the mock HTTP server functionality.
// REQ-GO-064: Go CLI shall provide mock HTTP server for adapter tests
func TestMockServer(t *testing.T) {
	server := NewMockServer()
	defer server.Close()

	// Configure a response
	server.ExpectGET("/test", MockResponse{
		StatusCode: 200,
		Body:       `{"status":"ok"}`,
	})

	// Make request
	resp, err := http.Get(server.URL + "/test")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != `{"status":"ok"}` {
		t.Errorf("Unexpected body: %s", body)
	}

	// Verify request was recorded
	if server.RequestCount() != 1 {
		t.Errorf("Expected 1 request, got %d", server.RequestCount())
	}
}

func TestMockServerRecordsRequests(t *testing.T) {
	server := NewMockServer()
	defer server.Close()

	server.ExpectPOST("/api/create", MockResponse{StatusCode: 201})

	// Make POST request with body
	reqBody := `{"name":"test"}`
	req, _ := http.NewRequest("POST", server.URL+"/api/create", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token123")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	resp.Body.Close()

	// Verify request was recorded
	lastReq := server.LastRequest()
	if lastReq == nil {
		t.Fatal("No request recorded")
	}

	if lastReq.Method != "POST" {
		t.Errorf("Expected method POST, got %s", lastReq.Method)
	}

	if lastReq.Path != "/api/create" {
		t.Errorf("Expected path /api/create, got %s", lastReq.Path)
	}

	if lastReq.Headers.Get("Authorization") != "Bearer token123" {
		t.Errorf("Expected Authorization header, got %s", lastReq.Headers.Get("Authorization"))
	}
}

func TestMockServerJSONResponse(t *testing.T) {
	server := NewMockServer()
	defer server.Close()

	// Configure response with struct body (will be JSON encoded)
	server.ExpectGET("/api/user", MockResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"id":   123,
			"name": "test user",
		},
	})

	resp, err := http.Get(server.URL + "/api/user")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}

	if result["name"] != "test user" {
		t.Errorf("Expected name 'test user', got %v", result["name"])
	}
}

func TestMockServerCustomHeaders(t *testing.T) {
	server := NewMockServer()
	defer server.Close()

	server.ExpectGET("/api/data", MockResponse{
		StatusCode: 200,
		Body:       "data",
		Headers: map[string]string{
			"X-Custom-Header": "custom-value",
			"Content-Type":    "text/plain",
		},
	})

	resp, err := http.Get(server.URL + "/api/data")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.Header.Get("X-Custom-Header") != "custom-value" {
		t.Errorf("Expected custom header, got %s", resp.Header.Get("X-Custom-Header"))
	}

	if resp.Header.Get("Content-Type") != "text/plain" {
		t.Errorf("Expected Content-Type text/plain, got %s", resp.Header.Get("Content-Type"))
	}
}

func TestMockServer404ForUnknownPath(t *testing.T) {
	server := NewMockServer()
	defer server.Close()

	resp, err := http.Get(server.URL + "/unknown")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 404 {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestMockServerMultipleMethods(t *testing.T) {
	server := NewMockServer()
	defer server.Close()

	server.ExpectGET("/resource", MockResponse{StatusCode: 200, Body: `{"action":"get"}`})
	server.ExpectPOST("/resource", MockResponse{StatusCode: 201, Body: `{"action":"create"}`})
	server.ExpectPATCH("/resource", MockResponse{StatusCode: 200, Body: `{"action":"update"}`})
	server.ExpectDELETE("/resource", MockResponse{StatusCode: 204})

	tests := []struct {
		method     string
		wantStatus int
	}{
		{"GET", 200},
		{"POST", 201},
		{"PATCH", 200},
		{"DELETE", 204},
	}

	client := &http.Client{}
	for _, tt := range tests {
		req, _ := http.NewRequest(tt.method, server.URL+"/resource", nil)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("%s request failed: %v", tt.method, err)
		}
		resp.Body.Close()

		if resp.StatusCode != tt.wantStatus {
			t.Errorf("%s: expected status %d, got %d", tt.method, tt.wantStatus, resp.StatusCode)
		}
	}
}

func TestMockServerRequestsForPath(t *testing.T) {
	server := NewMockServer()
	defer server.Close()

	server.ExpectGET("/api/v1/items", MockResponse{StatusCode: 200})
	server.ExpectGET("/api/v1/users", MockResponse{StatusCode: 200})

	// Make multiple requests
	_, _ = http.Get(server.URL + "/api/v1/items")
	_, _ = http.Get(server.URL + "/api/v1/items")
	_, _ = http.Get(server.URL + "/api/v1/users")

	itemRequests := server.RequestsForPath("/api/v1/items")
	if len(itemRequests) != 2 {
		t.Errorf("Expected 2 requests to /api/v1/items, got %d", len(itemRequests))
	}

	userRequests := server.RequestsForPath("/api/v1/users")
	if len(userRequests) != 1 {
		t.Errorf("Expected 1 request to /api/v1/users, got %d", len(userRequests))
	}
}

func TestMockServerReset(t *testing.T) {
	server := NewMockServer()
	defer server.Close()

	server.ExpectGET("/test", MockResponse{StatusCode: 200})
	_, _ = http.Get(server.URL + "/test")

	if server.RequestCount() != 1 {
		t.Error("Expected 1 request before reset")
	}

	server.Reset()

	if server.RequestCount() != 0 {
		t.Error("Expected 0 requests after reset")
	}

	// Response should be cleared too
	resp, _ := http.Get(server.URL + "/test")
	if resp.StatusCode != 404 {
		t.Error("Expected 404 after reset clears responses")
	}
}

func TestMockServerConcurrentAccess(t *testing.T) {
	server := NewMockServer()
	defer server.Close()

	server.ExpectGET("/concurrent", MockResponse{StatusCode: 200})

	// Make concurrent requests
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, _ = http.Get(server.URL + "/concurrent")
			done <- true
		}()
	}

	// Wait for all requests
	for i := 0; i < 10; i++ {
		<-done
	}

	if server.RequestCount() != 10 {
		t.Errorf("Expected 10 concurrent requests, got %d", server.RequestCount())
	}
}
