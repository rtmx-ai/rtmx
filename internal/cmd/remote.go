package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/rtmx-ai/rtmx/internal/config"
)

var (
	remoteRepo     string
	remotePath     string
	remoteDatabase string
)

var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "Manage remote RTM repositories",
	Long: `Manage remote RTM repository references for cross-repo dependency tracking.

Remotes allow you to reference requirements from other RTMX-enabled repositories
using the format sync:ALIAS:REQ-ID.

Examples:
  rtmx remote list
  rtmx remote add upstream --repo rtmx-ai/rtmx --path ../rtmx
  rtmx remote remove upstream`,
}

var remoteListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured remotes",
	RunE:  runRemoteList,
}

var remoteAddCmd = &cobra.Command{
	Use:   "add ALIAS",
	Short: "Add a remote RTM repository",
	Args:  cobra.ExactArgs(1),
	RunE:  runRemoteAdd,
}

var remoteRemoveCmd = &cobra.Command{
	Use:   "remove ALIAS",
	Short: "Remove a remote RTM repository",
	Args:  cobra.ExactArgs(1),
	RunE:  runRemoteRemove,
}

func init() {
	remoteAddCmd.Flags().StringVar(&remoteRepo, "repo", "", "GitHub repository (e.g. org/repo)")
	remoteAddCmd.Flags().StringVar(&remotePath, "path", "", "local path to cloned repository")
	remoteAddCmd.Flags().StringVar(&remoteDatabase, "database", ".rtmx/database.csv", "path to RTM database within repo")
	_ = remoteAddCmd.MarkFlagRequired("repo")

	remoteCmd.AddCommand(remoteListCmd)
	remoteCmd.AddCommand(remoteAddCmd)
	remoteCmd.AddCommand(remoteRemoveCmd)
	rootCmd.AddCommand(remoteCmd)
}

func runRemoteList(cmd *cobra.Command, _ []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	cfg, err := config.LoadFromDir(wd)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	remotes := cfg.RTMX.Sync.Remotes
	if len(remotes) == 0 {
		cmd.Println("No remotes configured.")
		cmd.Println("")
		cmd.Println("Add a remote with:")
		cmd.Println("  rtmx remote add ALIAS --repo ORG/REPO [--path PATH]")
		return nil
	}

	cmd.Printf("%-15s %-30s %-25s %s\n", "ALIAS", "REPO", "PATH", "STATUS")
	cmd.Printf("%-15s %-30s %-25s %s\n", "-----", "----", "----", "------")

	for alias, remote := range remotes {
		path := remote.Path
		if path == "" {
			path = "-"
		}

		status := "remote"
		if remote.Path != "" {
			dbPath := filepath.Join(remote.Path, remote.Database)
			if _, err := os.Stat(dbPath); err == nil {
				status = "available"
			} else {
				status = "not found"
			}
		}

		cmd.Printf("%-15s %-30s %-25s %s\n", alias, remote.Repo, path, status)
	}

	return nil
}

func runRemoteAdd(cmd *cobra.Command, args []string) error {
	alias := args[0]

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	cfg, err := config.LoadFromDir(wd)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.RTMX.Sync.Remotes == nil {
		cfg.RTMX.Sync.Remotes = make(map[string]config.SyncRemote)
	}

	if _, exists := cfg.RTMX.Sync.Remotes[alias]; exists {
		return fmt.Errorf("remote %q already exists (use 'rtmx remote remove %s' first)", alias, alias)
	}

	cfg.RTMX.Sync.Remotes[alias] = config.SyncRemote{
		Repo:     remoteRepo,
		Database: remoteDatabase,
		Path:     remotePath,
	}

	configPath := filepath.Join(wd, ".rtmx", "config.yaml")
	if found, err := config.FindConfig(wd); err == nil {
		configPath = found
	}

	if err := cfg.Save(configPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	cmd.Printf("Added remote %q -> %s\n", alias, remoteRepo)
	if remotePath != "" {
		cmd.Printf("  Local path: %s\n", remotePath)
	}

	return nil
}

func runRemoteRemove(cmd *cobra.Command, args []string) error {
	alias := args[0]

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	cfg, err := config.LoadFromDir(wd)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if _, exists := cfg.RTMX.Sync.Remotes[alias]; !exists {
		return fmt.Errorf("remote %q not found", alias)
	}

	delete(cfg.RTMX.Sync.Remotes, alias)

	configPath := filepath.Join(wd, ".rtmx", "config.yaml")
	if found, err := config.FindConfig(wd); err == nil {
		configPath = found
	}

	if err := cfg.Save(configPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	cmd.Printf("Removed remote %q\n", alias)

	return nil
}
