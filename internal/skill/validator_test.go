package skill_test

import (
	"strings"
	"testing"

	"github.com/darrenr/skills-cli/internal/skill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validSkill() *skill.Skill {
	return &skill.Skill{
		Frontmatter: skill.Frontmatter{
			Name:        "my-skill",
			Description: "A valid description.",
		},
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		modify     func(*skill.Skill)
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "valid skill",
			modify:  func(s *skill.Skill) {},
			wantErr: false,
		},
		// Name constraints
		{
			name:       "missing name",
			modify:     func(s *skill.Skill) { s.Name = "" },
			wantErr:    true,
			wantErrMsg: "name is required",
		},
		{
			name:       "name too long",
			modify:     func(s *skill.Skill) { s.Name = strings.Repeat("a", 65) },
			wantErr:    true,
			wantErrMsg: "name must be 64 characters or fewer",
		},
		{
			name:       "name exactly 64 chars is valid",
			modify:     func(s *skill.Skill) { s.Name = strings.Repeat("a", 64) },
			wantErr:    false,
		},
		{
			name:       "name with uppercase",
			modify:     func(s *skill.Skill) { s.Name = "MySkill" },
			wantErr:    true,
			wantErrMsg: "lowercase letters",
		},
		{
			name:       "name with leading hyphen",
			modify:     func(s *skill.Skill) { s.Name = "-my-skill" },
			wantErr:    true,
			wantErrMsg: "lowercase letters",
		},
		{
			name:       "name with trailing hyphen",
			modify:     func(s *skill.Skill) { s.Name = "my-skill-" },
			wantErr:    true,
			wantErrMsg: "lowercase letters",
		},
		{
			name:       "name with spaces",
			modify:     func(s *skill.Skill) { s.Name = "my skill" },
			wantErr:    true,
			wantErrMsg: "lowercase letters",
		},
		{
			name:    "single char name is valid",
			modify:  func(s *skill.Skill) { s.Name = "a" },
			wantErr: false,
		},
		{
			name:    "name with digits is valid",
			modify:  func(s *skill.Skill) { s.Name = "skill2" },
			wantErr: false,
		},
		{
			name:    "name with internal hyphens is valid",
			modify:  func(s *skill.Skill) { s.Name = "go-mcp-server-generator" },
			wantErr: false,
		},
		// Description constraints
		{
			name:       "missing description",
			modify:     func(s *skill.Skill) { s.Description = "" },
			wantErr:    true,
			wantErrMsg: "description is required",
		},
		{
			name:       "description too long",
			modify:     func(s *skill.Skill) { s.Description = strings.Repeat("x", 1025) },
			wantErr:    true,
			wantErrMsg: "description must be 1024 characters or fewer",
		},
		{
			name:    "description exactly 1024 chars is valid",
			modify:  func(s *skill.Skill) { s.Description = strings.Repeat("x", 1024) },
			wantErr: false,
		},
		// Compatibility constraints
		{
			name:       "compatibility too long",
			modify:     func(s *skill.Skill) { s.Compatibility = strings.Repeat("x", 501) },
			wantErr:    true,
			wantErrMsg: "compatibility must be 500 characters or fewer",
		},
		{
			name:    "compatibility exactly 500 chars is valid",
			modify:  func(s *skill.Skill) { s.Compatibility = strings.Repeat("x", 500) },
			wantErr: false,
		},
		// Multiple errors
		{
			name: "multiple errors reported",
			modify: func(s *skill.Skill) {
				s.Name = ""
				s.Description = ""
			},
			wantErr:    true,
			wantErrMsg: "name is required; description is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := validSkill()
			tt.modify(s)
			err := skill.Validate(s)
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrMsg != "" {
					for _, msg := range strings.Split(tt.wantErrMsg, "; ") {
						assert.Contains(t, err.Error(), msg)
					}
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateNameMatchesDir(t *testing.T) {
	t.Parallel()

	s := &skill.Skill{Frontmatter: skill.Frontmatter{Name: "my-skill"}}

	assert.NoError(t, skill.ValidateNameMatchesDir(s, "my-skill"))
	assert.Error(t, skill.ValidateNameMatchesDir(s, "other-skill"))
	assert.Error(t, skill.ValidateNameMatchesDir(s, "My-Skill"))
}
