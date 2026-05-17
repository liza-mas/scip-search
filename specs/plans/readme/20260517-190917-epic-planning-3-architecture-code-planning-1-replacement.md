# Code Plan: Symbol Discovery Query

Status: draft

## Source Context

Based on:
- `README.md#scip-symbol-format`
- `README.md#what-is-scip-search`
- `specs/epics/readme/20260517-134857-epic-planning-3.md`
- `specs/stories/readme/20260517-134857-epic-planning-3/cap-001-symbol-name-discovery.md`
- Architecture reference `specs/arch-plan/readme/20260517-170650-epic-planning-3-architecture.md` read from merge commit `9d8fb6318ef527b937d651e29a6b1a44a62953bb` because the file is referenced by blackboard state but absent from this worktree checkout.
- Prior plan `specs/plans/readme/20260517-184518-epic-planning-3-architecture-code-planning-0.md`
- Traversal planning tasks `epic-planning-2-architecture-code-planning-0` and `epic-planning-2-architecture-code-planning-1`

## Planning Boundary

This plan covers only successful `symbols --name <name>` discovery behavior over an already-loaded shared traversal view.

Out of scope: CLI command routing, shared `--index` handling, shared index loading, stdout/stderr/status behavior, runtime failures, raw SCIP protobuf parsing, traversal construction, package discovery, reference and implementation queries, source-file reads, package registry lookups, dependency graph traversal, ranking, fuzzy search, regex search, glob search, semantic search, case-folded search, and cross-index matching.

## Architectural Direction

Add the symbol discovery behavior in `internal/query/discovery` as pure query logic over traversal facts and the `internal/scipmodel` identity helpers from the prior plan. The query should accept a traversal view and literal `name` text, construct symbol candidates from traversal-provided symbol facts, match the `name` text case-sensitively against traversal display names and descriptor text derived from the full SCIP symbol, attach available definition context from traversal definition occurrences, sort by the exact full `symbol`, and return a typed success payload with an explicit `symbols` collection.

The query package must not own process streams, CLI parsing, index loading, SCIP bytes, traversal construction, source-file reads, or package/reference/implementation behavior. If the downstream checkout lacks the shared traversal view or the `internal/scipmodel` helpers, the coder should mark the implementation task blocked rather than broadening this task to create those foundations.

## Planned Coding Tasks

### Task 1 - Implement symbol discovery query behavior

**desc:** `scip-search` users can run `symbols --name <name>` against a loaded traversal view and receive a successful `symbols` payload containing every literal case-sensitive match with exact symbol, package identity, match context, stable order, and optional definition location.

**done_when:** Unit and fixture-backed query tests in `internal/query/discovery` pass for `symbols --name Supervisor`, `symbols --name Run`, `symbols --name DoesNotExist`, and a literal pattern-character query, asserting the top-level `symbols` collection is present, matches are all and only literal case-sensitive display-name or descriptor substring matches, entries are sorted ascending by exact `symbol`, each non-empty entry preserves the full SCIP `symbol`, `scheme`, `packageManager`, `packageName`, `packageVersion`, `matchText`, and `matchSource`, available definition context is emitted as `definition.documentPath` plus SCIP `definition.range`, missing definition context is omitted without failing the query, and no CLI routing, `--index` handling, shared stdout/stderr failure behavior, raw SCIP parsing, package discovery, reference, implementation, source-file read, ranking, fuzzy, regex, semantic, or cross-index behavior is introduced.

**scope:** In scope: `internal/query/discovery` symbol query implementation, successful symbol result payload value types, literal case-sensitive matching over traversal-provided symbol display names and descriptor text, stable ascending sort by exact full `symbol`, package identity fields supplied by `internal/scipmodel`, optional definition context from traversal-provided definition occurrences, and colocated tests using deterministic traversal fixtures or fakes. Out of scope: CLI command routing, shared `--index` handling, shared index loading, stdout/stderr/status behavior, raw SCIP protobuf parsing or traversal construction, package discovery, reference and implementation queries, source-file reads, package registry or dependency graph behavior, ranking, fuzzy, regex, glob, semantic, case-folded, or cross-index matching, and final e2e/golden coverage owned by `epic-planning-3-architecture-code-planning-3-replacement`.

**spec_ref:** README.md#scip-symbol-format

**Planned files:**
- `internal/query/discovery/symbols.go`
- `internal/query/discovery/symbols_test.go`

**Implementation notes:**
- Consume traversal-provided symbol facts and definition occurrence context through the shared traversal interfaces; do not inspect raw SCIP protobufs or index paths in this package.
- Use `internal/scipmodel` to preserve the exact full symbol string and derive `scheme`, `packageManager`, `packageName`, `packageVersion`, descriptor text, display-name-preferred match text, and related package identity values.
- Treat matching as `strings.Contains`-style literal, case-sensitive substring matching against display name and descriptor text. Characters such as `.`, `*`, `[`, `]`, `(`, `)`, `/`, and `#` must remain ordinary characters.
- If display name and descriptor text both match, report display-name match context. If only descriptor text matches, report descriptor match context.
- Represent match context with `matchText` and `matchSource`, where `matchSource` is one of `displayName` or `descriptor`.
- Represent available definition context as an optional `definition` object containing `documentPath` and `range` copied from traversal-preserved SCIP data. Omit `definition` when traversal has no definition occurrence; do not synthesize empty strings, sentinel ranges, or source-file-derived locations.
- Sort the final `symbols` slice by exact full `symbol` ascending after filtering and before returning the payload.
- Validation command for the task: `go test ./internal/query/discovery`.

## Dependency Plan

Task 1 depends on the discovery identity model implementation from `epic-planning-3-architecture-code-planning-0-coding-0-replacement`.

Task 1 also requires the shared traversal view planned by `epic-planning-2-architecture-code-planning-0` and definition occurrence access planned by `epic-planning-2-architecture-code-planning-1`. Concrete traversal coding task IDs are not available in this planning context; if the downstream coder's checkout does not contain those traversal interfaces, the coder should mark blocked on the traversal foundation rather than implementing raw SCIP parsing or traversal construction here.

The separate discovery fixture and golden JSON plan `epic-planning-3-architecture-code-planning-3-replacement` owns final e2e/golden coverage after symbol and package query implementations exist.

## Shared-File Audit

Task 1 is the only planned task in this plan and owns all files listed in its scope. No sibling `depends_on` edge is required inside this output array.

Task 1 must not edit CLI runtime files, shared index-loading files, traversal construction files, `internal/scipmodel` identity helper files, package discovery files, reference or implementation query files, README files, or final golden JSON fixture files. Those are dependency or sibling scopes.

## Test Impact

Task 1 must add colocated tests in `internal/query/discovery/symbols_test.go` for literal matching, multi-match ordering, `Supervisor`, `Run`, no-match empty results, literal pattern characters, exact symbol preservation, package identity fields, match context, available definition context, and missing-definition behavior.

End-to-end/golden validation through the normal CLI loading path is covered by sibling task `epic-planning-3-architecture-code-planning-3-replacement`, because this plan excludes CLI routing, `--index`, stdout/stderr behavior, and final golden files.

## Doc Impact

No user-facing documentation update is required in this plan. README already documents the `symbols --index <index-path> --name <name>` command and full SCIP symbol behavior; this task implements that existing documented behavior without changing the public contract text. Query-specific specs are already present in the epic and story documents listed above.

## Spec Compliance Matrix

| # | Requirement | Source | Task(s) | Status |
|---|-------------|--------|---------|--------|
| 1 | `symbols --name <name>` returns every symbol whose display name or descriptor text contains the supplied name. | `cap-001-symbol-name-discovery.md` AC-001-1; architecture Scope 2 | Task 1 | Covered |
| 2 | Matching is literal and case-sensitive, with pattern characters treated as ordinary text. | `cap-001-symbol-name-discovery.md` NFR-000-1 and AC-001-5; parent epic ASM-000-3 | Task 1 | Covered |
| 3 | Ambiguous partial-name queries are successful multi-result responses rather than failures. | `cap-001-symbol-name-discovery.md` AC-001-2; parent epic CAP-001 description | Task 1 | Covered |
| 4 | No-match symbol queries return a successful payload with an explicit empty `symbols` collection. | `cap-001-symbol-name-discovery.md` AC-001-3; parent epic NFR-000-3 | Task 1 | Covered |
| 5 | Result ordering is deterministic and independent of traversal iteration order. | `cap-001-symbol-name-discovery.md` AC-001-4 and ASM-000-2; architecture Scope 2 | Task 1 | Covered |
| 6 | Symbol results are sorted ascending by exact full `symbol`. | Assigned done_when; architecture Symbol Discovery Query key decision | Task 1 | Covered |
| 7 | The successful payload exposes a top-level `symbols` collection. | Assigned done_when; `cap-001-symbol-name-discovery.md` Interface I-001-001 | Task 1 | Covered |
| 8 | Each non-empty symbol entry preserves the exact full SCIP `symbol` string from the index. | README.md#scip-symbol-format; `cap-001-symbol-name-discovery.md` AC-002-1; parent epic NFR-000-1 | Task 1 | Covered |
| 9 | Each non-empty symbol entry includes package identity fields derived from the full SCIP symbol. | `cap-001-symbol-name-discovery.md` AC-002-3; prior plan `20260517-184518-epic-planning-3-architecture-code-planning-0.md` | Task 1 | Covered |
| 10 | Each non-empty symbol entry identifies the match context that caused the result to match the supplied name. | `cap-001-symbol-name-discovery.md` AC-002-2; architecture Scope 2 | Task 1 | Covered |
| 11 | Available definition context is included with document path and SCIP range when traversal provides it. | `cap-001-symbol-name-discovery.md` AC-002-4; traversal story `CAP-002-01-source-range-normalization.md` | Task 1 | Covered |
| 12 | Missing definition context does not turn a successful symbol match into an error and is not filled by source-file reads or sentinels. | `cap-001-symbol-name-discovery.md` AC-002-4b and ASM-000-3; assigned scope | Task 1 | Covered |
| 13 | The returned full symbol remains suitable for later symbol-based commands without transformation. | README.md#scip-symbol-format; `cap-001-symbol-name-discovery.md` AC-002-5 | Task 1 | Covered |
| 14 | Implementation stays over the shared traversal view and does not own CLI routing, `--index`, shared failures, raw SCIP parsing, traversal construction, package discovery, reference or implementation behavior, ranking, fuzzy, regex, semantic, source-file, or cross-index behavior. | Assigned scope; parent epic Out of Scope; architecture Constraints | Task 1 | Covered |
| E2E | e2e test coverage for new behavior | Cross-cutting | Sibling task `epic-planning-3-architecture-code-planning-3-replacement`: final discovery fixture and golden JSON validation through normal runtime path | Covered |
| DOC | Documentation updates for changed behavior | Cross-cutting | N/A: README already documents the command and full-symbol behavior; this plan implements existing user-facing specs without changing the public contract text | N/A |

## Pre-Submit Validation Checklist

- Re-read this plan and verify the output JSON fields are character-identical to Task 1 `desc`, `done_when`, `scope`, and `spec_ref`.
- Run `jq . specs/plans/readme/20260517-190917-epic-planning-3-architecture-code-planning-1-replacement-output.json`.
- Search this plan for `Task 1`, `epic-planning-3-architecture-code-planning-3-replacement`, and `epic-planning-3-architecture-code-planning-0-coding-0-replacement` references and confirm the referenced responsibility or dependency is explicitly stated.
- Run pre-commit on the plan and output JSON files.
- Commit only the plan and output JSON artifacts.
