package installer

import (
	"fmt"
	"os"
)

// RemoveOptions controls the behaviour of Remove.
type RemoveOptions struct {
	// Force skips the "not installed" error and still attempts directory removal.
	Force bool
}

// Remove deletes the skill directory and removes it from the manifest.
func Remove(name string, opts RemoveOptions, manifest *Manifest) error {
	installed := manifest.Find(name)
	if installed == nil {
		if opts.Force {
			return nil
		}
		return fmt.Errorf("skill %q is not installed", name)
	}

	if err := os.RemoveAll(installed.InstallPath); err != nil {
		return fmt.Errorf("remove %s: %w", installed.InstallPath, err)
	}

	manifest.Remove(name)
	if err := manifest.Save(); err != nil {
		return fmt.Errorf("save manifest: %w", err)
	}

	return nil
}
