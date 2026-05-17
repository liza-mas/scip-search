# User Stories: Occurrence Lookup View by Symbol

Status: review

## Goal
`scip-search` traversal exposes reusable occurrence lookup by full SCIP symbol while preserving document context, source ranges, and SCIP symbol-role data for downstream query planners.

## Parent Epic
specs/epics/readme/20260517-100328-epic-planning-2.md - Capability CAP-003

## Context
CAP-001 exposes document and occurrence inventories, and CAP-002 preserves occurrence source-location and role metadata. Reference and implementation planners need to ask for occurrences associated with a full SCIP symbol without each query command re-walking raw documents or reinterpreting SCIP occurrence fields. This document defines the traversal lookup contract only; query-specific selection, grouping, missing-symbol behavior, duplicate policy, and JSON schemas remain with sibling query epics.

## Personas
- **Story Writer**: A downstream planning agent writing reference and implementation query stories, needing a stable occurrence lookup contract without reopening SCIP traversal details.
- **CLI Maintainer**: A Go developer maintaining `scip-search`, needing occurrence lookup behavior that stays aligned with official SCIP occurrence and symbol-role fields.

## General information

Applies to: process-local occurrence lookup views derived from the loaded official SCIP index for query-planner traversal.

### References
- goal spec: README.md#language-support - Requires official SCIP bindings for parsing and traversal.
- goal spec: README.md#what-is-scip-search - Documents `references --symbol` and `implementations --symbol` as symbol-based query commands.
- parent epic: specs/epics/readme/20260517-100328-epic-planning-2.md#capability-cap-003---expose-occurrence-and-relationship-lookup-views - Requires occurrence lookup views needed by reference, definition, implementation, and related-symbol planners.
- prior story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-001-02-document-and-symbol-inventory.md - Defines document and occurrence inventories that supply lookup input.
- prior story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-002-01-source-range-normalization.md - Defines preserved document paths, position encoding, ranges, enclosing ranges, and symbol-role bitsets.
- official source: https://raw.githubusercontent.com/scip-code/scip/main/scip.proto - Documents `Occurrence.symbol`, `Occurrence.range`, `Occurrence.enclosing_range`, and `SymbolRole` bitset values.
- consistency check: specs/stories/readme/20260517-100328-epic-planning-2/CAP-001-01-scip-binding-input.md - Confirms traversal starts from the shared loaded SCIP boundary.
- consistency check: specs/stories/readme/20260517-100328-epic-planning-2/CAP-002-02-hover-metadata-access.md - Confirms occurrence metadata preservation remains separate from final response fields.

### Non-Functional Requirements
- NFR-000-1: Occurrence lookup data must remain traceable to official SCIP occurrence fields and must not be derived from source-file reads or custom persisted index data.
- NFR-000-2: The lookup view must be deterministic over the same loaded index so downstream query stories can specify repeatable traversal behavior.
- NFR-000-3: The lookup view must not define command-specific reference selection, implementation selection, missing-symbol behavior, result grouping, duplicate elimination, or final CLI JSON schemas.

### Related External Components
- Component C-001 - Official SCIP Go bindings: Go package exposing document and occurrence data from the loaded SCIP index.
- Component C-002 - Query planner traversal view: The process-local traversal view consumed by query-specific planning and implementation stories.

### Interfaces
- I-001-002 - Query planner traversal view (Interface 002 of Component C-001): Query planners inspect reusable occurrence lookup data derived from official SCIP occurrence inventories.

### Out of Scope
- Final algorithms for `references --symbol`, `implementations --symbol`, definition lookup, or related-symbol expansion.
- Missing-symbol command behavior, result grouping, duplicate elimination policy, response ordering, and final JSON field names.
- Relationship lookup, synthesized relationships, cross-index resolution, source-file reads, daemon behavior, and persistence beyond the current process.

### Assumptions
- **ASM-000-1**: Occurrence lookup keys are full SCIP symbol strings as supplied by `Occurrence.symbol`, not partial names or package prefixes. - *Why*: CAP-003 is for query planners after symbol-based commands have a SCIP symbol, while sibling epics own partial symbol matching and query semantics. - Confidence: HIGH
- **ASM-000-2**: Empty lookup results are valid traversal facts, but deciding whether an empty result means a missing symbol, no references, or another command-level outcome belongs to query-specific stories. - *Why*: CAP-003 excludes missing-symbol behavior while still requiring reusable lookup views. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Retrieve Occurrences for a SCIP Symbol

### References
- parent epic: specs/epics/readme/20260517-100328-epic-planning-2.md#capability-cap-003---expose-occurrence-and-relationship-lookup-views - Requires occurrences by symbol for downstream query planners.
- prior story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-001-02-document-and-symbol-inventory.md - Defines occurrence inventories and containing document context.
- official source: https://raw.githubusercontent.com/scip-code/scip/main/scip.proto - Documents `Occurrence.symbol`.

### User Story
**As a** Story Writer planning symbol-based query behavior, **I want to** rely on traversal exposing occurrences indexed by full SCIP symbol, **so that** reference and implementation stories can describe matched occurrence behavior without requiring each query to scan raw documents.

### Acceptance Criteria
- AC-001-1: Given traversal has a loaded SCIP index containing occurrences for a full SCIP symbol across one or more documents, when a query planner requests occurrences for that exact symbol, then every indexed occurrence whose `Occurrence.symbol` equals that symbol is available from the lookup view.
- AC-001-2: Given matching occurrences come from different SCIP documents, when the planner inspects the lookup result, then each occurrence remains associated with its containing document identity and relative path from the traversal document inventory.
- AC-001-3: Given the loaded index also contains occurrences for other symbols, when the planner requests occurrences for one exact symbol, then occurrences for other symbols remain distinguishable and are not included as matches for that lookup key.
- AC-001-4: Given no occurrence in the loaded index has the requested exact symbol, when the planner requests occurrences for that symbol, then traversal exposes an empty occurrence lookup result without deciding command-level missing-symbol behavior or failure semantics.
- AC-001-5: Given a planner uses occurrence lookup by symbol, when traversal serves lookup data, then traversal does not apply partial-name matching, package-prefix filtering, relationship expansion, final JSON shaping, result grouping, duplicate elimination, or cross-index lookup.

### Depends on:
Implementation ordering:
- Story document CAP-001-02-document-and-symbol-inventory.md - Occurrence lookup requires the shared document and occurrence inventories.
- Story document CAP-002-01-source-range-normalization.md - Lookup results need preserved document paths and source-location metadata.

Run time coupling:
- I-001-002 - Query planner traversal view

### Out of Scope
- Deciding whether definition, reference, implementation, or related-symbol query results should include a returned occurrence.
- Resolving partial symbols or package names to full SCIP symbols.
- Defining final command response fields or ordering.

### Assumptions
- **ASM-001-1**: Lookup preserves all matching occurrences for a symbol even when query-specific stories later choose to filter some roles out of a command result. - *Why*: CAP-003 supplies reusable traversal facts, while sibling query epics own final selection. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-002 - Preserve Symbol Role Data in Occurrence Lookup

### References
- parent epic: specs/epics/readme/20260517-100328-epic-planning-2.md#capability-cap-003---expose-occurrence-and-relationship-lookup-views - Requires definition and non-definition occurrence role data, plus import, read, write, test, and generated role bits.
- prior story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-002-01-source-range-normalization.md - Defines preserved occurrence range, enclosing range, and role metadata.
- official source: https://raw.githubusercontent.com/scip-code/scip/main/scip.proto - Documents `SymbolRole.Definition`, `Import`, `WriteAccess`, `ReadAccess`, `Generated`, `Test`, and `ForwardDefinition`.

### User Story
**As a** CLI Maintainer implementing shared traversal, **I want to** preserve SCIP symbol-role bitset data on occurrences returned from symbol lookup, **so that** query planners can distinguish definitions, non-definitions, imports, reads, writes, generated code, and test code without reinterpreting raw SCIP documents.

### Acceptance Criteria
- AC-002-1: Given a matched occurrence has the SCIP `Definition` role bit, when a planner inspects the occurrence through symbol lookup, then the definition role data remains available for downstream interpretation.
- AC-002-2: Given a matched occurrence does not have the SCIP `Definition` role bit, when a planner inspects the occurrence through symbol lookup, then the absence of the definition role remains distinguishable from a definition occurrence.
- AC-002-3: Given a matched occurrence has any of the SCIP `Import`, `WriteAccess`, `ReadAccess`, `Generated`, or `Test` role bits, when a planner inspects the occurrence through symbol lookup, then each present role bit remains available with that occurrence.
- AC-002-3b: Given a matched occurrence has multiple SCIP role bits set, when traversal exposes that occurrence from symbol lookup, then the role bits remain available together instead of being collapsed into a single role label.
- AC-002-4: Given the official SCIP data exposes additional role bits such as `ForwardDefinition`, when traversal exposes occurrence role data, then traversal preserves the role bitset for planner use without assigning command-specific meaning to that bit.
- AC-002-5: Given a planner uses occurrence role data from the lookup view, when traversal serves that data, then traversal does not decide which roles count as references, definitions, implementations, generated-code exclusions, test-code exclusions, or final command output fields.

### Depends on:
Implementation ordering:
- Story ST-001 - Retrieve Occurrences for a SCIP Symbol
- Story document CAP-002-01-source-range-normalization.md - Role preservation is defined with occurrence source-location metadata.

Run time coupling:
- I-001-002 - Query planner traversal view

### Out of Scope
- Command-specific interpretation of occurrence role combinations.
- Filtering generated or test occurrences from any command result.
- Formatting role data for final CLI JSON output.

### Assumptions
- **ASM-002-1**: Traversal preserves the symbol-role bitset rather than converting it to command-specific categories. - *Why*: CAP-003 requires role data for planners, while sibling query epics decide which roles their commands include. - Confidence: HIGH

### Open Questions
- None.
