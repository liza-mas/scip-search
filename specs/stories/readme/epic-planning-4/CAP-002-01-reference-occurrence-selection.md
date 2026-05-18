# User Stories: Reference Occurrence Selection

Status: review

## Goal
`references --symbol` selects only non-definition reference occurrences for the queried exact SCIP symbol and SCIP reference-related symbols.

## Parent Epic
specs/epics/readme/20260517-141006-epic-planning-4.md - Capability CAP-002

## Context
Reference queries run after shared CLI validation, index loading, and exact `--symbol` interpretation have succeeded. This document defines which indexed SCIP occurrences belong in a successful reference result set. It relies on traversal stories for occurrence and relationship lookup views, and leaves final JSON field names and ordering to the sibling CAP-002 result-shape story.

## Personas
- **Automation Agent**: An AI or script-driven caller running `scip-search` in a terminal or sandbox, needing reference locations that correspond to exact symbol use sites it can navigate to later.
- **CLI Maintainer**: A Go developer maintaining `scip-search`, needing bounded reference-selection semantics that stay aligned with SCIP occurrence roles and reference relationships.

## General information

Applies to: successful `references --symbol` occurrence selection.

### References
- goal spec: README.md#what-is-scip-search - Documents `scip-search references --index <index-path> --symbol <scip-symbol>` and successful structured JSON output.
- goal spec: README.md#complementary-existing-tool - Frames `references` as the query used to answer what calls or uses a symbol.
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#capability-cap-002---return-reference-occurrences-for-exact-symbols - Requires exact-symbol and reference-related occurrence selection, non-definition filtering, document paths, ranges, and role context.
- dependency story: specs/stories/readme/epic-planning-4/CAP-001-01-exact-symbol-input.md - Defines literal exact `--symbol` interpretation and top-level queried-symbol preservation.
- dependency story: specs/stories/readme/epic-planning-4/CAP-001-02-missing-symbol-results.md - Defines successful empty `references` collections for absent exact symbols and no command-specific matches.
- sibling boundary: specs/epics/readme/20260517-100328-epic-planning-2.md#capability-cap-003---expose-occurrence-and-relationship-lookup-views - Provides occurrence and relationship lookup views for reference planners.
- official source: https://raw.githubusercontent.com/scip-code/scip/main/scip.proto - Defines occurrence ranges, `SymbolRole.Definition`, and `Relationship.is_reference`.
- consistency scan: specs/stories/readme/epic-planning-4/ - Existing CAP-001 stories define exact-symbol and empty-result behavior but leave reference occurrence selection to this document.

### Non-Functional Requirements
- NFR-000-1: Reference selection must preserve full SCIP symbol strings exactly as supplied by the caller or exposed by traversal.
- NFR-000-2: Reference selection must use SCIP occurrence role information instead of source-file reads, snippets, or text search to distinguish references from definitions.
- NFR-000-3: Reference selection must not redefine command routing, shared flags, index loading, stdout/stderr stream rules, process status, raw SCIP parsing, traversal lookup construction, or final JSON formatting.

### Related External Components
- Component C-001 - SCIP symbol identifiers: Exact full symbol strings supplied through `--symbol` and attached to occurrences.
- Component C-002 - SCIP occurrence data: Document-contained symbol occurrences with schema-defined ranges and role bitsets.
- Component C-003 - SCIP relationship data: Symbol relationship edges that identify reference-related symbols for Find References behavior.

### Interfaces
- I-001-001 - Reference query contract (Interface 001 of Component C-001): The `references --symbol` query selects reference occurrences for the queried exact symbol.
- I-002-001 - Occurrence lookup view (Interface 001 of Component C-002): Query planning can inspect occurrences with their symbol, document, range, and role context.
- I-003-001 - Reference relationship lookup view (Interface 001 of Component C-003): Query planning can inspect SCIP relationships whose `is_reference` flag relates symbols for reference selection.

### Out of Scope
- Shared missing-flag, missing-value, unsupported-command, unreadable-index, malformed-index, stderr, stdout, and exit-status behavior.
- Partial `symbols --name` discovery, ambiguous partial-name resolution, fuzzy matching, regex matching, semantic similarity, ranking by relevance, or cross-index lookup.
- Source-file reads, snippet extraction, caller-function grouping, call hierarchy synthesis, graph visualization, or semantic ranking.
- Raw traversal lookup construction, SCIP protobuf parsing, hover extraction, fixture construction, final JSON field names, and implementation relationship results.

### Assumptions
- **ASM-000-1**: SCIP `is_reference` relationships are followed in either direction when traversal exposes the relationship set. - *Why*: The parent epic states that reference-related symbols should be included together for Find References behavior. - Confidence: MEDIUM
- **ASM-000-2**: An occurrence carrying `SymbolRole.Definition` is excluded from reference results even if it otherwise matches the selected exact or reference-related symbol set. - *Why*: The capability requires non-definition reference occurrences, and definition locations are symbol context rather than use sites. - Confidence: HIGH
- **ASM-000-3**: Duplicate reference relationships and duplicate candidate-symbol paths do not create duplicate reference entries. - *Why*: Automation expects occurrence results, not relationship-path results. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Build the Reference Candidate Symbol Set

### References
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#capability-cap-002---return-reference-occurrences-for-exact-symbols - Requires the query to include the exact queried symbol and SCIP reference-related symbols.
- dependency story: specs/stories/readme/epic-planning-4/CAP-001-01-exact-symbol-input.md - Defines the supplied `--symbol` value as one literal exact full SCIP symbol.
- sibling boundary: specs/epics/readme/20260517-100328-epic-planning-2.md#capability-cap-003---expose-occurrence-and-relationship-lookup-views - Provides relationship lookup views.

### User Story
**As an** Automation Agent running a follow-up reference query, **I want to** search references for the exact SCIP symbol I supplied and its SCIP reference-related symbols, **so that** references remain complete for symbol aliases or relationship-grouped symbols recorded by the index.

### Acceptance Criteria
- AC-001-1: Given the caller invokes `references --symbol` with a full SCIP symbol after shared runtime validation, when reference selection starts, then the candidate symbol set includes the exact caller-provided symbol.
- AC-001-2: Given traversal exposes a SCIP relationship with `is_reference` that connects the queried symbol to another full SCIP symbol, when reference selection builds the candidate symbol set, then the connected symbol is included as a reference-related candidate.
- AC-001-2b: Given traversal exposes the `is_reference` relationship with the queried symbol on either side of the relationship, when reference selection builds the candidate symbol set, then the connected symbol is included the same way.
- AC-001-3: Given duplicate relationships, cycles, or multiple relationship paths connect the same reference-related symbol, when candidate symbols are resolved, then each full SCIP symbol is represented once in the candidate set.
- AC-001-4: Given the loaded index has no exact symbol match and no reference-related symbols for the supplied value, when the query completes successfully under the shared runtime contract, then reference selection produces no selected occurrences rather than falling back to partial, fuzzy, regex, semantic, or containing-substring matches.

### Depends on:
Implementation ordering:
- Story document CAP-001-01-exact-symbol-input.md - Reference selection starts from the literal exact symbol query input.

Run time coupling:
- I-001-001 - Reference query contract
- I-003-001 - Reference relationship lookup view

### Out of Scope
- Defining the final `references` JSON entry fields.
- Defining shared runtime diagnostics or process status for invalid invocations or index-loading failures.
- Expanding candidate symbols through partial-name discovery, package discovery, text search, type hierarchy analysis, or source-file reads.

### Assumptions
- **ASM-001-1**: The candidate set is based on full SCIP symbol equality, not display names or descriptor fragments. - *Why*: CAP-001 requires exact full-symbol behavior, and this story only adds SCIP reference relationships to that exact base. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-002 - Select Non-Definition Occurrences

### References
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#capability-cap-002---return-reference-occurrences-for-exact-symbols - Requires source occurrences whose symbol matches the exact queried symbol or a reference-related symbol and whose role is not a definition.
- official source: https://raw.githubusercontent.com/scip-code/scip/main/scip.proto - Defines occurrence ranges and `SymbolRole.Definition`.
- sibling boundary: specs/epics/readme/20260517-100328-epic-planning-2.md#capability-cap-003---expose-occurrence-and-relationship-lookup-views - Provides occurrence lookup views.

### User Story
**As a** CLI Maintainer implementing reference traversal, **I want to** include only non-definition occurrences whose symbols are in the resolved reference candidate set, **so that** `references --symbol` returns use sites without mixing in definition locations.

### Acceptance Criteria
- AC-002-1: Given traversal exposes an occurrence whose symbol equals a candidate symbol and whose role context does not include `SymbolRole.Definition`, when reference selection runs, then that occurrence is selected as a reference result.
- AC-002-2: Given traversal exposes an occurrence whose symbol equals a candidate symbol and whose role context includes `SymbolRole.Definition`, when reference selection runs, then that occurrence is excluded from reference results.
- AC-002-3: Given traversal exposes an occurrence whose symbol is not in the candidate symbol set, when reference selection runs, then that occurrence is excluded from reference results even if it is a non-definition occurrence.
- AC-002-4: Given multiple documents contain selected reference occurrences for the candidate symbol set, when reference selection completes, then all selected occurrences remain associated with their traversal-provided document paths, SCIP ranges, matched occurrence symbols, and role context for result formatting.
- AC-002-5: Given all matching occurrences are definitions or the candidate symbols have no occurrences, when the query completes successfully under the shared runtime contract, then reference selection produces an empty selected-occurrence collection.

### Depends on:
Implementation ordering:
- Story ST-001 - Build the Reference Candidate Symbol Set

Run time coupling:
- I-001-001 - Reference query contract
- I-002-001 - Occurrence lookup view

### Out of Scope
- Defining definition-query behavior, implementation-query behavior, hover output, snippets, or source-file enrichment.
- Defining final result ordering or JSON field names.
- Defining fixtures or golden JSON cases.

### Assumptions
- **ASM-002-1**: Occurrences selected through a reference-related symbol retain the matched occurrence symbol rather than being rewritten to the originally queried symbol. - *Why*: The parent capability requires matched symbols in reference entries, and automation may need to distinguish exact-symbol matches from relationship-expanded matches. - Confidence: HIGH

### Open Questions
- None.
