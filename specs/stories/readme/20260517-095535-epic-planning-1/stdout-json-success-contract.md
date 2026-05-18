# User Stories: JSON Stdout Success Contract

Status: review

## Goal
Every successful `scip-search` query command writes only structured JSON to stdout, leaves stderr available for diagnostics, and exits with the documented success status.

## Parent Epic
specs/epics/readme/20260517-095535-epic-planning-1.md - Capability CAP-003

## Context
`scip-search` is used by automation that reads stdout as data. This document defines the shared success stream contract for all documented query commands while leaving each command's result fields to query-specific story documents.

## Personas
- **Automation agent**: A coding or planning agent running terminal commands inside concurrent worktrees, needing stable machine-readable output and deterministic failures.
- **CLI developer**: A Go developer implementing and testing `scip-search`, needing a bounded runtime contract that separates CLI behavior from query traversal.

## General information

Applies to: successful runtime paths for documented query commands.

### References
- goal spec: README.md#what-is-scip-search - Defines `scip-search` as a one-shot binary that loads an index, answers a query, prints structured JSON to stdout, and exits.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-003---report-machine-readable-results-and-actionable-failures - Defines JSON-only stdout on success and shared stream/status conventions.
- dependency story: specs/stories/readme/20260517-095535-epic-planning-1/cli-command-routing-and-usage.md - Defines the documented query command names.
- dependency story: specs/stories/readme/20260517-095535-epic-planning-1/cli-one-shot-process-lifecycle.md - Defines one selected command per process invocation.
- dependency story: specs/stories/readme/20260517-095535-epic-planning-1/scip-index-loading-boundary.md - Defines the loaded index boundary before query-specific execution.
- task: epic-planning-1-us-writing-2 - Limits this story work to shared stdout, stderr, exit-status, and runtime error conventions.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/ - Existing CAP-001 and CAP-002 story documents were read to preserve their invocation, lifecycle, and index-loading boundaries.

### Non-Functional Requirements
- NFR-000-1: Successful stdout must be suitable for automation consumption and must not contain progress text, warnings, logging, prompts, or human-oriented decoration.
- NFR-000-2: The shared success stream contract must apply consistently to `symbols`, `references`, `implementations`, and `packages`.

### Related External Components
- Component C-001 - SCIP index file: The on-disk input selected by the caller with `--index`.
- Component C-003 - Calling process environment: The shell, script, or agent invoking `scip-search` and observing stdout, stderr, and process status.

### Interfaces
- I-003-001 - CLI process contract (Interface 001 of Component C-003): The command-line invocation, stdout/stderr streams, and process exit status exposed by `scip-search`.

### Out of Scope
- Query-specific result field schemas for `symbols`, `packages`, `references`, or `implementations`.
- Pretty-printing, progress output, configurable logging, human UI output, version output, install behavior, release packaging, and distribution documentation.
- Failure diagnostics, nonzero failure status taxonomy, and shared runtime error case coverage, which are covered by sibling CAP-003 story documents.

### Assumptions
- **ASM-000-1**: A successful process status means exit status `0`. - *Why*: The epic requires documented process status and nonzero failure status; `0` is the standard CLI success status visible to automation callers. - Confidence: HIGH
- **ASM-000-2**: "Structured JSON" means stdout can be parsed as one complete JSON value for the selected command result. - *Why*: The source requires structured JSON for automation but leaves each query payload schema to sibling query epics. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Emit Parseable JSON on Successful Queries

### References
- goal spec: README.md#what-is-scip-search - Requires successful commands to print structured JSON to stdout.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-003---report-machine-readable-results-and-actionable-failures - Requires JSON-only stdout on success.

### User Story
**As an** automation agent parsing command output in a worktree, **I want to** receive one parseable JSON result on stdout when a query succeeds, **so that** downstream automation can consume successful results without filtering terminal text.

### Acceptance Criteria
- AC-001-1: Given any documented query command completes successfully, when the caller reads stdout, then stdout contains exactly one complete structured JSON value for that selected command's result.
- AC-001-2: Given a successful `symbols` command, when the caller parses stdout as JSON, then parsing succeeds without removing any non-JSON prefix, suffix, prompt, warning, or progress text.
- AC-001-3: Given a successful `references` command, when the caller parses stdout as JSON, then parsing succeeds without removing any non-JSON prefix, suffix, prompt, warning, or progress text.
- AC-001-4: Given a successful `implementations` command, when the caller parses stdout as JSON, then parsing succeeds without removing any non-JSON prefix, suffix, prompt, warning, or progress text.
- AC-001-5: Given a successful `packages` command, when the caller parses stdout as JSON, then parsing succeeds without removing any non-JSON prefix, suffix, prompt, warning, or progress text.

### Depends on:
Implementation ordering:
- Story document cli-command-routing-and-usage.md - The documented query commands must exist before their shared success stream behavior can be tested.
- Story document scip-index-loading-boundary.md - Successful query execution depends on the runtime reaching the loaded index boundary first.

Run time coupling:
- Interface I-003-001 - CLI process contract

### Out of Scope
- Defining the fields, ordering, or empty-result representation inside each query-specific JSON payload.
- Supporting multiple JSON result documents in one invocation.
- Adding progress or debug output to stdout.

### Assumptions
- **ASM-001-1**: The success contract permits each query command to choose its own JSON object or array shape in its own story document. - *Why*: CAP-003 owns stream purity while query epics own result schemas. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-002 - Keep Diagnostics off Successful Output Streams

### References
- goal spec: README.md#what-is-scip-search - Describes stdout as the destination for structured JSON.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-003---report-machine-readable-results-and-actionable-failures - Separates success stdout from failure diagnostics.

### User Story
**As a** CLI developer implementing the shared runtime, **I want to** keep successful command diagnostics out of stdout and stderr, **so that** automation callers can treat stdout as data and stderr as failure diagnostics.

### Acceptance Criteria
- AC-002-1: Given any documented query command completes successfully, when the caller reads stdout, then stdout contains no progress messages, logging lines, warnings, prompts, or explanatory text outside the structured JSON result.
- AC-002-2: Given any documented query command completes successfully, when the caller reads stderr, then stderr is empty for that invocation.
- AC-002-3: Given any documented query command completes successfully, when the caller observes process status, then the process exits with the documented success status.

### Depends on:
Implementation ordering:
- Story ST-001 - Emit Parseable JSON on Successful Queries

Run time coupling:
- Interface I-003-001 - CLI process contract

### Out of Scope
- Configurable logging modes.
- Human-readable alternate output modes.
- Query-specific result schema validation.

### Assumptions
- **ASM-002-1**: Warnings during an otherwise successful command are not part of the shared runtime contract unless a future story defines a separate machine-readable warning channel. - *Why*: The capability reserves stderr for runtime failures and requires JSON-only stdout on success. - Confidence: MEDIUM

### Open Questions
- None.
