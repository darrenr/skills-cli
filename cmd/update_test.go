package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/darrenr/skills-cli/internal/installer"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunUpdate_DryRun_DiscoversProjectSkillsWhenManifestMissing(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	projectDir := t.TempDir()
	origWD, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(projectDir))
	t.Cleanup(func() {
		_ = os.Chdir(origWD)
	})

	installPath := filepath.Join(projectDir, ".github", "skills", "git-commit")
	require.NoError(t, os.MkdirAll(installPath, 0o755))

	manifestPath, err := installer.DefaultManifestPath()
	require.NoError(t, err)

	cmd := &cobra.Command{}
	cmd.Flags().Bool("dry-run", false, "")
	require.NoError(t, cmd.Flags().Set("dry-run", "true"))

	out := captureStdout(t, func() {
		err := runUpdate(cmd, nil)
		require.NoError(t, err)
	})

	assert.Contains(t, out, "Would update: git-commit")
	assert.NotContains(t, out, "No skills are installed")

	reloaded, err := installer.LoadManifest(manifestPath)
	require.NoError(t, err)
	entry := reloaded.Find("git-commit")
	require.NotNil(t, entry)
	assert.Equal(t, filepath.Join(".github", "skills", "git-commit"), entry.InstallPath)
}
