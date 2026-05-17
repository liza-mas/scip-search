# User Stories: Relationship Lookup View by Symbol

Status: review

## Goal
`scip-search` traversal exposes reusable SCIP relationship lookup by symbol while preserving source symbol, target symbol, and schema-defined relationship edge kinds for downstream reference, definition, implementation, and related-symbol planners.

## Parent Epic
specs/epics/readme/20260517-100328-epic-planning-2.md - Capability CAP-003

## Context
CAP-001 exposes document-level and external symbol information inventories. SCIP symbol information can include relationships to other symbols that affect find-references, find-implementations, type-definition, and definition behavior. Query planners need reusable access to those relationship facts without owning SCIP symbol inventory traversal or deciding final query semantics in the traversal layer.

## Personas
- **Story Writer**: A downstream planning agent writing reference and implementation query stories, needing relationship lookup guarantees without encoding SCIP schema details in every query story.
- **CLI Maintainer**: A Go developer maintaining `scip-search`, needing relationship lookup behavior that stays faithful to official SCIP relationship fields and remains separate from command-specific algorithms.

## General information

Applies to: process-local relationship lookup views derived from document-level and external SCIP symbol information in the loaded index.

### References
- goal spec: README.md#language-support - Requires official SCIP bindings for parsing and traversal.
- goal spec: README.md#what-is-scip-search - Documents `references --symbol` and `implementations --symbol` as symbol-based query commands.
- parent epic: specs/epics/readme/20260517-100328-epic-planning-2.md#capability-cap-003---expose-occurrence-and-relationship-lookup-views - Requires relationship lookup views needed by reference, definition, implementation, and related-symbol planners.
- prior story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-001-02-document-and-symbol-inventory.md - Defines document-level and external symbol inventories that supply relationship data.
- official source: https://raw.githubusercontent.com/scip-code/scip/main/scip.proto - Documents `SymbolInformation.relationships` and `Relationship.is_reference`, `is_implementation`, `is_type_definition`, and `is_definition`.
- official source: https://pkg.go.dev/github.com/scip-code/scip/bindings/go/scip#Relationship - Documents relationship getters and helper behavior in the official Go binding.
- consistency check: specs/stories/readme/20260517-100328-epic-planning-2/CAP-001-01-scip-binding-input.md - Confirms traversal starts from official SCIP binding data after the shared runtime load boundary.

### Non-Functional Requirements
- NFR-000-1: Relationship lookup data must remain traceable to official SCIP `SymbolInformation.relationships` and `Relationship` fields.
- NFR-000-2: Relationship lookup views must be deterministic over the same loaded index so downstream query stories can specify repeatable traversal behavior.
- NFR-000-3: Relationship lookup views must not synthesize relationships that are absent from SCIP data or define query-specific reference, definition, implementation, grouping, duplicate, or JSON result behavior.

### Related External Components
- Component C-001 - Official SCIP Go bindings: Go package exposing symbol information and relationship data from the loaded SCIP index.
- Component C-002 - Query planner traversal view: The process-local traversal view consumed by query-specific planning and implementation stories.

### Interfaces
- I-001-002 - Query planner traversal view (Interface 002 of Component C-001): Query planners inspect reusable relationship lookup data derived from official SCIP symbol information inventories.

### Out of Scope
- Final algorithms for `references --symbol`, `implementations --symbol`, go-to-definition, go-to-type-definition, or related-symbol expansion.
- Missing-symbol command behavior, result grouping, duplicate elimination policy, response ordering, and final JSON field names.
- Synthesizing inverse relationships not represented by the lookup data, cross-index relationship resolution, daemon behavior, and persistence beyond the current process.

### Assumptions
- **ASM-000-1**: A relationship edge has a source symbol from the owning `SymbolInformation.symbol` and a target symbol from `Relationship.symbol`. - *Why*: SCIP stores relationships on symbol information entries and the relationship message names the related symbol. - Confidence: HIGH
- **ASM-000-2**: Traversal may provide lookup access for relationships by either source symbol or target symbol, provided it preserves original edge direction and flags. - *Why*: Reference and implementation planners may need relationships owned by the queried symbol or relationships that point at it, while CAP-003 leaves command-specific expansion rules out of scope. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Retrieve Relationships Owned by a SCIP Symbol

### References
- parent epic: specs/epics/readme/20260517-100328-epic-planning-2.md#capability-cap-003---expose-occurrence-and-relationship-lookup-views - Requires reusable relationship lookup views for query planners.
- prior story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-001-02-document-and-symbol-inventory.md - Defines document-level and external symbol inventories containing symbol information.
- official source: https://raw.githubusercontent.com/scip-code/scip/main/scip.proto - Documents `SymbolInformation.relationships` and `Relationship.symbol`.

### User Story
**As a** Story Writer planning relationship-aware query behavior, **I want to** rely on traversal exposing relationships owned by a full SCIP symbol, **so that** reference, definition, implementation, and related-symbol stories can describe planner behavior without scanning raw symbol information entries.

### Acceptance Criteria
- AC-001-1: Given traversal has a loaded SCIP index with document-level or external symbol information that owns relationships, when a query planner requests relationships for that exact owner symbol, then every relationship attached to that owner symbol is available from the lookup view.
- AC-001-2: Given a relationship is returned for an owner symbol, when the planner inspects the relationship, then the owner source symbol and the relationship target symbol remain distinguishable.
- AC-001-3: Given relationships exist for different owner symbols, when the planner requests relationships for one exact owner symbol, then relationships owned by other symbols remain distinguishable and are not included as owned relationships for that lookup key.
- AC-001-4: Given an owner symbol has no relationships in the loaded SCIP data, when the planner requests owned relationships for that symbol, then traversal exposes an empty relationship lookup result without deciding command-level missing-symbol behavior or failure semantics.
- AC-001-5: Given a planner uses relationships owned by a symbol, when traversal serves lookup data, then traversal does not apply reference expansion, implementation selection, definition fallback, type-definition behavior, duplicate elimination, final JSON shaping, or cross-index lookup.

### Depends on:
Implementation ordering:
- Story document CAP-001-02-document-and-symbol-inventory.md - Relationship lookup requires document-level and external symbol inventories.

Run time coupling:
- I-001-002 - Query planner traversal view

### Out of Scope
- Deciding which relationship edges a command follows.
- Synthesizing relationships for symbols that do not declare them.
- Defining final command response fields or ordering.

### Assumptions
- **ASM-001-1**: Relationships from document-level symbols and external symbols are both eligible for lookup when the official loaded index exposes them. - *Why*: CAP-001 exposes both symbol inventories, and CAP-003 does not narrow relationship lookup to only locally defined symbols. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-002 - Retrieve Relationships Targeting a SCIP Symbol

### References
- parent epic: specs/epics/readme/20260517-100328-epic-planning-2.md#capability-cap-003---expose-occurrence-and-relationship-lookup-views - Requires relationship lookup views needed by implementation and related-symbol planners.
- official source: https://raw.githubusercontent.com/scip-code/scip/main/scip.proto - Documents relationship targets through `Relationship.symbol` and relationship edge kinds.
- parent epic: specs/epics/readme/20260517-100328-epic-planning-2.md#capability-cap-003---expose-occurrence-and-relationship-lookup-views - Leaves final query expansion semantics to sibling query epics.

### User Story
**As a** CLI Maintainer implementing shared traversal, **I want to** expose relationships that target a full SCIP symbol while preserving their original owner symbols, **so that** implementation and related-symbol planners can find incoming SCIP relationship facts without inventing relationships or walking every symbol entry themselves.

### Acceptance Criteria
- AC-002-1: Given traversal has relationship data where one or more relationships target a full SCIP symbol, when a query planner requests relationships targeting that exact symbol, then every relationship whose `Relationship.symbol` equals that symbol is available from the lookup view.
- AC-002-2: Given a relationship is returned from target lookup, when the planner inspects it, then the original owner source symbol remains available alongside the target symbol.
- AC-002-3: Given multiple owner symbols declare relationships that target the same symbol, when the planner requests relationships targeting that symbol, then the lookup view exposes each source-to-target relationship fact separately without merging owners together.
- AC-002-4: Given no relationship targets the requested exact symbol, when the planner requests targeting relationships for that symbol, then traversal exposes an empty relationship lookup result without deciding command-level missing-symbol behavior or failure semantics.
- AC-002-5: Given target lookup exposes incoming relationship facts, when traversal serves those facts, then traversal does not decide whether incoming edges produce implementation results, reference results, definition results, related-symbol expansion, result grouping, duplicate elimination, or final JSON fields.

### Depends on:
Implementation ordering:
- Story ST-001 - Retrieve Relationships Owned by a SCIP Symbol

Run time coupling:
- I-001-002 - Query planner traversal view

### Out of Scope
- Treating incoming relationships as final command results.
- Creating synthetic reverse edges that lose original owner, target, or flag information.
- Resolving relationships across multiple SCIP indexes.

### Assumptions
- **ASM-002-1**: Target lookup is an access pattern over the same SCIP relationship facts as owner lookup, not a separate relationship model. - *Why*: CAP-003 excludes synthesized relationships while query planners still need efficient access to relationships pointing at a queried symbol. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-003 - Preserve SCIP Relationship Edge Kinds

### References
- parent epic: specs/epics/readme/20260517-100328-epic-planning-2.md#capability-cap-003---expose-occurrence-and-relationship-lookup-views - Requires relationship edge kinds needed by reference and implementation planners.
- official source: https://raw.githubusercontent.com/scip-code/scip/main/scip.proto - Documents `Relationship.is_reference`, `is_implementation`, `is_type_definition`, and `is_definition`.
- official source: https://pkg.go.dev/github.com/scip-code/scip/bindings/go/scip#Relationship - Documents official Go binding accessors for relationship fields.

### User Story
**As a** Story Writer planning relationship-aware command behavior, **I want to** rely on traversal preserving SCIP relationship edge-kind flags, **so that** query stories can specify which schema-defined relationships their planners follow without traversal collapsing or reclassifying the edges.

### Acceptance Criteria
- AC-003-1: Given a relationship has `is_reference` set in SCIP data, when a planner inspects that relationship through owner or target lookup, then the reference edge kind remains available for downstream interpretation.
- AC-003-2: Given a relationship has `is_implementation` set in SCIP data, when a planner inspects that relationship through owner or target lookup, then the implementation edge kind remains available for downstream interpretation.
- AC-003-3: Given a relationship has `is_type_definition` set in SCIP data, when a planner inspects that relationship through owner or target lookup, then the type-definition edge kind remains available for downstream interpretation.
- AC-003-4: Given a relationship has `is_definition` set in SCIP data, when a planner inspects that relationship through owner or target lookup, then the definition edge kind remains available for downstream interpretation.
- AC-003-4b: Given a relationship has multiple edge-kind flags set, when traversal exposes that relationship, then all present flags remain available together instead of being collapsed into one preferred kind.
- AC-003-5: Given a planner uses relationship edge-kind flags, when traversal serves those flags, then traversal does not decide final reference expansion, implementation expansion, go-to-definition behavior, go-to-type-definition behavior, command result grouping, duplicate elimination, or JSON fields.

### Depends on:
Implementation ordering:
- Story ST-001 - Retrieve Relationships Owned by a SCIP Symbol
- Story ST-002 - Retrieve Relationships Targeting a SCIP Symbol

Run time coupling:
- I-001-002 - Query planner traversal view

### Out of Scope
- Assigning command-specific precedence when multiple relationship flags are present.
- Inferring relationship flags from occurrence roles or symbol names.
- Filtering relationship flags for final command output.

### Assumptions
- **ASM-003-1**: Relationship edge-kind flags are preserved exactly as SCIP data exposes them, while command stories decide which flags matter for a specific query. - *Why*: CAP-003 requires surfacing relationship edges exactly enough for planners to choose expansion rules. - Confidence: HIGH

### Open Questions
- None.
