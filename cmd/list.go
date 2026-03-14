package cmd

import (
	"fmt"

	"github.com/darrenr/skills-cli/internal/installer"
	"github.com/darrenr/skills-cli/internal/registry"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available skills from the registry",
	Long: `List skills from the bundled registry (or cached registry if available).

Filter by category or source to narrow results.`,
	RunE: runList,
}

func init() {
	listCmd.Flags().String("category", "", "filter by category (e.g. git, golang, github)")
	listCmd.Flags().String("source", "", "filter by source repo (e.g. github/awesome-copilot)")
	listCmd.Flags().Bool("installed", false, "show only installed skills")
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
		var filtered []registry.SkillEntry
		for _, e := range results {
			if manifest.Find(e.Name) != nil {
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
