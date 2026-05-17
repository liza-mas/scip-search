# User Stories: Symbol Query Fixtures

Status: review

## Goal
`scip-search` maintainers can validate `symbols --name` behavior with deterministic SCIP fixtures and symbol-specific golden cases for literal matching, ambiguous names, empty results, and stable symbol ordering.

## Parent Epic
`specs/epics/readme/20260517-134857-epic-planning-3.md` - Capability CAP-003, "Validate discovery queries with fixtures"

## Context
Symbol discovery is useful only if automation can trust the returned full SCIP symbol strings across runs. This document defines the symbol-query fixture coverage needed to prove the CAP-001 symbol matching and payload stories without redefining shared SCIP traversal fixtures or shared runtime behavior.

## Personas
- **Automation Agent**: an AI or script-driven caller running `scip-search` in a terminal or sandbox, needing deterministic JSON results and exact full SCIP symbol strings for follow-up commands.
- **CLI Maintainer**: a Go developer maintaining `scip-search`, needing bounded query semantics and fixtures that keep symbol and package behavior stable across SCIP indexes.

## General information

Applies to: symbol-query fixture and golden-case expectations for successful `symbols --name` discovery.

### References
- goal spec: `README.md#scip-symbol-format` - Provides full SCIP symbol examples and requires partial name queries to return full SCIP symbol strings alongside results.
- goal spec: `README.md#what-is-scip-search` - Documents the `scip-search symbols --index <index-path> --name <name>` command form and structured JSON stdout.
- parent epic: `specs/epics/readme/20260517-134857-epic-planning-3.md#general-information` - Requires deterministic, small query-specific fixtures that use the same SCIP loading and traversal path as commands.
- parent epic: `specs/epics/readme/20260517-134857-epic-planning-3.md#capability-cap-003---validate-discovery-queries-with-fixtures` - Requires symbol fixtures for exact-looking partial matches, substring matches, overlapping names, empty results, stable ordering, and full SCIP symbol strings.
- sibling story: `specs/stories/readme/20260517-134857-epic-planning-3/cap-001-symbol-name-discovery.md` - Defines literal partial-name matching, ambiguous successful results, empty successful results, full symbol output, and symbol result context.
- consistency scan: `specs/stories/readme/20260517-134857-epic-planning-3/` - Existing CAP-001 and CAP-002 story documents establish query-specific fixture expectations while leaving shared runtime and traversal fixtures out of scope.

### Non-Functional Requirements
- NFR-000-1: Symbol-query fixtures must be deterministic and small enough for normal CLI test runs.
- NFR-000-2: Symbol-query validation must exercise the same SCIP loading and traversal path used by the `symbols` command.
- NFR-000-3: Golden symbol cases must preserve exact full SCIP symbol strings from the fixture index with no normalization or rewriting.
- NFR-000-4: Symbol fixture stories must not redefine shared command invocation, index loading, stdout/stderr placement, process status, or raw traversal fixture construction.

### Related External Components
- Component C-001 - SCIP symbol identifiers: Human-readable symbol strings made of scheme, package manager, package name, package version, and descriptors.
- Component C-002 - SCIP traversal view: The downstream query input from the traversal epic, including symbol inventories, document paths, and definition occurrences.
- Component C-003 - Query fixture set: Deterministic SCIP test data and expected JSON cases used by maintainers to validate symbol and package queries.

### Interfaces
- I-001-001 - Symbol discovery query contract (Interface 001 of Component C-001): The `symbols --name` query accepts a partial name and returns matched symbol results that include exact full SCIP symbol strings.

### Out of Scope
- Package query fixture expectations, package prefix filtering, package de-duplication, and package result JSON shape.
- Shared traversal fixtures for documents, occurrences, relationships, hover metadata, and source range preservation.
- Reference and implementation query fixtures, exact-symbol lookup behavior, missing exact-symbol behavior, and relationship traversal cases.
- Large real-world repository fixtures, performance benchmarks, external indexer installation, ctags fallback fixtures, and generated indexer workflows.
- Shared runtime fixtures for command routing, `--index`, stdout/stderr failures, exit status, and malformed index loading.

### Assumptions
- **ASM-000-1**: Symbol-query fixtures may reuse shared traversal fixture data as long as the query-specific cases own only command inputs and expected symbol-query payloads. - *Why*: CAP-003 explicitly depends on epic-planning-2 traversal fixtures while owning discovery-query behavior. - Confidence: HIGH
- **ASM-000-2**: Symbol golden cases can assert exact result order by the observable full `symbol` strings rather than by internal traversal order. - *Why*: CAP-001 already assumes stable lexical ordering by full SCIP symbol string, and CAP-003 requires stable output ordering. - Confidence: MEDIUM

### Open Questions
- None.

---

## Story ST-001 - Provide Symbol Fixture Data for Matching Cases

### References
- parent epic: `specs/epics/readme/20260517-134857-epic-planning-3.md#capability-cap-003---validate-discovery-queries-with-fixtures` - Requires fixtures with exact-looking partial matches, partial substring matches, overlapping names, and empty successful cases.
- sibling story: `specs/stories/readme/20260517-134857-epic-planning-3/cap-001-symbol-name-discovery.md#story-st-001---match-partial-symbol-names-deterministically` - Defines literal matching, successful ambiguity, empty results, and stable order.

### User Story
**As a** CLI Maintainer, **I want to** maintain deterministic symbol fixture data with known matching and non-matching symbol names, **so that** `symbols --name` validation can prove literal partial-name behavior without relying on external indexers or large repositories.

### Acceptance Criteria
- AC-001-1: Given the symbol-query fixture data, when maintainers inspect the indexed symbols used by the fixture, then it includes at least one symbol whose descriptor or display name exactly equals a likely `--name` query value.
- AC-001-2: Given the symbol-query fixture data, when maintainers inspect the indexed symbols used by the fixture, then it includes at least one symbol whose descriptor or display name contains a query value as a non-exact substring.
- AC-001-3: Given the symbol-query fixture data, when maintainers inspect the indexed symbols used by the fixture, then it includes multiple symbols with overlapping names that can all match the same supplied `--name` value.
- AC-001-4: Given the symbol-query fixture data, when maintainers inspect the indexed symbols used by the fixture, then it includes at least one valid `--name` query value that matches no indexed symbol.
- AC-001-5: Given the symbol-query fixture data, when maintainers inspect each expected non-empty symbol match, then each expected match records the exact full SCIP symbol string that should appear in query output.
- AC-001-6: Given the symbol-query fixture data is used by validation, when maintainers run the fixture-backed symbol cases repeatedly, then the fixture inputs do not depend on an external indexer installation or mutable repository state.

### Depends on:
Implementation ordering:
- Story document `cap-001-symbol-name-discovery.md` - Symbol matching and symbol payload behavior must be defined before fixture cases can assert them.

Run time coupling:
- I-001-001 - Symbol discovery query contract

### Out of Scope
- Defining package-query fixture data.
- Creating shared traversal fixture coverage for source ranges, hover text, relationships, or occurrence lookup behavior.
- Testing references, implementations, exact-symbol lookup, fuzzy matching, regex matching, semantic matching, or cross-index matching.

### Assumptions
- **ASM-001-1**: Fixture symbol names may be synthetic as long as their SCIP symbol strings are valid and deterministic. - *Why*: CAP-003 excludes large real-world fixtures and external indexer installation while requiring deterministic coverage. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-002 - Assert Symbol Query Golden Cases

### References
- parent epic: `specs/epics/readme/20260517-134857-epic-planning-3.md#capability-cap-003---validate-discovery-queries-with-fixtures` - Requires golden JSON cases proving full SCIP symbol strings, ordering, and empty collections.
- sibling story: `specs/stories/readme/20260517-134857-epic-planning-3/cap-001-symbol-name-discovery.md#story-st-002---return-full-scip-symbols-with-match-context` - Defines the expected full symbol and match context fields for symbol results.
- sibling story: `specs/stories/readme/20260517-134857-epic-planning-3/cap-001-symbol-name-discovery.md#story-st-003---validate-symbol-query-fixtures` - Requires golden cases for multi-match, empty, ordered, and full-symbol outputs.

### User Story
**As a** CLI Maintainer, **I want to** compare `symbols --name` results against symbol-specific golden JSON cases, **so that** future changes preserve successful ambiguity, empty results, exact full symbols, and deterministic ordering.

### Acceptance Criteria
- AC-002-1: Given the symbol-query fixture data contains a symbol whose name exactly equals the supplied query text, when maintainers run the matching golden case, then the expected JSON includes that exact full SCIP symbol string in the `symbols` collection.
- AC-002-2: Given the symbol-query fixture data contains a symbol whose name contains the supplied query text as a substring, when maintainers run the substring golden case, then the expected JSON includes that exact full SCIP symbol string in the `symbols` collection.
- AC-002-3: Given the symbol-query fixture data contains multiple overlapping names that match one supplied query text, when maintainers run the ambiguous-name golden case, then the expected JSON is a successful multi-result `symbols` collection rather than an error payload.
- AC-002-4: Given the symbol-query fixture data contains no match for a supplied query text, when maintainers run the empty-result golden case, then the expected JSON contains an empty `symbols` collection.
- AC-002-5: Given a symbol-query golden case returns more than one symbol, when maintainers compare the expected JSON to actual output, then the expected `symbols` order is stable and based on observable symbol result values.
- AC-002-6: Given a non-empty symbol-query golden case is evaluated, when maintainers inspect each expected result entry, then each entry includes the symbol result fields required by the CAP-001 payload story and does not require fields owned only by reference or implementation queries.

### Depends on:
Implementation ordering:
- Story ST-001 - Provide Symbol Fixture Data for Matching Cases
- Story document `cap-001-symbol-name-discovery.md` - Symbol payload fields must be defined before golden symbol entries can assert them.

Run time coupling:
- I-001-001 - Symbol discovery query contract

### Out of Scope
- Golden cases for `packages`, `references`, or `implementations`.
- Shared runtime error golden cases for bad flags, missing index paths, unreadable indexes, malformed SCIP files, stderr diagnostics, or exit statuses.
- Query result enrichment from source files, hover text, or relationship traversal.

### Assumptions
- **ASM-002-1**: Symbol golden JSON cases assert query-specific payload contents and rely on the shared runtime story for stdout-only success behavior. - *Why*: CAP-003 coordinates with epic-planning-1 but does not own shared stream contracts. - Confidence: HIGH

### Open Questions
- None.
