# Architecture Plan: CLI Runtime Contract

Status: review

## Goal

Define the shared Go CLI runtime boundary that lets every documented `scip-search` query command load one caller-selected SCIP index, report shared failures predictably, hand a loaded context to query-specific code, emit successful JSON only to stdout, and exit.

## Context

`scip-search` currently has product/specification artifacts but no committed Go module or CLI implementation. The runtime epic covers the process shell shared by later query epics: command routing, the required `--index` flag, one-shot execution, index loading, stream discipline, and exit status conventions. Query traversal and query-specific result schemas are deliberately outside this plan except for the boundary where a loaded index context and command arguments are handed to later query handlers.

### References

- Goal spec: `README.md#what-is-scip-search`
- Parent tasks: `epic-planning-1-us-writing-0`, `epic-planning-1-us-writing-1`, `epic-planning-1-us-writing-2`
- Parent epic: `specs/epics/readme/20260517-095535-epic-planning-1.md`
- Story specs: `specs/stories/readme/20260517-095535-epic-planning-1/*.md`
- Codebase: `.pre-commit-config.yaml`, `.editorconfig`, `README.md`, `specs/`
- Sibling context surveyed: `specs/epics/readme/20260517-100328-epic-planning-2.md`, `specs/epics/readme/20260517-134857-epic-planning-3.md`, `specs/epics/readme/20260517-141006-epic-planning-4.md`

### Constraints

- The product is a Go CLI binary named `scip-search`.
- The documented query commands for this runtime are exactly `symbols`, `references`, `implementations`, and `packages`.
- `--index <index-path>` is the only shared required runtime flag across documented query commands.
- Runtime scope must not define query-specific flags, traversal semantics, or result schemas beyond a generic successful JSON boundary.
- The runtime must read the caller-selected SCIP file directly through official SCIP Go bindings and must not compile, type-check, generate, update, cache, watch, or search default index locations.
- Successful runtime paths reserve stdout for one complete structured JSON value and keep stderr empty.
- Shared usage failures exit with status `2`; shared index-loading failures exit with status `3`; successful commands exit with status `0`.
- `.pre-commit-config.yaml` already exists in the worktree HEAD, so no `bootstrap-precommit` output entry is emitted.
- The existing pre-commit configuration expects Go files to work with project-owned Go tooling and make targets; the first runtime implementation scope owns the minimal Go module and make-target baseline needed by subsequent Go code.

### Assumptions

- **ASM-001**: The Go module path should align with the README installation repository, `github.com/liza-mas/scip-search`. - *Why*: The README uses `https://github.com/liza-mas/scip-search` for install and local clone examples, and no existing `go.mod` overrides it. - Confidence: HIGH
- **ASM-002**: Query handlers can be represented as an internal boundary that returns a marshalable result without prescribing the eventual result fields. - *Why*: The runtime epic owns stream and lifecycle behavior, while sibling query epics own payload content. - Confidence: HIGH

### Open Questions

- None.

## Components

### Go Project Baseline (`go.mod`, `go.sum`, `Makefile`)

**Responsibility:** Establish the repository as a Go CLI project that can be built, tested, vetted, and checked by the existing pre-commit hooks.

**Boundaries:**
- Exposes: module identity, dependency ownership, and make targets used by local validation and pre-commit.
- Depends on: the existing pre-commit hook contract and Go toolchain available to project-scoped commands.

**Key decisions:**
- Module baseline belongs with the first runtime implementation scope: no Go package can pass the existing hooks without project-owned module and make-target structure.
- The baseline should stay minimal: runtime and loader needs drive dependencies; release/install targets remain owned by the distribution epic.

### CLI Entrypoint (`cmd/scip-search/`)

**Responsibility:** Adapt process-level inputs and outputs to the internal runtime and terminate with the runtime-provided exit status.

**Boundaries:**
- Exposes: the executable command surface named `scip-search`.
- Depends on: internal CLI runtime package and process stdout/stderr streams.

**Key decisions:**
- Keep `main` thin: process termination stays at the outer boundary so internal packages remain directly testable without process exits.
- No daemon, prompt, server, watcher, or background lifecycle starts from the entrypoint.

### CLI Runtime (`internal/cli/`)

**Responsibility:** Own shared command routing, shared flag parsing, one-shot orchestration, and separation between usage failures, index-loading failures, and successful query execution.

**Boundaries:**
- Exposes: a single process-run boundary that accepts argv plus output streams and returns a process status.
- Depends on: runtime contracts, an index loader interface, and query command handlers.

**Key decisions:**
- Command routing is centralized for the four documented query names so shared invocation behavior cannot drift per command.
- The shared `--index` flag is parsed at the runtime layer before index loading; missing command, unsupported command, missing `--index`, and missing index values are usage failures.
- Extra command names after the selected command are treated as invocation shape errors rather than as a second query in the same process.

### Runtime Contracts (`internal/runtime/`)

**Responsibility:** Define shared error classes, exit statuses, result-writing behavior, and the query execution boundary used by `internal/cli`, `internal/scipindex`, and later query packages.

**Boundaries:**
- Exposes: status constants, diagnostic failure categories, JSON success writer behavior, loaded-query-context shape, and query handler contracts.
- Depends on: Go standard library stream and JSON facilities, plus the loaded index type exposed by the loader boundary.

**Key decisions:**
- The runtime owns stream purity: success writes exactly one JSON value to stdout and no stderr; shared runtime failures write diagnostics only to stderr and no stdout.
- Diagnostics are human-readable stderr text unless a future spec introduces a machine-readable error schema.
- Query-specific packages own the contents of successful results, but the runtime owns encoding them as one JSON value.

### SCIP Index Loader (`internal/scipindex/`)

**Responsibility:** Convert a caller-selected path into either a loaded SCIP index context or a shared index-loading failure before query traversal begins.

**Boundaries:**
- Exposes: a loader implementation behind the runtime loader interface.
- Depends on: filesystem reads and official SCIP Go bindings.

**Key decisions:**
- Path validation and SCIP parsing are one runtime boundary: nonexistent, unreadable, directory, and invalid SCIP inputs all fail before query handlers run.
- The loader never searches default locations and never mutates caller-owned index files.
- The loaded context is the only index input for the current process invocation and is not persisted beyond that invocation.

### Query Handler Boundary (`internal/query/` or command-specific internal packages)

**Responsibility:** Provide the later query-specific execution surface that receives loaded SCIP data and command-specific arguments, then returns a result for runtime JSON encoding.

**Boundaries:**
- Exposes: command handlers registered by name with the CLI runtime.
- Depends on: the loaded query context and runtime result contract.

**Key decisions:**
- This epic may define and test the boundary with fakes or minimal no-traversal handlers, but must not define real symbol, package, reference, or implementation traversal behavior.
- Later query epics can replace or fill handlers without changing process, stream, and loading contracts.

## Interfaces

### CLI Entrypoint -> CLI Runtime

**Contract:** Process argv and writable stdout/stderr streams cross into the runtime; the runtime returns an integer process status.
**Direction:** `cmd/scip-search` calls `internal/cli`.
**Invariants:** Entrypoint does not print diagnostics or JSON independently; all user-observable stream content comes from the runtime.

### CLI Runtime -> Runtime Contracts

**Contract:** Parsed invocation outcomes are represented as shared success, usage failure, or index-loading failure results with documented statuses.
**Direction:** `internal/cli` uses `internal/runtime`.
**Invariants:** Usage failures map to status `2`; index-loading failures map to status `3`; success maps to status `0`.

### CLI Runtime -> SCIP Index Loader

**Contract:** A syntactically supplied `--index` value crosses as the caller-selected path; the loader returns either a loaded query context or an index-loading failure.
**Direction:** `internal/cli` calls `internal/scipindex` through a narrow loader boundary.
**Invariants:** Loader is not called until command routing and shared flag validation pass; query handlers are not called unless loading succeeds.

### CLI Runtime -> Query Handler Boundary

**Contract:** The selected command name, command-specific remaining arguments, and loaded query context cross into a query handler; a successful handler result crosses back as a value for runtime JSON encoding.
**Direction:** `internal/cli` dispatches to the selected handler.
**Invariants:** Exactly one selected query command is executed per process invocation; query handlers do not write directly to stdout/stderr.

### Runtime Contracts -> Calling Process

**Contract:** Success produces one complete structured JSON value on stdout, empty stderr, and status `0`; shared runtime failures produce empty stdout, diagnostic stderr, and documented nonzero status.
**Direction:** `internal/runtime` writes to process streams through the runtime boundary.
**Invariants:** No progress text, warnings, prompts, or logging lines are mixed into successful stdout.

## Data Flow

```text
argv + stdout/stderr
  -> cmd/scip-search
  -> internal/cli command router
  -> shared --index validation
  -> internal/scipindex loader
  -> loaded query context
  -> selected query handler boundary
  -> runtime JSON encoder
  -> stdout + exit 0
```

Failure paths:

```text
missing/unsupported command or missing --index
  -> usage diagnostic
  -> stderr only + exit 2

nonexistent/unreadable/directory/invalid SCIP index
  -> index-loading diagnostic
  -> stderr only + exit 3
```

## Cross-Cutting Concerns

| Concern | Approach |
|---------|----------|
| Error handling | Classify failures at the boundary where they occur: invocation shape errors in `internal/cli`, index input errors in `internal/scipindex`, and stream/status mapping in `internal/runtime`. |
| Observability | Do not add progress logging or prompts. Diagnostics appear only on failure stderr. |
| Configuration | Only explicit command-line inputs are configured in this epic; no default index path, config file, environment lookup, cache directory, or persistent session state. |
| Testing | Unit-test routing and status mapping with injected loader/handler fakes; integration-test process streams across documented commands; loader tests use missing paths, directories, unreadable inputs where portable, invalid files, and a valid SCIP fixture. |
| Security | Treat index paths as caller-owned filesystem inputs; never mutate selected files; avoid shelling out to indexers; do not expose secrets in diagnostics. |
| Concurrency | Keep all state invocation-local so concurrent worktrees can pass different `--index` paths without shared mutable process state. |
| Dependency management | Add only dependencies required for the Go CLI and official SCIP binding boundary; release/install dependencies remain out of scope. |

## Decomposition

Each scope becomes a code-planning child task.

### Scope 1: Shared CLI Runtime And Stream Contract

**Component(s):** Go Project Baseline, CLI Entrypoint, CLI Runtime, Runtime Contracts, Query Handler Boundary

**Boundary:** In scope: Go module and validation baseline required by existing hooks, `scip-search` entrypoint, documented command routing for `symbols`, `references`, `implementations`, and `packages`, required shared `--index` flag parsing, usage failure classification, one-shot lifecycle, status constants, JSON success writer, stderr-only runtime diagnostics, and an injected no-traversal query handler boundary. Out of scope: official SCIP parsing, filesystem index validation beyond invoking the loader boundary, query-specific flags, symbol/package/reference/implementation traversal, query-specific result fields, version/install/release behavior, daemon/watch/MCP/UI behavior, and custom index formats.

**Desc:** `scip-search` has a shared Go CLI runtime shell that routes the four documented query commands, enforces the common `--index` invocation contract, provides one-shot process execution, and centralizes stdout/stderr/status behavior without implementing SCIP traversal.

**Done when:** The Go project baseline supports the existing pre-commit Go hooks, every documented query command is recognized through one shared runtime path, missing or unsupported commands and missing `--index` inputs fail with status `2` through stderr only, valid routed invocations call an injected loader and query handler exactly once for one selected command, successful handler results are emitted as one parseable JSON value on stdout with empty stderr and status `0`, and no query-specific traversal or result schema is defined.

**Depends on:** None.

### Scope 2: Caller-Selected SCIP Index Loading Boundary

**Component(s):** SCIP Index Loader, Runtime Contracts, CLI Runtime

**Boundary:** In scope: caller-selected index path validation after shared invocation succeeds, official SCIP Go binding load/parse boundary, loaded query context handoff to the selected command handler, index-loading diagnostics, status `3`, and cross-command coverage for missing, unreadable, directory, invalid, and valid SCIP inputs. Out of scope: generating, updating, caching, watching, compiling, type-checking, fallback ctags parsing, custom index formats, query traversal, query-specific result schemas, and release/install behavior.

**Desc:** `scip-search` has a caller-selected SCIP index loading boundary that validates the supplied `--index` path, parses readable SCIP input through official Go bindings, reports shared index-loading failures consistently, and exposes loaded context to query handlers before traversal-specific work begins.

**Done when:** For `symbols`, `references`, `implementations`, and `packages`, nonexistent, unreadable, directory, and invalid SCIP inputs fail before query handlers run with empty stdout, diagnostic stderr, and status `3`; readable valid SCIP input is parsed through the official SCIP Go binding boundary and passed as the only loaded context for the current invocation; selected index files are never generated, searched for, updated, deleted, cached, or parsed as a custom format.

**Depends on:** Scope 1.

### Spec Coverage

| Spec Requirement | Scope |
|------------------|-------|
| Thin Go binary invoked as `scip-search` | Scope 1 |
| Loads a SCIP index file at the caller-provided path | Scope 2 |
| Answers one query and exits as a one-shot process | Scope 1 |
| Prints structured JSON to stdout on success | Scope 1 |
| Uses official SCIP Go bindings and reads SCIP output directly | Scope 2 |
| `symbols`, `references`, `implementations`, and `packages` command forms exist | Scope 1 |
| Shared required `--index <index-path>` flag across commands | Scope 1 |
| Missing or invalid invocation shape fails before traversal | Scope 1 |
| Missing, unreadable, directory, or invalid SCIP input fails before traversal | Scope 2 |
| Shared runtime failures write diagnostics only to stderr with nonzero status | Scope 1 and Scope 2 |
| Concurrent worktrees can select explicit index paths without global defaults | Scope 2 |
| No daemon, watch mode, incremental updates, MCP server, UI, graph visualization, semantic similarity, vector storage, custom index format, index generation, compilation, or type-checking | Scope 1 and Scope 2 |
