# Claude Code instructions

This project keeps one set of rules for every AI assistant, in [`.ai/`](.ai/README.md).
The instructions are identical to [`AGENTS.md`](AGENTS.md) — follow that file.

**Read before writing code:**

1. [`.ai/project.md`](.ai/project.md) — what the service is and where its boundaries are.
2. [`.ai/structure.md`](.ai/structure.md) — the repository map and which layer owns what.
3. [`.ai/conventions.md`](.ai/conventions.md) — the rules of this codebase.
4. [`.ai/journal.md`](.ai/journal.md) — the last few entries, for the current state.

**The rules that break things when ignored:**

- Run `go test -count=1 ./tests/bdd/...` after any change to code, and report conflicts
  rather than rewriting expectations.
- `api/rest/v1/openapi.yaml` is the source of truth for the REST API. Edit it, regenerate,
  then implement the handler. Never hand-edit `internal/ui/api/v1/spec/*.gen.go`.
- The contract and everything under `internal/ui/web/static` are English, comments
  included. Go code and tests are commented in Russian.
- Comments explain why, never what, and never the change you just made.
- Append an entry to [`.ai/journal.md`](.ai/journal.md) when a task is done.
- Do not commit unless you were asked to.
