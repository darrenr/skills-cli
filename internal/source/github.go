package source

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	rawBase        = "https://raw.githubusercontent.com"
	apiBase        = "https://api.github.com"
	defaultTimeout = 30 * time.Second
)

// GitHubFetcher fetches files from public GitHub repositories via
// raw.githubusercontent.com. No authentication is required for public repos;
// setting Token enables a higher API rate limit.
type GitHubFetcher struct {
	client  *http.Client
	token   string
	baseURL string // overridden in tests
	apiURL  string // overridden in tests
}

// SetClient replaces the HTTP client (used in tests to inject a test server client).
func (f *GitHubFetcher) SetClient(c *http.Client) { f.client = c }

// SetBaseURL overrides the base URL (used in tests to point at a local httptest server).
func (f *GitHubFetcher) SetBaseURL(u string) { f.baseURL = u }

// SetAPIBaseURL overrides the GitHub API base URL (used in tests).
func (f *GitHubFetcher) SetAPIBaseURL(u string) { f.apiURL = u }

// NewGitHubFetcher creates a fetcher. token may be empty for unauthenticated
// access (60 req/hr limit). Set GITHUB_TOKEN or pass it directly for 5000/hr.
func NewGitHubFetcher(token string) *GitHubFetcher {
	return &GitHubFetcher{
		client: &http.Client{Timeout: defaultTimeout},
		token:  token,
		apiURL: apiBase,
	}
}

// FetchFile downloads a single file from a GitHub repository and returns its
// contents. repo is "owner/name", ref is a branch/tag/SHA, path is the file
// path within the repo.
func (f *GitHubFetcher) FetchFile(ctx context.Context, repo, ref, path string) ([]byte, error) {
	url := f.rawURL(repo, ref, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	if f.token != "" {
		req.Header.Set("Authorization", "Bearer "+f.token)
	}
	req.Header.Set("User-Agent", "skills-cli")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, &NotFoundError{Repo: repo, Ref: ref, Path: path}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch %s: unexpected status %d", url, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	return data, nil
}

// DownloadSkill fetches the SKILL.md (and any supporting files listed in
// extraFiles) from the remote repo and writes them into destDir, creating it
// if needed. Returns the paths of files written.
func (f *GitHubFetcher) DownloadSkill(ctx context.Context, repo, ref, skillPath, destDir string, extraFiles []string) ([]string, error) {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return nil, fmt.Errorf("create dest dir: %w", err)
	}

	files, err := f.listSkillFiles(ctx, repo, ref, skillPath)
	if err != nil {
		return nil, fmt.Errorf("list skill files: %w", err)
	}

	files = append(files, extraFiles...)
	files = uniqStrings(files)
	var written []string

	for _, relPath := range files {
		if !isSafeRelativePath(relPath) {
			return written, fmt.Errorf("unsafe relative path %q", relPath)
		}

		remotePath := strings.TrimSuffix(skillPath, "/") + "/" + relPath
		data, err := f.FetchFile(ctx, repo, ref, remotePath)
		if err != nil {
			if isNotFound(err) && relPath != "SKILL.md" {
				// Optional supporting files are allowed to be missing.
				continue
			}
			return written, err
		}
		dest := filepath.Join(destDir, relPath)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return written, fmt.Errorf("create parent dir for %s: %w", dest, err)
		}
		if err := os.WriteFile(dest, data, 0o644); err != nil {
			return written, fmt.Errorf("write %s: %w", dest, err)
		}
		written = append(written, dest)
	}
	return written, nil
}

type githubContentEntry struct {
	Type string `json:"type"`
	Path string `json:"path"`
}

// listSkillFiles returns file paths relative to skillPath, discovered recursively
// via the GitHub Contents API. SKILL.md must exist.
func (f *GitHubFetcher) listSkillFiles(ctx context.Context, repo, ref, skillPath string) ([]string, error) {
	trimmedSkillPath := strings.Trim(strings.TrimSpace(skillPath), "/")
	if trimmedSkillPath == "" {
		return nil, fmt.Errorf("skill path is required")
	}

	var files []string
	if err := f.walkContents(ctx, repo, ref, trimmedSkillPath, trimmedSkillPath, &files); err != nil {
		return nil, err
	}

	files = uniqStrings(files)
	sort.Strings(files)

	hasSkillMD := false
	for _, f := range files {
		if f == "SKILL.md" {
			hasSkillMD = true
			break
		}
	}
	if !hasSkillMD {
		return nil, &NotFoundError{Repo: repo, Ref: ref, Path: trimmedSkillPath + "/SKILL.md"}
	}

	return files, nil
}

func (f *GitHubFetcher) walkContents(ctx context.Context, repo, ref, rootSkillPath, currentPath string, files *[]string) error {
	repo = strings.TrimPrefix(repo, "github/")
	url := fmt.Sprintf("%s/repos/%s/contents/%s?ref=%s", strings.TrimSuffix(f.apiURL, "/"), repo, currentPath, ref)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	if f.token != "" {
		req.Header.Set("Authorization", "Bearer "+f.token)
	}
	req.Header.Set("User-Agent", "skills-cli")

	resp, err := f.client.Do(req)
	if err != nil {
		return fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &NotFoundError{Repo: repo, Ref: ref, Path: currentPath}
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch %s: unexpected status %d", url, resp.StatusCode)
	}

	var entries []githubContentEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return fmt.Errorf("decode contents response: %w", err)
	}

	for _, e := range entries {
		switch e.Type {
		case "file":
			rel, ok := strings.CutPrefix(e.Path, rootSkillPath+"/")
			if !ok {
				if e.Path == rootSkillPath {
					rel = filepath.Base(rootSkillPath)
				} else {
					continue
				}
			}
			if rel != "" {
				*files = append(*files, rel)
			}
		case "dir":
			if err := f.walkContents(ctx, repo, ref, rootSkillPath, e.Path, files); err != nil {
				return err
			}
		}
	}

	return nil
}

func uniqStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, v := range values {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		result = append(result, v)
	}
	return result
}

func isSafeRelativePath(path string) bool {
	if path == "" {
		return false
	}
	if filepath.IsAbs(path) {
		return false
	}
	clean := filepath.Clean(path)
	if clean == "." {
		return false
	}
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return false
	}
	return clean == path
}

// rawURL builds the raw file URL using the configured base (or the default).
func (f *GitHubFetcher) rawURL(repo, ref, path string) string {
	base := f.baseURL
	if base == "" {
		base = rawBase
	}
	repo = strings.TrimPrefix(repo, "github/")
	return fmt.Sprintf("%s/%s/%s/%s", base, repo, ref, strings.TrimPrefix(path, "/"))
}

// NotFoundError is returned when a file or skill does not exist in the repo.
type NotFoundError struct {
	Repo string
	Ref  string
	Path string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("not found: %s/%s@%s", e.Repo, e.Path, e.Ref)
}

func isNotFound(err error) bool {
	var nfe *NotFoundError
	return errors.As(err, &nfe)
}
