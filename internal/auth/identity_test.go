package auth

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// makeJWT builds a test JWT from a claims map (no real signature).
func makeJWT(t *testing.T, claims map[string]interface{}) string {
	t.Helper()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}
	payloadB64 := base64.RawURLEncoding.EncodeToString(payload)
	return header + "." + payloadB64 + ".fakesignature"
}

func writeTokenFile(t *testing.T, dir string, tokens TokenSet) string {
	t.Helper()
	path := filepath.Join(dir, "tokens.json")
	data, err := json.Marshal(tokens)
	if err != nil {
		t.Fatalf("marshal tokens: %v", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("write tokens: %v", err)
	}
	return path
}

func TestCurrentUser(t *testing.T) {
	rtmx.Req(t, "REQ-PLAN-008")

	tests := []struct {
		name      string
		setup     func(t *testing.T, dir string) string
		wantEmail string
		wantName  string
		wantSub   string
		wantErr   bool
		wantEmpty bool
	}{
		{
			name: "valid_id_token",
			setup: func(t *testing.T, dir string) string {
				jwt := makeJWT(t, map[string]interface{}{
					"sub":   "user-123",
					"email": "alice@example.com",
					"name":  "Alice Smith",
				})
				return writeTokenFile(t, dir, TokenSet{
					AccessToken: "access",
					IDToken:     jwt,
				})
			},
			wantSub:   "user-123",
			wantEmail: "alice@example.com",
			wantName:  "Alice Smith",
		},
		{
			name: "missing_token_file",
			setup: func(t *testing.T, dir string) string {
				return filepath.Join(dir, "nonexistent.json")
			},
			wantEmpty: true,
		},
		{
			name: "empty_id_token",
			setup: func(t *testing.T, dir string) string {
				return writeTokenFile(t, dir, TokenSet{
					AccessToken: "access",
					IDToken:     "",
				})
			},
			wantEmpty: true,
		},
		{
			name: "malformed_jwt",
			setup: func(t *testing.T, dir string) string {
				return writeTokenFile(t, dir, TokenSet{
					AccessToken: "access",
					IDToken:     "not-a-jwt",
				})
			},
			wantErr: true,
		},
		{
			name: "invalid_base64_payload",
			setup: func(t *testing.T, dir string) string {
				return writeTokenFile(t, dir, TokenSet{
					AccessToken: "access",
					IDToken:     "header.!!!invalid!!!.sig",
				})
			},
			wantErr: true,
		},
		{
			name: "partial_claims",
			setup: func(t *testing.T, dir string) string {
				jwt := makeJWT(t, map[string]interface{}{
					"sub": "user-456",
				})
				return writeTokenFile(t, dir, TokenSet{
					AccessToken: "access",
					IDToken:     jwt,
				})
			},
			wantSub:   "user-456",
			wantEmail: "",
			wantName:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := tt.setup(t, dir)

			identity, err := CurrentUser(path)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantEmpty {
				if identity.Sub != "" || identity.Email != "" || identity.Name != "" {
					t.Fatalf("expected empty identity, got %+v", identity)
				}
				return
			}

			if identity.Sub != tt.wantSub {
				t.Errorf("Sub = %q, want %q", identity.Sub, tt.wantSub)
			}
			if identity.Email != tt.wantEmail {
				t.Errorf("Email = %q, want %q", identity.Email, tt.wantEmail)
			}
			if identity.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", identity.Name, tt.wantName)
			}
		})
	}
}
