# GitHub Copilot Instructions — skills-cli

## Project Context

- **Module**: `github.com/darrenr/skills-cli`
- **From**: initial scaffold (go.mod only, `main` branch)
- **To**: full v1 CLI with registry, installer, and 128 bundled skills (`develop` branch)
- **Migration type**: Greenfield Go CLI — built from scratch following established conventions
- **Go version**: 1.26+

---

## Architecture

```
skills-cli/
├── main.go                   # entry point — calls cmd.Execute() only
├── cmd/                      # Cobra CLI commands
│   ├── root.go               # root command + Viper config init
│   ├── helpers.go            # shared registry loader + table/JSON formatter
│   └── <command>.go          # one file per subcommand
├── internal/
│   ├── skill/                # domain types, SKILL.md parser, validator
│   ├── registry/             # registry types, loader (embed+cache), search
│   ├── source/               # GitHub raw-content fetcher
│   └── installer/            # manifest, install, update, remove logic
└── registry/
    ├── embed.go              # //go:embed skills.json → var Skills []byte
    └── skills.json           # bundled registry seed (128 skills, 4 sources)
```

All business logic lives in `internal/`. The `cmd/` layer is thin: parse flags → call internal → print results.

---

## Mandatory Patterns

### 1. Package boundaries

| Package | Allowed to import | Forbidden |
|---|---|---|
| `internal/skill` | stdlib only | anything in `internal/` |
| `internal/registry` | stdlib, `internal/skill` | `cmd/`, `source`, `installer` |
| `internal/source` | stdlib | `internal/registry`, `installer` |
| `internal/installer` | stdlib, `internal/skill`, `internal/registry`, `internal/source` | `cmd/` |
| `cmd/` | all `internal/`, `registry/` | direct HTTP, file I/O |
| `registry/` | `embed` only | all others |

### 2. Error wrapping — always use `%w`

```go
// CORRECT
return nil, fmt.Errorf("load registry: %w", err)
return nil, fmt.Errorf("fetch %s: %w", name, err)

// WRONG — loses the error chain
return nil, fmt.Errorf("load registry: %v", err)
return errors.New("load failed")
```

### 3. Error type checks — use `errors.As`, never direct assertion

```go
// CORRECT
var nfe *source.NotFoundError
if errors.As(err, &nfe) { ... }

if os.IsNotExist(err) { ... }  // stdlib sentinel is fine

// WRONG
if err.(*source.NotFoundError) != nil { ... }
```

### 4. File paths — always `filepath.Join`, never string concatenation

```go
// CORRECT
dest := filepath.Join(opts.TargetDir, entry.Name)
path := filepath.Join(home, ".skills-cli", "manifest.json")

// WRONG
dest := opts.TargetDir + "/" + entry.Name
path := home + "/.skills-cli/manifest.json"
```

### 5. HTTP requests — always use `http.NewRequestWithContext`

```go
// CORRECT
req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

// WRONG — ignores context cancellation
req, err := http.NewRequest(http.MethodGet, url, nil)
```

### 6. Directory permissions

| Use case | Permission |
|---|---|
| Config / manifest directories (`~/.skills-cli/`) | `0o700` |
| Skill directories installed in project | `0o755` |
| Written files | `0o644` (skill files), `0o600` (manifest JSON) |

---

## New Command Pattern

Every new subcommand follows this structure:

```go
// cmd/<name>.go

var <name>Cmd = &cobra.Command{
    Use:   "<name> <required> [optional...]",
    Short: "One-line summary",
    Long: `Longer description.

Examples:
  skills-cli <name> foo
  skills-cli <name> bar --flag value`,
    Args: cobra.ExactArgs(1), // or MinimumNArgs, etc.
    RunE: run<Name>,          // always RunE, never Run
}

func init() {
    <name>Cmd.Flags().StringP("flag", "f", "default", "flag description")
    rootCmd.AddCommand(<name>Cmd)
}

func run<Name>(cmd *cobra.Command, args []string) error {
    // 1. Read flags
    flag, _ := cmd.Flags().GetString("flag")

    // 2. Load registry if needed
    r, err := loadRegistry()
    if err != nil {
        return fmt.Errorf("load registry: %w", err)
    }

    // 3. Load manifest if needed
    manifestPath, err := installer.DefaultManifestPath()
    if err != nil {
        return err
    }
    manifest, err := installer.LoadManifest(manifestPath)
    if err != nil {
        return fmt.Errorf("load manifest: %w", err)
    }

    // 4. Do work — errors printed to stderr, continue loop if batching
    // 5. Print results via printSkillEntries() or fmt.Printf
    return nil
}
```

**Rules:**
- Use `RunE` (returns error), never `Run`
- Print errors inside loops to `os.Stderr` so one failure doesn't abort the batch
- Call `loadRegistry()` from `cmd/helpers.go`, never construct a `Loader` directly in cmd
- Never touch `os.Exit` in command handlers — let `cmd.Execute()` handle it

---

## Registry Entry Pattern

New entries in `registry/skills.json`:

```jsonc
{
  "name": "skill-name",           // lowercase, hyphens only, matches folder name
  "description": "...",           // ≤300 chars in the registry (full desc in SKILL.md)
  "category": "category-slug",    // see categories below
  "tags": ["tag1", "tag2"],
  "license": "MIT",               // or "" if unknown
  "source": {
    "repo": "owner/repo",         // GitHub owner/repo — no https://
    "path": "path/to/skill-dir",  // path to the directory containing SKILL.md
    "ref": "main"                 // branch or tag
  }
}
```

**Existing categories:** `git` · `github` · `skills` · `docs` · `planning` · `code-quality` · `testing` · `mcp` · `ai` · `dotnet` · `csharp` · `java` · `javascript` · `python` · `rust` · `database` · `docker` · `creative`

**Validation before adding a registry entry:**
```bash
# Verify the SKILL.md actually exists at the declared path
curl -sf "https://raw.githubusercontent.com/<owner>/<repo>/<ref>/<path>/SKILL.md" | head -5
```

If the path returns 404, do **not** add the entry. `agent-customization` was removed because it was a VS Code built-in with no GitHub path.

### Adding A New Registry Source (Bulk)

When importing many entries from a new repo:

1. Discover candidate skill directories from upstream.
2. Validate each `.../SKILL.md` path before writing entries.
3. Extract descriptions from frontmatter (`description:`), do not use placeholder text.
4. Truncate descriptions to ~300 chars for registry readability.
5. Skip duplicate `name` values already present in `registry/skills.json`.
6. Map each skill into one of the existing category slugs.
7. Validate JSON (`jq empty registry/skills.json`) before replacing file contents.
8. Run `go test ./...` and spot-check `go run . list --source <repo>`.

Required safety rules for bulk updates:

- Use bounded-time network calls in loops (for example, `curl --max-time 20`) to avoid hanging sessions.
- Never leave fallback placeholder descriptions in committed registry entries.
- If a command partially truncates `registry/skills.json`, restore from git immediately before retrying.

---

## Test Patterns

### Required for every new `internal/` package

```go
package <pkg>_test  // external test package, not package <pkg>

import (
    "testing"
    "<module>/internal/<pkg>"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestFoo(t *testing.T) {
    t.Parallel()  // always parallel at top level

    tests := []struct { ... }{ ... }  // table-driven

    for _, tc := range tests {
        tc := tc  // capture loop variable
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()
            // ...
        })
    }
}
```

**Rules:**
- Always `t.Parallel()` at both the outer test and each sub-test
- Use `require.NoError` / `require.Error` for fatal assertions that block further checks
- Use `assert.*` for non-fatal checks (collect all failures in one run)
- Use `t.TempDir()` for filesystem tests — never hardcode temp paths
- Use `httptest.NewServer` for HTTP tests — never hit real GitHub in unit tests
- Inject `SetClient` / `SetBaseURL` into `GitHubFetcher` for test isolation

### Coverage target

All `internal/` packages must maintain **≥80% statement coverage** (`go test -race -coverprofile=coverage.out ./internal/...`).

---

## Obsolete Patterns to Avoid

| Avoid | Use instead |
|---|---|
| `destDir := baseDir + "/" + name` | `filepath.Join(baseDir, name)` |
| `err.(*SomeType)` | `errors.As(err, &target)` |
| `fmt.Errorf("...: %v", err)` | `fmt.Errorf("...: %w", err)` |
| `http.NewRequest(...)` | `http.NewRequestWithContext(ctx, ...)` |
| `cmd.Run = func(...)` | `cmd.RunE = func(...) error` |
| Constructing `registry.Loader` directly in `cmd/` | `loadRegistry()` from `cmd/helpers.go` |
| Registry entries without verifying the `path/SKILL.md` exists | Verify with `curl -sf` before adding |

---

## Build and Release

```bash
make build       # ./skills-cli with ldflags: Version, Commit, BuildDate
make test        # go test -race ./...
make test-cover  # coverage report → coverage.html
make install     # go install to $GOPATH/bin
```

Version is injected at link time:
```
-X github.com/darrenr/skills-cli/cmd.Version=$(git describe --tags)
-X github.com/darrenr/skills-cli/cmd.Commit=$(git rev-parse --short HEAD)
-X github.com/darrenr/skills-cli/cmd.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)
```

Releases via GoReleaser (`.goreleaser.yml`): darwin/linux/windows × amd64/arm64.

---

## Validation Checklist

Before merging to `main`:

- [ ] `go build ./...` clean
- [ ] `go test -race ./...` all green
- [ ] `internal/` coverage ≥80% (`make test-cover`)
- [ ] `skills-cli list` shows expected skill count
- [ ] `skills-cli install <skill>` + `remove <skill>` round-trip works
- [ ] Cross-compile: `GOOS=linux GOARCH=amd64 go build -o /dev/null .`
