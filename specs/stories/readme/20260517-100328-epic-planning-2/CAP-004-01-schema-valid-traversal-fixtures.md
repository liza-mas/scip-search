# User Stories: Schema-Valid Traversal Fixtures

Status: review

## Goal
Maintainers can build a small deterministic SCIP fixture set that is schema-valid and exercises the traversal data categories required by downstream query planners.

## Parent Epic
specs/epics/readme/20260517-100328-epic-planning-2.md - Capability CAP-004

## Context
CAP-001 through CAP-003 define the traversal view over loaded SCIP data, preserved location and hover metadata, and occurrence and relationship lookup facts. This document defines the shared fixture data those traversal stories can use for validation without installing external indexers or defining query command golden JSON.

## Personas
- **CLI Maintainer**: A Go developer maintaining `scip-search`, needing compact fixture data that catches traversal regressions while staying faithful to official SCIP schema fields.
- **Story Writer**: A downstream planning agent writing query implementation stories, needing one shared traversal fixture baseline so query stories do not recreate document, occurrence, symbol, relationship, range, and hover coverage.

## General information

Applies to: deterministic traversal fixtures for the process-local SCIP traversal layer.

### References
- goal spec: README.md#language-support - Requires official SCIP bindings for parsing and traversal and says `scip-search` reads SCIP output directly.
- goal spec: README.md#scip-symbol-format - Shows full SCIP symbol strings used by later symbol-based queries.
- parent epic: specs/epics/readme/20260517-100328-epic-planning-2.md#capability-cap-004---provide-traversal-fixtures-for-scip-data-coverage - Defines fixture coverage for documents, occurrences, symbols, relationships, ranges, and hover metadata.
- dependency story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-001-01-scip-binding-input.md - Defines traversal input from official SCIP binding data.
- dependency story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-001-02-document-and-symbol-inventory.md - Defines document, symbol, external symbol, and occurrence inventories that fixtures must exercise.
- dependency story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-002-01-source-range-normalization.md - Defines document path, position encoding, range, enclosing range, and role metadata that fixtures must exercise.
- dependency story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-002-02-hover-metadata-access.md - Defines symbol and occurrence hover metadata that fixtures must exercise.
- dependency story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-003-01-occurrence-lookup-view.md - Defines symbol-keyed occurrence lookup facts that fixtures must exercise.
- dependency story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-003-02-relationship-lookup-view.md - Defines relationship lookup and edge-kind facts that fixtures must exercise.
- official source: https://raw.githubusercontent.com/scip-code/scip/main/scip.proto - Defines the schema elements represented by the fixture payloads.
- consistency check: specs/stories/readme/20260517-134857-epic-planning-3/CAP-003-01-symbol-query-fixtures.md - Confirms query-specific fixture and golden behavior belongs to discovery-query scope, not this traversal fixture scope.

### Non-Functional Requirements
- NFR-000-1: Traversal fixtures must be deterministic and small enough to run in normal maintainer validation.
- NFR-000-2: Fixture payloads must be schema-valid SCIP data read through the same official binding path used by traversal.
- NFR-000-3: Fixture data must preserve exact SCIP symbol strings, document paths, ranges, role bitsets, relationships, and hover metadata without normalization for command-specific behavior.
- NFR-000-4: Fixture stories must not require external indexer installation, real repository fixtures, performance benchmarks, final command JSON golden files, or ctags fallback data.

### Related External Components
- Component C-001 - Official SCIP Go bindings: Go package used to load and expose schema-defined SCIP fixture data.
- Component C-002 - SCIP protobuf schema: Protocol definition that defines valid documents, occurrences, symbols, relationships, ranges, roles, and hover metadata.
- Component C-003 - Query planner traversal view: The shared in-memory traversal surface validated by the fixture data.

### Interfaces
- I-001-001 - SCIP traversal input (Interface 001 of Component C-001): Traversal receives schema-valid fixture data through the same official binding path as other loaded SCIP indexes.
- I-001-002 - Query planner traversal view (Interface 002 of Component C-001): Query planners inspect fixture-backed documents, symbols, occurrences, relationships, ranges, and hover metadata through reusable traversal views.

### Out of Scope
- Installing or invoking `scip-go`, `scip-typescript`, or any external language indexer during fixture setup.
- Full CLI command tests for `symbols`, `packages`, `references`, or `implementations`.
- Golden JSON files for final command response schemas, stdout/stderr behavior, process exit statuses, or shared runtime failures.
- Large real-world repository fixtures, performance benchmarks, cross-index fixtures, ctags fallback data, and custom persisted index formats.

### Assumptions
- **ASM-000-1**: The shared traversal fixture may be synthetic when it remains schema-valid SCIP data consumed through official bindings. - *Why*: CAP-004 explicitly excludes external indexer installation and real-world fixture scale while requiring deterministic SCIP coverage. - Confidence: HIGH
- **ASM-000-2**: One compact fixture set may contain multiple documents and symbols rather than one physical fixture per traversal category. - *Why*: CAP-004 requires category coverage, not separate files, and smaller fixtures reduce maintainer validation cost. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Provide Document and Symbol Fixture Coverage

### References
- parent epic: specs/epics/readme/20260517-100328-epic-planning-2.md#capability-cap-004---provide-traversal-fixtures-for-scip-data-coverage - Requires fixtures to cover documents, occurrences, symbols, relationships, ranges, and hover metadata.
- dependency story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-001-02-document-and-symbol-inventory.md - Requires document, document-symbol, external-symbol, and occurrence inventories.
- goal spec: README.md#scip-symbol-format - Shows full SCIP symbol strings preserved for later symbol-based queries.

### User Story
**As a** CLI Maintainer, **I want to** maintain a deterministic SCIP fixture with multiple documents, document symbols, external symbols, and occurrences, **so that** traversal inventory stories can be validated against realistic schema categories without relying on external indexers.

### Acceptance Criteria
- AC-001-1: Given maintainers run traversal fixture validation, when the fixture is loaded through the official SCIP binding path, then at least two SCIP documents are available with distinct relative paths and document identities.
- AC-001-2: Given the fixture documents are inspected through traversal, when maintainers check document metadata, then each fixture document preserves its language and position encoding for planner use.
- AC-001-3: Given the fixture contains document-level symbol information, when maintainers inspect symbol inventory through traversal, then at least one local full SCIP symbol string is available from a document symbol entry.
- AC-001-4: Given the fixture contains external symbol information, when maintainers inspect external symbol inventory through traversal, then at least one external full SCIP symbol string is available separately from document-level symbols.
- AC-001-5: Given the fixture contains occurrences for more than one symbol, when maintainers inspect occurrence inventory, then occurrences remain associated with their exact full SCIP symbol strings and containing documents.
- AC-001-6: Given fixture data is used for traversal validation, when maintainers inspect the fixture scope, then it does not require command routing, package-prefix filtering, partial symbol matching, reference result selection, implementation result selection, or final command JSON schemas.

### Depends on:
Implementation ordering:
- Story document CAP-001-01-scip-binding-input.md - Fixture validation must use the same loaded SCIP binding path as traversal input.
- Story document CAP-001-02-document-and-symbol-inventory.md - Fixture coverage must align with traversal inventory expectations.

Run time coupling:
- I-001-001 - SCIP traversal input
- I-001-002 - Query planner traversal view

### Out of Scope
- Choosing package query payload fields or package de-duplication rules.
- Resolving partial names or ranking symbol matches.
- Generating fixture data by invoking external language indexers.

### Assumptions
- **ASM-001-1**: Distinct fixture documents are enough to validate traversal document separation without modeling a large repository tree. - *Why*: CAP-004 requires deterministic coverage and excludes large real-world repository fixtures. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-002 - Provide Range, Role, and Hover Metadata Fixture Coverage

### References
- parent epic: specs/epics/readme/20260517-100328-epic-planning-2.md#capability-cap-004---provide-traversal-fixtures-for-scip-data-coverage - Requires fixtures to cover ranges and hover metadata.
- dependency story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-002-01-source-range-normalization.md - Requires occurrence ranges, enclosing ranges, document position encoding, and role bitsets.
- dependency story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-002-02-hover-metadata-access.md - Requires symbol documentation, signature documentation, display metadata, and occurrence override documentation.
- official source: https://raw.githubusercontent.com/scip-code/scip/main/scip.proto - Defines occurrence range, enclosing range, symbol roles, symbol information, and override documentation fields.

### User Story
**As a** Story Writer planning location and hover behavior, **I want to** rely on traversal fixtures that include source ranges, role bitsets, and hover metadata variants, **so that** downstream query stories can cite shared coverage instead of redefining SCIP metadata examples.

### Acceptance Criteria
- AC-002-1: Given maintainers inspect fixture occurrences through traversal, when occurrence ranges are checked, then the fixture includes at least one same-line SCIP range form and at least one multi-position SCIP range form.
- AC-002-2: Given fixture occurrences are inspected, when enclosing range metadata is checked, then at least one occurrence has an enclosing range and at least one occurrence preserves the absence of an enclosing range.
- AC-002-3: Given fixture occurrences are inspected, when symbol role metadata is checked, then the fixture covers at least one definition occurrence, at least one non-definition occurrence, and at least one occurrence with multiple SCIP role bits available together.
- AC-002-4: Given fixture symbol information is inspected, when hover-related symbol metadata is checked, then the fixture includes symbol kind, display name, documentation, and signature documentation for at least one symbol.
- AC-002-5: Given fixture occurrences are inspected, when occurrence-level hover metadata is checked, then at least one occurrence exposes override documentation and at least one occurrence preserves absent override documentation.
- AC-002-6: Given maintainers use fixture metadata coverage, when validation observes ranges, roles, or hover facts, then traversal preserves SCIP-provided values without rendering Markdown, reading source files, converting editor coordinates, or choosing command JSON fields.

### Depends on:
Implementation ordering:
- Story document CAP-002-01-source-range-normalization.md - Range and role fixture expectations must align with traversal source-location preservation.
- Story document CAP-002-02-hover-metadata-access.md - Hover fixture expectations must align with traversal metadata preservation.
- Story ST-001 - Provide Document and Symbol Fixture Coverage

Run time coupling:
- I-001-002 - Query planner traversal view

### Out of Scope
- Human hover rendering, Markdown formatting, signature display formatting, syntax highlighting, diagnostics display, and source-file reads.
- Deciding whether command results include or omit hover metadata.
- Interpreting role bitsets as final reference, definition, or implementation command behavior.

### Assumptions
- **ASM-002-1**: Fixture validation can assert presence and absence of optional metadata without requiring malformed SCIP inputs. - *Why*: CAP-004 asks for schema-valid payloads, so absent optional fields should be represented as valid absence rather than invalid data. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-003 - Provide Relationship Edge Fixture Coverage

### References
- parent epic: specs/epics/readme/20260517-100328-epic-planning-2.md#capability-cap-004---provide-traversal-fixtures-for-scip-data-coverage - Requires fixtures to cover relationships and specifically calls for at least one implementation relationship.
- dependency story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-003-02-relationship-lookup-view.md - Requires relationship owner lookup, target lookup, original direction, and SCIP edge-kind preservation.
- dependency story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-001-02-document-and-symbol-inventory.md - Defines document-level and external symbol information that can own relationship facts.
- official source: https://raw.githubusercontent.com/scip-code/scip/main/scip.proto - Defines `SymbolInformation.relationships` and relationship edge-kind flags.

### User Story
**As a** CLI Maintainer implementing relationship traversal, **I want to** include schema-valid SCIP relationship facts in the shared traversal fixture, **so that** relationship lookup stories can be validated against concrete owner, target, direction, and edge-kind data.

### Acceptance Criteria
- AC-003-1: Given maintainers inspect the shared traversal fixture, when relationship data is checked, then at least one document-level or external symbol information entry owns a SCIP relationship.
- AC-003-2: Given a fixture relationship is inspected through traversal, when owner and target symbols are checked, then the relationship preserves the owning source symbol, the relationship target symbol, and the original source-to-target direction.
- AC-003-3: Given the fixture relationship edge kinds are inspected, when implementation coverage is checked, then at least one relationship has the SCIP implementation relationship flag available for traversal validation.
- AC-003-4: Given the fixture relationship edge kinds are inspected, when reference, definition, and type-definition coverage are checked, then the fixture includes relationship facts that make each of those SCIP edge-kind flags available for traversal validation.
- AC-003-4b: Given a fixture relationship has multiple SCIP edge-kind flags set, when traversal exposes that relationship, then the fixture preserves the flags together instead of requiring separate single-kind relationships.
- AC-003-5: Given maintainers use relationship fixture data, when traversal validation observes relationship facts, then the fixture does not define final reference expansion, implementation selection, related-symbol behavior, duplicate elimination, response ordering, or command JSON fields.

### Depends on:
Implementation ordering:
- Story document CAP-003-02-relationship-lookup-view.md - Relationship fixture expectations must align with owner lookup, target lookup, and edge-kind preservation.
- Story ST-001 - Provide Document and Symbol Fixture Coverage

Run time coupling:
- I-001-002 - Query planner traversal view

### Out of Scope
- Synthesizing inverse relationships that are absent from SCIP data.
- Deciding which edge kinds a command follows for final query results.
- Resolving relationship targets across multiple SCIP indexes.

### Assumptions
- **ASM-003-1**: Edge-kind coverage can be satisfied by either separate relationship facts or a multi-flag relationship fact when the official SCIP data exposes multiple flags together. - *Why*: CAP-003 requires preserving all present flags and does not require one physical relationship per edge kind. - Confidence: HIGH

### Open Questions
- None.
