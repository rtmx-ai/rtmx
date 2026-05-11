package schema

// PhoenixSchema extends the core schema with 25 additional columns for
// defense and aerospace validation taxonomy: scope markers, technique
// markers, environment markers, metrics, and hardware columns.
var PhoenixSchema = CoreSchema.Extend("phoenix", []Column{
	// Scope markers (boolean)
	{Name: "scope_unit", Type: TypeBool, Description: "Single component isolation test"},
	{Name: "scope_integration", Type: TypeBool, Description: "Multiple components"},
	{Name: "scope_system", Type: TypeBool, Description: "End-to-end system test"},

	// Technique markers (boolean)
	{Name: "technique_nominal", Type: TypeBool, Description: "Happy path, typical parameters"},
	{Name: "technique_parametric", Type: TypeBool, Description: "Systematic parameter space exploration"},
	{Name: "technique_monte_carlo", Type: TypeBool, Description: "Random scenario testing"},
	{Name: "technique_stress", Type: TypeBool, Description: "Boundary/edge cases, extreme conditions"},

	// Environment markers (boolean)
	{Name: "env_simulation", Type: TypeBool, Description: "Pure software, synthetic signals"},
	{Name: "env_hil", Type: TypeBool, Description: "Real hardware, controlled signals"},
	{Name: "env_anechoic", Type: TypeBool, Description: "RF characterization chamber"},
	{Name: "env_static_field", Type: TypeBool, Description: "Outdoor, stationary targets"},
	{Name: "env_dynamic_field", Type: TypeBool, Description: "Outdoor, moving targets"},

	// Metrics columns
	{Name: "baseline_metric", Type: TypeFloat, Description: "Initial measurement value"},
	{Name: "current_metric", Type: TypeFloat, Description: "Current measurement value"},
	{Name: "target_metric", Type: TypeFloat, Description: "Target measurement value"},
	{Name: "metric_unit", Type: TypeString, Description: "Unit of measurement"},

	// Hardware columns
	{Name: "lead_time_weeks", Type: TypeFloat, Description: "Hardware procurement lead time"},
	{Name: "supplier_part", Type: TypeString, Description: "Supplier part number"},

	// Classification columns
	{Name: "dal_level", Type: TypeEnum, EnumValues: []string{"A", "B", "C", "D", "E"}, Description: "Design Assurance Level (DO-178C)"},
	{Name: "asil_level", Type: TypeEnum, EnumValues: []string{"QM", "A", "B", "C", "D"}, Description: "Automotive Safety Integrity Level (ISO 26262)"},
	{Name: "verification_objective", Type: TypeString, Description: "Formal verification objective"},
	{Name: "trace_to_srs", Type: TypeString, Description: "Software Requirements Specification trace"},
	{Name: "trace_to_hrd", Type: TypeString, Description: "Hardware Requirements Document trace"},
	{Name: "review_authority", Type: TypeString, Description: "Designated review authority"},
	{Name: "evidence_artifact", Type: TypeString, Description: "Path to verification evidence"},
})

func init() {
	PhoenixSchema.Description = "Phoenix extension schema for defense and aerospace validation taxonomy"
	Register(PhoenixSchema)
}
