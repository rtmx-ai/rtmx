package schema

// DO178CSchema extends the core schema with columns for DO-178C airborne
// software certification. Columns cover Design Assurance Level (DAL),
// software level objectives, verification independence, and traceability
// artifacts required by the standard.
var DO178CSchema = CoreSchema.Extend("do178c", []Column{
	// Classification
	{Name: "dal_level", Type: TypeEnum, EnumValues: []string{"A", "B", "C", "D", "E"}, Description: "Design Assurance Level"},
	{Name: "sw_level", Type: TypeEnum, EnumValues: []string{"1", "2", "3", "4", "5"}, Description: "Software level (Table A-1 through A-7)"},

	// Objectives
	{Name: "objective_id", Type: TypeString, Description: "DO-178C objective reference (e.g., A-2.1)"},
	{Name: "objective_text", Type: TypeString, Description: "Objective description from Table A-*"},
	{Name: "independence_required", Type: TypeBool, Description: "Requires independent verification"},
	{Name: "applicability", Type: TypeEnum, EnumValues: []string{"applicable", "not-applicable", "deactivated"}, Description: "Objective applicability for target DAL"},

	// Traceability
	{Name: "trace_to_srs", Type: TypeString, Description: "Software Requirements Specification trace"},
	{Name: "trace_to_sdd", Type: TypeString, Description: "Software Design Document trace"},
	{Name: "trace_to_src", Type: TypeString, Description: "Source code trace"},
	{Name: "trace_to_test", Type: TypeString, Description: "Test case trace"},

	// Verification
	{Name: "review_authority", Type: TypeString, Description: "Designated Engineering Representative (DER)"},
	{Name: "verification_method", Type: TypeEnum, EnumValues: []string{"review", "analysis", "test", "inspection"}, Description: "Verification method per DO-178C"},
	{Name: "structural_coverage", Type: TypeEnum, EnumValues: []string{"statement", "decision", "mcdc", "none"}, Description: "Required structural coverage level"},

	// Evidence
	{Name: "evidence_artifact", Type: TypeString, Description: "Path to verification evidence"},
	{Name: "ccb_approval", Type: TypeString, Description: "Configuration Control Board approval reference"},
	{Name: "problem_report", Type: TypeString, Description: "Associated problem report ID"},
})

func init() {
	DO178CSchema.Description = "DO-178C airborne software certification schema"
	Register(DO178CSchema)
}
