package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rtmx-ai/rtmx/internal/output"
	"github.com/rtmx-ai/rtmx/internal/tui/linkpicker"
	"github.com/spf13/cobra"
)

var linkCmd = &cobra.Command{
	Use:   "link",
	Short: "Link an issue store and bootstrap the RTM",
	Long: `Interactively choose an issue store (GitHub, Jira, Asana, Monday, GitLab,
or Tests) and run the same import path as bootstrap --from-<provider>.

On a TTY, a provider list is shown with an optional pre-selection inferred
from git remotes (github.com / gitlab hosts). Non-TTY environments should
use rtmx bootstrap --from-* instead.

Examples:
    rtmx link
    rtmx link --dry-run
    rtmx bootstrap          # same picker when no --from-* on a TTY`,
	RunE: runLink,
}

func init() {
	linkCmd.Flags().BoolVar(&bootstrapDryRun, "dry-run", false, "preview without writing files")
	linkCmd.Flags().BoolVar(&bootstrapMerge, "merge", false, "merge with existing RTM (default: replace)")
	linkCmd.Flags().StringVar(&bootstrapPrefix, "prefix", "REQ", "requirement ID prefix")
	rootCmd.AddCommand(linkCmd)
}

func runLink(cmd *cobra.Command, args []string) error {
	if noColor {
		output.DisableColor()
	}
	if !isTerminal() {
		cmd.Printf("%s\n", output.Color(
			"rtmx link requires a TTY. Use rtmx bootstrap --from-github|--from-jira|--from-asana|--from-monday|--from-gitlab|--from-tests",
			output.Yellow,
		))
		return NewExitError(1, "link requires a TTY")
	}
	if err := selectBootstrapSourceInteractive(cmd); err != nil {
		return err
	}
	if !bootstrapSourceSelected() {
		cmd.Println("Skipped linking an issue store.")
		return nil
	}
	return runBootstrap(cmd, args)
}

// runProviderPicker runs the Bubble Tea link picker. Overridable in tests.
var runProviderPicker = func(hint string) (string, error) {
	m := linkpicker.New(hint)
	p := tea.NewProgram(m)
	final, err := p.Run()
	if err != nil {
		return "", err
	}
	model, ok := final.(linkpicker.Model)
	if !ok {
		return "", fmt.Errorf("unexpected picker model type %T", final)
	}
	return model.Chosen(), nil
}

// selectBootstrapSourceInteractive infers a hint and runs the provider picker,
// then sets the matching bootstrap --from-* flag. Skip / cancel clears flags.
func selectBootstrapSourceInteractive(cmd *cobra.Command) error {
	remotes, _ := listGitRemotes("")
	hint := InferProviderHint(remotes)
	if hint != "" {
		cmd.Printf("Inferred provider hint from git remotes: %s\n", hint)
	}
	choice, err := runProviderPicker(hint)
	if err != nil {
		return fmt.Errorf("link picker: %w", err)
	}
	applyLinkChoice(choice)
	return nil
}

// applyLinkChoice sets bootstrap source flags from a picker ID.
// Returns false when the choice is skip/empty (no source selected).
func applyLinkChoice(id string) bool {
	bootstrapFromTests = false
	bootstrapFromGitHub = false
	bootstrapFromJira = false
	bootstrapFromAsana = false
	bootstrapFromMonday = false
	bootstrapFromGitLab = false
	switch id {
	case "github":
		bootstrapFromGitHub = true
	case "jira":
		bootstrapFromJira = true
	case "asana":
		bootstrapFromAsana = true
	case "monday":
		bootstrapFromMonday = true
	case "gitlab":
		bootstrapFromGitLab = true
	case "tests":
		bootstrapFromTests = true
	case "skip", "":
		return false
	default:
		return false
	}
	return true
}
