package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// UserIdentity holds the decoded identity from an OIDC ID token.
type UserIdentity struct {
	Sub   string `json:"sub"`   // OIDC subject identifier
	Email string `json:"email"` // User email
	Name  string `json:"name"`  // Display name
}

// CurrentUser loads the stored ID token and decodes JWT claims.
// Returns empty identity (not error) when no token file exists.
func CurrentUser(tokenStorePath string) (UserIdentity, error) {
	data, err := os.ReadFile(tokenStorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return UserIdentity{}, nil
		}
		return UserIdentity{}, fmt.Errorf("failed to read token store: %w", err)
	}

	var tokens TokenSet
	if err := json.Unmarshal(data, &tokens); err != nil {
		return UserIdentity{}, fmt.Errorf("failed to parse token store: %w", err)
	}

	if tokens.IDToken == "" {
		return UserIdentity{}, nil
	}

	return decodeIDToken(tokens.IDToken)
}

// decodeIDToken extracts claims from a JWT without signature verification.
// The OIDC provider validated the token at login time.
func decodeIDToken(jwt string) (UserIdentity, error) {
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		return UserIdentity{}, fmt.Errorf("malformed JWT: expected 3 parts, got %d", len(parts))
	}

	// Decode the payload (second part).
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return UserIdentity{}, fmt.Errorf("malformed JWT payload: %w", err)
	}

	var identity UserIdentity
	if err := json.Unmarshal(payload, &identity); err != nil {
		return UserIdentity{}, fmt.Errorf("failed to parse JWT claims: %w", err)
	}

	return identity, nil
}
