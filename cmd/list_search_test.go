package cmd

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/darrenr/skills-cli/internal/installer"
	"github.com/darrenr/skills-cli/internal/skill"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	orig := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	defer func() {
		os.Stdout = orig
	}()

	fn()

	require.NoError(t, w.Close())
	data, err := io.ReadAll(r)
	require.NoError(t, err)
	require.NoError(t, r.Close())
	return string(data)
}

func TestRunSearch_JSONOutputIsValid(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("category", "", "")
	cmd.Flags().String("source", "", "")
	cmd.Flags().Int("limit", 1, "")
	require.NoError(t, cmd.Flags().Set("limit", "1"))

	prevOutput := viper.GetString("output")
	viper.Set("output", "json")
	defer viper.Set("output", prevOutput)

	out := captureStdout(t, func() {
		err := runSearch(cmd, []string{"commit"})
		require.NoError(t, err)
	})

	var entries []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &entries))
	require.NotEmpty(t, entries)
	assert.NotEmpty(t, entries[0]["name"])
}

func TestRunList_JSONEmptyOutputIsValid(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("category", "", "")
	cmd.Flags().String("source", "", "")
	cmd.Flags().Bool("installed", false, "")
	require.NoError(t, cmd.Flags().Set("category", "does-not-exist"))

	prevOutput := viper.GetString("output")
	viper.Set("output", "json")
	defer viper.Set("output", prevOutput)

	out := captureStdout(t, func() {
		err := runList(cmd, nil)
		require.NoError(t, err)
	})

	var entries []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &entries))
	assert.Len(t, entries, 0)
}

func TestRunList_InstalledInCurrentProject_PrunesStaleAndExcludesOutsideProject(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	projectDir := t.TempDir()
	origWD, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(projectDir))
	t.Cleanup(func() {
		_ = os.Chdir(origWD)
	})

	projectInstalledPath := filepath.Join(projectDir, ".github", "skills", "git-commit")
	require.NoError(t, os.MkdirAll(projectInstalledPath, 0o755))

	outsideProjectPath := filepath.Join(homeDir, ".copilot", "skills", "create-readme")
	require.NoError(t, os.MkdirAll(outsideProjectPath, 0o755))

	staleProjectPath := filepath.Join(projectDir, ".github", "skills", "gh-cli")

	manifestPath, err := installer.DefaultManifestPath()
	require.NoError(t, err)
	manifest, err := installer.LoadManifest(manifestPath)
	require.NoError(t, err)
	manifest.Upsert(skill.InstalledSkill{Name: "git-commit", InstallPath: projectInstalledPath, InstalledAt: time.Now()})
	manifest.Upsert(skill.InstalledSkill{Name: "create-readme", InstallPath: outsideProjectPath, InstalledAt: time.Now()})
	manifest.Upsert(skill.InstalledSkill{Name: "gh-cli", InstallPath: staleProjectPath, InstalledAt: time.Now()})
	require.NoError(t, manifest.Save())

	cmd := &cobra.Command{}
	cmd.Flags().String("category", "", "")
	cmd.Flags().String("source", "", "")
	cmd.Flags().Bool("installed", false, "")
	require.NoError(t, cmd.Flags().Set("installed", "true"))

	prevOutput := viper.GetString("output")
	viper.Set("output", "json")
	defer viper.Set("output", prevOutput)

	out := captureStdout(t, func() {
		err := runList(cmd, nil)
		require.NoError(t, err)
	})

	var entries []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &entries))

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		name, ok := e["name"].(string)
		require.True(t, ok)
		names = append(names, name)
	}

	assert.Contains(t, names, "git-commit")
	assert.NotContains(t, names, "create-readme")
	assert.NotContains(t, names, "gh-cli")

	reloaded, err := installer.LoadManifest(manifestPath)
	require.NoError(t, err)
	assert.Nil(t, reloaded.Find("gh-cli"))
}

func TestRunList_InstalledInCurrentProject_FindsSkillFromProjectFolderWhenManifestMissing(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	projectDir := t.TempDir()
	origWD, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(projectDir))
	t.Cleanup(func() {
		_ = os.Chdir(origWD)
	})

	require.NoError(t, os.MkdirAll(filepath.Join(projectDir, ".github", "skills", "git-commit"), 0o755))

	manifestPath, err := installer.DefaultManifestPath()
	require.NoError(t, err)
	manifest, err := installer.LoadManifest(manifestPath)
	require.NoError(t, err)
	require.NoError(t, manifest.Save())

	cmd := &cobra.Command{}
	cmd.Flags().String("category", "", "")
	cmd.Flags().String("source", "", "")
	cmd.Flags().Bool("installed", false, "")
	require.NoError(t, cmd.Flags().Set("installed", "true"))

	prevOutput := viper.GetString("output")
	viper.Set("output", "json")
	defer viper.Set("output", prevOutput)

	out := captureStdout(t, func() {
		err := runList(cmd, nil)
		require.NoError(t, err)
	})

	var entries []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &entries))

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		name, ok := e["name"].(string)
		require.True(t, ok)
		names = append(names, name)
	}

	assert.Contains(t, names, "git-commit")
}
