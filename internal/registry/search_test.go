package registry_test

import (
	"testing"
	"time"

	"github.com/darrenr/skills-cli/internal/registry"
	"github.com/stretchr/testify/assert"
)

func testRegistry() *registry.Registry {
	return &registry.Registry{
		Version:   "1",
		UpdatedAt: time.Now(),
		Skills: []registry.SkillEntry{
			{Name: "git-commit", Description: "Write great commit messages.", Category: "git", Tags: []string{"git", "commits"}, Source: registry.Source{Repo: "github/awesome-copilot"}},
			{Name: "go-test", Description: "Write idiomatic Go tests.", Category: "golang", Tags: []string{"go", "testing"}, Source: registry.Source{Repo: "github/awesome-copilot"}},
			{Name: "conventional-commit", Description: "Conventional commit standard.", Category: "git", Tags: []string{"git", "commits", "conventional"}, Source: registry.Source{Repo: "anthropics/skills"}},
			{Name: "rust-patterns", Description: "Idiomatic Rust patterns.", Category: "rust", Tags: []string{"rust"}, Source: registry.Source{Repo: "anthropics/skills"}},
		},
	}
}

func TestSearch_NoFilter(t *testing.T) {
	t.Parallel()
	r := testRegistry()
	results := registry.Search(r, registry.SearchOptions{})
	assert.Len(t, results, 4)
}

func TestSearch_ByQuery_Name(t *testing.T) {
	t.Parallel()
	results := registry.Search(testRegistry(), registry.SearchOptions{Query: "commit"})
	assert.Len(t, results, 2)
}

func TestSearch_ByQuery_Description(t *testing.T) {
	t.Parallel()
	results := registry.Search(testRegistry(), registry.SearchOptions{Query: "idiomatic"})
	assert.Len(t, results, 2)
}

func TestSearch_ByQuery_Tag(t *testing.T) {
	t.Parallel()
	results := registry.Search(testRegistry(), registry.SearchOptions{Query: "testing"})
	assert.Len(t, results, 1)
	assert.Equal(t, "go-test", results[0].Name)
}

func TestSearch_ByCategory(t *testing.T) {
	t.Parallel()
	results := registry.Search(testRegistry(), registry.SearchOptions{Category: "git"})
	assert.Len(t, results, 2)
}

func TestSearch_BySource(t *testing.T) {
	t.Parallel()
	results := registry.Search(testRegistry(), registry.SearchOptions{Source: "anthropics/skills"})
	assert.Len(t, results, 2)
}

func TestSearch_ByQueryAndCategory(t *testing.T) {
	t.Parallel()
	results := registry.Search(testRegistry(), registry.SearchOptions{Query: "commit", Category: "git", Source: "anthropics/skills"})
	assert.Len(t, results, 1)
	assert.Equal(t, "conventional-commit", results[0].Name)
}

func TestSearch_CaseInsensitive(t *testing.T) {
	t.Parallel()
	results := registry.Search(testRegistry(), registry.SearchOptions{Query: "COMMIT"})
	assert.Len(t, results, 2)
}

func TestSearch_Limit(t *testing.T) {
	t.Parallel()
	results := registry.Search(testRegistry(), registry.SearchOptions{Limit: 2})
	assert.Len(t, results, 2)
}

func TestSearch_NoResults(t *testing.T) {
	t.Parallel()
	results := registry.Search(testRegistry(), registry.SearchOptions{Query: "python"})
	assert.Empty(t, results)
}
