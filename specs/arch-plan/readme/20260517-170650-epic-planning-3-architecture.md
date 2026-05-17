# Architecture Plan: Discovery Queries

Status: draft

## Goal

Implement symbol and package discovery as query-specific Go components over the shared SCIP traversal view, with deterministic JSON payloads and fixture-backed golden validation.

## Context

The repository currently contains the README, story specifications, and project quality configuration; no Go source tree is present yet. This plan defines the discovery-query structure that downstream code planners should apply when the shared runtime and traversal foundations from sibling epics are available.

Discovery starts after the shared CLI runtime has selected and loaded one SCIP index and after the traversal layer has exposed process-local inventories of symbols, external symbols, documents, occurrences, definition context, and range data. This task owns only `symbols --name`, `packages`, `packages --prefix`, and their successful discovery payloads and golden cases.

### References

- Goal spec: `README.md#scip-symbol-format`, `README.md#what-is-scip-search`, and `README.md#language-support`.
- Parent tasks: `epic-planning-3-us-writing-0`, `epic-planning-3-us-writing-1`, `epic-planning-3-us-writing-2`.
- Parent epic: `specs/epics/readme/20260517-134857-epic-planning-3.md`.
- Parent story outputs: `specs/epics/readme/20260517-134857-epic-planning-3.outputs.json`.
- Symbol stories: `specs/stories/readme/20260517-134857-epic-planning-3/cap-001-symbol-name-discovery.md`.
- Package stories: `specs/stories/readme/20260517-134857-epic-planning-3/CAP-002-01-package-inventory.md`, `CAP-002-02-package-prefix-filtering.md`, and `CAP-002-03-package-result-json-shape.md`.
- Fixture stories: `specs/stories/readme/20260517-134857-epic-planning-3/CAP-003-01-symbol-query-fixtures.md`, `CAP-003-02-package-query-fixtures.md`, and `CAP-003-03-golden-json-validation.md`.
- Runtime boundary references: `specs/stories/readme/20260517-095535-epic-planning-1/stdout-json-success-contract.md` and `scip-index-loading-boundary.md`.
- Traversal boundary references: `specs/epics/readme/20260517-100328-epic-planning-2.md`, `specs/epics/readme/20260517-100328-epic-planning-2-output.json`, `CAP-001-01-scip-binding-input.md`, and `CAP-001-02-document-and-symbol-inventory.md`.
- Codebase explored: repository file list, `.pre-commit-config.yaml`, `.editorconfig`, and all epic-planning-3 story documents listed above.

### Constraints

- Shared CLI command routing, `--index` parsing, path validation, load failures, stderr behavior, and success stream purity stay with epic-planning-1.
- SCIP parsing and traversal inventories stay with epic-planning-2 and must use official SCIP Go bindings.
- Reference and implementation behavior stays with epic-planning-4.
- Discovery queries must not introduce fuzzy, regex, glob, semantic, case-folded, cross-index, package-registry, dependency-graph, source-file, daemon, watcher, custom index, UI, MCP, embedding, or vector behavior.
- Successful no-match discovery returns explicit empty collections, not shared runtime failures.
- `.pre-commit-config.yaml` already exists in the worktree, so no `bootstrap-precommit` child scope is emitted.

### Assumptions

- **ASM-001**: Discovery implementation can consume a traversal package that exposes symbol information, external symbols, occurrences, document paths, and definition ranges without owning raw protobuf parsing - *Why*: epic-planning-2 defines this as the query-planner traversal view - Confidence: HIGH.
- **ASM-002**: The first four SCIP symbol components are sufficient to form the package identity and `packageKey` for package-bearing symbols - *Why*: README and package stories define the SCIP symbol format as scheme, package manager, package name, package version, then descriptors - Confidence: HIGH.
- **ASM-003**: Symbol query fixture data may share the same schema-valid SCIP fixture source as package query fixture data when golden cases remain command-specific - *Why*: CAP-003 permits shared deterministic fixtures and requires separation at expected-result level - Confidence: HIGH.

### Open Questions

- None.

---

## Components

### Runtime Command Adapter (`cmd/scip-search` / `internal/cli`)

**Responsibility:** Own process invocation, command routing, shared flags, index loading, and stdout/stderr behavior under the epic-planning-1 contract.

**Boundaries:**
- Exposes: selected command parameters and the loaded traversal view to query packages.
- Depends on: discovery query interfaces for `symbols --name` and `packages`.

**Key decisions:**
- Discovery code is called after shared runtime validation and loaded-index handoff: keeps query code free of shared error taxonomy and stream handling.
- Runtime writes one JSON value returned by discovery: preserves the shared JSON-only stdout contract while allowing query-specific payloads.

### Traversal View (`internal/scipview`)

**Responsibility:** Provide process-local, official-SCIP-derived inventories to query planners.

**Boundaries:**
- Exposes: symbol inventory, external symbol inventory, occurrence inventory, document paths, definition occurrence context, and source ranges.
- Depends on: official SCIP Go binding data from the loaded-index boundary.

**Key decisions:**
- Discovery queries depend on traversal interfaces, not raw protobuf fields: avoids duplicating SCIP parsing and keeps all commands aligned on one source of indexed facts.
- Traversal remains command-neutral: symbol matching, package filtering, ordering, and JSON fields live outside traversal.

### SCIP Identity Model (`internal/scipmodel`)

**Responsibility:** Interpret full SCIP symbol strings into reusable identity facts needed by discovery queries.

**Boundaries:**
- Exposes: exact full symbol string, scheme, package manager, package name, package version, descriptor text, display/match name, and `packageKey`.
- Depends on: symbol strings and optional display names supplied by traversal.

**Key decisions:**
- Preserve the original full symbol string on every symbol result: downstream `references` and `implementations` commands require unchanged symbol input.
- Build `packageKey` from the exact package prefix components without descriptors: package discovery needs comparison keys without forcing callers to parse full symbols.
- Keep matching text and package identity derivation together: both discovery commands need the same symbol-format interpretation and should not duplicate string parsing.

### Symbol Discovery Query (`internal/query/discovery`)

**Responsibility:** Resolve `symbols --name <name>` against traversal symbol facts and return deterministic symbol result payload data.

**Boundaries:**
- Exposes: a query operation that accepts literal `name` text and returns a `symbols` collection.
- Depends on: traversal view data, SCIP identity model, and optional definition context.

**Key decisions:**
- Match literal, case-sensitive query text against display name or descriptor text: follows CAP-001 and avoids unsupported pattern languages.
- Return all matches, including ambiguous partial names: the command is discovery, not disambiguation.
- Sort results by exact full `symbol` string ascending: deterministic over observable output values and independent of traversal iteration order.
- Represent missing definition context as absent optional context on an otherwise successful symbol entry: full symbol discovery remains useful without source location.

### Package Discovery Query (`internal/query/discovery`)

**Responsibility:** Enumerate package identities from indexed symbol facts, apply optional literal package-name prefix filtering, and return deterministic package result payload data.

**Boundaries:**
- Exposes: a query operation that accepts an optional literal prefix and returns a `packages` collection.
- Depends on: traversal view data and SCIP identity model.

**Key decisions:**
- Derive package candidates from indexed symbol facts exposed by traversal, including document and external symbol inventories: packages reflect what the selected index contains.
- De-duplicate by full package identity before final output: many symbols may share one package.
- Apply `--prefix` only to `packageName`: prevents accidental matches on scheme, package manager, version, descriptors, symbol names, or document paths.
- Sort results by exact `packageKey` ascending: deterministic and shared by unfiltered and filtered package queries.

### JSON Payload Contracts (`internal/output` or query result structs)

**Responsibility:** Carry query-specific successful payload shapes to the runtime JSON writer.

**Boundaries:**
- Exposes: `symbols` and `packages` result structs with stable field names.
- Depends on: discovery query results.

**Key decisions:**
- Keep payload structs query-specific while runtime owns JSON emission: shared stream behavior stays centralized.
- Symbol entries include exact `symbol`, match context, package identity, and optional definition location.
- Package entries include `scheme`, `packageManager`, `packageName`, `packageVersion`, and `packageKey`.

### Discovery Fixtures and Golden Cases (`internal/query/discovery/testdata` / `testdata/scip`)

**Responsibility:** Provide deterministic SCIP fixture inputs and expected successful JSON cases for discovery commands.

**Boundaries:**
- Exposes: small schema-valid SCIP fixture data plus per-command golden JSON expectations.
- Depends on: shared traversal fixture mechanics and the public CLI/runtime path used by commands.

**Key decisions:**
- Fixture data may be shared across symbol and package cases, but expected JSON cases stay command-specific: avoids cross-command coupling in tests.
- Golden validation compares parsed JSON values while asserting array order: object field order is not a semantic contract.
- Tests exercise the same SCIP loading and traversal path as commands: avoids validating a different in-memory shortcut.

---

## Interfaces

### Runtime Command Adapter -> Discovery Queries

**Contract:** Runtime supplies the loaded traversal view plus command-specific parameters: `name` for `symbols --name`, optional `prefix` for `packages`.
**Direction:** Runtime calls discovery query package after shared validation and index loading.
**Invariants:** Discovery receives exactly one loaded index view for the current process and does not emit stdout/stderr or decide shared exit status.

### Discovery Queries -> Traversal View

**Contract:** Discovery reads symbol, external symbol, occurrence, document, and definition context facts from traversal interfaces.
**Direction:** Discovery pulls immutable process-local facts from traversal.
**Invariants:** Discovery does not parse SCIP bytes, reopen the index file, compile code, generate indexes, or persist derived views across invocations.

### Discovery Queries -> SCIP Identity Model

**Contract:** Discovery passes exact full symbol strings and optional display names to identity helpers and receives package identity, descriptor/match text, and package key data.
**Direction:** Discovery calls identity helpers during candidate construction.
**Invariants:** Full symbol strings are never normalized or rewritten; package keys are derived from scheme, package manager, package name, and package version only.

### Discovery Queries -> JSON Payload Contracts

**Contract:** Discovery returns pure data structs for successful query payloads: `symbols` collection for symbol discovery and `packages` collection for package discovery.
**Direction:** Discovery produces structs; runtime encodes them as the single success JSON value.
**Invariants:** Empty successful results use explicit empty collections; no success diagnostics are included in payload streams; failure behavior remains upstream.

### Fixture Validation -> Runtime and Discovery

**Contract:** Tests execute discovery command paths against deterministic SCIP fixtures and compare parsed JSON to golden expectations.
**Direction:** Test harness invokes command/runtime path, captures stdout, parses JSON, and compares command-specific payloads.
**Invariants:** Fixture tests do not assert shared runtime failures, reference/implementation payloads, object field order, external indexer installation, or large real-world indexes.

---

## Data Flow

```
CLI args
  -> Runtime command adapter
  -> Shared index loading boundary
  -> Traversal view
  -> SCIP identity model
  -> Symbol or package discovery query
  -> Query-specific payload struct
  -> Runtime JSON stdout writer
```

Symbol discovery flow:

```
--name text
  -> symbol facts from traversal
  -> display/descriptor literal matching
  -> optional definition context attachment
  -> sort by full symbol
  -> {"symbols":[...]}
```

Package discovery flow:

```
optional --prefix
  -> symbol facts from traversal
  -> package identity extraction
  -> de-duplicate by package key
  -> optional packageName prefix filter
  -> sort by packageKey
  -> {"packages":[...]}
```

Fixture validation flow:

```
schema-valid SCIP fixture
  -> normal runtime loading and traversal
  -> discovery command invocation
  -> captured success JSON
  -> parsed JSON comparison with golden cases
```

---

## Cross-Cutting Concerns

| Concern | Approach |
|---------|----------|
| Error handling | Shared invocation, path, load, stdout/stderr, and exit failures remain in runtime. Discovery no-match and ambiguous-match cases are successful payloads. |
| Determinism | Sort symbol results by exact `symbol`; sort package results by exact `packageKey`; compare golden JSON values while asserting array order. |
| Configuration | No discovery-specific configuration. `--name` and optional `--prefix` are literal inputs accepted by shared CLI parsing. |
| Security | Discovery does not execute shell commands, read source files, query registries, follow dependency graphs, or load additional paths. It only consumes the already-loaded index view. |
| Performance | Build process-local candidate collections from traversal inventories once per invocation. Avoid repeated full traversal scans within a single query. |
| Testing | Use small deterministic SCIP fixtures through the same runtime loading and traversal path as the CLI commands. Keep symbol and package golden cases separate. |
| Observability | Successful commands emit only JSON stdout and empty stderr. Any diagnostics remain part of shared failure handling, not discovery query code. |

---

## Decomposition

Each scope becomes a code-planning child task.

### Scope 1: SCIP Discovery Identity Model

**Component(s):** SCIP Identity Model.
**Boundary:** In scope: reusable interpretation of full SCIP symbol strings into exact symbol value, display/descriptor match text, scheme, package manager, package name, package version, and package key for discovery queries. Out of scope: CLI routing, index loading, traversal construction, fuzzy/regex/semantic matching, reference or implementation lookup, and final command JSON emission.
**Desc:** `scip-search` maintainers have a shared discovery identity model that preserves full SCIP symbols exactly while exposing package identity components, descriptor or display match text, and package keys for symbol and package discovery.
**Done when:** Code planners can specify a falsifiable implementation where discovery code derives scheme, package manager, package name, package version, descriptor or display match text, and packageKey from fixture SCIP symbols while preserving the exact full symbol string unchanged for downstream commands.
**Scope:** Reusable SCIP symbol identity extraction and match-text/package-key helpers for discovery queries. Excludes CLI command routing, shared index loading, traversal construction, final JSON emission, reference or implementation behavior, and unsupported fuzzy, regex, semantic, or cross-index matching.
**Depends on:** none.

### Scope 2: Symbol Discovery Query

**Component(s):** Symbol Discovery Query, JSON Payload Contracts.
**Boundary:** In scope: `symbols --name <name>` literal matching, multi-match and empty successful behavior, stable ordering, exact full symbol preservation, package identity fields, match context, and optional definition location in the `symbols` payload. Out of scope: shared runtime failures, raw SCIP traversal construction, package listing, references, implementations, fuzzy/regex/semantic matching, and source-file enrichment.
**Desc:** `scip-search` users can run `symbols --name <name>` and receive deterministic successful `symbols` results that include every literal partial-name match, exact full SCIP symbol strings, package identity, match context, and available definition context.
**Done when:** Code planners can specify a falsifiable implementation where `symbols --name Supervisor`, `symbols --name Run`, and a no-match query over deterministic fixtures return JSON payloads with an explicit `symbols` collection, all and only literal case-sensitive matches, stable ascending order by full `symbol`, exact symbol strings, package identity fields, match context, and optional definition location when traversal provides it.
**Scope:** Symbol discovery query behavior and successful symbol payload data over the shared traversal view. Excludes CLI routing and `--index` handling, shared stdout/stderr and runtime failure behavior, raw SCIP parsing or traversal construction, package discovery, reference and implementation queries, source-file reads, ranking, fuzzy search, regex search, semantic search, and cross-index matching.
**Depends on:** Scope 1.

### Scope 3: Package Discovery Query

**Component(s):** Package Discovery Query, JSON Payload Contracts.
**Boundary:** In scope: `packages` all-package listing, optional `--prefix` literal package-name filtering, de-duplication by full package identity, stable ordering, empty successful results, and package JSON fields. Out of scope: shared runtime failures, raw SCIP traversal construction, registry queries, dependency graph analysis, version resolution, non-package filters, symbols payload behavior, references, and implementations.
**Desc:** `scip-search` users can run `packages` with or without `--prefix` and receive deterministic successful `packages` results with de-duplicated package identities filtered only by literal package-name prefix when provided.
**Done when:** Code planners can specify a falsifiable implementation where unfiltered `packages`, `packages --prefix github.com/liza-mas/`, `packages --prefix github.com/liza-mas/scip-search`, and a no-match prefix over deterministic fixtures return JSON payloads with an explicit `packages` collection, one entry per full package identity, stable ascending order by `packageKey`, exact `scheme`, `packageManager`, `packageName`, `packageVersion`, and `packageKey` fields, and no matches caused by prefix text outside `packageName`.
**Scope:** Package discovery query behavior and successful package payload data over the shared traversal view. Excludes CLI routing and `--index` handling, shared stdout/stderr and runtime failure behavior, raw SCIP parsing or traversal construction, symbol discovery, reference and implementation queries, package registry lookups, dependency graph traversal, version resolution, descriptor filters, regex/glob/fuzzy/semantic filters, and source-file reads.
**Depends on:** Scope 1, Scope 2.

### Scope 4: Discovery Fixtures and Golden JSON Validation

**Component(s):** Discovery Fixtures and Golden Cases.
**Boundary:** In scope: deterministic SCIP fixture coverage and successful golden JSON cases for symbol partial matching, ambiguous names, symbol no-match, package listing, package de-duplication, package prefix filtering, package no-match, and stable output ordering. Out of scope: shared runtime failure fixtures, raw traversal fixture ownership, reference or implementation query fixtures, large real-world fixtures, performance benchmarks, external indexer installation, and ctags fallback fixtures.
**Desc:** `scip-search` maintainers can validate symbol and package discovery against deterministic SCIP fixtures and golden JSON cases that cover matching, filtering, empty results, ambiguous names, de-duplication, exact symbol preservation, package identities, and stable ordering.
**Done when:** Code planners can specify a falsifiable validation plan where fixture-backed tests exercise the normal SCIP loading and traversal path and compare parsed JSON values for `symbols --name Supervisor`, `symbols --name Run`, `symbols --name DoesNotExist`, unfiltered `packages`, matching package prefixes, exact package-prefix narrowing, and no-match package prefixes, asserting stable array ordering and command-specific payload fields without asserting shared runtime failures or object field order.
**Scope:** Query-specific discovery fixtures, golden JSON cases, and validation coverage for symbol and package discovery. Excludes shared runtime failure fixtures, shared traversal fixture construction ownership, reference or implementation query fixtures, large real-world fixtures, performance benchmarks, external indexer installation, ctags fallback data, and alternate output formats.
**Depends on:** Scope 2, Scope 3.

### Spec Coverage

| Spec Requirement | Scope |
|------------------|-------|
| Preserve full SCIP symbol strings exactly for follow-up symbol-based commands | Scope 1, Scope 2 |
| Interpret SCIP package identity components from the symbol format | Scope 1, Scope 3 |
| `symbols --name <name>` supports deterministic literal partial-name matching | Scope 2 |
| Ambiguous symbol names return multiple successful matches rather than a failure | Scope 2 |
| Empty symbol matches return an explicit successful `symbols` collection | Scope 2 |
| Symbol entries include exact full `symbol`, match context, package identity, and available definition context | Scope 2 |
| `packages` lists indexed packages without a prefix | Scope 3 |
| `packages --prefix <prefix>` filters by literal package-name prefix only | Scope 3 |
| Package results are de-duplicated by full package identity | Scope 3 |
| Empty package prefix matches return an explicit successful `packages` collection | Scope 3 |
| Package entries expose `scheme`, `packageManager`, `packageName`, `packageVersion`, and `packageKey` | Scope 3 |
| Successful discovery output remains one structured JSON value and does not redefine shared runtime failures | Scope 2, Scope 3, Scope 4 |
| Fixtures cover symbol partial matching, ambiguous names, package filtering, de-duplication, empty successful results, and stable ordering | Scope 4 |
| Fixtures do not depend on external indexer installation or large real-world repositories | Scope 4 |

---

## Shared-File Audit

- Scope 1 owns shared discovery identity helpers. Scopes 2 and 3 depend on it to avoid parallel edits to shared parsing and package-key logic.
- Scopes 2 and 3 may both extend discovery query package structure and successful payload contracts. Scope 3 depends on Scope 2 to serialize likely shared file edits while still keeping package behavior as a separate planning domain.
- Scope 4 depends on Scopes 2 and 3 because fixture golden cases assert both command payloads and should validate final observable behavior after query contracts exist.

## Validation Plan

- Confirm the architecture document contains components, interfaces, data flow, cross-cutting concerns, decomposition, and spec coverage for all parent tasks.
- Validate output JSON entries with `jq`.
- Run pre-commit against the architecture document.
- Commit only the architecture document and submit the task output definitions for review.
