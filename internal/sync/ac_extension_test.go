package sync

import (
	"testing"

	"github.com/reearth/ygo/crdt"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestACExtensionDisabledLeavesV1Only(t *testing.T) {
	rtmx.Req(t, "REQ-DATA-005")
	doc := crdt.New()
	root := doc.GetMap(RequirementsMapName)
	doc.Transact(func(txn *crdt.Transaction) {
		entry := crdt.NewMapPrelim()
		entry.Set(txn, "status", "COMPLETE")
		root.Set(txn, "REQ-AUTH-001", entry)
	}, localOrigin{})

	ApplyACExtension(doc, false, "REQ-AUTH-001", `[{"ac_id":"AC-1"}]`, `[]`)

	keys := V1RequirementKeys(doc)
	if len(keys) != 1 || keys[0] != "REQ-AUTH-001" {
		t.Fatalf("v1 keys = %v", keys)
	}
	if ACExtensionEnabled(doc) {
		t.Fatal("capability must stay off when flag disabled")
	}
	// Disabled path must not materialize the additive maps as populated.
	if _, ok := doc.GetMap(ACMapName).Get("REQ-AUTH-001"); ok {
		t.Fatal("disabled flag wrote requirement_acs")
	}
}

func TestACExtensionEnabledIsAdditive(t *testing.T) {
	rtmx.Req(t, "REQ-DATA-005")
	doc := crdt.New()
	root := doc.GetMap(RequirementsMapName)
	doc.Transact(func(txn *crdt.Transaction) {
		entry := crdt.NewMapPrelim()
		entry.Set(txn, "status", "PARTIAL")
		root.Set(txn, "REQ-AUTH-001", entry)
	}, localOrigin{})

	before := decodeRequirement(mustGet(t, root, "REQ-AUTH-001"))
	ApplyACExtension(doc, true, "REQ-AUTH-001", `[{"ac_id":"AC-1","statement":"reject"}]`, `[{"binding_id":"TB-1","ac_id":"AC-1"}]`)

	after := decodeRequirement(mustGet(t, root, "REQ-AUTH-001"))
	if before["status"] != after["status"] || after["status"] != "PARTIAL" {
		t.Fatalf("v1 requirement mutated: before=%v after=%v", before, after)
	}
	if !ACExtensionEnabled(doc) {
		t.Fatal("expected capability flag")
	}
	acs, ok := doc.GetMap(ACMapName).Get("REQ-AUTH-001")
	if !ok || acs == "" {
		t.Fatal("missing requirement_acs")
	}
	if _, ok := doc.GetMap(BindingMapName).Get("REQ-AUTH-001"); !ok {
		t.Fatal("missing requirement_bindings")
	}
	// v1 reader still sees the original requirement id only.
	if got := V1RequirementKeys(doc); len(got) != 1 || got[0] != "REQ-AUTH-001" {
		t.Fatalf("v1 keys = %v", got)
	}
}

func mustGet(t *testing.T, m *crdt.YMap, key string) any {
	t.Helper()
	v, ok := m.Get(key)
	if !ok {
		t.Fatalf("missing %s", key)
	}
	return v
}
