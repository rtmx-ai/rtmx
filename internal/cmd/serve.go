package cmd

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/graph"
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

The --sync-url flag connects to a remote sync server for real-time
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
	serveCmd.Flags().StringVar(&serveSyncURL, "sync-url", "", "sync server URL for real-time collaboration")
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

	mux := NewDashboardMuxWithPath(db, cfg, dbPath)

	// Apply auth middleware when --auth is set
	handler := http.Handler(mux)
	if serveAuth != "" {
		apiKey := os.Getenv("RTMX_API_KEY")
		handler = authMiddleware(authConfig{Mode: serveAuth, APIKey: apiKey})(mux)
	}

	addr := fmt.Sprintf(":%d", servePort)
	cmd.Printf("Starting RTMX dashboard on http://localhost%s\n", addr)

	if serveAuth != "" {
		cmd.Printf("  Auth: %s\n", serveAuth)
	}
	if serveSyncURL != "" {
		cmd.Printf("  Sync: %s\n", serveSyncURL)
	}

	return http.ListenAndServe(addr, handler)
}

// NewDashboardMux creates an HTTP handler for the RTMX web dashboard.
func NewDashboardMux(db *database.Database, cfg *config.Config) http.Handler {
	return NewDashboardMuxWithPath(db, cfg, "")
}

// NewDashboardMuxWithPath creates an HTTP handler with an explicit database path for persistence.
func NewDashboardMuxWithPath(db *database.Database, cfg *config.Config, dbPath string) http.Handler {
	mux := http.NewServeMux()
	mu := &sync.Mutex{} // protects writes to db

	// Derive claims dir from dbPath
	claimsDir := ""
	if dbPath != "" {
		claimsDir = filepath.Join(filepath.Dir(dbPath), "claims")
	}

	mux.HandleFunc("/", handleDashboardShell(db))

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

	mux.HandleFunc("/api/requirements", handleAPIRequirements(db))
	mux.HandleFunc("/api/requirements/", handleAPIRequirementDetail(db, dbPath, mu))
	mux.HandleFunc("/api/graph", handleAPIGraph(db))
	mux.HandleFunc("/api/backlog", handleAPIBacklog(db))
	mux.HandleFunc("/api/releases", handleAPIReleases(db))
	mux.HandleFunc("/api/releases/", handleAPIReleaseDetail(db))
	mux.HandleFunc("/api/agents/claims", handleAPIAgentClaims(db, claimsDir, 15*time.Minute))

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"status":"ok"}`)
	})

	// Slack slash command endpoint
	slackSecret := os.Getenv("SLACK_SIGNING_SECRET")
	mux.HandleFunc("/api/slack/command", handleSlackSlashCommand(db, slackSecret))

	// Dashboard SPA and partials
	g := graph.NewGraph(db)
	registerDashboardRoutes(mux, db, g)

	return mux
}

