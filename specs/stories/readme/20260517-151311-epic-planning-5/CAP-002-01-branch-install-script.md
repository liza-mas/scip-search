# User Stories: Branch Source Install Through Installer

Status: review

## Goal
Source builders can request a branch install through the README installer command and receive either an installed `scip-search` executable built from source or an actionable source-build failure.

## Parent Epic
specs/epics/readme/20260517-151311-epic-planning-5.md - Capability CAP-002

## Context
This document covers the `BRANCH=<branch> bash` installer path from the README. It keeps branch source builds distinct from release binary downloads and local clone workflows while preserving the same user-facing installer entry point.

## Personas
- **Source Builder**: A developer or early adopter with Go and make available, installing from a branch before a release artifact exists.
- **Automation Agent**: A script-driven caller running terminal commands in constrained environments, needing deterministic install success and actionable non-interactive failures.

## General information

Applies to: branch-based source installation through the README installer invocation.

### References
- goal spec: README.md#installation - Defines `BRANCH=main` installer behavior and states that branch builds require Go and make.
- goal spec: README.md#what-is-scip-search - Defines the installed artifact as the one-shot `scip-search` Go binary.
- parent epic: specs/epics/readme/20260517-151311-epic-planning-5.md#personas - Defines the Source Builder and Automation Agent personas.
- parent epic: specs/epics/readme/20260517-151311-epic-planning-5.md#general-information - Defines distribution-wide NFRs, external components, interfaces, and out-of-scope boundaries.
- parent epic: specs/epics/readme/20260517-151311-epic-planning-5.md#capability-cap-002---build-and-install-from-source-workflows - Defines source-based branch and local clone workflows.
- task: epic-planning-5-us-writing-1 - Limits this story work to BRANCH source builds, local clone make install behavior, Go/make prerequisites, installed CLI outcome, and source-build failures.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/cli-command-routing-and-usage.md - Read for local story format and scope-boundary style; no existing CAP-002 story documents were present before this task.

### Non-Functional Requirements
- NFR-000-1: Branch source installs must not provision Go, make, or language indexers on the caller's host.
- NFR-000-2: Branch source installs must be automation-friendly: success and failure outcomes are non-interactive and observable from the invoking shell.
- NFR-000-3: Branch source install behavior must stay separate from release artifact selection and download behavior.

### Related External Components
- Component C-001 - Release installer: The user-facing `install.sh` workflow invoked from the README with `BRANCH`.
- Component C-003 - Local Go and make toolchain: The caller-provided tools required for source builds.
- Component C-004 - README installation documentation: The user-facing installation and verification guidance maintained with the distribution behavior.

### Interfaces
- I-001-001 - Installer invocation contract (Interface 001 of Component C-001): The shell invocation and environment options users rely on for branch installs.
- I-003-001 - Source build contract (Interface 001 of Component C-003): The caller-provided Go and make command surface used by source builders.

### Out of Scope
- Latest release or `VERSION` release artifact download behavior.
- Cross-compilation, CI release publication, hosted release setup, package manager formulas, signing, notarization, and operating systems beyond the README install scope.
- Local clone `make install` behavior, except where referenced as a separate CAP-002 story.
- Query command validation, SCIP index loading, SCIP traversal, and query result schemas.
- Installing Go, make, or language indexers.
- Defining exact `scip-search --version` output content.

### Assumptions
- **ASM-000-1**: A non-empty `BRANCH` option selects the source-build installer path instead of the release artifact path. - *Why*: The README labels `BRANCH=main` as "Build from a branch (requires Go and make)", and the capability separates source workflows from release artifact behavior. - Confidence: HIGH
- **ASM-000-2**: The branch install path may validate the installed executable by confirming that a runnable `scip-search` binary exists, without defining the exact version string. - *Why*: CAP-003 owns version output behavior, while this capability owns the installed CLI outcome. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Build and Install a Requested Branch Through the Installer

### References
- goal spec: README.md#installation - Defines the `BRANCH=main` installer example and Go/make prerequisite.
- parent epic: specs/epics/readme/20260517-151311-epic-planning-5.md#capability-cap-002---build-and-install-from-source-workflows - Defines branch source build behavior and failure visibility.

### User Story
**As a** Source Builder installing a pre-release branch from the terminal, **I want to** pass a requested branch to the README installer command, **so that** I can install a runnable `scip-search` CLI before a release artifact is available.

### Acceptance Criteria
- AC-001-1: Given Go and make are available on the caller's PATH and `BRANCH` names an available branch, when the README installer command runs with that `BRANCH`, then the installer builds `scip-search` from source and reports a successful source install.
- AC-001-2: Given the branch source install succeeds, when the caller looks for the installed command at the installer-selected binary location, then an executable named `scip-search` is present.
- AC-001-3: Given the branch source install succeeds, when the caller invokes the installed `scip-search` executable for installation verification, then the command can start without requiring a local clone to remain in the caller's current directory.
- AC-001-4: Given `BRANCH` is set, when the installer selects the install workflow, then the caller-observable behavior is a source build workflow requiring Go and make rather than a release artifact download workflow.
- AC-001-4b: Given `BRANCH` is set and Go is not available on the caller's PATH, when the installer validates source-build prerequisites, then it fails before claiming installation success and reports that Go is required for branch builds.
- AC-001-4c: Given `BRANCH` is set and make is not available on the caller's PATH, when the installer validates source-build prerequisites, then it fails before claiming installation success and reports that make is required for branch builds.
- AC-001-5: Given `BRANCH` names a branch that cannot be fetched or checked out, when the installer attempts the branch source workflow, then it fails with an actionable branch source-build error and does not report a successful installed binary.
- AC-001-6: Given the source build or install command fails after prerequisites pass, when the installer exits, then the caller sees a failed install outcome and no message that implies `scip-search` was successfully installed.

### Depends on:
Run time coupling:
- Interface I-001-001 - Installer invocation contract
- Interface I-003-001 - Source build contract

### Out of Scope
- Exact installer implementation strategy for retrieving source.
- Exact stderr text, numeric exit code taxonomy, or log formatting.
- Release binary URL construction, archive extraction, checksum verification, or fallback behavior.
- Exact `scip-search --version` fields.
- Running `symbols`, `packages`, `references`, or `implementations` as source install validation.

### Assumptions
- **ASM-001-1**: "Installation verification" for this story means the installed executable can be invoked as `scip-search`; exact version provenance belongs to CAP-003. - *Why*: The README verification command is shared, but this task excludes redefining version behavior. - Confidence: HIGH

### Open Questions
- None.
