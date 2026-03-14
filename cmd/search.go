package cmd

import (
	"fmt"

	"github.com/darrenr/skills-cli/internal/registry"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search skills by keyword",
	Long: `Search for skills by keyword. Matches against name, description, category, and tags.

Examples:
  skills-cli search git
  skills-cli search "commit messages" --category git
  skills-cli search go --limit 5`,
	Args: cobra.ExactArgs(1),
	RunE: runSearch,
}

func init() {
	searchCmd.Flags().String("category", "", "filter by category")
	searchCmd.Flags().String("source", "", "filter by source repo")
	searchCmd.Flags().Int("limit", 0, "maximum number of results (0 = no limit)")
	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := args[0]
	cat, _ := cmd.Flags().GetString("category")
	src, _ := cmd.Flags().GetString("source")
	limit, _ := cmd.Flags().GetInt("limit")
	output := viper.GetString("output")

	r, err := loadRegistry()
	if err != nil {
		return fmt.Errorf("load registry: %w", err)
	}

	results := registry.Search(r, registry.SearchOptions{
		Query:    query,
		Category: cat,
		Source:   src,
		Limit:    limit,
	})

	if len(results) == 0 {
		if isStructuredOutput(output) {
			return printSkillEntries(results, output)
		}
		fmt.Printf("No skills found matching %q.\n", query)
		return nil
	}

	if !isStructuredOutput(output) {
		fmt.Printf("Found %d skill(s) matching %q:\n\n", len(results), query)
	}
	return printSkillEntries(results, output)
}
