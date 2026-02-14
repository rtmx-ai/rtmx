package adapters

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/rtmx-ai/rtmx-go/internal/config"
)

// MockHTTPClient implements HTTPClient for testing.
type MockHTTPClient struct {
	Response *http.Response
	Err      error
	Requests []*http.Request
}

// Do records the request and returns the configured response.
func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	m.Requests = append(m.Requests, req)
	return m.Response, m.Err
}

// TestHTTPClientInterface validates that http.Client satisfies HTTPClient.
// REQ-GO-061: Go CLI shall provide HTTPClient interface for adapter testing
func TestHTTPClientInterface(t *testing.T) {
	// http.Client must satisfy HTTPClient interface
	var _ HTTPClient = &http.Client{}

	// MockHTTPClient must satisfy HTTPClient interface
	var _ HTTPClient = &MockHTTPClient{}
}

// TestDefaultHTTPClient validates the default HTTP client configuration.
func TestDefaultHTTPClient(t *testing.T) {
	client := DefaultHTTPClient()
	if client == nil {
		t.Fatal("DefaultHTTPClient returned nil")
	}

	// Should be usable as HTTPClient
	var _ HTTPClient = client
}

// TestWithHTTPClient validates HTTP client injection.
func TestWithHTTPClient(t *testing.T) {
	mockClient := &MockHTTPClient{
		Response: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString(`{"full_name":"owner/repo"}`)),
		},
	}

	cfg := config.GitHubAdapterConfig{
		Enabled:  true,
		Repo:     "owner/repo",
		TokenEnv: "TEST_TOKEN",
	}

	// Inject mock HTTP client and mock env getter
	adapter, err := NewGitHubAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string {
			if key == "TEST_TOKEN" {
				return "test-token"
			}
			return ""
		}),
	)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	// Test connection using injected mock
	success, msg := adapter.TestConnection()
	if !success {
		t.Errorf("TestConnection failed: %s", msg)
	}

	// Verify the mock received the request
	if len(mockClient.Requests) != 1 {
		t.Errorf("Expected 1 request, got %d", len(mockClient.Requests))
	}

	// Verify authorization header was set
	req := mockClient.Requests[0]
	if req.Header.Get("Authorization") != "token test-token" {
		t.Errorf("Expected Authorization header 'token test-token', got %s", req.Header.Get("Authorization"))
	}
}

// TestWithEnvGetter validates environment variable injection.
func TestWithEnvGetter(t *testing.T) {
	tests := []struct {
		name      string
		envVars   map[string]string
		wantErr   bool
		errSubstr string
	}{
		{
			name: "all env vars present",
			envVars: map[string]string{
				"TEST_TOKEN": "token-value",
			},
			wantErr: false,
		},
		{
			name:      "missing token",
			envVars:   map[string]string{},
			wantErr:   true,
			errSubstr: "token not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.GitHubAdapterConfig{
				Enabled:  true,
				Repo:     "owner/repo",
				TokenEnv: "TEST_TOKEN",
			}

			_, err := NewGitHubAdapter(&cfg, WithEnvGetter(func(key string) string {
				return tt.envVars[key]
			}))

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestJiraWithHTTPClient validates HTTP client injection for Jira adapter.
func TestJiraWithHTTPClient(t *testing.T) {
	mockClient := &MockHTTPClient{
		Response: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString(`{"key":"PROJ","name":"Project"}`)),
		},
	}

	cfg := config.JiraAdapterConfig{
		Enabled:  true,
		Server:   "https://test.atlassian.net",
		Project:  "PROJ",
		TokenEnv: "TEST_TOKEN",
		EmailEnv: "TEST_EMAIL",
	}

	adapter, err := NewJiraAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string {
			switch key {
			case "TEST_TOKEN":
				return "test-token"
			case "TEST_EMAIL":
				return "test@example.com"
			}
			return ""
		}),
	)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	// Test connection using injected mock
	success, msg := adapter.TestConnection()
	if !success {
		t.Errorf("TestConnection failed: %s", msg)
	}

	// Verify the mock received the request
	if len(mockClient.Requests) != 1 {
		t.Errorf("Expected 1 request, got %d", len(mockClient.Requests))
	}

	// Verify authorization header was set (Basic auth)
	req := mockClient.Requests[0]
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		t.Error("Expected Authorization header to be set")
	}
	if len(authHeader) < 10 || authHeader[:6] != "Basic " {
		t.Errorf("Expected Basic auth header, got %s", authHeader)
	}
}

// TestAdapterOptionsDefaults validates default options behavior.
func TestAdapterOptionsDefaults(t *testing.T) {
	opts := defaultOptions()

	if opts.httpClient == nil {
		t.Error("Default httpClient should not be nil")
	}

	if opts.getEnv == nil {
		t.Error("Default getEnv should not be nil")
	}
}

// TestApplyOptions validates option function application.
func TestApplyOptions(t *testing.T) {
	customClient := &MockHTTPClient{}
	customGetEnv := func(key string) string { return "custom-" + key }

	opts := applyOptions([]AdapterOption{
		WithHTTPClient(customClient),
		WithEnvGetter(customGetEnv),
	})

	if opts.httpClient != customClient {
		t.Error("Custom HTTP client not applied")
	}

	if opts.getEnv("test") != "custom-test" {
		t.Error("Custom getEnv not applied")
	}
}
