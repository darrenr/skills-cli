package registry

import "time"

// Source describes where a skill comes from in a remote repository.
type Source struct {
	Repo string `json:"repo"` // e.g. "github/awesome-copilot"
	Path string `json:"path"` // path within the repo, e.g. "skills/git-commit"
	Ref  string `json:"ref"`  // branch, tag, or commit SHA
}

// SkillEntry is a lightweight registry record for a single skill.
// It contains enough information to display in list/search output and
// to locate the full SKILL.md for download.
type SkillEntry struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	License     string    `json:"license,omitempty"`
	Source      Source    `json:"source"`
}

// Registry is the top-level container for the bundled or cached skill registry.
type Registry struct {
	Version   string       `json:"version"`
	UpdatedAt time.Time    `json:"updated_at"`
	Skills    []SkillEntry `json:"skills"`
}
