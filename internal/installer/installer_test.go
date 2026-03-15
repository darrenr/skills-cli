package installer_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/darrenr/skills-cli/internal/installer"
	"github.com/darrenr/skills-cli/internal/registry"
	"github.com/darrenr/skills-cli/internal/skill"
	"github.com/darrenr/skills-cli/internal/source"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func skillMDContent() string {
	return "---\nname: test-skill\ndescription: A test skill.\n---\n# Test"
}

func newSkillServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/owner/repo/contents/skills/test-skill" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]map[string]string{{
				"type": "file",
				"path": "skills/test-skill/SKILL.md",
			}})
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(skillMDContent()))
	}))
}

func newMissingServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/owner/repo/contents/skills/test-skill" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]map[string]string{})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func testEntry() registry.SkillEntry {
	return registry.SkillEntry{
		Name:        "test-skill",
		Description: "A test skill.",
		Source:      registry.Source{Repo: "owner/repo", Path: "skills/test-skill", Ref: "main"},
	}
}

func testFetcher(srv *httptest.Server) *source.GitHubFetcher {
	f := source.NewGitHubFetcher("")
	f.SetClient(srv.Client())
	f.SetBaseURL(srv.URL)
	f.SetAPIBaseURL(srv.URL)
	return f
}

func tempManifest(t *testing.T) *installer.Manifest {
	t.Helper()
	path := filepath.Join(t.TempDir(), "manifest.json")
	m, err := installer.LoadManifest(path)
	require.NoError(t, err)
	return m
}

// --- Manifest tests ---

func TestManifest_EmptyWhenFileAbsent(t *testing.T) {
	t.Parallel()
	m := tempManifest(t)
	assert.Empty(t, m.Skills)
}

func TestManifest_UpsertAndFind(t *testing.T) {
	t.Parallel()
	m := tempManifest(t)
	m.Upsert(installedSkill("my-skill", "/path/to/my-skill"))
	found := m.Find("my-skill")
	require.NotNil(t, found)
	assert.Equal(t, "/path/to/my-skill", found.InstallPath)
}

func TestManifest_UpsertOverwrites(t *testing.T) {
	t.Parallel()
	m := tempManifest(t)
	m.Upsert(installedSkill("my-skill", "/old/path"))
	m.Upsert(installedSkill("my-skill", "/new/path"))
	assert.Len(t, m.Skills, 1)
	assert.Equal(t, "/new/path", m.Find("my-skill").InstallPath)
}

func TestManifest_Remove(t *testing.T) {
	t.Parallel()
	m := tempManifest(t)
	m.Upsert(installedSkill("my-skill", "/path"))
	assert.True(t, m.Remove("my-skill"))
	assert.Nil(t, m.Find("my-skill"))
	assert.False(t, m.Remove("my-skill")) // second remove returns false
}

func TestManifest_SaveAndReload(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")

	m1, err := installer.LoadManifest(path)
	require.NoError(t, err)
	m1.Upsert(installedSkill("skill-a", "/a"))
	require.NoError(t, m1.Save())

	m2, err := installer.LoadManifest(path)
	require.NoError(t, err)
	require.Len(t, m2.Skills, 1)
	assert.Equal(t, "skill-a", m2.Skills[0].Name)
}

// --- Install tests ---

func TestInstall_Success(t *testing.T) {
	t.Parallel()
	srv := newSkillServer(t)
	defer srv.Close()

	dir := t.TempDir()
	m := tempManifest(t)
	path, err := installer.Install(context.Background(), testEntry(),
		installer.InstallOptions{TargetDir: dir}, testFetcher(srv), m)
	require.NoError(t, err)
	assert.DirExists(t, path)
	assert.FileExists(t, filepath.Join(path, "SKILL.md"))
	assert.NotNil(t, m.Find("test-skill"))
}

func TestInstall_AlreadyInstalledWithoutForce(t *testing.T) {
	t.Parallel()
	srv := newSkillServer(t)
	defer srv.Close()

	dir := t.TempDir()
	m := tempManifest(t)
	_, err := installer.Install(context.Background(), testEntry(),
		installer.InstallOptions{TargetDir: dir}, testFetcher(srv), m)
	require.NoError(t, err)

	// Second install without --force should fail.
	_, err = installer.Install(context.Background(), testEntry(),
		installer.InstallOptions{TargetDir: dir}, testFetcher(srv), m)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already installed")
}

func TestInstall_StaleManifestEntry_AllowsInstall(t *testing.T) {
	t.Parallel()
	srv := newSkillServer(t)
	defer srv.Close()

	dir := t.TempDir()
	m := tempManifest(t)

	// Simulate a stale manifest entry after manual directory deletion.
	stalePath := filepath.Join(dir, "test-skill")
	m.Upsert(installedSkill("test-skill", stalePath))
	require.NoError(t, m.Save())

	before := m.Find("test-skill")
	require.NotNil(t, before)
	assert.Equal(t, stalePath, before.InstallPath)

	path, err := installer.Install(context.Background(), testEntry(),
		installer.InstallOptions{TargetDir: dir}, testFetcher(srv), m)
	require.NoError(t, err)
	assert.Equal(t, stalePath, path)
	assert.FileExists(t, filepath.Join(path, "SKILL.md"))

	after := m.Find("test-skill")
	require.NotNil(t, after)
	assert.Equal(t, stalePath, after.InstallPath)
}

func TestInstall_ForceOverwrites(t *testing.T) {
	t.Parallel()
	srv := newSkillServer(t)
	defer srv.Close()

	dir := t.TempDir()
	m := tempManifest(t)
	_, err := installer.Install(context.Background(), testEntry(),
		installer.InstallOptions{TargetDir: dir}, testFetcher(srv), m)
	require.NoError(t, err)

	_, err = installer.Install(context.Background(), testEntry(),
		installer.InstallOptions{TargetDir: dir, Force: true}, testFetcher(srv), m)
	require.NoError(t, err)
}

func TestInstall_FetchError(t *testing.T) {
	t.Parallel()
	srv := newMissingServer(t)
	defer srv.Close()

	dir := t.TempDir()
	m := tempManifest(t)
	_, err := installer.Install(context.Background(), testEntry(),
		installer.InstallOptions{TargetDir: dir}, testFetcher(srv), m)
	require.Error(t, err)
}

// --- Update tests ---

func TestUpdate_Success(t *testing.T) {
	t.Parallel()
	srv := newSkillServer(t)
	defer srv.Close()

	dir := t.TempDir()
	m := tempManifest(t)
	reg := &registry.Registry{Skills: []registry.SkillEntry{testEntry()}}

	_, err := installer.Install(context.Background(), testEntry(),
		installer.InstallOptions{TargetDir: dir}, testFetcher(srv), m)
	require.NoError(t, err)

	path, err := installer.Update(context.Background(), "test-skill",
		installer.UpdateOptions{}, testFetcher(srv), reg, m)
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(path, "SKILL.md"))
}

func TestUpdate_NotInstalled(t *testing.T) {
	t.Parallel()
	srv := newSkillServer(t)
	defer srv.Close()

	m := tempManifest(t)
	reg := &registry.Registry{Skills: []registry.SkillEntry{testEntry()}}

	_, err := installer.Update(context.Background(), "test-skill",
		installer.UpdateOptions{}, testFetcher(srv), reg, m)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not installed")
}

func TestUpdate_DryRun(t *testing.T) {
	t.Parallel()
	srv := newSkillServer(t)
	defer srv.Close()

	dir := t.TempDir()
	m := tempManifest(t)
	reg := &registry.Registry{Skills: []registry.SkillEntry{testEntry()}}

	_, err := installer.Install(context.Background(), testEntry(),
		installer.InstallOptions{TargetDir: dir}, testFetcher(srv), m)
	require.NoError(t, err)

	before := m.InstalledAt("test-skill")
	_, err = installer.Update(context.Background(), "test-skill",
		installer.UpdateOptions{DryRun: true}, testFetcher(srv), reg, m)
	require.NoError(t, err)
	// InstalledAt should not change on dry run.
	assert.Equal(t, before, m.InstalledAt("test-skill"))
}

// --- Remove tests ---

func TestRemove_Success(t *testing.T) {
	t.Parallel()
	srv := newSkillServer(t)
	defer srv.Close()

	dir := t.TempDir()
	m := tempManifest(t)

	path, err := installer.Install(context.Background(), testEntry(),
		installer.InstallOptions{TargetDir: dir}, testFetcher(srv), m)
	require.NoError(t, err)

	require.NoError(t, installer.Remove("test-skill", installer.RemoveOptions{}, m))
	assert.NoDirExists(t, path)
	assert.Nil(t, m.Find("test-skill"))
}

func TestRemove_NotInstalled(t *testing.T) {
	t.Parallel()
	m := tempManifest(t)
	err := installer.Remove("unknown-skill", installer.RemoveOptions{}, m)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not installed")
}

func TestRemove_ForceOnMissing(t *testing.T) {
	t.Parallel()
	m := tempManifest(t)
	require.NoError(t, installer.Remove("unknown-skill", installer.RemoveOptions{Force: true}, m))
}

func TestRemove_CleansDirectory(t *testing.T) {
	t.Parallel()
	srv := newSkillServer(t)
	defer srv.Close()

	dir := t.TempDir()
	m := tempManifest(t)
	installPath, err := installer.Install(context.Background(), testEntry(),
		installer.InstallOptions{TargetDir: dir}, testFetcher(srv), m)
	require.NoError(t, err)

	// Write an extra file to ensure the whole directory is removed.
	require.NoError(t, os.WriteFile(filepath.Join(installPath, "extra.txt"), []byte("extra"), 0o644))

	require.NoError(t, installer.Remove("test-skill", installer.RemoveOptions{}, m))
	assert.NoDirExists(t, installPath)
}

// --- Extra manifest coverage ---

func TestDefaultManifestPath(t *testing.T) {
	t.Parallel()
	path, err := installer.DefaultManifestPath()
	require.NoError(t, err)
	assert.NotEmpty(t, path)
	assert.Contains(t, path, "manifest.json")
}

func TestManifest_InstalledAt_NotFound(t *testing.T) {
	t.Parallel()
	m := tempManifest(t)
	assert.Equal(t, time.Time{}, m.InstalledAt("nonexistent"))
}

func TestManifest_FindExistingInCurrentProject_AllowsDotDotPrefixDirectoryName(t *testing.T) {
	projectDir := t.TempDir()
	origWD, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(projectDir))
	t.Cleanup(func() {
		_ = os.Chdir(origWD)
	})

	installPath := filepath.Join(projectDir, "..cache", "edge-skill")
	require.NoError(t, os.MkdirAll(installPath, 0o755))

	m := tempManifest(t)
	m.Upsert(installedSkill("edge-skill", installPath))

	found, err := m.FindExistingInCurrentProject("edge-skill")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, installPath, found.InstallPath)
}

func TestManifest_FindExistingInCurrentProject_ExcludesOutsideProject(t *testing.T) {
	projectDir := t.TempDir()
	origWD, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(projectDir))
	t.Cleanup(func() {
		_ = os.Chdir(origWD)
	})

	outsidePath := filepath.Join(t.TempDir(), "outside-skill")
	require.NoError(t, os.MkdirAll(outsidePath, 0o755))

	m := tempManifest(t)
	m.Upsert(installedSkill("outside-skill", outsidePath))

	found, err := m.FindExistingInCurrentProject("outside-skill")
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestLoadManifest_InvalidJSON(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")
	require.NoError(t, os.WriteFile(path, []byte("not valid json"), 0o600))
	_, err := installer.LoadManifest(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse manifest")
}

// --- helpers ---

func installedSkill(name, path string) skill.InstalledSkill {
	return skill.InstalledSkill{
		Name:        name,
		InstallPath: path,
		InstalledAt: time.Now(),
	}
}
