package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/docmodel/migrate"
	"github.com/rtmx-ai/rtmx/internal/graph"
)

const (
	deliveryPRPolicy     = "one_pr_per_requirement"
	deliveryCommitPolicy = "one_commit_per_ac"
)

// deliverySkip is a requirement the planner did not select.
type deliverySkip struct {
	ReqID  string `json:"req_id"`
	Reason string `json:"reason"`
}

// deliveryPlan matches one rtmx loop tick. MCP next does not call GitHub
// and does not claim; skipped stays empty and idle writes no claim.
type deliveryPlan struct {
	Idle         bool           `json:"idle,omitempty"`
	ReqID        string         `json:"req_id,omitempty"`
	Decomposed   bool           `json:"decomposed"`
	Children     []string       `json:"children"`
	PRPolicy     string         `json:"pr_policy"`
	CommitPolicy string         `json:"commit_policy"`
	Skipped      []deliverySkip `json:"skipped"`
}

func (p deliveryPlan) MarshalJSON() ([]byte, error) {
	if p.Idle {
		return []byte(`{"idle":true}`), nil
	}
	children := p.Children
	if children == nil {
		children = []string{}
	}
	skipped := p.Skipped
	if skipped == nil {
		skipped = []deliverySkip{}
	}
	return json.Marshal(struct {
		ReqID        string         `json:"req_id"`
		Decomposed   bool           `json:"decomposed"`
		Children     []string       `json:"children"`
		PRPolicy     string         `json:"pr_policy"`
		CommitPolicy string         `json:"commit_policy"`
		Skipped      []deliverySkip `json:"skipped"`
	}{
		ReqID:        p.ReqID,
		Decomposed:   p.Decomposed,
		Children:     children,
		PRPolicy:     p.PRPolicy,
		CommitPolicy: p.CommitPolicy,
		Skipped:      skipped,
	})
}

var deliveryParentish = regexp.MustCompile(`(?i)closes?\s+only\s+when\s+children|parent\s+.*\s+children\s+complete|when\s+children\s+are\s+COMPLETE`)

func (s *Server) planDelivery(db *database.Database, webs []graph.Web, category string) deliveryPlan {
	req := pickDeliveryReq(db, webs, category)
	if req == nil {
		return deliveryPlan{Idle: true}
	}
	body := s.readRequirementMarkdown(req)
	children, decomposed := deliveryChildren(db, req, body)
	return deliveryPlan{
		ReqID:        req.ReqID,
		Decomposed:   decomposed,
		Children:     children,
		PRPolicy:     deliveryPRPolicy,
		CommitPolicy: deliveryCommitPolicy,
		Skipped:      []deliverySkip{},
	}
}

func pickDeliveryReq(db *database.Database, webs []graph.Web, category string) *database.Requirement {
	var best *database.Requirement
	for _, web := range webs {
		if category != "" && !webHasCategory(db, web, category) {
			continue
		}
		for _, id := range web.Unblocked {
			r := db.Get(id)
			if r == nil {
				continue
			}
			if category != "" && r.Category != category {
				continue
			}
			if best == nil || r.Priority.Weight() < best.Priority.Weight() ||
				(r.Priority.Weight() == best.Priority.Weight() && r.EffortWeeks < best.EffortWeeks) {
				best = r
			}
		}
	}
	return best
}

func webHasCategory(db *database.Database, web graph.Web, category string) bool {
	for _, id := range web.IDs {
		r := db.Get(id)
		if r != nil && r.Category == category {
			return true
		}
	}
	return false
}

func (s *Server) readRequirementMarkdown(req *database.Requirement) string {
	base := filepath.Dir(filepath.Dir(s.dbPath))
	var path string
	switch {
	case req.RequirementFile != "":
		path = req.RequirementFile
		if !filepath.IsAbs(path) {
			path = filepath.Join(base, path)
		}
	case s.cfg != nil:
		path = filepath.Join(s.cfg.RequirementsPath(base), req.Category, req.ReqID+".md")
	default:
		path = filepath.Join(base, ".rtmx", "requirements", req.Category, req.ReqID+".md")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(raw)
}

func deliveryChildren(db *database.Database, req *database.Requirement, body string) ([]string, bool) {
	acs := migrate.ParseAcceptanceCriteria(body)
	titles := deliveryDecompositionTitles(body)
	incompleteChildren := 0
	if req.Blocks != nil {
		for _, id := range req.Blocks.Slice() {
			child := db.Get(id)
			if child != nil && child.IsIncomplete() {
				incompleteChildren++
			}
		}
	}
	parentish := deliveryParentish.MatchString(req.RequirementText) || deliveryParentish.MatchString(body)
	atomic := len(acs) >= 1 && len(acs) <= 5 && !parentish && incompleteChildren == 0 && len(titles) == 0
	if atomic {
		return []string{}, false
	}
	n := len(titles)
	if n == 0 && len(acs) > 5 {
		n = len(acs)
	}
	if n == 0 && parentish && len(acs) > 0 {
		n = len(acs)
	}
	if n == 0 {
		n = 1
	}
	return deliveryChildIDs(req.ReqID, n, db), true
}

func deliveryChildIDs(parent string, n int, db *database.Database) []string {
	ids := make([]string, 0, n)
	letter := 0
	for len(ids) < n {
		suf := deliverySuffix(letter)
		letter++
		id := parent + suf
		if db.Get(id) != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids
}

func deliverySuffix(i int) string {
	if i < 26 {
		return string(rune('a' + i))
	}
	return deliverySuffix(i/26-1) + string(rune('a'+i%26))
}

func deliveryDecompositionTitles(md string) []string {
	lines := strings.Split(md, "\n")
	in := false
	var titles []string
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "## ") {
			if in {
				break
			}
			if strings.EqualFold(trim, "## Decomposition") {
				in = true
			}
			continue
		}
		if !in || !strings.HasPrefix(trim, "- ") {
			continue
		}
		title := strings.TrimSpace(strings.TrimPrefix(trim, "- "))
		if title != "" {
			titles = append(titles, title)
		}
	}
	return titles
}
