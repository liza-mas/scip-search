# User Stories: Document Symbol and Occurrence Inventory

Status: review

## Goal
`scip-search` traversal exposes reusable document, document-symbol, external-symbol, and occurrence inventories from official SCIP data so query planners can inspect indexed content without command-specific behavior.

## Parent Epic
specs/epics/readme/20260517-100328-epic-planning-2.md - Capability CAP-001

## Context
After traversal accepts the loaded SCIP index boundary, query planners need a document-centered inventory that preserves the SCIP data needed by later symbol, package, reference, and implementation stories. This document defines the inventory shape at the story level while leaving query filtering and final JSON schemas to sibling epics.

## Personas
- **Story Writer**: A downstream planning agent writing query implementation stories, needing a stable inventory contract for documents, symbols, and occurrences.
- **CLI Maintainer**: A Go developer maintaining `scip-search`, needing traversal expectations that stay aligned with official SCIP document, symbol, and occurrence fields.

## General information

Applies to: process-local inventories derived from the loaded official SCIP index for query-planner traversal.

### References
- goal spec: README.md#language-support - Requires official SCIP bindings for parsing and traversal.
- parent epic: specs/epics/readme/20260517-100328-epic-planning-2.md#capability-cap-001---load-scip-index-data-for-query-traversal - Requires document identity, language, relative paths, document-level symbols, external symbols, and document occurrences.
- task scope: epic-planning-2-us-writing-0 - Includes official SCIP binding input, index metadata, document inventory, document symbols, external symbols, and occurrence inventory; excludes query-specific filtering, final JSON schemas, and persistence beyond the current process.
- official source: https://pkg.go.dev/github.com/scip-code/scip/bindings/go/scip#Document - Documents SCIP document fields.
- official source: https://pkg.go.dev/github.com/scip-code/scip/bindings/go/scip#Occurrence - Documents SCIP occurrence fields.
- official source: https://pkg.go.dev/github.com/scip-code/scip/bindings/go/scip#SymbolInformation - Documents SCIP symbol information fields and relationships.

### Non-Functional Requirements
- NFR-000-1: Inventories must preserve official SCIP document, symbol, external symbol, and occurrence identities without requiring query planners to parse protobuf data.
- NFR-000-2: Inventories must be deterministic over the same loaded index so downstream query stories can specify repeatable behavior.
- NFR-000-3: Inventories must not define command-specific matching semantics, final JSON schemas, or persistence beyond the current process.

### Related External Components
- Component C-001 - Official SCIP Go bindings: Go package exposing document, occurrence, and symbol information fields from loaded SCIP data.
- Component C-002 - Query planner traversal view: The process-local view consumed by query-specific planning and implementation stories.

### Interfaces
- I-001-002 - Query planner traversal view (Interface 002 of Component C-001): Query planners inspect reusable document, symbol, external symbol, and occurrence inventories derived from official SCIP data.

### Out of Scope
- CLI command behavior, shared index-path validation, shared load failures, stdout/stderr, and process status.
- Query-specific filtering for `symbols`, `packages`, `references`, or `implementations`.
- Detailed range conversion, hover rendering, symbol-role interpretation, and relationship lookup beyond preserving occurrence and symbol inventories.
- Final JSON field names, response ordering, and persistence beyond the current process.

### Assumptions
- **ASM-000-1**: A deterministic inventory means repeatable traversal enumeration over the same loaded index; exact CLI result ordering remains owned by query-specific stories. - *Why*: CAP-001 requires planners to receive reusable inventories, while this task excludes final query result schemas and command-specific ordering. - Confidence: HIGH
- **ASM-000-2**: Document-level symbols and external symbols are both exposed because query planners may need to distinguish symbols declared in documents from symbols referenced from outside the indexed documents. - *Why*: CAP-001 names both document-level symbols and external symbols as in-scope data. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Expose Document Inventory

### References
- parent epic: specs/epics/readme/20260517-100328-epic-planning-2.md#capability-cap-001---load-scip-index-data-for-query-traversal - Requires document identity, language, relative paths, and occurrences.
- official source: https://pkg.go.dev/github.com/scip-code/scip/bindings/go/scip#Document - Documents SCIP document fields.
- goal spec: README.md#what-is-scip-search - Establishes query answers over the caller-selected index.

### User Story
**As a** Story Writer planning document-aware queries, **I want to** rely on a traversal document inventory containing each SCIP document's identity, language, relative path, and occurrence collection, **so that** query stories can describe results by indexed document without re-walking raw SCIP documents.

### Acceptance Criteria
- AC-001-1: Given traversal receives a loaded SCIP index containing multiple documents, when a query planner requests the document inventory, then every SCIP document is available with its document identity, language, and relative path as exposed by official SCIP document data.
- AC-001-2: Given a SCIP document has occurrences, when that document is represented in the inventory, then its occurrence collection is available to planners from the document entry.
- AC-001-2b: Given a SCIP document has no occurrences, when that document is represented in the inventory, then the document still appears with an empty occurrence collection.
- AC-001-3: Given multiple documents have distinct relative paths or languages, when the inventory is inspected, then those document attributes remain distinguishable for query-planner use.
- AC-001-4: Given a query planner uses the document inventory, when it inspects document data, then traversal does not apply symbol-name filtering, package-prefix filtering, reference selection, implementation selection, or final JSON result shaping.

### Depends on:
Implementation ordering:
- Story document CAP-001-01-scip-binding-input.md - Traversal must first accept the loaded SCIP binding input.

Run time coupling:
- I-001-002 - Query planner traversal view

### Out of Scope
- Formatting source ranges or hover documentation for result output.
- Deciding which occurrence roles count as definitions, references, or implementations.
- Hiding, merging, or filtering documents for a specific command.

### Assumptions
- **ASM-001-1**: The document inventory keeps documents addressable by official SCIP document identity and relative path, but this story does not require a specific lookup API shape. - *Why*: The source requires document identity and relative paths, while implementation details are not part of the story contract. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-002 - Expose Symbol and Occurrence Inventories

### References
- parent epic: specs/epics/readme/20260517-100328-epic-planning-2.md#capability-cap-001---load-scip-index-data-for-query-traversal - Requires document-level symbols, external symbols, and document occurrences.
- official source: https://pkg.go.dev/github.com/scip-code/scip/bindings/go/scip#SymbolInformation - Documents SCIP symbol information fields.
- official source: https://pkg.go.dev/github.com/scip-code/scip/bindings/go/scip#Occurrence - Documents SCIP occurrence fields.

### User Story
**As a** CLI Maintainer implementing shared traversal, **I want to** expose document symbols, external symbols, and occurrence inventories from the official SCIP data, **so that** query planners can build symbol, package, reference, and implementation behavior from one common source of indexed facts.

### Acceptance Criteria
- AC-002-1: Given a loaded SCIP index contains document-level symbols, when a query planner requests symbol inventory for a document, then every symbol information entry declared on that document is available from the traversal view.
- AC-002-2: Given a loaded SCIP index contains external symbols, when a query planner requests external symbol inventory, then every external symbol information entry exposed at index level is available from the traversal view.
- AC-002-3: Given a document occurrence references a SCIP symbol string, when a query planner inspects occurrence inventory, then the occurrence remains associated with that symbol string and its containing document.
- AC-002-4: Given symbol or occurrence inventories are exposed, when query planners consume them, then traversal preserves official SCIP symbol strings without resolving partial names, package prefixes, reference semantics, implementation semantics, or final JSON fields.
- AC-002-4b: Given the loaded index contains no document symbols, no external symbols, or no occurrences, when the corresponding inventory is requested, then traversal exposes an empty inventory for that category instead of inventing placeholder symbols or occurrences.

### Depends on:
Implementation ordering:
- Story document CAP-001-01-scip-binding-input.md - Traversal must first accept the loaded SCIP binding input.
- Story ST-001 - Expose Document Inventory

Run time coupling:
- I-001-002 - Query planner traversal view

### Out of Scope
- Partial symbol-name matching for `symbols --name`.
- Package identity extraction and `packages --prefix` filtering.
- Reference occurrence selection, implementation relationship selection, relationship graph traversal, source-range formatting, hover documentation rendering, and final query JSON schemas.

### Assumptions
- **ASM-002-1**: Occurrence inventory in CAP-001 preserves each occurrence's document and symbol association, while role, range, hover, and relationship interpretation remain outside this assigned scope. - *Why*: CAP-001 names occurrence inventory as in scope, while this task is limited to official binding input, index metadata, document inventory, document symbols, external symbols, and occurrence inventory. - Confidence: HIGH

### Open Questions
- None.
