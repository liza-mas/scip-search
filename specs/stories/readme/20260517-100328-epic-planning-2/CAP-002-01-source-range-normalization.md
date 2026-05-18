# User Stories: SCIP Source Location Preservation

Status: review

## Goal
`scip-search` traversal preserves SCIP document paths, occurrence ranges, enclosing ranges, position encodings, and symbol-role bitsets so downstream query planners can render location-bearing results without reparsing SCIP data.

## Parent Epic
specs/epics/readme/20260517-100328-epic-planning-2.md - Capability CAP-002

## Context
CAP-001 gives query planners process-local document and occurrence inventories from the loaded official SCIP index. This document narrows CAP-002 to source-location metadata on those existing inventories: document relative paths, occurrence ranges, enclosing ranges, document position encodings, and symbol-role bitsets. Query-specific stories still own which matches are selected and how command JSON names or formats those fields.

## Personas
- **Story Writer**: A downstream planning agent writing query implementation stories, needing precise source-location guarantees without reopening the SCIP schema.
- **CLI Maintainer**: A Go developer maintaining traversal internals, needing range and role preservation expectations that stay aligned with official SCIP semantics.

## General information

Applies to: source-location metadata preserved on the CAP-001 traversal view for documents and occurrences.

### References
- goal spec: README.md#language-support - Requires official SCIP bindings for parsing and traversal.
- parent epic: specs/epics/readme/20260517-100328-epic-planning-2.md#capability-cap-002---preserve-source-locations-and-hover-metadata - Requires document identity, occurrence range, enclosing range, symbol role bitsets, and range semantics to remain available to query planners.
- prior story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-001-01-scip-binding-input.md - Defines the traversal input boundary from the shared loaded SCIP index.
- prior story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-001-02-document-and-symbol-inventory.md - Defines document and occurrence inventories that this document enriches with source-location metadata.
- official source: https://raw.githubusercontent.com/scip-code/scip/main/scip.proto - Documents document relative paths, position encoding, occurrence range, symbol role bitsets, override documentation, and enclosing ranges.

### Non-Functional Requirements
- NFR-000-1: Source-location metadata must remain traceable to official SCIP document and occurrence schema fields.
- NFR-000-2: Traversal must preserve SCIP range semantics for query planners and must not convert them into editor-specific or command-specific coordinate conventions.
- NFR-000-3: Source-location preservation must not introduce source-file reads, cross-process persistence, daemon behavior, or query-specific filtering.

### Related External Components
- Component C-001 - Official SCIP Go bindings: Go package exposing document and occurrence fields from loaded SCIP data.
- Component C-002 - Query planner traversal view: The process-local traversal view consumed by query-specific planning and implementation stories.

### Interfaces
- I-001-002 - Query planner traversal view (Interface 002 of Component C-001): Query planners inspect reusable document, occurrence, range, and role metadata derived from official SCIP data.

### Out of Scope
- Final CLI JSON field names, ordering, pretty-printing, and human display formatting.
- Editor-specific coordinate conversion beyond preserving SCIP range values and the document position encoding needed to interpret them.
- Query-specific filtering for `symbols`, `packages`, `references`, or `implementations`.
- Reading source files from disk to supplement absent SCIP document text.
- Relationship lookup, implementation selection, hover rendering, diagnostics display, syntax highlighting display, and ctags fallback behavior.

### Assumptions
- **ASM-000-1**: "Preserve source ranges" means traversal exposes SCIP range values together with their document position encoding, not that traversal converts them to a different coordinate system. - *Why*: CAP-002 excludes editor-specific coordinate conversion and final command response formatting. - Confidence: HIGH
- **ASM-000-2**: Symbol role bitsets remain available as SCIP role data for downstream query planners, while deciding which roles count as definitions or references belongs to query-specific stories. - *Why*: CAP-002 requires preserving role bitsets, while sibling query epics own reference and implementation semantics. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Preserve Document Location Context

### References
- parent epic: specs/epics/readme/20260517-100328-epic-planning-2.md#capability-cap-002---preserve-source-locations-and-hover-metadata - Requires document identity and source-location data for matched SCIP symbols and occurrences.
- prior story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-001-02-document-and-symbol-inventory.md - Defines the document inventory that supplies document identity, language, relative path, and occurrence collections.
- official source: https://raw.githubusercontent.com/scip-code/scip/main/scip.proto - Defines document `relative_path` and `position_encoding`.

### User Story
**As a** Story Writer planning location-bearing query results, **I want to** rely on traversal occurrences retaining their containing document path and position encoding, **so that** query stories can specify source locations without redefining document context.

### Acceptance Criteria
- AC-001-1: Given traversal exposes a SCIP document from the loaded index, when a query planner inspects that document, then the document's SCIP relative path remains available with the document entry.
- AC-001-2: Given traversal exposes occurrences for a SCIP document, when a query planner inspects an occurrence, then the planner can determine the containing document for that occurrence and the containing document's relative path.
- AC-001-3: Given a SCIP document declares a position encoding, when a query planner inspects source-location metadata for occurrences in that document, then the document position encoding remains available for interpreting occurrence character offsets.
- AC-001-4: Given SCIP document text is absent or empty, when traversal exposes document location context, then traversal does not read source files from disk to supplement the missing document text.
- AC-001-5: Given a query planner uses document location context, when it prepares query-specific results, then traversal has not converted relative paths or positions into final CLI JSON fields, pretty-printed text, or editor-specific coordinates.

### Depends on:
Implementation ordering:
- Story document CAP-001-02-document-and-symbol-inventory.md - Source-location context depends on document and occurrence inventories existing first.

Run time coupling:
- I-001-002 - Query planner traversal view

### Out of Scope
- Resolving relative paths to absolute filesystem paths.
- Reading document text or source files to compute missing locations.
- Deciding final command field names or path display format.

### Assumptions
- **ASM-001-1**: Query planners need the document position encoding alongside occurrence ranges because SCIP character offsets are interpreted relative to that document-level encoding. - *Why*: The SCIP schema defines range character interpretation through the document position encoding. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-002 - Preserve Occurrence Ranges and Symbol Roles

### References
- parent epic: specs/epics/readme/20260517-100328-epic-planning-2.md#capability-cap-002---preserve-source-locations-and-hover-metadata - Requires occurrence ranges, enclosing ranges, and symbol-role bitsets to remain available.
- official source: https://raw.githubusercontent.com/scip-code/scip/main/scip.proto - Defines occurrence `range`, `symbol_roles`, and `enclosing_range` fields.
- prior story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-001-02-document-and-symbol-inventory.md - Defines the occurrence inventory that this story enriches with range and role metadata.

### User Story
**As a** CLI Maintainer implementing shared traversal, **I want to** preserve each occurrence's SCIP range, enclosing range, and symbol-role bitset, **so that** query planners can select and report matched occurrences using schema-defined location and role facts.

### Acceptance Criteria
- AC-002-1: Given a SCIP occurrence has a `range`, when a query planner inspects that occurrence through traversal, then the range value is available with the occurrence.
- AC-002-2: Given a SCIP occurrence range uses either the three-integer same-line form or the four-integer multi-position form, when traversal exposes the occurrence, then the planner can distinguish the preserved SCIP range values without traversal converting them to another coordinate convention.
- AC-002-3: Given a SCIP occurrence has an `enclosing_range`, when a query planner inspects that occurrence, then the enclosing range remains available separately from the occurrence range.
- AC-002-3b: Given a SCIP occurrence has no `enclosing_range`, when the occurrence is exposed, then traversal preserves that absence instead of inventing an enclosing range from the occurrence range.
- AC-002-4: Given a SCIP occurrence has `symbol_roles`, when a query planner inspects the occurrence, then the role bitset remains available as SCIP role data for downstream interpretation.
- AC-002-5: Given query planners inspect range or role metadata, when traversal serves that metadata, then traversal does not decide final JSON field names, result ordering, pretty-printing, editor coordinate conversion, reference filtering, or implementation filtering.

### Depends on:
Implementation ordering:
- Story ST-001 - Preserve Document Location Context

Run time coupling:
- I-001-002 - Query planner traversal view

### Out of Scope
- Interpreting which role combinations produce `symbols`, `references`, or `implementations` command results.
- Validating or repairing malformed SCIP range arrays beyond preserving loaded official SCIP data for planners.
- Displaying syntax highlighting, diagnostics, or hover popovers.

### Assumptions
- **ASM-002-1**: Absent optional occurrence range metadata remains distinguishable from traversal-invented defaults when the official SCIP binding exposes that distinction. - *Why*: Query planners need to know whether SCIP supplied an enclosing range instead of receiving a value manufactured by traversal. - Confidence: HIGH

### Open Questions
- None.
