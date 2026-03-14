package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/darrenr/skills-cli/internal/registry"
	registrydata "github.com/darrenr/skills-cli/registry"
	"gopkg.in/yaml.v3"
)

// loadRegistry returns the best available registry (embedded seed or cache).
func loadRegistry() (*registry.Registry, error) {
	loader := registry.NewLoader(registrydata.Skills, "")
	return loader.Load()
}

// printSkillEntries writes skill entries to stdout in the requested format.
func printSkillEntries(entries []registry.SkillEntry, format string) error {
	switch strings.ToLower(format) {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(entries)
	case "yaml":
		enc := yaml.NewEncoder(os.Stdout)
		defer enc.Close()
		return enc.Encode(entries)
	default: // table
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tCATEGORY\tSOURCE\tDESCRIPTION")
		fmt.Fprintln(w, "----\t--------\t------\t-----------")
		for _, e := range entries {
			desc := e.Description
			if len(desc) > 60 {
				desc = desc[:57] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", e.Name, e.Category, e.Source.Repo, desc)
		}
		return w.Flush()
	}
}


func isStructuredOutput(format string) bool {
	switch strings.ToLower(format) {
	case "json", "yaml":
		return true
	default:
		return false
	}
}
