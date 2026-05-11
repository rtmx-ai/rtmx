package database

// MigrationResult describes what changed when migrating a database schema.
type MigrationResult struct {
	// AddedColumns lists standard columns that were missing and added.
	AddedColumns []string

	// ExtraColumns lists non-standard columns preserved from the original.
	ExtraColumns []string

	// Migrated is true if any columns were added.
	Migrated bool
}

// DetectMigration checks whether the database header is missing any
// standard columns. Returns a MigrationResult describing the gap.
func (db *Database) DetectMigration() MigrationResult {
	header := db.Header()
	headerSet := make(map[string]bool, len(header))
	for _, h := range header {
		headerSet[normalizeColumnName(h)] = true
	}

	var result MigrationResult

	for _, std := range standardColumns {
		if !headerSet[std] {
			result.AddedColumns = append(result.AddedColumns, std)
		}
	}

	for _, h := range header {
		normalized := normalizeColumnName(h)
		isStandard := false
		for _, std := range standardColumns {
			if normalized == std {
				isStandard = true
				break
			}
		}
		if !isStandard {
			result.ExtraColumns = append(result.ExtraColumns, h)
		}
	}

	result.Migrated = len(result.AddedColumns) > 0
	return result
}

// Migrate adds any missing standard columns to the database header.
// Existing data is preserved -- new columns get empty default values.
// Returns a MigrationResult describing what was changed.
func (db *Database) Migrate() MigrationResult {
	result := db.DetectMigration()
	if !result.Migrated {
		return result
	}

	// Build new header with all standard columns plus any extras
	newHeader := make([]string, len(standardColumns))
	copy(newHeader, standardColumns)

	// Preserve extra columns at the end
	for _, extra := range result.ExtraColumns {
		newHeader = append(newHeader, extra)
	}

	db.originalHeader = newHeader
	db.dirty = true
	return result
}

// SchemaVersion returns a version identifier based on the number of
// standard columns present. This enables detection of schema drift
// between CLI versions.
func SchemaVersion() int {
	return len(standardColumns)
}

// DatabaseSchemaVersion returns the schema version of a loaded database
// based on how many standard columns its header contains.
func (db *Database) DatabaseSchemaVersion() int {
	header := db.Header()
	headerSet := make(map[string]bool, len(header))
	for _, h := range header {
		headerSet[normalizeColumnName(h)] = true
	}

	count := 0
	for _, std := range standardColumns {
		if headerSet[std] {
			count++
		}
	}
	return count
}

// NeedsMigration returns true if the database header is missing standard columns.
func (db *Database) NeedsMigration() bool {
	return db.DatabaseSchemaVersion() < SchemaVersion()
}
