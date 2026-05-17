# User Stories: CLI Shared Index Flag

Status: draft

## Goal
Every documented `scip-search` query command requires the shared `--index <index-path>` flag and handles missing or incomplete index flag usage consistently before query-specific behavior runs.

## Parent Epic
specs/epics/readme/20260517-095535-epic-planning-1.md - Capability CAP-001

## Context
The goal spec shows `--index` on every query command. This story document defines only the shared flag contract at invocation time; path validation and SCIP parsing belong to the index-loading capability.

## Personas
- **Automation agent**: A coding or planning agent running terminal commands inside concurrent worktrees, needing stable machine-readable output and deterministic failures.
- **CLI developer**: A Go developer implementing and testing `scip-search`, needing a bounded runtime contract that separates CLI behavior from query traversal.

## General information

Applies to: the shared `--index <index-path>` flag on `symbols`, `references`, `implementations`, and `packages`.

### References
- goal spec: README.md#what-is-scip-search, lines 58-80 - Lists each query command with the shared `--index <index-path>` flag.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#personas, lines 11-14 - Defines the automation agent and CLI developer personas.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#general-information, lines 15-57 - Defines explicit index-path input as an epic-wide runtime constraint.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-001---invoke-scip-search-commands-consistently, lines 59-93 - Includes shared flag behavior in the command invocation capability.
- consistency check: specs/stories/readme/ - No existing story documents were present in this domain when checked.

### Non-Functional Requirements
- NFR-000-1: The CLI must support concurrent worktrees by requiring callers to provide an explicit index path instead of assuming a global or repository-local index location.
- NFR-000-2: Shared flag failures must be deterministic for automation and must occur before SCIP traversal.

### Related External Components
Summary of all the external components referenced by this document:
- Component C-001 - SCIP index file: The on-disk input selected by the caller with `--index`.
- Component C-003 - Calling process environment: The shell, script, or agent that invokes `scip-search` and observes stdout, stderr, and exit status.

### Interfaces
Summary of all the external interfaces referenced by this document:
- I-003-001 - CLI process contract (Interface 001 of Component C-003): The command-line invocation, stdout/stderr streams, and process exit status exposed by `scip-search`.

### Out of Scope
- Validating whether the supplied index path exists, is readable, or contains valid SCIP data.
- Loading or traversing a SCIP index.
- Query-specific flags such as `--name`, `--symbol`, and `--prefix`.
- Query-specific result fields and filtering.
- Installation, release packaging, version command behavior, and user-facing distribution docs.
- Exact stderr wording, stdout JSON schema, and numeric exit-code taxonomy.

### Assumptions
- **ASM-000-1**: `--index` is required for all four query commands rather than defaulting to a conventional file path. - *Why*: The goal spec examples always pass `--index`, and the epic NFR requires explicit index paths for concurrent worktrees. - Confidence: HIGH
- **ASM-000-2**: The invocation contract requires a value after `--index`, but this story does not define what makes that value a valid SCIP file. - *Why*: CAP-001 owns shared flag shape, while CAP-002 owns index path validation and loading. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Require the Shared Index Flag

### References
- goal spec: README.md#what-is-scip-search, lines 64-72 - Shows every query command using `--index <index-path>`.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-001---invoke-scip-search-commands-consistently, lines 59-93 - Requires shared flag behavior across documented commands.

### User Story
**As an** automation agent, **I want to** pass an explicit SCIP index path with every query command, **so that** concurrent worktrees can select the intended index without relying on ambient repository state.

### Acceptance Criteria
- AC-001-1: Given `scip-search symbols` is invoked without `--index <index-path>`, when the runtime parses the invocation, then it fails through the shared usage-failure path.
- AC-001-2: Given `scip-search references` is invoked without `--index <index-path>`, when the runtime parses the invocation, then it fails through the shared usage-failure path.
- AC-001-3: Given `scip-search implementations` is invoked without `--index <index-path>`, when the runtime parses the invocation, then it fails through the shared usage-failure path.
- AC-001-4: Given `scip-search packages` is invoked without `--index <index-path>`, when the runtime parses the invocation, then it fails through the shared usage-failure path.
- AC-001-5: Given any documented query command includes `--index <index-path>`, when the runtime parses the invocation, then the caller does not receive a missing-index-flag or incomplete-index-flag usage failure.

### Depends on:
Implementation ordering:
- Story document specs/stories/readme/cli-command-routing-and-usage.md - The documented command names must exist before the shared flag can be required across them.

Run time coupling:
- Interface I-003-001 - CLI process contract

### Out of Scope
- Determining whether `<index-path>` exists or points to valid SCIP data.
- Query-specific flags and query traversal.

### Assumptions
- None.

### Open Questions
- None.

---

## Story ST-002 - Reject Incomplete Index Flag Usage

### References
- goal spec: README.md#what-is-scip-search, lines 64-72 - Documents `--index` as a flag with a required `<index-path>` value.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-001---invoke-scip-search-commands-consistently, lines 59-93 - Requires invalid invocation shape to be rejected consistently.

### User Story
**As an** automation agent, **I want to** receive a deterministic usage failure when `--index` has no path value, **so that** malformed commands do not proceed as ambiguous query executions.

### Acceptance Criteria
- AC-002-1: Given any documented query command is invoked with `--index` and no following value, when the runtime parses the invocation, then it fails through the shared usage-failure path.
- AC-002-2: Given any documented query command is invoked with an incomplete `--index` flag, when the runtime handles the failure, then the caller receives a usage failure rather than an index-loading failure or query result.
- AC-002-3: Given any documented query command is invoked with an incomplete `--index` flag, when the runtime handles the failure, then control returns to the calling process as a completed one-shot invocation.

### Depends on:
Implementation ordering:
- Story document specs/stories/readme/cli-command-routing-and-usage.md - The documented command names must exist before incomplete shared flag usage can be tested across them.

Run time coupling:
- Interface I-003-001 - CLI process contract

### Out of Scope
- Exact diagnostic wording, stdout/stderr stream contract, and numeric exit status details.
- Query-specific missing-flag failures.

### Assumptions
- None.

### Open Questions
- None.
