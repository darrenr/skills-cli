package skill

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// ErrNoFrontmatter is returned when no valid YAML frontmatter block is found.
var ErrNoFrontmatter = errors.New("no YAML frontmatter found")

// Parse reads SKILL.md content and returns the fully parsed Skill.
func Parse(content []byte) (*Skill, error) {
	fmBytes, body, err := splitFrontmatter(content)
	if err != nil {
		return nil, err
	}

	var front Frontmatter
	if err := yaml.Unmarshal(fmBytes, &front); err != nil {
		return nil, fmt.Errorf("invalid frontmatter YAML: %w", err)
	}

	return &Skill{
		Frontmatter: front,
		Body:        body,
	}, nil
}

// ParseFile reads a SKILL.md from disk and returns the parsed Skill.
func ParseFile(path string) (*Skill, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(content)
}

// splitFrontmatter splits SKILL.md content into YAML frontmatter bytes and
// the trimmed markdown body. The content must begin with a line containing
// only "---".
func splitFrontmatter(content []byte) (fm []byte, body string, err error) {
	lines := strings.Split(string(content), "\n")

	if len(lines) == 0 || strings.TrimRight(lines[0], "\r") != "---" {
		return nil, "", ErrNoFrontmatter
	}

	for i := 1; i < len(lines); i++ {
		if strings.TrimRight(lines[i], "\r") == "---" {
			frontmatter := strings.Join(lines[1:i], "\n")
			body := strings.TrimSpace(strings.Join(lines[i+1:], "\n"))
			return []byte(frontmatter), body, nil
		}
	}

	return nil, "", ErrNoFrontmatter
}
