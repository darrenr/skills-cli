package cmd

import (
	"fmt"
	"os"

	"github.com/darrenr/skills-cli/internal/installer"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:     "remove <name> [name...]",
	Aliases: []string{"rm", "uninstall"},
	Short:   "Remove installed skills",
	Long: `Delete one or more installed skills and remove them from the manifest.

Examples:
  skills-cli remove conventional-commit
  skills-cli rm go-test gh-cli
  skills-cli remove my-skill --force`,
	Args: cobra.MinimumNArgs(1),
	RunE: runRemove,
}

func init() {
	removeCmd.Flags().Bool("force", false, "suppress error if skill is not installed")
	rootCmd.AddCommand(removeCmd)
}

func runRemove(cmd *cobra.Command, args []string) error {
	force, _ := cmd.Flags().GetBool("force")

	manifestPath, err := installer.DefaultManifestPath()
	if err != nil {
		return err
	}
	manifest, err := installer.LoadManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("load manifest: %w", err)
	}

	for _, name := range args {
		fmt.Printf("Removing %s...\n", name)
		if err := installer.Remove(name, installer.RemoveOptions{Force: force}, manifest); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			continue
		}
		fmt.Printf("  ✓ removed\n")
	}
	return nil
}
