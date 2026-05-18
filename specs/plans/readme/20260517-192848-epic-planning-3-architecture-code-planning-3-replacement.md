# Code Plan: Discovery Fixture and Golden JSON Validation

Status: draft

## Source Context

Based on:
- `README.md#scip-symbol-format`
- `README.md#what-is-scip-search`
- `specs/epics/readme/20260517-134857-epic-planning-3.md`
- `specs/stories/readme/20260517-134857-epic-planning-3/CAP-003-01-symbol-query-fixtures.md`
- `specs/stories/readme/20260517-134857-epic-planning-3/CAP-003-02-package-query-fixtures.md`
- `specs/stories/readme/20260517-134857-epic-planning-3/CAP-003-03-golden-json-validation.md`
- Consistency references `specs/stories/readme/20260517-134857-epic-planning-3/cap-001-symbol-name-discovery.md`, `CAP-002-01-package-inventory.md`, `CAP-002-02-package-prefix-filtering.md`, and `CAP-002-03-package-result-json-shape.md`
- Architecture reference `specs/arch-plan/readme/20260517-170650-epic-planning-3-architecture.md` read from merge commit `9d8fb6318ef527b937d651e29a6b1a44a62953bb` because the file is referenced by blackboard state but absent from that planning worktree checkout.
- Prior plan `specs/plans/readme/20260517-190917-epic-planning-3-architecture-code-planning-1-replacement.md`
- Prior plan `specs/plans/readme/20260517-191918-epic-planning-3-architecture-code-planning-2-replacement.md`

## Planning Boundary

This plan covers query-specific discovery fixture data, golden JSON cases, and fixture-backed validation tests for successful `symbols --name` and `packages` discovery behavior.

Out of scope: shared runtime failure fixtures, shared traversal fixture construction ownership, raw SCIP traversal construction, query implementation behavior, reference and implementation query fixtures, large real-world fixtures, performance benchmarks, external indexer installation, ctags fallback data, alternate output formats, and object field order assertions.

## Architectural Direction

Add one integration-style validation task that uses shared fixture construction and the normal loaded-index command path instead of directly calling symbol or package query functions with in-memory fakes. The task should provide a small deterministic discovery fixture containing the symbol and package identities required by CAP-003, command-specific expected JSON cases, and tests that invoke the discovery commands through the same runtime loading and traversal path used by the CLI.

The fixture data may physically share one schema-valid SCIP fixture source for symbol and package cases, but the golden expectations must remain command-specific. Validation should parse stdout as JSON and compare JSON values, while separately asserting array ordering for `symbols` and `packages`. Tests must not depend on object field order and must not assert shared runtime failures, stderr diagnostics, reference or implementation payloads, or alternate output formats.

The task depends on the symbol discovery query and package discovery query implementations. It also requires the shared runtime loading boundary and traversal fixture mechanics from sibling epics; if those foundations are absent in the downstream checkout, the coder should mark the task blocked rather than creating raw SCIP loading, traversal construction, or shared runtime infrastructure here.

## Planned Coding Tasks

### Task 1 - Add discovery fixture and golden JSON validation

**desc:** `scip-search` maintainers can validate symbol and package discovery through normal SCIP loading and traversal using deterministic fixtures and command-specific golden JSON cases for matching, filtering, empty results, identities, and stable ordering.

**done_when:** Fixture-backed validation tests pass for `symbols --name Supervisor`, `symbols --name Run`, `symbols --name DoesNotExist`, unfiltered `packages`, `packages --prefix github.com/liza-mas/`, `packages --prefix github.com/liza-mas/scip-search`, and `packages --prefix github.com/no-match/`, invoking the normal SCIP loading and traversal path, parsing stdout JSON into values, comparing command-specific golden payloads, asserting stable `symbols` array order by exact full `symbol`, stable `packages` array order by exact `packageKey`, exact full symbol preservation, symbol match context and package identity fields, package de-duplication, package prefix filtering over `packageName` only, explicit empty result collections, and no assertions for shared runtime failures, stderr diagnostics, reference or implementation payloads, alternate output formats, or JSON object field order.

**scope:** In scope: query-specific deterministic discovery fixture data, command-specific golden JSON files for successful symbol and package discovery cases, validation tests that invoke the normal loaded-index command path and compare parsed JSON values, symbol cases for `Supervisor`, `Run`, and `DoesNotExist`, package cases for unfiltered listing, `github.com/liza-mas/`, `github.com/liza-mas/scip-search`, and `github.com/no-match/`, stable array ordering assertions, exact symbol and package identity payload assertions, and reuse of shared traversal fixture mechanics when present. Out of scope: shared command routing implementation, shared `--index` handling implementation, shared index-loading or malformed-index failure fixtures, raw SCIP traversal fixture construction ownership, symbol query implementation, package query implementation, reference or implementation query fixtures, large real-world fixtures, performance benchmarks, external indexer installation, ctags fallback data, alternate output formats, object field order assertions, source-file reads, package registry behavior, dependency graph behavior, fuzzy, regex, glob, semantic, case-folded, or cross-index matching.

**spec_ref:** README.md#scip-symbol-format

**Planned files:**
- `internal/query/discovery/discovery_golden_test.go`
- `internal/query/discovery/testdata/discovery.scip` or the repository's shared traversal fixture source equivalent
- `internal/query/discovery/testdata/golden/symbols-supervisor.json`
- `internal/query/discovery/testdata/golden/symbols-run.json`
- `internal/query/discovery/testdata/golden/symbols-does-not-exist.json`
- `internal/query/discovery/testdata/golden/packages-all.json`
- `internal/query/discovery/testdata/golden/packages-prefix-liza-mas.json`
- `internal/query/discovery/testdata/golden/packages-prefix-scip-search.json`
- `internal/query/discovery/testdata/golden/packages-prefix-no-match.json`

**Implementation notes:**
- Build or reuse a deterministic schema-valid SCIP fixture through the shared traversal fixture mechanism owned by epic-planning-2. Do not create a parallel raw SCIP parser, traversal builder, index loader, or CLI runtime in this task.
- Ensure the fixture contains at least these full symbols: `scip-go gomod github.com/liza-mas/liza . supervisor/Supervisor#`, `scip-go gomod github.com/liza-mas/liza . supervisor/SupervisorConfig#`, `scip-go gomod github.com/liza-mas/liza . supervisor/Run().`, and `scip-go gomod github.com/liza-mas/liza . agent/SupervisorAgent#`.
- Ensure the package fixture surface contains at least two symbols from package identity `scip-go gomod github.com/liza-mas/liza .`, at least one symbol from `scip-go gomod github.com/liza-mas/scip-search .`, and at least one symbol from `scip-go gomod github.com/sourcegraph/scip-bindings .`.
- Include a package fixture case where text such as `liza-mas` appears outside `packageName`, for example in descriptor text, and prove that `packages --prefix github.com/liza-mas/` does not include that package solely because of non-package-name text.
- Invoke commands through the normal runtime path with a fixture index path and capture successful stdout. Parse JSON before comparison; do not compare raw strings except where asserting exact scalar values inside the parsed payload.
- Assert `symbols --name Supervisor` returns all matching `Supervisor` display or descriptor cases and keeps them sorted by full symbol.
- Assert `symbols --name Run` returns the run symbol case and excludes unrelated symbols that merely share a package identity.
- Assert `symbols --name DoesNotExist` returns a successful payload with an explicit empty `symbols` array.
- Assert unfiltered `packages` returns each distinct package identity exactly once despite duplicate symbols.
- Assert package prefix filtering is literal, case-sensitive, and based on `packageName` only.
- Compare parsed golden JSON values rather than raw object field order.

**Validation command:** `go test ./internal/query/discovery/...`

## Dependency Plan

This task depends on `epic-planning-3-architecture-code-planning-1-replacement-coding-0` and `epic-planning-3-architecture-code-planning-2-replacement-coding-0` because the golden validation must exercise the implemented symbol and package discovery query surfaces rather than define those query behaviors itself.

The task also assumes the shared runtime loading boundary and traversal fixture mechanics from epic-planning-1 and epic-planning-2 are available in the coding checkout. If those foundations are absent, implementation should block rather than broadening into runtime, index loading, or traversal ownership.

## Shared-File Audit

This plan owns discovery golden validation files only. It must not modify symbol query implementation files, package query implementation files, `internal/scipmodel` helper files, shared CLI runtime files, shared index-loading files, traversal construction files, reference or implementation query files, README files, or shared runtime failure fixture files. Those are dependency or sibling scopes.

## Test Impact

Task 1 must add fixture-backed validation tests and golden JSON files under `internal/query/discovery` for the successful discovery command cases listed in `done_when`.

## Doc Impact

No user-facing documentation update is required in this plan. The task adds maintainer validation coverage for behavior already specified by the README, epic, and story documents; it does not change the public command contract text.

## Spec Compliance Matrix

| # | Requirement | Source | Task(s) | Status |
|---|-------------|--------|---------|--------|
| 1 | Symbol fixture data is deterministic, small enough for normal CLI validation, and uses the same SCIP loading and traversal path as `symbols`. | `CAP-003-01-symbol-query-fixtures.md` NFR-000-1 and NFR-000-3 | Task 1 | Covered |
| 2 | Symbol fixture data includes `Supervisor#`, `SupervisorConfig#`, `Run().`, and `SupervisorAgent#` full SCIP symbols for exact-looking, substring, and overlapping-name matches. | `CAP-003-01-symbol-query-fixtures.md` AC-001-1 | Task 1 | Covered |
| 3 | `symbols --name Supervisor` golden validation includes every display or descriptor match and treats overlapping names as a successful multi-result response. | `CAP-003-01-symbol-query-fixtures.md` AC-001-2; `CAP-003-03-golden-json-validation.md` AC-001-1 | Task 1 | Covered |
| 4 | `symbols --name Run` golden validation proves partial literal matching, includes the `supervisor/Run().` exact full symbol, and excludes non-matching `Supervisor` entries solely sharing the package. | `CAP-003-01-symbol-query-fixtures.md` AC-001-3; `CAP-003-03-golden-json-validation.md` AC-001-2 | Task 1 | Covered |
| 5 | Non-empty symbol golden entries assert exact full `symbol` values and CAP-001 match context fields. | `CAP-003-01-symbol-query-fixtures.md` AC-001-4; `CAP-003-03-golden-json-validation.md` AC-001-2 | Task 1 | Covered |
| 6 | Multi-result symbol golden entries assert stable ascending order by observable full `symbol`. | `CAP-003-01-symbol-query-fixtures.md` AC-001-5; `CAP-003-03-golden-json-validation.md` AC-001-4 | Task 1 | Covered |
| 7 | `symbols --name DoesNotExist` golden validation returns a successful payload with an explicit empty `symbols` collection. | `CAP-003-01-symbol-query-fixtures.md` AC-002-1 and AC-002-3; `CAP-003-03-golden-json-validation.md` AC-001-3 | Task 1 | Covered |
| 8 | Symbol no-match validation does not assert stderr diagnostics, nonzero status, shared runtime error fields, or suggestions. | `CAP-003-01-symbol-query-fixtures.md` AC-002-2 | Task 1 | Covered |
| 9 | Symbol golden validation asserts only successful query-specific `symbols` behavior and excludes shared runtime failures, package payloads, reference results, and implementation results. | `CAP-003-03-golden-json-validation.md` AC-001-5; assigned scope | Task 1 | Covered |
| 10 | Package fixture data is deterministic, small enough for normal CLI validation, and uses the same SCIP loading and traversal path as `packages`. | `CAP-003-02-package-query-fixtures.md` NFR-000-1 and NFR-000-3 | Task 1 | Covered |
| 11 | Package fixture data includes repeated `github.com/liza-mas/liza`, `github.com/liza-mas/scip-search`, and `github.com/sourcegraph/scip-bindings` package identities. | `CAP-003-02-package-query-fixtures.md` AC-001-1 | Task 1 | Covered |
| 12 | Unfiltered `packages` golden validation includes each distinct fixture package identity exactly once and proves duplicate symbols do not duplicate package entries. | `CAP-003-02-package-query-fixtures.md` AC-001-2; `CAP-003-03-golden-json-validation.md` AC-002-1 and AC-002-2 | Task 1 | Covered |
| 13 | Package identities that differ by package name, package manager, scheme, or package version remain separate entries. | `CAP-003-02-package-query-fixtures.md` AC-001-3 | Task 1 | Covered |
| 14 | Non-empty package golden entries assert `scheme`, `packageManager`, `packageName`, `packageVersion`, and `packageKey`. | `CAP-003-02-package-query-fixtures.md` AC-001-4; `CAP-003-03-golden-json-validation.md` AC-002-6 | Task 1 | Covered |
| 15 | Unfiltered package golden entries assert stable ascending order by observable `packageKey`. | `CAP-003-02-package-query-fixtures.md` AC-001-5; `CAP-003-03-golden-json-validation.md` AC-002-5 | Task 1 | Covered |
| 16 | `packages --prefix github.com/liza-mas/` golden validation includes only the two `github.com/liza-mas/` package identities. | `CAP-003-02-package-query-fixtures.md` AC-002-1; `CAP-003-03-golden-json-validation.md` AC-002-3 | Task 1 | Covered |
| 17 | `packages --prefix github.com/liza-mas/scip-search` golden validation includes only the exact package identity whose `packageName` starts with that literal prefix. | `CAP-003-02-package-query-fixtures.md` AC-002-2; assigned done_when | Task 1 | Covered |
| 18 | Prefix text outside `packageName` does not cause a package identity to match the `github.com/liza-mas/` prefix case. | `CAP-003-02-package-query-fixtures.md` AC-002-3 | Task 1 | Covered |
| 19 | `packages --prefix github.com/no-match/` golden validation returns a successful payload with an explicit empty `packages` collection. | `CAP-003-02-package-query-fixtures.md` AC-002-4; `CAP-003-03-golden-json-validation.md` AC-002-4 | Task 1 | Covered |
| 20 | Filtered package golden entries preserve the same stable package ordering as unfiltered package listing. | `CAP-003-02-package-query-fixtures.md` AC-002-5; `CAP-003-03-golden-json-validation.md` AC-002-5 | Task 1 | Covered |
| 21 | Golden JSON validation compares parsed JSON values and does not assert object field order. | `CAP-003-03-golden-json-validation.md` ASM-000-1; assigned done_when | Task 1 | Covered |
| 22 | Golden JSON validation expects no success diagnostics in stdout and asserts query-specific payload fields only. | `CAP-003-03-golden-json-validation.md` NFR-000-2 | Task 1 | Covered |
| 23 | Golden validation does not depend on external indexer installation or large real-world repositories. | `CAP-003-03-golden-json-validation.md` NFR-000-4; assigned scope | Task 1 | Covered |
| 24 | The task stays out of shared runtime failure fixtures, shared traversal fixture construction ownership, reference or implementation fixtures, alternate output formats, performance benchmarks, and ctags fallback data. | Assigned scope; CAP-003 story Out of Scope sections | Task 1 | Covered |
| E2E | e2e test coverage for new behavior | Cross-cutting | Task 1: fixture-backed tests invoke the normal SCIP loading and traversal path for both discovery commands | Covered |
| DOC | Documentation updates for changed behavior | Cross-cutting | N/A: maintainer validation coverage only; README and story specs already define the public behavior and no public contract changes | N/A |

## Pre-Submit Validation Checklist

- Re-read this plan and verify the output JSON fields are character-identical to Task 1 `desc`, `done_when`, `scope`, and `spec_ref`.
- Run `jq . specs/plans/readme/20260517-192848-epic-planning-3-architecture-code-planning-3-replacement-output.json`.
- Search this plan for `Task 1`, `epic-planning-3-architecture-code-planning-1-replacement-coding-0`, and `epic-planning-3-architecture-code-planning-2-replacement-coding-0` references and confirm the referenced responsibility or dependency is explicitly stated.
- Run pre-commit on the plan and output JSON files.
- Commit only the plan and output JSON artifacts.
