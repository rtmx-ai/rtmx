package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/rtmx-ai/rtmx-go/internal/config"
	syncpkg "github.com/rtmx-ai/rtmx-go/internal/sync"
)

var (
	grantRole              string
	grantCategories        []string
	grantIDs               []string
	grantExcludeCategories []string
	grantExpiresAt         string
)

var grantCmd = &cobra.Command{
	Use:   "grant",
	Short: "Manage access grants for remote collaborators",
	Long: `Manage grant delegations that control what requirement data
is visible to remote collaborators.

Grants define access roles and constraints:
  dependency_viewer  - See status and dependency graph only
  status_observer    - See status of all matching requirements
  requirement_editor - Full read access + propose changes
  admin              - Full access

Examples:
  rtmx grant create upstream --role status_observer --categories AUTH,API
  rtmx grant list
  rtmx grant revoke grant-upstream-1234567890`,
}

var grantCreateCmd = &cobra.Command{
	Use:   "create ALIAS",
	Short: "Create a grant for a remote collaborator",
	Args:  cobra.ExactArgs(1),
	RunE:  runGrantCreate,
}

var grantListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all grants",
	RunE:  runGrantList,
}

var grantRevokeCmd = &cobra.Command{
	Use:   "revoke GRANT-ID",
	Short: "Revoke a grant",
	Args:  cobra.ExactArgs(1),
	RunE:  runGrantRevoke,
}

func init() {
	grantCreateCmd.Flags().StringVar(&grantRole, "role", "", "access role (dependency_viewer, status_observer, requirement_editor, admin)")
	grantCreateCmd.Flags().StringSliceVar(&grantCategories, "categories", nil, "category whitelist")
	grantCreateCmd.Flags().StringSliceVar(&grantIDs, "ids", nil, "requirement ID whitelist")
	grantCreateCmd.Flags().StringSliceVar(&grantExcludeCategories, "exclude", nil, "category blacklist")
	grantCreateCmd.Flags().StringVar(&grantExpiresAt, "expires", "", "expiry date (YYYY-MM-DD)")
	_ = grantCreateCmd.MarkFlagRequired("role")

	grantCmd.AddCommand(grantCreateCmd)
	grantCmd.AddCommand(grantListCmd)
	grantCmd.AddCommand(grantRevokeCmd)
	rootCmd.AddCommand(grantCmd)
}

func runGrantCreate(cmd *cobra.Command, args []string) error {
	grantee := args[0]

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	cfg, err := config.LoadFromDir(wd)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Validate the grantee is a known remote
	if _, ok := cfg.RTMX.Sync.Remotes[grantee]; !ok {
		return fmt.Errorf("unknown remote %q (add it first with: rtmx remote add %s --repo ORG/REPO)", grantee, grantee)
	}

	if err := syncpkg.ValidateNewGrant(cfg.RTMX.Sync.Grants, grantee, grantRole); err != nil {
		return err
	}

	grant := config.SyncGrant{
		ID:        syncpkg.GenerateGrantID(grantee),
		Grantee:   grantee,
		Role:      grantRole,
		CreatedAt: time.Now().Format("2006-01-02"),
		Constraints: config.GrantConstraint{
			Categories:        grantCategories,
			RequirementIDs:    grantIDs,
			ExcludeCategories: grantExcludeCategories,
			ExpiresAt:         grantExpiresAt,
		},
	}

	cfg.RTMX.Sync.Grants = append(cfg.RTMX.Sync.Grants, grant)

	configPath := filepath.Join(wd, ".rtmx", "config.yaml")
	if found, err := config.FindConfig(wd); err == nil {
		configPath = found
	}

	if err := cfg.Save(configPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	cmd.Printf("Created grant %s\n", grant.ID)
	cmd.Printf("  Grantee: %s\n", grant.Grantee)
	cmd.Printf("  Role: %s (visibility: %s)\n", grant.Role, syncpkg.VisibilityForRole(grant.Role))
	if len(grant.Constraints.Categories) > 0 {
		cmd.Printf("  Categories: %v\n", grant.Constraints.Categories)
	}

	return nil
}

func runGrantList(cmd *cobra.Command, _ []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	cfg, err := config.LoadFromDir(wd)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	grants := cfg.RTMX.Sync.Grants
	if len(grants) == 0 {
		cmd.Println("No grants configured.")
		cmd.Println("")
		cmd.Println("Create a grant with:")
		cmd.Println("  rtmx grant create ALIAS --role ROLE")
		return nil
	}

	cmd.Printf("%-30s %-15s %-22s %-10s %s\n", "ID", "GRANTEE", "ROLE", "STATUS", "CONSTRAINTS")
	cmd.Printf("%-30s %-15s %-22s %-10s %s\n", "--", "-------", "----", "------", "-----------")

	for _, g := range grants {
		status := "active"
		if !syncpkg.IsGrantActive(g) {
			status = "expired"
		}

		constraints := ""
		if len(g.Constraints.Categories) > 0 {
			constraints = fmt.Sprintf("categories=%v", g.Constraints.Categories)
		}
		if g.Constraints.ExpiresAt != "" {
			if constraints != "" {
				constraints += " "
			}
			constraints += fmt.Sprintf("expires=%s", g.Constraints.ExpiresAt)
		}
		if constraints == "" {
			constraints = "-"
		}

		cmd.Printf("%-30s %-15s %-22s %-10s %s\n", g.ID, g.Grantee, g.Role, status, constraints)
	}

	return nil
}

func runGrantRevoke(cmd *cobra.Command, args []string) error {
	grantID := args[0]

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	cfg, err := config.LoadFromDir(wd)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	found := false
	var remaining []config.SyncGrant
	for _, g := range cfg.RTMX.Sync.Grants {
		if g.ID == grantID {
			found = true
			continue
		}
		remaining = append(remaining, g)
	}

	if !found {
		return fmt.Errorf("grant %q not found", grantID)
	}

	cfg.RTMX.Sync.Grants = remaining

	configPath := filepath.Join(wd, ".rtmx", "config.yaml")
	if fp, err := config.FindConfig(wd); err == nil {
		configPath = fp
	}

	if err := cfg.Save(configPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	cmd.Printf("Revoked grant %s\n", grantID)

	return nil
}
