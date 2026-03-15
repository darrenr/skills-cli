package installer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/darrenr/skills-cli/internal/registry"
	"github.com/darrenr/skills-cli/internal/skill"
	"github.com/darrenr/skills-cli/internal/source"
)

// InstallOptions controls the behaviour of Install.
type InstallOptions struct {
	// TargetDir is the base directory to install into (e.g. ".github/skills").
	// The skill will be placed at TargetDir/<skill-name>/.
	TargetDir string
	// Force overwrites an existing installation without error.
	Force bool
}

// Install downloads a skill from its source and writes it to TargetDir/<name>/.
// It updates the manifest on success.
func Install(ctx context.Context, entry registry.SkillEntry, opts InstallOptions, fetcher *source.GitHubFetcher, manifest *Manifest) (string, error) {
	destDir := filepath.Join(opts.TargetDir, entry.Name)

	existing := manifest.Find(entry.Name)
	if existing != nil {
		if _, err := os.Stat(existing.InstallPath); err != nil {
			if os.IsNotExist(err) {
				// Self-heal stale manifest entries when users manually delete skill folders.
				manifest.Remove(entry.Name)
				if err := manifest.Save(); err != nil {
					warnErr := fmt.Errorf("save manifest: %w", err)
					fmt.Fprintf(os.Stderr, "warning: %v\n", warnErr)
				}
				existing = nil
			} else {
				return "", fmt.Errorf("check existing install path %s: %w", existing.InstallPath, err)
			}
		}
	}
	if existing != nil && !opts.Force {
		return "", fmt.Errorf("skill %q is already installed at %s (use --force to overwrite)", entry.Name, existing.InstallPath)
	}

	written, err := fetcher.DownloadSkill(ctx, entry.Source.Repo, entry.Source.Ref, entry.Source.Path, destDir, nil)
	if err != nil {
		return "", fmt.Errorf("download %s: %w", entry.Name, err)
	}
	if len(written) == 0 {
		return "", fmt.Errorf("no files written for skill %q", entry.Name)
	}

	manifest.Upsert(skill.InstalledSkill{
		Name:        entry.Name,
		SourceRepo:  entry.Source.Repo,
		SourcePath:  entry.Source.Path,
		InstallPath: destDir,
		InstalledAt: time.Now(),
	})
	if err := manifest.Save(); err != nil {
		// Non-fatal: files are written, just warn.
		warnErr := fmt.Errorf("save manifest: %w", err)
		fmt.Fprintf(os.Stderr, "warning: %v\n", warnErr)
	}

	return destDir, nil
}
