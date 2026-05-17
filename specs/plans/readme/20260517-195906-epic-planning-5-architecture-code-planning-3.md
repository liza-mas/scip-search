# Code Plan: Distribution Documentation And Packaging Validation

Status: draft

## Source Context

Based on:
- `README.md#installation`
- `README.md#existing-indexers`
- `specs/epics/readme/20260517-151311-epic-planning-5.md#capability-cap-004---document-and-validate-distribution-workflows`
- `specs/stories/readme/20260517-151311-epic-planning-5/CAP-004-01-readme-installation-documentation.md`
- `specs/stories/readme/20260517-151311-epic-planning-5/CAP-004-02-packaging-and-release-validation.md`
- Architecture reference `specs/arch-plan/readme/20260517-170615-epic-planning-5-architecture.md#scope-4-distribution-documentation-and-packaging-validation` read from merge commit `25dc06d175819fe6c79d9faa5ff8f6b0ce43c811` because the file is referenced by blackboard state but absent from this worktree checkout.
- Prior plan `specs/plans/readme/20260517-190924-epic-planning-5-architecture-code-planning-0.md`, which owns the offline `scip-search --version` build identity contract consumed by documentation and validation.
- Prior plan `specs/plans/readme/20260517-193112-epic-planning-5-architecture-code-planning-1.md`, read from merge commit `6d96139642be9bcc380e2df787abaff8ed532be0` because the plan file is referenced by blackboard state but absent from this worktree checkout; it owns release installer behavior and release installer e2e coverage.
- Prior plan `specs/plans/readme/20260517-194106-epic-planning-5-architecture-code-planning-2.md`, which owns source install behavior and source install e2e coverage.
- Current worktree file discovery: `README.md`, `go.mod`, `Makefile`, `install.sh`, `cmd/scip-search/main.go`, `cmd/scip-search/main_test.go`, and `tests/install/release_selection_test.go` exist; no `docs/` directory exists at planning time.
- Targeted blackboard reads for `epic-planning-5-architecture-code-planning-2-coding-0`, `epic-planning-5-architecture-code-planning-2-coding-1`, and `epic-planning-5-architecture-code-planning-2-coding-2` returned `not_found`, so this plan references the merged source-install planning task where concrete source-install child task IDs are not yet available.

## Planning Boundary

This plan covers only README installation documentation, maintainer release validation command documentation, documentation drift checks, and a packaging validation command surface that composes distribution-scoped tests for release install, source install, custom install directories, actionable failure signals, docs drift, and `--version` verification.

Out of scope: changing `install.sh`, `make install`, `scip-search --version`, release artifacts, source build behavior, query command behavior, query-specific fixtures, traversal fixtures, real language indexer execution, unsupported packaging channels, release approval policy, changelog policy, hosted release setup, signing, and notarization.

## Architectural Direction

Keep documentation and validation as consumers of the distribution contracts from `epic-planning-5-architecture-code-planning-0`, `epic-planning-5-architecture-code-planning-1`, and `epic-planning-5-architecture-code-planning-2`. The README and maintainer validation docs should describe supported workflow commands and verification surfaces without re-owning installer internals, source build internals, or exact `--version` formatting.

Separate the public README from maintainer validation procedure details. The README should remain concise and user-facing, while a maintainer-facing validation document can enumerate release validation command sequences, expected success proof, and distribution-scoped failure interpretation.

Implement drift checks that compare documented command surfaces with the supported distribution workflows and fail as documentation mismatches. Drift checks should not execute real public release downloads or broad query validation; behavior-smoke coverage belongs to packaging validation and existing installer/source e2e tests.

Implement one distribution validation command or target that composes the release installer, source installer, `--version`, and documentation drift checks. The target should make distribution failures identifiable without adding a new machine-readable installer schema and without invoking `symbols`, `packages`, `references`, `implementations`, generated SCIP indexes, traversal fixtures, language indexers, or query golden JSON.

## Planned Coding Tasks

### Task 1 - Update installation and release validation documentation

**desc:** `scip-search` maintainers have README installation guidance and maintainer release validation commands that enumerate every supported distribution workflow and verify installs only through `--version`.

**done_when:** README and maintainer validation documentation are updated so the README installation section documents latest-release install, explicit `VERSION=<release>`, `BRANCH=<branch>`, custom `INSTALL_DIR=<directory>`, local clone `make install`, Go and make as caller-provided source prerequisites, separate language-indexer prerequisites for SCIP index generation, `scip-search --version` verification for all workflows, direct `<INSTALL_DIR>/scip-search --version` custom-directory verification, and no unsupported package manager, signing, hosted release setup, query command, SCIP index, language indexer execution, traversal fixture, or query golden JSON validation path.

**scope:** In scope: `README.md#installation`, wording that separates installing `scip-search` from installing language indexers, source prerequisite notes for `BRANCH` and local clone workflows, custom install directory verification guidance, and a new maintainer-facing distribution validation document such as `docs/release-validation.md` that lists latest release, explicit `VERSION`, custom `INSTALL_DIR`, `BRANCH`, local clone `make install`, and `--version` validation command sequences. Out of scope: changing installer behavior, source build behavior, version output formatting, release artifact hosting, release approval or changelog policy, signing or notarization, package manager docs, query command documentation, query validation, SCIP traversal validation, real language indexer execution, and packaging test implementation.

**spec_ref:** README.md#installation

**task_depends_on:** `epic-planning-5-architecture-code-planning-0`, `epic-planning-5-architecture-code-planning-1`, `epic-planning-5-architecture-code-planning-2`

**Planned files:**
- `README.md`
- `docs/release-validation.md`

**Implementation notes:**
- The README should keep language indexers as separate query-preparation prerequisites, not as prerequisites for installing or verifying the `scip-search` binary.
- Use placeholder values such as `<release>`, `<branch>`, and `<directory>` where that keeps examples reusable; keep commands copyable after substitution.
- Custom directory guidance must show direct invocation from the requested directory because `INSTALL_DIR` may not be on `PATH`.
- Maintainer validation docs should state that success proof is installed executable plus `--version`, not a query command or SCIP index.
- If the repository establishes a different maintainer-doc location before coding begins, use that existing location instead of creating a competing docs structure.
- Validation command for this task: `rg` checks over `README.md` and maintainer validation docs, plus the repository markdown or pre-commit checks for touched docs.

### Task 2 - Add documentation drift checks for install guidance

**desc:** `scip-search` maintainers can fail fast when README installation examples or maintainer validation commands drift from the supported distribution command surface.

**done_when:** Documentation drift tests pass by inspecting README installation guidance and maintainer validation documentation, asserting the documented command surface includes latest release install, explicit `VERSION=<release>`, `BRANCH=<branch>`, custom `INSTALL_DIR=<directory>`, local clone `make install`, `scip-search --version`, direct custom-directory version verification, Go and make source prerequisite notes, separate language-indexer prerequisite wording, and maintainer validation commands; the tests fail with distribution documentation mismatch diagnostics when required examples are missing, stale, or replaced by query-specific, traversal-specific, real-indexer, package-manager, signing, notarization, hosted-release-setup, or query-golden validation paths.

**scope:** In scope: documentation drift test code, expected supported distribution command-surface fixtures or constants, assertions over `README.md` and maintainer validation docs, and distribution-scoped failure messages for documentation mismatch. Out of scope: implementing or changing installer behavior, source install behavior, `--version` behavior, release/source e2e execution, public network release downloads, query command tests, SCIP index fixtures, traversal fixtures, real language indexer execution, release approval policy, changelog policy, signing, and notarization.

**spec_ref:** README.md#installation

**depends_on:** Task 1

**Planned files:**
- `tests/docs/install_docs_drift_test.go` or equivalent documentation drift test file.
- Optional small test fixture or helper under `tests/docs/` if the project test pattern favors fixtures.

**Implementation notes:**
- Treat docs drift as a documentation contract failure, not as an installer runtime failure.
- Prefer explicit expected command patterns over brittle full-section snapshot tests.
- Check for prohibited query-validation wording by command surface, not by banning all query examples elsewhere in the README.
- Keep drift tests independent from public network availability and hosted release setup.
- Validation command for this task: `go test ./tests/docs` or the equivalent project test target that owns docs drift tests.

### Task 3 - Add distribution packaging validation target

**desc:** `scip-search` maintainers can run one distribution validation command set that covers release install, source install, custom install directory, actionable failure signals, docs drift, and `--version` proof without invoking query or traversal validation.

**done_when:** A maintainer-facing distribution validation command, such as `make validate-distribution`, passes by running distribution-scoped packaging checks for release install, explicit `VERSION`, custom `INSTALL_DIR`, branch source install, local clone `make install`, installed-executable `--version` verification, release/source/custom-install failure diagnostics, and documentation drift; when a covered check fails, its output identifies a distribution validation category rather than a query runtime category, and the command set does not require generated SCIP indexes, real language indexer execution, query-specific fixtures, traversal fixtures, query golden JSON, unsupported package managers, signing, notarization, hosted release setup, or release approval policy.

**scope:** In scope: a maintainer-facing validation target or script, wiring to the existing or planned release installer e2e tests from `epic-planning-5-architecture-code-planning-1`, source install e2e tests from `epic-planning-5-architecture-code-planning-2`, `--version` executable smoke coverage from `epic-planning-5-architecture-code-planning-0`, documentation drift tests from Task 2, and packaging-level failure labels or diagnostics. Out of scope: changing release installer internals, source installer internals, `--version` formatting, README copy except consuming Task 1 documentation, adding query/traversal validation, executing real language indexers, public hosted release setup, release approval gates, changelog policy, signing, notarization, and package manager validation.

**spec_ref:** README.md#installation

**depends_on:** Task 2

**task_depends_on:** `epic-planning-5-architecture-code-planning-0-coding-1-after-version-implementation`, `epic-planning-5-architecture-code-planning-1-coding-2`, `epic-planning-5-architecture-code-planning-2`

**Planned files:**
- `Makefile`
- `scripts/validate-distribution.sh` or equivalent project-local validation script if the Makefile delegates validation orchestration.
- `tests/e2e/distribution_validation_test.go` or equivalent packaging validation target test, if the project uses tests to assert validation-target wiring.

**Implementation notes:**
- Compose existing release/source/version tests rather than duplicating query-independent installer behavior already owned by prior plans.
- If concrete source-install coding task IDs exist by coding time, depend on those tasks or consume their merged test targets instead of broadening this task to implement source install behavior.
- The validation command should be runnable by automation and non-interactive.
- Failure labels should help maintainers distinguish docs drift, release install packaging, source install packaging, custom directory, failure-signal, and version verification failures.
- Do not run `scip-search symbols`, `packages`, `references`, or `implementations` as part of distribution readiness.
- Validation command for this task: the new distribution validation command itself, plus the focused test that proves the command wiring.

## Dependency Plan

Task 1 has external `task_depends_on` relationships on the three prior distribution planning tasks because it documents the supported version, release installer, and source installer command surfaces without redefining them.

Task 2 depends on Task 1 because drift checks inspect the README and maintainer validation documentation introduced or updated there.

Task 3 depends on Task 2 because the distribution validation target must include docs drift checks. Task 3 also depends externally on the version executable smoke task and release installer e2e task from prior plans. The merged source-install code-planning task is listed as an external dependency because concrete child coding task IDs were not present in targeted blackboard reads at planning time; coders should consume the merged source-install tests once available rather than implementing source install behavior here.

No task in this plan changes installer, source build, or version behavior. If a coding checkout lacks the behavior or test targets from prior distribution plans, the affected coding task should block rather than broaden this plan into installer/source/version implementation.

## Shared-File Audit

Task 1 modifies `README.md` and creates or updates maintainer validation documentation.

Task 2 owns documentation drift tests and depends on Task 1 because it consumes Task 1's docs. It should not edit `README.md` or maintainer docs except to resolve test-discovered documentation omissions within the same approved documentation scope; if that happens, keep the Task 2 dependency on Task 1 and update only the missing documentation required for drift-test truth.

Task 3 modifies validation orchestration such as `Makefile` and optionally a validation script or target test. It depends on Task 2 for behavioral ordering and should not edit `README.md`, maintainer docs, `install.sh`, or source install internals.

`Makefile` is also owned by prior source-install planning. Task 3's dependency on `epic-planning-5-architecture-code-planning-2` serializes validation-target edits behind source-install planning; if the source-install coding tasks have concrete IDs by implementation time, the coder should use the merged source-install Makefile state rather than replacing it.

## Test Impact

Task 1 is documentation-only. It should run markdown/pre-commit checks and any existing docs checks, but it does not introduce behavior tests.

Task 2 adds documentation drift tests that inspect README and maintainer validation docs for required and prohibited distribution command surfaces.

Task 3 adds or wires a distribution packaging validation command that runs the existing version, release installer, source installer, and docs drift checks as one maintainer-facing validation surface.

## Doc Impact

Task 1 updates `README.md#installation` and adds or updates maintainer release validation documentation.

Task 2 and Task 3 may reference those docs in tests or validation commands but should not broaden user-facing documentation beyond the Task 1 scope.

## Spec Compliance Matrix

| # | Requirement | Source | Task(s) | Status |
|---|-------------|--------|---------|--------|
| 1 | README documents the latest-release macOS/Linux install command. | `CAP-004-01-readme-installation-documentation.md` ST-001 AC-001-1; assigned done_when | Task 1, Task 2 | Covered |
| 2 | README documents an explicit released-version installer example using `VERSION=<release>`. | `CAP-004-01-readme-installation-documentation.md` ST-001 AC-001-2; assigned done_when | Task 1, Task 2 | Covered |
| 3 | README documents a branch source-build installer example using `BRANCH=<branch>` and states Go and make are caller-provided prerequisites. | `CAP-004-01-readme-installation-documentation.md` ST-001 AC-001-3; assigned done_when | Task 1, Task 2 | Covered |
| 4 | README documents a custom-directory installer example using `INSTALL_DIR=<directory>`. | `CAP-004-01-readme-installation-documentation.md` ST-001 AC-001-4; assigned done_when | Task 1, Task 2 | Covered |
| 5 | README documents the local clone workflow ending in `make install`. | `CAP-004-01-readme-installation-documentation.md` ST-001 AC-001-5; assigned done_when | Task 1, Task 2 | Covered |
| 6 | README installation guidance does not introduce unsupported package manager, container, signing, notarization, hosted release setup, or query-validation installation paths. | `CAP-004-01-readme-installation-documentation.md` ST-001 AC-001-6 | Task 1, Task 2 | Covered |
| 7 | README states language indexers are separate tools needed for generating SCIP indexes and are not installed by the `scip-search` install workflow. | `CAP-004-01-readme-installation-documentation.md` ST-002 AC-002-1; assigned done_when | Task 1, Task 2 | Covered |
| 8 | README identifies Go and make as caller-provided prerequisites for `BRANCH` and local clone source builds. | `CAP-004-01-readme-installation-documentation.md` ST-002 AC-002-2; assigned done_when | Task 1, Task 2 | Covered |
| 9 | README release binary guidance does not require Go, make, a local clone, language indexers, or SCIP indexes for release install verification. | `CAP-004-01-readme-installation-documentation.md` ST-002 AC-002-3 | Task 1, Task 2 | Covered |
| 10 | README shows `scip-search --version` as the installation verification command for supported workflows. | `CAP-004-01-readme-installation-documentation.md` ST-002 AC-002-4; assigned done_when | Task 1, Task 2 | Covered |
| 11 | README custom `INSTALL_DIR` verification makes clear callers can invoke `<directory>/scip-search --version` without relying on `PATH`. | `CAP-004-01-readme-installation-documentation.md` ST-002 AC-002-4b; assigned done_when | Task 1, Task 2 | Covered |
| 12 | README verification guidance does not require `--index`, query subcommands, generated SCIP data, or network release lookup. | `CAP-004-01-readme-installation-documentation.md` ST-002 AC-002-5 | Task 1, Task 2 | Covered |
| 13 | Maintainer release validation commands include latest-release install followed by `scip-search --version`. | `CAP-004-02-packaging-and-release-validation.md` ST-001 AC-001-1; assigned done_when | Task 1, Task 2, Task 3 | Covered |
| 14 | Maintainer release validation commands include explicit `VERSION=<release>` install followed by version verification. | `CAP-004-02-packaging-and-release-validation.md` ST-001 AC-001-2; assigned done_when | Task 1, Task 2, Task 3 | Covered |
| 15 | Maintainer release validation commands include custom `INSTALL_DIR=<directory>` install followed by `<directory>/scip-search --version`. | `CAP-004-02-packaging-and-release-validation.md` ST-001 AC-001-3; assigned done_when | Task 1, Task 2, Task 3 | Covered |
| 16 | Maintainer release validation commands include branch source install followed by version verification. | `CAP-004-02-packaging-and-release-validation.md` ST-001 AC-001-4; assigned done_when | Task 1, Task 2, Task 3 | Covered |
| 17 | Maintainer release validation commands include local clone validation with `make install` and `scip-search --version`. | `CAP-004-02-packaging-and-release-validation.md` ST-001 AC-001-5; assigned done_when | Task 1, Task 2, Task 3 | Covered |
| 18 | Distribution validation success is based on installed executable and `--version`, not query commands, SCIP indexes, language indexers, traversal fixtures, or query golden JSON. | `CAP-004-02-packaging-and-release-validation.md` ST-001 AC-001-6; NFR-000-2 | Task 1, Task 2, Task 3 | Covered |
| 19 | Packaging tests cover latest-release install, explicit `VERSION`, and custom `INSTALL_DIR` release outcomes. | `CAP-004-02-packaging-and-release-validation.md` ST-002 AC-002-1; assigned done_when | Task 3 | Covered |
| 20 | Packaging tests cover branch source install and local clone `make install` using caller-provided Go and make. | `CAP-004-02-packaging-and-release-validation.md` ST-002 AC-002-2; assigned done_when | Task 3 | Covered |
| 21 | Packaging tests verify installed binaries through `--version`. | `CAP-004-02-packaging-and-release-validation.md` ST-002 AC-002-3; assigned done_when | Task 3 | Covered |
| 22 | Packaging tests fail with distribution documentation mismatch diagnostics when README examples drift from installer, source build, or version verification contracts. | `CAP-004-02-packaging-and-release-validation.md` ST-002 AC-002-4; assigned done_when | Task 2, Task 3 | Covered |
| 23 | Packaging coverage includes release install, source install, custom install directory, and `--version` verification. | `CAP-004-02-packaging-and-release-validation.md` ST-002 AC-002-5; assigned done_when | Task 3 | Covered |
| 24 | Packaging tests do not require generated SCIP indexes, real language indexer execution, query-specific fixtures, traversal fixtures, or query golden JSON. | `CAP-004-02-packaging-and-release-validation.md` ST-002 AC-002-6; assigned scope | Task 2, Task 3 | Covered |
| 25 | Packaging failure-signal validation observes installer or source-build failures without changing expectations to make failed installs appear successful. | `CAP-004-02-packaging-and-release-validation.md` ST-002 AC-002-6b; assigned done_when | Task 3 | Covered |
| 26 | Validation output lets maintainers distinguish install packaging failures from query runtime failures. | `CAP-004-02-packaging-and-release-validation.md` NFR-000-3 | Task 3 | Covered |
| 27 | Distribution validation is runnable by automation without an interactive UI. | `CAP-004-02-packaging-and-release-validation.md` NFR-000-1; epic NFR-000-2 | Task 3 | Covered |
| 28 | Documentation and validation do not change installer, source install, version, query, traversal, release approval, changelog, hosted release, signing, or notarization behavior. | assigned scope exclusions; architecture Scope 4 boundary | Task 1, Task 2, Task 3 | Covered |
| E2E | e2e test coverage for new behavior | Cross-cutting | Task 3 | Covered |
| DOC | Documentation updates for changed behavior | Cross-cutting | Task 1 | Covered |

## Pre-Submit Validation Checklist

- Re-read this plan and verify the output JSON fields are character-identical to each task's `desc`, `done_when`, `scope`, and `spec_ref`.
- Run `jq . specs/plans/readme/20260517-195906-epic-planning-5-architecture-code-planning-3-output.json`.
- Search this plan for `Task 1`, `Task 2`, `Task 3`, `epic-planning-5-architecture-code-planning-0`, `epic-planning-5-architecture-code-planning-1`, `epic-planning-5-architecture-code-planning-2`, `README.md`, `docs/release-validation.md`, `Makefile`, and `validate-distribution` references and confirm each responsibility, dependency, and exclusion is stated by the referenced task.
- Verify every output JSON `depends_on` index matches this plan's dependency order.
- Run pre-commit on the plan and output JSON files.
- Commit only the plan and output JSON artifacts.
