# User Stories: Version Output Contract

Status: review

## Goal
`scip-search --version` verifies the installed executable by returning version information successfully without requiring a SCIP index.

## Parent Epic
specs/epics/readme/20260517-151311-epic-planning-5.md - Capability CAP-003

## Context
The README uses `scip-search --version` as the installation verification command. This document defines the observable CLI behavior for that verification path while keeping query success JSON and runtime error contracts owned by the shared query runtime stories.

## Personas
- **Automation Agent**: an AI or script-driven caller running terminal commands in constrained environments, needing reproducible install commands and a quick way to verify the installed binary.
- **CLI Maintainer**: a Go developer preparing releases for macOS and Linux users, needing packaging checks and README examples that match the supported install behavior.

## General information

Applies to: top-level `scip-search --version` invocation behavior and observable output.

### References
- goal spec: README.md#installation - Defines `scip-search --version` as the verification command after install.
- goal spec: README.md#what-is-scip-search - Defines `scip-search` as a thin Go binary with one-shot process behavior.
- parent epic: specs/epics/readme/20260517-151311-epic-planning-5.md#capability-cap-003---report-cli-version-information - Defines version output for install verification and release/source build identity.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/cli-command-routing-and-usage.md - Confirms query command routing excludes version output.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/cli-shared-index-flag.md - Confirms `--index` is required only for documented query commands.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/stdout-json-success-contract.md - Confirms JSON stdout is scoped to successful query commands, not distribution verification.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/stderr-exit-error-contract.md - Confirms runtime failure contracts are scoped to shared query runtime failures.

### Non-Functional Requirements
- NFR-000-1: Version verification must be automation-friendly: deterministic, non-interactive, and observable through stdout, stderr, and process status.
- NFR-000-2: Version verification must not require a SCIP index file, language indexer execution, compilation, type-checking, daemon startup, watch mode, or network lookup.

### Related External Components
- Component C-004 - README installation documentation: The user-facing installation and verification guidance maintained with the distribution behavior.

### Interfaces
- I-004-001 - Version verification contract (Interface 001 of Component C-004): The command-line version output users and release automation use to confirm the installed executable.

### Out of Scope
- Query result JSON schemas for `symbols`, `packages`, `references`, or `implementations`.
- Shared query runtime failure status taxonomy, stderr diagnostics, or `--index` validation.
- Semantic versioning policy beyond reporting the installed build identity.
- Network release lookups, health checks that require SCIP indexes, release hosting, package manager formulas, signing, notarization, and container images.

### Assumptions
- **ASM-000-1**: `--version` is a top-level flag accepted by `scip-search` without a query command. - *Why*: The README shows `scip-search --version` as installation verification rather than as a query invocation. - Confidence: HIGH
- **ASM-000-2**: Successful version output is distribution verification output, not a query result payload. - *Why*: Query JSON contracts explicitly exclude version output, and CAP-003 keeps this capability separate from query command behavior. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Run Version Verification Without an Index

### References
- goal spec: README.md#installation - Shows `scip-search --version` as the verification command.
- parent epic: specs/epics/readme/20260517-151311-epic-planning-5.md#capability-cap-003---report-cli-version-information - Requires top-level version behavior and non-index execution.

### User Story
**As an** automation agent verifying an installed binary in a constrained environment, **I want to** run `scip-search --version` without providing an index path, **so that** I can confirm the executable is installed before preparing or selecting SCIP indexes.

### Acceptance Criteria
- AC-001-1: Given `scip-search --version` is invoked without a query command, when the CLI validates the invocation, then the invocation is accepted as a version verification request rather than rejected as a missing query command.
- AC-001-2: Given `scip-search --version` is invoked without `--index`, when the CLI runs, then the process does not report a missing-index usage failure.
- AC-001-3: Given `scip-search --version` is invoked from a directory that contains no selected SCIP index, when the CLI runs, then it completes without attempting to load or validate a SCIP index file.
- AC-001-4: Given `scip-search --version` is a valid invocation, when the process exits, then it exits with status `0`.

### Depends on:
Run time coupling:
- Interface I-004-001 - Version verification contract

### Out of Scope
- Query command routing for `symbols`, `packages`, `references`, and `implementations`.
- Supporting `--version` as a query subcommand.
- Defining default SCIP index locations.

### Assumptions
- **ASM-001-1**: Version verification succeeds independently of the caller's current working directory contents. - *Why*: The README positions the command as install verification, not project or index validation. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-002 - Emit Observable Version Information

### References
- goal spec: README.md#installation - Uses the version command to verify the installed executable.
- parent epic: specs/epics/readme/20260517-151311-epic-planning-5.md#capability-cap-003---report-cli-version-information - Requires output that identifies the installed build well enough to confirm installation and release provenance.

### User Story
**As a** CLI maintainer validating installation instructions, **I want to** see version information for the installed `scip-search` executable, **so that** release checks can prove the command on PATH is the expected binary.

### Acceptance Criteria
- AC-002-1: Given `scip-search --version` completes successfully, when the caller reads stdout, then stdout contains non-empty version information that identifies the executable as `scip-search`.
- AC-002-2: Given `scip-search --version` completes successfully, when the caller reads stdout, then the output includes the installed build identity exposed by the binary.
- AC-002-3: Given `scip-search --version` completes successfully, when the caller reads stderr, then stderr is empty for that invocation.
- AC-002-4: Given `scip-search --version` completes successfully, when the caller reads stdout, then stdout is version verification output and not a query result for `symbols`, `packages`, `references`, or `implementations`.

### Depends on:
Implementation ordering:
- Story ST-001 - Run Version Verification Without an Index

Run time coupling:
- Interface I-004-001 - Version verification contract

### Out of Scope
- Exact wording, field separators, color, pretty-printing, or a machine-readable schema for version output.
- Query JSON payload fields, ordering, and empty-result representations.
- Warnings or diagnostics for failed query invocations.

### Assumptions
- **ASM-002-1**: Install verification only requires the output to be stable enough for a caller to recognize `scip-search` and the build identity; exact formatting can remain an implementation detail. - *Why*: The epic requires build identity but does not prescribe a version output schema. - Confidence: MEDIUM

### Open Questions
- None.
