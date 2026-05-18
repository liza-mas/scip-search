# User Stories: stderr and Exit Status Failure Contract

Status: review

## Goal
Every shared runtime failure writes diagnostics only to stderr, writes no success JSON to stdout, and exits with a documented nonzero status.

## Parent Epic
specs/epics/readme/20260517-095535-epic-planning-1.md - Capability CAP-003

## Context
Automation callers need to distinguish failed invocations from successful empty results without parsing mixed streams. This document defines the shared failure stream and status contract for usage failures and index-loading failures inherited from CAP-001 and CAP-002.

## Personas
- **Automation agent**: A coding or planning agent running terminal commands inside concurrent worktrees, needing stable machine-readable output and deterministic failures.
- **CLI developer**: A Go developer implementing and testing `scip-search`, needing a bounded runtime contract that separates CLI behavior from query traversal.

## General information

Applies to: shared runtime failures for documented query commands.

### References
- goal spec: README.md#what-is-scip-search - Defines one-shot process behavior with structured JSON stdout on success.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-003---report-machine-readable-results-and-actionable-failures - Defines diagnostic-only stderr and consistent process status for failures.
- dependency story: specs/stories/readme/20260517-095535-epic-planning-1/cli-command-routing-and-usage.md - Defines missing and unsupported command usage failures.
- dependency story: specs/stories/readme/20260517-095535-epic-planning-1/cli-shared-index-flag.md - Defines missing `--index` and missing index value usage failures.
- dependency story: specs/stories/readme/20260517-095535-epic-planning-1/index-path-validation.md - Defines missing and unreadable selected index failures.
- dependency story: specs/stories/readme/20260517-095535-epic-planning-1/scip-index-loading-boundary.md - Defines invalid SCIP input failures.
- task: epic-planning-1-us-writing-2 - Requires stderr plus process status for all runtime failures.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/ - Existing CAP-001 and CAP-002 story documents were read to preserve their failure boundaries.

### Non-Functional Requirements
- NFR-000-1: Failure diagnostics must be automation-friendly by keeping stdout empty on shared runtime failures.
- NFR-000-2: The same shared failure class must use the same documented process status regardless of which documented query command triggered it.
- NFR-000-3: Failure handling must not require an interactive prompt, daemon lifecycle, watch loop, or retry session.

### Related External Components
- Component C-001 - SCIP index file: The on-disk input selected by the caller with `--index`.
- Component C-003 - Calling process environment: The shell, script, or agent invoking `scip-search` and observing stdout, stderr, and process status.

### Interfaces
- I-003-001 - CLI process contract (Interface 001 of Component C-003): The command-line invocation, stdout/stderr streams, and process exit status exposed by `scip-search`.

### Out of Scope
- Exact diagnostic wording beyond user-observable requirements for stream placement and actionable cause.
- Query-specific validation failures for `--name`, `--symbol`, `--prefix`, or query result semantics.
- Pretty-printing, progress output, configurable logging, human UI output, distribution documentation, install verification, and version output.

### Assumptions
- **ASM-000-1**: Shared usage failures use exit status `2`, and shared index-loading failures use exit status `3`. - *Why*: The epic calls for documented nonzero status, prior stories define distinct failure sources, and distinct statuses let automation separate invalid invocation from unusable index input without parsing stderr. - Confidence: MEDIUM
- **ASM-000-2**: Failure diagnostics are human-readable stderr text rather than structured JSON unless a future story explicitly defines a machine-readable error schema. - *Why*: The capability says structured JSON belongs to successful stdout and diagnostics belong to stderr. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Report Shared Runtime Failures on stderr Only

### References
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-003---report-machine-readable-results-and-actionable-failures - Requires diagnostic-only stderr for failures.
- dependency story: specs/stories/readme/20260517-095535-epic-planning-1/cli-command-routing-and-usage.md - Defines usage failures before traversal.
- dependency story: specs/stories/readme/20260517-095535-epic-planning-1/index-path-validation.md - Defines index path validation failures.
- dependency story: specs/stories/readme/20260517-095535-epic-planning-1/scip-index-loading-boundary.md - Defines invalid SCIP loading failures.

### User Story
**As an** automation agent running `scip-search` from scripts, **I want to** receive runtime failure diagnostics only on stderr, **so that** stdout remains a reliable success-data channel.

### Acceptance Criteria
- AC-001-1: Given a shared usage failure occurs, when the caller reads process streams, then stdout is empty and stderr contains a diagnostic for that failure.
- AC-001-2: Given a shared index-loading failure occurs, when the caller reads process streams, then stdout is empty and stderr contains a diagnostic for that failure.
- AC-001-3: Given any shared runtime failure occurs, when the caller reads stderr, then stderr does not contain a successful query result payload.
- AC-001-4: Given any shared runtime failure occurs, when the caller reads stdout, then stdout does not contain partial JSON, empty-result JSON, progress text, warnings, or diagnostics.

### Depends on:
Implementation ordering:
- Story document cli-command-routing-and-usage.md - Usage failure sources must be defined before shared failure stream placement can be tested.
- Story document cli-shared-index-flag.md - Shared index flag failures must be defined before shared failure stream placement can be tested.
- Story document index-path-validation.md - Index path failures must be defined before shared failure stream placement can be tested.
- Story document scip-index-loading-boundary.md - Invalid SCIP input failures must be defined before shared failure stream placement can be tested.

Run time coupling:
- Interface I-003-001 - CLI process contract

### Out of Scope
- Query-specific argument validation failures.
- Structured stderr error schemas.
- Recovery suggestions that require inspecting source code or generating indexes.

### Assumptions
- **ASM-001-1**: A diagnostic is sufficient when it identifies the shared failure class and the caller-owned input that caused the failure when such input exists. - *Why*: The capability requires actionable failures without prescribing exact text. - Confidence: MEDIUM

### Open Questions
- None.

---

## Story ST-002 - Use Documented Nonzero Status for Failures

### References
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-003---report-machine-readable-results-and-actionable-failures - Requires consistent failure reporting through process status.
- dependency story: specs/stories/readme/20260517-095535-epic-planning-1/cli-one-shot-process-lifecycle.md - Defines that each invocation exits after success or failure.

### User Story
**As a** CLI developer implementing the runtime contract, **I want to** assign documented nonzero process statuses to shared runtime failures, **so that** automation can branch on process status without parsing stderr first.

### Acceptance Criteria
- AC-002-1: Given a missing command or unsupported command usage failure occurs, when the process exits, then it exits with status `2`.
- AC-002-2: Given a missing `--index` flag or missing `--index` value usage failure occurs, when the process exits, then it exits with status `2`.
- AC-002-3: Given a nonexistent, unreadable, directory, or invalid SCIP selected index failure occurs, when the process exits, then it exits with status `3`.
- AC-002-4: Given the same shared failure class occurs for different documented query commands, when each process exits, then each process uses the same documented nonzero status for that failure class.
- AC-002-5: Given any shared runtime failure occurs, when the caller observes the process status, then the status is not the documented success status.

### Depends on:
Implementation ordering:
- Story ST-001 - Report Shared Runtime Failures on stderr Only
- Story document cli-one-shot-process-lifecycle.md - The process must terminate after failures before exit status can be observed.

Run time coupling:
- Interface I-003-001 - CLI process contract

### Out of Scope
- Query-specific validation status for result-filter arguments.
- Shell-specific signal, timeout, or process-killed statuses outside normal `scip-search` runtime failures.

### Assumptions
- **ASM-002-1**: Status `2` is reserved for shared invocation shape failures, while status `3` is reserved for shared failures that occur after an index path is selected but before query traversal starts. - *Why*: The runtime needs statuses that are documented, nonzero, and observable without query-specific behavior. - Confidence: MEDIUM

### Open Questions
- None.
