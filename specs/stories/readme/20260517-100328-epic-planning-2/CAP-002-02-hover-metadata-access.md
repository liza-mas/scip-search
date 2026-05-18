# User Stories: SCIP Hover Metadata Access

Status: review

## Goal
`scip-search` traversal preserves SCIP symbol metadata and occurrence override documentation so downstream query planners can attach hover-related facts to matched symbols and occurrences without defining final CLI response fields.

## Parent Epic
specs/epics/readme/20260517-100328-epic-planning-2.md - Capability CAP-002

## Context
CAP-001 gives traversal consumers symbol and occurrence inventories. CAP-002 also requires hover-related metadata for those same facts: symbol kind, display name, documentation, signature documentation, enclosing symbol metadata, and occurrence-level override documentation. This document keeps the contract at traversal level so query-specific stories can decide how, or whether, those values appear in command responses.

## Personas
- **Story Writer**: A downstream planning agent writing query implementation stories, needing hover metadata guarantees for symbols and matched occurrences without choosing CLI JSON fields.
- **CLI Maintainer**: A Go developer maintaining traversal internals, needing a bounded metadata contract that stays aligned with official SCIP symbol and occurrence fields.

## General information

Applies to: hover-related symbol and occurrence metadata preserved on the CAP-001 traversal view.

### References
- goal spec: README.md#language-support - Requires official SCIP bindings for parsing and traversal.
- parent epic: specs/epics/readme/20260517-100328-epic-planning-2.md#capability-cap-002---preserve-source-locations-and-hover-metadata - Requires symbol kind, display name, documentation, signature documentation, and occurrence override documentation to remain available.
- prior story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-001-02-document-and-symbol-inventory.md - Defines the symbol and occurrence inventories that this document enriches with hover metadata.
- official source: https://pkg.go.dev/github.com/scip-code/scip/bindings/go/scip#SymbolInformation - Documents symbol documentation, kind, display name, signature documentation, and enclosing symbol fields.
- official source: https://raw.githubusercontent.com/scip-code/scip/main/scip.proto - Documents occurrence override documentation.

### Non-Functional Requirements
- NFR-000-1: Hover metadata must remain traceable to official SCIP symbol and occurrence schema fields.
- NFR-000-2: Traversal must preserve hover metadata for planner use without rendering Markdown, formatting signatures, or selecting final CLI JSON field names.
- NFR-000-3: Hover metadata access must not introduce query-specific filtering, source-file reads, cross-process persistence, daemon behavior, or a custom index format.

### Related External Components
- Component C-001 - Official SCIP Go bindings: Go package exposing symbol and occurrence metadata from loaded SCIP data.
- Component C-002 - Query planner traversal view: The process-local traversal view consumed by query-specific planning and implementation stories.

### Interfaces
- I-001-002 - Query planner traversal view (Interface 002 of Component C-001): Query planners inspect reusable symbol and occurrence metadata derived from official SCIP data.

### Out of Scope
- Final CLI JSON field names, ordering, pretty-printing, Markdown rendering, hover popover rendering, and human display formatting.
- Choosing whether command results include symbol metadata, occurrence override documentation, or signature documentation.
- Query-specific matching for `symbols`, `packages`, `references`, or `implementations`.
- Relationship lookup behavior beyond leaving relationships to the sibling CAP-003 scope.
- Reading source files from disk to supplement absent symbol or occurrence documentation.

### Assumptions
- **ASM-000-1**: "Hover metadata" means traversal exposes the documentation-oriented SCIP fields that query planners may use later; it does not require traversal to render a hover UI or choose command response fields. - *Why*: CAP-002 requires access to metadata but excludes final CLI JSON fields and pretty-printing. - Confidence: HIGH
- **ASM-000-2**: Symbol metadata may come from document-level symbols or external symbols because CAP-001 exposes both inventories and CAP-002 does not narrow hover metadata to only local document symbols. - *Why*: SCIP external symbols may carry hover documentation for referenced symbols outside the indexed documents. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Expose Symbol Hover Metadata

### References
- parent epic: specs/epics/readme/20260517-100328-epic-planning-2.md#capability-cap-002---preserve-source-locations-and-hover-metadata - Requires symbol kind, display name, documentation, and signature documentation.
- prior story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-001-02-document-and-symbol-inventory.md - Defines document-level and external symbol inventories.
- official source: https://pkg.go.dev/github.com/scip-code/scip/bindings/go/scip#SymbolInformation - Documents `documentation`, `kind`, `display_name`, `signature_documentation`, and `enclosing_symbol`.

### User Story
**As a** Story Writer planning symbol-aware query behavior, **I want to** rely on traversal exposing each indexed symbol's hover-related metadata, **so that** query stories can decide result behavior without reparsing SCIP symbol information.

### Acceptance Criteria
- AC-001-1: Given traversal exposes a document-level symbol, when a query planner inspects that symbol, then the symbol kind and display name available from official SCIP symbol information remain available for planner use.
- AC-001-2: Given traversal exposes an external symbol, when a query planner inspects that symbol, then the symbol kind and display name available from official SCIP symbol information remain available for planner use.
- AC-001-3: Given a SCIP symbol has documentation entries, when a query planner inspects that symbol through traversal, then the documentation entries remain available with the symbol.
- AC-001-3b: Given a SCIP symbol has no documentation entries, when the symbol is exposed, then traversal preserves the absence or empty collection instead of inventing documentation from source files or symbol names.
- AC-001-4: Given a SCIP symbol has signature documentation, when a query planner inspects that symbol through traversal, then the signature documentation remains available with the symbol.
- AC-001-5: Given a SCIP symbol has enclosing symbol metadata, when a query planner inspects that symbol through traversal, then the enclosing symbol value remains available for planner use.
- AC-001-6: Given query planners inspect symbol hover metadata, when traversal serves that metadata, then traversal does not render Markdown, format signature text for display, decide command JSON fields, or apply query-specific symbol matching.

### Depends on:
Implementation ordering:
- Story document CAP-001-02-document-and-symbol-inventory.md - Hover metadata access depends on document-level and external symbol inventories existing first.

Run time coupling:
- I-001-002 - Query planner traversal view

### Out of Scope
- Resolving partial symbol names.
- Extracting package identities or filtering package prefixes.
- Rendering documentation, choosing final response fields, or deciding whether a command includes hover metadata.

### Assumptions
- **ASM-001-1**: Traversal preserves signature documentation as the official SCIP binding exposes it, including its language and text content when present, without flattening it into plain display text. - *Why*: The official binding models signature documentation as structured SCIP data, while CAP-002 excludes pretty-printing. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-002 - Preserve Occurrence Override Documentation

### References
- parent epic: specs/epics/readme/20260517-100328-epic-planning-2.md#capability-cap-002---preserve-source-locations-and-hover-metadata - Requires occurrence override documentation for matched SCIP occurrences.
- official source: https://raw.githubusercontent.com/scip-code/scip/main/scip.proto - Defines occurrence `override_documentation`.
- prior story document: specs/stories/readme/20260517-100328-epic-planning-2/CAP-001-02-document-and-symbol-inventory.md - Defines the occurrence inventory that this story enriches with occurrence-level documentation.

### User Story
**As a** CLI Maintainer implementing shared traversal, **I want to** preserve occurrence-level override documentation with the occurrence that supplied it, **so that** query planners can distinguish range-specific hover documentation from symbol-level documentation.

### Acceptance Criteria
- AC-002-1: Given a SCIP occurrence has override documentation, when a query planner inspects that occurrence through traversal, then the override documentation entries remain available with that occurrence.
- AC-002-2: Given a SCIP occurrence has override documentation and a referenced symbol also has symbol-level documentation, when traversal exposes both facts, then the occurrence override documentation remains distinguishable from the symbol-level documentation.
- AC-002-2b: Given a SCIP occurrence has no override documentation, when the occurrence is exposed, then traversal preserves the absence or empty collection instead of copying symbol documentation onto the occurrence.
- AC-002-3: Given occurrence override documentation is available, when a query planner inspects a matched occurrence, then traversal preserves the documentation without deciding whether it replaces, augments, or appears alongside symbol-level documentation in any command response.
- AC-002-4: Given query planners inspect occurrence hover metadata, when traversal serves that metadata, then traversal does not render Markdown, read source files, decide final JSON fields, select references, or select implementations.

### Depends on:
Implementation ordering:
- Story document CAP-001-02-document-and-symbol-inventory.md - Occurrence override documentation depends on the shared occurrence inventory.
- Story document CAP-002-01-source-range-normalization.md - Matched occurrence metadata depends on occurrence source-location preservation for downstream query planners.

Run time coupling:
- I-001-002 - Query planner traversal view

### Out of Scope
- Deciding command-specific precedence between occurrence override documentation and symbol-level documentation.
- Rendering documentation for humans.
- Relationship lookup, reference filtering, implementation filtering, and final response schemas.

### Assumptions
- **ASM-002-1**: Traversal keeps occurrence override documentation tied to the occurrence rather than merging it into symbol metadata. - *Why*: The SCIP schema defines override documentation on the occurrence, and CAP-002 distinguishes occurrence override documentation from symbol documentation. - Confidence: HIGH

### Open Questions
- None.
