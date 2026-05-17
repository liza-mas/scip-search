# User Stories: SCIP Binding Traversal Input

Status: review

## Goal
`scip-search` traversal receives the shared loaded-index boundary as official SCIP Go binding data and exposes index-level traversal input without taking over CLI loading or command behavior.

## Parent Epic
specs/epics/readme/20260517-100328-epic-planning-2.md - Capability CAP-001

## Context
The CLI runtime owns command routing, `--index` selection, path validation, and shared load failures. This document starts after that runtime boundary has accepted the caller-selected SCIP index and focuses on the traversal-facing input contract that query planners can reuse.

## Personas
- **Story Writer**: A downstream planning agent writing query implementation stories, needing a precise traversal boundary so command stories do not redefine SCIP parsing.
- **CLI Maintainer**: A Go developer maintaining `scip-search`, needing a small traversal contract that uses official SCIP schema types and avoids custom index formats.

## General information

Applies to: traversal input after the shared CLI runtime has successfully loaded the caller-selected SCIP index.

### References
- goal spec: README.md#language-support - Requires official SCIP bindings for parsing and traversal and says `scip-search` reads SCIP output directly.
- goal spec: README.md#what-is-scip-search - Defines one-shot loading of a caller-provided SCIP index before answering a query.
- parent epic: specs/epics/readme/20260517-100328-epic-planning-2.md#capability-cap-001---load-scip-index-data-for-query-traversal - Defines the traversal-ready view, official binding input, metadata, documents, symbols, external symbols, and occurrences.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/scip-index-loading-boundary.md - Confirms the shared runtime owns readable SCIP loading and invalid SCIP failure behavior.
- official source: https://pkg.go.dev/github.com/scip-code/scip/bindings/go/scip#IndexVisitor.ParseStreaming - Documents streaming parsing through the official Go SCIP binding.
- official source: https://pkg.go.dev/github.com/scip-code/scip/bindings/go/scip#Index - Documents index metadata, documents, and external symbols exposed by the official Go SCIP binding.

### Non-Functional Requirements
- NFR-000-1: Traversal input must use official SCIP Go binding data as the source of truth and must not introduce a custom index representation as the parsing source.
- NFR-000-2: Traversal input must remain reusable by all query planners and must not embed `symbols`, `packages`, `references`, or `implementations` command-specific matching or result-shaping behavior.
- NFR-000-3: Traversal input must remain process-local and must not introduce a daemon, watcher, incremental refresh, cache, or persisted derived index.

### Related External Components
- Component C-001 - Official SCIP Go bindings: Go package used by `scip-search` to access schema-defined index, document, symbol, and occurrence data.
- Component C-002 - Shared CLI runtime loaded-index boundary: Runtime-owned handoff that has already accepted the caller-selected SCIP index before traversal starts.

### Interfaces
- I-001-001 - SCIP traversal input (Interface 001 of Component C-001): Traversal receives official SCIP binding data from the shared loaded-index boundary and prepares it for query-planner use.
- I-001-002 - Runtime loaded-index handoff (Interface 002 of Component C-002): The CLI runtime supplies the successfully loaded index data and retains responsibility for command invocation, path validation, and shared load failures.

### Out of Scope
- CLI command routing, shared flags, `--index` parsing, index-path validation, stdout/stderr conventions, process status, and shared malformed-index errors.
- Query-specific symbol matching, package prefix filtering, reference selection, implementation selection, final JSON schemas, and command-specific ordering.
- Custom index formats, ctags fallback behavior, graph visualization, MCP server behavior, daemon behavior, file watching, incremental updates, and persistence beyond the current process.

### Assumptions
- **ASM-000-1**: The traversal input begins from the shared runtime loaded-index boundary, not from a raw file path. - *Why*: Epic CAP-001 excludes command invocation and path validation, while epic-planning-1 owns the caller-provided index path and shared loading boundary. - Confidence: HIGH
- **ASM-000-2**: "Official binding data" means the traversal contract remains traceable to official SCIP Go binding index, document, symbol, and occurrence fields, while exact internal wrapper names remain implementation-owned. - *Why*: The epic mandates official bindings but this story document should not prescribe private Go type names beyond the external boundary. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Accept the Runtime Loaded SCIP Boundary

### References
- parent epic: specs/epics/readme/20260517-100328-epic-planning-2.md#capability-cap-001---load-scip-index-data-for-query-traversal - Requires traversal to accept caller-selected SCIP index data without command-specific parsing.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/scip-index-loading-boundary.md - Defines the earlier runtime loading boundary.
- goal spec: README.md#language-support - Requires official SCIP bindings for parsing and traversal.

### User Story
**As a** CLI Maintainer implementing traversal, **I want to** receive the shared loaded SCIP index boundary as official SCIP binding data, **so that** query planners can consume one traversal entry point without owning protobuf parsing or file loading.

### Acceptance Criteria
- AC-001-1: Given the shared CLI runtime has successfully loaded the caller-selected SCIP index, when traversal starts, then traversal receives that loaded index as the only index input for the current process.
- AC-001-2: Given traversal receives loaded SCIP data, when a query planner asks for traversal access, then the planner can use the traversal boundary without reading the original index path or parsing protobuf bytes itself.
- AC-001-3: Given traversal starts from the loaded index boundary, when the selected command later plans a query, then command routing, flag parsing, path validation, malformed-index diagnostics, stdout, stderr, and process status remain governed by the shared runtime stories.
- AC-001-4: Given traversal receives official SCIP binding data, when it exposes the traversal boundary, then it does not require a custom index format, ctags fallback input, daemon, watcher, incremental refresh, or persisted derived index.

### Depends on:
Implementation ordering:
- Story document specs/stories/readme/20260517-095535-epic-planning-1/scip-index-loading-boundary.md - Traversal needs the runtime loaded-index handoff before this boundary can be implemented or tested.

Run time coupling:
- I-001-001 - SCIP traversal input
- I-001-002 - Runtime loaded-index handoff

### Out of Scope
- Opening or validating filesystem paths.
- Defining shared runtime errors or process-level output behavior.
- Implementing any query-specific filtering or JSON result schema.

### Assumptions
- **ASM-001-1**: Traversal may wrap official SCIP binding values for convenience, provided query planners can still trace exposed data to official SCIP schema fields. - *Why*: The capability requires official binding types and a reusable traversal-ready view, but does not require query planners to operate on raw protobuf structs directly. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-002 - Expose Index-Level Traversal Data

### References
- parent epic: specs/epics/readme/20260517-100328-epic-planning-2.md#capability-cap-001---load-scip-index-data-for-query-traversal - Requires index metadata, documents, external symbols, and occurrences to be available to planners.
- official source: https://pkg.go.dev/github.com/scip-code/scip/bindings/go/scip#Index - Documents index metadata, documents, and external symbols.
- goal spec: README.md#what-is-scip-search - Frames traversal as one-shot query support over a pre-built SCIP index.

### User Story
**As a** Story Writer planning query behavior, **I want to** rely on a traversal view that exposes index-level metadata, documents, and external symbols from the loaded SCIP index, **so that** query stories can describe planner behavior without reopening the SCIP schema.

### Acceptance Criteria
- AC-002-1: Given traversal receives a loaded SCIP index with metadata, when a query planner inspects the traversal view, then index-level metadata available through the official SCIP index binding is available for planner use.
- AC-002-2: Given traversal receives a loaded SCIP index with one or more documents, when a query planner requests the document inventory, then every SCIP document from the loaded index is represented once.
- AC-002-3: Given traversal receives a loaded SCIP index with external symbols, when a query planner requests external symbol inventory, then every external symbol exposed by the official SCIP index binding is available without scanning command-specific results.
- AC-002-4: Given traversal receives a loaded SCIP index, when a query planner uses index-level traversal data, then no query-specific filtering, result JSON shaping, or cross-process persistence is applied by this capability.
- AC-002-4b: Given the loaded SCIP index contains no documents or no external symbols, when the corresponding inventory is requested, then traversal exposes an empty inventory for that category instead of treating absence as a query failure.

### Depends on:
Implementation ordering:
- Story ST-001 - Accept the Runtime Loaded SCIP Boundary

Run time coupling:
- I-001-001 - SCIP traversal input

### Out of Scope
- Final names, fields, or ordering of CLI JSON results.
- Symbol-name partial matching, package-prefix filtering, reference lookup, implementation lookup, relationship traversal, range formatting, and hover rendering.
- Persisting metadata or inventories beyond the current process.

### Assumptions
- **ASM-002-1**: Empty document or external-symbol collections are valid traversal inventories when the official loaded index exposes them that way. - *Why*: CAP-001 asks traversal to expose inventory data; query-specific stories decide whether an empty inventory produces an empty result or a command-level outcome. - Confidence: HIGH

### Open Questions
- None.
