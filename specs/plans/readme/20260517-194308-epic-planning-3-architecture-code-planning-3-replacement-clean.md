# Code Plan: Discovery Fixture and Golden JSON Validation

Status: draft

## Source Context

Based on:
- `README.md#scip-symbol-format`
- `README.md#what-is-scip-search`
- `specs/epics/readme/20260517-134857-epic-planning-3.md`
- `specs/stories/readme/20260517-134857-epic-planning-3/cap-001-symbol-name-discovery.md`
- `specs/stories/readme/20260517-134857-epic-planning-3/CAP-002-01-package-inventory.md`
- `specs/stories/readme/20260517-134857-epic-planning-3/CAP-002-02-package-prefix-filtering.md`
- `specs/stories/readme/20260517-134857-epic-planning-3/CAP-002-03-package-result-json-shape.md`
- `specs/stories/readme/20260517-134857-epic-planning-3/CAP-003-01-symbol-query-fixtures.md`
- `specs/stories/readme/20260517-134857-epic-planning-3/CAP-003-02-package-query-fixtures.md`
- `specs/stories/readme/20260517-134857-epic-planning-3/CAP-003-03-golden-json-validation.md`
- Architecture reference `specs/arch-plan/readme/20260517-170650-epic-planning-3-architecture.md` read from merge commit `9d8fb6318ef527b937d651e29a6b1a44a62953bb` because the file is referenced by blackboard state but absent from this worktree checkout.
- Prior plan `specs/plans/readme/20260517-190917-epic-planning-3-architecture-code-planning-1-replacement.md`
- Prior plan `specs/plans/readme/20260517-191918-epic-planning-3-architecture-code-planning-2-replacement.md`

## Planning Boundary

This plan covers query-specific deterministic discovery fixture data and successful golden JSON validation for `symbols --name` and `packages` commands after the symbol query, package query, shared runtime, and traversal foundations exist.

Out of scope: shared command routing implementation, shared `--index` handling implementation, shared index-loading or malformed-index failure fixtures, raw SCIP traversal fixture construction ownership, symbol query implementation, package query implementation, reference or implementation query fixtures, large real-world fixtures, performance benchmarks, external indexer installation, ctags fallback data, alternate output formats, object field order assertions, source-file reads, package registry behavior, dependency graph behavior, fuzzy, regex, glob, semantic, case-folded, or cross-index matching.

## Architectural Direction

Add the final discovery validation as a thin integration layer over the existing runtime, loaded-index, traversal, and discovery query components. The fixture should be deterministic and schema-valid, but query validation must not own raw traversal fixture construction if the traversal epic has provided shared fixture mechanics. Golden tests should invoke the normal loaded-index command path, capture successful JSON, parse it, and compare JSON values against command-specific expected files while separately asserting array ordering.

The symbol and package golden cases should share fixture data when that keeps the fixture small, but expected JSON files and assertions stay command-specific. This keeps the cross-command validation cohesive without coupling the `symbols` and `packages` payload contracts.

If a downstream checkout lacks the normal loaded-index command path, traversal fixture mechanics, symbol query implementation, or package query implementation, the coder should mark the affected validation task blocked rather than implementing those foundations here.

## Planned Coding Tasks

### Task 1 - Add shared discovery fixture support

**desc:** `scip-search` maintainers have a reusable deterministic discovery SCIP fixture and golden-test harness input that can be loaded through the normal shared SCIP loading and traversal path for downstream symbol and package discovery validation.

**done_when:** Fixture support tests in the discovery validation package pass and prove the query-specific fixture is loaded through the normal shared SCIP loading and traversal path, reuses shared traversal fixture mechanics when available, exposes at least the full symbols `scip-go gomod github.com/liza-mas/liza . supervisor/Supervisor#`, `scip-go gomod github.com/liza-mas/liza . supervisor/SupervisorConfig#`, `scip-go gomod github.com/liza-mas/liza . supervisor/Run().`, and `scip-go gomod github.com/liza-mas/liza . agent/SupervisorAgent#`, includes repeated and distinct package identities `scip-go gomod github.com/liza-mas/liza .`, `scip-go gomod github.com/liza-mas/scip-search .`, and `scip-go gomod github.com/sourcegraph/scip-bindings .`, contains a non-package-name `liza-mas` descriptor case, remains small and deterministic, and introduces no command routing, shared `--index` handling, shared runtime failure fixtures, raw traversal fixture ownership, reference or implementation query data, source-file reads, external indexer installation, large real-world fixtures, ctags fallback data, or alternate output formats.

**scope:** In scope: query-specific deterministic discovery fixture data, fixture access helpers needed by command-specific golden tests, reuse of shared traversal fixture mechanics when present, fixture support validation proving the fixture can be loaded through the normal shared SCIP loading and traversal path, symbol data required by `Supervisor`, `Run`, and `DoesNotExist` cases, package data required by unfiltered, `github.com/liza-mas/`, `github.com/liza-mas/scip-search`, and `github.com/no-match/` cases, duplicate package identity data, distinct package identity data, and non-package-name prefix-negative data. Out of scope: command-specific golden JSON assertions, shared command routing implementation, shared `--index` handling implementation, shared index-loading or malformed-index failure fixtures, raw SCIP traversal fixture construction ownership, symbol query implementation, package query implementation, reference or implementation query fixtures, large real-world fixtures, performance benchmarks, external indexer installation, ctags fallback data, alternate output formats, object field order assertions, source-file reads, package registry behavior, dependency graph behavior, fuzzy, regex, glob, semantic, case-folded, or cross-index matching.

**spec_ref:** README.md#scip-symbol-format

**Planned files:**
- `internal/query/discovery/testdata/discovery.scip`
- `internal/query/discovery/discovery_fixture_test.go`

**Implementation notes:**
- Prefer the shared traversal fixture writer/loader from epic-planning-2 if it exists. Do not hand-roll protobuf construction when traversal-owned helpers are available.
- Keep the fixture artificial but schema-valid and deterministic; no external indexer invocation is required.
- Ensure package-only negative data proves prefix text outside `packageName` does not create package matches.

**Validation command:** `go test ./internal/query/discovery/...`

### Task 2 - Add symbol discovery golden validation

**desc:** `scip-search` maintainers can validate successful `symbols --name` discovery responses against command-specific golden JSON through the normal loaded-index command path.

**done_when:** Golden validation tests pass for `symbols --name Supervisor`, `symbols --name Run`, and `symbols --name DoesNotExist` against the deterministic discovery SCIP fixture, invoking the normal loaded-index command path and comparing parsed JSON values to command-specific golden files; the tests assert the top-level `symbols` collection, exact full SCIP `symbol` preservation, required `scheme`, `packageManager`, `packageName`, `packageVersion`, `matchText`, `matchSource`, and available `definition.documentPath` plus SCIP `definition.range` fields for non-empty entries, stable ascending array order by exact `symbol`, inclusion of every `Supervisor` display or descriptor match including ambiguous matches, exclusion of unrelated symbols from the `Run` case, explicit empty `symbols` for `DoesNotExist`, and no assertions about shared runtime failures, object field order, package payloads, references, implementations, source-file reads, fuzzy, regex, glob, semantic, case-folded, or cross-index behavior.

**scope:** In scope: symbol discovery golden JSON files, fixture-backed tests for `symbols --name Supervisor`, `symbols --name Run`, and `symbols --name DoesNotExist`, invocation through the normal loaded-index command path, parsed JSON value comparison, stable `symbols` array ordering assertions, exact full SCIP symbol preservation assertions, package identity field assertions, match context assertions, available definition context assertions, explicit empty symbol result assertions, and reuse of Task 1 fixture support. Out of scope: package golden cases, shared command routing implementation, shared `--index` handling implementation, shared index-loading or malformed-index failure fixtures, raw SCIP traversal fixture construction ownership, symbol query implementation, reference or implementation query fixtures, large real-world fixtures, performance benchmarks, external indexer installation, ctags fallback data, alternate output formats, object field order assertions, source-file reads, package registry behavior, dependency graph behavior, fuzzy, regex, glob, semantic, case-folded, or cross-index matching.

**spec_ref:** README.md#scip-symbol-format

**Planned files:**
- `internal/query/discovery/symbols_golden_test.go`
- `internal/query/discovery/testdata/golden/symbols-supervisor.json`
- `internal/query/discovery/testdata/golden/symbols-run.json`
- `internal/query/discovery/testdata/golden/symbols-does-not-exist.json`

**Implementation notes:**
- Compare parsed JSON values, not raw JSON text or object field order.
- Assert array order explicitly after decoding so stable ordering remains a tested contract.
- Use the real command/runtime path available in the repository rather than calling the pure symbol query function directly.

**Validation command:** `go test ./internal/query/discovery/...`

### Task 3 - Add package discovery golden validation

**desc:** `scip-search` maintainers can validate successful `packages` discovery responses against command-specific golden JSON through the normal loaded-index command path.

**done_when:** Golden validation tests pass for unfiltered `packages`, `packages --prefix github.com/liza-mas/`, `packages --prefix github.com/liza-mas/scip-search`, and `packages --prefix github.com/no-match/` against the deterministic discovery SCIP fixture, invoking the normal loaded-index command path and comparing parsed JSON values to command-specific golden files; the tests assert the top-level `packages` collection, unfiltered inclusion of each distinct full package identity exactly once despite duplicate symbols, stable ascending array order by exact `packageKey`, exact `scheme`, `packageManager`, `packageName`, `packageVersion`, and `packageKey` fields for non-empty entries, literal case-sensitive prefix filtering over `packageName` only, exclusion when prefix text appears only in scheme, package manager, package version, descriptor, symbol name, or document path, exact prefix narrowing to `github.com/liza-mas/scip-search`, explicit empty `packages` for `github.com/no-match/`, and no assertions about shared runtime failures, object field order, symbol payloads, references, implementations, source-file reads, package registry behavior, dependency graph behavior, version resolution, regex, glob, fuzzy, semantic, case-folded, or cross-index behavior.

**scope:** In scope: package discovery golden JSON files, fixture-backed tests for unfiltered `packages`, `packages --prefix github.com/liza-mas/`, `packages --prefix github.com/liza-mas/scip-search`, and `packages --prefix github.com/no-match/`, invocation through the normal loaded-index command path, parsed JSON value comparison, stable `packages` array ordering assertions, exact package identity payload assertions, de-duplication assertions, literal package-name prefix filtering assertions, exact package-prefix narrowing assertions, no-match package prefix assertions, and reuse of Task 1 fixture support. Out of scope: symbol golden cases, shared command routing implementation, shared `--index` handling implementation, shared index-loading or malformed-index failure fixtures, raw SCIP traversal fixture construction ownership, package query implementation, reference or implementation query fixtures, large real-world fixtures, performance benchmarks, external indexer installation, ctags fallback data, alternate output formats, object field order assertions, source-file reads, package registry behavior, dependency graph behavior, version resolution beyond the SCIP package version component, descriptor, scheme, package-manager, package-version, document-path, regex, glob, fuzzy, semantic, case-folded, or cross-index filtering.

**spec_ref:** README.md#scip-symbol-format

**Planned files:**
- `internal/query/discovery/packages_golden_test.go`
- `internal/query/discovery/testdata/golden/packages-all.json`
- `internal/query/discovery/testdata/golden/packages-liza-mas.json`
- `internal/query/discovery/testdata/golden/packages-scip-search.json`
- `internal/query/discovery/testdata/golden/packages-no-match.json`

**Implementation notes:**
- Compare parsed JSON values, not raw JSON text or object field order.
- Assert array order explicitly after decoding so package ordering remains observable.
- Do not add dependency graph, package registry, or source-file fixtures.

**Validation command:** `go test ./internal/query/discovery/...`

## Dependency Plan

Task 1 depends on the shared runtime loaded-index orchestration from `epic-planning-1-architecture-code-planning-0-coding-2`, the traversal view from `epic-planning-2-architecture-code-planning-0-coding-0`, and shared traversal fixture planning from `epic-planning-2-architecture-code-planning-2`.

Task 2 depends on Task 1 and on the symbol query implementation task `epic-planning-3-architecture-code-planning-1-replacement-coding-0-after-traversal`.

Task 3 depends on Task 1 and on the package query implementation task `epic-planning-3-architecture-code-planning-2-replacement-coding-0-after-traversal`.

Task 2 and Task 3 can run in parallel after Task 1 if their planned file scopes remain disjoint. If implementation discovers a shared golden helper file is needed beyond Task 1, the later coder should stop and request a dependency/order update rather than creating parallel shared-file edits.

## Shared-File Audit

Task 1 owns the shared discovery fixture and fixture support files.

Task 2 owns only symbol golden test files and symbol golden JSON files. It depends on Task 1 because it reads the shared fixture support.

Task 3 owns only package golden test files and package golden JSON files. It depends on Task 1 because it reads the shared fixture support.

No sibling dependency is required between Task 2 and Task 3 under the planned file split. Both must avoid editing each other's golden JSON files or query implementation files.

## Test Impact

Task 1 adds fixture support validation through the normal shared SCIP loading and traversal path.

Task 2 adds symbol command golden tests for `symbols --name Supervisor`, `symbols --name Run`, and `symbols --name DoesNotExist`.

Task 3 adds package command golden tests for unfiltered `packages`, `packages --prefix github.com/liza-mas/`, `packages --prefix github.com/liza-mas/scip-search`, and `packages --prefix github.com/no-match/`.

## Doc Impact

No README or user-facing documentation update is required. This plan adds maintainer validation for behavior already documented by README and the story/spec artifacts; it does not change the public CLI contract.

## Spec Compliance Matrix

| # | Requirement | Source | Task(s) | Status |
|---|-------------|--------|---------|--------|
| 1 | Discovery fixture data is deterministic, small, schema-valid, and runs through the same SCIP loading and traversal path as commands. | EP-003 NFR-000-4; CAP-003-01 NFR-000-1/NFR-000-3; CAP-003-02 NFR-000-1/NFR-000-3; CAP-003-03 NFR-000-3/NFR-000-4 | Task 1 | Covered |
| 2 | Fixture validation reuses shared traversal fixture mechanics when present and does not own raw traversal fixture construction. | Assigned scope; EP-003 CAP-003 Out of Scope; architecture Scope 4 | Task 1 | Covered |
| 3 | Symbol fixture includes `Supervisor`, `SupervisorConfig`, `Run`, and `SupervisorAgent` full SCIP symbols for exact-looking, substring, and ambiguous matches. | CAP-003-01 ST-001 AC-001-1 | Task 1, Task 2 | Covered |
| 4 | `symbols --name Supervisor` golden validation includes every matching display or descriptor result, including ambiguous matches, with no ambiguity failure. | CAP-003-01 ST-001 AC-001-2; CAP-003-03 ST-001 AC-001-1; assigned done_when | Task 2 | Covered |
| 5 | `symbols --name Run` golden validation proves partial literal matching and excludes unrelated `Supervisor` symbols that only share a package. | CAP-003-01 ST-001 AC-001-3; CAP-003-03 ST-001 AC-001-2; assigned done_when | Task 2 | Covered |
| 6 | `symbols --name DoesNotExist` golden validation returns a successful payload with an explicit empty `symbols` collection. | CAP-003-01 ST-002 AC-002-1 through AC-002-3; CAP-003-03 ST-001 AC-001-3; assigned done_when | Task 2 | Covered |
| 7 | Non-empty symbol golden entries preserve exact full SCIP `symbol` strings and command-specific match context fields. | CAP-003-01 ST-001 AC-001-4; cap-001-symbol-name-discovery ST-002 AC-002-1/AC-002-2; assigned scope | Task 2 | Covered |
| 8 | Symbol golden validation asserts stable ascending array order by exact full `symbol`. | CAP-003-01 ST-001 AC-001-5; CAP-003-03 ST-001 AC-001-4; assigned done_when | Task 2 | Covered |
| 9 | Symbol golden validation asserts command-specific payload fields without asserting shared runtime failures, package payloads, references, implementations, or object field order. | CAP-003-03 ST-001 AC-001-5; CAP-003-03 ASM-000-1; assigned done_when | Task 2 | Covered |
| 10 | Package fixture includes repeated `github.com/liza-mas/liza` symbols and distinct `github.com/liza-mas/scip-search` and `github.com/sourcegraph/scip-bindings` package identities. | CAP-003-02 ST-001 AC-001-1; assigned scope | Task 1, Task 3 | Covered |
| 11 | Unfiltered `packages` golden validation includes every distinct package identity exactly once despite duplicate symbols. | CAP-003-02 ST-001 AC-001-2; CAP-003-03 ST-002 AC-002-1/AC-002-2; assigned done_when | Task 3 | Covered |
| 12 | Non-empty package golden entries assert exact `scheme`, `packageManager`, `packageName`, `packageVersion`, and `packageKey` fields. | CAP-003-02 ST-001 AC-001-4; CAP-003-03 ST-002 AC-002-6; assigned done_when | Task 3 | Covered |
| 13 | Package golden validation asserts stable ascending array order by observable `packageKey`. | CAP-003-02 ST-001 AC-001-5; CAP-003-03 ST-002 AC-002-5; assigned done_when | Task 3 | Covered |
| 14 | `packages --prefix github.com/liza-mas/` golden validation includes only the two `github.com/liza-mas/` package identities. | CAP-003-02 ST-002 AC-002-1; CAP-003-03 ST-002 AC-002-3; assigned done_when | Task 3 | Covered |
| 15 | `packages --prefix github.com/liza-mas/scip-search` golden validation narrows to the exact matching package identity. | CAP-003-02 ST-002 AC-002-2; assigned done_when | Task 3 | Covered |
| 16 | Prefix text outside `packageName` does not cause a package match. | CAP-003-02 ST-002 AC-002-3; CAP-002-02 ST-001 AC-001-3; assigned done_when | Task 1, Task 3 | Covered |
| 17 | `packages --prefix github.com/no-match/` golden validation returns a successful payload with an explicit empty `packages` collection. | CAP-003-02 ST-002 AC-002-4; CAP-003-03 ST-002 AC-002-4; assigned done_when | Task 3 | Covered |
| 18 | Filtered package golden cases preserve stable package ordering. | CAP-003-02 ST-002 AC-002-5; assigned done_when | Task 3 | Covered |
| 19 | Golden validation compares parsed JSON values and must not assert object field order. | CAP-003-03 ASM-000-1; assigned done_when | Task 2, Task 3 | Covered |
| 20 | Validation excludes shared runtime failure fixtures, reference and implementation query fixtures, source-file reads, registry/dependency graph behavior, unsupported match modes, large real-world fixtures, external indexers, ctags fallback data, and alternate output formats. | Assigned scope; EP-003 CAP-003 Out of Scope; CAP-003-01 Out of Scope; CAP-003-02 Out of Scope; CAP-003-03 Out of Scope | Task 1, Task 2, Task 3 | Covered |
| E2E | e2e test coverage for new behavior | Cross-cutting | Task 2 and Task 3: fixture-backed command-path golden validation for successful discovery commands | Covered |
| DOC | Documentation updates for changed behavior | Cross-cutting | N/A: maintainer validation only; README and story specs already document the discovery command behavior and this plan does not change the user-facing contract | N/A |

## Pre-Submit Validation Checklist

- Re-read this plan and verify the output JSON fields are character-identical to each task's `desc`, `done_when`, `scope`, and `spec_ref`.
- Run `jq . specs/plans/readme/20260517-194308-epic-planning-3-architecture-code-planning-3-replacement-clean-output.json`.
- Search this plan for `Task 1`, `Task 2`, `Task 3`, `epic-planning-3-architecture-code-planning-1-replacement-coding-0-after-traversal`, and `epic-planning-3-architecture-code-planning-2-replacement-coding-0-after-traversal` references and confirm the referenced responsibility, dependency, or exclusion is explicitly stated.
- Verify Task 2 and Task 3 share only Task 1 fixture support and do not share planned writable files with each other.
- Run pre-commit on the plan and output JSON files.
- Commit only the plan and output JSON artifacts.
