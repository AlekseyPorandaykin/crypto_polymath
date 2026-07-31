# Development rules

Rules that this codebase actually follows, with the reason attached. Where a rule is
enforced by a test, the test is named — if you disagree with a rule, the test is the thing
to argue with.

## 1. Language policy

Two audiences, two languages, and the split is deliberate.

| Where | Language | Enforced by |
| --- | --- | --- |
| Go code: comments, doc-comments, test names and failure messages | Russian | convention |
| BDD feature files (`tests/bdd/features/*.feature`) | Russian Gherkin (`# language: ru`) | convention |
| `api/rest/v1/openapi.yaml` — descriptions **and** YAML comments | English | `TestContractIsEnglish` |
| `internal/ui/web/static/**` — markup, UI strings, CSS and JS comments | English | `TestStaticFilesAreEnglish` |
| `domain.IndicatorDescriptions` | English | `TestIndicatorDescriptionsAreEnglish` |

The contract and the static files are read by strangers: clients generate code from the
contract, and anyone can open `/docs/api/openapi.yaml` or the page source. Go comments are
read by the people who maintain the service.

## 2. Comments

- A comment explains **why**, or a constraint the code cannot show. Never what the next
  line does.
- Never explain the change you are making. `// added validation here` is noise the moment
  the pull request merges.
- Match the density and tone of the file you are editing. Files in `core/` and
  `internal/ui/` open with a doc-comment explaining what the unit is for and which
  trade-offs it encodes — keep that shape when adding a file.

## 3. Testing

**After any change to code, run the BDD suite:**

```bash
go test -count=1 ./tests/bdd/...
```

If it fails, report the conflict instead of quietly rewriting the expectation: name the
scenario, the violated expectation and offer the choice between fixing the code and
updating the test.

Then extend the tests around what you changed:

- new behaviour in `core/` → a scenario in `tests/bdd/features/` plus unit tests;
- changed invariants → update the fuzz tests (`core/trading`);
- changed output format → refresh the golden files (`make test-golden-update`);
- changed contract, pages or static assets → extend `api/rest/v1/spec_test.go` or
  `internal/ui/web/pages_test.go`.

BDD coverage mirrors the `core/` packages: `core/trading` → `trading.feature`,
`simulate_addon.feature`, `stop_loss.feature`; `core/indicator/calculator` →
`indicators.feature`; `core/candle_indicator` → `heiken_ashi.feature`; `core/analysis` →
`analysis.feature`. Step definitions sit in `tests/bdd/*_test.go`.

Every test file carries a doc-comment saying why those tests exist and which class of
mistake they catch. Feature files carry Russian comments on each `Правило:` block
explaining the business idea. Failure messages state the consequence, not just the
mismatch — `"описание из %d символов будет обрезан в выдаче"` is useful, `"want != got"`
is not.

Useful targets (see `Makefile` for the rest): `make test-unit`, `make test-bdd`,
`make test-race`, `make test-cover`, `make test-fuzz`, `make bench-core`.

### Known pitfalls

- `go test ./internal/ui/...` can hang: tests under `internal/ui/daemon` reach the network.
  Run the packages you need instead (`./internal/ui/web/`, `./internal/ui/api/v1/impl/`).
- `go vet ./...` reports `unreachable code` in
  `internal/ui/daemon/candlesticks_consumer/app.go` — pre-existing, unrelated to your
  change. Vet the packages you touched.

## 4. The REST API is contract-first

`api/rest/v1/openapi.yaml` is the source of truth. The order is always:

1. Edit the contract.
2. Regenerate:

```bash
cd internal/ui/api/v1/spec
oapi-codegen -old-config-style -package spec -generate types,skip-prune -o types.gen.go ../../../../../api/rest/v1/openapi.yaml
oapi-codegen -old-config-style -package spec -generate server -o server.gen.go ../../../../../api/rest/v1/openapi.yaml
```

3. Implement the new method on `*impl.Handler` — `var _ spec.ServerInterface = (*Handler)(nil)`
   will not compile until you do.

Never hand-edit `*.gen.go`. Regeneration can surface drift that accumulated earlier
(fields turning into pointers because the contract made them optional); fix the handler,
do not bend the contract back.

Every operation documents `429`, and `TestEveryOperationDocumentsRateLimitResponse` counts
them, so a new endpoint without a `429` response fails the build.

## 5. What the API promises about tokens

- Anonymous callers: 10 requests per minute per IP address, shared across all endpoints,
  answered with `429` and a `Retry-After` header.
- Callers with `X-Token`: the quota of their pricing plan.

The token is not validated yet. That state stays inside `internal/ui/api/v1/impl` — the
contract and the interface describe tokens as issued with a plan, and
`TestContractDoesNotAdvertiseUncheckedToken` keeps wording like "not validated" or
"arbitrary string" out of the contract. The gap is in the code, and the code is what should
catch up.

## 6. Frontend constraints

- **The landing page carries exactly one script**: `static/js/contact.js`, for the contact
  dialog. Everything else on the page is static markup, because search crawlers and AI
  crawlers usually do not execute JavaScript — content built by a script does not exist for
  them. `TestLandingScriptsAreLimitedToContactForm` guards this.
- **No build step.** The tools page loads Vue and PrimeVue through an `importmap` with the
  version pinned in two places; both must match, or `provide/inject` inside PrimeVue breaks
  silently (`TestToolsPagePinsSameVueVersion`).
- PrimeVue runs against the runtime build without a template compiler, so components are
  written with `render()`/`h()`, not string templates.
- Numbers, dates and percentages go through `static/js/format.js`: numbers in `en-US`,
  dates in `en-GB` with a spelled-out month. Do not add a second formatting convention.
- Field limits in forms repeat the contract's `minLength`/`maxLength`
  (`TestContactFormLimitsMatchContract`): the browser should reject what the server would
  reject.
- SEO invariants are tested: the canonical host is the same everywhere, `/tools` stays
  `noindex`, structured data must match the visible text, and `robots.txt`, `sitemap.xml`
  and `llms.txt` are served. When you add a page, extend those tests.
- Colours and spacing come from the CSS variables in `static/css/app.css`. Charts read the
  same palette by value, since Chart.js cannot use CSS variables.

## 7. Formatting and dependencies

- `make go-fix` — `go mod tidy`, `gci`, `gofumpt`.
- `make go-linters` — vet, gofmt, goimports, golangci-lint.
- Dependencies are vendored: after changing `go.mod`, run `go mod tidy && go mod vendor`.

## 8. Definition of done

1. `go build ./...` passes.
2. Tests for the packages you touched pass, and `go test -count=1 ./tests/bdd/...` is green.
3. Tests and documentation are updated together with the code, not afterwards.
4. An entry is appended to [`journal.md`](journal.md).
5. Nothing is committed unless the user asked for a commit.
