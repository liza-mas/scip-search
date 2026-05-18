# Architecture Plan: SCIP Traversal Layer

Status: review

## Goal

Define the shared Go traversal layer that turns the runtime-loaded official SCIP index into process-local document, symbol, occurrence, relationship, range, and hover views for query planners without taking ownership of command behavior.

## Context

`scip-search` is planned as a thin Go CLI that loads one caller-selected SCIP index, answers one query, prints JSON, and exits. The CLI runtime architecture owns command routing, `--index` parsing, stream/status behavior, and the official binding load boundary. This plan starts after that boundary: traversal receives the loaded official SCIP data and prepares reusable in-memory views for later symbol, package, reference, and implementation query planners.

There is currently no committed Go module or traversal implementation in this worktree. Existing artifacts are product/specification documents, `.pre-commit-config.yaml`, `.editorconfig`, and architecture/story plans. The plan therefore defines package boundaries for downstream code-planners while preserving the sibling epic boundary: traversal supplies facts, query epics choose matching, filtering, result grouping, duplicate policy, and final JSON fields.

### References

- Goal spec: `README.md#language-support`, `README.md#what-is-scip-search`, `README.md#scip-symbol-format`, `README.md#out-of-scope`
- Parent tasks: `epic-planning-2-us-writing-0`, `epic-planning-2-us-writing-1`, `epic-planning-2-us-writing-2`, `epic-planning-2-us-writing-3`
- Parent epic: `specs/epics/readme/20260517-100328-epic-planning-2.md`
- Story specs: `specs/stories/readme/20260517-100328-epic-planning-2/*.md`
- Runtime boundary surveyed: `specs/arch-plan/readme/20260517-170052-epic-planning-1-architecture.md`
- Codebase surveyed: `.pre-commit-config.yaml`, `.editorconfig`, `README.md`, `specs/`

### Constraints

- Traversal must use official SCIP Go binding data as the source of truth and must not introduce a custom persisted index format.
- Traversal starts from the runtime loaded-index handoff owned by epic-planning-1, not from a raw path or command flag.
- Traversal must stay process-local: no daemon, watcher, incremental refresh, cache, cross-run persistence, generated index, compiler/type-checker invocation, MCP server, UI, vector store, or ctags fallback behavior.
- Traversal must not define query-specific matching, filtering, missing-symbol behavior, result grouping, duplicate elimination, response ordering, or final CLI JSON schemas.
- Source-location metadata must preserve SCIP range values and document position encodings rather than convert them to editor-specific coordinates.
- Hover metadata must remain available as schema facts and must not be rendered, flattened, or merged across symbol and occurrence boundaries.
- Fixture data must be deterministic, schema-valid SCIP data consumed through the same official binding path as other indexes and small enough for normal test runs.
- `.pre-commit-config.yaml` already exists in the worktree HEAD, so no `bootstrap-precommit` output entry is emitted.

### Assumptions

- **ASM-001**: The traversal package can wrap official SCIP binding values in query-planner-facing structs while keeping every exposed fact traceable to official SCIP fields. - *Why*: The stories allow convenience wrappers but require official binding data as the source of truth and prohibit command-specific custom formats. - Confidence: HIGH
- **ASM-002**: A single document-centered traversal view should own both inventory fields and metadata fields. - *Why*: Documents, occurrences, ranges, roles, hover metadata, and symbol references are one shared set of SCIP facts; splitting location and hover into a separate model would force downstream planners to join the same facts repeatedly. - Confidence: HIGH
- **ASM-003**: Synthetic fixtures are acceptable when encoded as schema-valid SCIP payloads and loaded through the same official binding path. - *Why*: CAP-004 excludes external indexer installation and large real-world fixtures while requiring deterministic SCIP coverage. - Confidence: HIGH

### Open Questions

- None.

## Components

### Runtime Loaded Index Boundary (`internal/scipindex/`)

**Responsibility:** Supply traversal with the official SCIP index data already accepted by the shared CLI runtime.

**Boundaries:**
- Exposes: a loaded invocation-scoped context containing official SCIP binding index data and any runtime-owned load metadata.
- Depends on: filesystem loading and official SCIP Go bindings, as defined by the CLI runtime architecture.

**Key decisions:**
- Traversal consumes this boundary instead of accepting paths or bytes: this preserves epic-planning-1 ownership of command invocation, index-path validation, malformed-index diagnostics, stream behavior, and exit status.
- The loaded context remains invocation-local: query planners cannot reuse traversal data across processes or assume global state.

### Traversal View (`internal/traversal/`)

**Responsibility:** Build and expose a reusable, process-local view over official SCIP documents, symbols, external symbols, occurrences, source-location metadata, hover metadata, and relationships.

**Boundaries:**
- Exposes: document inventory, symbol inventory, external symbol inventory, occurrence inventory, occurrence lookup by full SCIP symbol, relationship lookup by owner and target full SCIP symbols, and schema-preserved metadata access.
- Depends on: the runtime loaded index boundary and official SCIP binding types.

**Key decisions:**
- Keep traversal command-agnostic: it exposes facts and lookup access patterns, while query packages decide `symbols`, `packages`, `references`, and `implementations` semantics.
- Use exact full SCIP symbol strings as lookup keys: partial-name and package-prefix matching belong to query-specific epics.
- Preserve absence distinctly for optional SCIP fields such as enclosing ranges, override documentation, symbol documentation, and empty inventories: query planners need to know what SCIP supplied versus what traversal invented.
- Preserve role and relationship flags as bitsets/flags rather than command labels: reference, definition, implementation, generated, test, and other meanings remain downstream decisions.

### Traversal Facts (`internal/traversal/`)

**Responsibility:** Define the data records crossing from traversal into query planners.

**Boundaries:**
- Exposes: immutable or read-only fact records for documents, symbols, occurrences, ranges, hover metadata, and relationships.
- Depends on: official SCIP schema field meanings and the traversal view build process.

**Key decisions:**
- Document facts carry SCIP document identity, relative path, language, position encoding, document symbols, and occurrences.
- Occurrence facts carry containing document context, full symbol string, range, optional enclosing range, role bitset, and override documentation.
- Symbol facts carry full symbol string, local-vs-external source, kind, display name, documentation entries, signature documentation, enclosing symbol metadata, and relationships.
- Relationship facts carry original owner/source symbol, target symbol, and all SCIP edge-kind flags; target lookup is an access pattern over the same facts, not synthesized inverse data.

### Traversal Fixtures (`internal/traversal/testdata/` and `internal/traversal/traversaltest/`)

**Responsibility:** Provide deterministic schema-valid SCIP payloads and helper accessors for traversal tests.

**Boundaries:**
- Exposes: compact SCIP fixture data loaded through the official binding path and helper functions usable by traversal tests.
- Depends on: official SCIP Go bindings and the traversal view contract.

**Key decisions:**
- Store fixture payloads under traversal-owned test paths so query-specific epics can reference the shared baseline without turning it into a command JSON golden set.
- Include multiple documents and symbols in one compact fixture set where practical: CAP-004 requires category coverage, not one physical fixture per category.
- Keep fixture helper code out of production packages except through `_test.go` or explicitly test-only package boundaries, preserving the existing pre-commit test-helper constraint.

### Query Planner Boundary (`internal/query/` or command-specific internal packages)

**Responsibility:** Consume traversal facts and choose query-specific behavior in later epics.

**Boundaries:**
- Exposes: planner code for `symbols`, `packages`, `references`, and `implementations` in sibling query scopes.
- Depends on: traversal view contracts and the runtime query handler boundary.

**Key decisions:**
- Query planners receive traversal facts instead of raw protobuf documents when they implement command behavior.
- Query planners own final response fields, matching algorithms, result ordering, grouping, duplicate policy, generated/test filtering, and missing-symbol behavior.

## Interfaces

### Runtime Loaded Index Boundary -> Traversal View

**Contract:** The runtime supplies the already loaded official SCIP index data for the current invocation; traversal builds a process-local view from that data.
**Direction:** `internal/cli` or query handler setup calls `internal/traversal` with the loaded context from `internal/scipindex`.
**Invariants:** Traversal does not open paths, parse command flags, emit diagnostics, write stdout/stderr, or decide exit status.

### Traversal View -> Query Planner Boundary

**Contract:** Query planners can enumerate documents, local symbols, external symbols, and occurrences, and can inspect source-location and hover metadata associated with those facts.
**Direction:** Query-specific packages call traversal view methods or receive traversal facts from handler setup.
**Invariants:** Traversal exposes schema facts without applying command-specific filters, final JSON field names, sorting, grouping, or duplicate removal.

### Traversal View -> Occurrence Lookup

**Contract:** A full SCIP symbol string crosses as the lookup key; traversal returns all occurrences whose `Occurrence.symbol` exactly matches that key, each retaining document context, range, enclosing range absence/presence, role bitset, and override documentation.
**Direction:** Query planners call traversal lookup accessors.
**Invariants:** Empty lookup results remain successful traversal facts and do not imply command-level missing-symbol behavior.

### Traversal View -> Relationship Lookup

**Contract:** A full SCIP symbol string crosses as either an owner/source key or a target key; traversal returns relationship facts preserving original owner, target, direction, and every SCIP edge-kind flag.
**Direction:** Query planners call traversal relationship accessors.
**Invariants:** Target lookup does not synthesize inverse relationships or collapse multiple owners; command planners decide which edges matter.

### Traversal Fixtures -> Traversal View

**Contract:** Fixture payloads are loaded through the same official SCIP binding path as runtime-loaded indexes and then observed through traversal view APIs.
**Direction:** Traversal tests load fixture data, build the traversal view, and assert category coverage.
**Invariants:** Fixture validation does not assert final command JSON, stdout/stderr behavior, process status, or query-specific selection semantics.

## Data Flow

```text
caller-selected SCIP file
  -> internal/scipindex official binding loader
  -> loaded invocation context
  -> internal/traversal view builder
  -> document/symbol/occurrence/relationship facts
  -> query planners
  -> query-specific results owned by sibling epics
```

Traversal view build:

```text
official scip.Index
  -> document inventory with relative paths, languages, position encodings, occurrences
  -> symbol inventories from document-level and external symbol information
  -> occurrence facts keyed by exact full SCIP symbol
  -> relationship facts keyed by owner symbol and target symbol
  -> metadata access for ranges, roles, hover docs, signatures, and override docs
```

Fixture validation:

```text
schema-valid SCIP fixture payload
  -> same official binding loader path
  -> traversal view
  -> category assertions for documents, symbols, occurrences, ranges, roles, relationships, and hover metadata
```

## Cross-Cutting Concerns

| Concern | Approach |
|---------|----------|
| Error handling | Traversal accepts only successfully loaded official SCIP data. Build-time validation errors, if any, return to the caller as internal errors for the runtime/query layer to classify; traversal does not emit process diagnostics. Empty inventories and empty lookup results are data, not errors. |
| Observability | Do not add progress logs, prompts, or tracing output in traversal. Tests and coverage failures should name missing fixture categories clearly. |
| Configuration | Traversal has no configuration beyond the loaded index supplied for the current invocation. No default paths, environment lookup, cache directory, or persistent session state. |
| Testing | Unit-test traversal view construction and lookups through deterministic schema-valid fixtures. Use fixture coverage tests for required SCIP categories; query-specific JSON goldens remain with query epics. |
| Security | Treat SCIP data as caller-supplied input already accepted by the loader. Do not shell out to indexers, read source files to supplement metadata, mutate selected indexes, or include sensitive file contents in diagnostics. |
| Concurrency | All traversal state is invocation-local and immutable/read-only after build so concurrent CLI invocations and worktrees can use different explicit index paths without shared mutable state. |
| Performance | Build simple in-memory maps for symbol-keyed occurrence and relationship lookup to avoid repeated full document scans by query planners. No performance benchmark is required in this epic. |
| Dependency management | Use the official SCIP Go binding already required by the loader. Do not add non-SCIP parsing libraries or custom persisted index dependencies. |

## Decomposition

Each scope becomes a code-planning child task.

### Scope 1: Traversal View and Metadata Facts

**Component(s):** Runtime Loaded Index Boundary, Traversal View, Traversal Facts

**Boundary:** In scope: `internal/traversal` accepts the shared runtime loaded SCIP boundary, uses official SCIP Go binding data as source truth, builds process-local inventories for index metadata, documents, document symbols, external symbols, and occurrences, and exposes document paths, languages, position encodings, occurrence ranges, enclosing-range absence/presence, symbol-role bitsets, symbol kind, display name, documentation, signature documentation, enclosing symbol metadata, and occurrence override documentation. Out of scope: raw path loading, command routing, stdout/stderr/status behavior, query-specific matching or filtering, final CLI JSON schemas, source-file reads, custom persisted formats, daemon/watch/cache behavior, relationship lookup indexes, fixture authoring, and ctags fallback behavior.

**Desc:** `scip-search` has a reusable traversal view that accepts the shared runtime loaded-index boundary, preserves official SCIP document, symbol, occurrence, source-location, role, and hover metadata facts, and exposes them to query planners without command-specific behavior.

**Done when:** Query planners can build a traversal view from the runtime loaded official SCIP index and enumerate index metadata, every document, every document-level symbol, every external symbol, and every occurrence exactly as process-local traversal facts; each occurrence fact retains its containing document path and position encoding plus SCIP range, optional enclosing range presence/absence, role bitset, and override documentation; each symbol fact retains full SCIP symbol string, local/external source, kind, display name, documentation entries, signature documentation, and enclosing symbol metadata; traversal does not read source files, open index paths, write process streams, persist derived data, or define any command-specific matching or JSON result behavior.

**Depends on:** Existing task `epic-planning-1-architecture-code-planning-1`.

### Scope 2: Occurrence and Relationship Lookup Views

**Component(s):** Traversal View, Traversal Facts, Query Planner Boundary

**Boundary:** In scope: symbol-keyed occurrence lookup for exact full SCIP symbols, relationship fact extraction from document-level and external symbol information, relationship lookup by owner/source symbol, relationship lookup by target symbol, and preservation of source symbol, target symbol, original direction, and all SCIP edge-kind flags including reference, implementation, type-definition, and definition. Out of scope: partial symbol resolution, package-prefix filtering, reference or implementation query algorithms, missing-symbol command behavior, synthesized relationships, cross-index lookup, duplicate elimination policy, result grouping, response ordering, final JSON schemas, and fixture authoring.

**Desc:** `scip-search` traversal exposes reusable lookup views for exact-symbol occurrences and SCIP relationships so reference, definition, implementation, and related-symbol planners can query traversal facts without scanning raw documents or symbol information.

**Done when:** Given a traversal view from Scope 1, query planners can request all occurrences for an exact full SCIP symbol and receive every matching occurrence with document context, source-location metadata, role bitset, and hover override metadata preserved; planners can request relationships owned by an exact full SCIP symbol and relationships targeting an exact full SCIP symbol and receive relationship facts preserving owner/source symbol, target symbol, original direction, and every SCIP relationship edge-kind flag; lookups for absent symbols return empty traversal results without deciding command-level missing-symbol behavior; traversal does not synthesize relationships or define final reference, implementation, definition, grouping, duplicate, ordering, or JSON semantics.

**Depends on:** Scope 1.

### Scope 3: Traversal Fixtures and Coverage Validation

**Component(s):** Traversal Fixtures, Traversal View, Traversal Facts

**Boundary:** In scope: compact deterministic schema-valid SCIP fixture payloads loaded through the official binding path; fixture helper accessors under traversal-owned test boundaries; coverage validation for documents, languages, paths, position encodings, local symbols, external symbols, occurrences, exact-symbol lookup, relationship owner and target lookup, same-line and multi-position ranges, enclosing range presence/absence, definition and non-definition role cases, multi-bit roles, relationship edge-kind flags, symbol hover metadata, signature documentation, and occurrence override documentation. Out of scope: installing or invoking external indexers, large real-world fixtures, performance benchmarks, ctags fallback fixtures, command-specific fixtures, final command JSON golden files, stdout/stderr/status validation, and query-specific selection or grouping behavior.

**Desc:** `scip-search` has deterministic schema-valid traversal fixtures and coverage validation proving the shared traversal view exposes the SCIP data categories needed by downstream query planners.

**Done when:** Maintainer tests can load the shared fixture set through the same official SCIP binding path used by traversal, build the traversal view, and prove at least two documents with distinct relative paths, document language and position encoding, at least one document-level symbol, at least one external symbol, occurrences for more than one full SCIP symbol, empty lookup behavior for an absent symbol, same-line and multi-position range forms, present and absent enclosing ranges, definition and non-definition occurrences, a multi-bit role occurrence, symbol kind/display name/documentation/signature documentation, present and absent occurrence override documentation, relationship owner and target lookup, original relationship direction, and reference, implementation, definition, and type-definition edge-kind flags; coverage failures identify the missing traversal category and do not assert command JSON, stdout/stderr, exit status, or query-specific semantics.

**Depends on:** Scope 1, Scope 2.

### Spec Coverage

| Spec Requirement | Scope |
|------------------|-------|
| Uses official SCIP Go bindings for parsing and traversal | Scope 1, Scope 3 |
| Reads SCIP output directly through the runtime-loaded boundary | Scope 1, Scope 3 |
| Query planners receive traversal-ready data without owning protobuf parsing | Scope 1 |
| Exposes index metadata, documents, document symbols, external symbols, and occurrences | Scope 1 |
| Preserves document paths, languages, and position encodings | Scope 1, Scope 3 |
| Preserves occurrence ranges, enclosing ranges, and role bitsets | Scope 1, Scope 3 |
| Preserves symbol kind, display name, documentation, signature documentation, enclosing symbol metadata, and occurrence override documentation | Scope 1, Scope 3 |
| Exposes occurrence lookup by exact full SCIP symbol | Scope 2, Scope 3 |
| Exposes relationship lookup by owner/source and target symbols | Scope 2, Scope 3 |
| Preserves relationship owner, target, original direction, and reference, implementation, type-definition, and definition edge-kind flags | Scope 2, Scope 3 |
| Provides deterministic schema-valid traversal fixtures for documents, symbols, occurrences, relationships, ranges, roles, and hover metadata | Scope 3 |
| Keeps traversal process-local without daemon, watch mode, incremental update, cache, custom index format, embeddings, vector store, UI, MCP server, source-file supplementation, or ctags fallback | Scope 1, Scope 2, Scope 3 |
| Leaves command-specific matching, filtering, result grouping, duplicate policy, missing-symbol behavior, response ordering, and final JSON schemas to sibling query epics | Scope 1, Scope 2, Scope 3 |
