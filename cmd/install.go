package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/darrenr/skills-cli/internal/installer"
	"github.com/darrenr/skills-cli/internal/registry"
	"github.com/darrenr/skills-cli/internal/source"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install <name> [name...]",
	Short: "Install one or more skills into your project",
	Long: `Download and install skills from the registry into your project.

By default skills are installed to .github/skills/<name>/ (Copilot project scope).
Use --target to change the install location.

Examples:
  skills-cli install conventional-commit
  skills-cli install git-commit go-test --target project-agents
  skills-cli install my-skill --force`,
	Args: cobra.MinimumNArgs(1),
	RunE: runInstall,
}

func init() {
	installCmd.Flags().StringP("target", "t", "project-copilot", "install target: project-copilot, project-agents, project-claude, global-copilot, global-agents, global-claude")
	installCmd.Flags().Bool("force", false, "overwrite an existing installation")
	rootCmd.AddCommand(installCmd)
}

func runInstall(cmd *cobra.Command, args []string) error {
	target, _ := cmd.Flags().GetString("target")
	force, _ := cmd.Flags().GetBool("force")

	targetDir, err := resolveTargetDir(target)
	if err != nil {
		return err
	}

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

	token := os.Getenv("GITHUB_TOKEN")
	fetcher := source.NewGitHubFetcher(token)

	for _, name := range args {
		results := registry.Search(r, registry.SearchOptions{Query: name, Limit: 10})
		entry, found := findExact(results, name)
		if !found {
			fmt.Fprintf(os.Stderr, "skill %q not found in registry\n", name)
			continue
		}

		fmt.Printf("Installing %s...\n", name)
		path, err := installer.Install(context.Background(), entry,
			installer.InstallOptions{TargetDir: targetDir, Force: force},
			fetcher, manifest)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			continue
		}
		fmt.Printf("  ✓ installed to %s\n", path)
	}
	return nil
}

// resolveTargetDir maps a target name to a base directory path.
func resolveTargetDir(target string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	switch target {
	case "project-copilot":
		return ".github/skills", nil
	case "project-agents":
		return ".agents/skills", nil
	case "project-claude":
		return ".claude/skills", nil
	case "global-copilot":
		return filepath.Join(home, ".copilot", "skills"), nil
	case "global-agents":
		return filepath.Join(home, ".agents", "skills"), nil
	case "global-claude":
		return filepath.Join(home, ".claude", "skills"), nil
	default:
		return "", fmt.Errorf("unknown target %q: use project-copilot, project-agents, project-claude, global-copilot, global-agents, or global-claude", target)
	}
}
