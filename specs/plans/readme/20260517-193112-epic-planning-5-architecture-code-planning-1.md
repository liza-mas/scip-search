# Code Plan: Release Installer Workflows

Status: draft

## Source Context

Based on:
- `README.md#installation`
- `specs/epics/readme/20260517-151311-epic-planning-5.md#capability-cap-001---install-released-scip-search-binaries`
- `specs/stories/readme/20260517-151311-epic-planning-5/CAP-001-01-release-install-script.md`
- `specs/stories/readme/20260517-151311-epic-planning-5/CAP-001-02-release-install-failure-and-verification.md`
- `specs/stories/readme/20260517-151311-epic-planning-5/CAP-003-01-version-output-contract.md`
- `specs/stories/readme/20260517-151311-epic-planning-5/CAP-003-02-build-metadata-for-release-and-source-builds.md`
- `specs/stories/readme/20260517-151311-epic-planning-5/CAP-004-02-packaging-and-release-validation.md`
- Architecture reference `specs/arch-plan/readme/20260517-170615-epic-planning-5-architecture.md#scope-2-release-installer-workflows` read from merge commit `25dc06d175819fe6c79d9faa5ff8f6b0ce43c811` because the file is referenced by blackboard state but absent from this worktree checkout.
- Prior plan `specs/plans/readme/20260517-190924-epic-planning-5-architecture-code-planning-0.md`, which owns the `scip-search --version` build identity contract that release installer verification consumes.
- Current worktree file discovery: `go.mod`, `Makefile`, and `cmd/scip-search` exist; no `install.sh` or release installer tests were present at planning time.

## Planning Boundary

This plan covers only the release installer workflow for latest release, explicit `VERSION`, default and custom `INSTALL_DIR`, release-install failure diagnostics, executable availability, and release-install tests that verify success through the installed binary's `--version` output.

Out of scope: `BRANCH`, local clone `make install`, source provenance, source install failures, query command execution, SCIP index generation, release hosting setup, signing, package managers, language indexer installation, README documentation, release validation command documentation, and docs drift checks.

## Architectural Direction

Create one `install.sh` release-mode boundary that selects a released artifact, installs exactly one `scip-search` executable, and reports observable success only after the executable is available in the selected binary directory. Keep the release artifact lookup testable through local or injectable release metadata/artifact sources so installer tests do not depend on hosted release setup.

Release workflow selection should stay separate from source workflow selection. `VERSION` selects released artifacts only. `BRANCH` and local clone behavior are explicitly deferred to sibling source-install planning. The release installer must not require Go, make, a local clone, language indexers, SCIP indexes, or query execution.

Use `scip-search --version` as the verification boundary provided by `epic-planning-5-architecture-code-planning-0`. If a coding checkout lacks an implemented version identity contract that can identify controlled release builds, coders should block rather than inventing a second version contract inside the installer task.

## Planned Coding Tasks

### Task 1 - Implement release artifact selection in `install.sh`

**desc:** `scip-search` release installer users can install latest or explicit released binaries on supported macOS/Linux through `install.sh` using a testable release artifact boundary that never selects branch or source workflows.

**done_when:** Unit tests for `install.sh` release selection pass using controlled release metadata/artifacts, asserting supported macOS/Linux latest installs choose newest release, `VERSION=<release>` chooses exactly the requested release, unsupported OS/architecture and unavailable latest/version exit nonzero with diagnostics naming the platform or `VERSION`, and no Go, make, local clone, language indexer, SCIP index, `BRANCH` workflow, or query command is required.

**scope:** In scope: creating or extending `install.sh` release-mode selection for default latest and `VERSION`, supported macOS/Linux platform artifact mapping, injectable or overridable release metadata/artifact locations for tests, clear release selection diagnostics, and shell-facing tests for release selection and release-selection failure behavior. Out of scope: `BRANCH` or source build paths, local clone `make install`, build metadata implementation, hosted release setup or publication, signing policy, package managers, language indexer installation, SCIP index generation, query command execution, README documentation, release validation command documentation, and docs drift checks.

**spec_ref:** README.md#installation

**task_depends_on:** `epic-planning-5-architecture-code-planning-0`

**Planned files:**
- `install.sh`
- `tests/install/release_selection_test.go` or equivalent installer release-selection test file.

**Implementation notes:**
- Treat unset `VERSION` as latest released artifact selection, not default branch source installation.
- Treat `VERSION` as a released version identifier; do not accept branch names or arbitrary Git refs through this variable.
- Keep release metadata and artifact source injectable for tests so controlled local artifacts can prove behavior without release hosting setup.
- Detect supported macOS/Linux platform and CPU artifact names inside the installer boundary; unsupported platform or missing artifact is a nonzero release-install failure.
- Do not add `BRANCH` behavior in this task; source workflow selection belongs to `epic-planning-5-architecture-code-planning-2`.
- Validation command for the task: `go test ./tests/install` or the equivalent project test target that owns installer release-selection tests.

### Task 2 - Install selected release into default or custom binary directory

**desc:** `scip-search` release installer users can install the selected release artifact into the default or requested binary directory and receive an observable executable path that verifies with `--version`.

**done_when:** Installer tests pass asserting a selected release artifact is installed as an executable `scip-search` in the installer default directory when `INSTALL_DIR` is unset and in `<INSTALL_DIR>/scip-search` when set, success output identifies the installed filesystem path and selected release, `<installed>/scip-search --version` exits `0` and identifies that release, unusable install directories exit nonzero with diagnostics naming `INSTALL_DIR`, and failed installs do not claim success or tell callers to verify a missing binary.

**scope:** In scope: install directory defaulting, `INSTALL_DIR` handling, directory/file write and executable mode behavior, success output that names the installed executable path and release identity, post-install `--version` verification against the installed binary in release-mode tests, and unusable install directory failure handling. Out of scope: release selection internals already owned by Task 1 except consuming its selected artifact boundary, `BRANCH` or source workflows, local clone installs, exact `--version` output formatting beyond release identity observability, query runtime validation, README documentation, broad packaging validation, and host tool provisioning.

**spec_ref:** README.md#installation

**depends_on:** Task 1

**task_depends_on:** `epic-planning-5-architecture-code-planning-0`

**Planned files:**
- `install.sh`
- `tests/install/release_install_test.go` or equivalent installer release-install test file.

**Implementation notes:**
- The installer-defined default directory should be documented in tests as an observable behavior, but this task does not need to update README.
- Custom `INSTALL_DIR` success must be verifiable by direct `<INSTALL_DIR>/scip-search --version` invocation without requiring the directory to be on `PATH`.
- Success output should name the final executable path so automation does not have to infer where the binary landed.
- Failure before install completion must not print a success message or a verification command for a missing binary.
- Do not repair permissions, edit shell profiles, or install extra helper binaries.
- Validation command for the task: `go test ./tests/install` or the equivalent project test target that owns installer release-install tests.

### Task 3 - Add release installer end-to-end coverage with controlled artifacts

**desc:** `scip-search` maintainers can validate the release installer end to end with controlled release artifacts for latest, explicit `VERSION`, custom `INSTALL_DIR`, and release-install failure cases using `--version` as the only executable proof.

**done_when:** End-to-end release installer tests pass from temporary directories with controlled release artifacts, covering latest release install, explicit `VERSION` install, custom `INSTALL_DIR` direct invocation, unsupported platform or unavailable release/version failure, and unusable install directory failure; success cases assert the installed executable path exists, is executable, and `--version` identifies the installed release, while failure cases assert nonzero status, actionable diagnostics, no success claim, and no requirement for Go, make, local clone, SCIP index, language indexer, or query command execution.

**scope:** In scope: black-box or integration tests for `install.sh` release workflows using local/controlled release artifacts or test doubles, temporary install directories, direct `<INSTALL_DIR>/scip-search --version` verification, stderr/stdout/status assertions for release-install failures, and isolation from network-dependent hosted release setup. Out of scope: implementing installer behavior, source/`BRANCH`/local clone workflows, README drift checks or maintainer validation command docs, query fixtures, traversal fixtures, real language indexer execution, package manager validation, signing, notarization, and release approval policy.

**spec_ref:** README.md#installation

**depends_on:** Task 2

**Planned files:**
- `tests/e2e/release_installer_test.go` or equivalent release installer end-to-end test file.

**Implementation notes:**
- Use controlled release artifacts whose binaries expose release identities through the version contract from `epic-planning-5-architecture-code-planning-0`.
- Test latest, explicit `VERSION`, and custom `INSTALL_DIR` as user-observable installer invocations rather than only internal helper functions.
- Failure assertions should check process status and actionable diagnostics without requiring an exact numeric exit taxonomy.
- Do not duplicate broad packaging/docs drift checks; those belong to `epic-planning-5-architecture-code-planning-3`.
- Validation command for the task: `go test ./tests/e2e` or the equivalent project test target that owns release installer e2e tests.

## Dependency Plan

Task 1 has an external `task_depends_on` relationship on `epic-planning-5-architecture-code-planning-0` because release artifacts must be able to identify controlled release builds through `scip-search --version`. If the prior version contract is not implemented in a coding checkout, Task 1 should block rather than defining installer-local version behavior.

Task 2 depends on Task 1 because installation consumes the selected release artifact boundary and both tasks modify `install.sh`. Task 2 also has an external `task_depends_on` relationship on `epic-planning-5-architecture-code-planning-0` because its success criteria verify the installed binary by `--version`.

Task 3 depends on Task 2 because end-to-end release installer coverage needs both release selection and install-directory behavior to exist. Task 3 does not modify `install.sh`; it validates the release installer through the executable/script boundary.

Sibling `epic-planning-5-architecture-code-planning-2` owns `BRANCH` and local clone source install workflows and depends on this release installer plan. Sibling `epic-planning-5-architecture-code-planning-3` owns README documentation, maintainer validation commands, packaging-wide tests, and docs drift checks.

## Shared-File Audit

Task 1 and Task 2 both modify `install.sh`, so Task 2 depends on Task 1.

Task 1 and Task 2 use separate installer test files in the plan. If coders consolidate installer tests into one shared file, Task 2's dependency on Task 1 still serializes the shared test file edits.

Task 3 owns only release installer e2e coverage and depends on Task 2 for behavioral ordering. It should not edit `install.sh`.

No planned task modifies README or source-install files, preserving the scope boundary with `epic-planning-5-architecture-code-planning-2` and `epic-planning-5-architecture-code-planning-3`.

## Test Impact

Task 1 adds installer release-selection tests for latest, explicit `VERSION`, supported platform mapping, unsupported platform, and unavailable release/version failures.

Task 2 adds installer install-directory tests for default directory, custom `INSTALL_DIR`, executable mode, success path observability, installed-binary `--version`, and unusable install directory failures.

Task 3 adds end-to-end release installer tests over controlled artifacts covering the release workflows and release-install failure cases required by the assigned scope.

## Doc Impact

No README or user-facing documentation update is planned in this task because the assigned scope is release installer behavior and release-install tests, while sibling `epic-planning-5-architecture-code-planning-3` owns README installation guidance, release-facing validation command documentation, packaging-wide tests, and docs drift checks.

## Spec Compliance Matrix

| # | Requirement | Source | Task(s) | Status |
|---|-------------|--------|---------|--------|
| 1 | Install the latest released binary by default on supported macOS/Linux when `VERSION` is unset. | `CAP-001-01-release-install-script.md` ST-001 AC-001-1; assigned done_when | Task 1, Task 3 | Covered |
| 2 | Expose the filesystem path of a successful latest-release install. | `CAP-001-01-release-install-script.md` ST-001 AC-001-2; architecture Scope 2 | Task 2, Task 3 | Covered |
| 3 | Produce an installed file that the operating system treats as executable. | `CAP-001-01-release-install-script.md` ST-001 AC-001-3 | Task 2, Task 3 | Covered |
| 4 | Release install success must not require a local clone, Go, make, language indexers, or a SCIP index file. | `CAP-001-01-release-install-script.md` ST-001 AC-001-4; NFR-000-2 | Task 1, Task 2, Task 3 | Covered |
| 5 | `VERSION=<release>` installs exactly the requested available release rather than silently installing latest. | `CAP-001-01-release-install-script.md` ST-002 AC-002-1, AC-002-2; assigned done_when | Task 1, Task 3 | Covered |
| 6 | Successful explicit-version installs identify both the requested version and installed executable path. | `CAP-001-01-release-install-script.md` ST-002 AC-002-3 | Task 2, Task 3 | Covered |
| 7 | `VERSION` release installs must not require a local clone, Go, make, language indexers, or a SCIP index file. | `CAP-001-01-release-install-script.md` ST-002 AC-002-4 | Task 1, Task 2, Task 3 | Covered |
| 8 | `INSTALL_DIR=<directory>` installs `scip-search` into the requested usable directory. | `CAP-001-01-release-install-script.md` ST-003 AC-003-1; assigned done_when | Task 2, Task 3 | Covered |
| 9 | `<INSTALL_DIR>/scip-search` is directly executable after a custom-directory install. | `CAP-001-01-release-install-script.md` ST-003 AC-003-2 | Task 2, Task 3 | Covered |
| 10 | `INSTALL_DIR` can be combined with `VERSION` to install the requested release in the requested directory. | `CAP-001-01-release-install-script.md` ST-003 AC-003-3 | Task 1, Task 2, Task 3 | Covered |
| 11 | Custom install directory success is observable without relying on `PATH`. | `CAP-001-01-release-install-script.md` ST-003 AC-003-4 | Task 2, Task 3 | Covered |
| 12 | Unsupported platform or unavailable latest release failures return nonzero and identify the release-install condition. | `CAP-001-02-release-install-failure-and-verification.md` ST-001 AC-001-1; assigned done_when | Task 1, Task 3 | Covered |
| 13 | Unavailable `VERSION` failures return nonzero and identify the requested version as the failing input. | `CAP-001-02-release-install-failure-and-verification.md` ST-001 AC-001-2; assigned done_when | Task 1, Task 3 | Covered |
| 14 | Unusable `INSTALL_DIR` failures return nonzero and identify the requested install directory. | `CAP-001-02-release-install-failure-and-verification.md` ST-001 AC-001-3; assigned done_when | Task 2, Task 3 | Covered |
| 15 | Failed release installs do not claim success or instruct verification of a binary that was not installed. | `CAP-001-02-release-install-failure-and-verification.md` ST-001 AC-001-4 | Task 2, Task 3 | Covered |
| 16 | Failure diagnostics stay distribution-scoped and do not require generating a SCIP index, installing a language indexer, building from a branch, or cloning the repository to understand remediation. | `CAP-001-02-release-install-failure-and-verification.md` ST-001 AC-001-5 | Task 1, Task 2, Task 3 | Covered |
| 17 | Successful latest-release installs verify through `scip-search --version` as a released `scip-search` version. | `CAP-001-02-release-install-failure-and-verification.md` ST-002 AC-002-1 | Task 2, Task 3 | Covered |
| 18 | Successful explicit `VERSION` installs verify through `--version` as the same requested release. | `CAP-001-02-release-install-failure-and-verification.md` ST-002 AC-002-2 | Task 2, Task 3 | Covered |
| 19 | Successful `INSTALL_DIR` installs verify through direct `<INSTALL_DIR>/scip-search --version`. | `CAP-001-02-release-install-failure-and-verification.md` ST-002 AC-002-3 | Task 2, Task 3 | Covered |
| 20 | Release verification does not require a local clone, Go, make, language indexers, SCIP index generation, or query command execution. | `CAP-001-02-release-install-failure-and-verification.md` ST-002 AC-002-4; assigned scope | Task 2, Task 3 | Covered |
| 21 | Installed executable availability is observable independently of any SCIP index file. | `CAP-001-02-release-install-failure-and-verification.md` ST-002 AC-002-5 | Task 2, Task 3 | Covered |
| 22 | Release workflows remain separate from `BRANCH`, source install, and local clone `make install` behavior. | assigned scope exclusions; architecture Scope 2 boundary | Task 1, Task 2, Task 3 | Covered |
| 23 | Release installer tests may use controlled artifacts without owning hosted release setup. | architecture ASM-003; `CAP-004-02-packaging-and-release-validation.md` ASM-000-1 | Task 1, Task 3 | Covered |
| E2E | e2e test coverage for new behavior | Cross-cutting | Task 3 | Covered |
| DOC | Documentation updates for changed behavior | Cross-cutting | N/A: README and release-facing documentation are explicitly assigned to sibling `epic-planning-5-architecture-code-planning-3` | N/A |

## Pre-Submit Validation Checklist

- Re-read this plan and verify the output JSON fields are character-identical to each task's `desc`, `done_when`, `scope`, and `spec_ref`.
- Run `jq . specs/plans/readme/20260517-193112-epic-planning-5-architecture-code-planning-1-output.json`.
- Search this plan for `Task 1`, `Task 2`, `Task 3`, `epic-planning-5-architecture-code-planning-0`, `epic-planning-5-architecture-code-planning-2`, `epic-planning-5-architecture-code-planning-3`, and `install.sh` references and confirm each responsibility, dependency, and exclusion is stated by the referenced task.
- Run pre-commit on the plan and output JSON files.
- Commit only the plan and output JSON artifacts.
