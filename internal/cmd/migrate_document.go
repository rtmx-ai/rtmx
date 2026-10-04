package cmd

import (
	"fmt"

	"github.com/rtmx-ai/rtmx/internal/docmodel/migrate"
	"github.com/spf13/cobra"
)

var (
	migrateDocInput    string
	migrateDocOutput   string
	migrateDocPrevious string
)

var migrateDocumentCmd = &cobra.Command{
	Use:   "document",
	Short: "Migrate CSV+Markdown ACs into structured JSONL + companion stubs",
	Long: `Convert a CSV+Markdown fixture (or spike layout) into the requirement
document model: requirements.jsonl, requirements.document.json, companion
Markdown stubs, and a CSV projection.

Does not change the project's default on-disk CSV. Remigrate with --previous
to keep stable ac_id values when AC statement text is unchanged (ADR-0007).

Examples:
  rtmx migrate document --input path/to/fixture --output path/to/out
  rtmx migrate document --input fixture --output out --previous out/requirements.jsonl`,
	RunE: runMigrateDocument,
}

func init() {
	migrateDocumentCmd.Flags().StringVar(&migrateDocInput, "input", "", "input directory with database.csv and requirements/*.md")
	migrateDocumentCmd.Flags().StringVar(&migrateDocOutput, "output", "", "output directory for JSONL/document/companions")
	migrateDocumentCmd.Flags().StringVar(&migrateDocPrevious, "previous", "", "optional previous requirements.jsonl for stable ac_id rematerialization")
	_ = migrateDocumentCmd.MarkFlagRequired("input")
	_ = migrateDocumentCmd.MarkFlagRequired("output")
	migrateCmd.AddCommand(migrateDocumentCmd)
}

func runMigrateDocument(cmd *cobra.Command, _ []string) error {
	fresh, err := migrate.MigrateFixtureDir(migrateDocInput)
	if err != nil {
		return fmt.Errorf("migrate input: %w", err)
	}
	reqs := fresh
	if migrateDocPrevious != "" {
		prev, err := migrate.LoadJSONL(migrateDocPrevious)
		if err != nil {
			return fmt.Errorf("load --previous: %w", err)
		}
		reqs = migrate.Remigrate(prev, fresh)
	}
	if err := migrate.ExportDir(migrateDocOutput, reqs); err != nil {
		return err
	}
	cmd.Printf("Wrote structured document store to %s\n", migrateDocOutput)
	cmd.Printf("  requirements.jsonl (%d)\n", len(reqs))
	cmd.Printf("  requirements.document.json\n")
	cmd.Printf("  companions/\n")
	cmd.Printf("  database.projection.csv\n")
	cmd.Println("Default project CSV unchanged (no 2.0 flip).")
	return nil
}
