# AI context for Crypto Polymath

This folder is the shared context for AI coding assistants — Cursor, Codex, Claude Code
and anything else that reads a repository before touching it. One folder rather than one
file per tool: the rules are about the project, not about the assistant, and three copies
of the same rule drift apart within a week.

Every tool reaches this folder through a thin entry point of its own:

| Tool | Entry point |
| --- | --- |
| Cursor | `.cursor/context.md` (boot map) + `.cursor/rules/ai-context.mdc` (always applied) |
| Codex | `AGENTS.md` in the repository root |
| Claude Code | `CLAUDE.md` in the repository root |

Those files hold no full rule set themselves. They point here. Cursor’s
`.cursor/context.md` is a short orientation only — detailed rules stay in this folder.

## Reading order

1. `project.md` — what this service is, what it does and where its boundaries are.
2. `structure.md` — the repository map: which layer owns what and where new code belongs.
3. `conventions.md` — the rules of this codebase: language policy, contract-first API,
   testing, the frontend constraints. **Read this before writing code.**
4. `journal.md` — what changed recently and why. Read the last few entries to learn the
   current state; append an entry when you finish a task.

Deep architecture lives in [`docs/overview.md`](../docs/overview.md) (data flow, queues,
the calculation cascade, the database schema). This folder does not duplicate it — it
tells you when to go and read it.

## Hard rules, in short

These are the ones that break the build or the site when ignored. The details, and the
reasoning, are in `conventions.md`.

- Run `go test -count=1 ./tests/bdd/...` after any change to code. Report conflicts
  instead of quietly adapting the tests.
- Never hand-edit generated files (`internal/ui/api/v1/spec/*.gen.go`). Change
  `api/rest/v1/openapi.yaml` and regenerate.
- The API contract and everything under `internal/ui/web/static` are English only —
  comments included. Go code and tests are commented in Russian.
- Comments explain why, not what. No comment that narrates the change you just made.
- Append an entry to `journal.md` when a task is done.
