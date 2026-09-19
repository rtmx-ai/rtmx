package sync

import (
	"github.com/reearth/ygo/crdt"
)

// Additive AC extension map names (REQ-DATA-005). v1 clients read only
// RequirementsMapName and ignore these.
const (
	MetadataMapName       = "metadata"
	ACMapName             = "requirement_acs"
	BindingMapName        = "requirement_bindings"
	CapabilityFlagKey     = "capability"
	CapabilityRTMDocV0    = "rtm-doc/v0"
	SchemaVersionKey      = "schema_version"
	SchemaVersionACExt    = "1.1.0-ac"
)

// ApplyACExtension writes requirement_acs and requirement_bindings when
// enabled. When disabled it does not create those maps, so v1 rooms stay
// unchanged. Existing requirements entries are never rewritten.
func ApplyACExtension(doc *crdt.Doc, enabled bool, reqID, acsJSON, bindingsJSON string) {
	if doc == nil || !enabled || reqID == "" {
		return
	}
	meta := doc.GetMap(MetadataMapName)
	acs := doc.GetMap(ACMapName)
	bindings := doc.GetMap(BindingMapName)
	doc.Transact(func(txn *crdt.Transaction) {
		meta.Set(txn, CapabilityFlagKey, CapabilityRTMDocV0)
		meta.Set(txn, SchemaVersionKey, SchemaVersionACExt)
		acs.Set(txn, reqID, acsJSON)
		bindings.Set(txn, reqID, bindingsJSON)
	}, localOrigin{})
}

// V1RequirementKeys returns requirement IDs visible to a sync-protocol-v1
// reader. Unknown maps are ignored.
func V1RequirementKeys(doc *crdt.Doc) []string {
	if doc == nil {
		return nil
	}
	return doc.GetMap(RequirementsMapName).Keys()
}

// ACExtensionEnabled reports whether this document advertised the capability.
func ACExtensionEnabled(doc *crdt.Doc) bool {
	if doc == nil {
		return false
	}
	meta := doc.GetMap(MetadataMapName)
	v, ok := meta.Get(CapabilityFlagKey)
	if !ok {
		return false
	}
	s, ok := v.(string)
	return ok && s == CapabilityRTMDocV0
}
