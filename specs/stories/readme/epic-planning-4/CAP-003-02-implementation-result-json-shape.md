# User Stories: Implementation Result JSON Shape

Status: review

## Goal
Successful `implementations --symbol` payloads expose stable implementation entries with implementation symbols, incoming relationship basis, and definition document paths and SCIP ranges when traversal data provides them.

## Parent Epic
specs/epics/readme/20260517-141006-epic-planning-4.md - Capability CAP-003

## Context
Implementation relationship selection identifies which implementation symbols qualify for a query. This document defines the automation-facing successful JSON payload for those results, including stable ordering, empty-result behavior, and handling implementation symbols with and without available definition locations.

## Personas
- **Automation Agent**: An AI or script-driven caller parsing `scip-search` output in a terminal or sandbox, needing stable implementation symbols and source locations it can feed into later navigation or editing workflows.
- **CLI Maintainer**: A Go developer maintaining `scip-search`, needing a bounded query-specific JSON contract that does not infer source locations absent from SCIP traversal data.

## General information

Applies to: successful JSON payloads for exact `implementations --symbol ... --json` queries after implementation relationship selection.

### References
- goal spec: README.md#what-is-scip-search - Documents `implementations --symbol` and explicit `--json` output mode.
- goal spec: README.md#language-support - Requires official SCIP bindings and direct SCIP output reads.
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#general-information - Requires stable payloads, explicit empty collections, document-relative paths, 0-based SCIP ranges, and no source-file reads to infer missing locations.
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#capability-cap-003---return-implementation-relationship-results - Requires implementation symbols, incoming relationship basis, definition locations when available, document paths, and ranges when present.
- official source: https://raw.githubusercontent.com/scip-code/scip/main/scip.proto - Defines occurrence ranges, document paths, symbol roles, and implementation relationships.
- dependency story: specs/stories/readme/epic-planning-4/CAP-001-01-exact-symbol-input.md - Defines top-level queried-symbol preservation for successful `implementations --symbol` payloads.
- dependency story: specs/stories/readme/epic-planning-4/CAP-001-02-missing-symbol-results.md - Defines successful empty `implementations` collections for missing and no-match exact-symbol outcomes.
- sibling story: specs/stories/readme/epic-planning-4/CAP-003-01-implementation-relationship-selection.md - Defines which incoming SCIP implementation relationships produce implementation entries.
- consistency check: specs/stories/readme/20260517-134857-epic-planning-3/CAP-002-03-package-result-json-shape.md - Provides precedent for query-specific JSON fields under the shared success contract.

### Non-Functional Requirements
- NFR-000-1: Successful implementation payloads must be stable for automation across repeated runs over the same index, including deterministic entry ordering.
- NFR-000-2: Implementation entry fields must preserve SCIP symbol strings, document-relative paths, and schema-defined 0-based ranges exactly as provided by traversal data.
- NFR-000-3: Missing implementation definition locations must not be filled by reading source files, guessing descriptor locations, or synthesizing ranges.
- NFR-000-4: These stories must not redefine shared stdout/stderr stream behavior, process status, shared runtime failures, traversal view construction, reference result schemas, symbol discovery, package discovery, or fixture coverage.

### Related External Components
- Component C-001 - SCIP symbol identifiers: Exact full symbol strings supplied through `--symbol` and returned in implementation result payloads.
- Component C-002 - SCIP occurrence data: Document-relative source ranges and definition roles used to locate implementation symbols when available.
- Component C-003 - SCIP relationship data: Incoming implementation relationships that explain why each implementation entry is present.
- Component C-004 - Calling process environment: The shell, script, or agent that parses successful JSON stdout under the shared runtime contract.

### Interfaces
- I-001-002 - Implementation query contract (Interface 002 of Component C-001): The `implementations --symbol` query returns a top-level `symbol` value and an `implementations` collection for the queried exact symbol.
- I-002-001 - Implementation location source (Interface 001 of Component C-002): Traversal data exposes definition document paths and SCIP ranges for implementation symbols when those occurrences are available.
- I-003-001 - Implementation relationship source (Interface 001 of Component C-003): Traversal data exposes the incoming implementation relationship basis for each implementation entry.
- I-004-001 - CLI process contract (Interface 001 of Component C-004): The shared process contract that exposes successful JSON-producing query payloads as parseable JSON on stdout.

### Out of Scope
- Shared invocation errors, index-loading errors, stderr diagnostics, stdout stream purity, and process status taxonomy.
- Selecting which relationships qualify as implementation results beyond the sibling selection story.
- Source-file reads, synthesized locations, hover text, reference occurrences, type-definition traversal, full hierarchy graphs, package dependency graphs, semantic similarity, or default one-line output.
- Golden fixture authoring and fixture generation.

### Assumptions
- **ASM-000-1**: Implementation entries use `implementationSymbol` for the implementer symbol, `relationship` for the incoming relationship basis, `documentPath` for the definition document when available, and `range` for the SCIP range when available. - *Why*: CAP-003 requires these payload concepts, and explicit names keep the command-specific JSON contract parseable for automation. - Confidence: MEDIUM
- **ASM-000-2**: When a qualifying implementation symbol has no available definition location, its entry omits `documentPath` and `range` rather than using guessed, empty-string, or sentinel values. - *Why*: The parent epic forbids source-file reads to infer missing locations and requires ranges only when present. - Confidence: MEDIUM

### Open Questions
- None.

---

## Story ST-001 - Expose Implementation Entries in Stable JSON

### References
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#capability-cap-003---return-implementation-relationship-results - Requires implementation symbols, incoming relationship basis, stable ordering, and successful payloads.
- dependency story: specs/stories/readme/epic-planning-4/CAP-001-01-exact-symbol-input.md - Defines top-level queried-symbol preservation.
- dependency story: specs/stories/readme/epic-planning-4/CAP-001-02-missing-symbol-results.md - Defines the `implementations` collection used for empty and no-match results.
- sibling story: specs/stories/readme/epic-planning-4/CAP-003-01-implementation-relationship-selection.md - Defines which selected relationships produce implementation entries.

### User Story
**As an** Automation Agent parsing implementation-query output, **I want to** receive a stable `implementations` collection with implementation symbols and their incoming relationship basis, **so that** I can compare and navigate implementation results deterministically.

### Acceptance Criteria
- AC-001-1: Given a successful `implementations --symbol ... --json` query returns one or more implementation results, when the caller parses the query-specific JSON payload, then the payload includes top-level `symbol` equal to the caller-provided queried symbol and an `implementations` collection.
- AC-001-2: Given an implementation entry appears in the `implementations` collection, when the caller inspects that entry, then it includes `implementationSymbol` equal to the qualifying implementer symbol.
- AC-001-3: Given an implementation entry appears in the `implementations` collection, when the caller inspects that entry, then it includes a `relationship` value that identifies the entry as based on an incoming SCIP implementation relationship to the queried symbol.
- AC-001-4: Given the same loaded index and queried symbol are used across repeated successful runs, when the caller compares the `implementations` collection order, then entries appear in deterministic order.
- AC-001-5: Given multiple implementation entries are returned, when the caller observes their order, then the ordering is stable by implementation symbol and, for entries with equal implementation symbols, by available definition location.
- AC-001-5b: Given the successful query has no implementation results, when the caller parses the payload, then `implementations` is present as an empty collection and no implementation diagnostics are printed in the query-specific JSON payload.

### Depends on:
Implementation ordering:
- Story document CAP-001-01-exact-symbol-input.md - Queried-symbol preservation must be defined before this payload can reference it.
- Story document CAP-001-02-missing-symbol-results.md - Empty successful `implementations` collection behavior must be defined before this document extends the non-empty payload.
- Story document CAP-003-01-implementation-relationship-selection.md - Result entries depend on selected incoming implementation relationships.

Run time coupling:
- I-001-002 - Implementation query contract
- I-003-001 - Implementation relationship source
- I-004-001 - CLI process contract

### Out of Scope
- Shared stdout stream purity, stderr behavior, process status, runtime failures, or command routing.
- Returning outgoing implemented targets, references, type definitions, hover text, package data, hierarchy graphs, or semantic matches.
- Defining fixture and golden-case coverage.

### Assumptions
- **ASM-001-1**: Stable ordering by implementation symbol, then available definition location, is sufficient for automation because implementation symbols are mandatory and locations are optional. - *Why*: CAP-003 requires stable ordering but does not prescribe a ranking signal, and deterministic lexical fields avoid relevance scoring. - Confidence: MEDIUM

### Open Questions
- None.

---

## Story ST-002 - Include Definition Locations and Ranges When Available

### References
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#general-information - Requires document-relative paths and schema-defined 0-based ranges without source-file inference.
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#capability-cap-003---return-implementation-relationship-results - Requires definition locations and ranges when available and successful handling when unavailable.
- official source: https://raw.githubusercontent.com/scip-code/scip/main/scip.proto - Defines SCIP occurrence ranges and symbol definition roles.

### User Story
**As an** Automation Agent navigating to implementation code, **I want to** receive definition document paths and SCIP ranges for implementation symbols when available, **so that** I can jump to concrete implementation locations without inspecting source files first.

### Acceptance Criteria
- AC-002-1: Given a qualifying implementation symbol has an available SCIP definition occurrence, when the caller parses its implementation entry, then the entry includes the definition `documentPath` from the loaded index.
- AC-002-2: Given a qualifying implementation symbol has an available SCIP definition occurrence with a range, when the caller parses its implementation entry, then the entry includes the SCIP `range` for that definition.
- AC-002-3: Given an implementation entry includes `documentPath` or `range`, when the caller compares those values to the loaded SCIP traversal data, then they preserve the document-relative path and schema-defined 0-based range values without source-file-derived adjustment.
- AC-002-4: Given a qualifying implementation symbol has no available definition occurrence, when the caller parses its implementation entry, then the entry still includes `implementationSymbol` and incoming `relationship` basis without failing the successful query.
- AC-002-4b: Given a qualifying implementation symbol has no available definition occurrence, when the caller inspects its implementation entry, then the entry does not include a guessed document path, guessed range, source-file-derived location, or sentinel location.
- AC-002-5: Given one implementation symbol has a definition location and another qualifying implementation symbol does not, when the caller runs `implementations --symbol`, then both qualifying implementation symbols can appear in the same successful `implementations` collection.

### Depends on:
Implementation ordering:
- Story ST-001 - Expose Implementation Entries in Stable JSON

Run time coupling:
- I-001-002 - Implementation query contract
- I-002-001 - Implementation location source
- I-003-001 - Implementation relationship source

### Out of Scope
- Reading source files to infer definition locations.
- Returning hover documentation, enclosing symbol metadata, reference locations, type definitions, or alternative location formats.
- Treating missing definition locations as shared runtime failures.

### Assumptions
- **ASM-002-1**: Definition location lookup prefers SCIP definition occurrences for the implementation symbol and does not use non-definition occurrences as substitute locations. - *Why*: CAP-003 asks for definition locations when available, and the epic's general information distinguishes SCIP definition roles from other occurrences. - Confidence: MEDIUM

### Open Questions
- None.
