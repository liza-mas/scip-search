# Code Plan: Traversal Fixtures and Coverage Validation

Status: draft

## Source Context

Based on:
- `README.md#language-support`, `README.md#scip-symbol-format`, and `README.md#out-of-scope`
- Blackboard task `epic-planning-2-architecture-code-planning-2`
- `specs/epics/readme/20260517-100328-epic-planning-2.md`
- `specs/arch-plan/readme/20260517-172511-epic-planning-2-architecture.md`
- Prior dependency plan `specs/plans/readme/20260517-194839-epic-planning-2-architecture-code-planning-0.md`
- Prior dependency plan `specs/plans/readme/20260517-195805-epic-planning-2-architecture-code-planning-1.md`
- `specs/stories/readme/20260517-100328-epic-planning-2/CAP-004-01-schema-valid-traversal-fixtures.md`
- `specs/stories/readme/20260517-100328-epic-planning-2/CAP-004-02-fixture-coverage-validation.md`
- Supporting lookup story documents `CAP-003-01-occurrence-lookup-view.md` and `CAP-003-02-relationship-lookup-view.md`

## Planning Boundary

This plan covers Scope 3 from the traversal architecture: deterministic schema-valid SCIP fixture payloads and traversal-owned coverage validation over the shared traversal view and lookup APIs produced by prior scopes.

In scope: compact deterministic schema-valid SCIP fixture payloads loaded through the official binding path; fixture helper accessors under traversal-owned test boundaries; coverage validation for documents, languages, paths, position encodings, local symbols, external symbols, occurrences, exact-symbol lookup, relationship owner and target lookup, same-line and multi-position ranges, enclosing range presence/absence, definition and non-definition role cases, multi-bit roles, relationship edge-kind flags, symbol hover metadata, signature documentation, and occurrence override documentation.

Out of scope: installing or invoking external indexers, large real-world fixtures, performance benchmarks, ctags fallback fixtures, command-specific fixtures, final command JSON golden files, stdout/stderr/status validation, and query-specific selection or grouping behavior.

If a coder starts before the traversal metadata facts and lookup views from `epic-planning-2-architecture-code-planning-0` and `epic-planning-2-architecture-code-planning-1` exist in the target worktree, Task 1 must block instead of recreating traversal facts, occurrence lookup, relationship lookup, raw SCIP loading, command runtime behavior, or query-specific command fixtures.

## Architectural Direction

The implementation should keep fixtures inside traversal-owned test boundaries and validate them through the same loaded-index and traversal view path that query planners use:

```text
schema-valid SCIP fixture payload
  -> official SCIP binding load path used by traversal tests
  -> internal/traversal View from prior scopes
  -> traversal fixture helpers
  -> category-level coverage assertions
```

Fixture helper APIs should return traversal facts or stable fixture handles for tests, not command response DTOs. Coverage tests should fail with missing traversal categories, such as missing document coverage, missing range coverage, or missing relationship edge-kind coverage. They must not assert command JSON, stdout, stderr, process status, symbol name matching, package prefix filtering, reference selection, implementation selection, grouping, duplicate policy, or ordering semantics.

## Planned Coding Tasks

### Task 1 - Create schema-valid traversal fixture set and helpers

**desc:** `scip-search` has a compact traversal-owned SCIP fixture set and test helper boundary that load schema-valid fixture payloads through the same official binding path used by traversal.

**done_when:** Unit tests in `internal/traversal` or `internal/traversal/traversaltest` pass asserting the shared fixture set loads through the official SCIP binding path used by traversal, builds the traversal view without command runtime behavior, and exposes helper accessors for at least two documents with distinct relative paths, document language and position encoding, at least one document-level symbol, at least one external symbol, occurrences for more than one full SCIP symbol, same-line and multi-position range forms, present and absent enclosing ranges, definition and non-definition occurrences, a multi-bit role occurrence, symbol kind, display name, documentation, signature documentation, present and absent occurrence override documentation, and relationship facts with owner, target, original direction, and reference, implementation, definition, and type-definition edge-kind flags; the fixture setup does not invoke external indexers, read source files, define command JSON goldens, assert stdout/stderr/status, or encode query-specific selection or grouping behavior.

**scope:** In scope: traversal-owned fixture payloads under `internal/traversal/testdata` or equivalent, traversal test helper accessors under `internal/traversal/traversaltest` or equivalent test-only boundaries, schema-valid official SCIP binding data, compact deterministic documents, local and external symbols, occurrences, ranges, role bitsets, hover metadata, override documentation, relationship owner and target facts, relationship edge-kind flags, and colocated tests proving the helpers load and build the traversal view. Out of scope: production fixture APIs, external indexer installation or invocation, large real-world repositories, performance benchmarks, ctags fallback data, command-specific fixtures, final command JSON golden files, stdout/stderr/status assertions, raw path loading owned by traversal, source-file reads, query matching/filtering, result grouping, duplicate policy, response ordering, and final CLI JSON schemas.

**spec_ref:** specs/stories/readme/20260517-100328-epic-planning-2/CAP-004-01-schema-valid-traversal-fixtures.md

**task_depends_on:** `epic-planning-2-architecture-code-planning-0-coding-3`, `epic-planning-2-architecture-code-planning-1-coding-2`

**Planned files:**
- `internal/traversal/testdata/*` or equivalent traversal-owned fixture payload files
- `internal/traversal/traversaltest/*` or equivalent test-only fixture helper package
- `internal/traversal/*_test.go` only if the helper smoke test belongs beside traversal package tests

**Implementation notes:**
- Prefer one compact fixture set that covers the required categories instead of many category-specific payloads.
- Keep helper names category-oriented and internal to tests so downstream query packages can reuse fixture handles without treating them as command goldens.
- Validation command for the task: `go test ./internal/traversal/...`.

### Task 2 - Validate traversal fixture coverage categories

**desc:** `scip-search` maintainer tests validate the shared traversal fixture set through traversal view APIs and fail with the missing SCIP traversal category when required coverage is absent.

**done_when:** Unit tests in `internal/traversal` pass asserting fixture coverage validation loads the Task 1 shared fixture through the official binding path, builds the traversal view, and proves at least two distinct document paths with language and position encoding, at least one document-level symbol, at least one external symbol, occurrences for more than one full SCIP symbol, empty lookup behavior for an absent symbol, exact-symbol occurrence lookup preserving containing document and symbol, relationship owner and target lookup preserving source symbol, target symbol, original direction, reference flag, implementation flag, definition flag, type-definition flag, same-line and multi-position range forms, present and absent enclosing ranges, definition and non-definition occurrences, a multi-bit role occurrence, symbol kind, display name, documentation, signature documentation, present and absent occurrence override documentation, and category-named failure messages for missing coverage; the validation does not assert command JSON, stdout, stderr, exit status, command-level missing-symbol behavior, query-specific selection, grouping, duplicate elimination, ordering, or package or partial-name matching semantics.

**scope:** In scope: traversal fixture coverage tests, category-level assertion helpers, validation through the shared traversal view and exact-symbol occurrence plus owner and target relationship lookup APIs, absent lookup checks for traversal-empty results, coverage failure messages that identify missing traversal categories, and dependency on the Task 1 fixture helper boundary. Out of scope: creating additional fixture payload categories beyond Task 1, command-specific golden JSON files, end-to-end CLI execution, stdout/stderr/status assertions, malformed-index diagnostics, command-level missing-symbol outcomes, reference or implementation result selection, symbol partial-name matching, package prefix filtering, result grouping, duplicate elimination, response ordering, performance benchmarking, external indexer setup, ctags fallback coverage, and final CLI JSON schemas.

**spec_ref:** specs/stories/readme/20260517-100328-epic-planning-2/CAP-004-02-fixture-coverage-validation.md

**Depends on:** Task 1.

**Planned files:**
- `internal/traversal/fixture_coverage_test.go` or equivalent traversal package coverage test
- `internal/traversal/traversaltest/*` only if category assertion helpers belong beside the fixture helper package

**Implementation notes:**
- Assert category coverage through traversal facts and lookup APIs, not by inspecting a custom manifest as source truth.
- Use explicit category names in failure messages so missing coverage points maintainers to the fixture gap.
- Validation command for the task: `go test ./internal/traversal/...`.

## Dependency Plan

Task 1 depends on existing concrete tasks `epic-planning-2-architecture-code-planning-0-coding-3` and `epic-planning-2-architecture-code-planning-1-coding-2` because the fixture helper must build the prior traversal view with metadata facts and exercise the prior occurrence and relationship lookup APIs instead of creating those APIs here.

Task 2 depends on Task 1 because coverage validation needs the shared fixture payloads and helper boundary before it can prove category coverage. Task 2 may share `internal/traversal/traversaltest/*` with Task 1 if category assertion helpers are colocated there, so the dependency also prevents parallel edits to the helper package.

## Shared-File Audit

| File or package | Tasks | Dependency |
|-----------------|-------|------------|
| `internal/traversal/traversaltest/*` | Task 1, Task 2 optional | Task 1 -> Task 2 |
| `internal/traversal/testdata/*` | Task 1 | None needed |
| `internal/traversal/fixture_coverage_test.go` | Task 2 | None needed |

No two sibling tasks in this plan modify the same file or package without a dependency chain.

## Test Impact

Task 1 adds traversal-owned unit tests or helper smoke tests proving schema-valid fixture payloads load through the official binding path, build the traversal view, and expose helper accessors for documents, symbols, occurrences, ranges, roles, hover metadata, override documentation, relationships, and edge-kind flags.

Task 2 adds traversal fixture coverage tests proving the shared fixture covers every required data category and reports missing categories clearly, including absent exact-symbol lookup behavior, range variants, optional metadata presence and absence, role cases, relationship owner and target lookup, original direction, and edge-kind flags.

No e2e command tests are planned in this scope because traversal fixtures are internal shared test assets and must not define command stdout, stderr, process status, final JSON schemas, or query-specific selection semantics. Query command e2e coverage belongs to sibling query plans.

## Doc Impact

No README or user-facing documentation task is planned. This scope adds internal traversal fixtures and maintainer validation tests, and the source requirements are already captured in the epic, architecture plan, story documents, and this code plan. User-visible command behavior and response schemas are explicitly out of scope.

## Spec Compliance Matrix

| # | Requirement | Source | Task(s) | Status |
|---|-------------|--------|---------|--------|
| 1 | Fixture payloads are deterministic, compact, schema-valid SCIP data read through the same official binding path used by traversal. | CAP-004-01 NFR-000-1, NFR-000-2; README.md#language-support; assigned scope | Task 1, Task 2 | Covered |
| 2 | Fixture setup must not install or invoke external language indexers, use large real-world repositories, add performance benchmarks, use ctags fallback data, or create custom persisted index formats. | CAP-004-01 NFR-000-4; CAP-004-01 Out of Scope; assigned scope | Task 1, Task 2 | Covered |
| 3 | Fixture data preserves exact SCIP symbol strings, document paths, ranges, role bitsets, relationships, and hover metadata without command-specific normalization. | CAP-004-01 NFR-000-3 | Task 1, Task 2 | Covered |
| 4 | Fixture validation loads at least two SCIP documents with distinct relative paths and document identities. | CAP-004-01 ST-001 AC-001-1; CAP-004-02 ST-001 AC-001-1; assigned done_when | Task 1, Task 2 | Covered |
| 5 | Each fixture document preserves language and position encoding for planner use. | CAP-004-01 ST-001 AC-001-2; CAP-004-02 ST-002 AC-002-1 | Task 1, Task 2 | Covered |
| 6 | Fixture symbol inventory includes at least one document-level full SCIP symbol and at least one external full SCIP symbol as distinct traversal facts. | CAP-004-01 ST-001 AC-001-3, AC-001-4; CAP-004-02 ST-001 AC-001-2 | Task 1, Task 2 | Covered |
| 7 | Fixture occurrences cover more than one full SCIP symbol and remain associated with exact symbols and containing documents. | CAP-004-01 ST-001 AC-001-5; CAP-004-02 ST-001 AC-001-3; assigned done_when | Task 1, Task 2 | Covered |
| 8 | Fixture and validation scope excludes command routing, package-prefix filtering, partial symbol matching, reference or implementation result selection, grouping, ordering, duplicate policy, and final command JSON schemas. | CAP-004-01 ST-001 AC-001-6; CAP-004-02 ST-001 AC-001-7; assigned scope | Task 1, Task 2 | Covered |
| 9 | Fixture occurrences include at least one same-line SCIP range form and at least one multi-position SCIP range form, and validation proves both forms are preserved as SCIP range values. | CAP-004-01 ST-002 AC-002-1; CAP-004-02 ST-002 AC-002-2; assigned done_when | Task 1, Task 2 | Covered |
| 10 | Fixture occurrences include both present and absent enclosing range metadata, and validation proves the two cases remain distinguishable. | CAP-004-01 ST-002 AC-002-2; CAP-004-02 ST-002 AC-002-3 | Task 1, Task 2 | Covered |
| 11 | Fixture role coverage includes at least one definition occurrence, at least one non-definition occurrence, and at least one occurrence with multiple SCIP role bits available together. | CAP-004-01 ST-002 AC-002-3; CAP-004-02 ST-002 AC-002-4; assigned done_when | Task 1, Task 2 | Covered |
| 12 | Fixture symbol metadata includes symbol kind, display name, documentation, and signature documentation for at least one symbol. | CAP-004-01 ST-002 AC-002-4; CAP-004-02 ST-002 AC-002-5; assigned done_when | Task 1, Task 2 | Covered |
| 13 | Fixture occurrence hover metadata includes present override documentation for at least one occurrence and absent override documentation for at least one other occurrence. | CAP-004-01 ST-002 AC-002-5; CAP-004-02 ST-002 AC-002-6; assigned done_when | Task 1, Task 2 | Covered |
| 14 | Range, role, and hover fixture validation must not render Markdown, read source files, convert coordinates for editors, choose final command response fields, or decide command metadata inclusion. | CAP-004-01 ST-002 AC-002-6; CAP-004-02 ST-002 AC-002-7 | Task 1, Task 2 | Covered |
| 15 | Fixture relationships include at least one document-level or external symbol information entry that owns a SCIP relationship. | CAP-004-01 ST-003 AC-003-1 | Task 1, Task 2 | Covered |
| 16 | Fixture relationship facts preserve owning source symbol, target symbol, and original source-to-target direction. | CAP-004-01 ST-003 AC-003-2; CAP-004-02 ST-001 AC-001-5; assigned done_when | Task 1, Task 2 | Covered |
| 17 | Fixture relationships cover implementation, reference, definition, and type-definition edge-kind flags. | CAP-004-01 ST-003 AC-003-3, AC-003-4; CAP-004-02 ST-001 AC-001-6; assigned done_when | Task 1, Task 2 | Covered |
| 18 | Fixture relationships preserve multiple edge-kind flags together when present instead of requiring collapsed single-kind relationships. | CAP-004-01 ST-003 AC-003-4b; CAP-003-02 ST-003 AC-003-4b | Task 1, Task 2 | Covered |
| 19 | Relationship fixture data and validation do not define final reference expansion, implementation selection, related-symbol behavior, duplicate elimination, response ordering, or command JSON fields. | CAP-004-01 ST-003 AC-003-5 | Task 1, Task 2 | Covered |
| 20 | Coverage validation proves occurrence lookup for a known full SCIP symbol preserves every matching occurrence's containing document and exact full SCIP symbol. | CAP-004-02 ST-001 AC-001-3; CAP-003-01 ST-001 AC-001-1, AC-001-2 | Task 2 | Covered |
| 21 | Coverage validation proves lookup for an absent full SCIP symbol returns an empty traversal result without treating it as a command-level missing-symbol failure. | CAP-004-02 ST-001 AC-001-4; CAP-003-01 ST-001 AC-001-4; assigned done_when | Task 2 | Covered |
| 22 | Coverage validation proves relationship owner and target lookup preserve source symbols, target symbols, and original direction. | CAP-004-02 ST-001 AC-001-5; CAP-003-02 ST-001 AC-001-1 through AC-001-3; CAP-003-02 ST-002 AC-002-1 through AC-002-3 | Task 2 | Covered |
| 23 | Coverage validation failures identify the missing traversal data category well enough for maintainers to repair the shared fixture. | CAP-004-02 NFR-000-3; assigned done_when | Task 2 | Covered |
| 24 | Coverage validation must exercise traversal behavior through official SCIP binding data rather than inspecting a custom fixture manifest as the source of truth. | CAP-004-02 NFR-000-2 | Task 1, Task 2 | Covered |
| 25 | Coverage validation must not assert final command JSON schemas, stdout/stderr behavior, process statuses, performance budgets, or query-specific selection semantics. | CAP-004-02 NFR-000-4; assigned done_when | Task 2 | Covered |
| E2E | e2e test coverage for new behavior | Cross-cutting | N/A: internal traversal fixture assets and category-level maintainer tests only; command e2e behavior is owned by sibling query/runtime plans. | N/A |
| DOC | Documentation updates for changed behavior | Cross-cutting | N/A: internal fixture and validation contract is documented by this plan and existing epic/story specs; no user-visible command behavior, install flow, or README response schema changes are introduced. | N/A |

## Pre-Submit Validation Checklist

- Re-read this plan and verify the output JSON fields are character-identical to each task's `desc`, `done_when`, `scope`, and `spec_ref`.
- Run `jq . specs/plans/readme/20260517-201903-epic-planning-2-architecture-code-planning-2-output.json`.
- Search this plan for every `Task N` cross-reference and confirm the referenced task states the corresponding responsibility, dependency, or exclusion.
- Confirm every shared file or package listed in more than one task has a dependency chain.
- Run pre-commit on the plan and output JSON files.
- Commit only the plan and output JSON artifacts.
