# User Stories: Golden JSON Validation

Status: review

## Goal
`scip-search` maintainers can run golden JSON validation for both discovery commands and prove stable successful payloads for symbol matching, package filtering, empty results, de-duplication, ambiguous names, and output ordering.

## Parent Epic
`specs/epics/readme/20260517-134857-epic-planning-3.md` - Capability CAP-003, "Validate discovery queries with fixtures"

## Context
Symbol and package fixture data only protects behavior if maintainers can compare command results with expected JSON. This document defines the cross-command golden validation expectations for successful discovery query payloads while leaving shared runtime stream behavior to epic-planning-1 and raw SCIP traversal fixture construction to epic-planning-2.

## Personas
- **Automation Agent**: an AI or script-driven caller running `scip-search` in a terminal or sandbox, needing deterministic JSON results that can be compared across runs.
- **CLI Maintainer**: a Go developer maintaining `scip-search`, needing fixture-backed golden cases that keep discovery query behavior stable across code changes.

## General information

Applies to: golden JSON validation coverage for successful `symbols --name` and `packages` discovery commands.

### References
- goal spec: `README.md#what-is-scip-search` - Requires `scip-search` commands to print structured JSON to stdout and lists the `symbols` and `packages` command forms.
- goal spec: `README.md#scip-symbol-format` - Requires full SCIP symbol strings for symbol discovery and defines package identity components used by package discovery.
- parent epic: `specs/epics/readme/20260517-134857-epic-planning-3.md#general-information` - Requires deterministic ordering, explicit empty result collections, and query-specific fixtures through the same SCIP loading and traversal path used by commands.
- parent epic: `specs/epics/readme/20260517-134857-epic-planning-3.md#capability-cap-003---validate-discovery-queries-with-fixtures` - Requires golden JSON cases proving full SCIP symbol strings, package identities, ordering, and empty collections.
- sibling story: `specs/stories/readme/20260517-134857-epic-planning-3/CAP-003-01-symbol-query-fixtures.md` - Defines symbol-query fixture and symbol-specific golden case coverage.
- sibling story: `specs/stories/readme/20260517-134857-epic-planning-3/CAP-003-02-package-query-fixtures.md` - Defines package-query fixture and package-specific golden case coverage.
- consistency story: `specs/stories/readme/20260517-095535-epic-planning-1/stdout-json-success-contract.md` - Owns stdout stream purity and shared successful process behavior.
- consistency scan: `specs/stories/readme/20260517-134857-epic-planning-3/` - Existing CAP-001 and CAP-002 story documents define the query-specific payloads that golden JSON validation must assert.

### Non-Functional Requirements
- NFR-000-1: Golden JSON cases must be deterministic across repeated runs against the same fixture index.
- NFR-000-2: Golden JSON validation must exercise the same SCIP loading and traversal path used by the discovery commands.
- NFR-000-3: Golden JSON cases must assert query-specific payload contents without redefining shared stdout/stderr, exit status, or runtime error behavior.
- NFR-000-4: Golden JSON cases must remain small enough for normal CLI test runs and must not require external indexer installation.

### Related External Components
- Component C-001 - SCIP symbol identifiers: Human-readable symbol strings and package identity components used by discovery payloads.
- Component C-002 - SCIP traversal view: The downstream query input from the traversal epic.
- Component C-003 - Query fixture set: Deterministic SCIP test data and expected JSON cases used by maintainers to validate symbol and package queries.
- Component C-004 - CLI process contract: The shared successful runtime contract that provides a parseable JSON value on stdout.

### Interfaces
- I-001-001 - Symbol discovery query contract (Interface 001 of Component C-001): The `symbols --name` query accepts a partial name and returns matched symbol results that include exact full SCIP symbol strings.
- I-001-002 - Package discovery query contract (Interface 002 of Component C-001): The `packages` query returns package identities from the index, optionally filtered by package-name prefix.
- I-004-001 - Successful JSON stdout contract (Interface 001 of Component C-004): The shared runtime contract that successful query commands emit structured JSON on stdout.

### Out of Scope
- Shared runtime fixtures for command routing, `--index`, stdout/stderr failures, exit status, and malformed index loading.
- Shared traversal fixtures for documents, occurrences, relationships, hover metadata, and source range preservation.
- Reference and implementation query fixtures, exact-symbol lookup behavior, missing exact-symbol behavior, relationship traversal cases, and location/range JSON for those commands.
- Large real-world repository fixtures, performance benchmarks, external indexer installation, ctags fallback fixtures, generated indexer workflows, alternate output formats, and human-readable tables.

### Assumptions
- **ASM-000-1**: Cross-command golden validation can live in a separate story document from command-specific fixture coverage because it coordinates expected JSON comparison behavior for both discovery commands. - *Why*: CAP-003 lists a distinct golden JSON validation story document in addition to symbol and package fixture documents. - Confidence: HIGH
- **ASM-000-2**: Golden JSON validation asserts successful query payloads and does not need separate error golden cases for no-match results because no-match discovery is represented as successful empty collections. - *Why*: The epic and CAP-001/CAP-002 stories treat no-match discovery as successful, while shared runtime failures are owned by epic-planning-1. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Validate Successful Discovery Payloads Against Golden JSON

### References
- parent epic: `specs/epics/readme/20260517-134857-epic-planning-3.md#capability-cap-003---validate-discovery-queries-with-fixtures` - Requires golden JSON cases for symbol and package discovery queries.
- sibling story: `specs/stories/readme/20260517-134857-epic-planning-3/CAP-003-01-symbol-query-fixtures.md#story-st-002---assert-symbol-query-golden-cases` - Defines symbol-specific golden cases.
- sibling story: `specs/stories/readme/20260517-134857-epic-planning-3/CAP-003-02-package-query-fixtures.md#story-st-002---assert-package-query-golden-cases` - Defines package-specific golden cases.
- consistency story: `specs/stories/readme/20260517-095535-epic-planning-1/stdout-json-success-contract.md` - Defines shared successful JSON stdout behavior that validation must not duplicate.

### User Story
**As a** CLI Maintainer, **I want to** validate successful discovery command output against committed golden JSON cases, **so that** symbol and package query behavior remains stable when implementation details change.

### Acceptance Criteria
- AC-001-1: Given the deterministic query fixture set is available, when maintainers run golden validation for `symbols --name`, then validation compares the command's successful query payload to committed symbol golden JSON cases.
- AC-001-2: Given the deterministic query fixture set is available, when maintainers run golden validation for `packages` without `--prefix`, then validation compares the command's successful query payload to a committed unfiltered package golden JSON case.
- AC-001-3: Given the deterministic query fixture set is available, when maintainers run golden validation for `packages --prefix <prefix>`, then validation compares the command's successful query payload to committed filtered package golden JSON cases.
- AC-001-4: Given any committed discovery-query golden case is evaluated, when actual output differs from expected JSON in a query-specific field, collection entry, or entry order, then validation reports that golden case as failing.
- AC-001-5: Given any committed discovery-query golden case is evaluated, when shared runtime behavior such as stdout stream purity or process status is already provided by the runtime contract, then this validation treats that behavior as a dependency rather than redefining separate runtime expectations.

### Depends on:
Implementation ordering:
- Story document `CAP-003-01-symbol-query-fixtures.md` - Symbol golden cases must exist before cross-command golden validation can include them.
- Story document `CAP-003-02-package-query-fixtures.md` - Package golden cases must exist before cross-command golden validation can include them.

Run time coupling:
- I-001-001 - Symbol discovery query contract
- I-001-002 - Package discovery query contract
- I-004-001 - Successful JSON stdout contract

### Out of Scope
- Golden validation for `references`, `implementations`, exact-symbol lookup, missing exact-symbol behavior, or relationship traversal.
- Golden validation for runtime failure payloads, stderr diagnostics, exit statuses, malformed invocation, missing index paths, unreadable indexes, or invalid SCIP files.
- Snapshot approval workflows, auto-updating golden files, pretty-printing policy, or alternate output formats.

### Assumptions
- **ASM-001-1**: Maintainers compare parsed JSON values for semantic equality while still treating collection order as significant. - *Why*: The epic requires structured JSON and stable output ordering, while object member ordering is not the user-visible contract. - Confidence: MEDIUM

### Open Questions
- None.

---

## Story ST-002 - Cover Required Discovery Edge Cases Across Golden JSON

### References
- parent epic: `specs/epics/readme/20260517-134857-epic-planning-3.md#capability-cap-003---validate-discovery-queries-with-fixtures` - Requires coverage for matching, filtering, empty results, ambiguous names, full SCIP symbols, package identities, ordering, and empty collections.
- task scope: `epic-planning-3-us-writing-2` - Requires fixture stories to cover symbol partial matching, ambiguous names, package prefix filtering, de-duplication, empty successful results, and stable output ordering for both discovery commands.
- sibling story: `specs/stories/readme/20260517-134857-epic-planning-3/cap-001-symbol-name-discovery.md` - Defines symbol query behavior covered by symbol golden cases.
- sibling story: `specs/stories/readme/20260517-134857-epic-planning-3/CAP-002-01-package-inventory.md` - Defines package de-duplication and ordering covered by package golden cases.
- sibling story: `specs/stories/readme/20260517-134857-epic-planning-3/CAP-002-02-package-prefix-filtering.md` - Defines package prefix filtering and empty filtered results covered by package golden cases.

### User Story
**As a** CLI Maintainer, **I want to** see the required discovery edge cases represented across the committed golden JSON suite, **so that** reviewers can verify CAP-003 coverage without reverse-engineering fixture internals.

### Acceptance Criteria
- AC-002-1: Given the committed discovery-query golden suite, when maintainers review symbol cases, then the suite includes a partial symbol match case, an ambiguous multi-symbol case, an empty successful `symbols` case, and a multi-result ordering case.
- AC-002-2: Given the committed discovery-query golden suite, when maintainers review package cases, then the suite includes an unfiltered de-duplication case, a package-name prefix filtering case, an empty successful `packages` case, and a multi-result ordering case.
- AC-002-3: Given the committed discovery-query golden suite, when maintainers review non-empty symbol cases, then each expected symbol entry includes the exact full SCIP symbol string required for later symbol-based commands.
- AC-002-4: Given the committed discovery-query golden suite, when maintainers review non-empty package cases, then each expected package entry includes the package identity fields and exact package key required by CAP-002.
- AC-002-5: Given the committed discovery-query golden suite, when maintainers review all cases together, then no case depends on reference queries, implementation queries, external indexer installation, large real-world fixtures, performance benchmarks, or ctags fallback behavior.

### Depends on:
Implementation ordering:
- Story ST-001 - Validate Successful Discovery Payloads Against Golden JSON
- Story document `CAP-003-01-symbol-query-fixtures.md` - Symbol coverage must be defined before suite-level coverage can assert it.
- Story document `CAP-003-02-package-query-fixtures.md` - Package coverage must be defined before suite-level coverage can assert it.

Run time coupling:
- I-001-001 - Symbol discovery query contract
- I-001-002 - Package discovery query contract

### Out of Scope
- Adding coverage for capabilities outside symbol and package discovery.
- Requiring golden cases to describe how SCIP fixture files are generated or serialized.
- Requiring performance thresholds, large fixture matrices, or language-indexer compatibility matrices.

### Assumptions
- **ASM-002-1**: Suite-level coverage can be demonstrated by the committed golden cases and their names or metadata, without adding a separate human-readable coverage report. - *Why*: The task requires fixture stories and golden JSON cases, not a reporting feature. - Confidence: MEDIUM

### Open Questions
- None.
