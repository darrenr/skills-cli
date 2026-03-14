package cmd

import (
	"encoding/json"
	"io"
	"os"
	"testing"

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
