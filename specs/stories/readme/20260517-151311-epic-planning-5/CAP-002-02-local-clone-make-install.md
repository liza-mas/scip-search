# User Stories: Local Clone Make Install

Status: review

## Goal
Source builders can run `make install` from a local clone and receive either an installed `scip-search` executable built from that checkout or an actionable source-build failure.

## Parent Epic
specs/epics/readme/20260517-151311-epic-planning-5.md - Capability CAP-002

## Context
This document covers the README local clone install path. It describes the user-observable behavior of `make install` for contributors and early adopters who already have a clone, Go, and make, without expanding into release downloads or query command behavior.

## Personas
- **Source Builder**: A developer or early adopter with Go and make available, installing from a branch or local clone before a release artifact exists.
- **CLI Maintainer**: A Go developer preparing distribution workflows, needing local build commands that contributors can run and maintain.

## General information

Applies to: source installation from an existing local clone through `make install`.

### References
- goal spec: README.md#installation - Defines cloning the repository and running `make install` from the local clone.
- goal spec: README.md#what-is-scip-search - Defines the installed artifact as the one-shot `scip-search` Go binary.
- parent epic: specs/epics/readme/20260517-151311-epic-planning-5.md#personas - Defines Source Builder and CLI Maintainer personas.
- parent epic: specs/epics/readme/20260517-151311-epic-planning-5.md#general-information - Defines distribution-wide NFRs, external components, interfaces, and out-of-scope boundaries.
- parent epic: specs/epics/readme/20260517-151311-epic-planning-5.md#capability-cap-002---build-and-install-from-source-workflows - Defines local clone source build behavior and failure visibility.
- task: epic-planning-5-us-writing-1 - Limits this story work to local clone make install behavior, Go/make prerequisites, installed CLI outcome, and source-build failures.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/cli-command-routing-and-usage.md - Read for local story format and scope-boundary style; no existing CAP-002 story documents were present before this task.

### Non-Functional Requirements
- NFR-000-1: Local clone installs must not provision Go, make, or language indexers on the caller's host.
- NFR-000-2: Local clone installs must be automation-friendly: success and failure outcomes are non-interactive and observable from the invoking shell.
- NFR-000-3: Local clone install behavior must stay separate from release artifact selection and download behavior.

### Related External Components
- Component C-003 - Local Go and make toolchain: The caller-provided tools required for source builds.
- Component C-004 - README installation documentation: The user-facing installation and verification guidance maintained with the distribution behavior.

### Interfaces
- I-003-001 - Source build contract (Interface 001 of Component C-003): The local clone build and install command surface used by source builders.

### Out of Scope
- Latest release, `VERSION`, or `BRANCH` installer behavior.
- Release artifact hosting, archive extraction, checksums, signing, notarization, package manager formulas, and CI release publication.
- Cross-compilation or installing for platforms beyond the current local toolchain.
- Query command validation, SCIP index loading, SCIP traversal, and query result schemas.
- Installing Go, make, or language indexers.
- Defining exact `scip-search --version` output content.

### Assumptions
- **ASM-000-1**: `make install` installs the code from the caller's current checkout, including whatever branch or commit the clone is already on. - *Why*: The README local clone instructions run `make install` after entering the clone and do not introduce another source selector. - Confidence: HIGH
- **ASM-000-2**: The local clone install path may validate the installed executable by confirming that a runnable `scip-search` binary exists, without defining the exact version string. - *Why*: CAP-003 owns version output behavior, while this capability owns the installed CLI outcome. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Build and Install From a Local Clone With Make

### References
- goal spec: README.md#installation - Defines `git clone`, entering the clone, and `make install`.
- parent epic: specs/epics/readme/20260517-151311-epic-planning-5.md#capability-cap-002---build-and-install-from-source-workflows - Defines local clone source build behavior and failure visibility.

### User Story
**As a** Source Builder working from a local clone, **I want to** run `make install` in that checkout, **so that** I can install the `scip-search` CLI produced from the source I am evaluating.

### Acceptance Criteria
- AC-001-1: Given the caller is in a local `scip-search` clone and Go and make are available on the caller's PATH, when the caller runs `make install`, then the command builds `scip-search` from the current checkout and reports a successful local source install.
- AC-001-2: Given `make install` succeeds, when the caller looks for the installed command at the make-selected binary location, then an executable named `scip-search` is present.
- AC-001-3: Given `make install` succeeds, when the caller invokes the installed `scip-search` executable for installation verification outside the clone directory, then the installed command can start without requiring the clone as the current working directory.
- AC-001-4: Given Go is not available on the caller's PATH, when the caller runs `make install`, then the command fails before claiming installation success and reports that Go is required for local source installs.
- AC-001-4b: Given make is not available on the caller's PATH, when the caller attempts the README local clone workflow, then the workflow cannot proceed through `make install` and the caller is not directed to any host tool provisioning performed by `scip-search`.
- AC-001-5: Given the source build or install step fails after prerequisites pass, when `make install` exits, then the caller sees a failed install outcome and no message that implies `scip-search` was successfully installed.
- AC-001-6: Given the caller runs `make install` from a local clone, when the workflow selects its install behavior, then the caller-observable behavior is a local source build rather than a release artifact download workflow.

### Depends on:
Run time coupling:
- Interface I-003-001 - Source build contract

### Out of Scope
- The preceding `git clone` operation beyond the README setup context.
- Exact make target implementation, build flags, install directory defaults, stderr text, or numeric exit code taxonomy.
- Release artifact URL construction, archive extraction, checksum verification, or fallback behavior.
- Exact `scip-search --version` fields.
- Running `symbols`, `packages`, `references`, or `implementations` as local source install validation.

### Assumptions
- **ASM-001-1**: "Installation verification" for this story means the installed executable can be invoked as `scip-search`; exact version provenance belongs to CAP-003. - *Why*: The README verification command is shared, but this task excludes redefining version behavior. - Confidence: HIGH

### Open Questions
- None.
