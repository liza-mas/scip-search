# Code Plan: Shared CLI Runtime Shell

Status: draft

## Source Context

Based on:
- `README.md#what-is-scip-search`
- `README.md#out-of-scope`
- `specs/epics/readme/20260517-095535-epic-planning-1.md`
- `specs/arch-plan/readme/20260517-170052-epic-planning-1-architecture.md`
- `specs/stories/readme/20260517-095535-epic-planning-1/cli-command-routing-and-usage.md`
- `specs/stories/readme/20260517-095535-epic-planning-1/cli-shared-index-flag.md`
- `specs/stories/readme/20260517-095535-epic-planning-1/cli-one-shot-process-lifecycle.md`
- `specs/stories/readme/20260517-095535-epic-planning-1/stdout-json-success-contract.md`
- `specs/stories/readme/20260517-095535-epic-planning-1/stderr-exit-error-contract.md`
- `specs/stories/readme/20260517-095535-epic-planning-1/shared-runtime-error-cases.md`
- Blackboard task `repo-go-module-baseline-coding-0`, which is MERGED and establishes the existing `scip-search` Go module baseline.

## Planning Boundary

This plan covers the shared runtime shell for the four documented query commands: `symbols`, `references`, `implementations`, and `packages`.

In scope: runtime status constants, JSON success writer, stderr-only diagnostics, shared command routing, required shared `--index` parsing, one-shot loader-to-handler orchestration with injected fakes, process entrypoint wiring, black-box invocation coverage, and user-facing documentation for the shared runtime contract.

Out of scope: official SCIP parsing, filesystem index validation beyond invoking the loader boundary, query-specific flags, symbol/package/reference/implementation traversal, query-specific result fields, default index discovery, caching, daemon/watch/MCP/UI behavior, version/install/release behavior, and custom index formats.

## Architectural Direction

Keep the process boundary thin and move all observable runtime behavior behind testable internal packages:

```text
cmd/scip-search
  -> internal/cli
       -> internal/runtime
       -> injected loader boundary
       -> injected query handler boundary
```

`internal/runtime` should own status codes and stream discipline. `internal/cli` should own command selection, shared `--index` parsing, loader/handler invocation order, and orchestration. `cmd/scip-search` should only adapt `os.Args`, stdout/stderr, and returned status to process exit behavior.

The loader and handlers are injected boundaries in this scope. Coders should use fakes to prove the runtime calls the loader and selected handler exactly once for valid routed invocations. They must not add a production SCIP traversal, result schema, or filesystem parser in this plan.

## Planned Coding Tasks

### Task 1 - Define shared runtime stream and status contracts

**desc:** `scip-search` has shared runtime status constants, stderr-only diagnostic helpers, and a JSON success writer that encode one marshalable handler result without defining query result fields.

**done_when:** Unit tests in `internal/runtime` pass asserting `StatusOK` is `0`, `StatusUsage` is `2`, and `StatusIndexLoad` is `3`; usage and index-loading diagnostics write only to a provided stderr writer while leaving stdout untouched; the success writer emits exactly one parseable JSON value plus a trailing newline to stdout, leaves stderr empty, returns status `0`, and exposes no command-specific result struct fields.

**scope:** In scope: `internal/runtime` status constants, shared failure classification/types, diagnostic writer behavior, JSON success writer behavior, and colocated unit tests with anonymous map/struct results. Out of scope: command routing, flag parsing, process exits, filesystem index validation, official SCIP parsing, query traversal, command-specific result schemas, installer/version behavior, logging configuration, and docs.

**spec_ref:** specs/stories/readme/20260517-095535-epic-planning-1/stdout-json-success-contract.md

**Planned files:**
- `internal/runtime/status.go`
- `internal/runtime/streams.go`
- `internal/runtime/streams_test.go`

**Implementation notes:**
- Use only Go standard library JSON and writer APIs.
- Keep result encoding generic over any marshalable value.
- Keep diagnostic text human-readable on stderr; do not introduce a JSON error schema.
- Validation command for the task: `go test ./internal/runtime`.

### Task 2 - Route documented commands through one shared CLI path

**desc:** `scip-search` has one shared CLI router that recognizes `symbols`, `references`, `implementations`, and `packages` and rejects missing or unsupported command names as usage failures before loader or handler execution.

**done_when:** Unit tests in `internal/cli` pass asserting `symbols`, `references`, `implementations`, and `packages` all route through the same command registry path; root invocation with no command and unsupported commands with plausible flags return status `2`, write a diagnostic only to stderr, leave stdout empty, and do not call the injected loader or any query handler.

**scope:** In scope: `internal/cli` command registry/router, missing-command and unsupported-command usage failure handling, injected loader/handler test doubles, and tests proving no loader or handler calls occur before a selected supported command exists. Out of scope: shared `--index` parsing after command selection, filesystem index validation, official SCIP parsing, query-specific flags, query traversal, result schemas, process-level `os.Exit`, version/install behavior, and docs.

**spec_ref:** specs/stories/readme/20260517-095535-epic-planning-1/cli-command-routing-and-usage.md

**Depends on:** Task 1.

**Planned files:**
- `internal/cli/runtime.go`
- `internal/cli/runtime_test.go`

**Implementation notes:**
- Represent the supported command set once, not as separate switch copies per behavior.
- Keep unsupported command failures in the usage class even when callers pass plausible flags.
- Validation command for the task: `go test ./internal/cli`.

### Task 3 - Enforce shared index flag and injected one-shot execution

**desc:** `scip-search` has shared `--index` parsing and one-shot runtime orchestration that passes the selected index path and opaque query arguments to exactly one injected loader and the selected query handler.

**done_when:** Unit tests in `internal/cli` pass for every documented command asserting missing `--index` and `--index` without a value return status `2` through stderr only; a valid invocation with `--index <path>` calls the injected loader exactly once with `<path>`, calls only the selected handler exactly once with the loaded context and remaining opaque arguments, rejects an obvious second documented command token as a usage failure, emits the handler result as one parseable JSON value on stdout with empty stderr and status `0`, and introduces no SCIP traversal, filesystem path validation, or command-specific result fields.

**scope:** In scope: `internal/cli` shared flag parser, invocation-local selected index path, one-shot loader-to-handler orchestration, opaque query-argument forwarding, additional documented-command token guard where it can be identified by shared parsing, success JSON emission through `internal/runtime`, and unit tests with fakes. Out of scope: query-specific validation for `--name`, `--symbol`, or `--prefix`, filesystem index validation beyond invoking the loader boundary, official SCIP parsing, traversal, command-specific schemas, default index locations, caching, daemon/watch/MCP/UI behavior, version/install behavior, and docs.

**spec_ref:** specs/stories/readme/20260517-095535-epic-planning-1/cli-shared-index-flag.md

**Depends on:** Task 1 and Task 2.

**Planned files:**
- `internal/cli/runtime.go`
- `internal/cli/runtime_test.go`

**Implementation notes:**
- Treat query-specific arguments as opaque after shared parsing; do not require `--name`, `--symbol`, or `--prefix`.
- The selected index path must not persist outside the `Run` call.
- The loader should be invoked only after command and shared flag validation pass.
- The selected handler should be invoked only after the loader returns a loaded context.
- Validation command for the task: `go test ./internal/cli`.

### Task 4 - Wire the process entrypoint and black-box invocation tests

**desc:** `scip-search` command entrypoint delegates process IO and exit status to the shared runtime path and has black-box coverage for user-observable shared invocation behavior.

**done_when:** Tests for `cmd/scip-search` pass by executing or invoking the command entrypoint with real stdout/stderr buffers and asserting no-command, unsupported-command, and each documented command missing `--index` exit with status `2`, stdout empty, and stderr non-empty; the entrypoint delegates to `internal/cli` instead of printing or encoding independently; no background loop, prompt, daemon, watcher, index generation, compilation, type-checking, or query traversal code is introduced; `go test ./...` passes.

**scope:** In scope: `cmd/scip-search` entrypoint wiring, testable main/run adapter if needed, process stream/status tests for shared invocation failures, and `go test ./...` validation. Out of scope: successful production query execution before the real index loader exists, filesystem index validation, official SCIP parsing, query-specific flags, query traversal, result schemas, install/version behavior, and README changes.

**spec_ref:** README.md#what-is-scip-search

**Depends on:** Task 3.

**Planned files:**
- `cmd/scip-search/main.go`
- `cmd/scip-search/main_test.go`

**Implementation notes:**
- Keep `main` thin enough that stream/status behavior is testable without calling `os.Exit` in unit tests.
- Do not implement a placeholder production loader that pretends arbitrary paths are valid SCIP indexes.
- Validation command for the task: `go test ./...`.

### Task 5 - Document the shared runtime shell behavior

**desc:** `scip-search` README documents the shared query runtime failure and stream contract without adding query traversal or result schema details.

**done_when:** README changes document the shared `--index` requirement, success stdout as one JSON value with empty stderr and status `0`, shared usage failures for missing/unsupported command or missing `--index` as stderr-only status `2`, and note that index-loading failures are owned by the subsequent loader boundary; no query-specific result fields, install/version changes, or traversal behavior are documented; markdown/pre-commit checks pass for the touched docs.

**scope:** In scope: README user-facing command runtime notes for shared invocation, streams, and statuses introduced by this runtime shell. Out of scope: query-specific result schemas, symbol/package/reference/implementation traversal semantics, official SCIP loader implementation details, install/version docs, release packaging, ctags fallback docs, and code changes.

**spec_ref:** README.md#what-is-scip-search

**Depends on:** Task 4.

**Planned files:**
- `README.md`

**Implementation notes:**
- Keep the documentation aligned with the behavior available after this runtime shell, and explicitly leave loader failure specifics to the loader-boundary task.
- Validation command for the task: `pre-commit run --files README.md`.

## Dependency Plan

Task 1 depends on the merged `repo-go-module-baseline-coding-0` task because it assumes `go test ./...` already runs in module mode. Task 2 depends on Task 1 because usage failures must reuse the shared status and diagnostic contract. Task 3 depends on Task 1 and Task 2 because shared `--index` parsing builds on selected-command routing and success/failure stream behavior. Task 4 depends on Task 3 because the process entrypoint should only adapt the completed internal runtime path. Task 5 depends on Task 4 so docs describe implemented user-visible behavior.

## Shared-File Audit

| File or package | Tasks | Dependency |
|-----------------|-------|------------|
| `internal/runtime/*` | Task 1 | Single owner |
| `internal/cli/runtime.go`, `internal/cli/runtime_test.go` | Task 2, Task 3 | Task 3 depends on Task 2 |
| `cmd/scip-search/main.go`, `cmd/scip-search/main_test.go` | Task 4 | Single owner |
| `README.md` | Task 5 | Single owner |

No two sibling tasks without a dependency modify the same files.

## Test Impact

Task 1 adds unit tests for runtime status and stream helpers. Task 2 adds unit tests for command routing and missing/unsupported command usage failures. Task 3 adds unit tests for shared `--index` parsing, one-shot loader/handler orchestration, and generic JSON success emission with fakes. Task 4 adds black-box entrypoint tests for observable shared invocation failures and runs `go test ./...`. Task 5 runs documentation pre-commit checks only.

## Doc Impact

Task 5 updates README because this plan introduces user-visible command invocation, stream, and status behavior. Query-specific result schema docs remain owned by sibling query tasks, and official SCIP loader failure details remain owned by the loader-boundary task.

## Spec Compliance Matrix

| # | Requirement | Source | Task(s) | Status |
|---|-------------|--------|---------|--------|
| 1 | The Go project baseline supports existing pre-commit Go hooks and `go test ./...` in module mode. | Assigned done_when; architecture Scope 1; merged task `repo-go-module-baseline-coding-0` | Task 1, Task 4 | Covered |
| 2 | The process exposes a CLI binary entrypoint named `scip-search`. | README.md#what-is-scip-search; architecture CLI Entrypoint | Task 4 | Covered |
| 3 | The documented query commands are exactly `symbols`, `references`, `implementations`, and `packages` for this runtime shell. | README.md#what-is-scip-search; cli-command-routing-and-usage AC-001-1 through AC-001-4 | Task 2 | Covered |
| 4 | Missing query command is a shared usage failure before loader or traversal work. | cli-command-routing-and-usage AC-002-1; shared-runtime-error-cases AC-001-1 | Task 2, Task 4 | Covered |
| 5 | Unsupported command names are shared usage failures even when plausible flags are present. | cli-command-routing-and-usage AC-002-2 and AC-002-2b; shared-runtime-error-cases AC-001-2 | Task 2, Task 4 | Covered |
| 6 | Every documented query command requires shared `--index`. | README.md#what-is-scip-search; cli-shared-index-flag AC-001-1 through AC-001-4 | Task 3, Task 4 | Covered |
| 7 | `--index` without a value is a usage failure before loader or traversal work. | cli-shared-index-flag AC-002-1b; shared-runtime-error-cases AC-001-3 through AC-001-6 | Task 3 | Covered |
| 8 | A supplied `--index <index-path>` becomes the selected index path for the current invocation and does not persist into later invocations. | cli-shared-index-flag AC-002-1 and AC-002-2; architecture CLI Runtime -> SCIP Index Loader invariant | Task 3 | Covered |
| 9 | Loader is not called until command routing and shared flag validation pass. | architecture CLI Runtime -> SCIP Index Loader invariant; assigned done_when | Task 2, Task 3 | Covered |
| 10 | Valid routed invocations call the injected loader exactly once with the selected path. | Assigned done_when; architecture CLI Runtime -> SCIP Index Loader | Task 3 | Covered |
| 11 | Valid routed invocations call only the selected query handler exactly once with loaded context and opaque command args. | Assigned done_when; architecture CLI Runtime -> Query Handler Boundary | Task 3 | Covered |
| 12 | One process invocation handles only one selected query command and exits after success or shared failure. | README.md#what-is-scip-search; cli-one-shot-process-lifecycle AC-001-1 and AC-001-2 | Task 3, Task 4 | Covered |
| 13 | Additional documented command tokens are not executed as a second query in the same process. | cli-one-shot-process-lifecycle AC-001-2b | Task 3 | Covered |
| 14 | Runtime does not prompt for missing input or start daemon, server, watcher, background indexing, compilation, or type-checking work. | README.md#out-of-scope; cli-one-shot-process-lifecycle AC-002-1 through AC-002-3 | Task 4 | Covered |
| 15 | Shared success status is `0`. | stdout-json-success-contract ASM-000-1; assigned done_when | Task 1, Task 3 | Covered |
| 16 | Shared usage failure status is `2`. | stderr-exit-error-contract AC-002-1 and AC-002-2; assigned done_when | Task 1, Task 2, Task 3, Task 4 | Covered |
| 17 | Shared index-loading failure status constant is available as `3` for the subsequent loader-boundary scope. | stderr-exit-error-contract AC-002-3; architecture Runtime Contracts | Task 1 | Covered |
| 18 | Successful handler results are emitted as exactly one parseable JSON value on stdout. | README.md#what-is-scip-search; stdout-json-success-contract AC-001-1 through AC-001-5; assigned done_when | Task 1, Task 3 | Covered |
| 19 | Successful invocations leave stderr empty. | stdout-json-success-contract AC-002-2; assigned done_when | Task 1, Task 3 | Covered |
| 20 | Successful stdout contains no progress text, warnings, prompts, logging, or explanatory text outside JSON. | stdout-json-success-contract AC-002-1; epic NFR-000-3 | Task 1, Task 3 | Covered |
| 21 | Shared usage failures write diagnostics only to stderr and no stdout. | stderr-exit-error-contract AC-001-1, AC-001-3, AC-001-4; assigned done_when | Task 1, Task 2, Task 3, Task 4 | Covered |
| 22 | Diagnostics are human-readable stderr text, not a structured stderr JSON schema. | stderr-exit-error-contract ASM-000-2 | Task 1 | Covered |
| 23 | Query-specific flags such as `--name`, `--symbol`, and `--prefix` are not validated in this shared runtime scope. | cli-command-routing-and-usage Out of Scope; shared-runtime-error-cases Out of Scope; assigned scope | Task 3 | Covered |
| 24 | No query-specific traversal or result schema is defined. | Assigned done_when and scope; stdout-json-success-contract Out of Scope | Task 1, Task 3, Task 4, Task 5 | Covered |
| 25 | Filesystem index validation and official SCIP parsing are not implemented beyond invoking the loader boundary. | Assigned scope; architecture Scope 2 owns loading boundary | Task 3, Task 4 | Covered |
| 26 | Version/install/release behavior remains out of this runtime shell. | Assigned scope; epic Out of Scope; README.md#installation owned by epic-planning-5 | Task 2, Task 3, Task 4, Task 5 | Covered |
| 27 | Documentation covers user-visible shared runtime behavior introduced by the plan. | Cross-cutting doc deliverable rule; README.md#what-is-scip-search | Task 5 | Covered |
| E2E | e2e test coverage for new behavior | Cross-cutting | Task 4 | Covered |
| DOC | Documentation updates for changed behavior | Cross-cutting | Task 5 | Covered |

## Pre-Submit Validation Checklist

- Re-read this plan and verify the output JSON fields are character-identical to each task's `desc`, `done_when`, `scope`, and `spec_ref`.
- Run `jq . specs/plans/readme/20260517-191414-epic-planning-1-architecture-code-planning-0-output.json`.
- Search this plan for every `Task N` cross-reference and confirm the referenced task states the corresponding responsibility or dependency.
- Confirm every shared file listed in more than one task has a dependency chain.
- Run pre-commit on the plan and output JSON files.
- Commit only the plan and output JSON artifacts.
