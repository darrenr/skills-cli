package installer

import (
	"context"
	"fmt"
	"time"

	"github.com/darrenr/skills-cli/internal/registry"
	"github.com/darrenr/skills-cli/internal/skill"
	"github.com/darrenr/skills-cli/internal/source"
)

// UpdateOptions controls the behaviour of Update.
type UpdateOptions struct {
	// DryRun reports what would change without writing files.
	DryRun bool
}

// Update re-downloads a skill from its source. The skill must already be in
// the manifest. Returns the install path.
func Update(ctx context.Context, name string, opts UpdateOptions, fetcher *source.GitHubFetcher, reg *registry.Registry, manifest *Manifest) (string, error) {
	installed := manifest.Find(name)
	if installed == nil {
		return "", fmt.Errorf("skill %q is not installed", name)
	}

	// Find the current registry entry for this skill.
	results := registry.Search(reg, registry.SearchOptions{Query: name, Limit: 10})
	var entry *registry.SkillEntry
	for i := range results {
		if results[i].Name == name {
			entry = &results[i]
			break
		}
	}
	if entry == nil {
		return "", fmt.Errorf("skill %q not found in registry", name)
	}

	if opts.DryRun {
		return installed.InstallPath, nil
	}

	written, err := fetcher.DownloadSkill(ctx, entry.Source.Repo, entry.Source.Ref, entry.Source.Path, installed.InstallPath, nil)
	if err != nil {
		return "", fmt.Errorf("download %s: %w", name, err)
	}
	if len(written) == 0 {
		return "", fmt.Errorf("no files written for skill %q", name)
	}

	manifest.Upsert(skill.InstalledSkill{
		Name:        name,
		SourceRepo:  entry.Source.Repo,
		SourcePath:  entry.Source.Path,
		InstallPath: installed.InstallPath,
		InstalledAt: time.Now(),
	})
	if err := manifest.Save(); err != nil {
		return installed.InstallPath, fmt.Errorf("save manifest: %w", err)
	}

	return installed.InstallPath, nil
}
