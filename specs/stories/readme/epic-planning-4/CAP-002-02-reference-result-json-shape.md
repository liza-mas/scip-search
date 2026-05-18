# User Stories: Reference Result JSON Shape

Status: review

## Goal
Successful `references --symbol` responses expose a stable `references` payload whose entries identify matched symbols, document paths, SCIP ranges, and role context.

## Parent Epic
specs/epics/readme/20260517-141006-epic-planning-4.md - Capability CAP-002

## Context
The shared runtime contract owns stdout stream purity on success, diagnostics on failure, and process status. CAP-001 owns top-level queried-symbol preservation and successful empty no-match responses. This document defines only the `references --json` payload fields and ordering needed once reference occurrence selection has produced zero or more selected occurrences.

## Personas
- **Automation Agent**: An AI or script-driven caller parsing `scip-search references` output, needing stable source locations and matched symbols for later code-navigation or editing workflows.
- **CLI Maintainer**: A Go developer maintaining `scip-search`, needing a bounded query-specific JSON contract that can be validated by deterministic golden cases.

## General information

Applies to: successful `references --symbol ... --json` payloads.

### References
- goal spec: README.md#what-is-scip-search - Documents the `references --symbol` command form and explicit `--json` output mode.
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#capability-cap-002---return-reference-occurrences-for-exact-symbols - Requires reference entries with matched occurrence symbol, document path, SCIP range, role context, and originally queried symbol in the response envelope.
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#general-information - Requires deterministic ordering and explicit empty result collections for successful no-match cases.
- dependency story: specs/stories/readme/epic-planning-4/CAP-001-01-exact-symbol-input.md - Defines top-level queried-symbol preservation.
- dependency story: specs/stories/readme/epic-planning-4/CAP-001-02-missing-symbol-results.md - Defines empty `references` collections for no-match outcomes.
- dependency story: specs/stories/readme/epic-planning-4/CAP-002-01-reference-occurrence-selection.md - Defines the selected occurrence set that this payload exposes.
- consistency check: specs/stories/readme/20260517-134857-epic-planning-3/CAP-002-03-package-result-json-shape.md - Provides same-project precedent for query-specific JSON field assumptions under the shared stdout contract.
- official source: https://raw.githubusercontent.com/scip-code/scip/main/scip.proto - Defines document-relative occurrence ranges and occurrence role data.
- consistency scan: specs/stories/readme/epic-planning-4/ - Existing CAP-001 stories define exact-symbol and empty-result behavior but leave non-empty reference entry fields and ordering to this document.

### Non-Functional Requirements
- NFR-000-1: Reference result JSON must be deterministic for repeated runs over the same loaded SCIP index and query symbol.
- NFR-000-2: Reference result JSON must expose source locations from SCIP data without requiring callers to read source files or infer ranges from snippets.
- NFR-000-3: Reference result stories must not redefine shared stream placement, diagnostic behavior, process status, command routing, index loading, reference occurrence selection, or implementation result schemas.

### Related External Components
- Component C-001 - SCIP symbol identifiers: Exact full symbol strings supplied through `--symbol` and attached to selected reference occurrences.
- Component C-002 - SCIP occurrence data: Document-contained occurrences with schema-defined ranges and role context.
- Component C-003 - Calling process environment: The shell, script, or agent that parses successful JSON stdout under the shared runtime contract.

### Interfaces
- I-001-001 - Reference query contract (Interface 001 of Component C-001): The `references --symbol` query returns a `references` collection for the queried exact symbol.
- I-002-001 - Selected reference occurrence contract (Interface 001 of Component C-002): Selected occurrences expose matched symbol, document path, SCIP range, and role context for JSON formatting.
- I-003-001 - CLI process contract (Interface 001 of Component C-003): The shared successful runtime contract provides a parseable JSON value on stdout for explicit JSON-producing modes.

### Out of Scope
- Shared successful stdout stream rules, stderr behavior, process status, runtime failures, command routing, or index-path behavior.
- Selecting reference occurrences, following reference relationships, interpreting exact input, or defining no-match success as a command outcome.
- Result schemas for `symbols`, `packages`, or `implementations`.
- Source snippets, hover text, caller-function grouping, call hierarchy synthesis, semantic ranking, graph output, or default one-line output.

### Assumptions
- **ASM-000-1**: The reference query result is a JSON object with top-level `symbol` and `references` fields. - *Why*: CAP-001 already assumes top-level `symbol`, and CAP-002 needs a command-specific collection for reference entries. - Confidence: MEDIUM
- **ASM-000-2**: Reference entry fields are named `symbol`, `documentPath`, `range`, and `roles`. - *Why*: These names expose the parent epic's matched occurrence symbol, document path, SCIP range, and role context in automation-readable fields while staying consistent with camelCase query payload precedent. - Confidence: MEDIUM
- **ASM-000-3**: `range` preserves SCIP's document-relative, 0-based range values as a single field rather than converting them to editor- or source-file-derived coordinates. - *Why*: The parent epic requires SCIP range semantics and excludes source-file reads. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Expose Reference Entries in JSON

### References
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#capability-cap-002---return-reference-occurrences-for-exact-symbols - Requires reference entries with matched occurrence symbol, document path, SCIP range, and role context.
- dependency story: specs/stories/readme/epic-planning-4/CAP-002-01-reference-occurrence-selection.md - Defines which occurrences are selected before JSON formatting.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/stdout-json-success-contract.md - Owns stdout stream purity and leaves query-specific fields to this document.

### User Story
**As an** Automation Agent parsing successful reference output, **I want to** receive each selected reference occurrence with its matched symbol, document path, SCIP range, and role context, **so that** I can navigate to the use site without reparsing the SCIP index or reading source files.

### Acceptance Criteria
- AC-001-1: Given a successful `references --symbol ... --json` query returns one or more selected reference occurrences, when the caller parses the query-specific JSON payload, then the top-level payload contains `symbol` equal to the caller-provided query symbol and a `references` collection.
- AC-001-2: Given a selected reference occurrence appears in the `references` collection, when the caller inspects that entry, then it exposes `symbol`, `documentPath`, `range`, and `roles`.
- AC-001-3: Given a reference entry was selected from the exact queried symbol or a reference-related symbol, when the caller reads the entry's `symbol`, then it equals the matched occurrence symbol from traversal, not a display name, descriptor fragment, or rewritten query symbol.
- AC-001-4: Given a reference entry includes `documentPath`, when the caller compares it to traversal data, then it preserves the SCIP document path associated with the occurrence.
- AC-001-5: Given a reference entry includes `range`, when the caller compares it to traversal data, then it preserves the SCIP occurrence range using SCIP document-relative 0-based range semantics.
- AC-001-6: Given a reference entry includes `roles`, when the caller inspects it, then it exposes the occurrence role context used for selection and does not include a selected entry whose roles contain `SymbolRole.Definition`.

### Depends on:
Implementation ordering:
- Story document CAP-002-01-reference-occurrence-selection.md - The selected reference occurrence set must be defined before its JSON fields can be validated.

Run time coupling:
- I-001-001 - Reference query contract
- I-002-001 - Selected reference occurrence contract
- I-003-001 - CLI process contract

### Out of Scope
- Defining shared stdout, stderr, diagnostics, or process status.
- Adding snippets, hover text, source-file content, caller function names, ranking scores, or graph edges to reference entries.
- Rewriting matched symbols to the queried symbol for reference-related matches.

### Assumptions
- **ASM-001-1**: `roles` can be empty when SCIP exposes no non-definition role bits for a selected occurrence. - *Why*: The selection rule only requires absence of `SymbolRole.Definition`; an unlabelled non-definition occurrence is still a reference occurrence in SCIP use-site terms. - Confidence: MEDIUM

### Open Questions
- None.

---

## Story ST-002 - Preserve Stable Reference Result Ordering and Empty Payloads

### References
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#capability-cap-002---return-reference-occurrences-for-exact-symbols - Requires deterministic reference occurrence results and empty successful results.
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#general-information - Requires deterministic ordering and explicit empty result collections for successful no-match cases.
- dependency story: specs/stories/readme/epic-planning-4/CAP-001-02-missing-symbol-results.md - Defines successful empty `references` collections for no-match outcomes.

### User Story
**As an** Automation Agent comparing reference results across repeated runs, **I want to** receive reference entries in a deterministic order with explicit empty collections, **so that** automation can diff and consume results without treating ordering noise or absence as failure.

### Acceptance Criteria
- AC-002-1: Given a successful `references --symbol` query returns multiple reference entries from the same loaded index, when the caller parses the `references` collection, then entries are ordered deterministically by `documentPath`, then range start position, then range end position, then matched `symbol`.
- AC-002-2: Given two successful `references --symbol` runs use the same loaded index and query symbol, when the selected reference occurrences are unchanged, then the order and field values of the `references` collection are the same across both payloads.
- AC-002-3: Given reference selection produces no selected occurrences because the exact symbol is absent, has no references, or only has definition occurrences, when the query completes successfully under the shared runtime contract, then the payload includes top-level `symbol` equal to the caller-provided value and `references` as an empty collection.
- AC-002-4: Given a successful empty reference result is emitted, when the caller parses stdout under the shared runtime contract, then the absence of references is represented by the empty `references` collection rather than omitted fields, human diagnostic text, stderr output, or a shared runtime failure.
- AC-002-5: Given selected occurrences have identical document path and range values but different matched symbols, when the payload is ordered, then the matched `symbol` tie-breaker keeps ordering stable.

### Depends on:
Implementation ordering:
- Story ST-001 - Expose Reference Entries in JSON
- Story document CAP-001-02-missing-symbol-results.md - Empty reference collection semantics must remain consistent with the shared symbol-query no-match story.

Run time coupling:
- I-001-001 - Reference query contract
- I-003-001 - CLI process contract

### Out of Scope
- Defining shared success stream purity, failure diagnostics, or exit statuses.
- Ranking references by semantic relevance, caller function, recency, package, language, or relationship distance.
- Defining fixture files or golden JSON case locations.

### Assumptions
- **ASM-002-1**: Sorting by source location before matched symbol is the observable stable order for non-empty reference results. - *Why*: It gives automation a predictable navigation order without introducing semantic ranking, and it uses fields already required by the parent capability. - Confidence: MEDIUM

### Open Questions
- None.
