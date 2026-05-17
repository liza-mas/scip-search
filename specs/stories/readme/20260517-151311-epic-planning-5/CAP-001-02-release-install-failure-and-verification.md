# User Stories: Release Install Failures and Verification Outcome

Status: review

## Goal
`scip-search` release installer users and automation can distinguish actionable install failures from successful installs and verify the installed executable by version without a local clone or SCIP index.

## Parent Epic
specs/epics/readme/20260517-151311-epic-planning-5.md - Capability CAP-001

## Context
This document covers the observable failure and verification behavior around release installs. It uses `scip-search --version` only as the user-visible verification command and leaves the detailed version output contract to the sibling version-information capability.

## Personas
- **Automation Agent**: An AI or script-driven caller running terminal commands in constrained macOS or Linux environments, needing deterministic failure signals and a quick installed-binary verification command.
- **CLI Maintainer**: A Go developer preparing releases for macOS and Linux users, needing failure and verification cases that can be validated without exercising query execution.

## General information

Applies to: release installer failures and post-install executable verification for CAP-001.

### References
- goal spec: README.md#installation, lines 131-148 - Defines latest release, `VERSION`, and `INSTALL_DIR` install invocations.
- goal spec: README.md#installation, lines 157-160 - Defines `scip-search --version` as the verification command.
- parent epic: specs/epics/readme/20260517-151311-epic-planning-5.md#capability-cap-001---install-released-scip-search-binaries - Requires actionable install failures and installed executable verification.
- task: epic-planning-5-us-writing-0 - Requires install failure reporting and version-based verification without requiring a local clone or SCIP index generation.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/cli-command-routing-and-usage.md - Read for story format and scope-boundary style.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/stderr-exit-error-contract.md - Read for stream and status boundary style without importing query-runtime failure semantics.

### Non-Functional Requirements
- NFR-000-1: Release install failures must be usable by automation through a nonzero process result and human-readable diagnostic text.
- NFR-000-2: Verification must not require SCIP language indexers, SCIP index files, query execution, Go, make, or a local repository clone.
- NFR-000-3: Verification behavior must remain compatible with the sibling version-information contract rather than redefining exact version output in this document.

### Related External Components
- Component C-001 - Release installer: The user-facing `install.sh` workflow invoked from the README for latest release, specific version, and custom directory installs.
- Component C-002 - Release artifacts: The published binaries or archives consumed by release install workflows.
- Component C-003 - Calling process environment: The macOS/Linux shell environment invoking the installer and observing process result plus output streams.
- Component C-004 - README installation documentation: The user-facing installation and verification guidance maintained with distribution behavior.

### Interfaces
- I-001-001 - Installer invocation contract (Interface 001 of Component C-001): The shell invocation and environment options users rely on for release and custom-directory installs.
- I-004-001 - Version verification contract (Interface 001 of Component C-004): The command-line version output users and release automation use to confirm the installed executable.

### Out of Scope
- Branch builds, local clone builds, Go or make source-build behavior, and `BRANCH` handling.
- Installing SCIP language indexers, generating SCIP indexes, or validating query execution.
- Package manager formulas, container images, signing, notarization, and release hosting setup.
- Exact `--version` output formatting, beyond requiring that the installed release can be identified.
- Exact numeric installer exit status taxonomy.

### Assumptions
- **ASM-000-1**: The installer communicates normal failures through a nonzero process result and diagnostic output, without requiring a machine-readable error schema. - *Why*: The epic requires actionable install failures for automation but does not define an installer JSON contract. - Confidence: MEDIUM
- **ASM-000-2**: CAP-003 owns the detailed `scip-search --version` output contract; CAP-001 only requires that verification can identify the installed release well enough to confirm the install. - *Why*: The parent epic has a separate capability for reporting CLI version information. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Report Actionable Release Install Failures

### References
- goal spec: README.md#installation, lines 131-148 - Defines supported release install inputs whose failures must be actionable.
- parent epic: specs/epics/readme/20260517-151311-epic-planning-5.md#capability-cap-001---install-released-scip-search-binaries - Requires release install failure reporting.

### User Story
**As an** Automation Agent installing `scip-search` from a script, **I want to** receive actionable release installer failures, **so that** my workflow can stop on unsupported or unusable install conditions instead of assuming the CLI is available.

### Acceptance Criteria
- AC-001-1: Given the installer cannot install a latest release for the current macOS or Linux environment, when the install fails, then the process reports a nonzero result and a diagnostic that identifies the unsupported or unavailable release-install condition.
- AC-001-2: Given `VERSION` names a release that the installer cannot fetch or match to a release artifact, when the install fails, then the process reports a nonzero result and a diagnostic that identifies the requested version as the failing input.
- AC-001-3: Given `INSTALL_DIR` is unusable by the installer for the current caller, when the install fails, then the process reports a nonzero result and a diagnostic that identifies the requested install directory as the failing input.
- AC-001-4: Given a release install fails before an executable is installed, when the caller checks for success, then the installer does not report the install as successful and does not instruct the caller to verify a binary that was not installed.
- AC-001-5: Given a release install fails, when the caller reviews required remediation, then the diagnostic does not require generating a SCIP index, installing a language indexer, building from a branch, or cloning the repository to understand the install failure.

### Depends on:
Run time coupling:
- Interface I-001-001 - Installer invocation contract

### Out of Scope
- Exact diagnostic wording or exact numeric exit status.
- Retrying downloads, selecting mirrors, or defining release hosting behavior.
- Query runtime stderr/stdout semantics.

### Assumptions
- **ASM-001-1**: Unsupported platform, unavailable requested version, and unusable install directory are the release-install failure classes that must be explicitly user-actionable for CAP-001. - *Why*: They correspond to macOS/Linux scope, `VERSION`, and `INSTALL_DIR` behavior required by the task. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-002 - Verify the Installed Executable by Version

### References
- goal spec: README.md#installation, lines 157-160 - Defines `scip-search --version` as the verification command.
- parent epic: specs/epics/readme/20260517-151311-epic-planning-5.md#capability-cap-001---install-released-scip-search-binaries - Requires version-based installed executable verification.

### User Story
**As an** Automation Agent verifying a release install, **I want to** run the installed `scip-search --version`, **so that** I can confirm the installed executable is present and corresponds to the expected release without needing a SCIP index.

### Acceptance Criteria
- AC-002-1: Given a latest-release install succeeds, when the caller runs the installed `scip-search --version`, then the command succeeds and identifies the installed build as a released `scip-search` version.
- AC-002-2: Given an explicit `VERSION` install succeeds, when the caller runs the installed `scip-search --version`, then the command succeeds and identifies the same release version requested by `VERSION`.
- AC-002-3: Given an `INSTALL_DIR` install succeeds, when the caller runs `<INSTALL_DIR>/scip-search --version`, then the command succeeds without requiring the directory to be on `PATH`.
- AC-002-4: Given any successful release install, when the caller runs version verification, then verification does not require a local `scip-search` clone, Go, make, language indexers, SCIP index generation, or query command execution.
- AC-002-5: Given version verification succeeds, when the caller evaluates install success, then the installed executable is observable as available independently of any SCIP index file.

### Depends on:
Implementation ordering:
- Story document CAP-001-01-release-install-script.md - Successful release install behavior must exist before installed executable verification can be exercised.

Run time coupling:
- Interface I-004-001 - Version verification contract

### Out of Scope
- Exact `--version` output fields, ordering, or formatting beyond identifying the installed release.
- Verifying query results after installation.
- Verifying source-built or branch-built binaries.

### Assumptions
- **ASM-002-1**: A successful `--version` invocation is sufficient to prove the installed executable can start for verification purposes, even though query commands require separate SCIP index inputs. - *Why*: The README verification step is `scip-search --version`, and query behavior is owned by other epics. - Confidence: HIGH

### Open Questions
- None.
