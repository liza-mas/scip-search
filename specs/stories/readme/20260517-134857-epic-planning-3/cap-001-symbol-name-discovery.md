# User Stories: Symbol Name Discovery

Status: draft

## Goal
`scip-search symbols --name <name>` returns deterministic successful symbol-discovery results for literal partial-name queries, grouped by package identity to minimize repeated package text while preserving enough context to reconstruct exact full SCIP symbol strings for later symbol-based commands.

## Parent Epic
`specs/epics/readme/20260517-134857-epic-planning-3.md` - Capability CAP-001, "Resolve partial symbol names"

## Context
Automation callers often know a source-level symbol name but not the full SCIP symbol identifier. This story set defines the query-specific behavior for discovering candidate SCIP symbols from `symbols --name`; shared command invocation, index loading, stdout/stderr, exit status, and raw SCIP traversal contracts remain owned by sibling epics. The default payload is compact for coding-agent token efficiency; `--flat` preserves the earlier self-contained symbol-entry shape.

## Personas
- **Automation Agent**: an AI or script-driven caller running `scip-search` in a terminal or sandbox, needing deterministic JSON results and exact full SCIP symbol strings for follow-up commands.
- **CLI Maintainer**: a Go developer maintaining `scip-search`, needing bounded query semantics and fixtures that keep symbol discovery behavior stable across SCIP indexes.

## General information

Applies to: all symbol name discovery stories in this document.

### References
- goal spec: `README.md#scip-symbol-format` - Defines SCIP symbol structure and requires partial name queries to return full SCIP symbols alongside results.
- goal spec: `README.md#what-is-scip-search` - Documents `scip-search symbols --index <index-path> --name <name>` and successful structured JSON output.
- goal spec: `README.md#complementary-existing-tool` - Identifies `symbols` as the query used to find where a named symbol is defined.
- parent epic: `specs/epics/readme/20260517-134857-epic-planning-3.md#general-information` - Provides shared personas, non-functional requirements, boundaries, and assumptions for symbol and package discovery.
- parent epic: `specs/epics/readme/20260517-134857-epic-planning-3.md#capability-cap-001---resolve-partial-symbol-names` - Defines partial symbol-name matching, multi-result behavior, returned full symbol strings, package identity, and definition context.
- consistency scan: `specs/stories/readme/20260517-134857-epic-planning-3/` - No existing same-domain story documents were present before this file was written.

### Non-Functional Requirements
- NFR-000-1: Default result entries preserve the exact SCIP symbol descriptor and package key needed to reconstruct the full SCIP symbol string from the index, with no normalization or rewriting.
- NFR-000-2: Successful result ordering is deterministic for the same index and query, independent of traversal iteration order.
- NFR-000-3: No-match and multi-match outcomes are successful discovery results, not query-specific runtime failures.
- NFR-000-4: This document does not redefine shared runtime behavior owned by `epic-planning-1` or raw SCIP traversal behavior owned by `epic-planning-2`.

### Related External Components
- Component C-001 - SCIP symbol identifiers: Human-readable symbol strings made of scheme, package manager, package name, package version, and descriptors.
- Component C-002 - SCIP traversal view: The downstream query input from the traversal epic, including symbol inventories, package identity, document paths, and definition locations.
- Component C-003 - Symbol query fixture set: Deterministic SCIP test data and expected JSON cases used by maintainers to validate symbol discovery queries.

### Interfaces
- I-001-001 - Symbol discovery query contract (Interface 001 of Component C-001): The `symbols --name` query accepts a partial name and returns a successful JSON payload with a `packages` collection whose entries contain package identity fields and nested matched symbol descriptors. The optional `--flat` flag returns the compatibility `symbols` collection of matched full SCIP symbol entries.

### Out of Scope
- `references --symbol`, `implementations --symbol`, exact-symbol lookup, reference occurrence selection, implementation relationship traversal, and relationship expansion from a matched symbol.
- Fuzzy search, regex search, semantic similarity, ranking by relevance, cross-index matching, and reading source files to supplement missing SCIP data.
- Shared invocation failures, index-loading failures, stdout/stderr stream rules, process exit behavior, and shared runtime error payloads.
- SCIP protobuf parsing, traversal view construction, source range extraction, hover extraction, and raw occurrence lookup construction.
- Package discovery behavior for the `packages` command.

### Assumptions
- **ASM-000-1**: Partial-name matching is literal and case-sensitive - *Why*: SCIP symbol identifiers preserve source spelling and the parent epic excludes fuzzy, regex, semantic, and other non-literal search modes - Confidence: MEDIUM.
- **ASM-000-2**: Stable ordering is by the exact full SCIP symbol string in ascending lexical order - *Why*: the parent epic requires deterministic ordering by observable result values and does not require preserving SCIP traversal order - Confidence: MEDIUM.
- **ASM-000-3**: A missing definition location is represented in the successful symbol entry as absent context rather than as an error - *Why*: the parent epic says definition location is included when traversal provides one, while successful discovery is still valuable when the full symbol string is known - Confidence: HIGH.

### Open Questions
- None.

---

## Story ST-001 - Match Partial Symbol Names Deterministically

### References
- goal spec: `README.md#scip-symbol-format` - Requires partial name queries such as `--name Supervisor` to resolve to full SCIP symbols.
- parent epic: `specs/epics/readme/20260517-134857-epic-planning-3.md#capability-cap-001---resolve-partial-symbol-names` - Defines literal partial-name matching, all-match behavior, successful ambiguity, empty results, and stable ordering.

### User Story
**As an** Automation Agent, **I want to** query `symbols --name <name>` with a partial source-level symbol name, **so that** I can deterministically discover every matching full SCIP symbol candidate without choosing a candidate blindly.

### Acceptance Criteria
- AC-001-1: Given a loaded index with one or more symbols whose descriptor or display name contains the supplied `--name` text, when the Automation Agent runs `symbols --name <name>`, then the successful response includes every matching symbol candidate nested under the matching package identity in the `packages` collection.
- AC-001-2: Given a loaded index with multiple symbols that match the same supplied `--name` text, when the Automation Agent runs the query, then the response is a successful multi-result response and does not fail only because the partial name is ambiguous.
- AC-001-3: Given a loaded index with no symbols whose descriptor or display name contains the supplied `--name` text, when the Automation Agent runs the query, then the successful response contains an empty `packages` collection.
- AC-001-4: Given a loaded index and query that produce multiple matches, when the Automation Agent repeats the same query against the same index, then package entries and nested symbol entries appear in the same order every time.
- AC-001-5: Given a supplied `--name` value that contains characters commonly used by pattern syntaxes, when the Automation Agent runs the query, then those characters are matched literally and are not interpreted as regex, fuzzy, semantic, or glob syntax.

### Depends on:
Run time coupling:
- Interface I-001-001 - Symbol discovery query contract

### Out of Scope
- Ranking candidates by relevance or selecting a single best match.
- Case-insensitive matching, fuzzy matching, regex matching, semantic matching, and cross-index matching.
- Shared command-line argument validation and index-loading failures.

### Assumptions
- **ASM-001-1**: Descriptor and display-name matching are both acceptable match sources for the same result - *Why*: the parent epic says the query matches symbol name or descriptor text and returns the matched display or descriptor name - Confidence: HIGH.

### Open Questions
- None.

---

## Story ST-002 - Return Reconstructable SCIP Symbols With Match Context

### References
- goal spec: `README.md#scip-symbol-format` - Requires full SCIP symbol strings so callers can pass them to later `references` or `implementations` commands.
- parent epic: `specs/epics/readme/20260517-134857-epic-planning-3.md#capability-cap-001---resolve-partial-symbol-names` - Requires each result to expose the exact full `symbol` string, matched display or descriptor name, package identity, and definition location when traversal provides one.

### User Story
**As an** Automation Agent, **I want to** receive package-grouped symbol descriptors and matching context in each symbol discovery result, **so that** I can choose the correct candidate and reconstruct its exact symbol string for later symbol-based commands without spending tokens on repeated package identity fields.

### Acceptance Criteria
- AC-002-1: Given a matching SCIP symbol in the loaded index, when the Automation Agent runs `symbols --name <name>`, then the containing package entry includes `packageKey`, `scheme`, `packageManager`, `packageName`, and `packageVersion`, and the nested symbol entry includes the symbol `descriptor` exactly as it appears after the package identity prefix.
- AC-002-2: Given a matching SCIP symbol in the loaded index, when the Automation Agent reads a returned nested symbol entry, then the entry identifies the matched symbol name or descriptor text that caused the result to match the supplied `--name` value.
- AC-002-3: Given a nested symbol entry and its containing package entry, when the Automation Agent concatenates `packageKey`, one space, and `descriptor`, then the result is the exact full SCIP symbol string from the index.
- AC-002-4: Given traversal provides a definition location for a matching symbol, when the Automation Agent reads the returned nested symbol entry, then the entry includes the definition document path and source range for that symbol.
- AC-002-4b: Given traversal does not provide a definition location for a matching symbol, when the Automation Agent reads the returned nested symbol entry, then the entry still includes the exact descriptor and available match context without turning the successful match into an error.
- AC-002-5: Given a caller needs the earlier self-contained result shape, when the Automation Agent runs `symbols --name <name> --flat`, then each returned `symbols` entry includes a `symbol` field whose value is the exact full SCIP symbol string from the index and includes package identity fields on that same entry.

### Depends on:
Run time coupling:
- Interface I-001-001 - Symbol discovery query contract

Implementation ordering:
- Story ST-001 - Match Partial Symbol Names Deterministically

### Out of Scope
- Defining the response shape for `references --symbol` or `implementations --symbol`.
- Fetching hover documentation or reading source files to enrich match context.
- Defining shared stdout, stderr, or process status behavior.

### Assumptions
- **ASM-002-1**: Package identity in symbol results comes from the full SCIP symbol identifier, not from package discovery output - *Why*: callers need symbol-local context, and package listing is a sibling capability - Confidence: HIGH.

### Open Questions
- None.

---

## Story ST-003 - Validate Symbol Query Fixtures

### References
- parent epic: `specs/epics/readme/20260517-134857-epic-planning-3.md#capability-cap-001---resolve-partial-symbol-names` - Requires deterministic matching, multi-result behavior, empty results, returned full symbols, and definition context for symbol discovery.
- parent epic: `specs/epics/readme/20260517-134857-epic-planning-3.md#general-information` - Requires deterministic, small query-specific fixtures that use the same SCIP loading and traversal path as commands.

### User Story
**As a** CLI Maintainer, **I want to** validate `symbols --name` against deterministic symbol-query fixtures and golden JSON cases, **so that** future changes preserve literal matching, successful ambiguity, empty results, stable ordering, compact grouping, and exact full SCIP symbol reconstruction.

### Acceptance Criteria
- AC-003-1: Given the symbol discovery fixture set, when maintainers run the symbol query validation cases, then at least one golden successful response proves a partial literal name can match multiple full SCIP symbols.
- AC-003-2: Given the symbol discovery fixture set, when maintainers run the symbol query validation cases, then at least one golden successful response proves an unmatched partial name returns an empty `packages` collection.
- AC-003-3: Given the symbol discovery fixture set, when maintainers run the symbol query validation cases, then the golden responses assert deterministic ordering by observable package keys and reconstructed full symbol strings.
- AC-003-4: Given the symbol discovery fixture set, when maintainers run the symbol query validation cases, then each non-empty golden nested symbol entry asserts the exact SCIP descriptor and match context fields defined by Story ST-002, and the containing package entry asserts the package identity fields needed to reconstruct the full symbol.
- AC-003-5: Given the symbol discovery fixture set, when maintainers run the symbol query validation cases, then the cases do not assert shared runtime failures, package query behavior, reference query behavior, implementation query behavior, fuzzy search, regex search, semantic similarity, or cross-index behavior.

### Depends on:
Implementation ordering:
- Story ST-001 - Match Partial Symbol Names Deterministically
- Story ST-002 - Return Full SCIP Symbols With Match Context

### Out of Scope
- Package query fixture expectations.
- Reference and implementation query fixture expectations.
- Constructing raw SCIP traversal fixtures outside the traversal epic's shared fixture path.

### Assumptions
- **ASM-003-1**: Symbol-query fixture expectations can live with this capability while shared fixture construction remains outside this document - *Why*: this task scope includes symbol-query fixture expectations but excludes raw SCIP traversal construction - Confidence: HIGH.

### Open Questions
- None.
