# Code Plan: CLI Version Build Identity

Status: draft

## Source Context

Based on:
- `README.md#installation`
- `README.md#what-is-scip-search`
- `specs/epics/readme/20260517-151311-epic-planning-5.md#capability-cap-003---report-cli-version-information`
- `specs/stories/readme/20260517-151311-epic-planning-5/CAP-003-01-version-output-contract.md`
- `specs/stories/readme/20260517-151311-epic-planning-5/CAP-003-02-build-metadata-for-release-and-source-builds.md`
- Architecture reference `specs/arch-plan/readme/20260517-170615-epic-planning-5-architecture.md#scope-1-version-and-build-identity` read from merge commit `25dc06d175819fe6c79d9faa5ff8f6b0ce43c811` because the file is referenced by blackboard state but absent from this worktree checkout.
- Active shared CLI runtime plan `epic-planning-1-architecture-code-planning-0`, which owns the Go module baseline, `cmd/scip-search` entrypoint, documented query routing, shared `--index` contract, stdout/stderr/status behavior, and injected loader/query handler boundary.
- Current worktree file discovery: no checked-in `go.mod`, `cmd/`, `internal/`, `Makefile`, or `tests/` implementation paths were present at planning time.

## Planning Boundary

This plan covers only the top-level `scip-search --version` invocation and the offline build identity contract that makes version output distinguish release builds from source builds.

Out of scope: query command JSON schemas, query runtime error taxonomy, query-specific flags, SCIP index path validation, SCIP data loading, traversal, network release lookups, installer workflow selection, source-install workflow selection, README documentation, packaging validation, release hosting setup, signing, package managers, and language indexer installation.

## Architectural Direction

Create a narrow version boundary for build identity and route `--version` before any query command validation or index loading path. The version boundary should expose a deterministic formatter for installed build identity and keep the metadata source local to the binary, using implementation-local build variables or equivalent offline data supplied by release/source build tooling.

The CLI should treat `--version` as distribution verification output, not as a query result. It must write human-readable stdout that includes `scip-search`, identifies whether the binary is release-built or source-built, and includes the relevant release identity or source provenance available to the build. It must exit `0` with empty stderr, even when no query command or `--index` is present.

`epic-planning-1-architecture-code-planning-0` owns creation of the shared CLI runtime shell and `cmd/scip-search` entrypoint. This plan depends on that task and extends its top-level pre-routing path so version handling and query routing do not create competing executable entrypoints. If the shared runtime shell has not landed in the coding checkout, the coder must block rather than create a parallel `cmd/scip-search` ownership path.

The implementation may introduce internal version-package files needed for this behavior after the shared runtime shell exists. It must not implement query routing, query handlers, SCIP parsing, network lookup, installer behavior, or README documentation in this scope.

## Planned Coding Tasks

### Task 1 - Implement offline version identity and top-level `--version`

**desc:** `scip-search` users can run top-level `scip-search --version` and receive offline build identity output that identifies `scip-search`, distinguishes release builds from source builds, and bypasses query command and index validation.

**done_when:** Unit tests for the CLI/version boundary pass with release and source build metadata cases, asserting `scip-search --version` without a query command or `--index` exits `0`, writes non-empty stdout containing `scip-search` and the supplied release identity or source provenance, writes empty stderr, identifies release builds differently from source builds, does not emit query JSON, and does not call any SCIP index loader, query handler, or network lookup boundary.

**scope:** In scope: extending the shared CLI runtime pre-routing path from `epic-planning-1-architecture-code-planning-0` with top-level `--version` flag handling before query validation, an internal build identity value/formatter, implementation-local build metadata variables or equivalent offline metadata boundary, release-vs-source provenance rendering, and colocated unit tests for stream/status behavior and no-loader/no-network guarantees. Out of scope: creating a competing `cmd/scip-search` entrypoint or separate runtime shell, query command routing beyond preserving `--version` precedence, query result schemas, shared runtime error taxonomy, SCIP index parsing/loading, traversal, installer workflow selection, source install workflow selection, README documentation, packaging validation, release hosting, signing, package managers, and language indexer installation.

**spec_ref:** README.md#installation

**task_depends_on:** `epic-planning-1-architecture-code-planning-0`

**Planned files:**
- Shared runtime entrypoint file owned by `epic-planning-1-architecture-code-planning-0` and extended here for pre-routing `--version` behavior, expected path `cmd/scip-search/main.go`.
- `internal/version/version.go`
- `internal/version/version_test.go`
- CLI/runtime unit test file colocated with the shared runtime shell, expected path `cmd/scip-search/main_test.go` or equivalent existing runtime test path.

**Implementation notes:**
- Consume the shared runtime shell from `epic-planning-1-architecture-code-planning-0`; do not create an alternate binary entrypoint.
- Parse and satisfy `--version` before enforcing a query command or `--index`.
- Keep version output out of the query JSON writer path.
- Keep build identity lookup offline and deterministic; do not call HTTP, release APIs, git remotes, SCIP loaders, or language indexers at runtime.
- Model build provenance as explicit release/source identity rather than inferring from install path.
- If the shared runtime checkout path or injected loader/query handler boundary differs from the expected path names above, adapt to the merged runtime structure while preserving the dependency and ownership boundary.
- If the checkout still lacks the shared CLI runtime shell, mark the coding task blocked rather than broadening into unrelated query runtime setup.
- Validation command for the task: `go test ./cmd/scip-search ./internal/version`.

### Task 2 - Add no-index `--version` executable smoke coverage

**desc:** `scip-search` maintainers can validate the built executable's `--version` behavior end to end without a SCIP index, query command, network access, or installer workflow.

**done_when:** End-to-end tests build or invoke controlled release-identity and source-provenance `scip-search` executables, run each as `scip-search --version` from a temporary directory with no selected SCIP index, and pass while asserting exit status `0`, empty stderr, non-empty stdout containing `scip-search`, release output that includes the controlled release identity, source output that includes controlled source provenance and does not masquerade as a release, and no query command, SCIP index fixture, language indexer, installer script, or remote lookup is required.

**scope:** In scope: distribution-scoped executable smoke tests for `--version`, controlled release/source build metadata inputs, no-index working directory setup, stdout/stderr/status assertions, and test isolation from query fixtures and installer workflows. Out of scope: implementing version formatting or CLI routing, query command behavior, SCIP fixture generation, traversal validation, installer release/source workflow tests, README documentation, packaging validation scripts, hosted release lookup, and language indexer execution.

**spec_ref:** README.md#installation

**Planned files:**
- `tests/e2e/version_test.go`

**Implementation notes:**
- Build controlled test binaries with local metadata inputs, such as Go linker variables if Task 1 exposes them, without requiring hosted release artifacts.
- Run from an empty temporary directory to prove `--index` and default index discovery are not involved.
- Assert observable process behavior through the executable boundary rather than calling internal version helpers directly.
- Do not use the installer or `make install`; installer and packaging validation are sibling scopes.
- Validation command for the task: `go test ./tests/e2e`.

## Dependency Plan

Task 1 has an external `task_depends_on` relationship on `epic-planning-1-architecture-code-planning-0` because that active shared CLI runtime plan owns the Go module baseline, `cmd/scip-search` entrypoint, and injected loader/query handler boundary. This task extends that runtime's top-level pre-routing path for `--version` instead of owning a second entrypoint.

Task 2 depends on Task 1 because executable smoke coverage needs the version boundary and build metadata inputs implemented first.

No task in this plan depends on installer or README work. Sibling installer/source/documentation tasks depend on the version contract from this plan, not the other way around.

## Shared-File Audit

Task 1 may modify the shared CLI entrypoint file expected at `cmd/scip-search/main.go`, but that file is externally owned by `epic-planning-1-architecture-code-planning-0`. The explicit `task_depends_on` relationship serializes this plan behind the shared runtime plan and makes Task 1 an extension of the shared pre-routing path.

Task 1 owns internal version implementation files. Task 2 owns only e2e test files and depends on Task 1 to consume its build metadata contract.

No file is modified by both Task 1 and Task 2, so no sibling dependency is needed for shared-file conflict prevention. The Task 2 dependency exists for behavioral ordering, not shared-file serialization.

If a coder sees another merged task has moved the runtime entrypoint to a different file, Task 1 should follow that merged extension point. If no coherent shared runtime extension point exists, the coder should block for orchestration rather than creating a parallel binary entrypoint.

## Test Impact

Task 1 adds unit tests for version identity formatting, top-level stream/status behavior, release/source distinction, and no call into loader/query/network boundaries.

Task 2 adds executable smoke coverage for `scip-search --version` from an index-free directory with controlled release and source metadata.

## Doc Impact

No README or user-facing documentation update is planned in this task because the assigned scope explicitly excludes README documentation and sibling `epic-planning-5-architecture-code-planning-3` owns documentation and packaging validation alignment.

## Spec Compliance Matrix

| # | Requirement | Source | Task(s) | Status |
|---|-------------|--------|---------|--------|
| 1 | Accept `scip-search --version` as a top-level invocation without a query command. | `CAP-003-01-version-output-contract.md` ST-001 AC-001-1; architecture Scope 1 | Task 1, Task 2 | Covered |
| 2 | Accept `scip-search --version` without `--index` and avoid missing-index usage failure. | `CAP-003-01-version-output-contract.md` ST-001 AC-001-2; assigned done_when | Task 1, Task 2 | Covered |
| 3 | Complete version verification without attempting to load or validate a SCIP index file. | `CAP-003-01-version-output-contract.md` ST-001 AC-001-3; assigned done_when | Task 1, Task 2 | Covered |
| 4 | Exit with status `0` for valid `scip-search --version`. | `CAP-003-01-version-output-contract.md` ST-001 AC-001-4; assigned done_when | Task 1, Task 2 | Covered |
| 5 | Write non-empty stdout that identifies the executable as `scip-search`. | `CAP-003-01-version-output-contract.md` ST-002 AC-002-1; assigned done_when | Task 1, Task 2 | Covered |
| 6 | Include the installed build identity exposed by the binary in version stdout. | `CAP-003-01-version-output-contract.md` ST-002 AC-002-2; architecture CLI Version Surface | Task 1, Task 2 | Covered |
| 7 | Leave stderr empty for successful `--version`. | `CAP-003-01-version-output-contract.md` ST-002 AC-002-3; assigned done_when | Task 1, Task 2 | Covered |
| 8 | Keep version output separate from query result output for `symbols`, `packages`, `references`, and `implementations`. | `CAP-003-01-version-output-contract.md` ST-002 AC-002-4; architecture constraints | Task 1 | Covered |
| 9 | Provide offline build provenance from the installed binary without network lookup during `--version`. | `CAP-003-02-build-metadata-for-release-and-source-builds.md` NFR-000-1; ST-001 AC-001-3; assigned done_when | Task 1, Task 2 | Covered |
| 10 | Support automation comparing observed version output with the install workflow that produced the binary. | `CAP-003-02-build-metadata-for-release-and-source-builds.md` NFR-000-2 | Task 1, Task 2 | Covered |
| 11 | Identify latest-release installed binaries as release builds. | `CAP-003-02-build-metadata-for-release-and-source-builds.md` ST-001 AC-001-1 | Task 1, Task 2 | Covered |
| 12 | Include an explicit requested release identity for `VERSION=<release>` builds. | `CAP-003-02-build-metadata-for-release-and-source-builds.md` ST-001 AC-001-2 | Task 1, Task 2 | Covered |
| 13 | Report executable build identity independently of custom install directory path. | `CAP-003-02-build-metadata-for-release-and-source-builds.md` ST-001 AC-001-3b | Task 1, Task 2 | Covered |
| 14 | Identify branch-built executables as source builds rather than released binaries. | `CAP-003-02-build-metadata-for-release-and-source-builds.md` ST-002 AC-002-1 | Task 1, Task 2 | Covered |
| 15 | Include source provenance available from local `make install` builds. | `CAP-003-02-build-metadata-for-release-and-source-builds.md` ST-002 AC-002-2 | Task 1, Task 2 | Covered |
| 16 | Prevent source-built executables without release identity from masquerading as released binaries. | `CAP-003-02-build-metadata-for-release-and-source-builds.md` ST-002 AC-002-3 | Task 1, Task 2 | Covered |
| 17 | Make source-built and release-installed executables distinguishable through version output. | `CAP-003-02-build-metadata-for-release-and-source-builds.md` ST-002 AC-002-4; assigned done_when | Task 1, Task 2 | Covered |
| 18 | Do not require SCIP language indexer execution, compilation, type-checking, daemon startup, watch mode, health checks with SCIP indexes, or query execution for version verification. | `CAP-003-01-version-output-contract.md` NFR-000-2; epic CAP-003 out of scope | Task 1, Task 2 | Covered |
| 19 | Keep installer workflow selection, release install behavior, source install behavior, packaging validation, and README documentation outside this version implementation scope. | assigned scope; architecture Scope 1 boundary | Task 1, Task 2 | Covered |
| 20 | Avoid competing ownership of the shared `cmd/scip-search` entrypoint and consume the shared runtime pre-routing extension point. | Rejection feedback for `20260517-190045-epic-planning-5-architecture-code-planning-0.md`; active task `epic-planning-1-architecture-code-planning-0` scope | Task 1 | Covered |
| E2E | e2e test coverage for new behavior | Cross-cutting | Task 2 | Covered |
| DOC | Documentation updates for changed behavior | Cross-cutting | N/A: README documentation is explicitly excluded from this task and assigned to sibling `epic-planning-5-architecture-code-planning-3` | N/A |

## Pre-Submit Validation Checklist

- Re-read this plan and verify the output JSON fields are character-identical to each task's `desc`, `done_when`, `scope`, and `spec_ref`.
- Run `jq . specs/plans/readme/20260517-190924-epic-planning-5-architecture-code-planning-0-output.json`.
- Search this plan for `Task 1`, `Task 2`, `epic-planning-1-architecture-code-planning-0`, and `cmd/scip-search/main.go` references and confirm each responsibility, dependency, and exclusion is stated by the referenced task.
- Run pre-commit on the plan and output JSON files.
- Commit only the plan and output JSON artifacts.
