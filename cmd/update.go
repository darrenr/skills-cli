package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/darrenr/skills-cli/internal/installer"
	"github.com/darrenr/skills-cli/internal/skill"
	"github.com/darrenr/skills-cli/internal/source"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update [name]",
	Short: "Update installed skills to their latest version",
	Long: `Re-download one or all installed skills from their source repositories.

If no name is given, all installed skills are updated.

Examples:
  skills-cli update conventional-commit
  skills-cli update
  skills-cli update my-skill --dry-run`,
	RunE: runUpdate,
}

func init() {
	updateCmd.Flags().Bool("dry-run", false, "show what would be updated without making changes")
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	r, err := loadRegistry()
	if err != nil {
		return fmt.Errorf("load registry: %w", err)
	}

	manifestPath, err := installer.DefaultManifestPath()
	if err != nil {
		return err
	}
	manifest, err := installer.LoadManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("load manifest: %w", err)
	}

	scannedSkills, err := scanProjectInstalledSkills()
	if err != nil {
		return err
	}

	manifestChanged := false
	for name, installPath := range scannedSkills {
		existing, err := manifest.FindExistingInCurrentProject(name)
		if err != nil {
			return fmt.Errorf("resolve installed skill %q: %w", name, err)
		}
		if existing == nil || existing.InstallPath != installPath {
			manifest.Upsert(skill.InstalledSkill{
				Name:        name,
				InstallPath: installPath,
				InstalledAt: time.Now(),
			})
			manifestChanged = true
		}
	}
	if manifestChanged {
		if err := manifest.Save(); err != nil {
			return fmt.Errorf("save manifest: %w", err)
		}
	}

	token := os.Getenv("GITHUB_TOKEN")
	fetcher := source.NewGitHubFetcher(token)

	targets := args
	if len(targets) == 0 {
		for name := range scannedSkills {
			targets = append(targets, name)
		}
		sort.Strings(targets)
	}

	if len(targets) == 0 {
		fmt.Println("No skills are installed.")
		return nil
	}

	for _, name := range targets {
		if dryRun {
			fmt.Printf("Would update: %s\n", name)
			continue
		}
		fmt.Printf("Updating %s...\n", name)
		path, err := installer.Update(context.Background(), name,
			installer.UpdateOptions{DryRun: dryRun},
			fetcher, r, manifest)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			continue
		}
		fmt.Printf("  ✓ updated at %s\n", path)
	}
	return nil
}
