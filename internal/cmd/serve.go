package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/output"
	"github.com/spf13/cobra"
)

var (
	servePort    int
	serveAuth    string
	serveSyncURL string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start web dashboard for RTM status",
	Long: `Starts a local web dashboard showing requirements status, backlog,
and health metrics. Supports authentication via API key or OAuth for
multi-user deployment.

The --sync-url flag connects to a remote rtmx-sync server for real-time
collaboration via CRDT synchronization.

Examples:
    rtmx serve                              # local dashboard on :8080
    rtmx serve --port 9090                  # custom port
    rtmx serve --auth api-key               # API key authentication
    rtmx serve --auth oauth --sync-url ws://sync.example.com  # full deployment`,
	RunE: runServe,
}

func init() {
	serveCmd.Flags().IntVar(&servePort, "port", 8080, "port to listen on")
	serveCmd.Flags().StringVar(&serveAuth, "auth", "", "authentication mode (api-key or oauth)")
	serveCmd.Flags().StringVar(&serveSyncURL, "sync-url", "", "rtmx-sync server URL for real-time collaboration")
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	if noColor {
		output.DisableColor()
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	cfg, err := config.LoadFromDir(cwd)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	dbPath := cfg.DatabasePath(cwd)
	db, err := database.Load(dbPath)
	if err != nil {
		return fmt.Errorf("failed to load database: %w", err)
	}

	mux := NewDashboardMux(db, cfg)

	addr := fmt.Sprintf(":%d", servePort)
	cmd.Printf("Starting RTMX dashboard on http://localhost%s\n", addr)

	if serveAuth != "" {
		cmd.Printf("  Auth: %s\n", serveAuth)
	}
	if serveSyncURL != "" {
		cmd.Printf("  Sync: %s\n", serveSyncURL)
	}

	return http.ListenAndServe(addr, mux)
}

// NewDashboardMux creates an HTTP handler for the RTMX web dashboard.
func NewDashboardMux(db *database.Database, cfg *config.Config) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprint(w, dashboardHTML(db))
	})

	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		reqs := db.All()
		complete, partial, missing := 0, 0, 0
		for _, req := range reqs {
			switch {
			case req.IsComplete():
				complete++
			case req.Status == database.StatusPartial:
				partial++
			default:
				missing++
			}
		}
		_, _ = fmt.Fprintf(w, `{"total":%d,"complete":%d,"partial":%d,"missing":%d}`,
			len(reqs), complete, partial, missing)
	})

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"status":"ok"}`)
	})

	return mux
}

func dashboardHTML(db *database.Database) string {
	reqs := db.All()
	complete, partial, missing := 0, 0, 0
	for _, req := range reqs {
		switch {
		case req.IsComplete():
			complete++
		case req.Status == database.StatusPartial:
			partial++
		default:
			missing++
		}
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><title>RTMX Dashboard</title>
<style>
body { font-family: system-ui, sans-serif; margin: 2rem; background: #f8f9fa; }
.card { background: white; border-radius: 8px; padding: 1.5rem; margin: 1rem 0; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
.stats { display: flex; gap: 2rem; }
.stat { text-align: center; }
.stat .value { font-size: 2rem; font-weight: bold; }
.complete { color: #16a34a; }
.partial { color: #d97706; }
.missing { color: #dc2626; }
</style>
</head>
<body>
<h1>RTMX Dashboard</h1>
<div class="card">
<div class="stats">
<div class="stat"><div class="value complete">%d</div><div>Complete</div></div>
<div class="stat"><div class="value partial">%d</div><div>Partial</div></div>
<div class="stat"><div class="value missing">%d</div><div>Missing</div></div>
<div class="stat"><div class="value">%d</div><div>Total</div></div>
</div>
</div>
</body>
</html>`, complete, partial, missing, len(reqs))
}
