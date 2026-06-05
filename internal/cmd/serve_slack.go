package cmd

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rtmx-ai/rtmx/internal/database"
)

// slackBlock represents a Slack Block Kit block.
type slackBlock struct {
	Type string      `json:"type"`
	Text *slackText  `json:"text,omitempty"`
	Fields []slackText `json:"fields,omitempty"`
}

// slackText represents a text object in Block Kit.
type slackText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// slackResponse is the Block Kit response for a slash command.
type slackResponse struct {
	ResponseType string       `json:"response_type"`
	Blocks       []slackBlock `json:"blocks"`
}

// handleSlackSlashCommand creates an HTTP handler for Slack slash commands.
// It verifies Slack request signatures and routes subcommands:
//   /rtmx status   -- project status summary
//   /rtmx backlog  -- top backlog items
//   /rtmx req REQ-ID -- requirement detail
func handleSlackSlashCommand(db *database.Database, signingSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeAPIError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		// Read body for signature verification
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "failed to read request body")
			return
		}

		// Verify Slack signature
		if signingSecret != "" {
			if !verifySlackSignature(r.Header, body, signingSecret) {
				writeAPIError(w, http.StatusUnauthorized, "invalid signature")
				return
			}
		}

		// Parse form data from body
		values, err := parseFormBody(string(body))
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "failed to parse form data")
			return
		}

		commandText := strings.TrimSpace(values.Get("text"))
		parts := strings.Fields(commandText)

		var resp slackResponse
		if len(parts) == 0 {
			resp = slackStatusResponse(db)
		} else {
			switch parts[0] {
			case "status":
				resp = slackStatusResponse(db)
			case "backlog":
				resp = slackBacklogResponse(db)
			case "req":
				if len(parts) < 2 {
					resp = slackErrorResponse("Usage: /rtmx req REQ-ID")
				} else {
					resp = slackReqDetailResponse(db, parts[1])
				}
			default:
				resp = slackErrorResponse(fmt.Sprintf("Unknown command: %s. Try: status, backlog, req REQ-ID", parts[0]))
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}

// verifySlackSignature verifies the Slack request signature using HMAC-SHA256.
func verifySlackSignature(headers http.Header, body []byte, signingSecret string) bool {
	timestamp := headers.Get("X-Slack-Request-Timestamp")
	signature := headers.Get("X-Slack-Signature")

	if timestamp == "" || signature == "" {
		return false
	}

	// Reject requests older than 5 minutes to prevent replay attacks
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}
	if abs(time.Now().Unix()-ts) > 300 {
		return false
	}

	// Compute expected signature
	baseString := fmt.Sprintf("v0:%s:%s", timestamp, string(body))
	mac := hmac.New(sha256.New, []byte(signingSecret))
	mac.Write([]byte(baseString))
	expected := "v0=" + hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expected))
}

func abs(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}

// slackStatusResponse builds the Block Kit response for /rtmx status.
func slackStatusResponse(db *database.Database) slackResponse {
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

	total := len(reqs)
	pct := 0.0
	if total > 0 {
		pct = float64(complete) / float64(total) * 100.0
	}

	return slackResponse{
		ResponseType: "in_channel",
		Blocks: []slackBlock{
			{
				Type: "header",
				Text: &slackText{Type: "plain_text", Text: "RTMX Project Status"},
			},
			{
				Type: "section",
				Fields: []slackText{
					{Type: "mrkdwn", Text: fmt.Sprintf("*Total:* %d", total)},
					{Type: "mrkdwn", Text: fmt.Sprintf("*Completion:* %.0f%%", pct)},
					{Type: "mrkdwn", Text: fmt.Sprintf("*Complete:* %d", complete)},
					{Type: "mrkdwn", Text: fmt.Sprintf("*Partial:* %d", partial)},
					{Type: "mrkdwn", Text: fmt.Sprintf("*Missing:* %d", missing)},
				},
			},
		},
	}
}

// slackBacklogResponse builds the Block Kit response for /rtmx backlog.
func slackBacklogResponse(db *database.Database) slackResponse {
	opts := database.FilterOptions{}
	isComplete := false
	opts.IsComplete = &isComplete
	incomplete := db.Filter(opts)

	// Show top 5 items by priority
	limit := 5
	if len(incomplete) < limit {
		limit = len(incomplete)
	}

	blocks := []slackBlock{
		{
			Type: "header",
			Text: &slackText{Type: "plain_text", Text: "RTMX Backlog"},
		},
		{
			Type: "section",
			Text: &slackText{
				Type: "mrkdwn",
				Text: fmt.Sprintf("Showing top %d of %d incomplete requirements", limit, len(incomplete)),
			},
		},
	}

	for i := 0; i < limit; i++ {
		req := incomplete[i]
		text := fmt.Sprintf("*%s* [%s] %s\n_%s_ | %s | %.1fw",
			req.ReqID, req.Status.String(), truncateStr(req.RequirementText, 60),
			req.Priority.String(), req.Category, req.EffortWeeks)
		blocks = append(blocks, slackBlock{
			Type: "section",
			Text: &slackText{Type: "mrkdwn", Text: text},
		})
	}

	return slackResponse{
		ResponseType: "in_channel",
		Blocks:       blocks,
	}
}

// slackReqDetailResponse builds the Block Kit response for /rtmx req REQ-ID.
func slackReqDetailResponse(db *database.Database, reqID string) slackResponse {
	req := db.Get(reqID)
	if req == nil {
		return slackErrorResponse(fmt.Sprintf("Requirement not found: %s", reqID))
	}

	deps := req.Dependencies.Slice()
	blocks := req.Blocks.Slice()

	depsStr := "None"
	if len(deps) > 0 {
		depsStr = strings.Join(deps, ", ")
	}
	blocksStr := "None"
	if len(blocks) > 0 {
		blocksStr = strings.Join(blocks, ", ")
	}

	return slackResponse{
		ResponseType: "in_channel",
		Blocks: []slackBlock{
			{
				Type: "header",
				Text: &slackText{Type: "plain_text", Text: reqID},
			},
			{
				Type: "section",
				Text: &slackText{Type: "mrkdwn", Text: req.RequirementText},
			},
			{
				Type: "section",
				Fields: []slackText{
					{Type: "mrkdwn", Text: fmt.Sprintf("*Status:* %s", req.Status.String())},
					{Type: "mrkdwn", Text: fmt.Sprintf("*Priority:* %s", req.Priority.String())},
					{Type: "mrkdwn", Text: fmt.Sprintf("*Category:* %s", req.Category)},
					{Type: "mrkdwn", Text: fmt.Sprintf("*Phase:* %d", req.Phase)},
					{Type: "mrkdwn", Text: fmt.Sprintf("*Effort:* %.1f weeks", req.EffortWeeks)},
					{Type: "mrkdwn", Text: fmt.Sprintf("*Assignee:* %s", req.Assignee)},
				},
			},
			{
				Type: "section",
				Fields: []slackText{
					{Type: "mrkdwn", Text: fmt.Sprintf("*Dependencies:* %s", depsStr)},
					{Type: "mrkdwn", Text: fmt.Sprintf("*Blocks:* %s", blocksStr)},
				},
			},
		},
	}
}

// slackErrorResponse returns an ephemeral error message.
func slackErrorResponse(msg string) slackResponse {
	return slackResponse{
		ResponseType: "ephemeral",
		Blocks: []slackBlock{
			{
				Type: "section",
				Text: &slackText{Type: "mrkdwn", Text: msg},
			},
		},
	}
}

// parseFormBody is a minimal URL form parser for Slack payloads.
type formValues map[string]string

func (f formValues) Get(key string) string {
	return f[key]
}

func parseFormBody(body string) (formValues, error) {
	values := make(formValues)
	for _, pair := range strings.Split(body, "&") {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			key := urlDecode(parts[0])
			val := urlDecode(parts[1])
			values[key] = val
		}
	}
	return values, nil
}

// urlDecode performs basic URL percent-decoding.
func urlDecode(s string) string {
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '+' {
			result.WriteByte(' ')
		} else if s[i] == '%' && i+2 < len(s) {
			hi := unhex(s[i+1])
			lo := unhex(s[i+2])
			if hi >= 0 && lo >= 0 {
				result.WriteByte(byte(hi<<4 | lo))
				i += 2
			} else {
				result.WriteByte(s[i])
			}
		} else {
			result.WriteByte(s[i])
		}
	}
	return result.String()
}

func unhex(c byte) int {
	switch {
	case '0' <= c && c <= '9':
		return int(c - '0')
	case 'a' <= c && c <= 'f':
		return int(c - 'a' + 10)
	case 'A' <= c && c <= 'F':
		return int(c - 'A' + 10)
	}
	return -1
}
