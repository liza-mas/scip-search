# User Stories: Reference Query Fixtures

Status: review

## Goal
Maintainers can validate `scip-search references --symbol <scip-symbol>` against deterministic SCIP fixtures and golden JSON cases that prove exact-symbol reference behavior, reference-related symbol expansion, source locations, ranges, empty results, and stable ordering.

## Parent Epic
specs/epics/readme/20260517-141006-epic-planning-4.md - Capability CAP-004

## Context
Reference query behavior is defined by CAP-001 and CAP-002. This document defines only the query-specific fixture and golden-case expectations that keep that behavior stable. Shared runtime fixtures, raw SCIP traversal fixture construction, and discovery-query fixtures remain owned by sibling stories and epics.

## Personas
- **CLI Maintainer**: A Go developer maintaining `scip-search`, needing bounded reference-query fixtures that catch regressions in exact symbol handling, occurrence selection, payload fields, and ordering.
- **Automation Agent**: An AI or script-driven caller parsing `scip-search references` output in a terminal or sandbox, needing deterministic JSON locations it can feed into later navigation or editing workflows.

## General information

Applies to: reference-query fixture coverage for successful `references --symbol` queries.

### References
- goal spec: README.md#what-is-scip-search - Documents `scip-search references --index <index-path> --symbol <scip-symbol>` and structured JSON success output.
- goal spec: README.md#scip-symbol-format - Defines exact full SCIP symbols and command chaining from discovery to symbol-based commands.
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#general-information - Requires stable successful payloads, explicit empty collections, source document paths, SCIP ranges, and small deterministic query-specific fixtures.
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#capability-cap-004---validate-symbol-based-queries-with-fixtures - Requires reference fixtures for exact symbols, missing symbols, reference occurrences, reference-related symbols, locations, ranges, stable ordering, and empty results.
- dependency story: specs/stories/readme/epic-planning-4/CAP-001-01-exact-symbol-input.md - Defines literal exact `references --symbol` input and queried-symbol preservation.
- dependency story: specs/stories/readme/epic-planning-4/CAP-001-02-missing-symbol-results.md - Defines successful empty `references` payloads for missing exact symbols and no-match outcomes.
- dependency story: specs/stories/readme/epic-planning-4/CAP-002-01-reference-occurrence-selection.md - Defines direct and reference-related non-definition occurrence selection.
- dependency story: specs/stories/readme/epic-planning-4/CAP-002-02-reference-result-json-shape.md - Defines reference entry fields, source paths, SCIP ranges, role context, stable ordering, and empty payload behavior.
- sibling boundary: specs/epics/readme/20260517-100328-epic-planning-2.md#capability-cap-004---provide-traversal-fixtures-for-scip-data-coverage - Owns shared traversal fixture construction for documents, occurrences, symbols, relationships, ranges, and hover metadata.
- consistency check: specs/stories/readme/20260517-134857-epic-planning-3/CAP-003-03-golden-json-validation.md - Provides same-project precedent for successful query golden JSON validation while excluding reference-query cases from discovery scope.

### Non-Functional Requirements
- NFR-000-1: Reference fixture data must be deterministic, small enough for normal CLI validation, and independent of external indexer installation.
- NFR-000-2: Reference golden cases must validate observable `references` command payloads through the same SCIP loading and traversal path used by the command.
- NFR-000-3: Reference golden cases must preserve exact full SCIP symbols, document-relative paths, SCIP 0-based ranges, role context, explicit empty collections, and deterministic array ordering.
- NFR-000-4: Reference fixture stories must not redefine shared command routing, `--index` handling, stdout/stderr stream rules, process status, raw traversal fixture construction, symbol/package discovery behavior, implementation-query behavior, or performance coverage.

### Related External Components
- Component C-001 - SCIP symbol identifiers: Exact full symbol strings supplied through `--symbol` and attached to selected reference occurrences.
- Component C-002 - SCIP occurrence data: Document-contained occurrences with source document paths, SCIP ranges, and role context.
- Component C-003 - SCIP reference relationships: Relationship edges marked as reference relationships and used to include reference-related symbols.
- Component C-004 - Query fixture set: Deterministic SCIP test data and expected JSON cases used by maintainers to validate reference behavior.

### Interfaces
- I-001-001 - Reference query contract (Interface 001 of Component C-001): The `references --symbol` query returns a top-level `symbol` value and a `references` collection for the queried exact symbol.
- I-002-001 - Selected reference occurrence contract (Interface 001 of Component C-002): Selected occurrences expose matched symbol, document path, SCIP range, and role context for JSON formatting.
- I-003-001 - Reference relationship lookup view (Interface 001 of Component C-003): Traversal exposes reference relationships that connect the queried symbol to reference-related symbols.

### Out of Scope
- Shared runtime fixtures for command routing, missing flags, invalid invocations, index-path validation, unreadable indexes, malformed indexes, stdout/stderr failures, and process status.
- Shared traversal fixture construction for raw SCIP documents, occurrences, relationships, ranges, hover metadata, and reusable lookup views.
- Symbol discovery fixtures, package discovery fixtures, implementation query fixtures, large real-world fixtures, performance benchmarks, external indexer installation, and ctags fallback fixtures.
- Source-file reads, snippet extraction, call hierarchy synthesis, semantic ranking, graph output, alternate output formats, and human-readable tables.

### Assumptions
- **ASM-000-1**: The reference query fixture may extend the shared traversal fixture set with query-specific symbols, occurrences, and reference relationships rather than defining a separate raw fixture format. - *Why*: The parent epic says query-specific fixtures can extend shared traversal coverage, while epic-planning-2 owns reusable traversal fixture construction. - Confidence: HIGH
- **ASM-000-2**: Reference golden JSON validation compares parsed JSON values and reference array order, not object member order. - *Why*: The successful payload contract is machine-readable JSON content with stable result ordering; object field order is not a reliable JSON semantic. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Cover Direct and Reference-Related Reference Fixture Cases

### References
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#capability-cap-004---validate-symbol-based-queries-with-fixtures - Requires exact symbol, missing symbol, direct reference occurrences, reference-related symbols, locations, ranges, ordering, and empty results in fixture coverage.
- dependency story: specs/stories/readme/epic-planning-4/CAP-002-01-reference-occurrence-selection.md - Defines candidate symbol expansion and non-definition occurrence selection.
- dependency story: specs/stories/readme/epic-planning-4/CAP-001-02-missing-symbol-results.md - Defines successful empty `references` collections for missing and no-match symbols.

### User Story
**As a** CLI Maintainer, **I want to** validate `references --symbol` against a deterministic fixture containing direct references, reference-related symbol references, definitions, and no-match symbols, **so that** future changes cannot silently alter reference occurrence selection.

### Acceptance Criteria
- AC-001-1: Given the reference fixture contains the exact queried symbol `scip-go gomod github.com/liza-mas/scip-search . query/Service#`, when maintainers run the fixture-backed reference query cases, then the fixture provides deterministic data for a found exact symbol.
- AC-001-2: Given the fixture contains at least one non-definition occurrence for the exact queried symbol in a source document such as `internal/query/caller.go`, when maintainers run `references --symbol` for that symbol, then the fixture provides a direct reference occurrence with a deterministic document path, SCIP range, matched symbol, and role context.
- AC-001-3: Given the fixture contains a SCIP reference relationship between the queried symbol and a related symbol such as `scip-go gomod github.com/liza-mas/scip-search . query/ServiceAlias#`, plus a non-definition occurrence of the related symbol, when maintainers run the same reference query, then the fixture provides deterministic data for reference-related occurrence inclusion.
- AC-001-4: Given the fixture contains a definition occurrence for the queried symbol, when maintainers run the reference query case, then the fixture provides data proving definition occurrences are available but excluded from reference results.
- AC-001-5: Given the fixture contains unrelated symbols, unrelated reference relationships, or occurrences in other documents, when maintainers run the reference query case, then those unrelated fixture records do not produce reference entries.
- AC-001-6: Given the fixture contains no exact symbol or reference-related symbol for `scip-go gomod github.com/liza-mas/scip-search . query/Missing#`, when maintainers run the missing-symbol reference case, then the expected successful result is an explicit empty `references` collection.
- AC-001-7: Given the fixture contains a symbol with only definition occurrences or no non-definition reference occurrences, when maintainers run its reference query case, then the expected successful result is an explicit empty `references` collection.

### Depends on:
Implementation ordering:
- Story document CAP-001-01-exact-symbol-input.md - Reference fixture cases start from exact full SCIP symbols.
- Story document CAP-001-02-missing-symbol-results.md - Missing and no-match reference payload behavior must be defined before fixture cases can validate it.
- Story document CAP-002-01-reference-occurrence-selection.md - Reference candidate and occurrence selection must be defined before fixture cases can validate it.

Run time coupling:
- I-001-001 - Reference query contract
- I-002-001 - Selected reference occurrence contract
- I-003-001 - Reference relationship lookup view

### Out of Scope
- Building the shared traversal fixture serializer, raw SCIP helper API, or reusable occurrence and relationship lookup views.
- Validating final golden JSON object field names beyond the selected occurrence concepts; golden payload assertions are covered by ST-002.
- Validating implementation relationships, package discovery, symbol discovery, or shared runtime failures.

### Assumptions
- **ASM-001-1**: Artificial but valid SCIP symbols are acceptable for the reference fixture when they preserve SCIP symbol structure and command-observable behavior. - *Why*: The parent epic excludes external indexer installation and requires deterministic fixtures rather than generated real-world indexes. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-002 - Validate Reference Golden JSON Cases

### References
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#capability-cap-004---validate-symbol-based-queries-with-fixtures - Requires golden JSON cases for reference occurrences, missing symbols, locations, ranges, stable ordering, and empty results.
- dependency story: specs/stories/readme/epic-planning-4/CAP-002-02-reference-result-json-shape.md - Defines `symbol`, `references`, reference entry fields, range semantics, ordering, and empty payload behavior.
- story: Story ST-001 - Defines the reference fixture cases that golden JSON must validate.

### User Story
**As a** CLI Maintainer, **I want to** validate successful `references --symbol` responses against golden JSON cases, **so that** exact symbols, matched symbols, paths, ranges, roles, empty results, and ordering stay stable for automation.

### Acceptance Criteria
- AC-002-1: Given the direct-and-related reference fixture case for `scip-go gomod github.com/liza-mas/scip-search . query/Service#`, when maintainers run golden validation, then the expected JSON contains top-level `symbol` equal to that exact query value and a `references` collection containing both direct and reference-related selected occurrences.
- AC-002-2: Given a selected reference entry appears in the golden JSON, when maintainers inspect it, then the entry asserts the matched occurrence `symbol`, traversal-provided `documentPath`, fixture-defined SCIP `range`, and `roles` values required by the reference result contract.
- AC-002-3: Given the fixture includes a selected occurrence for the reference-related symbol, when maintainers inspect the golden JSON, then that entry's `symbol` remains the matched related symbol rather than being rewritten to the originally queried symbol.
- AC-002-4: Given the fixture includes a definition occurrence for the queried symbol, when maintainers inspect the golden JSON for the reference case, then no `references` entry is present for that definition occurrence.
- AC-002-5: Given the fixture produces multiple reference entries, when maintainers compare the golden JSON, then the `references` array order is asserted as stable by `documentPath`, range start position, range end position, and matched `symbol`.
- AC-002-6: Given the missing exact symbol reference case is validated, when maintainers inspect the expected JSON, then it contains top-level `symbol` equal to the missing query value and `references` as an explicit empty collection.
- AC-002-7: Given the no-reference successful case for a symbol with only definition occurrences or no selected occurrences is validated, when maintainers inspect the expected JSON, then it contains the queried `symbol` and an explicit empty `references` collection.
- AC-002-8: Given reference golden validation runs, when maintainers inspect expected cases, then they assert only successful query-specific `references` payload behavior and do not assert shared runtime failure payloads, stderr text, implementation results, symbol discovery results, or package query payloads.

### Depends on:
Implementation ordering:
- Story ST-001 - Reference fixture cases must be defined before golden JSON can assert their outputs.
- Story document CAP-002-02-reference-result-json-shape.md - Reference JSON fields and ordering must be defined before golden JSON can validate them.

Run time coupling:
- I-001-001 - Reference query contract
- I-002-001 - Selected reference occurrence contract

### Out of Scope
- Object member order as a JSON semantic assertion.
- Golden cases for runtime failures, malformed indexes, invalid invocations, implementations, packages, symbols, or raw traversal data.
- Adding snippets, hover text, source-file content, caller function names, ranking scores, or graph edges to expected reference entries.

### Assumptions
- **ASM-002-1**: Reference golden cases can share the same physical deterministic SCIP fixture source as implementation or discovery fixtures if the expected cases remain separated by command. - *Why*: The parent epic requires query-specific coverage but does not require one fixture file per command. - Confidence: HIGH

### Open Questions
- None.
