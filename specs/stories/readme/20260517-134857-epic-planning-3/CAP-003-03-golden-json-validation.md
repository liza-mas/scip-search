# User Stories: Golden JSON Validation for Discovery Queries

Status: review

## Goal
Maintainers can validate symbol JSON modes and package discovery commands against golden JSON cases that prove stable successful payloads for matching, filtering, de-duplication, ambiguous names, empty results, and output ordering.

## Parent Epic
`specs/epics/readme/20260517-134857-epic-planning-3.md` - Capability CAP-003, "Validate discovery queries with fixtures"

## Context
CAP-003 needs fixture-backed expected JSON for package discovery and for the explicit symbol JSON modes. This document defines the cross-command golden validation expectations while leaving the default one-line symbol output matrix to CAP-003-01 and the package fixture matrix to CAP-003-02. The shared runtime epic owns stdout stream placement and runtime errors; these stories only assert successful query-specific JSON payloads.

## Personas
- **CLI Maintainer**: a Go developer maintaining `scip-search`, needing golden JSON cases that make discovery-query regressions obvious in normal validation runs.
- **Automation Agent**: an AI or script-driven caller parsing `scip-search` output, needing stable JSON collections whose ordering and empty states are predictable.

## General information

Applies to: successful golden JSON validation for `symbols --nested-json`, `symbols --json`, and package discovery queries.

### References
- goal spec: `README.md#what-is-scip-search` - Requires explicit JSON modes to print structured JSON for query results.
- goal spec: `README.md#scip-symbol-format` - Defines full SCIP symbols and package identity components used by golden results.
- parent epic: `specs/epics/readme/20260517-134857-epic-planning-3.md#general-information` - Requires deterministic ordering, explicit empty collections, no diagnostics in stdout, and small deterministic query-specific fixtures.
- parent epic: `specs/epics/readme/20260517-134857-epic-planning-3.md#capability-cap-003---validate-discovery-queries-with-fixtures` - Requires golden JSON cases for symbol and package discovery validation.
- consistency check: `specs/stories/readme/20260517-134857-epic-planning-3/CAP-003-01-symbol-query-fixtures.md` - Defines symbol fixture cases that golden JSON must cover.
- consistency check: `specs/stories/readme/20260517-134857-epic-planning-3/CAP-003-02-package-query-fixtures.md` - Defines package fixture cases that golden JSON must cover.
- consistency check: `specs/stories/readme/20260517-095535-epic-planning-1/stdout-json-success-contract.md` - Owns shared successful JSON stdout behavior.
- consistency check: `specs/stories/readme/20260517-095535-epic-planning-1/stderr-exit-error-contract.md` - Owns shared runtime failure behavior outside these successful golden cases.

### Non-Functional Requirements
- NFR-000-1: Golden JSON cases must be deterministic for repeated validation against the same fixture data.
- NFR-000-2: Golden JSON cases must assert query-specific payload fields only, while preserving the shared runtime contract by expecting no success diagnostics in stdout.
- NFR-000-3: Golden JSON validation must exercise the same SCIP loading and traversal path used by the discovery commands.
- NFR-000-4: Golden JSON validation must be small enough for normal maintainer test runs and must not depend on external indexer installation or large real-world repositories.

### Related External Components
- Component C-001 - SCIP symbol identifiers: Human-readable symbol strings and package identity components used in discovery query results.
- Component C-003 - Query fixture set: Deterministic SCIP test data and expected JSON cases used by maintainers to validate discovery queries.
- Component C-004 - Calling process environment: The shell, script, or agent that parses successful JSON stdout under the shared runtime contract.

### Interfaces
- I-001-001 - Symbol discovery query contract (Interface 001 of Component C-001): `symbols --name --nested-json` returns a compact successful JSON payload with a `packages` collection and nested symbol descriptors; `symbols --name --json` returns a compatibility `symbols` collection.
- I-001-002 - Package discovery query contract (Interface 002 of Component C-001): The `packages` query returns a successful JSON payload with a `packages` collection.
- I-004-001 - CLI process contract (Interface 001 of Component C-004): The shared successful runtime contract that provides parseable JSON on stdout.

### Out of Scope
- Defining shared JSON envelope mechanics, stdout/stderr placement rules, nonzero status behavior, invocation errors, index-loading errors, or malformed SCIP errors.
- Constructing raw SCIP fixtures for traversal metadata beyond what discovery queries need.
- Reference and implementation query golden cases, exact-symbol missing behavior, relationship traversal cases, and location/range JSON for those commands.
- Default one-line symbol output, pretty-print options, human tables, performance benchmarks, large real-world fixtures, external indexer installation, and ctags fallback fixtures.

### Assumptions
- **ASM-000-1**: Golden JSON validation compares normalized JSON values rather than relying on object field order. - *Why*: The user-observable contract is parseable JSON content and array ordering, while object member order is not a reliable JSON semantic. - Confidence: HIGH
- **ASM-000-2**: Golden cases can be organized per command while sharing one deterministic fixture source. - *Why*: CAP-003 needs validation coverage for both discovery commands but does not require separate physical fixture files. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Validate Symbol Discovery Golden JSON

### References
- parent epic: `specs/epics/readme/20260517-134857-epic-planning-3.md#capability-cap-003---validate-discovery-queries-with-fixtures` - Requires golden JSON cases for matching, ambiguous names, empty results, and stable ordering.
- story document: `specs/stories/readme/20260517-134857-epic-planning-3/CAP-003-01-symbol-query-fixtures.md` - Defines symbol query fixture cases and expected coverage.
- consistency check: `specs/stories/readme/20260517-134857-epic-planning-3/cap-001-symbol-name-discovery.md` - Defines symbol payload fields and matching semantics.

### User Story
**As a** CLI Maintainer, **I want to** validate `symbols --name --nested-json` and `symbols --name --json` successful responses against golden JSON, **so that** package grouping, exact full-symbol reconstruction, match context, ambiguous results, empty results, and result ordering stay stable.

### Acceptance Criteria
- AC-001-1: Given the symbol fixture golden case for `symbols --name Supervisor --nested-json`, when maintainers run golden validation, then the expected JSON contains a top-level `packages` collection with every ambiguous matching symbol descriptor nested under its package identity and no package-only or unrelated symbol entries.
- AC-001-2: Given the symbol fixture golden case for a partial substring query such as `symbols --name Run --nested-json`, when maintainers run golden validation, then the expected JSON proves partial literal matching and exact full SCIP symbol reconstruction from `packageKey + " " + descriptor`.
- AC-001-3: Given the symbol fixture golden case for `symbols --name DoesNotExist --nested-json`, when maintainers run golden validation, then the expected JSON contains an explicit empty `packages` collection.
- AC-001-4: Given any symbol golden case with multiple results, when maintainers compare the expected JSON, then array order is asserted as stable ascending order by observable package key and reconstructed full symbol value.
- AC-001-5: Given symbol golden validation runs, when maintainers inspect the expected JSON cases, then they assert only successful query-specific `symbols` payload behavior and do not assert shared runtime failure payloads, stderr text, reference results, implementation results, or package query payloads.

### Depends on:
Implementation ordering:
- `CAP-003-01-symbol-query-fixtures.md` - Symbol fixture cases must be defined before their golden JSON cases can be validated.
- `cap-001-symbol-name-discovery.md` - Symbol matching and payload contracts must be defined before golden JSON can assert them.

Run time coupling:
- Interface I-001-001 - Symbol discovery query contract
- Interface I-004-001 - CLI process contract

### Out of Scope
- Golden cases for package discovery, references, implementations, runtime failures, malformed indexes, and invalid invocations.
- Object field order as a semantic assertion.

### Assumptions
- **ASM-001-1**: Nested symbol golden JSON cases use the same `packages` collection shape for non-empty and empty successful results. - *Why*: CAP-001 and the parent epic require explicit empty collections for no-match JSON discovery, and `--nested-json` groups symbols under packages. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-002 - Validate Package Discovery Golden JSON

### References
- parent epic: `specs/epics/readme/20260517-134857-epic-planning-3.md#capability-cap-003---validate-discovery-queries-with-fixtures` - Requires golden JSON cases for filtering, empty results, de-duplication, and stable ordering.
- story document: `specs/stories/readme/20260517-134857-epic-planning-3/CAP-003-02-package-query-fixtures.md` - Defines package query fixture cases and expected coverage.
- consistency check: `specs/stories/readme/20260517-134857-epic-planning-3/CAP-002-01-package-inventory.md` - Defines package inventory and de-duplication semantics.
- consistency check: `specs/stories/readme/20260517-134857-epic-planning-3/CAP-002-02-package-prefix-filtering.md` - Defines prefix filtering and empty-result semantics.
- consistency check: `specs/stories/readme/20260517-134857-epic-planning-3/CAP-002-03-package-result-json-shape.md` - Defines package payload fields.

### User Story
**As a** CLI Maintainer, **I want to** validate `packages` successful responses against golden JSON, **so that** package inventory, prefix filtering, de-duplication, empty results, and result ordering stay stable.

### Acceptance Criteria
- AC-002-1: Given the unfiltered `packages` fixture golden case, when maintainers run golden validation, then the expected JSON contains a top-level `packages` collection with one entry for each distinct fixture package identity.
- AC-002-2: Given a package fixture includes multiple symbols from the same package identity, when maintainers inspect the unfiltered golden JSON, then the duplicated package identity appears once in `packages`.
- AC-002-3: Given the golden case for `packages --prefix github.com/liza-mas/`, when maintainers run golden validation, then expected JSON includes only package entries whose `packageName` begins with that literal prefix.
- AC-002-4: Given the golden case for `packages --prefix github.com/no-match/`, when maintainers run golden validation, then expected JSON contains an explicit empty `packages` collection.
- AC-002-5: Given any package golden case with multiple results, when maintainers compare expected JSON, then array order is asserted as stable ascending order by observable `packageKey`.
- AC-002-6: Given package golden validation runs, when maintainers inspect expected JSON cases, then each non-empty package entry asserts `scheme`, `packageManager`, `packageName`, `packageVersion`, and `packageKey` without asserting symbol descriptors, source locations, dependency graph data, or registry metadata.

### Depends on:
Implementation ordering:
- `CAP-003-02-package-query-fixtures.md` - Package fixture cases must be defined before their golden JSON cases can be validated.
- `CAP-002-01-package-inventory.md` - Package identity and de-duplication semantics must be defined before golden JSON can assert them.
- `CAP-002-02-package-prefix-filtering.md` - Prefix filtering semantics must be defined before filtered golden JSON can assert them.
- `CAP-002-03-package-result-json-shape.md` - Package JSON fields must be defined before golden JSON can assert them.

Run time coupling:
- Interface I-001-002 - Package discovery query contract
- Interface I-004-001 - CLI process contract

### Out of Scope
- Golden cases for symbol discovery, references, implementations, runtime failures, malformed indexes, invalid invocations, dependency graph queries, or package registry queries.
- Object field order as a semantic assertion.

### Assumptions
- **ASM-002-1**: Package golden JSON cases use the same `packages` collection shape for unfiltered, filtered, and empty successful results. - *Why*: CAP-002 defines one package query payload shape across package discovery modes. - Confidence: HIGH

### Open Questions
- None.
