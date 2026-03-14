package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	cacheTTL      = 24 * time.Hour
	cacheFileName = "registry.json"
	cacheDirName  = ".skills-cli"
)

// ErrNoRegistry is returned when no registry data can be found (embedded or cached).
var ErrNoRegistry = errors.New("no registry available")

// Loader manages loading the skill registry from an embedded JSON seed
// and a user-level cache that can be refreshed from remote sources.
type Loader struct {
	embedded []byte // embedded JSON from //go:embed (may be nil)
	cacheDir string // directory for cached registry.json
}

// NewLoader creates a Loader. embedded may be nil if no bundled registry
// has been generated yet (e.g. during early development). cacheDir defaults
// to ~/.skills-cli if empty.
func NewLoader(embedded []byte, cacheDir string) *Loader {
	if cacheDir == "" {
		home, _ := os.UserHomeDir()
		cacheDir = filepath.Join(home, cacheDirName)
	}
	return &Loader{embedded: embedded, cacheDir: cacheDir}
}

// Load returns the best available registry:
//  1. Cached file, if it exists, is within TTL, and is valid JSON.
//  2. Embedded JSON seed.
//  3. Stale cached file (past TTL but still readable).
//  4. ErrNoRegistry if nothing is available.
//
// A corrupt cache file (invalid JSON) is always returned as an error
// regardless of TTL, so the caller knows to delete or refresh it.
func (l *Loader) Load() (*Registry, error) {
	cachePath := filepath.Join(l.cacheDir, cacheFileName)

	r, fresh, cacheErr := l.loadCached(cachePath)
	if cacheErr == nil && fresh {
		return r, nil
	}
	// Corrupt cache — surface immediately, don't silently ignore.
	if isParseError(cacheErr) {
		return nil, cacheErr
	}

	if len(l.embedded) > 0 {
		return parseRegistry(l.embedded)
	}

	// Fall back to stale cache rather than nothing.
	if cacheErr == nil && !fresh {
		return r, nil
	}

	return nil, ErrNoRegistry
}

// Save writes the given registry to the cache file, creating the directory
// if needed.
func (l *Loader) Save(r *Registry) error {
	if err := os.MkdirAll(l.cacheDir, 0o700); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal registry: %w", err)
	}
	cachePath := filepath.Join(l.cacheDir, cacheFileName)
	return os.WriteFile(cachePath, data, 0o600)
}

// loadCached reads the registry from disk. Returns (registry, isFresh, error).
// isFresh is true when the file's modification time is within cacheTTL.
func (l *Loader) loadCached(path string) (*Registry, bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, false, err
	}
	fresh := time.Since(info.ModTime()) <= cacheTTL
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false, err
	}
	r, err := parseRegistry(data)
	if err != nil {
		return nil, false, err
	}
	return r, fresh, nil
}

// isParseError reports whether err originated from JSON parsing.
func isParseError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "parse registry JSON")
}

func parseRegistry(data []byte) (*Registry, error) {
	var r Registry
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("parse registry JSON: %w", err)
	}
	return &r, nil
}
