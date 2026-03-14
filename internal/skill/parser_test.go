package skill_test

import (
	"testing"

	"github.com/darrenr/skills-cli/internal/skill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      string
		wantErr    error
		wantErrMsg string
		wantName   string
		wantDesc   string
		wantLic    string
		wantCompat string
		wantMeta   map[string]string
		wantBody   string
	}{
		{
			name: "valid full skill",
			input: `---
name: my-skill
description: A test skill.
license: Apache-2.0
compatibility: Works everywhere.
metadata:
  author: testauthor
  version: "1.0"
---
# My Skill

Body content here.`,
			wantName:   "my-skill",
			wantDesc:   "A test skill.",
			wantLic:    "Apache-2.0",
			wantCompat: "Works everywhere.",
			wantMeta:   map[string]string{"author": "testauthor", "version": "1.0"},
			wantBody:   "# My Skill\n\nBody content here.",
		},
		{
			name: "minimal valid skill",
			input: `---
name: minimal
description: Minimal description.
---
# Minimal`,
			wantName: "minimal",
			wantDesc: "Minimal description.",
			wantBody: "# Minimal",
		},
		{
			name: "empty body",
			input: `---
name: no-body
description: Skill with no body.
---`,
			wantName: "no-body",
			wantDesc: "Skill with no body.",
			wantBody: "",
		},
		{
			name: "body with multiple sections",
			input: `---
name: rich-skill
description: Has a rich body.
---
# Rich Skill

## Usage

Do things.

## Examples

- Example one`,
			wantName: "rich-skill",
			wantDesc: "Has a rich body.",
			wantBody: "# Rich Skill\n\n## Usage\n\nDo things.\n\n## Examples\n\n- Example one",
		},
		{
			name: "no frontmatter",
			input: `# Just markdown

No frontmatter here.`,
			wantErr: skill.ErrNoFrontmatter,
		},
		{
			name: "unclosed frontmatter",
			input: `---
name: broken
description: Missing closing delimiter.
`,
			wantErr: skill.ErrNoFrontmatter,
		},
		{
			name: "invalid YAML",
			input: `---
name: [invalid
---
# Body`,
			wantErrMsg: "invalid frontmatter YAML",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := skill.Parse([]byte(tt.input))

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			if tt.wantErrMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMsg)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantName, got.Name)
			assert.Equal(t, tt.wantDesc, got.Description)
			assert.Equal(t, tt.wantLic, got.License)
			assert.Equal(t, tt.wantCompat, got.Compatibility)
			assert.Equal(t, tt.wantMeta, got.Metadata)
			assert.Equal(t, tt.wantBody, got.Body)
		})
	}
}

func TestParseFile(t *testing.T) {
	t.Parallel()

	t.Run("reads and parses valid file", func(t *testing.T) {
		t.Parallel()
		got, err := skill.ParseFile("testdata/valid.md")
		require.NoError(t, err)
		assert.Equal(t, "test-skill", got.Name)
		assert.Equal(t, "A test skill for unit testing the parser.", got.Description)
		assert.Equal(t, "Apache-2.0", got.License)
		assert.NotEmpty(t, got.Body)
	})

	t.Run("returns error for missing file", func(t *testing.T) {
		t.Parallel()
		_, err := skill.ParseFile("testdata/does-not-exist.md")
		require.Error(t, err)
	})
}
