package skill

import "time"

// Frontmatter holds the parsed YAML frontmatter fields of a SKILL.md file.
// Field names match the Agent Skills open specification.
type Frontmatter struct {
	Name          string            `yaml:"name"`
	Description   string            `yaml:"description"`
	License       string            `yaml:"license,omitempty"`
	Compatibility string            `yaml:"compatibility,omitempty"`
	Metadata      map[string]string `yaml:"metadata,omitempty"`
	AllowedTools  string            `yaml:"allowed-tools,omitempty"`
}

// Skill is the fully parsed representation of a SKILL.md file.
type Skill struct {
	Frontmatter
	Body string // Markdown body after the frontmatter delimiters.
}

// InstalledSkill records a skill that has been installed locally.
type InstalledSkill struct {
	Name        string    `json:"name"`
	SourceRepo  string    `json:"source_repo"`
	SourcePath  string    `json:"source_path"`
	InstallPath string    `json:"install_path"`
	InstalledAt time.Time `json:"installed_at"`
}

// Standard install target base paths (relative to project root or $HOME for Personal* targets).
const (
	// Project-scoped targets — relative to the project root.
	TargetProjectCopilot = ".github/skills"
	TargetProjectAgents  = ".agents/skills"
	TargetProjectClaude  = ".claude/skills"

	// Personal targets — relative to $HOME.
	TargetPersonalCopilot = ".copilot/skills"
	TargetPersonalAgents  = ".agents/skills"
	TargetPersonalClaude  = ".claude/skills"
)
