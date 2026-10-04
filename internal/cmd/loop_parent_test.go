package cmd

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func rtmxModuleRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller")
	}
	return filepath.Join(filepath.Dir(file), "..", "..")
}

func TestParentDeliveryLoop(t *testing.T) {
	rtmx.Req(t, "REQ-ORCH-019")
	root := rtmxModuleRoot(t)

	dbFile, err := os.Open(filepath.Join(root, ".rtmx", "database.csv"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = dbFile.Close() }()
	rows, err := csv.NewReader(dbFile).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	status := map[string]string{}
	for _, row := range rows[1:] {
		if len(row) > 8 {
			status[row[0]] = row[8]
		}
	}
	for _, id := range []string{"REQ-ORCH-019a", "REQ-ORCH-019b", "REQ-ORCH-019c"} {
		if status[id] != "COMPLETE" {
			t.Fatalf("parent cannot close: %s is %q", id, status[id])
		}
	}

	feature, err := os.ReadFile(filepath.Join(root, "features", "agent_delivery_loop.feature"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(feature)
	for _, needle := range []string{
		"rtmx loop",
		"rtmx loop install",
		"launchctl bootstrap",
		"systemctl --user",
		"does not need managed-sync OAuth or Stripe",
	} {
		if !strings.Contains(text, needle) {
			t.Fatalf("deploy path missing %q", needle)
		}
	}

	// A timer wake with no new merge must not tick, and the runner waits.
	waits := 0
	prevWait := loopWait
	loopWait = func() { waits++ }
	t.Cleanup(func() { loopWait = prevWait })
	withLoopWatch(t, &scriptedMerged{}, 2)
	loopWait = func() { waits++ }

	db := testDBHeader +
		"REQ-EX-060,CLI,Commands,Waiting,Pass,mod,TestA,Unit Test,MISSING,HIGH,1,,1.0,,,,,,,\n"
	dir := setupNextTestProject(t, db)
	writeLoopReqMD(t, dir, "REQ-EX-060", 2)
	if err := SaveLoopCursor(LoopCursorPath(dir), LoopCursor{}); err != nil {
		t.Fatal(err)
	}
	out, err := runLoopCmd(t, dir, "loop")
	if err != nil {
		t.Fatalf("loop: %v\n%s", err, out)
	}
	if strings.Contains(out, `"req_id"`) {
		t.Fatalf("timer elapsed must not start a tick:\n%s", out)
	}
	if waits != 2 {
		t.Fatalf("expected a wait per poll, got %d", waits)
	}
	claims, _ := filepath.Glob(filepath.Join(dir, ".rtmx", "claims", "*.json"))
	if len(claims) != 0 {
		t.Fatalf("idle loop invented work: %v", claims)
	}

	// An edge's first mutations are claim and decomposition, not a commit.
	edges := []LoopEdge{{ReqID: "REQ-MERGED", Source: "pr", PRNumber: 4}}
	prevEdges := loopEdgesOverride
	loopEdgesOverride = &edges
	t.Cleanup(func() { loopEdgesOverride = prevEdges })
	coarse := testDBHeader +
		"REQ-EX-061,CLI,Commands,Coarse parent,Pass,mod,TestA,Unit Test,MISSING,HIGH,1,,1.0,,,,,,,\n"
	work := setupNextTestProject(t, coarse)
	writeLoopReqMD(t, work, "REQ-EX-061", 6)
	out, err = runLoopCmd(t, work, "loop", "--once")
	if err != nil {
		t.Fatalf("tick: %v\n%s", err, out)
	}
	if !strings.Contains(out, `"decomposed":true`) || !strings.Contains(out, "REQ-EX-061") {
		t.Fatalf("tick plan: %s", out)
	}
	if _, err := os.Stat(filepath.Join(work, ".rtmx", "claims", "REQ-EX-061.json")); err != nil {
		t.Fatal(err)
	}
	kids, _ := filepath.Glob(filepath.Join(work, ".rtmx", "requirements", "CLI", "REQ-EX-061*.md"))
	if len(kids) < 2 {
		t.Fatalf("decompose did not write children: %v", kids)
	}
	if _, err := os.Stat(filepath.Join(work, ".git")); !os.IsNotExist(err) {
		t.Fatal("tick created a git repository")
	}
}
