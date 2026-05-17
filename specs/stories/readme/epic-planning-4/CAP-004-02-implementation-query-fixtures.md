# User Stories: Implementation Query Fixtures

Status: review

## Goal
Maintainers can validate `scip-search implementations --symbol <scip-symbol>` against deterministic SCIP fixtures and golden JSON cases that prove incoming implementation relationships, available and unavailable definition locations, source paths, ranges, empty results, and stable ordering.

## Parent Epic
specs/epics/readme/20260517-141006-epic-planning-4.md - Capability CAP-004

## Context
Implementation query behavior is defined by CAP-001 and CAP-003. This document defines only the query-specific fixture and golden-case expectations that keep that behavior stable. Shared runtime fixtures, raw SCIP traversal fixture construction, reference fixtures, and discovery-query fixtures remain owned by sibling stories and epics.

## Personas
- **CLI Maintainer**: A Go developer maintaining `scip-search`, needing bounded implementation-query fixtures that catch regressions in relationship selection, location handling, payload fields, and ordering.
- **Automation Agent**: An AI or script-driven caller parsing `scip-search implementations` output in a terminal or sandbox, needing deterministic implementation symbols and locations it can feed into later navigation or editing workflows.

## General information

Applies to: implementation-query fixture coverage for successful `implementations --symbol` queries.

### References
- goal spec: README.md#what-is-scip-search - Documents `scip-search implementations --index <index-path> --symbol <scip-symbol>` and structured JSON success output.
- goal spec: README.md#scip-symbol-format - Defines exact full SCIP symbols and command chaining from discovery to symbol-based commands.
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#general-information - Requires stable successful payloads, explicit empty collections, source document paths, SCIP ranges, no source-file reads for missing locations, and small deterministic query-specific fixtures.
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#capability-cap-004---validate-symbol-based-queries-with-fixtures - Requires implementation fixtures for exact symbols, missing symbols, implementation relationships, locations, ranges, stable ordering, empty results, and relationships without locations where representable.
- dependency story: specs/stories/readme/epic-planning-4/CAP-001-01-exact-symbol-input.md - Defines literal exact `implementations --symbol` input and queried-symbol preservation.
- dependency story: specs/stories/readme/epic-planning-4/CAP-001-02-missing-symbol-results.md - Defines successful empty `implementations` payloads for missing exact symbols and no-match outcomes.
- dependency story: specs/stories/readme/epic-planning-4/CAP-003-01-implementation-relationship-selection.md - Defines incoming implementation relationship selection and exclusion of outgoing or synthesized results.
- dependency story: specs/stories/readme/epic-planning-4/CAP-003-02-implementation-result-json-shape.md - Defines implementation entry fields, relationship basis, available definition locations, missing-location behavior, ordering, and empty payload behavior.
- sibling boundary: specs/epics/readme/20260517-100328-epic-planning-2.md#capability-cap-004---provide-traversal-fixtures-for-scip-data-coverage - Owns shared traversal fixture construction for documents, occurrences, symbols, relationships, ranges, and hover metadata.
- consistency check: specs/stories/readme/20260517-134857-epic-planning-3/CAP-003-03-golden-json-validation.md - Provides same-project precedent for successful query golden JSON validation while excluding implementation-query cases from discovery scope.

### Non-Functional Requirements
- NFR-000-1: Implementation fixture data must be deterministic, small enough for normal CLI validation, and independent of external indexer installation.
- NFR-000-2: Implementation golden cases must validate observable `implementations` command payloads through the same SCIP loading and traversal path used by the command.
- NFR-000-3: Implementation golden cases must preserve exact full SCIP symbols, incoming relationship basis, document-relative paths when available, SCIP 0-based ranges when available, explicit empty collections, and deterministic array ordering.
- NFR-000-4: Implementation fixture stories must not redefine shared command routing, `--index` handling, stdout/stderr stream rules, process status, raw traversal fixture construction, symbol/package discovery behavior, reference-query behavior, source-file reads, synthesized relationships, or performance coverage.

### Related External Components
- Component C-001 - SCIP symbol identifiers: Exact full symbol strings supplied through `--symbol` and returned as implementation result symbols.
- Component C-002 - SCIP occurrence data: Definition occurrences with source document paths and SCIP ranges used to locate implementation symbols when available.
- Component C-003 - SCIP implementation relationships: Relationship edges marked as implementation relationships and used to select incoming implementation results.
- Component C-004 - Query fixture set: Deterministic SCIP test data and expected JSON cases used by maintainers to validate implementation behavior.

### Interfaces
- I-001-002 - Implementation query contract (Interface 002 of Component C-001): The `implementations --symbol` query returns a top-level `symbol` value and an `implementations` collection for the queried exact symbol.
- I-002-001 - Implementation location source (Interface 001 of Component C-002): Traversal data exposes definition document paths and SCIP ranges for implementation symbols when those occurrences are available.
- I-003-001 - Implementation relationship source (Interface 001 of Component C-003): Traversal data exposes incoming implementation relationships used as the basis for implementation entries.

### Out of Scope
- Shared runtime fixtures for command routing, missing flags, invalid invocations, index-path validation, unreadable indexes, malformed indexes, stdout/stderr failures, and process status.
- Shared traversal fixture construction for raw SCIP documents, occurrences, relationships, ranges, hover metadata, and reusable lookup views.
- Symbol discovery fixtures, package discovery fixtures, reference query fixtures, large real-world fixtures, performance benchmarks, external indexer installation, and ctags fallback fixtures.
- Synthesizing implementation relationships, source-file reads to infer locations, full hierarchy graphs, package dependency graphs, semantic similarity, graph output, alternate output formats, and human-readable tables.

### Assumptions
- **ASM-000-1**: The implementation query fixture may extend the shared traversal fixture set with query-specific symbols, definition occurrences, and implementation relationships rather than defining a separate raw fixture format. - *Why*: The parent epic says query-specific fixtures can extend shared traversal coverage, while epic-planning-2 owns reusable traversal fixture construction. - Confidence: HIGH
- **ASM-000-2**: Implementation golden JSON validation compares parsed JSON values and implementation array order, not object member order. - *Why*: The successful payload contract is machine-readable JSON content with stable result ordering; object field order is not a reliable JSON semantic. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Cover Implementation Relationship Fixture Cases

### References
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#capability-cap-004---validate-symbol-based-queries-with-fixtures - Requires exact symbol, missing symbol, implementation relationship, available location, unavailable location where representable, ranges, ordering, and empty result fixture coverage.
- dependency story: specs/stories/readme/epic-planning-4/CAP-003-01-implementation-relationship-selection.md - Defines incoming implementation relationship selection and exclusion of outgoing or synthesized results.
- dependency story: specs/stories/readme/epic-planning-4/CAP-003-02-implementation-result-json-shape.md - Defines location and missing-location behavior for implementation entries.
- dependency story: specs/stories/readme/epic-planning-4/CAP-001-02-missing-symbol-results.md - Defines successful empty `implementations` collections for missing and no-match symbols.

### User Story
**As a** CLI Maintainer, **I want to** validate `implementations --symbol` against a deterministic fixture containing incoming implementation relationships with and without available definition locations, **so that** future changes cannot silently alter implementation relationship selection or location handling.

### Acceptance Criteria
- AC-001-1: Given the implementation fixture contains the exact queried target symbol `scip-go gomod github.com/liza-mas/scip-search . query/Handler#`, when maintainers run the fixture-backed implementation query cases, then the fixture provides deterministic data for a found exact symbol.
- AC-001-2: Given the fixture contains an implementer symbol such as `scip-go gomod github.com/liza-mas/scip-search . query/httpHandler#` with an incoming SCIP implementation relationship pointing at the queried target symbol, when maintainers run `implementations --symbol` for the target, then the fixture provides a deterministic implementation result.
- AC-001-3: Given the implementation fixture contains a definition occurrence for that implementer symbol in a source document such as `internal/query/http_handler.go`, when maintainers run the implementation query case, then the fixture provides deterministic definition `documentPath` and SCIP `range` data for the implementation entry.
- AC-001-4: Given the shared traversal fixture can represent a relationship-only implementation symbol without an available definition occurrence, when the fixture includes such an implementer, then the implementation query case proves the relationship still produces a successful implementation entry without location fields.
- AC-001-4b: Given the shared traversal fixture cannot represent a relationship-only implementation symbol without a definition occurrence, when maintainers review the fixture coverage, then that unavailable-location case is recorded as not representable by the shared fixture layer rather than replaced with guessed or source-file-derived locations.
- AC-001-5: Given the fixture contains an outgoing implementation relationship from the queried symbol to a different target, when maintainers run the implementation query for the queried symbol, then that outgoing target does not produce an implementation result.
- AC-001-6: Given the fixture contains symbols with similar names, package proximity, reference relationships, or occurrences but no incoming implementation relationship pointing at the queried symbol, when maintainers run the implementation query case, then those records do not produce implementation entries.
- AC-001-7: Given the fixture contains no exact symbol or incoming implementation relationship for `scip-go gomod github.com/liza-mas/scip-search . query/MissingHandler#`, when maintainers run the missing-symbol implementation case, then the expected successful result is an explicit empty `implementations` collection.
- AC-001-8: Given the fixture contains a symbol with no incoming implementation relationships, when maintainers run its implementation query case, then the expected successful result is an explicit empty `implementations` collection.

### Depends on:
Implementation ordering:
- Story document CAP-001-01-exact-symbol-input.md - Implementation fixture cases start from exact full SCIP symbols.
- Story document CAP-001-02-missing-symbol-results.md - Missing and no-match implementation payload behavior must be defined before fixture cases can validate it.
- Story document CAP-003-01-implementation-relationship-selection.md - Incoming relationship selection must be defined before fixture cases can validate it.
- Story document CAP-003-02-implementation-result-json-shape.md - Available and missing location payload behavior must be defined before fixture cases can validate it.

Run time coupling:
- I-001-002 - Implementation query contract
- I-002-001 - Implementation location source
- I-003-001 - Implementation relationship source

### Out of Scope
- Building the shared traversal fixture serializer, raw SCIP helper API, or reusable occurrence and relationship lookup views.
- Validating final golden JSON object field names beyond the implementation concepts; golden payload assertions are covered by ST-002.
- Validating reference relationships, package discovery, symbol discovery, or shared runtime failures.

### Assumptions
- **ASM-001-1**: Artificial but valid SCIP symbols are acceptable for the implementation fixture when they preserve SCIP symbol structure and command-observable behavior. - *Why*: The parent epic excludes external indexer installation and requires deterministic fixtures rather than generated real-world indexes. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-002 - Validate Implementation Golden JSON Cases

### References
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#capability-cap-004---validate-symbol-based-queries-with-fixtures - Requires golden JSON cases for implementation relationships, locations, missing locations where representable, ranges, stable ordering, and empty results.
- dependency story: specs/stories/readme/epic-planning-4/CAP-003-02-implementation-result-json-shape.md - Defines `symbol`, `implementations`, implementation entry fields, relationship basis, optional location fields, ordering, and empty payload behavior.
- story: Story ST-001 - Defines the implementation fixture cases that golden JSON must validate.

### User Story
**As a** CLI Maintainer, **I want to** validate successful `implementations --symbol` responses against golden JSON cases, **so that** implementation symbols, relationship basis, available paths, ranges, absent locations, empty results, and ordering stay stable for automation.

### Acceptance Criteria
- AC-002-1: Given the implementation fixture case for `scip-go gomod github.com/liza-mas/scip-search . query/Handler#`, when maintainers run golden validation, then the expected JSON contains top-level `symbol` equal to that exact query value and an `implementations` collection containing each qualifying incoming implementation result.
- AC-002-2: Given an implementation entry with an available definition location appears in the golden JSON, when maintainers inspect it, then the entry asserts `implementationSymbol`, incoming `relationship` basis, traversal-provided `documentPath`, and fixture-defined SCIP `range` values required by the implementation result contract.
- AC-002-3: Given a qualifying implementation symbol has no available definition location and the shared fixture can represent that case, when maintainers inspect its golden JSON entry, then the entry includes `implementationSymbol` and incoming `relationship` basis without guessed `documentPath`, guessed `range`, empty-string location, or sentinel location values.
- AC-002-4: Given the fixture includes an outgoing implementation relationship from the queried symbol to another target, when maintainers inspect the golden JSON for the queried symbol, then no `implementations` entry is present for that outgoing target.
- AC-002-5: Given the fixture produces multiple implementation entries, when maintainers compare the golden JSON, then the `implementations` array order is asserted as stable by `implementationSymbol` and, for equal implementation symbols, by available definition location.
- AC-002-6: Given the missing exact symbol implementation case is validated, when maintainers inspect the expected JSON, then it contains top-level `symbol` equal to the missing query value and `implementations` as an explicit empty collection.
- AC-002-7: Given the no-implementation successful case for a symbol with no incoming implementation relationships is validated, when maintainers inspect the expected JSON, then it contains the queried `symbol` and an explicit empty `implementations` collection.
- AC-002-8: Given implementation golden validation runs, when maintainers inspect expected cases, then they assert only successful query-specific `implementations` payload behavior and do not assert shared runtime failure payloads, stderr text, reference results, symbol discovery results, or package query payloads.

### Depends on:
Implementation ordering:
- Story ST-001 - Implementation fixture cases must be defined before golden JSON can assert their outputs.
- Story document CAP-003-02-implementation-result-json-shape.md - Implementation JSON fields, optional locations, and ordering must be defined before golden JSON can validate them.

Run time coupling:
- I-001-002 - Implementation query contract
- I-002-001 - Implementation location source
- I-003-001 - Implementation relationship source

### Out of Scope
- Object member order as a JSON semantic assertion.
- Golden cases for runtime failures, malformed indexes, invalid invocations, references, packages, symbols, or raw traversal data.
- Adding hover text, source-file content, reference occurrences, hierarchy graphs, ranking scores, package data, or graph edges to expected implementation entries.

### Assumptions
- **ASM-002-1**: Implementation golden cases can share the same physical deterministic SCIP fixture source as reference or discovery fixtures if the expected cases remain separated by command. - *Why*: The parent epic requires query-specific coverage but does not require one fixture file per command. - Confidence: HIGH

### Open Questions
- None.
