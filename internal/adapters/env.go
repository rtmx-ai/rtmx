package adapters

import "os"

// defaultGetEnv is the production environment variable getter.
func defaultGetEnv(key string) string {
	return os.Getenv(key)
}
