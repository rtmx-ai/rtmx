package cmd

import (
	"encoding/json"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/graph"
	"github.com/rtmx-ai/rtmx/internal/orchestration"
)

// apiRequirement is the JSON representation of a requirement for the REST API.
type apiRequirement struct {
	ReqID           string   `json:"req_id"`
	Category        string   `json:"category"`
	Subcategory     string   `json:"subcategory"`
	RequirementText string   `json:"requirement_text"`
	Status          string   `json:"status"`
	Priority        string   `json:"priority"`
	Phase           int      `json:"phase"`
	EffortWeeks     float64  `json:"effort_weeks"`
	Assignee        string   `json:"assignee"`
	Sprint          string   `json:"sprint"`
	Dependencies    []string `json:"dependencies"`
	Blocks          []string `json:"blocks"`
	StartedDate     string   `json:"started_date"`
	CompletedDate   string   `json:"completed_date"`
}

// apiPagination describes the pagination state in the response.
type apiPagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// apiFiltersApplied records which filters were active.
type apiFiltersApplied struct {
	Category *string `json:"category"`
	Status   *string `json:"status"`
	Priority *string `json:"priority"`
	Version  *string `json:"version"`
	Assignee *string `json:"assignee"`
	Search   *string `json:"search"`
}

// apiRequirementsResponse is the top-level response for GET /api/requirements.
type apiRequirementsResponse struct {
	Requirements []apiRequirement  `json:"requirements"`
	Pagination   apiPagination     `json:"pagination"`
	Filters      apiFiltersApplied `json:"filters_applied"`
}

// toAPIRequirement converts a database Requirement to the API representation.
func toAPIRequirement(r *database.Requirement) apiRequirement {
	deps := r.Dependencies.Slice()
	if deps == nil {
		deps = []string{}
	}
	blocks := r.Blocks.Slice()
	if blocks == nil {
		blocks = []string{}
	}
	sort.Strings(deps)
	sort.Strings(blocks)

	return apiRequirement{
		ReqID:           r.ReqID,
		Category:        r.Category,
		Subcategory:     r.Subcategory,
		RequirementText: r.RequirementText,
		Status:          r.Status.String(),
		Priority:        r.Priority.String(),
		Phase:           r.Phase,
		EffortWeeks:     r.EffortWeeks,
		Assignee:        r.Assignee,
		Sprint:          r.Sprint,
		Dependencies:    deps,
		Blocks:          blocks,
		StartedDate:     r.StartedDate,
		CompletedDate:   r.CompletedDate,
	}
}

var validSortFields = map[string]bool{
	"req_id":       true,
	"category":     true,
	"priority":     true,
	"status":       true,
	"effort_weeks": true,
	"phase":        true,
}

// handleAPIRequirements handles GET /api/requirements.
func handleAPIRequirements(db *database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeAPIError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		q := r.URL.Query()

		// Parse filter options
		opts := database.FilterOptions{}
		var filters apiFiltersApplied

		if v := q.Get("category"); v != "" {
			opts.Category = v
			filters.Category = &v
		}
		if v := q.Get("status"); v != "" {
			status, err := database.ParseStatus(v)
			if err != nil {
				writeAPIError(w, http.StatusBadRequest, "invalid status: "+v)
				return
			}
			opts.Status = &status
			s := status.String()
			filters.Status = &s
		}
		if v := q.Get("priority"); v != "" {
			priority, err := database.ParsePriority(v)
			if err != nil {
				writeAPIError(w, http.StatusBadRequest, "invalid priority: "+v)
				return
			}
			opts.Priority = &priority
			p := priority.String()
			filters.Priority = &p
		}
		if v := q.Get("version"); v != "" {
			opts.TargetVersion = v
			filters.Version = &v
		}
		if v := q.Get("assignee"); v != "" {
			opts.Assignee = v
			filters.Assignee = &v
		}

		search := q.Get("search")
		if search != "" {
			filters.Search = &search
		}

		// Sort parameters
		sortField := q.Get("sort")
		if sortField == "" {
			sortField = "req_id"
		}
		if !validSortFields[sortField] {
			writeAPIError(w, http.StatusBadRequest, "invalid sort field: "+sortField)
			return
		}

		order := strings.ToLower(q.Get("order"))
		if order == "" {
			order = "asc"
		}
		if order != "asc" && order != "desc" {
			writeAPIError(w, http.StatusBadRequest, "invalid order: must be asc or desc")
			return
		}

		// Pagination parameters
		page := 1
		if v := q.Get("page"); v != "" {
			p, err := strconv.Atoi(v)
			if err != nil || p < 1 {
				writeAPIError(w, http.StatusBadRequest, "invalid page: must be >= 1")
				return
			}
			page = p
		}

		perPage := 50
		if v := q.Get("per_page"); v != "" {
			pp, err := strconv.Atoi(v)
			if err != nil || pp < 1 {
				writeAPIError(w, http.StatusBadRequest, "invalid per_page: must be >= 1")
				return
			}
			if pp > 200 {
				pp = 200
			}
			perPage = pp
		}

		// Filter
		reqs := db.Filter(opts)

		// Apply full-text search
		if search != "" {
			searchLower := strings.ToLower(search)
			var matched []*database.Requirement
			for _, req := range reqs {
				if strings.Contains(strings.ToLower(req.ReqID), searchLower) ||
					strings.Contains(strings.ToLower(req.RequirementText), searchLower) ||
					strings.Contains(strings.ToLower(req.Notes), searchLower) {
					matched = append(matched, req)
				}
			}
			reqs = matched
		}

		// Sort
		sortRequirements(reqs, sortField, order == "desc")

		// Paginate
		total := len(reqs)
		totalPages := int(math.Ceil(float64(total) / float64(perPage)))
		if totalPages < 1 {
			totalPages = 1
		}

		start := (page - 1) * perPage
		if start >= total {
			reqs = nil
		} else {
			end := start + perPage
			if end > total {
				end = total
			}
			reqs = reqs[start:end]
		}

		// Build response
		apiReqs := make([]apiRequirement, 0, len(reqs))
		for _, req := range reqs {
			apiReqs = append(apiReqs, toAPIRequirement(req))
		}

		resp := apiRequirementsResponse{
			Requirements: apiReqs,
			Pagination: apiPagination{
				Page:       page,
				PerPage:    perPage,
				Total:      total,
				TotalPages: totalPages,
			},
			Filters: filters,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}

// sortRequirements sorts a slice of requirements by the given field.
func sortRequirements(reqs []*database.Requirement, field string, desc bool) {
	sort.SliceStable(reqs, func(i, j int) bool {
		var less bool
		switch field {
		case "req_id":
			less = reqs[i].ReqID < reqs[j].ReqID
		case "category":
			less = reqs[i].Category < reqs[j].Category
		case "priority":
			less = reqs[i].Priority.Weight() < reqs[j].Priority.Weight()
		case "status":
			less = reqs[i].Status.Weight() < reqs[j].Status.Weight()
		case "effort_weeks":
			less = reqs[i].EffortWeeks < reqs[j].EffortWeeks
		case "phase":
			less = reqs[i].Phase < reqs[j].Phase
		default:
			less = reqs[i].ReqID < reqs[j].ReqID
		}
		if desc {
			return !less
		}
		return less
	})
}

// writeAPIError writes a JSON error response.
func writeAPIError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// writeJSON writes a JSON response with 200 OK.
func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// --- REQ-API-002: Requirement Detail Endpoint ---

// apiRequirementDetail is the full requirement with extra fields for detail view.
type apiRequirementDetail struct {
	apiRequirement
	Notes            string `json:"notes"`
	TargetValue      string `json:"target_value"`
	TestModule       string `json:"test_module"`
	TestFunction     string `json:"test_function"`
	ValidationMethod string `json:"validation_method"`
	RequirementFile  string `json:"requirement_file"`
	ExternalID       string `json:"external_id"`
}

// apiDepSummary is a lightweight requirement reference for dependency lists.
type apiDepSummary struct {
	ReqID           string `json:"req_id"`
	Status          string `json:"status"`
	RequirementText string `json:"requirement_text"`
}

// apiDependencyDetail provides dependency context for a requirement.
type apiDependencyDetail struct {
	Upstream                  []apiDepSummary `json:"upstream"`
	Downstream                []apiDepSummary `json:"downstream"`
	TransitiveUpstreamCount   int             `json:"transitive_upstream_count"`
	TransitiveDownstreamCount int             `json:"transitive_downstream_count"`
	AllUpstreamComplete       bool            `json:"all_upstream_complete"`
}

// apiRequirementDetailResponse is the response for GET /api/requirements/:id.
type apiRequirementDetailResponse struct {
	Requirement      apiRequirementDetail `json:"requirement"`
	DependencyDetail apiDependencyDetail  `json:"dependency_detail"`
}

func toAPIRequirementDetail(r *database.Requirement) apiRequirementDetail {
	return apiRequirementDetail{
		apiRequirement:   toAPIRequirement(r),
		Notes:            r.Notes,
		TargetValue:      r.TargetValue,
		TestModule:       r.TestModule,
		TestFunction:     r.TestFunction,
		ValidationMethod: r.ValidationMethod,
		RequirementFile:  r.RequirementFile,
		ExternalID:       r.ExternalID,
	}
}

func buildDependencyDetail(db *database.Database, req *database.Requirement) apiDependencyDetail {
	g := graph.NewGraph(db)

	var upstream []apiDepSummary
	for dep := range req.Dependencies {
		if r := db.Get(dep); r != nil {
			upstream = append(upstream, apiDepSummary{
				ReqID: r.ReqID, Status: r.Status.String(), RequirementText: r.RequirementText,
			})
		}
	}
	if upstream == nil {
		upstream = []apiDepSummary{}
	}
	sort.Slice(upstream, func(i, j int) bool { return upstream[i].ReqID < upstream[j].ReqID })

	var downstream []apiDepSummary
	for blk := range req.Blocks {
		if r := db.Get(blk); r != nil {
			downstream = append(downstream, apiDepSummary{
				ReqID: r.ReqID, Status: r.Status.String(), RequirementText: r.RequirementText,
			})
		}
	}
	if downstream == nil {
		downstream = []apiDepSummary{}
	}
	sort.Slice(downstream, func(i, j int) bool { return downstream[i].ReqID < downstream[j].ReqID })

	transUp := g.TransitiveDependencies(req.ReqID)
	transDown := g.TransitiveDependents(req.ReqID)

	allComplete := true
	for _, id := range transUp {
		if r := db.Get(id); r != nil && r.IsIncomplete() {
			allComplete = false
			break
		}
	}
	if len(transUp) == 0 {
		allComplete = true
	}

	return apiDependencyDetail{
		Upstream:                  upstream,
		Downstream:                downstream,
		TransitiveUpstreamCount:   len(transUp),
		TransitiveDownstreamCount: len(transDown),
		AllUpstreamComplete:       allComplete,
	}
}

// handleAPIRequirementDetail handles GET /api/requirements/:id.
func handleAPIRequirementDetail(db *database.Database, dbPath string, mu *sync.Mutex) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqID := strings.TrimPrefix(r.URL.Path, "/api/requirements/")
		if reqID == "" {
			writeAPIError(w, http.StatusBadRequest, "missing requirement ID")
			return
		}

		switch r.Method {
		case http.MethodGet:
			req := db.Get(reqID)
			if req == nil {
				writeAPIError(w, http.StatusNotFound, "requirement not found: "+reqID)
				return
			}
			resp := apiRequirementDetailResponse{
				Requirement:      toAPIRequirementDetail(req),
				DependencyDetail: buildDependencyDetail(db, req),
			}
			writeJSON(w, resp)

		case http.MethodPatch:
			handleRequirementPatch(w, r, db, dbPath, reqID, mu)

		default:
			writeAPIError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	}
}

// --- REQ-API-003: Requirement Update Endpoint ---

var mutableFields = map[string]bool{
	"status": true, "assignee": true, "sprint": true, "priority": true, "notes": true,
}

func handleRequirementPatch(w http.ResponseWriter, r *http.Request, db *database.Database, dbPath string, reqID string, mu *sync.Mutex) {
	req := db.Get(reqID)
	if req == nil {
		writeAPIError(w, http.StatusNotFound, "requirement not found: "+reqID)
		return
	}

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	// Validate all fields are mutable
	for key := range body {
		if !mutableFields[key] {
			writeAPIError(w, http.StatusBadRequest, "field not mutable: "+key)
			return
		}
	}

	// Validate status if provided
	if v, ok := body["status"]; ok {
		s, isStr := v.(string)
		if !isStr {
			writeAPIError(w, http.StatusBadRequest, "status must be a string")
			return
		}
		if _, err := database.ParseStatus(s); err != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid status: "+s)
			return
		}
	}

	// Validate priority if provided
	if v, ok := body["priority"]; ok {
		s, isStr := v.(string)
		if !isStr {
			writeAPIError(w, http.StatusBadRequest, "priority must be a string")
			return
		}
		if _, err := database.ParsePriority(s); err != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid priority: "+s)
			return
		}
	}

	mu.Lock()
	defer mu.Unlock()

	// Handle completed_date auto-management
	if statusVal, ok := body["status"]; ok {
		newStatus, _ := database.ParseStatus(statusVal.(string))
		oldStatus := req.Status
		if newStatus == database.StatusComplete && oldStatus != database.StatusComplete {
			body["completed_date"] = time.Now().UTC().Format("2006-01-02")
		} else if newStatus != database.StatusComplete && oldStatus == database.StatusComplete {
			body["completed_date"] = ""
		}
	}

	if err := db.Update(reqID, body); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "update failed: "+err.Error())
		return
	}

	if err := db.Save(dbPath); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "save failed: "+err.Error())
		return
	}

	// Return updated requirement detail
	updated := db.Get(reqID)
	resp := apiRequirementDetailResponse{
		Requirement:      toAPIRequirementDetail(updated),
		DependencyDetail: buildDependencyDetail(db, updated),
	}
	writeJSON(w, resp)
}

// --- REQ-API-004: Dependency Graph Endpoint ---

type apiGraphNode struct {
	ID          string  `json:"id"`
	Category    string  `json:"category"`
	Status      string  `json:"status"`
	Priority    string  `json:"priority"`
	EffortWeeks float64 `json:"effort_weeks"`
	Label       string  `json:"label"`
	Blocked     bool    `json:"blocked"`
	Depth       int     `json:"depth"`
}

type apiGraphEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`
}

type apiGraphMetadata struct {
	TotalNodes         int      `json:"total_nodes"`
	TotalEdges         int      `json:"total_edges"`
	CriticalPath       []string `json:"critical_path"`
	CriticalPathLength int      `json:"critical_path_length"`
	MaxDepth           int      `json:"max_depth"`
	IndependentWebs    int      `json:"independent_webs"`
}

type apiGraphResponse struct {
	Nodes    []apiGraphNode   `json:"nodes"`
	Edges    []apiGraphEdge   `json:"edges"`
	Metadata apiGraphMetadata `json:"metadata"`
}

func handleAPIGraph(db *database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeAPIError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		q := r.URL.Query()
		categoryFilter := q.Get("category")
		statusFilter := q.Get("status")
		rootFilter := q.Get("root")
		depthLimit := 0
		if v := q.Get("depth"); v != "" {
			if d, err := strconv.Atoi(v); err == nil && d > 0 {
				depthLimit = d
			}
		}

		g := graph.NewGraph(db)
		layers := g.Layers()

		// Build depth map
		depthMap := make(map[string]int)
		maxDepth := 0
		for d, layer := range layers {
			for _, id := range layer {
				depthMap[id] = d
				if d > maxDepth {
					maxDepth = d
				}
			}
		}

		// Determine which nodes to include
		includeSet := make(map[string]bool)
		if rootFilter != "" {
			if db.Get(rootFilter) == nil {
				writeAPIError(w, http.StatusNotFound, "root requirement not found: "+rootFilter)
				return
			}
			// Include root + transitive deps + transitive dependents
			includeSet[rootFilter] = true
			for _, id := range g.TransitiveDependencies(rootFilter) {
				includeSet[id] = true
			}
			for _, id := range g.TransitiveDependents(rootFilter) {
				includeSet[id] = true
			}
			// Apply depth limit from root
			if depthLimit > 0 {
				rootDepth := depthMap[rootFilter]
				for id := range includeSet {
					d := depthMap[id]
					dist := d - rootDepth
					if dist < 0 {
						dist = -dist
					}
					if dist > depthLimit {
						delete(includeSet, id)
					}
				}
			}
		} else {
			for _, req := range db.All() {
				includeSet[req.ReqID] = true
			}
		}

		var nodes []apiGraphNode
		var edges []apiGraphEdge

		for _, req := range db.All() {
			if !includeSet[req.ReqID] {
				continue
			}
			if categoryFilter != "" && req.Category != categoryFilter {
				continue
			}
			if statusFilter != "" && req.Status.String() != strings.ToUpper(statusFilter) {
				continue
			}

			label := req.RequirementText
			if len(label) > 60 {
				label = label[:57] + "..."
			}

			nodes = append(nodes, apiGraphNode{
				ID:          req.ReqID,
				Category:    req.Category,
				Status:      req.Status.String(),
				Priority:    req.Priority.String(),
				EffortWeeks: req.EffortWeeks,
				Label:       label,
				Blocked:     g.IsBlocked(req.ReqID),
				Depth:       depthMap[req.ReqID],
			})
		}
		if nodes == nil {
			nodes = []apiGraphNode{}
		}

		// Build node set for edge filtering
		nodeSet := make(map[string]bool)
		for _, n := range nodes {
			nodeSet[n.ID] = true
		}

		for _, req := range db.All() {
			if !nodeSet[req.ReqID] {
				continue
			}
			for blk := range req.Blocks {
				if nodeSet[blk] {
					edges = append(edges, apiGraphEdge{
						From: req.ReqID, To: blk, Type: "blocks",
					})
				}
			}
		}
		if edges == nil {
			edges = []apiGraphEdge{}
		}

		cp := g.CriticalPath()
		if cp == nil {
			cp = []string{}
		}
		webs := g.DetectWebs()

		resp := apiGraphResponse{
			Nodes: nodes,
			Edges: edges,
			Metadata: apiGraphMetadata{
				TotalNodes:         len(nodes),
				TotalEdges:         len(edges),
				CriticalPath:       cp,
				CriticalPathLength: len(cp),
				MaxDepth:           maxDepth,
				IndependentWebs:    len(webs),
			},
		}
		writeJSON(w, resp)
	}
}

// --- REQ-API-005: Backlog Endpoint ---

type apiBacklogItem struct {
	ReqID                string  `json:"req_id"`
	RequirementText      string  `json:"requirement_text"`
	Priority             string  `json:"priority"`
	Status               string  `json:"status"`
	EffortWeeks          float64 `json:"effort_weeks"`
	Blocked              bool    `json:"blocked"`
	BlocksCount          int     `json:"blocks_count"`
	TransitiveBlockCount int     `json:"transitive_blocks_count"`
	Category             string  `json:"category"`
	Assignee             string  `json:"assignee"`
	Sprint               string  `json:"sprint"`
}

type apiBacklogSection struct {
	Name  string           `json:"name"`
	Items []apiBacklogItem `json:"items"`
}

type apiBacklogSummary struct {
	TotalIncomplete    int     `json:"total_incomplete"`
	TotalEffortWeeks   float64 `json:"total_effort_weeks"`
	UnblockedCount     int     `json:"unblocked_count"`
	BlockedCount       int     `json:"blocked_count"`
}

type apiBacklogResponse struct {
	View     string              `json:"view"`
	Sections []apiBacklogSection `json:"sections"`
	Summary  apiBacklogSummary   `json:"summary"`
}

func handleAPIBacklog(db *database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeAPIError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		q := r.URL.Query()
		view := q.Get("view")
		if view == "" {
			view = "all"
		}
		categoryFilter := q.Get("category")
		versionFilter := q.Get("version")
		limit := 50
		if v := q.Get("limit"); v != "" {
			if l, err := strconv.Atoi(v); err == nil && l > 0 {
				limit = l
			}
		}

		g := graph.NewGraph(db)
		blocking := g.BlockingAnalysis()

		// Get incomplete requirements
		opts := database.FilterOptions{}
		isComplete := false
		opts.IsComplete = &isComplete
		if categoryFilter != "" {
			opts.Category = categoryFilter
		}
		if versionFilter != "" {
			opts.TargetVersion = versionFilter
		}
		incomplete := db.Filter(opts)

		// Convert to backlog items
		toItem := func(req *database.Requirement) apiBacklogItem {
			transBlocks := 0
			if c, ok := blocking[req.ReqID]; ok {
				transBlocks = c
			}
			return apiBacklogItem{
				ReqID:                req.ReqID,
				RequirementText:      req.RequirementText,
				Priority:             req.Priority.String(),
				Status:               req.Status.String(),
				EffortWeeks:          req.EffortWeeks,
				Blocked:              g.IsBlocked(req.ReqID),
				BlocksCount:          len(g.Dependents(req.ReqID)),
				TransitiveBlockCount: transBlocks,
				Category:             req.Category,
				Assignee:             req.Assignee,
				Sprint:               req.Sprint,
			}
		}

		var sections []apiBacklogSection

		switch view {
		case "critical":
			var items []apiBacklogItem
			for _, req := range incomplete {
				if (req.Priority == database.PriorityP0 || req.Priority == database.PriorityHigh) && blocking[req.ReqID] >= 2 {
					items = append(items, toItem(req))
				}
			}
			sort.Slice(items, func(i, j int) bool { return items[i].TransitiveBlockCount > items[j].TransitiveBlockCount })
			sections = []apiBacklogSection{{Name: "Critical Path", Items: limitItems(items, limit)}}

		case "quick-wins":
			var items []apiBacklogItem
			for _, req := range incomplete {
				if req.EffortWeeks <= 1.0 && req.Priority.IsHighPriority() && !g.IsBlocked(req.ReqID) {
					items = append(items, toItem(req))
				}
			}
			sort.Slice(items, func(i, j int) bool { return items[i].EffortWeeks < items[j].EffortWeeks })
			sections = []apiBacklogSection{{Name: "Quick Wins", Items: limitItems(items, limit)}}

		case "blockers":
			var items []apiBacklogItem
			for _, req := range incomplete {
				if blocking[req.ReqID] > 0 {
					items = append(items, toItem(req))
				}
			}
			sort.Slice(items, func(i, j int) bool { return items[i].TransitiveBlockCount > items[j].TransitiveBlockCount })
			sections = []apiBacklogSection{{Name: "Blockers", Items: limitItems(items, limit)}}

		default: // "all"
			// Critical path items
			var critItems []apiBacklogItem
			critSet := make(map[string]bool)
			for _, req := range incomplete {
				if blocking[req.ReqID] >= 2 {
					critItems = append(critItems, toItem(req))
					critSet[req.ReqID] = true
				}
			}
			sort.Slice(critItems, func(i, j int) bool { return critItems[i].TransitiveBlockCount > critItems[j].TransitiveBlockCount })

			// Quick wins
			var qwItems []apiBacklogItem
			qwSet := make(map[string]bool)
			for _, req := range incomplete {
				if !critSet[req.ReqID] && req.EffortWeeks <= 1.0 && !g.IsBlocked(req.ReqID) {
					qwItems = append(qwItems, toItem(req))
					qwSet[req.ReqID] = true
				}
			}
			sort.Slice(qwItems, func(i, j int) bool { return qwItems[i].EffortWeeks < qwItems[j].EffortWeeks })

			// Remaining
			var remaining []apiBacklogItem
			for _, req := range incomplete {
				if !critSet[req.ReqID] && !qwSet[req.ReqID] {
					remaining = append(remaining, toItem(req))
				}
			}

			sections = []apiBacklogSection{
				{Name: "Critical Path", Items: safeItems(critItems)},
				{Name: "Quick Wins", Items: safeItems(qwItems)},
				{Name: "Remaining", Items: safeItems(remaining)},
			}
			// Apply limit across all sections
			total := 0
			for i := range sections {
				remaining := limit - total
				if remaining <= 0 {
					sections[i].Items = []apiBacklogItem{}
				} else if len(sections[i].Items) > remaining {
					sections[i].Items = sections[i].Items[:remaining]
				}
				total += len(sections[i].Items)
			}
		}

		// Compute summary
		totalEffort := 0.0
		unblockedCount := 0
		blockedCount := 0
		for _, req := range incomplete {
			totalEffort += req.EffortWeeks
			if g.IsBlocked(req.ReqID) {
				blockedCount++
			} else {
				unblockedCount++
			}
		}

		resp := apiBacklogResponse{
			View:     view,
			Sections: sections,
			Summary: apiBacklogSummary{
				TotalIncomplete:  len(incomplete),
				TotalEffortWeeks: totalEffort,
				UnblockedCount:   unblockedCount,
				BlockedCount:     blockedCount,
			},
		}
		writeJSON(w, resp)
	}
}

func limitItems(items []apiBacklogItem, limit int) []apiBacklogItem {
	if items == nil {
		return []apiBacklogItem{}
	}
	if len(items) > limit {
		return items[:limit]
	}
	return items
}

func safeItems(items []apiBacklogItem) []apiBacklogItem {
	if items == nil {
		return []apiBacklogItem{}
	}
	return items
}

// --- REQ-API-006: Releases Endpoint ---

type apiVersionSummary struct {
	Version       string  `json:"version"`
	Label         string  `json:"label,omitempty"`
	Total         int     `json:"total"`
	Complete      int     `json:"complete"`
	Partial       int     `json:"partial"`
	Missing       int     `json:"missing"`
	CompletionPct float64 `json:"completion_pct"`
	GateStatus    string  `json:"gate_status,omitempty"`
}

type apiReleasesResponse struct {
	Versions []apiVersionSummary `json:"versions"`
}

type apiReleaseDetailResponse struct {
	Version          string             `json:"version"`
	GateStatus       string             `json:"gate_status"`
	GateFailures     []string           `json:"gate_failures"`
	Requirements     []apiRequirement   `json:"requirements"`
	Summary          apiReleaseSummary  `json:"summary"`
}

type apiReleaseSummary struct {
	Total                int     `json:"total"`
	Complete             int     `json:"complete"`
	CompletionPct        float64 `json:"completion_pct"`
	TotalEffortRemaining float64 `json:"total_effort_remaining"`
}

func handleAPIReleases(db *database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeAPIError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		byVersion := db.ByVersion()
		versions := db.Versions()

		var summaries []apiVersionSummary
		for _, v := range versions {
			reqs := byVersion[v]
			s := computeVersionSummary(v, reqs)
			summaries = append(summaries, s)
		}
		// Add unversioned group
		if reqs, ok := byVersion[""]; ok && len(reqs) > 0 {
			s := computeVersionSummary("", reqs)
			s.Label = "unversioned"
			summaries = append(summaries, s)
		}

		if summaries == nil {
			summaries = []apiVersionSummary{}
		}

		writeJSON(w, apiReleasesResponse{Versions: summaries})
	}
}

func handleAPIReleaseDetail(db *database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeAPIError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		version := strings.TrimPrefix(r.URL.Path, "/api/releases/")
		if version == "" {
			writeAPIError(w, http.StatusBadRequest, "missing version")
			return
		}

		byVersion := db.ByVersion()
		reqs, ok := byVersion[version]
		if !ok || len(reqs) == 0 {
			writeAPIError(w, http.StatusNotFound, "version not found: "+version)
			return
		}

		// Gate check
		var failures []string
		var incompleteIDs []string
		complete := 0
		effortRemaining := 0.0
		for _, req := range reqs {
			if req.IsComplete() {
				complete++
			} else {
				incompleteIDs = append(incompleteIDs, req.ReqID)
				effortRemaining += req.EffortWeeks
			}
		}

		gateStatus := "PASS"
		if len(incompleteIDs) > 0 {
			gateStatus = "FAIL"
			failures = append(failures, strconv.Itoa(len(incompleteIDs))+" requirements not COMPLETE: "+strings.Join(incompleteIDs, ", "))
		}
		if failures == nil {
			failures = []string{}
		}

		apiReqs := make([]apiRequirement, 0, len(reqs))
		for _, req := range reqs {
			apiReqs = append(apiReqs, toAPIRequirement(req))
		}

		pct := 0.0
		if len(reqs) > 0 {
			pct = math.Round(float64(complete) / float64(len(reqs)) * 100.0)
		}

		resp := apiReleaseDetailResponse{
			Version:      version,
			GateStatus:   gateStatus,
			GateFailures: failures,
			Requirements: apiReqs,
			Summary: apiReleaseSummary{
				Total:                len(reqs),
				Complete:             complete,
				CompletionPct:        pct,
				TotalEffortRemaining: effortRemaining,
			},
		}
		writeJSON(w, resp)
	}
}

func computeVersionSummary(version string, reqs []*database.Requirement) apiVersionSummary {
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

	pct := 0.0
	if len(reqs) > 0 {
		pct = math.Round(float64(complete) / float64(len(reqs)) * 100.0)
	}

	s := apiVersionSummary{
		Version:       version,
		Total:         len(reqs),
		Complete:      complete,
		Partial:       partial,
		Missing:       missing,
		CompletionPct: pct,
	}
	if version != "" {
		if missing == 0 && partial == 0 {
			s.GateStatus = "PASS"
		} else {
			s.GateStatus = "FAIL"
		}
	}
	return s
}

// --- REQ-API-007: Agent Claims Endpoint ---

type apiClaim struct {
	ReqID           string `json:"req_id"`
	AgentID         string `json:"agent_id"`
	ClaimedAt       string `json:"claimed_at"`
	LastHeartbeat   string `json:"last_heartbeat"`
	Stale           bool   `json:"stale"`
	RequirementText string `json:"requirement_text"`
}

type apiClaimsSummary struct {
	TotalActive int      `json:"total_active"`
	StaleCount  int      `json:"stale_count"`
	Agents      []string `json:"agents"`
}

type apiClaimsResponse struct {
	ActiveClaims []apiClaim       `json:"active_claims"`
	Summary      apiClaimsSummary `json:"summary"`
}

func handleAPIAgentClaims(db *database.Database, claimsDir string, staleTimeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeAPIError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		store, err := orchestration.NewClaimStore(claimsDir)
		if err != nil {
			writeJSON(w, apiClaimsResponse{
				ActiveClaims: []apiClaim{},
				Summary:      apiClaimsSummary{Agents: []string{}},
			})
			return
		}

		claims, err := store.List()
		if err != nil {
			writeJSON(w, apiClaimsResponse{
				ActiveClaims: []apiClaim{},
				Summary:      apiClaimsSummary{Agents: []string{}},
			})
			return
		}

		now := time.Now().UTC()
		agentSet := make(map[string]bool)
		staleCount := 0
		var apiClaims []apiClaim

		for _, c := range claims {
			stale := now.Sub(c.ClaimedAt) > staleTimeout
			if stale {
				staleCount++
			}
			agentSet[c.AgentID] = true

			reqText := ""
			if req := db.Get(c.ReqID); req != nil {
				reqText = req.RequirementText
			}

			apiClaims = append(apiClaims, apiClaim{
				ReqID:           c.ReqID,
				AgentID:         c.AgentID,
				ClaimedAt:       c.ClaimedAt.Format(time.RFC3339),
				LastHeartbeat:   c.ClaimedAt.Format(time.RFC3339),
				Stale:           stale,
				RequirementText: reqText,
			})
		}
		if apiClaims == nil {
			apiClaims = []apiClaim{}
		}

		agents := make([]string, 0, len(agentSet))
		for a := range agentSet {
			agents = append(agents, a)
		}
		sort.Strings(agents)

		resp := apiClaimsResponse{
			ActiveClaims: apiClaims,
			Summary: apiClaimsSummary{
				TotalActive: len(apiClaims),
				StaleCount:  staleCount,
				Agents:      agents,
			},
		}
		writeJSON(w, resp)
	}
}
