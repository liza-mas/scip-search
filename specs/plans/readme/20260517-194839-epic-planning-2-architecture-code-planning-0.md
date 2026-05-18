# Code Plan: Traversal View and Metadata Facts

Status: draft

## Source Context

Based on:
- `README.md#language-support`, `README.md#what-is-scip-search`, and `README.md#out-of-scope`
- `specs/epics/readme/20260517-100328-epic-planning-2.md`
- `specs/arch-plan/readme/20260517-172511-epic-planning-2-architecture.md`
- `specs/stories/readme/20260517-100328-epic-planning-2/CAP-001-01-scip-binding-input.md`
- `specs/stories/readme/20260517-100328-epic-planning-2/CAP-001-02-document-and-symbol-inventory.md`
- `specs/stories/readme/20260517-100328-epic-planning-2/CAP-002-01-source-range-normalization.md`
- `specs/stories/readme/20260517-100328-epic-planning-2/CAP-002-02-hover-metadata-access.md`
- Prior dependency plan `specs/plans/readme/20260517-192724-epic-planning-1-architecture-code-planning-1.md`
- Blackboard task `epic-planning-1-architecture-code-planning-1`, which supplies the runtime loaded SCIP index boundary this plan consumes.
- External binding check: pkg.go.dev for `github.com/scip-code/scip/bindings/go/scip` showed the package at module version `v0.7.1` on 2026-05-17, with `Index.GetMetadata`, `Index.GetDocuments`, `Index.GetExternalSymbols`, `Document.GetRelativePath`, `Document.GetLanguage`, `Document.GetPositionEncoding`, `Document.GetOccurrences`, `Document.GetSymbols`, `Occurrence.GetRange`, `Occurrence.GetEnclosingRange`, `Occurrence.GetSymbol`, `Occurrence.GetSymbolRoles`, `Occurrence.GetOverrideDocumentation`, and `SymbolInformation` getters for symbol, kind, display name, documentation, signature documentation, enclosing symbol, and relationships. The SCIP schema source documents range arrays, document position encoding, role bitsets, enclosing ranges, and occurrence override documentation.

## Planning Boundary

This plan covers Scope 1 from the traversal architecture: the reusable `internal/traversal` view and metadata facts. Traversal starts after the runtime has successfully loaded the caller-selected SCIP index through `internal/scipindex`.

In scope: `internal/traversal` accepts the shared runtime loaded SCIP boundary, uses official SCIP Go binding data as source truth, builds process-local inventories for index metadata, documents, document symbols, external symbols, and occurrences, and exposes document paths, languages, position encodings, occurrence ranges, enclosing-range absence/presence, symbol-role bitsets, symbol kind, display name, documentation, signature documentation, enclosing symbol metadata, and occurrence override documentation.

Out of scope: raw path loading, command routing, stdout/stderr/status behavior, query-specific matching or filtering, final CLI JSON schemas, source-file reads, custom persisted formats, daemon/watch/cache behavior, relationship lookup indexes, fixture authoring, and ctags fallback behavior.

## Architectural Direction

The implementation should add a small command-neutral package boundary:

```text
internal/scipindex loaded context
  -> internal/traversal.NewView(...)
       -> index metadata fact
       -> document facts
       -> local document symbol facts
       -> external symbol facts
       -> occurrence facts
  -> query planners
```

`internal/traversal` should expose immutable or read-only facts derived from official SCIP binding values. The package can use wrapper structs for planner ergonomics, but every field must remain traceable to a SCIP binding getter or schema field. It must not re-open the selected index path, read source files, write process streams, cache across invocations, or decide command-level meanings for roles, partial names, packages, references, implementations, or JSON fields.

If a coder starts before `epic-planning-1-architecture-code-planning-1` has provided a usable loaded-index context in the target worktree, Task 1 must block instead of recreating path loading, protobuf parsing, CLI routing, status handling, or the loader boundary in `internal/traversal`.

## Planned Coding Tasks

### Task 1 - Build traversal view from loaded index and expose document inventory

**desc:** `scip-search` has an `internal/traversal` view builder that accepts the runtime loaded SCIP index boundary and exposes invocation-local index metadata and document inventories without owning loading or command behavior.

**done_when:** Unit tests in `internal/traversal` pass asserting a loaded official SCIP index from `internal/scipindex` can build a traversal view; planners can enumerate index metadata and every document exactly once with relative path, language, position encoding, and empty or populated occurrence collections preserved; empty document inventories remain data rather than runtime errors; the builder does not accept raw paths, read index files, write stdout/stderr, persist derived data, or apply command-specific filtering or JSON shaping.

**scope:** In scope: `internal/traversal` package entrypoint, `View` or equivalent builder API, index metadata fact type, document fact type, integration with the `internal/scipindex` loaded context, and colocated unit tests using in-memory official SCIP binding data. Out of scope: raw SCIP path loading or protobuf byte parsing, CLI routing, stream/status behavior, symbol metadata beyond document inventory references, occurrence lookup maps, relationship lookup indexes, fixture authoring, source-file reads, custom persisted formats, daemon/watch/cache behavior, query matching/filtering, and final command JSON schemas.

**spec_ref:** specs/stories/readme/20260517-100328-epic-planning-2/CAP-001-01-scip-binding-input.md

**task_depends_on:** `epic-planning-1-architecture-code-planning-1`

**Planned files:**
- `internal/traversal/view.go`
- `internal/traversal/view_test.go`
- `internal/traversal/facts.go` if fact type extraction keeps the view readable

**Implementation notes:**
- Accept the loaded context from `internal/scipindex`; do not add a path, byte slice, or `io.Reader` loader API to traversal.
- Preserve document position encoding alongside relative path and language so later occurrence facts can refer to the containing document context.
- Validation command for the task: `go test ./internal/traversal`.

### Task 2 - Expose symbol and occurrence inventories

**desc:** `scip-search` traversal exposes document-level symbol, external symbol, and occurrence inventories from official SCIP data while preserving full symbol strings, local or external source, and containing document context.

**done_when:** Unit tests in `internal/traversal` pass asserting every document-level `SymbolInformation` entry and every index-level external symbol is represented as a symbol fact with the full SCIP symbol string and local-or-external source preserved; every document occurrence is represented as an occurrence fact associated with its containing document and referenced SCIP symbol string; empty document-symbol, external-symbol, and occurrence inventories are exposed as empty collections without placeholder facts; repeated enumeration over the same loaded index is deterministic; traversal does not resolve partial names, package prefixes, reference semantics, implementation semantics, or final JSON fields.

**scope:** In scope: document-level symbol facts, external symbol facts, occurrence facts, local/external source tagging, symbol string preservation, occurrence-to-document association, deterministic process-local enumeration, and colocated unit tests. Out of scope: source-location range enrichment beyond fields already exposed by Task 1, hover metadata enrichment, relationship lookup indexes, exact-symbol occurrence lookup maps, query-specific matching/filtering, final command JSON schemas, source-file reads, raw path loading, custom persisted formats, daemon/watch/cache behavior, and fixture authoring.

**spec_ref:** specs/stories/readme/20260517-100328-epic-planning-2/CAP-001-02-document-and-symbol-inventory.md

**Depends on:** Task 1.

**Planned files:**
- `internal/traversal/facts.go`
- `internal/traversal/view.go`
- `internal/traversal/view_test.go`

**Implementation notes:**
- Keep document-level symbols distinguishable from external symbols because downstream discovery and reference planners need to know where SCIP supplied the metadata.
- Do not add symbol-keyed lookup maps here; exact-symbol lookup is owned by the sibling Scope 2 code plan.
- Validation command for the task: `go test ./internal/traversal`.

### Task 3 - Preserve occurrence location and role metadata

**desc:** `scip-search` traversal occurrence facts preserve SCIP source-location metadata, enclosing-range presence or absence, and symbol-role bitsets together with the containing document path and position encoding.

**done_when:** Unit tests in `internal/traversal` pass asserting each occurrence fact retains the containing document relative path and position encoding, the exact SCIP range value, the ability to distinguish three-integer same-line ranges from four-integer multi-position ranges, optional enclosing range presence or absence without inventing defaults, and the raw SCIP symbol-role bitset; missing or empty SCIP document text does not cause traversal to read source files; traversal does not convert coordinates, decide definition/reference/implementation meanings, format paths or positions for JSON, or apply command-specific filtering.

**scope:** In scope: occurrence range preservation, enclosing range optionality, role bitset preservation, containing document path and position encoding access from occurrence facts, source-file-read avoidance tests, and colocated unit tests. Out of scope: validating or repairing malformed SCIP range arrays beyond preserving loaded official data, relationship lookup, implementation or reference selection, syntax highlighting or diagnostics display, final command JSON fields, source-file reads, raw path loading, custom persisted formats, daemon/watch/cache behavior, and fixture authoring.

**spec_ref:** specs/stories/readme/20260517-100328-epic-planning-2/CAP-002-01-source-range-normalization.md

**Depends on:** Task 2.

**Planned files:**
- `internal/traversal/facts.go`
- `internal/traversal/view.go`
- `internal/traversal/view_test.go`

**Implementation notes:**
- Preserve the SCIP range array form as data for query planners rather than normalizing it into an editor coordinate model.
- Represent absent enclosing ranges distinctly from present empty or present non-empty values when the official binding exposes that distinction.
- Validation command for the task: `go test ./internal/traversal`.

### Task 4 - Preserve symbol and occurrence hover metadata

**desc:** `scip-search` traversal symbol and occurrence facts preserve SCIP hover metadata, including symbol kind, display name, documentation, signature documentation, enclosing symbol metadata, and occurrence override documentation, without rendering or merging it.

**done_when:** Unit tests in `internal/traversal` pass asserting document-level and external symbol facts retain full SCIP symbol string, local-or-external source, kind, display name, documentation entries, signature documentation, and enclosing symbol metadata; occurrence facts retain present override documentation entries and preserve absent override documentation as absence or an empty collection without copying symbol documentation; symbol-level documentation and occurrence override documentation remain distinguishable; traversal does not render Markdown, flatten signature documentation into display text, read source files for missing docs, choose command response fields, or apply query-specific matching.

**scope:** In scope: hover metadata fields on symbol facts, occurrence override documentation fields, preservation of empty or absent documentation, distinction between symbol-level and occurrence-level documentation, and colocated unit tests. Out of scope: relationship lookup behavior, command-specific documentation precedence, Markdown rendering, signature pretty-printing, final CLI JSON schemas, query-specific matching/filtering, source-file reads, raw path loading, custom persisted formats, daemon/watch/cache behavior, and fixture authoring.

**spec_ref:** specs/stories/readme/20260517-100328-epic-planning-2/CAP-002-02-hover-metadata-access.md

**Depends on:** Task 3.

**Planned files:**
- `internal/traversal/facts.go`
- `internal/traversal/view.go`
- `internal/traversal/view_test.go`

**Implementation notes:**
- Preserve signature documentation as structured SCIP data as exposed by the official binding; do not collapse it into plain display text.
- Keep occurrence override documentation attached to the occurrence rather than merging it into symbol metadata.
- Validation command for the task: `go test ./internal/traversal`.

## Dependency Plan

Task 1 depends on `epic-planning-1-architecture-code-planning-1` because traversal must consume the runtime loaded SCIP index boundary instead of recreating loading. Task 2 depends on Task 1 because symbol and occurrence facts are exposed through the traversal view and document inventory. Task 3 depends on Task 2 because range and role metadata enrich occurrence facts. Task 4 depends on Task 3 because hover metadata enriches the same symbol and occurrence facts and shares the same production files.

The dependency chain is intentionally serial: all tasks touch `internal/traversal/view.go`, `internal/traversal/facts.go`, and `internal/traversal/view_test.go`, so parallel coding would create avoidable merge conflicts.

## Shared-File Audit

| File or package | Tasks | Dependency |
|-----------------|-------|------------|
| `internal/traversal/view.go` | Task 1, Task 2, Task 3, Task 4 | Task 1 -> Task 2 -> Task 3 -> Task 4 |
| `internal/traversal/facts.go` | Task 1 optional, Task 2, Task 3, Task 4 | Task 1 -> Task 2 -> Task 3 -> Task 4 |
| `internal/traversal/view_test.go` | Task 1, Task 2, Task 3, Task 4 | Task 1 -> Task 2 -> Task 3 -> Task 4 |

No two sibling tasks in this plan modify the same file without a dependency chain.

## Test Impact

Task 1 adds unit tests for the traversal builder and index/document inventory. Task 2 adds unit tests for document-level symbols, external symbols, occurrence inventories, symbol source tagging, empty inventories, and deterministic enumeration. Task 3 adds unit tests for occurrence document context, range forms, enclosing-range presence/absence, role bitsets, coordinate preservation, and source-file-read avoidance. Task 4 adds unit tests for symbol hover metadata, signature documentation, enclosing symbol metadata, occurrence override documentation, and absence handling.

No e2e command tests are planned in this scope because traversal is an internal package and must not define command stdout, stderr, status, or final JSON behavior. Query command e2e coverage belongs to sibling query plans after traversal facts exist.

## Doc Impact

No README or user-facing documentation task is planned. This scope adds an internal traversal API and the source requirements are already captured in the epic, architecture plan, story documents, and this code plan. User-visible command behavior and response schemas are explicitly out of scope.

## Spec Compliance Matrix

| # | Requirement | Source | Task(s) | Status |
|---|-------------|--------|---------|--------|
| 1 | Traversal starts from the shared runtime loaded-index boundary, not from a raw path or protobuf parser owned by traversal. | CAP-001-01 ST-001 AC-001-1, AC-001-2; architecture Scope 1 Boundary | Task 1 | Covered |
| 2 | Traversal keeps command routing, flag parsing, path validation, malformed-index diagnostics, stdout, stderr, and process status in the shared runtime boundary. | CAP-001-01 ST-001 AC-001-3; assigned scope out of scope | Task 1 | Covered |
| 3 | Traversal uses official SCIP Go binding data as the source of truth and does not introduce a custom index format. | README.md#language-support; CAP-001-01 NFR-000-1; architecture Constraints | Task 1, Task 2, Task 3, Task 4 | Covered |
| 4 | Traversal remains process-local and does not add daemon, watcher, incremental refresh, cache, or persisted derived indexes. | CAP-001-01 NFR-000-3; README.md#out-of-scope | Task 1, Task 2, Task 3, Task 4 | Covered |
| 5 | Query planners can inspect index-level metadata available through the official SCIP index binding. | CAP-001-01 ST-002 AC-002-1; assigned done_when | Task 1 | Covered |
| 6 | Every SCIP document from the loaded index is represented once in the document inventory. | CAP-001-01 ST-002 AC-002-2; CAP-001-02 ST-001 AC-001-1 | Task 1 | Covered |
| 7 | Empty document inventories remain valid traversal data rather than query failures. | CAP-001-01 ST-002 AC-002-4b; assigned done_when | Task 1 | Covered |
| 8 | Document facts preserve document identity, relative path, language, position encoding, and empty or populated occurrence collections. | CAP-001-02 ST-001 AC-001-1, AC-001-2, AC-001-2b, AC-001-3; CAP-002-01 ST-001 AC-001-1, AC-001-3 | Task 1 | Covered |
| 9 | Traversal does not hide, merge, filter, or result-shape documents for any specific command. | CAP-001-02 ST-001 AC-001-4; CAP-002-01 ST-001 AC-001-5 | Task 1 | Covered |
| 10 | Every document-level symbol information entry is available from traversal. | CAP-001-02 ST-002 AC-002-1; assigned done_when | Task 2, Task 4 | Covered |
| 11 | Every external symbol information entry is available from traversal. | CAP-001-01 ST-002 AC-002-3; CAP-001-02 ST-002 AC-002-2; assigned done_when | Task 2, Task 4 | Covered |
| 12 | Symbol facts preserve full SCIP symbol strings and distinguish document-local from external source. | CAP-001-02 ST-002 AC-002-4; architecture Traversal Facts; assigned done_when | Task 2, Task 4 | Covered |
| 13 | Empty document-symbol, external-symbol, and occurrence inventories are exposed as empty collections without placeholders. | CAP-001-02 ST-002 AC-002-4b | Task 2 | Covered |
| 14 | Every document occurrence is available as a traversal fact associated with its containing document and referenced SCIP symbol string. | CAP-001-02 ST-002 AC-002-3; assigned done_when | Task 2 | Covered |
| 15 | Inventories are deterministic over the same loaded index for repeatable downstream planning. | CAP-001-02 NFR-000-2 | Task 2 | Covered |
| 16 | Traversal preserves occurrence containing document path and document position encoding for source-location interpretation. | CAP-002-01 ST-001 AC-001-2, AC-001-3; assigned done_when | Task 3 | Covered |
| 17 | Traversal preserves exact SCIP occurrence ranges, including distinguishable three-integer same-line and four-integer multi-position forms. | CAP-002-01 ST-002 AC-002-1, AC-002-2; SCIP schema range docs | Task 3 | Covered |
| 18 | Traversal preserves enclosing range presence separately from absence and does not invent defaults. | CAP-002-01 ST-002 AC-002-3, AC-002-3b; assigned done_when | Task 3 | Covered |
| 19 | Traversal preserves raw SCIP symbol-role bitsets without deciding command meanings. | CAP-002-01 ST-002 AC-002-4, AC-002-5; assigned done_when | Task 3 | Covered |
| 20 | Traversal does not convert source coordinates to editor-specific or command-specific conventions. | CAP-002-01 NFR-000-2; ST-001 AC-001-5; ST-002 AC-002-5 | Task 3 | Covered |
| 21 | Traversal does not read source files to supplement absent or empty SCIP document text. | CAP-002-01 ST-001 AC-001-4; assigned scope | Task 3, Task 4 | Covered |
| 22 | Symbol facts preserve kind and display name for both document-level and external symbols. | CAP-002-02 ST-001 AC-001-1, AC-001-2; assigned done_when | Task 4 | Covered |
| 23 | Symbol facts preserve documentation entries and absent or empty documentation without inventing docs. | CAP-002-02 ST-001 AC-001-3, AC-001-3b | Task 4 | Covered |
| 24 | Symbol facts preserve signature documentation as SCIP data and enclosing symbol metadata. | CAP-002-02 ST-001 AC-001-4, AC-001-5; assigned done_when | Task 4 | Covered |
| 25 | Occurrence facts preserve override documentation when present. | CAP-002-02 ST-002 AC-002-1; assigned done_when | Task 4 | Covered |
| 26 | Occurrence override documentation remains distinguishable from symbol-level documentation. | CAP-002-02 ST-002 AC-002-2, AC-002-3 | Task 4 | Covered |
| 27 | Absent occurrence override documentation remains absent or empty without copying symbol documentation. | CAP-002-02 ST-002 AC-002-2b | Task 4 | Covered |
| 28 | Traversal does not render Markdown, pretty-print signatures, choose command response fields, or apply query-specific symbol matching. | CAP-002-02 NFR-000-2; ST-001 AC-001-6; ST-002 AC-002-4 | Task 4 | Covered |
| 29 | Traversal does not resolve partial names, package prefixes, reference semantics, implementation semantics, result ordering, grouping, duplicate policy, or final JSON schemas. | CAP-001-01 NFR-000-2; CAP-001-02 NFR-000-3; assigned scope | Task 1, Task 2, Task 3, Task 4 | Covered |
| 30 | Relationship lookup indexes and fixture authoring remain out of this plan. | Architecture Scope 1 Out of scope; assigned scope | Task 1, Task 2, Task 3, Task 4 | Covered |
| E2E | e2e test coverage for new behavior | Cross-cutting | N/A: internal traversal facts only; command e2e behavior is owned by sibling query/runtime plans after traversal exists. | N/A |
| DOC | Documentation updates for changed behavior | Cross-cutting | N/A: internal package contract is documented by this plan and existing epic/story specs; no user-visible command behavior or README schema changes are introduced. | N/A |

## Pre-Submit Validation Checklist

- Re-read this plan and verify the output JSON fields are character-identical to each task's `desc`, `done_when`, `scope`, and `spec_ref`.
- Run `jq . specs/plans/readme/20260517-194839-epic-planning-2-architecture-code-planning-0-output.json`.
- Search this plan for every `Task N` cross-reference and confirm the referenced task states the corresponding responsibility, dependency, or exclusion.
- Confirm every shared file listed in more than one task has a dependency chain.
- Run pre-commit on the plan and output JSON files.
- Commit only the plan and output JSON artifacts.
