package registry_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/darrenr/skills-cli/internal/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleRegistry() *registry.Registry {
	return &registry.Registry{
		Version:   "1",
		UpdatedAt: time.Now(),
		Skills: []registry.SkillEntry{
			{Name: "git-commit", Description: "Write great commits.", Category: "git", Source: registry.Source{Repo: "github/awesome-copilot", Path: "skills/git-commit", Ref: "main"}},
			{Name: "go-test", Description: "Write Go tests.", Category: "golang", Source: registry.Source{Repo: "github/awesome-copilot", Path: "skills/go-test", Ref: "main"}},
		},
	}
}

func marshalRegistry(t *testing.T, r *registry.Registry) []byte {
	t.Helper()
	data, err := json.Marshal(r)
	require.NoError(t, err)
	return data
}

func TestLoader_LoadFromEmbedded(t *testing.T) {
	t.Parallel()
	embedded := marshalRegistry(t, sampleRegistry())
	l := registry.NewLoader(embedded, t.TempDir())
	r, err := l.Load()
	require.NoError(t, err)
	assert.Len(t, r.Skills, 2)
	assert.Equal(t, "git-commit", r.Skills[0].Name)
}

func TestLoader_LoadFromCache(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// Write a fresh cache file.
	reg := sampleRegistry()
	data := marshalRegistry(t, reg)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "registry.json"), data, 0o600))

	l := registry.NewLoader(nil, dir)
	r, err := l.Load()
	require.NoError(t, err)
	assert.Len(t, r.Skills, 2)
}

func TestLoader_StaleCache_FallsBackToEmbedded(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// Write a stale cache (mtime 48h ago).
	stalePath := filepath.Join(dir, "registry.json")
	staleReg := &registry.Registry{Version: "stale", Skills: []registry.SkillEntry{{Name: "old-skill", Description: "Old."}}}
	data := marshalRegistry(t, staleReg)
	require.NoError(t, os.WriteFile(stalePath, data, 0o600))
	staleTime := time.Now().Add(-48 * time.Hour)
	require.NoError(t, os.Chtimes(stalePath, staleTime, staleTime))

	// Embedded has fresh data.
	embedded := marshalRegistry(t, sampleRegistry())
	l := registry.NewLoader(embedded, dir)
	r, err := l.Load()
	require.NoError(t, err)
	// Should load embedded, not stale cache.
	assert.Equal(t, "1", r.Version)
}

func TestLoader_StaleCache_UsedWhenNoEmbedded(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	stalePath := filepath.Join(dir, "registry.json")
	staleReg := &registry.Registry{Version: "stale", Skills: []registry.SkillEntry{{Name: "old-skill", Description: "Old."}}}
	data := marshalRegistry(t, staleReg)
	require.NoError(t, os.WriteFile(stalePath, data, 0o600))
	staleTime := time.Now().Add(-48 * time.Hour)
	require.NoError(t, os.Chtimes(stalePath, staleTime, staleTime))

	l := registry.NewLoader(nil, dir)
	r, err := l.Load()
	require.NoError(t, err)
	assert.Equal(t, "stale", r.Version)
}

func TestLoader_NoRegistryAvailable(t *testing.T) {
	t.Parallel()
	l := registry.NewLoader(nil, t.TempDir())
	_, err := l.Load()
	require.ErrorIs(t, err, registry.ErrNoRegistry)
}

func TestLoader_Save(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	l := registry.NewLoader(nil, dir)
	require.NoError(t, l.Save(sampleRegistry()))

	// Load it back.
	r, err := l.Load()
	require.NoError(t, err)
	assert.Len(t, r.Skills, 2)
}

func TestLoader_InvalidJSON(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "registry.json"), []byte("not json"), 0o600))
	l := registry.NewLoader(nil, dir)
	_, err := l.Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse registry JSON")
}
