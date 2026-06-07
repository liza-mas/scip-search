---
phase: COMMITTED
work_type: feature
rca_required: false
red_test_required: false
required_reviewers: [codex, claude]
plan_revision: 3
analysis_revision: 0
red_test_round: 0
code_review_round: 1
phase_updated_at: "2026-06-07T01:53:25Z"
worktree: /home/tangi/Workspace/scip-search/.worktrees/20260607-graph-and-impact
agents:
  doer:
    role: doer
    status: APPROVED
    last_seen: "2026-06-07T01:53:25Z"
    reviewed_analysis_revision: null
    analysis_verdict: null
    reviewed_plan_revision: null
    plan_verdict: null
    reviewed_red_test_round: null
    red_test_verdict: null
    reviewed_code_round: null
    code_verdict: null
  codex:
    role: reviewer
    status: APPROVED
    last_seen: "2026-06-07T01:33:18Z"
    reviewed_analysis_revision: null
    analysis_verdict: null
    reviewed_plan_revision: 3
    plan_verdict: APPROVED
    reviewed_red_test_round: null
    red_test_verdict: null
    reviewed_code_round: 1
    code_verdict: APPROVED
  claude:
    role: reviewer
    status: APPROVED
    last_seen: "2026-06-07T01:28:16Z"
    reviewed_analysis_revision: null
    analysis_verdict: null
    reviewed_plan_revision: 3
    plan_verdict: APPROVED
    reviewed_red_test_round: null
    red_test_verdict: null
    reviewed_code_round: 1
    code_verdict: APPROVED
---

# Adversarial Pairing Blackboard

## Goal

Implement the scip-search graph and impact upgrade described in the 2026-06-07 pairing discussion:

- Add a compact static graph capability around SCIP symbols, positioned as static dependency/reference graph plus best-effort call-like relationships, not a complete runtime call graph.
- Preserve the current one-shot CLI model: every query reads only the caller-supplied --index file and exits.
- Keep output Markdown/text-first and agent-oriented, with optional JSON only where useful for machine consumers.
- Avoid a centralized index or hidden index discovery.
- Treat the existing symbols/references/implementations/packages commands as lookup primitives, not a missing lookup layer.
- Evolve toward impact support that can identify changed symbols and likely review/test boundaries while making stale-index limits explicit.

## Evidence

- Current CLI already supports lookup primitives: symbols, references, implementations, and packages.
- README and v2 goal document emphasize deterministic Bash-first navigation, explicit --index input, one-line default output, and stale-index transparency.
- SCIP exposes symbol occurrences, roles, syntax kinds, relationships, ranges, and enclosing metadata, but these are static index facts and may vary by indexer/language.
- Pairing conclusion: implement graph/impact as static, compact, and honest about degraded precision when containment or relationship metadata is absent.
- Optional local smoke indexes exist at /home/tangi/Workspace/omni/go.scip, /home/tangi/Workspace/omni/python.scip, and /home/tangi/Workspace/omni/typescript.scip; they load with the current packages command but are external mutable artifacts, not deterministic test fixtures.

## Plan Revisions

### Plan revision 1 - static graph and symbol impact

#### Understanding

The existing commands already provide lookup primitives. The upgrade should add higher-level composition over those facts:

- graph: a static symbol-centered boundary with incoming references, outgoing contained dependencies, and explicit SCIP relationship edges.
- callers: a compact incoming-only view over the same graph data.
- callees: a compact outgoing-only view over the same graph data.
- impact: an initial symbol-based review/test boundary view. impact --diff stays out of this coding pass because reading Git/worktree diff state would weaken the current query-command invariant that the command reads only the supplied --index file.

#### Scope

Planned code/docs touchlist:

- internal/query/graph/graph.go: new graph query payloads and text/JSON formatting helpers.
- internal/query/graph/graph_test.go: behavior tests for incoming, outgoing, relationships, missing containment, stable order, and JSON shape.
- cmd/scip-search/main.go: register graph, callers, callees, and impact; parse flags; reuse existing symbol/name resolution.
- cmd/scip-search/main_test.go: command routing/help/usage/output tests for the new commands.
- internal/cli/runtime.go and internal/cli/runtime_test.go: documented command registry and help text updates.
- README.md: document the static graph/impact contract, examples, and non-runtime-call-graph caveat.

#### Behavior Contract

- All new commands require --index <index-path> through the existing runtime path.
- graph, callers, callees, and impact accept repeated --symbol and --name values and de-duplicate resolved exact symbols using existing CLI helper behavior.
- Output flags are --one-line and --json; default is compact Markdown-ish text, not raw JSON.
- No new command searches for, generates, updates, caches, or centralizes indexes.
- graph does not read source files. It uses SCIP occurrence ranges, roles, relationships, and enclosing ranges only.
- Incoming references reuse existing relationship-aware reference semantics so direct is_reference related symbols are included consistently with references.
- Outgoing dependencies are collected only from non-definition occurrences inside the target definition occurrence enclosing range. If no definition/enclosing range is available, output says outgoing is unavailable rather than guessing.
- Relationship edges are reported separately from occurrence-derived incoming/outgoing facts so type-definition/implementation/reference relationships are not confused with call-like dependencies.
- impact initially reports symbol impact from the index: definition, incoming dependents, outgoing dependencies, implementation relationships, likely tests inferred from indexed test-role occurrences and test-like paths. It must label these as static index-derived hints.
- impact --diff is documented as a follow-up design decision, not implemented in this pass.

#### Output Sketch

Default graph output should remain compact and agent-oriented:

```text
[SCIP Static Graph]
Symbol: Alpha
Defined: cmd/alpha.go:3:7

Incoming:
- pkg/Beta# pkg/beta.go:13:5

Outgoing:
- pkg/Beta# pkg/beta.go:13:5

Relationships:
- reference -> pkg/Beta#
- implementation -> pkg/Impl#
```

When containment is absent:

```text
Outgoing:
- unavailable: definition has no enclosing range in this index
```

#### Tests

- Unit tests use deterministic repo-owned SCIP fixtures, not /home/tangi/Workspace/omni/*.scip.
- Add paired positive/negative tests:
  - Incoming includes exact and direct reference-related symbols.
  - Outgoing includes non-definition occurrences inside enclosing range and excludes definitions/self occurrences outside the range.
  - Missing definition/enclosing range produces an explicit unavailable reason.
  - Relationship edges preserve all SCIP relationship flags and stable order.
  - JSON has explicit empty arrays and unavailable reason fields rather than omitted ambiguous state.
- CLI tests cover help text, routing, usage errors, name resolution, duplicate outputs, one-line output, and JSON output.

#### Optional Smoke Validation

After deterministic tests pass, use the external mutable indexes only as manual smoke evidence if available:

```bash
scip-search packages --index /home/tangi/Workspace/omni/go.scip
scip-search packages --index /home/tangi/Workspace/omni/python.scip
scip-search packages --index /home/tangi/Workspace/omni/typescript.scip
```

These must not become required tests because they are outside the repo and may change independently.

#### Validation

### Doer validation - code review round 1 - 2026-06-07T01:25:17Z

Final validation order followed plan revision 3:

```text
pre-commit run --files README.md cmd/scip-search/main.go cmd/scip-search/main_test.go internal/cli/runtime.go internal/cli/runtime_test.go internal/query/graph/graph.go internal/query/graph/graph_test.go
```

Result: passed all hooks, including gofmt, goimports, go vet, staticcheck, testhelper check, go mod tidy, and duplicate-code check.

```text
go test ./internal/query/graph ./cmd/scip-search ./internal/cli
```

Result: `Go test: 262 passed in 3 packages`.

```text
go test ./...
```

Result: `Go test: 418 passed in 17 packages`.

Optional external index smoke checks with the worktree code path:

```text
go run ./cmd/scip-search packages --index /home/tangi/Workspace/omni/go.scip
```

Result: `scip-go gomod github.com/tangi/diagnosis-design/cli 462588106cf2`.

```text
go run ./cmd/scip-search packages --index /home/tangi/Workspace/omni/python.scip
```

Result: `scip-python python omni-api 0.1.0`; `scip-python python python-stdlib 3.11`.

```text
go run ./cmd/scip-search packages --index /home/tangi/Workspace/omni/typescript.scip
```

Result: `scip-typescript npm frontend 0.0.0`.
 Commands

Planned validation order after implementation:

```bash
go test ./internal/query/graph ./cmd/scip-search ./internal/cli
go test ./...
pre-commit run --files <touched files>
```

Optional smoke validation, if the external indexes remain present:

```bash
scip-search packages --index /home/tangi/Workspace/omni/go.scip
scip-search packages --index /home/tangi/Workspace/omni/python.scip
scip-search packages --index /home/tangi/Workspace/omni/typescript.scip
```

#### Risks

- Static SCIP facts are not runtime call graph facts. Mitigation: command names and docs say static graph / dependencies / hints, not complete runtime calls.
- Outgoing precision depends on indexer containment metadata. Mitigation: explicit unavailable state when definition/enclosing range is absent.
- External indexes can contain generated/cache paths, as observed in the Go smoke probe. Mitigation: formatting treats paths as opaque SCIP document paths and does not assume worktree readability.
- impact --diff is strategically useful but not index-only. Mitigation: defer it behind a separate design gate rather than smuggling Git/worktree reads into this pass.



### Plan revision 2 - static graph and symbol impact

#### Changes from revision 1

Addresses reviewer blockers:

- Default output remains one-line and pipeline-friendly for every new query command.
- Multi-line Markdown output is available only behind an explicit `--markdown` flag.
- Outgoing containment is specified as definition-occurrence `EnclosingRange` bounding other occurrence `Range` values in the same document.
- `impact` receives explicit test-hint semantics and dedicated tests.
- Multi-symbol text and JSON layouts are specified.

#### Understanding

The existing commands already provide lookup primitives. The upgrade should add higher-level composition over SCIP facts without changing the one-shot index contract or the default line-oriented output contract:

- `graph`: a static symbol-centered boundary with incoming references, outgoing contained dependencies, and explicit SCIP relationship edges.
- `callers`: static incoming/dependents view over graph data. Help text and labels must say static SCIP references/dependents, not runtime callers.
- `callees`: static outgoing/dependencies view over graph data. Help text and labels must say static SCIP dependencies, not runtime callees.
- `impact`: symbol-based review/test boundary hints from the supplied SCIP index.

`impact --diff` remains out of scope for this coding pass because reading Git/worktree diff state would weaken the current query-command invariant that commands read only the supplied `--index` file.

#### Scope

Planned code/docs touchlist:

- `internal/query/graph/graph.go`: new graph/impact payloads and text/JSON formatting helpers.
- `internal/query/graph/graph_test.go`: behavior tests for incoming, outgoing, relationships, impact hints, missing containment, stable order, and JSON shape.
- `cmd/scip-search/main.go`: register `graph`, `callers`, `callees`, and `impact`; parse flags; reuse existing symbol/name resolution.
- `cmd/scip-search/main_test.go`: command routing/help/usage/output tests for the new commands.
- `internal/cli/runtime.go` and `internal/cli/runtime_test.go`: documented command registry and help text updates.
- `README.md`: document static graph/impact behavior, examples, default one-line output, explicit `--markdown`, and non-runtime-call-graph caveat.

Implementation should avoid broad refactors. If `cmd/scip-search/main.go` growth becomes hard to review, extract only a small shared graph command handler/parser rather than refactoring existing command families.

#### Output Contract

- Default output for `graph`, `callers`, `callees`, and `impact` is one-line text.
- Each default line is independently parseable and starts with `<path>:<line>:<column>` where a location exists, matching current command style.
- Use `?:0:0` for relationship-only or unavailable facts without a source location.
- `--json` returns one JSON value.
- `--markdown` returns the compact multi-line agent-oriented view. `--markdown`, `--one-line`, and `--json` are mutually exclusive.
- Multi-symbol default text concatenates independently parseable one-line rows, de-duplicated where rows are identical.
- Multi-symbol JSON follows the existing pattern: top-level `symbols` and per-symbol `queries`.
- Multi-symbol Markdown emits repeated sections separated by one blank line.

One-line examples:

```text
cmd/alpha.go:3:7 symbol="scip-go ... cmd/Alpha()." kind=definition
pkg/beta.go:13:5 symbol="scip-go ... pkg/Beta#" direction=incoming roles=8
pkg/beta.go:13:5 symbol="scip-go ... pkg/Beta#" direction=outgoing roles=8
?:0:0 symbol="scip-go ... pkg/Impl#" relationship=implementation direction=outgoing
?:0:0 symbol="scip-go ... cmd/Alpha()." direction=outgoing unavailable="definition has no enclosing range in this index"
```

Markdown example behind `--markdown`:

```text
[SCIP Static Graph]
Symbol: Alpha
Defined: cmd/alpha.go:3:7

Incoming:
- pkg/Beta# pkg/beta.go:13:5

Outgoing:
- pkg/Beta# pkg/beta.go:13:5

Relationships:
- reference -> pkg/Beta#
- implementation -> pkg/Impl#
```

#### Graph Semantics

- All new commands require `--index <index-path>` through the existing runtime path.
- `graph`, `callers`, `callees`, and `impact` accept repeated `--symbol` and `--name` values and de-duplicate resolved exact symbols using existing helper behavior.
- No new command searches for, generates, updates, caches, or centralizes indexes.
- `graph` does not read source files. It uses SCIP occurrence ranges, roles, relationships, and enclosing ranges only.
- Definition location is the first stable sorted definition occurrence for the queried symbol, if present.
- Incoming references reuse existing relationship-aware reference semantics so direct `is_reference` related symbols are included consistently with `references`.
- Outgoing dependencies use this exact algorithm:
  - Find the target symbol definition occurrence in deterministic order.
  - Require that definition occurrence to have `HasEnclosingRange == true` and a non-empty `EnclosingRange` that acts as the bounding box for the definition body/scope.
  - Scan occurrences in the same document.
  - Include non-definition occurrences whose `Range` falls within the definition occurrence `EnclosingRange`.
  - Exclude the queried symbol itself unless it appears as a non-definition recursive/self dependency inside the enclosing range.
  - If no definition or no usable enclosing range exists, emit an explicit unavailable reason instead of guessing.
- Relationship edges are reported separately from occurrence-derived incoming/outgoing facts so type-definition/implementation/reference relationships are not confused with call-like dependencies.

#### Impact Semantics

`impact` is a compact static impact-hint view over the same graph data:

- Definition: target definition location, if indexed.
- Review: incoming dependents plus implementation/type-definition/reference relationships targeting or owned by the symbol.
- Dependencies: outgoing contained dependencies from `graph`.
- Tests: deterministic hints from indexed facts, not filesystem reads.

Test-hint rules:

- SCIP role signal: any incoming, outgoing, or exact-symbol occurrence with `SymbolRole_Test` set is a test hint. The hint location is that occurrence location.
- Path signal: any incoming, outgoing, or exact-symbol occurrence whose document path matches one of these fixed language-agnostic patterns is a test hint:
  - contains `/test/` or `/tests/`
  - basename starts with `test_`
  - basename ends with `_test.go`
  - basename ends with `.test.ts`, `.test.tsx`, `.spec.ts`, `.spec.tsx`, `.test.js`, `.spec.js`
  - basename ends with `_test.py`
- Dedupe test hints by document path and location.
- If no role or path signal is available, output an explicit empty tests collection rather than inferring from naming similarity.
- JSON must distinguish `testRole` and `testPath` reasons so callers can judge confidence.

#### Tests

Unit tests use deterministic repo-owned fixtures, not `/home/tangi/Workspace/omni/*.scip`.

Graph tests:

- Incoming includes exact and direct reference-related symbols.
- Outgoing includes non-definition occurrences inside the definition occurrence enclosing range.
- Outgoing excludes definitions and occurrences outside the definition occurrence enclosing range.
- Missing definition and missing enclosing range each produce explicit unavailable reasons.
- Relationship edges preserve all SCIP relationship flags and stable order.
- JSON has explicit empty arrays and unavailable reason fields rather than omitted ambiguous state.

Impact tests:

- Definition, review, dependency, implementation, and relationship sections are populated from indexed facts.
- `SymbolRole_Test` occurrence creates a test hint with reason `testRole`.
- Fixed test-like path patterns create test hints with reason `testPath`.
- Non-test paths and absent role bits produce an explicit empty tests collection.
- Unavailable outgoing graph state is surfaced in impact rather than hidden.
- Multi-symbol impact JSON uses `symbols` plus per-symbol `queries`.

CLI tests:

- Help text and documented command registry include the new commands.
- Missing `--index`, unsupported flags, duplicate output flags, and missing symbol/name inputs fail with status `2` before handler execution where applicable.
- `--name` resolution, repeated symbols, default one-line output, `--json`, and `--markdown` are covered.
- `callers` and `callees` are verified as filtered graph views rather than separate traversal logic.

#### Optional Smoke Validation

After deterministic tests pass, use the external mutable indexes only as manual smoke evidence if available:

```bash
scip-search packages --index /home/tangi/Workspace/omni/go.scip
scip-search packages --index /home/tangi/Workspace/omni/python.scip
scip-search packages --index /home/tangi/Workspace/omni/typescript.scip
```

These must not become required tests because they are outside the repo and may change independently.

#### Validation Commands

Planned validation order after implementation:

```bash
go test ./internal/query/graph ./cmd/scip-search ./internal/cli
go test ./...
pre-commit run --files <touched files>
```

Optional smoke validation, if the external indexes remain present:

```bash
scip-search packages --index /home/tangi/Workspace/omni/go.scip
scip-search packages --index /home/tangi/Workspace/omni/python.scip
scip-search packages --index /home/tangi/Workspace/omni/typescript.scip
```

#### Risks

- Static SCIP facts are not runtime call graph facts. Mitigation: command names remain caveated in help/README and output labels say static dependencies/dependents.
- `callers`/`callees` names carry runtime-call implications. Mitigation: implement as aliases/views with explicit static labels; revisit names if reviewers still object.
- Outgoing precision depends on indexer containment metadata. Mitigation: explicit unavailable state when definition `EnclosingRange` is absent.
- Test hints are heuristics. Mitigation: reasons are explicit (`testRole`, `testPath`) and no hint is emitted when neither signal exists.
- External indexes can contain generated/cache paths. Mitigation: formatting treats paths as opaque SCIP document paths and does not assume worktree readability.
- `impact --diff` is strategically useful but not index-only. Mitigation: defer it behind a separate design gate.



### Plan revision 3 - validation-order amendment

Revision 3 inherits plan revision 2 in full, with one correction requested by codex and acknowledged by claude: validation order follows the Pairing contract DoD order.

Corrected validation order after implementation:

```bash
pre-commit run --files <touched files>
go test ./internal/query/graph ./cmd/scip-search ./internal/cli
go test ./...
```

Optional smoke validation remains after deterministic validation, if the external indexes are present:

```bash
scip-search packages --index /home/tangi/Workspace/omni/go.scip
scip-search packages --index /home/tangi/Workspace/omni/python.scip
scip-search packages --index /home/tangi/Workspace/omni/typescript.scip
```

No semantic, scope, output-contract, or test-coverage changes from revision 2.

## Plan Reviews


### codex review - plan revision 1 - CHANGES_REQUESTED - 2026-06-07T00:45:00Z

Reviewed target: plan revision 1.

Based on: blackboard plan revision 1, README.md runtime contract, specs/goals/20260519-v2.md NFR/FT-003, cmd/scip-search/main.go, internal/cli/runtime.go, references/implementations query patterns, and traversal facts/tests.

Findings:

1. Blocking: the proposed default graph text output conflicts with the existing CLI output contract. The plan says `Output flags are --one-line and --json; default is compact Markdown-ish text` and sketches a multi-line `[SCIP Static Graph]` response. README.md states that by default query commands write one-line text output, and specs/goals/20260519-v2.md NFR-000-2 / FT-003 require default text to remain line-oriented and Bash-pipeline friendly. Required revision: either make graph/callers/callees/impact default one-line output with one independently parseable result per line and reserve multi-line Markdown for an explicit flag, or explicitly record a spec/docs contract change that the human has approved before implementation.

2. Blocking: `impact` is in scope but its behavior is not testable enough in the plan. The tests section covers incoming/outgoing/relationships/JSON for graph-like behavior, but not impact-specific definition, incoming dependents, outgoing dependencies, implementation relationships, `SymbolRole_Test` handling, test-like path inference, unavailable states, empty results, or JSON shape. Required revision: add impact unit tests and CLI tests that pin those semantics, especially the distinction between SCIP test-role hints and path-name heuristics.

Non-blocking note: the `callers` / `callees` command names are usable only if help text and output labels keep saying these are static SCIP-derived dependency/reference views, not runtime call graph facts.

### claude review — plan revision 1 — CHANGES_REQUESTED — 2026-06-07T00:45:30Z

Reviewed target: plan revision 1.

Based on: blackboard plan revision 1, cmd/scip-search/main.go (529 lines), internal/cli/runtime.go (documentedCommands, helpText, Handler/Runtime), internal/traversal/facts.go (Occurrence.HasEnclosingRange, EnclosingRange, Symbol.EnclosingSymbol), internal/traversal/view.go (View methods), internal/query/references/references.go (referenceCandidateSymbols pattern).

I agree with codex's two blocking findings. Adding my own:

**Blocking findings:**

1. **Default output format contract violation (concurs with codex #1).** The plan proposes multi-line Markdown-ish default output, but the existing runtime contract (README.md "Runtime Contract" section) says "by default, query commands write one-line text output." Every existing command defaults to `--one-line`. The plan must either (a) keep `--one-line` as the default and add an explicit `--text` or `--markdown` flag for the multi-line format, or (b) get an explicit human decision to change the default output contract for these new commands. This needs resolution before implementation — it affects the help text, the README, and the test assertions.

2. **Outgoing dependency source ambiguity.** The behavior contract says "outgoing dependencies are collected only from non-definition occurrences inside the target definition occurrence enclosing range." But the traversal layer has two different containment concepts: `Occurrence.EnclosingRange` (a line range on each occurrence) and `Symbol.EnclosingSymbol` (a parent symbol identifier on `SymbolInformation`). The plan must specify which one is used for the "definition's enclosing range" that bounds outgoing collection. Based on codebase inspection: the definition occurrence's own `Range` field locates the definition, and `Occurrence.EnclosingRange` on that definition occurrence (if `HasEnclosingRange` is true) likely spans the function/method body — this is what you want. But `EnclosingRange` on *other* occurrences within the same document could also be used to check containment. The plan should be explicit: "find the definition occurrence for the target symbol, use its EnclosingRange as the bounding box, then collect all other occurrences in the same document whose Range falls within that bounding box." This eliminates implementation ambiguity.

3. **impact test-inference heuristic unspecified.** The plan says "likely tests inferred from indexed test-role occurrences and test-like paths" but doesn't define either concept. SCIP `SymbolRole` has a `Test` role (bit flag), but not all indexers emit it. "Test-like paths" could mean `_test.go` (Go), `test_*.py` (Python), `*.test.ts` (TypeScript) — but this project is language-agnostic. The plan needs to specify: (a) what SCIP role bits indicate test context, (b) what path patterns are used and whether they're language-aware or a fixed set, and (c) what happens when neither signal is available. Without this, impact's test boundary output is not deterministically testable.

**Non-blocking comments:**

4. **callers/callees naming.** Concurs with codex's non-blocking note. The names carry call-graph semantics this tool explicitly isn't. `dependents`/`dependencies` or `incoming`/`outgoing` would be more honest. Not blocking — the README caveat may suffice.

5. **main.go growth.** Adding 4 commands to a 529-line file will push it past 800 lines. The new commands reuse the same `--symbol`/`--name`/output-mode parsing via `parseAndResolveSymbolSet`. Consider whether the plan should note extracting shared flag parsing or whether that's an implementation detail to handle during coding.

6. **Multi-symbol output layout.** The output sketch shows single-symbol graph output. When `--name` resolves to multiple symbols, the text format for multiple `[SCIP Static Graph]` blocks should be specified. The JSON pattern (wrap in `queries` array) matches existing multi-symbol commands, but the text layout isn't addressed.

### codex review - plan revision 2 - CHANGES_REQUESTED - 2026-06-07T00:54:31Z

Reviewed target: plan revision 2.

Based on: blackboard plan revision 2, prior plan reviews, README.md runtime contract, specs/goals/20260519-v2.md, and the Pairing contract DoD order of operations.

Revision 2 resolves my previous blockers on default one-line output and impact-specific test semantics.

Blocking finding:

1. The validation command order still conflicts with the Pairing contract. The plan lists `go test ./internal/query/graph ./cmd/scip-search ./internal/cli`, then `go test ./...`, then `pre-commit run --files <touched files>`. The contract requires pre-commit on touched files before running tests or DONE. Required revision: put `pre-commit run --files <touched files>` before the Go test commands in the planned validation order, then keep the focused and full Go tests after it.

### claude review — plan revision 2 — APPROVED — 2026-06-07T00:55:45Z

Reviewed target: plan revision 2.

Based on: blackboard plan revision 2 diff against revision 1, verified against codebase (traversal/facts.go, traversal/view.go, references/references.go, cmd/scip-search/main.go, internal/cli/runtime.go).

All three blocking findings from my revision 1 review are resolved:

1. **Default output format** — resolved. Default is now one-line for all new commands. Multi-line Markdown is behind explicit `--markdown` flag, mutually exclusive with `--one-line` and `--json`. Preserves the existing CLI output contract.

2. **Outgoing dependency source** — resolved. Graph Semantics section specifies the exact algorithm: find definition occurrence, require `HasEnclosingRange == true` with non-empty `EnclosingRange` as bounding box, scan same-document occurrences, include non-definition occurrences whose `Range` falls within that box. Matches the traversal layer's `Occurrence.HasEnclosingRange`/`EnclosingRange` fields.

3. **impact test-inference heuristic** — resolved. Impact Semantics section specifies `SymbolRole_Test` as the role signal, a fixed set of language-agnostic path patterns (enumerated), deduplication by document path and location, explicit empty collection when neither signal exists, and JSON distinguishing `testRole`/`testPath` reasons. Deterministically testable.

Previous non-blocking comments also addressed: multi-symbol layout specified, main.go growth acknowledged, callers/callees naming kept with static-label caveat.

**On codex's r2 blocker (validation command order):** I agree the plan should put pre-commit before tests to match the DoD order of operations. However, I consider this a trivial editorial fix rather than a design-level blocker — the plan's substantive design, semantics, and test coverage are sound. Approving the plan design; the doer should fix the validation order in the next revision.

No new blocking findings.

### codex review - plan revision 3 - APPROVED - 2026-06-07T01:01:32Z

Reviewed target: plan revision 3.

Based on: blackboard plan revision 3 and prior review trail for revisions 1-2.

Verdict: APPROVED.

Revision 3 makes the requested validation-order correction: pre-commit on touched files now precedes the focused and full Go test commands. It inherits revision 2's resolved output contract, graph semantics, impact semantics, and test coverage plan. No remaining blocking findings from codex.

### claude review — plan revision 3 — APPROVED — 2026-06-07T01:02:06Z

Reviewed target: plan revision 3.

Revision 3 applies the validation-order correction I acknowledged in my r2 review. Pre-commit now precedes tests, matching the DoD order of operations. No semantic changes from revision 2 which I already approved. No new findings.


## Implementation Notes

### Doer implementation note - code review round 1 - 2026-06-07T01:25:17Z

Implemented in worktree `/home/tangi/Workspace/scip-search/.worktrees/20260607-graph-and-impact` on branch `work/20260607-graph-and-impact`.

Staged files:

- `internal/query/graph/graph.go`: static graph and impact payload/query/formatting implementation.
- `internal/query/graph/graph_test.go`: graph containment, unavailable-state, relationship, impact, and test-hint coverage.
- `cmd/scip-search/main.go`: `graph`, `callers`, `callees`, and `impact` command handlers with `--one-line`, `--json`, and `--markdown`.
- `cmd/scip-search/main_test.go`: production CLI coverage for graph/impact commands.
- `internal/cli/runtime.go` and `internal/cli/runtime_test.go`: documented command/help updates.
- `README.md`: runtime contract and static graph/impact documentation.

Staged diff stat:

```text
README.md                          |  27 +-
cmd/scip-search/main.go            | 259 ++++++++++++++
cmd/scip-search/main_test.go       |  96 +++++-
internal/cli/runtime.go            |  24 +-
internal/cli/runtime_test.go       |  48 ++-
internal/query/graph/graph.go      | 679 +++++++++++++++++++++++++++++++++++++
internal/query/graph/graph_test.go | 308 +++++++++++++++++
7 files changed, 1414 insertions(+), 27 deletions(-)
```


## Code Review Rounds


### codex code review - staged diff round 1 - APPROVED - 2026-06-07T01:33:18Z

Review: Standard - Approve

Reviewed target: staged diff round 1 in `/home/tangi/Workspace/scip-search/.worktrees/20260607-graph-and-impact`.

Blockers: None.
Concerns: None.
Suggestions: 0.

Overall: The staged implementation matches plan revision 3: default graph-family output remains one-line, Markdown is explicit, graph/impact stay index-only, unavailable outgoing state is explicit, and tests cover the central graph/impact semantics. No P0-P2 findings.

Validation run by codex:

```text
git diff --cached --check
```

Result: passed.

```text
go test ./internal/query/graph ./cmd/scip-search ./internal/cli
```

Result: `Go test: 262 passed in 3 packages`.

```text
go test ./...
```

Result: `Go test: 418 passed in 17 packages`.

Blast Radius: Medium - new CLI commands and query behavior, no schema/auth/data mutation.
Confidence: high for staged diff round 1.
Next step: wait for remaining required reviewer verdict, then doer can advance when the code gate predicate is satisfied.


### claude code review — staged diff round 1 — APPROVED — 2026-06-07T01:28:16Z

Reviewed target: staged diff, code_review_round 1.

Reviewed in Standard mode. 7 files, 1414 insertions, 27 deletions. Diff read in full.

Based on: git diff --cached in worktree, verified against plan revision 3 and existing codebase patterns.

**Verdict: APPROVED**

The implementation is clean, well-structured, and faithful to the approved plan. No P0-P2 issues found.

**P0 (Security):** No issues. No external input beyond the existing --index path, no new file reads, no injection surfaces.

**P1 (Correctness):** Implementation correctly follows the approved plan:
- firstDefinition (graph.go:242) finds the stable sorted definition occurrence.
- outgoingOccurrences (graph.go:289) checks HasEnclosingRange and EnclosingRange, scans same-document occurrences, filters by rangeWithin, excludes definitions — matches the plan's exact algorithm.
- testReasons (graph.go:384) checks SymbolRole_Test bit and isTestPath with the fixed language-agnostic pattern set from the plan.
- nonNilOccurrences/nonNilRelationships/nonNilTestHints ensure JSON arrays are never null.
- Callers/Callees are filtered views over Query output, not separate traversals — matches plan.

**P2 (Data integrity):** No issues. All output is read-only from the index. Deduplication via seen maps with stable keys. Deterministic sort on all output collections.

**P3 (Architecture):** Clean separation: internal/query/graph owns payloads and formatting, cmd/scip-search/main.go owns CLI routing. New graphHandler struct with kind enum avoids duplicating four handler types. Reuses resolveSymbolSet and parseGraphSymbolSetArgs follows existing patterns.

**P4 (Performance):** One note — outgoingOccurrences (graph.go:304) calls view.Occurrences() which returns all occurrences in the entire index, then filters by DocumentPath. For large indexes this is O(total occurrences) rather than O(document occurrences). The View type already has OccurrencesForSymbol but not a per-document accessor. This is not a bug and the current approach is correct, but could become a hot path for large indexes with many documents. Not blocking — the one-shot CLI model means this runs once and exits.

**P5 (Maintainability):** Well-structured. Test fixture at graph_test.go:162 is compact and covers the key scenarios (definition with enclosing range, definition without enclosing range, missing definition, test role, test path, outside-range occurrences). CLI tests verify one-line default, markdown, and JSON outputs.

No blocking findings.

## Validation

### Doer commit - 2026-06-07T01:53:25Z

Committed approved staged implementation in `/home/tangi/Workspace/scip-search/.worktrees/20260607-graph-and-impact`.

```text
ba314fb feat: add static graph and impact commands
```

## Decisions

- 2026-06-07T00:39:35Z: Human approved repair of the malformed blackboard, including restoring agents.codex.reviewed_code_round and the Markdown body.
- 2026-06-07T00:40:11Z: Human waived future approval gates until ready-to-commit while preserving reviewer verdict gates.
- 2026-06-07T01:51:56Z: Required reviewers approved code review round 1; doer moved phase to READY_TO_COMMIT and is awaiting human commit approval.
- 2026-06-07T01:53:25Z: Human approved commit; doer committed approved staged implementation as `ba314fb` and moved phase to COMMITTED. Next integration steps require separate approval: rebase worktree branch onto base, merge to base, then remove worktree and merged topic branch.
