# User Stories: Build Metadata for Release and Source Builds

Status: review

## Goal
`scip-search --version` distinguishes released binaries from source-built binaries well enough for users and automation to confirm build provenance.

## Parent Epic
specs/epics/readme/20260517-151311-epic-planning-5.md - Capability CAP-003

## Context
The distribution epic includes release install, branch build, and local clone workflows. This document defines the provenance requirements for the build identity shown by `scip-search --version` without prescribing how release or source metadata is embedded.

## Personas
- **Automation Agent**: an AI or script-driven caller running terminal commands in constrained environments, needing reproducible install commands and a quick way to verify the installed binary.
- **CLI Maintainer**: a Go developer preparing releases for macOS and Linux users, needing packaging checks and README examples that match the supported install behavior.
- **Source Builder**: a developer or early adopter with Go and make available, installing from a branch or local clone before a release artifact exists.

## General information

Applies to: release and source build identity exposed by `scip-search --version`.

### References
- goal spec: README.md#installation - Defines latest release, specific version, branch build, local clone install, and `scip-search --version` verification workflows.
- parent epic: specs/epics/readme/20260517-151311-epic-planning-5.md#capability-cap-003---report-cli-version-information - Requires version output for release and source build provenance.
- related story: specs/stories/readme/20260517-151311-epic-planning-5/CAP-003-01-version-output-contract.md - Defines the top-level version invocation and stream/status behavior this provenance appears through.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/stdout-json-success-contract.md - Confirms successful query JSON schemas are out of scope for version output.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/stderr-exit-error-contract.md - Confirms shared runtime error contracts are out of scope for version output.

### Non-Functional Requirements
- NFR-000-1: Build provenance must be available offline from the installed binary when `scip-search --version` runs.
- NFR-000-2: Build provenance must support automation comparing the observed version output with the install workflow that produced the binary.

### Related External Components
- Component C-001 - Release installer: The user-facing `install.sh` workflow invoked from the README for latest release, specific version, branch, and custom-directory installs.
- Component C-002 - Release artifacts: The published binaries or archives consumed by release install workflows.
- Component C-003 - Local Go and make toolchain: The caller-provided tools required for branch and local clone source builds.
- Component C-004 - README installation documentation: The user-facing installation and verification guidance maintained with the distribution behavior.

### Interfaces
- I-001-001 - Installer invocation contract (Interface 001 of Component C-001): The shell invocation and environment options users rely on for release, branch, and custom-directory installs.
- I-003-001 - Source build contract (Interface 001 of Component C-003): The local clone build and install command surface used by source builders.
- I-004-001 - Version verification contract (Interface 001 of Component C-004): The command-line version output users and release automation use to confirm the installed executable.

### Out of Scope
- Semantic versioning policy beyond reporting the installed build identity.
- Network calls to confirm releases, fetch registries, or compare against remote source state during `--version`.
- Release hosting provider setup, CI provider selection, signing, notarization, package manager formulas, container images, and operating systems beyond macOS/Linux install scope.
- Query payload schemas, query fixtures, `--index` handling, runtime stderr diagnostics, and runtime failure status taxonomy.

### Assumptions
- **ASM-000-1**: Release builds and source builds may expose different kinds of build identity, as long as the caller can distinguish release provenance from source provenance. - *Why*: The README defines both release and source workflows, while the epic avoids prescribing a specific metadata implementation. - Confidence: HIGH
- **ASM-000-2**: Source build provenance can include an explicit source-build marker plus whatever source ref or revision the build workflow supplies. - *Why*: Branch and local clone workflows may not have the same release version identity as published artifacts. - Confidence: MEDIUM

### Open Questions
- None.

---

## Story ST-001 - Report Release Build Identity

### References
- goal spec: README.md#installation - Defines latest release and specific `VERSION` install workflows followed by `scip-search --version` verification.
- parent epic: specs/epics/readme/20260517-151311-epic-planning-5.md#capability-cap-003---report-cli-version-information - Requires installed build identity for release provenance.

### User Story
**As an** automation agent installing a released macOS or Linux binary, **I want to** see the release build identity in `scip-search --version`, **so that** I can verify the installed executable matches the requested release workflow.

### Acceptance Criteria
- AC-001-1: Given a released `scip-search` binary was installed through the latest-release workflow, when the caller runs `scip-search --version`, then the output identifies the installed executable as a release build.
- AC-001-2: Given a released `scip-search` binary was installed through `VERSION=<release>`, when the caller runs `scip-search --version`, then the output includes the requested release identity.
- AC-001-3: Given release automation validates an installed binary, when it reads `scip-search --version`, then it can compare the observed build identity with the release artifact being validated without querying the network.
- AC-001-3b: Given a released binary was installed into a custom install directory, when that executable is invoked with `--version`, then the output still reports the build identity of that executable rather than the install directory path.

### Depends on:
Implementation ordering:
- Story document CAP-003-01-version-output-contract.md - The version command invocation must exist before release build identity can be observed through it.

Run time coupling:
- Interface I-001-001 - Installer invocation contract
- Interface I-004-001 - Version verification contract

### Out of Scope
- Defining the release numbering scheme or whether versions follow semantic versioning.
- Checking remote release registries or update availability.
- Signing, checksum, notarization, package manager, and container image behavior.

### Assumptions
- **ASM-001-1**: "Requested release identity" means the release identifier selected by the install workflow, not a promise about semantic version ordering. - *Why*: The task excludes semantic versioning policy beyond reported build identity. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-002 - Report Source Build Provenance

### References
- goal spec: README.md#installation - Defines branch build and local clone install workflows that require Go and make.
- parent epic: specs/epics/readme/20260517-151311-epic-planning-5.md#capability-cap-003---report-cli-version-information - Requires source build provenance in version output.

### User Story
**As a** source builder installing from a branch or local clone, **I want to** see source build provenance in `scip-search --version`, **so that** I can distinguish my locally built executable from a published release artifact.

### Acceptance Criteria
- AC-002-1: Given `scip-search` was built from a requested branch workflow, when the caller runs `scip-search --version`, then the output identifies the executable as a source build rather than a released binary.
- AC-002-2: Given `scip-search` was installed from a local clone with `make install`, when the caller runs `scip-search --version`, then the output includes source provenance available from that local build.
- AC-002-3: Given a source-built executable lacks a published release identity, when the caller runs `scip-search --version`, then the output does not masquerade as a released binary.
- AC-002-4: Given an automation workflow compares source-built and release-installed executables, when it reads each executable's version output, then it can distinguish which executable came from source and which came from a release artifact.

### Depends on:
Implementation ordering:
- Story document CAP-003-01-version-output-contract.md - The version command invocation must exist before source build provenance can be observed through it.

Run time coupling:
- Interface I-003-001 - Source build contract
- Interface I-004-001 - Version verification contract

### Out of Scope
- Requiring a network lookup to verify the branch, commit, or release state.
- Requiring a SCIP index, language indexer, generated fixture, or query command to validate source provenance.
- Prescribing linker flags, build variables, file formats, or metadata storage mechanisms.

### Assumptions
- **ASM-002-1**: A source build can be considered distinguishable when the version output clearly avoids claiming release provenance and includes available source-build identity. - *Why*: The epic requires enough identity to confirm source build provenance but does not mandate exact metadata fields. - Confidence: MEDIUM

### Open Questions
- None.
