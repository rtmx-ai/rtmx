package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/docmodel/migrate"
	"github.com/rtmx-ai/rtmx/internal/output"
	"github.com/spf13/cobra"
)

var decomposeDryRun bool

var decomposeCmd = &cobra.Command{
	Use:   "decompose REQ-ID",
	Short: "Split a coarse requirement into child requirements",
	Long: `Turn a coarse requirement into child CSV rows and Markdown files.

A requirement is already atomic when it has 1–5 acceptance criteria,
is not described as a parent that closes only when children complete,
and does not already list incomplete children in blocks.

Otherwise children are written as {parent}a, {parent}b, … and the parent
gains dependencies on those children. Idempotent: a second run adds nothing.

Does not create commits or pull requests.`,
	Args: cobra.ExactArgs(1),
	RunE: runDecompose,
}

func init() {
	decomposeCmd.Flags().BoolVar(&decomposeDryRun, "dry-run", false, "print the plan without writing")
	rootCmd.AddCommand(decomposeCmd)
}

var parentClosesOnly = regexp.MustCompile(`(?i)closes?\s+only\s+when\s+children|parent\s+.*\s+children\s+complete|when\s+children\s+are\s+COMPLETE`)

type decomposePlan struct {
	ParentID    string
	Atomic      bool
	Reason      string
	ChildIDs    []string
	ChildTitles []string
	ChildACs    [][]migrate.AC
}

func runDecompose(cmd *cobra.Command, args []string) error {
	if noColor {
		output.DisableColor()
	}
	reqID := args[0]
	cwd, err := os.Getwd()
	if err != nil {
		return err
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
	req := db.Get(reqID)
	if req == nil {
		return fmt.Errorf("requirement %s not found", reqID)
	}
	if req.Status == database.StatusComplete {
		return fmt.Errorf("refuse to decompose COMPLETE requirement %s", reqID)
	}

	mdPath := resolveRequirementFile(cwd, req)
	mdBody := ""
	if mdPath != "" {
		if b, err := os.ReadFile(mdPath); err == nil {
			mdBody = string(b)
		}
	}
	plan, err := planDecompose(req, mdBody, db)
	if err != nil {
		return err
	}
	if plan.Atomic {
		cmd.Printf("%s already atomic (%s)\n", reqID, plan.Reason)
		return nil
	}

	cmd.Printf("Decompose %s → %d children\n", reqID, len(plan.ChildIDs))
	for i, id := range plan.ChildIDs {
		cmd.Printf("  %s  %s\n", id, plan.ChildTitles[i])
	}
	if decomposeDryRun {
		cmd.Println("Dry run — no changes written")
		return nil
	}

	if err := applyDecompose(cwd, db, req, plan, mdPath); err != nil {
		return err
	}
	if err := db.Save(dbPath); err != nil {
		return fmt.Errorf("save database: %w", err)
	}
	cmd.Printf("%s Wrote %d children; parent depends on them\n", output.Color("✓", output.Green), len(plan.ChildIDs))
	return nil
}

func resolveRequirementFile(cwd string, req *database.Requirement) string {
	if req.RequirementFile != "" {
		p := req.RequirementFile
		if !filepath.IsAbs(p) {
			p = filepath.Join(cwd, p)
		}
		return p
	}
	return filepath.Join(cwd, ".rtmx", "requirements", req.Category, req.ReqID+".md")
}

func planDecompose(req *database.Requirement, mdBody string, db *database.Database) (*decomposePlan, error) {
	acs := migrate.ParseAcceptanceCriteria(mdBody)
	decompTitles := parseDecompositionSection(mdBody)

	incompleteChildren := 0
	for _, id := range req.Blocks.Slice() {
		child := db.Get(id)
		if child != nil && child.IsIncomplete() {
			incompleteChildren++
		}
	}

	parentish := parentClosesOnly.MatchString(req.RequirementText) || parentClosesOnly.MatchString(mdBody)
	atomic := len(acs) >= 1 && len(acs) <= 5 && !parentish && incompleteChildren == 0 && len(decompTitles) == 0
	if atomic {
		return &decomposePlan{ParentID: req.ReqID, Atomic: true, Reason: fmt.Sprintf("%d ACs, no decomposition section", len(acs))}, nil
	}

	// Already has children from a prior decompose — idempotent.
	existingKids := existingChildIDs(req.ReqID, db)
	if len(existingKids) > 0 && incompleteChildren > 0 {
		return &decomposePlan{ParentID: req.ReqID, Atomic: true, Reason: "children already present", ChildIDs: existingKids}, nil
	}

	titles := decompTitles
	var childACs [][]migrate.AC
	if len(titles) == 0 {
		if len(acs) <= 5 && !parentish {
			// Coarse only because zero ACs — invent a single child from requirement text.
			if len(acs) == 0 {
				titles = []string{truncateTitle(req.RequirementText, 80)}
				childACs = [][]migrate.AC{{{ID: "AC-1", Statement: req.RequirementText}}}
			} else {
				// >5 is the only other coarse path without section; handled below
				titles = nil
			}
		}
		if len(acs) > 5 {
			titles = make([]string, len(acs))
			childACs = make([][]migrate.AC, len(acs))
			for i, ac := range acs {
				titles[i] = truncateTitle(ac.Statement, 80)
				childACs[i] = []migrate.AC{{ID: "AC-1", Statement: ac.Statement}}
			}
		}
	}
	if len(titles) == 0 && parentish && len(acs) > 0 {
		// Parentish with ACs: one child per AC.
		titles = make([]string, len(acs))
		childACs = make([][]migrate.AC, len(acs))
		for i, ac := range acs {
			titles[i] = truncateTitle(ac.Statement, 80)
			childACs[i] = []migrate.AC{{ID: "AC-1", Statement: ac.Statement}}
		}
	}
	if len(titles) == 0 {
		return nil, fmt.Errorf("%s is coarse but has no ACs or ## Decomposition section to split", req.ReqID)
	}
	if childACs == nil {
		childACs = make([][]migrate.AC, len(titles))
		// Distribute ACs across children when decomposition section lists titles.
		if len(acs) == 0 {
			for i := range titles {
				childACs[i] = []migrate.AC{{ID: "AC-1", Statement: titles[i]}}
			}
		} else if len(acs) == len(titles) {
			for i := range titles {
				childACs[i] = []migrate.AC{{ID: "AC-1", Statement: acs[i].Statement}}
			}
		} else {
			for i := range titles {
				childACs[i] = []migrate.AC{{ID: "AC-1", Statement: titles[i]}}
			}
		}
	}

	ids := allocateChildIDs(req.ReqID, len(titles), db)
	return &decomposePlan{
		ParentID:    req.ReqID,
		ChildIDs:    ids,
		ChildTitles: titles,
		ChildACs:    childACs,
	}, nil
}

func existingChildIDs(parent string, db *database.Database) []string {
	var out []string
	for _, r := range db.All() {
		if strings.HasPrefix(r.ReqID, parent) && r.ReqID != parent && isChildSuffix(r.ReqID[len(parent):]) {
			out = append(out, r.ReqID)
		}
	}
	return out
}

func isChildSuffix(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < 'a' || r > 'z' {
			return false
		}
	}
	return true
}

func allocateChildIDs(parent string, n int, db *database.Database) []string {
	ids := make([]string, 0, n)
	letter := 0
	for len(ids) < n {
		suf := childSuffix(letter)
		letter++
		id := parent + suf
		if db.Get(id) != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids
}

func childSuffix(i int) string {
	// a, b, … z, aa, ab, …
	if i < 26 {
		return string(rune('a' + i))
	}
	return childSuffix(i/26-1) + string(rune('a'+i%26))
}

func parseDecompositionSection(md string) []string {
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
		if !in || trim == "" {
			continue
		}
		if m := regexp.MustCompile(`^\s*[-*]|\d+[.)]\s+`).FindString(line); m != "" || strings.HasPrefix(trim, "- ") {
			title := trim
			title = regexp.MustCompile(`^\d+[.)]\s+`).ReplaceAllString(title, "")
			title = strings.TrimPrefix(title, "- ")
			title = strings.TrimPrefix(title, "* ")
			title = regexp.MustCompile(`^\[[ xX]\]\s*`).ReplaceAllString(title, "")
			title = strings.TrimSpace(title)
			if title != "" {
				titles = append(titles, title)
			}
		}
	}
	return titles
}

func truncateTitle(s string, n int) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) <= n {
		return s
	}
	cut := s[:n]
	if i := strings.LastIndexFunc(cut, unicode.IsSpace); i > n/2 {
		cut = cut[:i]
	}
	return strings.TrimSpace(cut) + "…"
}

func applyDecompose(cwd string, db *database.Database, parent *database.Requirement, plan *decomposePlan, parentMD string) error {
	cat := parent.Category
	if cat == "" {
		cat = "GENERAL"
	}
	reqDir := filepath.Join(cwd, ".rtmx", "requirements", cat)
	if err := os.MkdirAll(reqDir, 0o755); err != nil {
		return err
	}
	for i, id := range plan.ChildIDs {
		acs := plan.ChildACs[i]
		body := renderChildMarkdown(id, plan.ChildTitles[i], acs)
		path := filepath.Join(reqDir, id+".md")
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			return err
		}
		rel := filepath.ToSlash(filepath.Join(".rtmx", "requirements", cat, id+".md"))
		child := &database.Requirement{
			ReqID:           id,
			Category:        cat,
			Subcategory:     parent.Subcategory,
			RequirementText: plan.ChildTitles[i],
			Status:          database.StatusMissing,
			Priority:        parent.Priority,
			Phase:           parent.Phase,
			Dependencies:    database.NewStringSet(),
			Blocks:          database.NewStringSet(),
			RequirementFile: rel,
			ValidationMethod: "System Test",
		}
		parent.Blocks.Add(id)
		parent.Dependencies.Add(id)
		if err := db.Add(child); err != nil {
			return err
		}
	}
	return nil
}

func renderChildMarkdown(id, title string, acs []migrate.AC) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s: %s\n\n", id, title)
	b.WriteString("## Acceptance Criteria\n\n")
	for i, ac := range acs {
		stmt := ac.Statement
		if stmt == "" {
			stmt = title
		}
		fmt.Fprintf(&b, "%d. %s\n", i+1, stmt)
	}
	b.WriteString("\n")
	return b.String()
}
