# Code Plan: Source Install Workflows

Status: draft

## Source Context

Based on:
- `README.md#installation`
- `specs/epics/readme/20260517-151311-epic-planning-5.md#capability-cap-002---build-and-install-from-source-workflows`
- `specs/stories/readme/20260517-151311-epic-planning-5/CAP-002-01-branch-install-script.md`
- `specs/stories/readme/20260517-151311-epic-planning-5/CAP-002-02-local-clone-make-install.md`
- `specs/stories/readme/20260517-151311-epic-planning-5/CAP-003-02-build-metadata-for-release-and-source-builds.md`
- `specs/stories/readme/20260517-151311-epic-planning-5/CAP-004-02-packaging-and-release-validation.md`
- Architecture reference `specs/arch-plan/readme/20260517-170615-epic-planning-5-architecture.md#scope-3-source-install-workflows` read from merge commit `25dc06d175819fe6c79d9faa5ff8f6b0ce43c811` because the file is referenced by blackboard state but absent from this worktree checkout.
- Prior plan `specs/plans/readme/20260517-190924-epic-planning-5-architecture-code-planning-0.md`, which owns the offline `scip-search --version` build identity contract that source installs must populate.
- Prior plan `specs/plans/readme/20260517-193112-epic-planning-5-architecture-code-planning-1.md`, which owns release-mode `install.sh` selection and install-directory behavior that the `BRANCH` source workflow must extend without changing release semantics.
- Current worktree file discovery: `go.mod`, `Makefile`, and `cmd/scip-search` exist; `install.sh` is not present in this checkout even though prior release installer planning owns creating it.

## Planning Boundary

This plan covers only source install workflows: `BRANCH=<branch>` through `install.sh`, caller-provided Go and make prerequisite checks, local clone `make install`, source build/install output, source provenance handoff to `--version`, and source-install tests.

Out of scope: release binary download behavior, release artifact fallback, release artifact selection internals, cross-compilation, CI publication, package manager formulas, Go or make installation, language indexer installation, SCIP indexes, query command validation, README documentation, release-facing validation command documentation, and broad packaging drift checks.

## Architectural Direction

Keep local source build behavior in `Makefile` and have the branch installer workflow reuse that source install boundary rather than creating a separate release-like source installation path. `make install` should build the current checkout, install one executable `scip-search`, and supply source provenance to the existing version identity boundary from `epic-planning-5-architecture-code-planning-0`.

Extend the release installer entry point from `epic-planning-5-architecture-code-planning-1` with a distinct `BRANCH` workflow selector. When `BRANCH` is set, the installer must require caller-provided Go and make, retrieve the requested branch into a temporary source checkout, build/install from that source path, and never fall back to a release artifact if branch retrieval, build, or install fails.

Both source workflows should prove success through an installed executable and `--version` source provenance, not through query commands or SCIP indexes. If a coding checkout lacks the prior version contract or release installer extension point, coders should block rather than inventing parallel `--version` output or a second installer entry point.

## Planned Coding Tasks

### Task 1 - Implement local clone `make install` source build

**desc:** `scip-search` source builders can run `make install` from a local clone and install the current checkout as a runnable source-built `scip-search` executable with caller-provided Go and make.

**done_when:** Makefile/source-install tests pass asserting `make install` checks for caller-provided Go, builds the current checkout instead of selecting a release artifact, installs an executable `scip-search` at the make-selected install path, reports successful local source install only after the executable exists, passes source provenance into `scip-search --version`, verifies the installed executable from outside the clone directory, and fails nonzero without a success claim for missing Go, build failure, or install failure.

**scope:** In scope: `Makefile` source install target behavior, caller PATH Go prerequisite check inside the make-driven source workflow, current-checkout build command, install directory/path handling for local source installs, source provenance build metadata handoff to the existing version boundary, local source install diagnostics, and colocated tests for success, missing Go, build failure, install failure, executable mode, and outside-clone verification. Out of scope: release artifact lookup or download, `install.sh` `BRANCH` selection, missing make before make can run except documenting that it is a caller prerequisite, exact `--version` formatting beyond source provenance observability, query command execution, SCIP index loading, language indexer installation, README documentation, release validation docs, packaging-wide drift checks, Go installation, and make installation.

**spec_ref:** README.md#installation

**task_depends_on:** `epic-planning-5-architecture-code-planning-0-coding-0-after-runtime-shell`

**Planned files:**
- `Makefile`
- `tests/install/source_make_install_test.go` or equivalent local source install test file.

**Implementation notes:**
- Build from the working checkout that runs `make install`; do not fetch releases or inspect release metadata.
- Use the source provenance mechanism exposed by the version identity task, such as build variables or equivalent offline metadata inputs.
- Success output should name the installed executable path so automation can invoke it directly.
- Missing `make` is observable before the Makefile executes; this task should not attempt to install or provision make.
- If the version boundary from `epic-planning-5-architecture-code-planning-0` is unavailable, block rather than defining a second source provenance format.
- Validation command for the task: `go test ./tests/install` or the equivalent project test target that owns source make-install tests.

### Task 2 - Add `BRANCH` source workflow to `install.sh`

**desc:** `scip-search` source builders can run the installer with `BRANCH=<branch>` and receive an installed source-built executable from that requested branch or a nonzero actionable source-install failure.

**done_when:** Installer source-workflow tests pass asserting non-empty `BRANCH` selects a source build path requiring caller-provided Go and make, an available branch is fetched or checked out into controlled temporary source, the branch source is built and installed as executable `scip-search` through the shared source install boundary, success output identifies the installed executable path and branch/source provenance, the installed executable verifies with `--version` outside the source checkout, and missing Go, missing make, unavailable branch, build failure, or install failure exit nonzero without claiming success, provisioning host tools, or falling back to release artifacts.

**scope:** In scope: extending `install.sh` with `BRANCH` source workflow selection, source prerequisite checks for Go and make, branch retrieval/checkout failure handling, invocation of the shared source build/install path, temporary source checkout cleanup expectations, source install success output, source provenance handoff, and installer tests for successful branch install plus missing Go, missing make, unavailable branch, build failure, install failure, and release-fallback prevention. Out of scope: release-mode latest or `VERSION` selection except preserving existing behavior from `epic-planning-5-architecture-code-planning-1`, local clone `make install` internals except consuming the source install boundary from Task 1, hosted release setup, cross-compilation, package manager formulas, Go or make installation, language indexer installation, SCIP indexes, query command execution, README documentation, release validation command docs, and packaging-wide drift checks.

**spec_ref:** README.md#installation

**depends_on:** Task 1

**task_depends_on:** `epic-planning-5-architecture-code-planning-1-coding-1`

**Planned files:**
- `install.sh`
- `tests/install/source_branch_install_test.go` or equivalent installer source-workflow test file.

**Implementation notes:**
- Treat unset `BRANCH` as the release workflow owned by `epic-planning-5-architecture-code-planning-1`; this task should only add behavior when `BRANCH` is non-empty.
- Do not silently retry with a release artifact after branch, build, or install failure.
- Keep branch retrieval testable through local controlled repositories or overrides so tests do not require hosted release setup.
- Validate Go and make before reporting source install progress that implies a build can proceed.
- If `install.sh` from the release installer plan is absent in the coding checkout, block rather than creating a competing installer surface that ignores prior release semantics.
- Validation command for the task: `go test ./tests/install` or the equivalent project test target that owns installer source-workflow tests.

### Task 3 - Add source install executable smoke coverage

**desc:** `scip-search` maintainers can validate branch installer and local clone source installs end to end with controlled source checkouts using `--version` as the executable proof and no query or SCIP-index dependencies.

**done_when:** End-to-end source install tests pass from temporary directories with controlled source repositories, covering local clone `make install`, `BRANCH=<branch>` installer install, missing Go, missing make, unavailable branch, build failure, and install failure; success cases assert the installed executable exists, is executable, runs outside the source checkout with `--version`, and reports source provenance rather than release identity, while failure cases assert nonzero status, actionable diagnostics, no success claim, no host Go/make provisioning, no release artifact fallback, and no SCIP index, language indexer, or query command execution.

**scope:** In scope: black-box or integration tests for source installation through `make install` and `install.sh BRANCH`, controlled temporary source repositories, direct installed-executable `--version` verification outside the source checkout, failure-path process status and diagnostic assertions, and isolation from release artifacts, SCIP indexes, language indexers, and query commands. Out of scope: implementing Makefile or installer internals, release install e2e coverage, README drift checks or maintainer validation command docs, query fixtures, traversal fixtures, real language indexer execution, package manager validation, signing, notarization, and release approval policy.

**spec_ref:** README.md#installation

**depends_on:** Task 2

**Planned files:**
- `tests/e2e/source_install_test.go` or equivalent source install end-to-end test file.

**Implementation notes:**
- Exercise user-observable commands from temporary directories rather than only internal helper functions.
- Use controlled PATH/tool shims to prove missing Go and missing make failures without mutating the host environment.
- Use local controlled branch repositories or fixtures; do not require public network access or hosted release setup.
- Keep success proof to installed executable plus `--version`; do not run `symbols`, `packages`, `references`, or `implementations`.
- Validation command for the task: `go test ./tests/e2e` or the equivalent project test target that owns source installer e2e tests.

## Dependency Plan

Task 1 has an external `task_depends_on` relationship on `epic-planning-5-architecture-code-planning-0-coding-0-after-runtime-shell` because local source builds must hand source provenance to the existing offline `--version` boundary. If the version contract is not implemented in a coding checkout, Task 1 should block rather than inventing a separate provenance contract.

Task 2 depends on Task 1 because branch installs should consume the same source build/install boundary as local clone installs. Task 2 also has an external `task_depends_on` relationship on `epic-planning-5-architecture-code-planning-1-coding-1` because it modifies `install.sh` after the release installer workflow owns release selection and install-directory behavior.

Task 3 depends on Task 2 because end-to-end source install coverage needs both local source install tooling and the `BRANCH` installer workflow to exist.

Sibling `epic-planning-5-architecture-code-planning-3` owns README documentation, maintainer validation commands, packaging-wide tests, and docs drift checks. This plan only creates behavior-focused source install tests and e2e smoke coverage.

## Shared-File Audit

Task 1 modifies `Makefile` and source make-install tests.

Task 2 modifies `install.sh` and installer source-workflow tests. It depends externally on `epic-planning-5-architecture-code-planning-1-coding-1` because that prior release-installer implementation owns creating/extending `install.sh` for release behavior; this plan must extend the same file without changing release-mode semantics.

Task 3 owns only source install e2e coverage and depends on Task 2 for behavioral ordering. It should not modify `Makefile` or `install.sh`.

No file is modified by more than one sibling task in this plan. If coders consolidate source installer and source make-install tests into one shared file, Task 2's dependency on Task 1 and Task 3's dependency on Task 2 serialize the likely shared test edits.

## Test Impact

Task 1 adds local source `make install` tests for source build success, source provenance, executable mode/path, outside-clone verification, missing Go, build failure, and install failure.

Task 2 adds installer source workflow tests for `BRANCH` selection, Go/make prerequisites, branch retrieval failure, source build/install invocation, success output, source provenance, build/install failure, and release fallback prevention.

Task 3 adds executable end-to-end source install coverage for local clone and branch installer workflows plus source-install failure cases.

## Doc Impact

No README or maintainer documentation update is planned in this task because the assigned scope is source install behavior and source-install tests, while sibling `epic-planning-5-architecture-code-planning-3` owns README installation guidance, release-facing validation command documentation, packaging-wide tests, and docs drift checks.

## Spec Compliance Matrix

| # | Requirement | Source | Task(s) | Status |
|---|-------------|--------|---------|--------|
| 1 | `BRANCH=<branch>` installer workflow builds `scip-search` from source when Go and make are available and the branch exists. | `CAP-002-01-branch-install-script.md` ST-001 AC-001-1; assigned done_when | Task 2, Task 3 | Covered |
| 2 | Successful branch source install reports a successful source install only after an executable `scip-search` is present at the installer-selected binary location. | `CAP-002-01-branch-install-script.md` ST-001 AC-001-1, AC-001-2; architecture Scope 3 | Task 2, Task 3 | Covered |
| 3 | Branch-installed executable can be invoked for verification outside the source checkout and does not require the clone as current directory. | `CAP-002-01-branch-install-script.md` ST-001 AC-001-3; assigned done_when | Task 2, Task 3 | Covered |
| 4 | Non-empty `BRANCH` selects a source build workflow requiring Go and make rather than a release artifact download workflow. | `CAP-002-01-branch-install-script.md` ST-001 AC-001-4; architecture Installer Entry Point -> Source Build Tooling | Task 2, Task 3 | Covered |
| 5 | Missing Go in the branch installer path fails before claiming success and reports Go is required. | `CAP-002-01-branch-install-script.md` ST-001 AC-001-4b; assigned done_when | Task 2, Task 3 | Covered |
| 6 | Missing make in the branch installer path fails before claiming success and reports make is required. | `CAP-002-01-branch-install-script.md` ST-001 AC-001-4c; assigned done_when | Task 2, Task 3 | Covered |
| 7 | Unavailable branch fails nonzero with an actionable branch source-build error and no installed-binary success claim. | `CAP-002-01-branch-install-script.md` ST-001 AC-001-5; assigned done_when | Task 2, Task 3 | Covered |
| 8 | Branch source build or install failure after prerequisites pass fails nonzero without implying `scip-search` was installed. | `CAP-002-01-branch-install-script.md` ST-001 AC-001-6; assigned done_when | Task 2, Task 3 | Covered |
| 9 | Branch source installs do not provision Go, make, or language indexers on the caller host. | `CAP-002-01-branch-install-script.md` NFR-000-1; epic NFR-000-1 | Task 2, Task 3 | Covered |
| 10 | Branch source install behavior remains separate from release artifact selection and download behavior. | `CAP-002-01-branch-install-script.md` NFR-000-3; assigned scope exclusions | Task 2, Task 3 | Covered |
| 11 | Local clone `make install` builds `scip-search` from the current checkout when Go and make are available and reports a successful local source install. | `CAP-002-02-local-clone-make-install.md` ST-001 AC-001-1; assigned done_when | Task 1, Task 3 | Covered |
| 12 | Successful local clone `make install` produces an executable named `scip-search` at the make-selected binary location. | `CAP-002-02-local-clone-make-install.md` ST-001 AC-001-2; architecture Local Source Install Tooling | Task 1, Task 3 | Covered |
| 13 | Local clone installed executable can be invoked outside the clone directory. | `CAP-002-02-local-clone-make-install.md` ST-001 AC-001-3; assigned done_when | Task 1, Task 3 | Covered |
| 14 | Missing Go in local clone source install fails before claiming success and reports Go is required. | `CAP-002-02-local-clone-make-install.md` ST-001 AC-001-4; assigned done_when | Task 1, Task 3 | Covered |
| 15 | Missing make in the local clone workflow cannot proceed through `make install` and is not remediated by host tool provisioning from `scip-search`. | `CAP-002-02-local-clone-make-install.md` ST-001 AC-001-4b; assigned done_when | Task 3 | Covered |
| 16 | Local clone source build or install failure after prerequisites pass fails nonzero without implying `scip-search` was installed. | `CAP-002-02-local-clone-make-install.md` ST-001 AC-001-5; assigned done_when | Task 1, Task 3 | Covered |
| 17 | Local clone `make install` is a local source build workflow rather than a release artifact download workflow. | `CAP-002-02-local-clone-make-install.md` ST-001 AC-001-6; assigned scope exclusions | Task 1, Task 3 | Covered |
| 18 | Source-built binaries expose source build provenance through `scip-search --version` and do not masquerade as released binaries. | `CAP-003-02-build-metadata-for-release-and-source-builds.md` ST-002 AC-002-1, AC-002-2, AC-002-3, AC-002-4; assigned done_when | Task 1, Task 2, Task 3 | Covered |
| 19 | Source workflows use caller-provided Go and make and never install or provision those host tools. | Epic CAP-002 ASM-002-1; assigned done_when | Task 1, Task 2, Task 3 | Covered |
| 20 | Source install validation avoids query command execution, SCIP indexes, language indexers, query fixtures, and traversal validation. | `CAP-002-01-branch-install-script.md` out of scope; `CAP-002-02-local-clone-make-install.md` out of scope; `CAP-004-02-packaging-and-release-validation.md` NFR-000-2 | Task 1, Task 2, Task 3 | Covered |
| 21 | Source workflow failures are non-interactive, automation-friendly, nonzero, and actionable. | Epic NFR-000-2; assigned done_when | Task 1, Task 2, Task 3 | Covered |
| 22 | Source install implementation preserves prior release installer behavior and does not redefine release binary download behavior. | assigned scope; prior plan `epic-planning-5-architecture-code-planning-1` boundary | Task 2, Task 3 | Covered |
| E2E | e2e test coverage for new behavior | Cross-cutting | Task 3 | Covered |
| DOC | Documentation updates for changed behavior | Cross-cutting | N/A: README documentation and maintainer validation docs are explicitly assigned to sibling `epic-planning-5-architecture-code-planning-3` | N/A |

## Pre-Submit Validation Checklist

- Re-read this plan and verify the output JSON fields are character-identical to each task's `desc`, `done_when`, `scope`, and `spec_ref`.
- Run `jq . specs/plans/readme/20260517-194106-epic-planning-5-architecture-code-planning-2-output.json`.
- Search this plan for `Task 1`, `Task 2`, `Task 3`, `epic-planning-5-architecture-code-planning-0-coding-0-after-runtime-shell`, `epic-planning-5-architecture-code-planning-1-coding-1`, `epic-planning-5-architecture-code-planning-3`, `Makefile`, and `install.sh` references and confirm each responsibility, dependency, and exclusion is stated by the referenced task.
- Run pre-commit on the plan and output JSON files.
- Commit only the plan and output JSON artifacts.
