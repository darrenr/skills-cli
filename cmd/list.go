package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/darrenr/skills-cli/internal/installer"
	"github.com/darrenr/skills-cli/internal/registry"
	"github.com/darrenr/skills-cli/internal/skill"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available skills from the registry",
	Long: `List skills from the bundled registry (or cached registry if available).

Filter by category or source to narrow results.

Examples:
  skills-cli list
  skills-cli list --installed
  skills-cli list --category mcp
  skills-cli list --source github/awesome-copilot`,
	RunE: runList,
}

func init() {
	listCmd.Flags().String("category", "", "filter by category (e.g. git, mcp, github)")
	listCmd.Flags().String("source", "", "filter by source repo (e.g. github/awesome-copilot)")
	listCmd.Flags().Bool("installed", false, "show only installed skills in the current project")
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	cat, _ := cmd.Flags().GetString("category")
	src, _ := cmd.Flags().GetString("source")
	installedOnly, _ := cmd.Flags().GetBool("installed")
	output := viper.GetString("output")

	r, err := loadRegistry()
	if err != nil {
		return fmt.Errorf("load registry: %w", err)
	}

	results := registry.Search(r, registry.SearchOptions{
		Category: cat,
		Source:   src,
	})

	if installedOnly {
		manifestPath, err := installer.DefaultManifestPath()
		if err != nil {
			return err
		}
		manifest, err := installer.LoadManifest(manifestPath)
		if err != nil {
			return fmt.Errorf("load manifest: %w", err)
		}

		installedNames := make(map[string]struct{})

		for _, s := range append([]skill.InstalledSkill(nil), manifest.Skills...) {
			installed, err := manifest.FindExistingInCurrentProject(s.Name)
			if err != nil {
				return fmt.Errorf("resolve installed skill %q: %w", s.Name, err)
			}
			if installed != nil {
				installedNames[installed.Name] = struct{}{}
			}
		}

		scannedSkills, err := scanProjectInstalledSkills()
		if err != nil {
			return err
		}
		for name := range scannedSkills {
			installedNames[name] = struct{}{}
		}

		var filtered []registry.SkillEntry
		for _, e := range results {
			if _, ok := installedNames[e.Name]; ok {
				filtered = append(filtered, e)
			}
		}
		results = filtered
	}

	if len(results) == 0 {
		if isStructuredOutput(output) {
			return printSkillEntries(results, output)
		}
		fmt.Println("No skills found.")
		return nil
	}

	return printSkillEntries(results, output)
}

func scanProjectInstalledSkills() (map[string]string, error) {
	skills := make(map[string]string)
	for _, base := range []string{skill.TargetProjectCopilot, skill.TargetProjectAgents, skill.TargetProjectClaude} {
		dirs, err := os.ReadDir(base)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("read installed skills dir %s: %w", filepath.Clean(base), err)
		}
		for _, d := range dirs {
			if d.IsDir() {
				skills[d.Name()] = filepath.Join(base, d.Name())
			}
		}
	}
	return skills, nil
}
