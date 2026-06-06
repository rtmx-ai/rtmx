package cmd

import "net/http"

// authConfig holds authentication settings for the dashboard.
type authConfig struct {
	Mode   string // "", "api-key", "oauth"
	APIKey string // required when Mode == "api-key"
}

// authMiddleware returns middleware that enforces authentication.
// When cfg.Mode is empty, all requests pass through.
// When cfg.Mode is "api-key", requests must include the correct
// Authorization: Bearer <key> header.
func authMiddleware(cfg authConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.Mode == "" {
				next.ServeHTTP(w, r)
				return
			}

			if cfg.Mode == "api-key" {
				auth := r.Header.Get("Authorization")
				if auth == "" {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				// Expect "Bearer <key>"
				const prefix = "Bearer "
				if len(auth) <= len(prefix) || auth[:len(prefix)] != prefix {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				key := auth[len(prefix):]
				if key != cfg.APIKey {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			// Unknown auth mode -- deny by default
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		})
	}
}
