package schema

// ISO26262Schema extends the core schema with columns for ISO 26262
// automotive functional safety. Columns cover Automotive Safety Integrity
// Level (ASIL), safety goals, fault tolerance, and verification artifacts
// required by the standard.
var ISO26262Schema = CoreSchema.Extend("iso26262", []Column{
	// Classification
	{Name: "asil_level", Type: TypeEnum, EnumValues: []string{"QM", "A", "B", "C", "D"}, Description: "Automotive Safety Integrity Level"},
	{Name: "safety_goal_id", Type: TypeString, Description: "Associated safety goal reference"},
	{Name: "functional_safety_req", Type: TypeString, Description: "Functional safety requirement ID"},

	// Hazard Analysis
	{Name: "hazard_id", Type: TypeString, Description: "Hazard and risk analysis reference"},
	{Name: "severity", Type: TypeEnum, EnumValues: []string{"S0", "S1", "S2", "S3"}, Description: "Severity class (S0=no injury to S3=life-threatening)"},
	{Name: "exposure", Type: TypeEnum, EnumValues: []string{"E0", "E1", "E2", "E3", "E4"}, Description: "Probability of exposure"},
	{Name: "controllability", Type: TypeEnum, EnumValues: []string{"C0", "C1", "C2", "C3"}, Description: "Controllability class"},

	// Architecture
	{Name: "fault_tolerance", Type: TypeInt, Description: "Required fault tolerance level (0-2)"},
	{Name: "diagnostic_coverage", Type: TypeEnum, EnumValues: []string{"low", "medium", "high"}, Description: "Diagnostic coverage requirement"},
	{Name: "decomposition_target", Type: TypeString, Description: "ASIL decomposition target (e.g., B(D)=B+QM)"},

	// Verification
	{Name: "verification_method", Type: TypeEnum, EnumValues: []string{"review", "analysis", "simulation", "test"}, Description: "Verification method per ISO 26262"},
	{Name: "test_method", Type: TypeEnum, EnumValues: []string{"requirements-based", "fault-injection", "back-to-back", "interface"}, Description: "Test method classification"},

	// Evidence
	{Name: "evidence_artifact", Type: TypeString, Description: "Path to verification evidence"},
	{Name: "safety_case_ref", Type: TypeString, Description: "Safety case section reference"},
	{Name: "confirmation_review", Type: TypeString, Description: "Confirmation review reference"},
})

func init() {
	ISO26262Schema.Description = "ISO 26262 automotive functional safety schema"
	Register(ISO26262Schema)
}
