# skills-cli — Progress Tracker

> Track implementation progress, capture decisions, and note issues as we build.

## Status

| Phase | Status | Notes |
|-------|--------|-------|
| Phase 0: Prerequisite Skills | ✅ Done | 4 skills installed, git + go mod initialized |
| Phase 1: Scaffold & Core Types | ✅ Done | All types, parser, validator + 25 tests |
| Phase 2: Registry & Source Mgmt | ✅ Done (v1 scope) | Loader, search, GitHub fetcher + tests (`sources` deferred) |
| Phase 3: Core Commands | ✅ Done | list/search/info/install/update/remove + manifest |
| Phase 4: Config & Polish | ✅ Done (v1 scope) | config cmd, --installed, --output table/json/yaml, Makefile, GoReleaser |
| Phase 5: Registry Bootstrap | ✅ Done | Registry bootstrap complete (snapshot: 2026-03-14), .gitignore, README |

Legend: ⬜ Not started · 🟡 In progress · ✅ Done · ❌ Blocked

## Phase 0: Prerequisite Skills

- [x] Create `.github/skills/` directory
- [x] Install `conventional-commit` skill
- [x] Install `make-skill-template` skill
- [x] Install `go-mcp-server-generator` skill
- [x] Install `gh-cli` skill
- [x] Initialize git repo + first commit
- [x] Initialize Go module

## Phase 1: Scaffold & Core Types

- [x] Create directory structure (`cmd/`, `internal/`, `registry/`)
- [x] `main.go` with Cobra root command
- [x] `internal/skill/types.go` — core types
- [x] `internal/skill/parser.go` — SKILL.md parser
- [x] `internal/skill/parser_test.go` — parser tests with testdata fixtures
- [x] `internal/skill/validator.go` — spec validation
- [x] `internal/skill/validator_test.go` — validator tests

## Phase 2: Registry & Source Management

- [x] `internal/registry/types.go` — registry types
- [x] `internal/registry/loader.go` — embedded + cached loader
- [x] `internal/registry/loader_test.go`
- [x] `internal/registry/search.go` — substring search
- [x] `internal/registry/search_test.go`
- [x] `internal/source/github.go` — raw URL fetcher
- [x] `internal/source/github_test.go`
- [ ] `cmd/sources.go` — sources command (deferred to v2)

## Phase 3: Core Commands

- [x] `cmd/list.go`
- [x] `cmd/search.go`
- [x] `cmd/info.go`
- [x] `cmd/install.go`
- [x] `internal/installer/installer.go`
- [x] `internal/installer/installer_test.go`
- [x] `cmd/update.go`
- [x] `internal/installer/updater.go`
- [x] `internal/installer/manifest.go` + `installer_test.go` covers updater/remover
- [x] `cmd/remove.go`
- [x] `internal/installer/remover.go`

## Phase 4: Config & Polish

- [x] `cmd/config.go` — config get/set/list (Viper-backed, no separate internal package needed)
- [x] `cmd/version.go`
- [x] `cmd/completion.go` (built into Cobra automatically)
- [x] `--output` flag support (table/json)
- [x] `--installed` flag on `list`
- [x] `Makefile`
- [x] `.goreleaser.yml`

## Phase 5: Registry Bootstrap

- [x] `registry/embed.go` — `//go:embed skills.json`
- [x] `registry/skills.json` — snapshot on 2026-03-14: 128 skills across 4 sources
  - 48 from `github/awesome-copilot`
  - 15 from `anthropics/skills`
  - 31 from `dotnet/skills` (MIT licensed)
  - 34 from `openai/skills` (`skills/.curated`, MIT licensed)
- [x] `.gitignore` — excludes binary and coverage files
- [x] `README.md` — comprehensive docs
- [x] `.github/copilot-instructions.md` — project conventions for Copilot
- [x] Module path fixed: `github.com/darrenr/skills-cli`
- [x] Removed `agent-customization` (VS Code built-in, no GitHub path)

## Post-v1 (Backlog)

- [ ] `cmd/sources.go` — sync registry from configured upstream sources
- [ ] **Registry auto-refresh** — two options under consideration:
  - **Option A** *(simpler)*: fetch `registry/skills.json` from `https://raw.githubusercontent.com/darrenr/skills-cli/main/registry/skills.json` and save to `~/.skills-cli/registry.json`. Stale check (24h TTL) already exists in the loader; `Loader.Save()` already exists but is never called. Add a background refresh on any command + `skills-cli registry refresh` for manual force. Skills lag behind a repo commit but work offline with the embedded seed.
  - **Option B** *(dynamic)*: query each source repo's API (`github/awesome-copilot`, `anthropics/skills`, `dotnet/skills`, `openai/skills`) directly and rebuild the registry. Always current with upstream. Needs rate-limit handling and is what `cmd/sources.go` was intended to be. Can be bolted on top of Option A later.
  - Recommended: ship Option A first; Option B as a follow-on.
- [ ] `skills-cli catalog` — categorized summary (count of skills per category)
- [ ] `skills-cli validate <path>` — validate a local SKILL.md against the spec
- [ ] `skills-cli init <name>` — scaffold a new skill from template
- [ ] Merge `develop` → `main` and tag `v1.0.0`

---

## Decision Log

| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-03-14 | Cobra + Viper for CLI framework | Most popular Go CLI combo, used by kubectl/gh/docker |
| 2026-03-14 | CLI commands only, no TUI | Keep it simple and focused |
| 2026-03-14 | No auth required by default | Public repos accessible via raw GitHub URLs; zero setup friction |
| 2026-03-14 | All standard install locations | Support .github/, .agents/, .claude/ for both project and personal scopes |
| 2026-03-14 | Bundled + cached registry | Ship embedded JSON, update remotely, cache locally with TTL |
| 2026-03-14 | Install 4 prerequisite skills | conventional-commit, make-skill-template, go-mcp-server-generator, gh-cli |
| 2026-03-14 | stdlib `text/tabwriter` instead of tablewriter | No external dep needed for simple column output |
| 2026-03-14 | Substring search instead of fuzzy (sahilm/fuzzy) | Simpler, no dependency, sufficient for skill names/descriptions |
| 2026-03-14 | No progress bar (schollz/progressbar) | Installs are a single file fetch — progress bar adds no value |
| 2026-03-14 | No `internal/config/` package | Viper used directly in `cmd/`; no config logic complex enough to warrant its own package |
| 2026-03-14 | `registry/skills.json` not `registry/sources.json` | Name reflects content (skills, not sources) |
| 2026-03-14 | Added `dotnet/skills` as third registry source | 31 high-quality MIT-licensed .NET skills from the official Microsoft team |
| 2026-03-14 | Added `openai/skills` as fourth registry source | 34 curated OpenAI skills from `skills/.curated` (name collisions skipped) |
| 2026-03-14 | `cmd/sources.go` deferred to v2 | Single curated registry covers v1 needs; dynamic source management adds complexity |

## Notes

_Capture observations, issues, and ideas as we go._

### 2026-03-14 — Planning Complete
- Researched Agent Skills ecosystem thoroughly
- Identified gap: no standalone Go CLI for browse → install workflow
- Key source repos: `github/awesome-copilot` (150+ skills), `anthropics/skills` (official reference)
- Existing tools: `skills-hub` (gh extension), `skrills` (Rust, validation/sync focus)
- `agentskills.io/specification` is the canonical SKILL.md format spec

### 2026-03-14 — v1 Build Complete
- All v1-scoped phases implemented and tested on `develop` branch
- Registry snapshot on 2026-03-14: 128 skills bundled from 4 sources (48 awesome-copilot, 15 anthropics/skills, 31 dotnet/skills, 34 openai/skills)
- All `internal/` packages ≥85% statement coverage with race detector
- Cross-compiled clean for darwin/linux/windows × amd64/arm64
- End-to-end install→update→remove verified in /tmp test project
- Module path corrected to `github.com/darrenr/skills-cli`
- `develop` is 13 commits ahead of `main`, ready to merge and tag v1.0.0
