# Code Plan: Caller-Selected SCIP Index Loading Boundary

Status: draft

## Source Context

Based on:
- `README.md#what-is-scip-search`
- `README.md#language-support`
- `README.md#out-of-scope`
- `specs/epics/readme/20260517-095535-epic-planning-1.md`
- `specs/arch-plan/readme/20260517-170052-epic-planning-1-architecture.md`
- `specs/stories/readme/20260517-095535-epic-planning-1/index-path-validation.md`
- `specs/stories/readme/20260517-095535-epic-planning-1/scip-index-loading-boundary.md`
- `specs/stories/readme/20260517-095535-epic-planning-1/stderr-exit-error-contract.md`
- `specs/stories/readme/20260517-095535-epic-planning-1/shared-runtime-error-cases.md`
- `specs/stories/readme/20260517-095535-epic-planning-1/stdout-json-success-contract.md`
- Prior code plan `specs/plans/readme/20260517-191414-epic-planning-1-architecture-code-planning-0.md`
- Blackboard task `epic-planning-1-architecture-code-planning-0`, which defines the shared runtime shell this plan extends.
- External binding check: pkg.go.dev for `github.com/scip-code/scip/bindings/go/scip` showed the official SCIP Go package at module version `v0.7.1` on 2026-05-17, and `google.golang.org/protobuf/proto` remains the Go protobuf marshal/unmarshal boundary.

## Planning Boundary

This plan covers the runtime boundary that turns a caller-selected `--index` path into loaded SCIP data, or an index-loading failure, before any query-specific traversal runs.

In scope: caller-selected index path validation after shared invocation succeeds, official SCIP Go binding load/parse boundary, loaded query context handoff to the selected command handler, index-loading diagnostics, status `3`, and cross-command coverage for nonexistent, unreadable, directory, invalid, and valid SCIP inputs.

Out of scope: creating the shared CLI runtime shell from the prior plan, generating indexes, updating indexes, caching, watching, compiling, type-checking, fallback ctags parsing, custom index formats, query traversal, query-specific result schemas, and release/install behavior.

## Architectural Direction

Extend the shared runtime shell from `epic-planning-1-architecture-code-planning-0` without recreating it:

```text
cmd/scip-search
  -> internal/cli
       -> internal/runtime status/handler contracts
       -> internal/scipindex production loader
            -> filesystem read of caller-selected path
            -> official scip Go binding protobuf parse
       -> selected query handler boundary
```

`internal/scipindex` should own path-level validation and official SCIP parsing. `internal/cli` should keep owning command routing, shared `--index` parsing, stream/status mapping, and the rule that handlers are not called unless loading succeeds. Query packages should only receive the loaded context; they must not parse files or choose index paths.

If a coder starts before the runtime extension points promised by `epic-planning-1-architecture-code-planning-0` exist, the coding task must block instead of implementing command routing, status constants, JSON stream helpers, or the entrypoint as part of this scope.

## Planned Coding Tasks

### Task 1 - Implement official SCIP index loader

**desc:** `scip-search` has a production `internal/scipindex` loader that validates the caller-selected path, reads only that file, parses SCIP bytes through the official Go bindings, and returns a loaded runtime context or an index-loading failure.

**done_when:** Unit tests for `internal/scipindex` pass asserting nonexistent paths, directory paths, injected unreadable file-open errors, and readable invalid bytes all return the shared index-loading failure class without calling query handlers, mutating the selected path, searching fallback locations, or attempting custom-format parsing; a readable valid temp `.scip` file generated with `github.com/scip-code/scip/bindings/go/scip` and `google.golang.org/protobuf/proto` parses successfully into the loaded context with the caller-selected path preserved and official SCIP index data available.

**scope:** In scope: `internal/scipindex` production loader, minimal runtime loaded-index context type or adapter needed by the existing loader interface, Go module dependency updates for the official SCIP binding and protobuf APIs, and colocated loader unit tests with temp files plus an injected unreadable-open test double. Out of scope: shared command routing, shared `--index` parsing, process entrypoint behavior, query-specific flags, query traversal, query result schemas, generated index fixtures, fallback ctags parsing, default index discovery, caching, watching, compilation, type-checking, release/install behavior, and README changes.

**spec_ref:** specs/stories/readme/20260517-095535-epic-planning-1/scip-index-loading-boundary.md

**task_depends_on:** `epic-planning-1-architecture-code-planning-0`

**Planned files:**
- `go.mod`
- `go.sum`
- `internal/scipindex/loader.go`
- `internal/scipindex/loader_test.go`
- `internal/runtime/index.go` if the prior runtime shell does not already expose a loaded-context shape

**Implementation notes:**
- Prefer `github.com/scip-code/scip/bindings/go/scip` for the official index type and `google.golang.org/protobuf/proto` for binary protobuf decode or fixture generation.
- Keep the selected path as metadata on the loaded context so diagnostics and downstream traversal can identify the caller-selected input without re-reading or rediscovering it.
- Do not introduce a parser for any format other than SCIP protobuf bytes.
- Validation command for the task: `go test ./internal/scipindex ./internal/runtime`.

### Task 2 - Wire loader failures and loaded context into the shared CLI runtime

**desc:** `scip-search` routes the selected `--index` path through the production SCIP loader after shared invocation validation, reports every loader failure as status `3` through stderr only, and calls only the selected query handler with the loaded context on successful load.

**done_when:** Unit tests in `internal/cli` pass for `symbols`, `references`, `implementations`, and `packages` asserting loader failures representing nonexistent, unreadable, directory, and invalid SCIP inputs return status `3`, write a diagnostic only to stderr, leave stdout empty, and do not call any query handler; successful loads call only the selected handler exactly once with the loaded context from the loader and emit the handler result through the existing JSON success path; tests also assert the runtime does not search for, generate, update, delete, cache, or replace the selected index path.

**scope:** In scope: `internal/cli` production loader integration, loader-error to `StatusIndexLoad` mapping, diagnostic propagation, handler-not-called assertions for loader failures, loaded context handoff to the selected handler, and unit tests using loader and handler fakes. Out of scope: implementing the loader itself, shared command routing or shared `--index` parsing already owned by the prior runtime shell plan, query-specific argument validation, SCIP traversal, query result schemas, fallback ctags parsing, default index discovery, caching, watching, compilation, type-checking, release/install behavior, and README changes.

**spec_ref:** specs/stories/readme/20260517-095535-epic-planning-1/stderr-exit-error-contract.md

**Depends on:** Task 1.

**task_depends_on:** `epic-planning-1-architecture-code-planning-0`

**Planned files:**
- `internal/cli/runtime.go`
- `internal/cli/runtime_test.go`

**Implementation notes:**
- Keep loader invocation after command and shared `--index` validation so usage failures remain status `2`.
- The query handler boundary should receive exactly the loaded context returned by the loader for the current invocation; do not create a second context from globals, defaults, or cached state.
- Validation command for the task: `go test ./internal/cli`.

### Task 3 - Cover index-loading behavior end to end across documented commands

**desc:** `scip-search` has end-to-end command coverage proving the production runtime reports index-loading failures and valid SCIP loads consistently for `symbols`, `references`, `implementations`, and `packages`.

**done_when:** Command-level tests pass by invoking the `scip-search` entrypoint or testable run adapter for each documented command and asserting nonexistent, directory, and invalid selected index paths produce empty stdout, non-empty diagnostic stderr, and status `3`; unreadable selected index behavior is covered for each documented command through the same runtime path with a deterministic loader-open failure; valid temp `.scip` inputs generated through the official SCIP Go binding reach the query handler boundary and produce parseable JSON success with status `0` and empty stderr; `go test ./...` passes.

**scope:** In scope: black-box or entrypoint-level tests under `cmd/scip-search`, test helpers for generating minimal valid SCIP temp files through official bindings, deterministic unreadable-open coverage through the existing runtime injection point when OS permissions are not portable, and full-package `go test ./...` validation. Out of scope: adding query traversal assertions, defining query-specific JSON fields, invoking language indexers, committing generated binary indexes, default index lookup, caching, watching, compilation, type-checking, install/version behavior, and README changes.

**spec_ref:** specs/stories/readme/20260517-095535-epic-planning-1/shared-runtime-error-cases.md

**Depends on:** Task 2.

**task_depends_on:** `epic-planning-1-architecture-code-planning-0`

**Planned files:**
- `cmd/scip-search/main_test.go`
- `cmd/scip-search/testhelpers_test.go` if helper extraction is needed

**Implementation notes:**
- Keep successful result assertions generic: parseable JSON, empty stderr, status `0`, and evidence the handler boundary received loaded context. Do not assert symbol/package/reference/implementation result schemas.
- Do not commit generated `.scip` files; create them in temp directories during tests from official SCIP binding structs.
- Validation command for the task: `go test ./...`.

### Task 4 - Document index-loading failures and loaded-index boundary

**desc:** `scip-search` README documents the caller-selected SCIP index loading contract, including status `3` failure cases and the official binding boundary, without documenting traversal behavior or query-specific result schemas.

**done_when:** README changes document that every documented query command reads only the caller-supplied `--index` file, nonexistent, unreadable, directory, and invalid SCIP inputs fail before query execution with empty stdout, diagnostic stderr, and status `3`, valid SCIP inputs are loaded through the official SCIP Go bindings before the selected query handler runs, and `scip-search` never generates, searches for, updates, deletes, caches, watches, compiles, type-checks, or parses custom index formats for the selected index; no query-specific result fields, traversal semantics, install/version behavior, or ctags fallback behavior are added; documentation pre-commit checks pass for `README.md`.

**scope:** In scope: README user-facing notes for explicit index loading, status `3` shared loading failures, official SCIP Go binding load boundary, and excluded index side effects. Out of scope: code changes, query-specific result schema documentation, symbol/package/reference/implementation traversal docs, install/version/release docs, ctags fallback docs, and language indexer installation guidance beyond existing README content.

**spec_ref:** README.md#what-is-scip-search

**Depends on:** Task 3.

**task_depends_on:** `epic-planning-1-architecture-code-planning-0`

**Planned files:**
- `README.md`

**Implementation notes:**
- Place the new notes near the existing command examples and runtime description so automation users see the stream/status contract before installation details.
- Validation command for the task: `pre-commit run --files README.md`.

## Dependency Plan

All tasks depend on the prior shared runtime shell plan, `epic-planning-1-architecture-code-planning-0`. Task 1 can run first because it owns the production loader package and dependency updates. Task 2 depends on Task 1 because it wires the real loader result and failure type into the existing CLI runtime. Task 3 depends on Task 2 because end-to-end coverage needs the production runtime wiring. Task 4 depends on Task 3 so user-facing documentation describes implemented and validated behavior.

If the concrete runtime code from `epic-planning-1-architecture-code-planning-0` has not been implemented when these coding tasks start, coders should mark the relevant task blocked rather than expanding this plan to own the shared runtime shell.

## Shared-File Audit

| File or package | Tasks | Dependency |
|-----------------|-------|------------|
| `go.mod`, `go.sum` | Task 1 | Single owner |
| `internal/scipindex/*` | Task 1 | Single owner |
| `internal/runtime/index.go` | Task 1 | Single owner |
| `internal/cli/runtime.go`, `internal/cli/runtime_test.go` | Task 2 | Single owner in this plan; cross-plan dependency on `epic-planning-1-architecture-code-planning-0` |
| `cmd/scip-search/main_test.go`, `cmd/scip-search/testhelpers_test.go` | Task 3 | Single owner in this plan; cross-plan dependency on `epic-planning-1-architecture-code-planning-0` |
| `README.md` | Task 4 | Single owner |

No two sibling tasks in this plan modify the same file without a dependency chain.

## Test Impact

Task 1 adds loader unit tests for path validation, unreadable-open injection, invalid SCIP bytes, and valid SCIP protobuf parsing through official bindings. Task 2 adds CLI unit tests for status `3`, stderr-only diagnostics, handler suppression on loader failure, and loaded-context handoff. Task 3 adds end-to-end command coverage across `symbols`, `references`, `implementations`, and `packages`, plus `go test ./...`. Task 4 runs documentation pre-commit checks only.

## Doc Impact

Task 4 updates README because this plan changes user-visible failure behavior for selected index files and documents the official SCIP loading boundary. No separate spec update is planned because the behavior is already specified in the epic, architecture plan, and story documents cited above.

## Spec Compliance Matrix

| # | Requirement | Source | Task(s) | Status |
|---|-------------|--------|---------|--------|
| 1 | Loading starts only after shared command and `--index` invocation validation succeeds. | Architecture `CLI Runtime -> SCIP Index Loader`; index-path-validation context | Task 2, Task 3 | Covered |
| 2 | The caller-selected `--index <index-path>` is the only index input for the current invocation. | scip-index-loading-boundary AC-001-2; architecture Scope 2 | Task 1, Task 2, Task 3, Task 4 | Covered |
| 3 | A nonexistent selected index path fails before query traversal. | index-path-validation AC-001-1 | Task 1, Task 2, Task 3, Task 4 | Covered |
| 4 | Missing selected index paths do not trigger default-location search, repository-local inference, index generation, or index updates. | index-path-validation AC-001-2 | Task 1, Task 2, Task 4 | Covered |
| 5 | A selected path that cannot be opened for reading fails before query traversal. | index-path-validation AC-002-1 | Task 1, Task 2, Task 3, Task 4 | Covered |
| 6 | Directory or other non-file selected inputs fail as index-loading failures before traversal. | index-path-validation AC-002-2; stderr-exit-error-contract AC-002-3 | Task 1, Task 2, Task 3, Task 4 | Covered |
| 7 | Unreadable or invalid selected index paths are not generated, rewritten, deleted, or otherwise mutated. | index-path-validation AC-002-3; scip-index-loading-boundary AC-002-4 | Task 1, Task 2, Task 4 | Covered |
| 8 | Readable valid SCIP input is loaded through the official SCIP Go binding boundary. | README.md#language-support; scip-index-loading-boundary AC-001-1 | Task 1, Task 3, Task 4 | Covered |
| 9 | Loaded SCIP index data reaches the selected query execution boundary. | scip-index-loading-boundary AC-001-1 | Task 2, Task 3 | Covered |
| 10 | Loading success does not itself define query-specific result fields or traversal behavior. | scip-index-loading-boundary AC-001-4; stdout-json-success-contract Out of Scope | Task 1, Task 2, Task 3, Task 4 | Covered |
| 11 | A readable invalid SCIP input file fails before query traversal. | scip-index-loading-boundary AC-002-1 | Task 1, Task 2, Task 3, Task 4 | Covered |
| 12 | Invalid SCIP input does not produce a successful query result or empty-result success. | scip-index-loading-boundary AC-002-2; shared-runtime-error-cases AC-002-5 | Task 2, Task 3 | Covered |
| 13 | Invalid SCIP input does not trigger custom-format parsing or ctags fallback behavior. | scip-index-loading-boundary AC-002-3 | Task 1, Task 4 | Covered |
| 14 | Shared index-loading failures write no stdout and diagnostic stderr. | stderr-exit-error-contract AC-001-2 and AC-001-4 | Task 2, Task 3, Task 4 | Covered |
| 15 | Nonexistent, unreadable, directory, and invalid SCIP selected index failures exit with status `3`. | stderr-exit-error-contract AC-002-3 | Task 2, Task 3, Task 4 | Covered |
| 16 | The same index-loading failure class uses the same status across documented commands. | stderr-exit-error-contract AC-002-4 | Task 2, Task 3 | Covered |
| 17 | `symbols` covers nonexistent, unreadable, directory, and invalid selected index failures before traversal. | shared-runtime-error-cases AC-002-1 | Task 2, Task 3 | Covered |
| 18 | `references` covers nonexistent, unreadable, directory, and invalid selected index failures before traversal. | shared-runtime-error-cases AC-002-2 | Task 2, Task 3 | Covered |
| 19 | `implementations` covers nonexistent, unreadable, directory, and invalid selected index failures before traversal. | shared-runtime-error-cases AC-002-3 | Task 2, Task 3 | Covered |
| 20 | `packages` covers nonexistent, unreadable, directory, and invalid selected index failures before traversal. | shared-runtime-error-cases AC-002-4 | Task 2, Task 3 | Covered |
| 21 | Runtime does not compile, type-check, generate indexes, update indexes, start a watcher, start a daemon, cache indexes, or require a custom index format. | scip-index-loading-boundary NFR-000-2; assigned scope | Task 1, Task 2, Task 3, Task 4 | Covered |
| 22 | Selected index files are never searched for, generated, updated, deleted, cached, watched, or parsed as a custom format. | Assigned done_when and scope | Task 1, Task 2, Task 4 | Covered |
| 23 | Official SCIP loading is separate from sibling traversal work for documents, occurrences, symbols, relationships, ranges, and hover data. | Epic Out of Scope; scip-index-loading-boundary Out of Scope | Task 1, Task 2, Task 3, Task 4 | Covered |
| 24 | Query-specific flags and schemas remain out of this loading-boundary plan. | Assigned scope; shared-runtime-error-cases Out of Scope | Task 1, Task 2, Task 3, Task 4 | Covered |
| E2E | e2e test coverage for new behavior | Cross-cutting | Task 3 | Covered |
| DOC | Documentation updates for changed behavior | Cross-cutting | Task 4 | Covered |

## Pre-Submit Validation Checklist

- Re-read this plan and verify the output JSON fields are character-identical to each task's `desc`, `done_when`, `scope`, and `spec_ref`.
- Run `jq . specs/plans/readme/20260517-192724-epic-planning-1-architecture-code-planning-1-output.json`.
- Search this plan for every `Task N` cross-reference and confirm the referenced task states the corresponding responsibility or dependency.
- Confirm every shared file listed in more than one task has a dependency chain.
- Run pre-commit on the plan and output JSON files.
- Commit only the plan and output JSON artifacts.
