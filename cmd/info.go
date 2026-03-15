package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/darrenr/skills-cli/internal/registry"
	"github.com/darrenr/skills-cli/internal/source"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <skill-name>",
	Short: "Show detailed information about a skill",
	Long: `Display metadata and the full SKILL.md content for a skill.

The skill entry is looked up in the registry, then the SKILL.md is fetched
from the source repository and displayed. Requires internet access.`,
	Args: cobra.ExactArgs(1),
	RunE: runInfo,
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

func runInfo(cmd *cobra.Command, args []string) error {
	name := args[0]

	r, err := loadRegistry()
	if err != nil {
		return fmt.Errorf("load registry: %w", err)
	}

	results := registry.Search(r, registry.SearchOptions{Query: name, Limit: 10})
	entry, found := findExact(results, name)
	if !found {
		return fmt.Errorf("skill %q not found in registry", name)
	}

	printEntryHeader(entry)

	// Fetch full SKILL.md from GitHub.
	token := os.Getenv("GITHUB_TOKEN")
	fetcher := source.NewGitHubFetcher(token)
	skillMDPath := strings.TrimSuffix(entry.Source.Path, "/") + "/SKILL.md"
	data, err := fetcher.FetchFile(context.Background(), entry.Source.Repo, entry.Source.Ref, skillMDPath)
	if err != nil {
		warnErr := fmt.Errorf("could not fetch SKILL.md: %w", err)
		fmt.Fprintf(os.Stderr, "\nwarning: %v\n", warnErr)
		return nil
	}

	fmt.Println("\n--- SKILL.md ---")
	fmt.Println(string(data))
	return nil
}

func findExact(entries []registry.SkillEntry, name string) (registry.SkillEntry, bool) {
	for _, e := range entries {
		if e.Name == name {
			return e, true
		}
	}
	return registry.SkillEntry{}, false
}

func printEntryHeader(e registry.SkillEntry) {
	fmt.Printf("Name:        %s\n", e.Name)
	fmt.Printf("Description: %s\n", e.Description)
	if e.Category != "" {
		fmt.Printf("Category:    %s\n", e.Category)
	}
	if len(e.Tags) > 0 {
		fmt.Printf("Tags:        %s\n", strings.Join(e.Tags, ", "))
	}
	if e.License != "" {
		fmt.Printf("License:     %s\n", e.License)
	}
	fmt.Printf("Source:      %s/%s@%s\n", e.Source.Repo, e.Source.Path, e.Source.Ref)
}
