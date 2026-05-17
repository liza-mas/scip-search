# User Stories: Missing Exact Symbol Results

Status: review

## Goal
`references --symbol` and `implementations --symbol` return successful empty result payloads when the caller provides an exact full SCIP symbol that has no matches in the loaded index.

## Parent Epic
specs/epics/readme/20260517-141006-epic-planning-4.md - Capability CAP-001

## Context
A caller can legitimately ask about a full SCIP symbol that is absent from the selected index or has no query-specific related results. This document defines that no-match outcome as successful query data for both symbol-based commands, while leaving shared invocation and index-loading failures to the runtime stories.

## Personas
- **Automation Agent**: An AI or script-driven caller running `scip-search` in a terminal or sandbox, needing to distinguish successful absence from failed command execution.
- **CLI Maintainer**: A Go developer maintaining `scip-search`, needing one no-match contract shared by reference and implementation queries without expanding command-specific traversal semantics.

## General information

Applies to: successful no-match payloads for exact `references --symbol` and `implementations --symbol` queries.

### References
- goal spec: README.md#scip-symbol-format - Defines full SCIP symbol strings that can be passed from discovery to later symbol-based commands.
- goal spec: README.md#what-is-scip-search - Requires query commands to print structured JSON and documents `references --symbol` and `implementations --symbol`.
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#capability-cap-001---resolve-exact-symbol-query-inputs - Requires absent exact symbols to be successful no-match results with empty command-specific collections.
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#general-information - Assumes successful exact-symbol queries with no matching symbol or related results return empty result collections instead of shared runtime errors.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/stdout-json-success-contract.md - Owns successful JSON stdout stream behavior.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/stderr-exit-error-contract.md - Owns shared runtime failures and keeps them separate from successful empty results.

### Non-Functional Requirements
- NFR-000-1: Empty exact-symbol results must be deterministic and machine-readable, with explicit empty collections rather than omitted result fields or human diagnostic text.
- NFR-000-2: Empty exact-symbol result payloads must preserve the queried full SCIP symbol exactly so automation can record which query produced the absence result.
- NFR-000-3: These stories must not redefine shared invocation errors, shared index-loading errors, stderr diagnostics, process status taxonomy, command-specific reference occurrence selection, implementation traversal, or fixture coverage.

### Related External Components
- Component C-001 - SCIP symbol identifiers: Exact full symbol strings supplied through `--symbol` and returned in successful no-match payloads.
- Component C-003 - Calling process environment: The shell, script, or agent invoking `scip-search` and distinguishing successful empty JSON from shared runtime failures.

### Interfaces
- I-001-001 - Reference query contract (Interface 001 of Component C-001): The `references --symbol` query returns a `references` collection for the queried exact symbol.
- I-001-002 - Implementation query contract (Interface 002 of Component C-001): The `implementations --symbol` query returns an `implementations` collection for the queried exact symbol.
- I-003-001 - CLI process contract (Interface 001 of Component C-003): The shared process contract that exposes successful query payloads as parseable JSON on stdout and shared runtime failures outside this scope.

### Out of Scope
- Shared missing-flag, missing-value, unsupported-command, unreadable-index, malformed-index, stderr, stdout, and exit-status behavior.
- Partial `symbols --name` discovery, ambiguous partial-name resolution, fuzzy matching, regex matching, semantic similarity, ranking by relevance, or cross-index lookup.
- Selecting reference occurrences, including related reference symbols, traversing implementation relationships, resolving implementation definition locations, ordering non-empty results, or defining fixture/golden coverage.
- Source-file reads to infer missing locations, alternate output formats, daemon/watch behavior, MCP behavior, UI, embeddings, vector search, or ctags fallback behavior.

### Assumptions
- **ASM-000-1**: The no-match payload shape uses top-level command-specific collections named `references` and `implementations`. - *Why*: The documented command names provide clear automation-facing collection names, while sibling CAP-002 and CAP-003 stories can define non-empty entry fields. - Confidence: MEDIUM
- **ASM-000-2**: A loaded index that contains the queried symbol but has no command-specific matches uses the same successful empty collection shape as a loaded index with no exact symbol match. - *Why*: CAP-001 groups missing exact symbols and missing related results as successful no-match outcomes, and automation observes both as empty result data. - Confidence: MEDIUM

### Open Questions
- None.

---

## Story ST-001 - Return an Empty references Collection for Missing Exact Symbols

### References
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#capability-cap-001---resolve-exact-symbol-query-inputs - Requires exact symbols absent from the traversal view to return successful no-match results.
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#general-information - Defines empty successful result collections for no matching symbol or related results.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/stderr-exit-error-contract.md - Owns shared runtime failures, which this story must not redefine.

### User Story
**As an** Automation Agent running a reference query against a selected index, **I want to** receive a successful empty `references` payload when my exact full symbol has no reference matches, **so that** automation can treat absence as data instead of a command failure.

### Acceptance Criteria
- AC-001-1: Given the caller invokes `references --symbol` with a syntactically valid full SCIP symbol that has no exact symbol match in the loaded index traversal view, when the query completes successfully under the shared runtime contract, then the query-specific JSON payload includes `symbol` equal to the caller-provided value and `references` as an empty collection.
- AC-001-2: Given the caller invokes `references --symbol` with a syntactically valid full SCIP symbol that exists in the loaded index but has no command-specific reference results, when the query completes successfully under the shared runtime contract, then the query-specific JSON payload includes `symbol` equal to the caller-provided value and `references` as an empty collection.
- AC-001-3: Given an exact-symbol reference query has no matches, when the caller observes the successful payload, then the payload does not include reference entries derived from partial-name, fuzzy, regex, semantic, case-insensitive, or containing-substring matches.
- AC-001-4: Given an exact-symbol reference query has no matches, when the caller parses the successful stdout JSON defined by the shared runtime contract, then the no-match outcome is represented as structured result data rather than a shared runtime diagnostic.

### Depends on:
Run time coupling:
- I-001-001 - Reference query contract
- I-003-001 - CLI process contract

### Out of Scope
- Exact diagnostic text, failure status, stderr placement, missing `--symbol`, missing `--index`, unreadable index, malformed index, or unsupported-command behavior.
- Defining non-empty reference entry fields, reference occurrence role selection, related reference expansion, source path/range fields, result ordering, or golden fixtures.

### Assumptions
- **ASM-001-1**: The same empty `references` collection shape applies whether the queried exact symbol is absent or present without reference results. - *Why*: The parent epic treats both as successful no-match outcomes and does not require callers to distinguish the cause at this capability layer. - Confidence: MEDIUM

### Open Questions
- None.

---

## Story ST-002 - Return an Empty implementations Collection for Missing Exact Symbols

### References
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#capability-cap-001---resolve-exact-symbol-query-inputs - Requires exact symbols absent from the traversal view to return successful no-match results.
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#general-information - Defines empty successful result collections for no matching symbol or related results.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/stderr-exit-error-contract.md - Owns shared runtime failures, which this story must not redefine.

### User Story
**As an** Automation Agent running an implementation query against a selected index, **I want to** receive a successful empty `implementations` payload when my exact full symbol has no implementation matches, **so that** automation can treat absence as data instead of a command failure.

### Acceptance Criteria
- AC-002-1: Given the caller invokes `implementations --symbol` with a syntactically valid full SCIP symbol that has no exact symbol match in the loaded index traversal view, when the query completes successfully under the shared runtime contract, then the query-specific JSON payload includes `symbol` equal to the caller-provided value and `implementations` as an empty collection.
- AC-002-2: Given the caller invokes `implementations --symbol` with a syntactically valid full SCIP symbol that exists in the loaded index but has no command-specific implementation results, when the query completes successfully under the shared runtime contract, then the query-specific JSON payload includes `symbol` equal to the caller-provided value and `implementations` as an empty collection.
- AC-002-3: Given an exact-symbol implementation query has no matches, when the caller observes the successful payload, then the payload does not include implementation entries derived from partial-name, fuzzy, regex, semantic, case-insensitive, or containing-substring matches.
- AC-002-4: Given an exact-symbol implementation query has no matches, when the caller parses the successful stdout JSON defined by the shared runtime contract, then the no-match outcome is represented as structured result data rather than a shared runtime diagnostic.

### Depends on:
Run time coupling:
- I-001-002 - Implementation query contract
- I-003-001 - CLI process contract

### Out of Scope
- Exact diagnostic text, failure status, stderr placement, missing `--symbol`, missing `--index`, unreadable index, malformed index, or unsupported-command behavior.
- Defining non-empty implementation entry fields, relationship traversal rules, implementation definition location resolution, source path/range fields, result ordering, or golden fixtures.

### Assumptions
- **ASM-002-1**: The same empty `implementations` collection shape applies whether the queried exact symbol is absent or present without implementation results. - *Why*: The parent epic treats both as successful no-match outcomes and does not require callers to distinguish the cause at this capability layer. - Confidence: MEDIUM

### Open Questions
- None.
