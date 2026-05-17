# User Stories: Traversal Fixture Coverage Validation

Status: review

## Goal
Maintainers can validate that the shared traversal fixture set covers every SCIP data category required by downstream query planners before query-specific stories rely on it.

## Parent Epic
specs/epics/readme/20260517-100328-epic-planning-2.md - Capability CAP-004

## Context
The schema-valid traversal fixture set is only useful if maintainers can prove it exercises the traversal contracts from CAP-001 through CAP-003. This document defines coverage validation over the shared fixture data while leaving command-specific query fixtures and final JSON goldens to sibling epics.

## Personas
- **CLI Maintainer**: A Go developer maintaining `scip-search`, needing fast fixture coverage checks that fail when shared traversal data no longer exercises required SCIP categories.
- **Story Writer**: A downstream planning agent writing query implementation stories, needing validated fixture coverage for documents, occurrences, symbols, relationships, ranges, and hover metadata.

## General information

Applies to: validation of the CAP-004 shared traversal fixture set.

### References
- goal spec: README.md#language-support - Requires `scip-search` to read SCIP output directly through official bindings.
- parent epic: specs/epics/readme/20260517-100328-epic-planning-2.md#capability-cap-004---provide-traversal-fixtures-for-scip-data-coverage - Requires traversal fixture coverage for downstream query planners.
- story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-004-01-schema-valid-traversal-fixtures.md - Defines the fixture data categories that coverage validation checks.
- dependency story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-001-02-document-and-symbol-inventory.md - Defines document, symbol, external symbol, and occurrence inventory coverage.
- dependency story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-002-01-source-range-normalization.md - Defines range, enclosing range, position encoding, and role coverage.
- dependency story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-002-02-hover-metadata-access.md - Defines hover metadata coverage.
- dependency story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-003-01-occurrence-lookup-view.md - Defines occurrence lookup coverage by exact full SCIP symbol.
- dependency story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-003-02-relationship-lookup-view.md - Defines relationship lookup and edge-kind coverage.
- sibling boundary: specs/stories/readme/20260517-134857-epic-planning-3/CAP-003-03-golden-json-validation.md - Owns discovery-query JSON golden validation, which this traversal coverage validation must not redefine.

### Non-Functional Requirements
- NFR-000-1: Coverage validation must be deterministic over the same fixture payloads and suitable for normal maintainer test runs.
- NFR-000-2: Coverage validation must exercise traversal behavior through official SCIP binding data rather than inspecting a custom fixture manifest as the source of truth.
- NFR-000-3: Coverage failures must identify the missing traversal data category well enough for maintainers to repair the shared fixture.
- NFR-000-4: Coverage validation must not assert final command JSON schemas, stdout/stderr behavior, process statuses, performance budgets, or query-specific selection semantics.

### Related External Components
- Component C-001 - Official SCIP Go bindings: Go package used to load schema-valid fixture data and expose traversal facts.
- Component C-002 - Query planner traversal view: The reusable process-local traversal surface whose fixture coverage is validated.
- Component C-003 - Shared traversal fixture set: Deterministic SCIP payloads used by traversal and downstream query-planning stories.

### Interfaces
- I-001-002 - Query planner traversal view (Interface 002 of Component C-001): Coverage validation observes fixture-backed traversal facts through the same views query planners use.

### Out of Scope
- Creating additional command-specific fixtures for `symbols`, `packages`, `references`, or `implementations`.
- Golden JSON validation for command responses, shared runtime stdout/stderr validation, exit statuses, missing-symbol command behavior, and malformed-index diagnostics.
- Installing external indexers, generating fixtures from large repositories, benchmarking traversal, and ctags fallback fixture validation.

### Assumptions
- **ASM-000-1**: Coverage validation can use category-level assertions over traversal facts instead of asserting a final command response payload. - *Why*: CAP-004 validates shared traversal coverage, while query-specific epics own command outputs and goldens. - Confidence: HIGH
- **ASM-000-2**: Maintainer-facing coverage failures should name missing categories such as document, occurrence, relationship, range, or hover metadata rather than exposing low-level protobuf implementation details. - *Why*: The story value is fast fixture repair, and exact internal assertion mechanics remain implementation-owned. - Confidence: MEDIUM

### Open Questions
- None.

---

## Story ST-001 - Validate Traversal Inventory and Lookup Coverage

### References
- parent epic: specs/epics/readme/20260517-100328-epic-planning-2.md#capability-cap-004---provide-traversal-fixtures-for-scip-data-coverage - Requires fixture validation for documents, occurrences, symbols, and relationships.
- story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-004-01-schema-valid-traversal-fixtures.md - Defines the shared fixture data to validate.
- dependency story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-003-01-occurrence-lookup-view.md - Defines occurrence lookup behavior by exact full SCIP symbol.
- dependency story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-003-02-relationship-lookup-view.md - Defines relationship lookup by owner and target symbol plus edge-kind preservation.

### User Story
**As a** CLI Maintainer, **I want to** run fixture coverage validation that proves traversal exposes document, symbol, occurrence, and relationship lookup facts, **so that** downstream query planners can rely on the shared fixture baseline without each command rebuilding it.

### Acceptance Criteria
- AC-001-1: Given maintainers run traversal fixture coverage validation, when the shared fixture is loaded, then validation proves at least two documents are exposed through the traversal document inventory.
- AC-001-2: Given coverage validation inspects symbol inventories, when local and external symbols are checked, then validation proves at least one document-level symbol and at least one external symbol are available as distinct traversal facts.
- AC-001-3: Given coverage validation inspects occurrence lookup, when it queries a full SCIP symbol known to exist in the fixture, then validation proves every matching occurrence remains associated with its containing document and exact full SCIP symbol.
- AC-001-4: Given coverage validation inspects occurrence lookup for a full SCIP symbol absent from occurrences, when the lookup runs, then validation proves traversal exposes an empty lookup result without treating it as a command-level missing-symbol failure.
- AC-001-5: Given coverage validation inspects relationship lookup, when owner and target symbol access patterns are checked, then validation proves relationship source symbols, target symbols, and original direction remain distinguishable.
- AC-001-6: Given fixture relationships include schema-defined edge-kind coverage, when coverage validation inspects relationship facts, then validation proves reference, implementation, definition, and type-definition flags are each available from the shared fixture.
- AC-001-7: Given fixture coverage validation runs, when maintainers inspect its scope, then it does not assert final `references`, `implementations`, `symbols`, or `packages` result selection, grouping, duplicate elimination, ordering, or JSON fields.

### Depends on:
Implementation ordering:
- Story document CAP-004-01-schema-valid-traversal-fixtures.md - Coverage validation requires the shared fixture set to exist.
- Story document CAP-003-01-occurrence-lookup-view.md - Occurrence lookup validation requires symbol-keyed lookup behavior.
- Story document CAP-003-02-relationship-lookup-view.md - Relationship validation requires owner and target lookup behavior.

Run time coupling:
- I-001-002 - Query planner traversal view

### Out of Scope
- Asserting command-level missing-symbol outcomes.
- Validating reference or implementation query result payloads.
- Synthesizing inverse relationships or cross-index relationship facts.

### Assumptions
- **ASM-001-1**: Coverage validation may choose named fixture symbols for lookup assertions, provided those names are fixture-local examples and do not define user-facing query semantics. - *Why*: Deterministic coverage needs stable fixture handles, while query-specific stories own public command behavior. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-002 - Validate Range and Hover Metadata Coverage

### References
- parent epic: specs/epics/readme/20260517-100328-epic-planning-2.md#capability-cap-004---provide-traversal-fixtures-for-scip-data-coverage - Requires fixture validation for ranges and hover metadata.
- story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-004-01-schema-valid-traversal-fixtures.md - Defines range, role, and hover fixture variants.
- dependency story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-002-01-source-range-normalization.md - Defines preserved source-location metadata.
- dependency story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-002-02-hover-metadata-access.md - Defines preserved symbol and occurrence hover metadata.

### User Story
**As a** Story Writer planning query metadata behavior, **I want to** rely on coverage validation for range and hover metadata variants, **so that** downstream stories can reference proven traversal facts instead of restating SCIP schema coverage.

### Acceptance Criteria
- AC-002-1: Given maintainers run traversal fixture coverage validation, when source-location coverage is checked, then validation proves fixture documents expose relative paths and position encodings through traversal.
- AC-002-2: Given occurrence ranges are checked, when coverage validation inspects fixture occurrences, then validation proves same-line and multi-position SCIP range forms are preserved as SCIP range values.
- AC-002-3: Given enclosing range coverage is checked, when coverage validation inspects fixture occurrences, then validation proves both present enclosing ranges and absent enclosing ranges remain distinguishable.
- AC-002-4: Given role coverage is checked, when coverage validation inspects fixture occurrences, then validation proves definition, non-definition, and multi-bit role cases remain available as SCIP role bitsets.
- AC-002-5: Given symbol hover coverage is checked, when coverage validation inspects fixture symbols, then validation proves symbol kind, display name, documentation, and signature documentation are available for at least one symbol.
- AC-002-6: Given occurrence hover coverage is checked, when coverage validation inspects fixture occurrences, then validation proves occurrence override documentation is available for at least one occurrence and absent for at least one other occurrence.
- AC-002-7: Given range or hover coverage validation runs, when maintainers inspect the assertions, then validation does not render Markdown, read source files, convert coordinates for editors, choose final command response fields, or decide whether command results include hover metadata.

### Depends on:
Implementation ordering:
- Story document CAP-004-01-schema-valid-traversal-fixtures.md - Range and hover coverage validation requires the shared fixture variants to exist.
- Story document CAP-002-01-source-range-normalization.md - Source-location coverage validation requires preserved range and role behavior.
- Story document CAP-002-02-hover-metadata-access.md - Hover coverage validation requires preserved symbol and occurrence metadata behavior.

Run time coupling:
- I-001-002 - Query planner traversal view

### Out of Scope
- Formatting source locations for final JSON.
- Rendering documentation for humans.
- Deciding command-specific metadata inclusion, filtering, or ordering.

### Assumptions
- **ASM-002-1**: Coverage validation treats preserved absence of optional metadata as a covered behavior when the fixture intentionally includes that absence. - *Why*: CAP-002 stories require absence to remain distinguishable instead of traversal inventing values. - Confidence: HIGH

### Open Questions
- None.
