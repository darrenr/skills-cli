# skills-cli

[![Build](https://img.shields.io/github/actions/workflow/status/darrenr/skills-cli/build.yml?style=flat-square)](https://github.com/darrenr/skills-cli/actions)
[![Go version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat-square&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue?style=flat-square)](LICENSE)

> Browse, search, install, and manage Agent Skills (SKILL.md) from curated public registries — no authentication required.

Agent Skills are the open standard ([agentskills.io](https://agentskills.io)) for reusable AI agent instructions. `skills-cli` gives you a single command-line tool to discover and install them across all major AI coding agents: VS Code Copilot, Claude Code, Codex CLI, Cursor, Gemini CLI, and more.

## Features

- **Bundled curated registry** — seeded from `github/awesome-copilot`, `anthropics/skills`, `dotnet/skills`, and `openai/skills`
- **Browse and search** — filter by category, source, or free-text keyword
- **Multi-target install** — project scope (`.github/skills/`, `.agents/skills/`, `.claude/skills/`) or personal global scope
- **Track installed skills** — manifest at `~/.skills-cli/manifest.json`
- **Update and remove** — keep skills current or clean up what you no longer need
- **Single static binary** — no runtime, no Docker, no `gh` CLI required

## Installation

### Homebrew (macOS/Linux)

```bash
brew install darrenr/tap/skills-cli
```

### Go install

```bash
go install github.com/darrenr/skills-cli@latest
```

### Download a binary

Grab the latest release for your platform from the [Releases page](https://github.com/darrenr/skills-cli/releases).

### Build from source

```bash
git clone https://github.com/darrenr/skills-cli.git
cd skills-cli
make build
```

## Quick start

```bash
# Browse bundled skills
skills-cli list

# Filter by category
skills-cli list --category dotnet

# Search by keyword
skills-cli search "commit messages"

# See the full SKILL.md content for a skill
skills-cli info conventional-commit

# Install into your project (Copilot scope → .github/skills/)
skills-cli install conventional-commit

# Install for Claude Code instead
skills-cli install mcp-builder --target project-claude

# Update all installed skills
skills-cli update

# Remove a skill
skills-cli remove conventional-commit
```

## Commands

| Command | Description |
|---|---|
| `list` | List skills from the registry |
| `search <query>` | Search by keyword, category, or tags |
| `info <name>` | Show full SKILL.md content fetched from source |
| `install <name...>` | Download and install one or more skills |
| `update [name]` | Update one or all installed skills |
| `remove <name...>` | Remove installed skills |
| `config get/set/list` | Manage configuration |
| `version` | Print version information |

### Install targets

Use `--target` with `install` to choose where the skill is written:

| Target | Path | Scope |
|---|---|---|
| `project-copilot` *(default)* | `.github/skills/<name>/` | Project — VS Code Copilot |
| `project-agents` | `.agents/skills/<name>/` | Project — cross-agent |
| `project-claude` | `.claude/skills/<name>/` | Project — Claude Code |
| `global-copilot` | `~/.copilot/skills/<name>/` | Personal — VS Code Copilot |
| `global-agents` | `~/.agents/skills/<name>/` | Personal — cross-agent |
| `global-claude` | `~/.claude/skills/<name>/` | Personal — Claude Code |

### Common flags

```
--output, -o    Output format: table (default), json
--config        Path to config file (default: ~/.skills-cli/config.yaml)
```

## Registry

The bundled registry (`registry/skills.json`) is embedded at compile time and covers four curated sources:

| Source | Focus |
|---|---|
| [`github/awesome-copilot`](https://github.com/github/awesome-copilot) | Git, GitHub, docs, languages, MCP, testing |
| [`anthropics/skills`](https://github.com/anthropics/skills) | Claude API, MCP builder, frontend, file formats, creative |
| [`dotnet/skills`](https://github.com/dotnet/skills) | .NET diagnostics, MSBuild, upgrade migrations, EF Core |
| [`openai/skills`](https://github.com/openai/skills) | Curated OpenAI-focused skills for docs, platform workflows, integrations, and tooling |
| [`github/vuejs-ai/skills`](https://github.com/vuejs-ai/skills) | Vue-focused guidance: composables, router, pinia, testing, options-api, JSX, debugging |

The exact number of bundled skills can change as the registry is refreshed.


### Adding A New Source

When adding a new source repo to `registry/skills.json`, use this checklist:

1. Verify each candidate skill path has a real `SKILL.md`.
2. Extract `name` + frontmatter `description` from upstream `SKILL.md`.
3. Keep descriptions concise in the registry (max ~300 chars).
4. Map every entry to an existing category slug in this project.
5. Skip duplicate names (or rename only with explicit project decision).
6. Ensure `source.repo`, `source.path`, and `source.ref` are all valid.
7. Run tests and spot-check CLI output.

Suggested validation flow:

```bash
# 1) Confirm path exists
curl -sf "https://raw.githubusercontent.com/<owner>/<repo>/<ref>/<path>/SKILL.md" | head -5

# 2) Verify registry JSON is valid
jq empty registry/skills.json

# 3) Check new source entries render correctly
go run . list --source <owner/repo> | head -n 40

# 4) Run test suite
go test ./...
```

Notes:

- Do not use placeholder descriptions for bulk imports.
- Use bounded network calls (for example, `curl --max-time 20`) for large fetch loops.
- Keep category assignments intentional so `list --category` remains useful.

### Categories

`git` · `github` · `skills` · `docs` · `planning` · `code-quality` · `testing` · `mcp` · `ai` · `dotnet` · `csharp` · `java` · `javascript` · `python` · `rust` · `database` · `docker` · `creative`

> [!TIP]
> Set `GITHUB_TOKEN` in your environment to raise GitHub API rate limits when using the `info` command frequently.

## Configuration

`skills-cli` reads `~/.skills-cli/config.yaml` on startup. You can manage it with the `config` subcommand:

```bash
skills-cli config set output json
skills-cli config get output
skills-cli config list
```

Environment variables with the `SKILLS_` prefix override config file values:

```bash
SKILLS_OUTPUT=json skills-cli list
```

## What are Agent Skills?

An Agent Skill is a folder containing a `SKILL.md` file with YAML frontmatter and markdown instructions, plus optional `scripts/`, `references/`, and `assets/` subdirectories.

```
my-skill/
  SKILL.md          # Required — frontmatter + instructions
  scripts/          # Optional — helper scripts the agent can run
  references/       # Optional — reference documents
  assets/           # Optional — images, templates
```

Skills follow a *progressive loading* pattern: the agent reads only the `name` and `description` (~100 tokens) until it decides the skill is relevant, then loads the full body (<5000 tokens), then fetches resources as needed.

Supported by 10+ AI coding agents including VS Code Copilot, Claude Code, Codex CLI, Cursor, and Gemini CLI.

## Development

```bash
make test          # Run all tests with race detector
make test-cover    # Produce an HTML coverage report
make lint          # Run golangci-lint
make build         # Build ./skills-cli with version ldflags
```

> [!NOTE]
> Go 1.26+ is required. All packages pass tests with the race detector enabled; internal package coverage is ≥85%.

## Related projects

- [agentskills.io](https://agentskills.io) — The open Agent Skills specification
- [github/awesome-copilot](https://github.com/github/awesome-copilot) — Curated Copilot skills and instructions
- [anthropics/skills](https://github.com/anthropics/skills) — Official Anthropic Agent Skills
- [dotnet/skills](https://github.com/dotnet/skills) — Official Microsoft .NET Agent Skills
- [openai/skills](https://github.com/openai/skills) — Official OpenAI skills repository
