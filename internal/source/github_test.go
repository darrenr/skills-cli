package source_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/darrenr/skills-cli/internal/source"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestFetcher returns a GitHubFetcher whose HTTP client is wired to srv.
func newTestFetcher(srv *httptest.Server) *source.GitHubFetcher {
	f := source.NewGitHubFetcher("")
	f.SetClient(srv.Client())
	f.SetBaseURL(srv.URL)
	f.SetAPIBaseURL(srv.URL)
	return f
}

func TestFetchFile_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/owner/repo/main/skills/my-skill/SKILL.md", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("---\nname: my-skill\n---\n# Body"))
	}))
	defer srv.Close()

	f := newTestFetcher(srv)
	data, err := f.FetchFile(context.Background(), "owner/repo", "main", "skills/my-skill/SKILL.md")
	require.NoError(t, err)
	assert.Contains(t, string(data), "my-skill")
}

func TestFetchFile_NotFound(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	f := newTestFetcher(srv)
	_, err := f.FetchFile(context.Background(), "owner/repo", "main", "skills/missing/SKILL.md")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFetchFile_ServerError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	f := newTestFetcher(srv)
	_, err := f.FetchFile(context.Background(), "owner/repo", "main", "skills/my-skill/SKILL.md")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestFetchFile_AuthHeader(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer mytoken", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("content"))
	}))
	defer srv.Close()

	f := source.NewGitHubFetcher("mytoken")
	f.SetClient(srv.Client())
	f.SetBaseURL(srv.URL)
	_, err := f.FetchFile(context.Background(), "owner/repo", "main", "path/file.md")
	require.NoError(t, err)
}

func TestFetchFile_ContextCancelled(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	f := newTestFetcher(srv)
	_, err := f.FetchFile(ctx, "owner/repo", "main", "path/file.md")
	require.Error(t, err)
}

func TestFetchFile_RepoOwnerGithub_IsNotStripped(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/github/awesome-copilot/main/skills/git-commit/SKILL.md", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	f := newTestFetcher(srv)
	_, err := f.FetchFile(context.Background(), "github/awesome-copilot", "main", "skills/git-commit/SKILL.md")
	require.NoError(t, err)
}

func TestDownloadSkill_ProviderStyleGithubRepo_IsNormalized(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/vuejs-ai/skills/contents/skills/vue-best-practices":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]map[string]string{{
				"type": "file",
				"path": "skills/vue-best-practices/SKILL.md",
			}})
			return
		case "/vuejs-ai/skills/main/skills/vue-best-practices/SKILL.md":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("---\nname: vue-best-practices\n---\n# Body"))
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}))
	defer srv.Close()

	dir := t.TempDir()
	f := newTestFetcher(srv)
	written, err := f.DownloadSkill(context.Background(), "github/vuejs-ai/skills", "main", "skills/vue-best-practices", dir, nil)
	require.NoError(t, err)
	require.Len(t, written, 1)
}

func TestDownloadSkill_WritesFiles(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/owner/repo/contents/skills/my-skill" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]map[string]string{{
				"type": "file",
				"path": "skills/my-skill/SKILL.md",
			}})
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("content of " + filepath.Base(r.URL.Path)))
	}))
	defer srv.Close()

	dir := t.TempDir()
	f := newTestFetcher(srv)
	written, err := f.DownloadSkill(context.Background(), "owner/repo", "main", "skills/my-skill", dir, nil)
	require.NoError(t, err)
	require.Len(t, written, 1)

	data, err := os.ReadFile(filepath.Join(dir, "SKILL.md"))
	require.NoError(t, err)
	assert.Equal(t, "content of SKILL.md", string(data))
}

func TestDownloadSkill_RecursivelyWritesNestedFiles(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/owner/repo/contents/skills/my-skill":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]map[string]string{
				{"type": "file", "path": "skills/my-skill/SKILL.md"},
				{"type": "dir", "path": "skills/my-skill/scripts"},
			})
			return
		case "/repos/owner/repo/contents/skills/my-skill/scripts":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]map[string]string{
				{"type": "file", "path": "skills/my-skill/scripts/setup.sh"},
			})
			return
		case "/owner/repo/main/skills/my-skill/SKILL.md":
			_, _ = w.Write([]byte("---\nname: my-skill\ndescription: desc\n---\n# Body"))
			return
		case "/owner/repo/main/skills/my-skill/scripts/setup.sh":
			_, _ = w.Write([]byte("#!/bin/sh\necho hi\n"))
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}))
	defer srv.Close()

	dir := t.TempDir()
	f := newTestFetcher(srv)
	written, err := f.DownloadSkill(context.Background(), "owner/repo", "main", "skills/my-skill", dir, nil)
	require.NoError(t, err)
	require.Len(t, written, 2)

	assert.FileExists(t, filepath.Join(dir, "SKILL.md"))
	assert.FileExists(t, filepath.Join(dir, "scripts", "setup.sh"))
}

func TestDownloadSkill_MissingOptionalFileSkipped(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/owner/repo/contents/skills/my-skill" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]map[string]string{{
				"type": "file",
				"path": "skills/my-skill/SKILL.md",
			}})
			return
		}
		if filepath.Base(r.URL.Path) == "SKILL.md" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("skill content"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	dir := t.TempDir()
	f := newTestFetcher(srv)
	written, err := f.DownloadSkill(context.Background(), "owner/repo", "main", "skills/my-skill", dir, []string{"scripts/setup.sh"})
	require.NoError(t, err)
	assert.Len(t, written, 1) // only SKILL.md
}

func TestDownloadSkill_MissingSkillMdIsError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/owner/repo/contents/skills/my-skill" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]map[string]string{})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	dir := t.TempDir()
	f := newTestFetcher(srv)
	_, err := f.DownloadSkill(context.Background(), "owner/repo", "main", "skills/my-skill", dir, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
