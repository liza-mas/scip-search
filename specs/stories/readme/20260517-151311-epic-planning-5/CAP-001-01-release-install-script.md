# User Stories: Install Released Binaries with VERSION and INSTALL_DIR

Status: review

## Goal
`scip-search` release installer users can install the latest release or a requested release version on macOS/Linux into the intended binary directory and observe the installed executable.

## Parent Epic
specs/epics/readme/20260517-151311-epic-planning-5.md - Capability CAP-001

## Context
This document covers the successful release-install path from the README install command. It defines the user-visible behavior for latest-release installs, explicit `VERSION` installs, and caller-selected `INSTALL_DIR` installs without pulling in source-build, branch-build, package-manager, query execution, or SCIP index generation workflows.

## Personas
- **Automation Agent**: An AI or script-driven caller running terminal commands in constrained macOS or Linux environments, needing reproducible install commands and a quick way to find the installed binary.
- **CLI Maintainer**: A Go developer preparing releases for macOS and Linux users, needing installer behavior that can be validated against release artifacts and README examples.

## General information

Applies to: successful release binary installation through the installer invocation contract.

### References
- goal spec: README.md#installation, lines 131-148 - Defines latest release install, `VERSION`, and `INSTALL_DIR` examples.
- parent epic: specs/epics/readme/20260517-151311-epic-planning-5.md#capability-cap-001---install-released-scip-search-binaries - Defines release install scope, success observability, and exclusions.
- task: epic-planning-5-us-writing-0 - Requires latest-release, explicit `VERSION`, `INSTALL_DIR`, executable availability, and verification scope without local clone or SCIP index generation.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/cli-command-routing-and-usage.md - Read for story format and scope-boundary style.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/stderr-exit-error-contract.md - Read for failure-boundary style used by sibling runtime stories.

### Non-Functional Requirements
- NFR-000-1: Release install workflows must be usable from non-interactive shell automation on macOS and Linux.
- NFR-000-2: Release install success must not require Go, make, a local repository clone, language indexer installation, or SCIP index generation.
- NFR-000-3: Release install behavior must remain separate from query command behavior and query-specific runtime validation.

### Related External Components
- Component C-001 - Release installer: The user-facing `install.sh` workflow invoked from the README for latest release, specific version, and custom directory installs.
- Component C-002 - Release artifacts: The published binaries or archives consumed by release install workflows.
- Component C-003 - Calling process environment: The macOS/Linux shell environment invoking the installer and later resolving or executing `scip-search`.

### Interfaces
- I-001-001 - Installer invocation contract (Interface 001 of Component C-001): The shell invocation and environment options users rely on for release and custom-directory installs.

### Out of Scope
- Branch builds, local clone builds, Go or make source-build behavior, and `BRANCH` handling.
- Installing SCIP language indexers, generating SCIP indexes, or validating query execution.
- Package manager formulas, container images, signing, notarization, and release hosting setup.
- Exact release artifact archive layout, checksum implementation, or download transport internals.

### Assumptions
- **ASM-000-1**: The installer has an installer-defined default binary directory when `INSTALL_DIR` is not set. - *Why*: The README latest-release and `VERSION` examples omit `INSTALL_DIR`, while the custom-directory example sets it explicitly. - Confidence: HIGH
- **ASM-000-2**: Successful release installs identify the installed binary path in a user-observable way. - *Why*: The capability requires observable installed-executable success, and automation needs to know which executable to verify. - Confidence: MEDIUM

### Open Questions
- None.

---

## Story ST-001 - Install the Latest Released Binary by Default

### References
- goal spec: README.md#installation, lines 131-134 - Defines the quick install command as the latest release path for macOS/Linux.
- parent epic: specs/epics/readme/20260517-151311-epic-planning-5.md#capability-cap-001---install-released-scip-search-binaries - Defines default latest release selection and installed executable success.

### User Story
**As an** Automation Agent running the README install command on macOS or Linux, **I want to** install the latest released `scip-search` binary without naming a version, **so that** I can provision the CLI in a fresh environment without cloning the repository.

### Acceptance Criteria
- AC-001-1: Given the installer is invoked on supported macOS or Linux without `VERSION`, when the install succeeds, then the installed `scip-search` executable corresponds to the latest released version available to the installer.
- AC-001-2: Given the latest-release install succeeds, when the caller inspects the installer result, then the caller can identify the filesystem path of the installed `scip-search` executable.
- AC-001-3: Given the latest-release install succeeds, when the caller executes the installed file directly, then the operating system treats it as an executable command.
- AC-001-4: Given the latest-release install succeeds, when the caller reviews required local inputs, then the workflow has not required a local `scip-search` clone, Go, make, language indexers, or a SCIP index file.

### Depends on:
Run time coupling:
- Interface I-001-001 - Installer invocation contract

### Out of Scope
- Choosing release hosting URLs or artifact naming conventions.
- Verifying query commands after install.
- Installing from a branch or local clone.

### Assumptions
- **ASM-001-1**: "Latest released version" means the newest published release artifact visible to the installer at install time, not the repository's default branch. - *Why*: The README labels this path "latest release" and assigns branch builds to a separate option. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-002 - Install an Explicit Released Version

### References
- goal spec: README.md#installation, lines 137-138 - Defines `VERSION=v1.0.0` as the specific-version install option.
- parent epic: specs/epics/readme/20260517-151311-epic-planning-5.md#capability-cap-001---install-released-scip-search-binaries - Requires explicit version release install behavior.

### User Story
**As an** Automation Agent pinning tool versions in a scripted environment, **I want to** set `VERSION` when invoking the release installer, **so that** repeated installs use the requested `scip-search` release instead of drifting to the latest release.

### Acceptance Criteria
- AC-002-1: Given the installer is invoked on supported macOS or Linux with `VERSION` set to an available released version, when the install succeeds, then the installed `scip-search` executable corresponds to that requested release version.
- AC-002-2: Given `VERSION` is set to an available released version, when the install succeeds, then the installer does not silently install a different latest release instead.
- AC-002-3: Given an explicit-version install succeeds, when the caller inspects the installer result, then the caller can identify the requested version and the filesystem path of the installed executable.
- AC-002-4: Given an explicit-version install succeeds, when the caller reviews required local inputs, then the workflow has not required a local `scip-search` clone, Go, make, language indexers, or a SCIP index file.

### Depends on:
Run time coupling:
- Interface I-001-001 - Installer invocation contract

### Out of Scope
- Installing unreleased branches or commit SHAs through `VERSION`.
- Defining semantic-version parsing rules beyond accepting a released version identifier supported by the installer.
- Defining the exact `--version` output format used for post-install verification.

### Assumptions
- **ASM-002-1**: `VERSION` refers only to released `scip-search` versions, not branch names or arbitrary Git refs. - *Why*: The README separates "Specific version" from "Build from a branch". - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-003 - Install into a Caller-Selected Directory

### References
- goal spec: README.md#installation, lines 146-148 - Defines `INSTALL_DIR=~/.local/bin` as the custom-directory install option.
- parent epic: specs/epics/readme/20260517-151311-epic-planning-5.md#capability-cap-001---install-released-scip-search-binaries - Requires chosen binary directory behavior.

### User Story
**As an** Automation Agent installing tools in a constrained environment, **I want to** set `INSTALL_DIR` for the release installer, **so that** I can place `scip-search` in a binary directory controlled by my workflow.

### Acceptance Criteria
- AC-003-1: Given the installer is invoked on supported macOS or Linux with `INSTALL_DIR` set to a usable binary directory, when the install succeeds, then the `scip-search` executable is installed in that requested directory.
- AC-003-2: Given `INSTALL_DIR` is set and the install succeeds, when the caller executes `<INSTALL_DIR>/scip-search`, then the operating system treats it as an executable command.
- AC-003-3: Given `INSTALL_DIR` is set together with `VERSION`, when the install succeeds, then the requested release version is installed in the requested directory.
- AC-003-4: Given `INSTALL_DIR` is set, when the install succeeds, then success is observable without relying on the directory already being on `PATH`.

### Depends on:
Run time coupling:
- Interface I-001-001 - Installer invocation contract

### Out of Scope
- Editing shell profile files or permanently modifying `PATH`.
- Defining ownership, permission repair, or directory-creation policy for unusable install directories.
- Installing additional helper binaries.

### Assumptions
- **ASM-003-1**: A "usable binary directory" is a caller-selected location where the installer can place an executable file for the current user. - *Why*: The README allows custom directories but does not specify permission or creation behavior. - Confidence: MEDIUM

### Open Questions
- None.
