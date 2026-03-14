# skills-cli — Implementation Plan

> A Go CLI for browsing, searching, installing, updating, and removing Agent Skills (SKILL.md files) from curated GitHub repositories.

## Overview

Build a standalone Go CLI (`skills-cli`) that aggregates SKILL.md files from public GitHub repos (primarily `github/awesome-copilot` and `anthropics/skills`), lets users browse and search them, and install/update/remove skills in local projects. Uses Cobra+Viper, requires no auth by default, ships as a single cross-platform binary.

## Background

### What are Agent Skills?

- **Open standard** ([agentskills.io](https://agentskills.io)) maintained by Anthropic
- Skills are folders containing `SKILL.md` (YAML frontmatter + markdown) plus optional `scripts/`, `references/`, `assets/`
- Supported by: VS Code Copilot, Claude Code, Codex CLI, Cursor, Gemini CLI, JetBrains Junie, and 10+ more tools
- **Progressive loading**: agents read name+description (~100 tokens) → load full SKILL.md (<5000 tokens) → load resources as needed

### SKILL.md Format

```yaml
---
name: skill-name          # Required: 1-64 chars, lowercase alnum + hyphens, must match folder
description: '...'        # Required: max 1024 chars
license: Apache-2.0       # Optional
compatibility: '...'      # Optional: max 500 chars
metadata:                  # Optional: arbitrary key-value pairs
  author: org
  version: "1.0"
allowed-tools: '...'      # Optional, experimental
---
# Markdown body with instructions, examples, guidelines
```

### Install Locations

| Path | Scope |
|------|-------|
| `.github/skills/<name>/` | Project (Copilot) |
| `.agents/skills/<name>/` | Project (cross-agent) |
| `.claude/skills/<name>/` | Project (Claude) |
| `~/.copilot/skills/<name>/` | Personal (Copilot) |
| `~/.agents/skills/<name>/` | Personal (cross-agent) |
| `~/.claude/skills/<name>/` | Personal (Claude) |

### Why Build This?

| Tool | What it does | Gap |
|------|-------------|-----|
| `skills-hub` (gh extension) | Web catalog + `gh` extension | Requires `gh` CLI, shell script, not standalone |
| `skrills` (Rust) | Validates/syncs across CLIs | No remote browsing/installing |
| `awesome-copilot` MCP server | Docker-based install | Requires MCP setup |
| **skills-cli (ours)** | Browse → search → install → update → remove | Single Go binary, no dependencies |

## Design Decisions

- **Framework**: Cobra + Viper
- **Interface**: CLI commands only (no TUI)
- **Auth**: No token required — public repos via raw GitHub URLs. Optional `GITHUB_TOKEN` for higher rate limits.
- **Install targets**: All standard locations
- **Registry**: Bundled JSON via `//go:embed` + remote updates cached at `~/.skills-cli/registry.json`
- **Default sources**: `github/awesome-copilot`, `anthropics/skills`, `dotnet/skills`, `openai/skills`, `github/vuejs-ai/skills` (user-configurable)

## Implementation Phases

### Phase 0: Install Prerequisite Skills

Install reference skills to assist Copilot during development:

| Skill | Source | Purpose |
|-------|--------|---------|
| `conventional-commit` | `github/awesome-copilot` | Consistent commit messages |
| `make-skill-template` | `github/awesome-copilot` | SKILL.md spec reference (our core domain) |
| `go-mcp-server-generator` | `github/awesome-copilot` | Idiomatic Go project patterns |
| `gh-cli` | `github/awesome-copilot` | GitHub API interaction reference |

Steps:
1. Create `.github/skills/` directory
2. Download and install each skill's `SKILL.md` (and any supporting files)
3. Initialize git repo with first commit
4. Initialize Go module: `go mod init github.com/darrenr/skills-cli`

### Phase 1: Project Scaffold & Core Types

1. Set up directory structure:
   ```
   skills-cli/
   ├── cmd/                 # Cobra commands
   │   ├── root.go
   │   ├── list.go
   │   ├── search.go
   │   ├── install.go
   │   ├── update.go
   │   ├── remove.go
   │   ├── info.go
   │   ├── sources.go
   │   └── version.go
   ├── internal/
   │   ├── config/          # Viper config management
   │   ├── registry/        # Registry loading, caching, updating
   │   ├── skill/           # SKILL.md parsing, validation
   │   ├── source/          # GitHub fetching, raw URL access
   │   └── installer/       # Install/update/remove logic
   ├── registry/            # Bundled registry data
   │   └── sources.json
   ├── main.go
   ├── go.mod
   └── go.sum
   ```
2. Define core types (`internal/skill/types.go`):
   - `Skill`: Name, Description, License, Compatibility, Metadata, SourceRepo, SourcePath, Category, Tags
   - `SkillFrontmatter`: parsed YAML frontmatter
   - `InstalledSkill`: tracks source and install location
3. Implement SKILL.md parser (`internal/skill/parser.go`):
   - Parse YAML frontmatter via `gopkg.in/yaml.v3`
   - Validate against Agent Skills spec

### Phase 2: Registry & Source Management

*Depends on Phase 1*

4. Design registry types (`internal/registry/types.go`):
   - `Registry`: Version, UpdatedAt, Skills[]
   - `SkillEntry`: Name, Description, Category, Tags, Source, License
   - `Source`: Repo, Path, Ref
5. Implement registry loader (`internal/registry/loader.go`):
   - Load bundled registry from embedded JSON
   - Fetch remote updates, cache at `~/.skills-cli/registry.json` with TTL
6. Implement GitHub source manager (`internal/source/github.go`):
   - Fetch via `raw.githubusercontent.com` (no auth)
   - Optional GitHub API client with token
   - Download full skill directory
7. `skills-cli sources` command: `list`, `add`, `remove`, `sync` *(deferred to post-v1)*

### Phase 3: Core Commands

*Depends on Phase 2*

8. `skills-cli list` — flags: `--category`, `--source`, `--installed`, `--json`
9. `skills-cli search <query>` — substring match; flags: `--category`, `--source`, `--limit`
10. `skills-cli info <skill-name>` — fetch + display full SKILL.md
11. `skills-cli install <name> [name...]` — flags: `--target`, `--force`
12. `skills-cli update [name]` — re-download from registry source; flag: `--dry-run`
13. `skills-cli remove <name>` — delete + update manifest; flag: `--force`

### Phase 4: Config & Polish

*Parallel with Phase 3*

14. Viper config: `~/.skills-cli/config.yaml`, `config set/get` commands
15. `version` command with build info (ldflags)
16. Shell completions: bash, zsh, fish, powershell
17. Output format support: `--output table|json|yaml`
18. Makefile: `build`, `test`, `lint`, `install`
19. goreleaser config for cross-platform releases

### Phase 5: Registry Bootstrap

*Parallel with Phase 3/4*

20. Script to generate initial registry from awesome-copilot + anthropics/skills
21. Embed generated registry via `//go:embed`

### Recommended Future Features

- `skills-cli validate [path]` — validate local SKILL.md against spec
- `skills-cli init <name>` — scaffold new skill from template
- `skills-cli catalog` — categorized summary (counts per category)
- `skills-cli diff <name>` — diff installed vs latest version

## Key Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/spf13/viper` | Config management |
| `gopkg.in/yaml.v3` | YAML frontmatter parsing |
| `text/tabwriter` (stdlib) | Table output |
| Substring match (internal logic) | Search implementation |
| `github.com/stretchr/testify` | Test assertions + mocking |

## Testing Strategy

**Approach**: Idiomatic Go — standard `testing` package + `testify`. Every `internal/` package gets `_test.go` files alongside the code.

**Patterns**:
- Table-driven tests with `t.Run()` subtests
- `t.TempDir()` for filesystem tests
- `httptest.NewServer` for HTTP mocking
- `t.Parallel()` for independent tests
- Interfaces for dependency injection (`SourceFetcher`, `RegistryLoader`)
- `testdata/` directories for fixtures

**Tests per package**:

| Package | Key tests |
|---------|-----------|
| `internal/skill/` | Parser: valid/invalid frontmatter, name validation, description limits, body extraction |
| `internal/registry/` | Loader: embedded/cached, TTL expiry. Search: substring + filters |
| `internal/source/` | HTTP mocking for raw GitHub URLs, 404/timeout, auth headers |
| `internal/installer/` | Install to temp dirs, manifest tracking, force overwrite, update, remove |
| `cmd/config.go` | Defaults, config file loading, env var override |
| `cmd/` | Integration tests via `cobra.Command.Execute()` |

**Targets**: `go test ./...`, `go test -race ./...`, coverage >80% on `internal/`

## Verification Checklist

1. `go build ./...` compiles
2. `go test ./...` — all green
3. `go test -race ./...` — no races
4. Coverage >80% on `internal/`
5. `skills-cli list` shows bundled registry
6. `skills-cli search testing` returns results
7. `skills-cli install git-commit` creates `.github/skills/git-commit/`
8. `skills-cli info git-commit` displays SKILL.md
9. `skills-cli update git-commit` detects + re-downloads
10. `skills-cli remove git-commit` cleans up
11. Cross-compile: darwin/arm64, darwin/amd64, linux/amd64

### Post-v1 Verification

1. `skills-cli validate .github/skills/my-skill/` validates format
2. `skills-cli sources sync` fetches from all sources

## Scope

**Included (v1)**: Browse, search, install, update, remove skills from public GitHub repos; curated bundled registry sources; all standard install locations; local caching; cross-platform binary.

**Excluded (for now)**: Private repos, dynamic source management (`sources` command), local `validate` command, `init` scaffolding command, MCP server, TUI, multi-CLI sync, security scanning, VS Code extension.
