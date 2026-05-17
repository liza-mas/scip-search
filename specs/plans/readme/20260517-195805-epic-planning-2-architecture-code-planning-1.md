# Code Plan: Occurrence and Relationship Lookup Views

Status: draft

## Source Context

Based on:
- `README.md#language-support`, `README.md#what-is-scip-search`, and `README.md#out-of-scope`
- Blackboard task `epic-planning-2-architecture-code-planning-1`
- `specs/epics/readme/20260517-100328-epic-planning-2.md`
- `specs/arch-plan/readme/20260517-172511-epic-planning-2-architecture.md`
- Prior dependency plan `specs/plans/readme/20260517-194839-epic-planning-2-architecture-code-planning-0.md`
- `specs/stories/readme/20260517-100328-epic-planning-2/CAP-003-01-occurrence-lookup-view.md`
- `specs/stories/readme/20260517-100328-epic-planning-2/CAP-003-02-relationship-lookup-view.md`
- Supporting metadata story documents `CAP-002-01-source-range-normalization.md` and `CAP-002-02-hover-metadata-access.md`

## Planning Boundary

This plan covers Scope 2 from the traversal architecture: exact full SCIP symbol occurrence lookup and SCIP relationship lookup views over the traversal facts produced by Scope 1.

In scope: symbol-keyed occurrence lookup for exact full SCIP symbols, relationship fact extraction from document-level and external symbol information, relationship lookup by owner/source symbol, relationship lookup by target symbol, and preservation of source symbol, target symbol, original direction, and all SCIP edge-kind flags including reference, implementation, type-definition, and definition.

Out of scope: partial symbol resolution, package-prefix filtering, reference or implementation query algorithms, missing-symbol command behavior, synthesized relationships, cross-index lookup, duplicate elimination policy, result grouping, response ordering, final JSON schemas, and fixture authoring.

If a coder starts before `epic-planning-2-architecture-code-planning-0-coding-0` has provided the Scope 1 traversal view and metadata facts in the target worktree, Task 1 must block instead of recreating Scope 1, raw SCIP loading, runtime behavior, or fixture authoring here.

## Architectural Direction

The implementation should extend the existing command-neutral `internal/traversal` view without creating a query package:

```text
internal/traversal View from Scope 1
  -> occurrence facts indexed by exact Occurrence.symbol
  -> relationship facts extracted from local and external SymbolInformation.relationships
  -> relationship lookup by owner/source symbol
  -> relationship lookup by target symbol over the same facts
  -> query planners
```

Occurrence lookup should be an access pattern over the existing occurrence facts. Relationship target lookup should be an access pattern over the same relationship facts used by owner lookup, not synthesized inverse relationships. Query planners decide which occurrences or edges matter for references, definitions, implementations, related symbols, grouping, ordering, duplicate handling, and JSON schemas.

## Planned Coding Tasks

### Task 1 - Expose exact-symbol occurrence lookup

**desc:** `scip-search` traversal exposes exact-symbol occurrence lookup over existing occurrence facts while preserving every matched occurrence's document, location, role, and override-documentation metadata.

**done_when:** Unit tests in `internal/traversal` pass asserting a planner can request occurrences for an exact full SCIP symbol and receive every occurrence whose SCIP occurrence symbol equals that key, including matches across multiple documents; each returned occurrence preserves its containing document identity, relative path, position encoding, SCIP range, enclosing-range presence or absence, raw symbol-role bitset, and occurrence override documentation exactly as exposed by the existing traversal facts; occurrences for other symbols are not returned for the requested key; an absent symbol returns an empty occurrence result without runtime or command-level missing-symbol behavior; traversal does not apply partial-name matching, package-prefix filtering, relationship expansion, result grouping, duplicate elimination, response ordering, cross-index lookup, or final JSON shaping.

**scope:** In scope: occurrence lookup API on `internal/traversal` view, exact full SCIP symbol map construction from existing occurrence facts, preservation of returned occurrence fact metadata, empty-result behavior for absent symbols, deterministic lookup tests, and colocated unit tests. Out of scope: creating the base traversal view and metadata facts from Scope 1, raw SCIP loading, fixture authoring, relationship lookup, partial symbol resolution, package-prefix filtering, reference or implementation command algorithms, command-level missing-symbol behavior, source-file reads, duplicate elimination policy, grouping, ordering, cross-index lookup, and final command JSON schemas.

**spec_ref:** specs/stories/readme/20260517-100328-epic-planning-2/CAP-003-01-occurrence-lookup-view.md

**task_depends_on:** `epic-planning-2-architecture-code-planning-0-coding-0`

**Planned files:**
- `internal/traversal/view.go`
- `internal/traversal/facts.go` if lookup result aliases or helper types are needed
- `internal/traversal/view_test.go`

**Implementation notes:**
- Use exact string equality against the SCIP occurrence symbol supplied by Scope 1 facts.
- Preserve slice contents as traversal facts; do not clone into command-specific DTOs.
- Validation command for the task: `go test ./internal/traversal`.

### Task 2 - Expose SCIP relationship facts and owner lookup

**desc:** `scip-search` traversal exposes SCIP relationship facts owned by exact full SCIP symbols from both document-level and external symbol information while preserving owner, target, direction, and edge-kind flags.

**done_when:** Unit tests in `internal/traversal` pass asserting every relationship attached to a document-level or external SCIP `SymbolInformation` entry is represented as a relationship fact keyed by the owning full SCIP symbol; planners can request relationships owned by an exact full SCIP symbol and receive every relationship attached to that owner; each fact preserves the owner/source symbol, target symbol, original owner-to-target direction, and every SCIP edge-kind flag including reference, implementation, type-definition, and definition, with multiple flags preserved together; relationships owned by other symbols are excluded from that owner lookup; an owner with no relationships returns an empty relationship result without runtime or command-level missing-symbol behavior; traversal does not synthesize relationships, infer flags from occurrences, apply reference or implementation selection, collapse flags into command labels, group or de-duplicate results, define ordering policy, shape final JSON, or perform cross-index lookup.

**scope:** In scope: relationship fact type or fields in `internal/traversal`, extraction from existing document-level and external symbol facts, owner/source symbol lookup accessor or equivalent view method, preservation of owner/source symbol, target symbol, original direction, reference flag, implementation flag, type-definition flag, definition flag, multi-flag relationships, empty owner-result behavior, and colocated unit tests. Out of scope: target-symbol lookup, synthetic inverse relationships, occurrence role interpretation, command-specific reference, implementation, definition, type-definition, related-symbol, missing-symbol, grouping, duplicate, ordering, or JSON semantics, fixture authoring, source-file reads, raw path loading, custom persisted formats, daemon/watch/cache behavior, and ctags fallback behavior.

**spec_ref:** specs/stories/readme/20260517-100328-epic-planning-2/CAP-003-02-relationship-lookup-view.md

**Depends on:** Task 1.

**Planned files:**
- `internal/traversal/facts.go`
- `internal/traversal/view.go`
- `internal/traversal/view_test.go`

**Implementation notes:**
- Treat owner/source as the containing `SymbolInformation.symbol` and target as `Relationship.symbol`.
- Preserve boolean edge-kind flags as independent facts so multi-flag relationships remain representable.
- Validation command for the task: `go test ./internal/traversal`.

### Task 3 - Expose SCIP relationship lookup by target

**desc:** `scip-search` traversal exposes target-symbol relationship lookup over the same relationship facts used by owner lookup without synthesizing inverse edges or losing original direction.

**done_when:** Unit tests in `internal/traversal` pass asserting query planners can request relationships targeting an exact full SCIP symbol and receive every relationship fact whose SCIP relationship target equals that key; each returned fact remains the original source-to-target relationship with owner/source symbol, target symbol, original direction, and reference, implementation, type-definition, definition, and multi-flag edge-kind data preserved; relationships from multiple owners to the same target remain separate facts without merging owners; relationships targeting other symbols are excluded from that target lookup; an absent target symbol returns an empty relationship result without runtime or command-level missing-symbol behavior; traversal does not synthesize reverse relationships, decide whether incoming edges are implementation, reference, definition, or related-symbol command results, group or de-duplicate results, define ordering policy, shape final JSON, or perform cross-index lookup.

**scope:** In scope: target-symbol relationship lookup accessor or equivalent view method over Task 2 relationship facts, target lookup map construction, preservation of original owner/source symbol and target symbol, preservation of original direction and all edge-kind flags returned through target lookup, empty target-result behavior, multi-owner same-target behavior, and colocated unit tests. Out of scope: relationship fact extraction beyond what Task 2 owns, synthetic inverse edge creation, command-specific implementation/reference/definition/type-definition/related-symbol selection, missing-symbol command behavior, duplicate elimination policy, result grouping, response ordering, final CLI JSON schemas, fixture authoring, source-file reads, raw path loading, custom persisted formats, daemon/watch/cache behavior, and ctags fallback behavior.

**spec_ref:** specs/stories/readme/20260517-100328-epic-planning-2/CAP-003-02-relationship-lookup-view.md

**Depends on:** Task 2.

**Planned files:**
- `internal/traversal/view.go`
- `internal/traversal/view_test.go`
- `internal/traversal/facts.go` if target lookup helpers require shared relationship fact accessors

**Implementation notes:**
- Build target lookup from the Task 2 relationship fact collection; do not invert direction in the fact itself.
- Preserve every owner-to-target fact separately when multiple owners target the same symbol.
- Validation command for the task: `go test ./internal/traversal`.

## Dependency Plan

Task 1 depends on existing concrete task `epic-planning-2-architecture-code-planning-0-coding-0` because occurrence lookup must extend the Scope 1 traversal view and occurrence metadata facts instead of creating them here. Task 2 depends on Task 1 because both tasks modify the same `internal/traversal` files and should integrate serially to avoid parallel merge conflicts. Task 3 depends on Task 2 because target lookup must reuse the same relationship facts as owner lookup and because it modifies the same package files.

## Shared-File Audit

| File or package | Tasks | Dependency |
|-----------------|-------|------------|
| `internal/traversal/view.go` | Task 1, Task 2, Task 3 | Task 1 -> Task 2 -> Task 3 |
| `internal/traversal/facts.go` | Task 1 optional, Task 2, Task 3 optional | Task 1 -> Task 2 -> Task 3 |
| `internal/traversal/view_test.go` | Task 1, Task 2, Task 3 | Task 1 -> Task 2 -> Task 3 |

No two sibling tasks in this plan modify the same file without a dependency chain.

## Test Impact

Task 1 adds unit tests for exact full SCIP symbol occurrence lookup, cross-document matches, non-matching symbols, preserved document/source-location/role/override-documentation metadata, absent-symbol empty results, and excluded query semantics.

Task 2 adds unit tests for relationship extraction from document-level and external symbol information, owner/source lookup, absent owner empty results, preserved original direction, every required SCIP edge-kind flag, multi-flag edges, and excluded query semantics.

Task 3 adds unit tests for target lookup over the same relationship facts, absent target empty results, preserved original direction and edge-kind flags through target lookup, multiple owner facts targeting the same symbol, excluded non-target relationships, no synthesized reverse edges, and excluded query semantics.

No e2e command tests are planned in this scope because the behavior is an internal traversal lookup API and must not define command stdout, stderr, status, result grouping, ordering, duplicate handling, or final JSON behavior. Query command e2e coverage belongs to sibling query plans after traversal facts exist.

## Doc Impact

No README or user-facing documentation task is planned. This scope adds internal traversal lookup APIs and the source requirements are already captured in the epic, architecture plan, story documents, and this code plan. User-visible command behavior and response schemas are explicitly out of scope.

## Spec Compliance Matrix

| # | Requirement | Source | Task(s) | Status |
|---|-------------|--------|---------|--------|
| 1 | Occurrence lookup keys are exact full SCIP symbol strings from `Occurrence.symbol`. | CAP-003-01 ASM-000-1; ST-001 AC-001-1; assigned scope | Task 1 | Covered |
| 2 | Occurrence lookup returns every occurrence whose symbol exactly equals the requested key, including matches across one or more documents. | CAP-003-01 ST-001 AC-001-1; assigned done_when | Task 1 | Covered |
| 3 | Occurrence lookup keeps each returned occurrence associated with containing document identity and relative path. | CAP-003-01 ST-001 AC-001-2; CAP-002-01 ST-001 AC-001-2 | Task 1 | Covered |
| 4 | Occurrence lookup results preserve source-location metadata including document position encoding, SCIP range, and enclosing-range presence or absence. | CAP-002-01 ST-001 AC-001-3; CAP-002-01 ST-002 AC-002-1 through AC-002-3b; assigned done_when | Task 1 | Covered |
| 5 | Occurrence lookup results preserve raw SCIP symbol-role bitsets without command-specific interpretation. | CAP-003-01 ST-002 AC-002-1 through AC-002-5; assigned done_when | Task 1 | Covered |
| 6 | Occurrence lookup results preserve occurrence override documentation and keep hover metadata available from the existing occurrence fact. | CAP-002-02 ST-002 AC-002-1 through AC-002-4; assigned done_when | Task 1 | Covered |
| 7 | Occurrences for other symbols remain distinguishable and are not included for the requested exact symbol key. | CAP-003-01 ST-001 AC-001-3 | Task 1 | Covered |
| 8 | Absent occurrence symbols return an empty traversal result without deciding command-level missing-symbol behavior or failure semantics. | CAP-003-01 ST-001 AC-001-4; ASM-000-2; assigned done_when | Task 1 | Covered |
| 9 | Occurrence lookup does not apply partial-name matching, package-prefix filtering, relationship expansion, final JSON shaping, result grouping, duplicate elimination, or cross-index lookup. | CAP-003-01 ST-001 AC-001-5; NFR-000-3; assigned scope | Task 1 | Covered |
| 10 | Relationship facts are extracted from document-level and external SCIP symbol information. | CAP-003-02 ST-001 AC-001-1; ASM-001-1; assigned scope | Task 2 | Covered |
| 11 | Relationship facts preserve owner/source full SCIP symbol and target full SCIP symbol as distinguishable fields. | CAP-003-02 ST-001 AC-001-2; ASM-000-1; assigned done_when | Task 2 | Covered |
| 12 | Relationship facts preserve original owner-to-target direction. | CAP-003-02 ST-002 AC-002-2; architecture Relationship Lookup invariant; assigned done_when | Task 2 | Covered |
| 13 | Owner/source lookup by exact full SCIP symbol returns every relationship attached to that owner and excludes relationships owned by other symbols. | CAP-003-02 ST-001 AC-001-1, AC-001-3 | Task 2 | Covered |
| 14 | Absent owner symbols return empty relationship results without deciding command-level missing-symbol behavior or failure semantics. | CAP-003-02 ST-001 AC-001-4 | Task 2 | Covered |
| 15 | Target lookup by exact full SCIP symbol returns every relationship whose target equals that symbol. | CAP-003-02 ST-002 AC-002-1; assigned done_when | Task 3 | Covered |
| 16 | Target lookup preserves each original owner/source symbol and exposes multiple owners targeting the same symbol as separate relationship facts. | CAP-003-02 ST-002 AC-002-2, AC-002-3 | Task 3 | Covered |
| 17 | Absent target symbols return empty relationship results without deciding command-level missing-symbol behavior or failure semantics. | CAP-003-02 ST-002 AC-002-4 | Task 3 | Covered |
| 18 | Target lookup is an access pattern over the same relationship facts and does not synthesize inverse relationships. | CAP-003-02 ASM-002-1; ST-002 Out of Scope; assigned scope | Task 3 | Covered |
| 19 | Relationship lookup preserves `is_reference`, `is_implementation`, `is_type_definition`, and `is_definition` edge-kind flags. | CAP-003-02 ST-003 AC-003-1 through AC-003-4; assigned scope | Task 2, Task 3 | Covered |
| 20 | Relationship lookup preserves multiple edge-kind flags on one relationship instead of collapsing to a preferred kind. | CAP-003-02 ST-003 AC-003-4b | Task 2, Task 3 | Covered |
| 21 | Relationship lookup does not decide final reference expansion, implementation expansion, go-to-definition behavior, go-to-type-definition behavior, related-symbol expansion, command result grouping, duplicate elimination, ordering, cross-index lookup, or JSON fields. | CAP-003-02 ST-001 AC-001-5; ST-002 AC-002-5; ST-003 AC-003-5; assigned scope | Task 2, Task 3 | Covered |
| 22 | Lookup views remain deterministic and traceable to official SCIP fields, with no custom persisted index data, source-file reads, daemon behavior, or cross-process persistence. | CAP-003-01 NFR-000-1, NFR-000-2; CAP-003-02 NFR-000-1, NFR-000-2; epic NFR-000-1 and NFR-000-2 | Task 1, Task 2, Task 3 | Covered |
| 23 | Fixture authoring remains out of this plan. | Architecture Scope 2 Boundary; assigned scope | Task 1, Task 2, Task 3 | Covered |
| E2E | e2e test coverage for new behavior | Cross-cutting | N/A: internal traversal lookup API only; command e2e behavior is owned by sibling query/runtime plans after traversal exists. | N/A |
| DOC | Documentation updates for changed behavior | Cross-cutting | N/A: internal package contract is documented by this plan and existing epic/story specs; no user-visible command behavior or README schema changes are introduced. | N/A |

## Pre-Submit Validation Checklist

- Re-read this plan and verify the output JSON fields are character-identical to each task's `desc`, `done_when`, `scope`, and `spec_ref`.
- Run `jq . specs/plans/readme/20260517-195805-epic-planning-2-architecture-code-planning-1-output.json`.
- Search this plan for every `Task N` cross-reference and confirm the referenced task states the corresponding responsibility, dependency, or exclusion.
- Confirm every shared file listed in more than one task has a dependency chain.
- Run pre-commit on the plan and output JSON files.
- Commit only the plan and output JSON artifacts.
