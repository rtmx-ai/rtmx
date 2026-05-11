package schema

// CoreSchema defines the 21 standard columns used by all RTMX databases.
var CoreSchema = New("core", []Column{
	{Name: "req_id", Type: TypeString, Required: true, Description: "Unique requirement identifier"},
	{Name: "category", Type: TypeString, Required: true, Description: "High-level grouping"},
	{Name: "subcategory", Type: TypeString, Description: "Detailed classification"},
	{Name: "requirement_text", Type: TypeString, Required: true, Description: "Human-readable requirement description"},
	{Name: "target_value", Type: TypeString, Description: "Quantitative acceptance criteria"},
	{Name: "test_module", Type: TypeString, Description: "Test file path"},
	{Name: "test_function", Type: TypeString, Description: "Test function name"},
	{Name: "validation_method", Type: TypeString, Description: "Testing approach"},
	{Name: "status", Type: TypeEnum, EnumValues: []string{"COMPLETE", "PARTIAL", "MISSING", "NOT_STARTED"}, Description: "Requirement status"},
	{Name: "priority", Type: TypeEnum, EnumValues: []string{"P0", "HIGH", "MEDIUM", "LOW"}, Description: "Priority level"},
	{Name: "phase", Type: TypeInt, Description: "Development phase number"},
	{Name: "notes", Type: TypeString, Description: "Additional context"},
	{Name: "effort_weeks", Type: TypeFloat, Description: "Estimated effort in weeks"},
	{Name: "dependencies", Type: TypeSet, Description: "Pipe-separated dependency IDs"},
	{Name: "blocks", Type: TypeSet, Description: "Pipe-separated blocked IDs"},
	{Name: "assignee", Type: TypeString, Description: "Person responsible"},
	{Name: "sprint", Type: TypeString, Description: "Target version"},
	{Name: "started_date", Type: TypeDate, Description: "When work began"},
	{Name: "completed_date", Type: TypeDate, Description: "When completed"},
	{Name: "requirement_file", Type: TypeString, Description: "Path to spec markdown"},
	{Name: "external_id", Type: TypeString, Description: "Cross-repo reference"},
})

func init() {
	CoreSchema.Description = "RTMX core schema with 21 standard columns"
}
