package installer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/darrenr/skills-cli/internal/skill"
)

const manifestFileName = "manifest.json"

// Manifest tracks all locally installed skills.
type Manifest struct {
	Skills []skill.InstalledSkill `json:"skills"`
	path   string
}

// LoadManifest reads the manifest from path. Returns an empty manifest if the
// file does not exist yet.
func LoadManifest(path string) (*Manifest, error) {
	m := &Manifest{path: path}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return m, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	if err := json.Unmarshal(data, m); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	m.path = path
	return m, nil
}

// DefaultManifestPath returns ~/.skills-cli/manifest.json.
func DefaultManifestPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".skills-cli", manifestFileName), nil
}

// Save writes the manifest to disk.
func (m *Manifest) Save() error {
	if err := os.MkdirAll(filepath.Dir(m.path), 0o700); err != nil {
		return fmt.Errorf("create manifest dir: %w", err)
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}
	return os.WriteFile(m.path, data, 0o600)
}

// Find returns the installed skill with the given name, or nil if not found.
func (m *Manifest) Find(name string) *skill.InstalledSkill {
	for i := range m.Skills {
		if m.Skills[i].Name == name {
			return &m.Skills[i]
		}
	}
	return nil
}

// Upsert adds or replaces the entry for s.Name.
func (m *Manifest) Upsert(s skill.InstalledSkill) {
	for i := range m.Skills {
		if m.Skills[i].Name == s.Name {
			m.Skills[i] = s
			return
		}
	}
	m.Skills = append(m.Skills, s)
}

// Remove deletes the entry for name. Returns false if it was not present.
func (m *Manifest) Remove(name string) bool {
	for i, s := range m.Skills {
		if s.Name == name {
			m.Skills = append(m.Skills[:i], m.Skills[i+1:]...)
			return true
		}
	}
	return false
}

// InstalledAt returns the install timestamp for name, or zero time if not installed.
func (m *Manifest) InstalledAt(name string) time.Time {
	if s := m.Find(name); s != nil {
		return s.InstalledAt
	}
	return time.Time{}
}

// FindExisting returns an installed skill only if its install path currently
// exists. If the path is missing, the stale manifest entry is removed.
func (m *Manifest) FindExisting(name string) (*skill.InstalledSkill, error) {
	s := m.Find(name)
	if s == nil {
		return nil, nil
	}

	if _, err := os.Stat(s.InstallPath); err != nil {
		if os.IsNotExist(err) {
			m.Remove(name)
			if err := m.Save(); err != nil {
				return nil, fmt.Errorf("save manifest: %w", err)
			}
			return nil, nil
		}
		return nil, fmt.Errorf("check existing install path %s: %w", s.InstallPath, err)
	}

	return s, nil
}

// FindExistingInCurrentProject returns an installed skill only if it exists on
// disk and is inside the current working directory.
func (m *Manifest) FindExistingInCurrentProject(name string) (*skill.InstalledSkill, error) {
	projectDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get current directory: %w", err)
	}

	s, err := m.FindExisting(name)
	if err != nil || s == nil {
		return s, err
	}

	absInstallPath, err := filepath.Abs(s.InstallPath)
	if err != nil {
		return nil, fmt.Errorf("resolve install path %s: %w", s.InstallPath, err)
	}
	absProjectDir, err := filepath.Abs(projectDir)
	if err != nil {
		return nil, fmt.Errorf("resolve project path %s: %w", projectDir, err)
	}
	absInstallPath, err = filepath.EvalSymlinks(absInstallPath)
	if err != nil {
		return nil, fmt.Errorf("eval install path symlinks %s: %w", absInstallPath, err)
	}
	absProjectDir, err = filepath.EvalSymlinks(absProjectDir)
	if err != nil {
		return nil, fmt.Errorf("eval project path symlinks %s: %w", absProjectDir, err)
	}

	rel, err := filepath.Rel(absProjectDir, absInstallPath)
	if err != nil {
		return nil, fmt.Errorf("compute path relation %s to %s: %w", absProjectDir, absInstallPath, err)
	}

	if rel == "." {
		return s, nil
	}

	if filepath.IsAbs(rel) || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return nil, nil
	}

	return s, nil
}
