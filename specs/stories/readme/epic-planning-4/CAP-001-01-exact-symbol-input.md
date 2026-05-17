# User Stories: Exact Symbol Input

Status: review

## Goal
`references --symbol` and `implementations --symbol` accept caller-provided full SCIP symbols as literal exact identifiers and echo the queried symbol in successful payloads.

## Parent Epic
specs/epics/readme/20260517-141006-epic-planning-4.md - Capability CAP-001

## Context
Symbol-based queries are follow-up commands after discovery has returned a full SCIP symbol string. This document defines the common input contract for both `references --symbol` and `implementations --symbol`; command-specific occurrence selection and relationship traversal are covered by sibling story documents.

## Personas
- **Automation Agent**: An AI or script-driven caller running `scip-search` in a terminal or sandbox, needing exact full-symbol queries and stable JSON it can feed into later code-navigation workflows.
- **CLI Maintainer**: A Go developer maintaining `scip-search`, needing bounded symbol-query semantics that do not drift into partial discovery behavior.

## General information

Applies to: successful exact-symbol input handling for `references --symbol` and `implementations --symbol`.

### References
- goal spec: README.md#scip-symbol-format - Defines full SCIP symbol strings and says discovery results can feed later `references` or `implementations` calls.
- goal spec: README.md#what-is-scip-search - Documents `scip-search references --index <index-path> --symbol <scip-symbol>` and `scip-search implementations --index <index-path> --symbol <scip-symbol>`.
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#capability-cap-001---resolve-exact-symbol-query-inputs - Requires exact full-symbol input, no partial-name matching, and queried-symbol preservation.
- sibling boundary: specs/epics/readme/20260517-134857-epic-planning-3.md#capability-cap-001---resolve-partial-symbol-names - Owns `symbols --name` partial discovery before callers choose an exact symbol.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/stdout-json-success-contract.md - Owns shared JSON stdout success stream behavior while leaving query-specific fields to query stories.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/stderr-exit-error-contract.md - Owns shared runtime failures, which this document does not redefine.
- consistency check: specs/stories/readme/20260517-134857-epic-planning-3/CAP-002-03-package-result-json-shape.md - Provides same-domain precedent for query-specific JSON payload fields under the shared success contract.

### Non-Functional Requirements
- NFR-000-1: Exact-symbol input behavior must preserve full SCIP symbol strings without normalization, case folding, trimming internal whitespace, descriptor rewriting, or package component rewriting.
- NFR-000-2: Successful payloads must include the queried full symbol exactly as supplied so automation can correlate a result with the command invocation.
- NFR-000-3: These stories must not redefine command routing, `--index` loading, stdout/stderr stream rules, process status, shared runtime failures, reference occurrence selection, implementation traversal, or fixture coverage.

### Related External Components
- Component C-001 - SCIP symbol identifiers: Exact full symbol strings supplied through `--symbol` and returned in successful query payloads.
- Component C-003 - Calling process environment: The shell, script, or agent invoking `scip-search` and parsing successful JSON stdout under the shared runtime contract.

### Interfaces
- I-001-001 - Reference query contract (Interface 001 of Component C-001): The `references --symbol` query accepts an exact full SCIP symbol and returns reference-query results for that exact symbol.
- I-001-002 - Implementation query contract (Interface 002 of Component C-001): The `implementations --symbol` query accepts an exact full SCIP symbol and returns implementation-query results for that exact symbol.
- I-003-001 - CLI process contract (Interface 001 of Component C-003): The shared process contract that exposes successful query payloads as parseable JSON on stdout.

### Out of Scope
- Partial `symbols --name` matching, ambiguous partial-name resolution, fuzzy matching, regex matching, semantic similarity, ranking by relevance, or cross-index lookup.
- Shared missing-flag, missing-value, unsupported-command, unreadable-index, malformed-index, stderr, stdout, and exit-status behavior.
- Choosing reference occurrences, including related reference symbols, traversing implementation relationships, resolving implementation definition locations, or defining fixture/golden coverage.
- Alternate output formats, pretty printing, source-file reads, daemon/watch behavior, MCP behavior, UI, embeddings, vector search, or ctags fallback behavior.

### Assumptions
- **ASM-000-1**: Exact `--symbol` comparison is literal and case-sensitive. - *Why*: SCIP symbol identifiers are exact strings that callers can pass unchanged from discovery results. - Confidence: HIGH
- **ASM-000-2**: Both symbol-based command payloads expose the queried symbol through a top-level `symbol` field. - *Why*: The capability requires preserving the queried full symbol in successful payloads, and a shared field lets automation correlate results across both commands without inspecting command-specific collections. - Confidence: MEDIUM

### Open Questions
- None.

---

## Story ST-001 - Treat references --symbol as an Exact Symbol Query

### References
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#capability-cap-001---resolve-exact-symbol-query-inputs - Requires `references --symbol` to interpret `--symbol` as an exact full SCIP symbol.
- goal spec: README.md#what-is-scip-search - Documents the `references --symbol <scip-symbol>` command form.
- sibling boundary: specs/epics/readme/20260517-134857-epic-planning-3.md#capability-cap-001---resolve-partial-symbol-names - Owns partial symbol discovery outside this command.

### User Story
**As an** Automation Agent running a follow-up reference query, **I want to** pass the full SCIP symbol returned by discovery to `references --symbol` unchanged, **so that** I can ask for references to one exact symbol without triggering partial-name resolution.

### Acceptance Criteria
- AC-001-1: Given the caller invokes `references --symbol` with a full SCIP symbol string, when the reference query runs after shared runtime validation, then the query treats the provided `--symbol` value as one literal exact symbol identifier.
- AC-001-2: Given a successful `references --symbol` query, when the caller parses the query-specific JSON payload, then the payload includes a top-level `symbol` value equal to the exact `--symbol` argument supplied by the caller.
- AC-001-3: Given the supplied `--symbol` value is a substring, simple name, descriptor fragment, package fragment, or differently cased variant of an indexed symbol, when `references --symbol` runs, then it does not expand that value through partial-name, fuzzy, regex, semantic, or case-insensitive matching.
- AC-001-4: Given multiple indexed symbols contain the same substring as the supplied `--symbol` value, when the value is not an exact full symbol match, then `references --symbol` does not return references for those containing symbols.

### Depends on:
Run time coupling:
- I-001-001 - Reference query contract
- I-003-001 - CLI process contract

### Out of Scope
- Defining which occurrence roles count as reference results.
- Including related reference symbols or reference-related SCIP relationships.
- Defining source path, range, hover, or ordering fields for reference results.
- Defining shared runtime failures for missing `--symbol`, malformed invocation shape, invalid index input, stdout, stderr, or process status.

### Assumptions
- **ASM-001-1**: The command does not reject a syntactically valid full symbol solely because the loaded index has no exact match for it. - *Why*: CAP-001 defines missing exact symbols as successful no-match results, which are specified in the sibling missing-symbol story. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-002 - Treat implementations --symbol as an Exact Symbol Query

### References
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#capability-cap-001---resolve-exact-symbol-query-inputs - Requires `implementations --symbol` to interpret `--symbol` as an exact full SCIP symbol.
- goal spec: README.md#what-is-scip-search - Documents the `implementations --symbol <scip-symbol>` command form.
- sibling boundary: specs/epics/readme/20260517-134857-epic-planning-3.md#capability-cap-001---resolve-partial-symbol-names - Owns partial symbol discovery outside this command.

### User Story
**As an** Automation Agent running a follow-up implementation query, **I want to** pass the full SCIP symbol returned by discovery to `implementations --symbol` unchanged, **so that** I can ask for implementations of one exact symbol without triggering partial-name resolution.

### Acceptance Criteria
- AC-002-1: Given the caller invokes `implementations --symbol` with a full SCIP symbol string, when the implementation query runs after shared runtime validation, then the query treats the provided `--symbol` value as one literal exact symbol identifier.
- AC-002-2: Given a successful `implementations --symbol` query, when the caller parses the query-specific JSON payload, then the payload includes a top-level `symbol` value equal to the exact `--symbol` argument supplied by the caller.
- AC-002-3: Given the supplied `--symbol` value is a substring, simple name, descriptor fragment, package fragment, or differently cased variant of an indexed symbol, when `implementations --symbol` runs, then it does not expand that value through partial-name, fuzzy, regex, semantic, or case-insensitive matching.
- AC-002-4: Given multiple indexed symbols contain the same substring as the supplied `--symbol` value, when the value is not an exact full symbol match, then `implementations --symbol` does not return implementations for those containing symbols.

### Depends on:
Run time coupling:
- I-001-002 - Implementation query contract
- I-003-001 - CLI process contract

### Out of Scope
- Defining which SCIP relationship edges count as implementation results.
- Resolving implementation symbols to definition locations.
- Defining source path, range, hover, or ordering fields for implementation results.
- Defining shared runtime failures for missing `--symbol`, malformed invocation shape, invalid index input, stdout, stderr, or process status.

### Assumptions
- **ASM-002-1**: The command does not reject a syntactically valid full symbol solely because the loaded index has no exact match for it. - *Why*: CAP-001 defines missing exact symbols as successful no-match results, which are specified in the sibling missing-symbol story. - Confidence: HIGH

### Open Questions
- None.
