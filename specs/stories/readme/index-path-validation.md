# User Stories: Index Path Validation Failures

Status: review

## Goal
`scip-search` rejects caller-selected index paths that cannot be opened as SCIP input before query-specific traversal starts.

## Parent Epic
specs/epics/readme/20260517-095535-epic-planning-1.md - Capability CAP-002

## Context
The runtime receives an explicit `--index` value from the shared CLI invocation contract. This document covers filesystem-level index input failures after the flag value has been accepted and before the official SCIP parse boundary owns the file contents.

## Personas
- **Automation agent**: A coding or planning agent running terminal commands inside concurrent worktrees, needing stable machine-readable output and deterministic failures.
- **CLI developer**: A Go developer implementing and testing `scip-search`, needing a bounded runtime contract that separates CLI behavior from query traversal.

## General information

Applies to: caller-provided index path validation before SCIP parsing.

### References
- goal spec: README.md#what-is-scip-search - Requires `scip-search` to load the SCIP index file at the path provided as an argument before answering a query.
- goal spec: README.md#out-of-scope - Excludes daemon/watch mode, incremental updates, and custom index formats.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-002---load-the-caller-selected-scip-index - Defines caller-selected index loading and clear loading failures before query execution.
- dependency story: specs/stories/readme/20260517-095535-epic-planning-1/cli-shared-index-flag.md - Defines when a syntactically supplied `--index` value reaches the index-loading boundary.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/cli-command-routing-and-usage.md and specs/stories/readme/20260517-095535-epic-planning-1/cli-one-shot-process-lifecycle.md - Confirmed command routing and lifecycle behavior remain owned by CAP-001.

### Non-Functional Requirements
- NFR-000-1: Index path validation must not compile code, type-check code, generate indexes, update indexes, start watchers, or search default index locations.
- NFR-000-2: Index path failures must occur before query-specific traversal begins.

### Related External Components
- Component C-001 - SCIP index file: The on-disk input selected by the caller with `--index`.
- Component C-003 - Calling process environment: The shell, script, or agent invoking `scip-search` and observing stdout, stderr, and process status.

### Interfaces
- I-003-001 - CLI process contract (Interface 001 of Component C-003): The command-line invocation, stdout/stderr streams, and process exit status exposed by `scip-search`.

### Out of Scope
- Missing `--index` flag behavior, missing flag values, unsupported commands, and one-shot lifecycle rules owned by CAP-001 story documents.
- Parsing readable files as SCIP data, invalid SCIP payload handling, traversal-ready SCIP views, and query result schemas.
- Exact stderr wording, exact numeric exit code taxonomy, and JSON error schema owned by the shared runtime error contract.
- Generating indexes, installing language indexers, caching, watching, incremental updates, default index discovery, custom formats, and ctags fallback behavior.

### Assumptions
- **ASM-000-1**: Filesystem-level failures include a path that does not exist, a path that is not a regular file, and a path the process cannot open for reading. - *Why*: CAP-002 names missing and unreadable SCIP inputs, while directories are caller-provided paths that cannot be consumed as a SCIP index file. - Confidence: HIGH
- **ASM-000-2**: These failures are index-loading failures rather than usage failures once `--index` has a syntactic value. - *Why*: CAP-001 hands supplied `--index` values to the index-loading boundary and leaves file outcomes to CAP-002. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Reject Missing Index Files

### References
- goal spec: README.md#what-is-scip-search - Requires loading the caller-provided index file path before answering a query.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-002---load-the-caller-selected-scip-index - Covers distinguishing missing index input from successful loading.
- dependency story: specs/stories/readme/20260517-095535-epic-planning-1/cli-shared-index-flag.md - Defines that supplied index path values reach this loading boundary.

### User Story
**As an** automation agent running `scip-search` in concurrent worktrees, **I want to** be told when my selected index path does not exist, **so that** I can fix the index location before relying on query output.

### Acceptance Criteria
- AC-001-1: Given a documented query command is invoked with `--index <index-path>` and `<index-path>` does not identify any filesystem entry visible to the process, when index path validation runs, then the process reports an index loading failure for the selected path before query traversal starts.
- AC-001-1b: Given the selected path is relative, when it does not identify any filesystem entry from the invocation environment, then the process reports the same missing index loading failure before query traversal starts.

### Depends on:
Implementation ordering:
- Story document cli-shared-index-flag.md - The runtime must accept a syntactic `--index` value before filesystem validation can apply to it.

Run time coupling:
- Interface I-003-001 - CLI process contract

### Out of Scope
- Defining the exact diagnostic text or numeric status for the failure.
- Searching fallback locations or generating an index when the selected path is missing.
- Validating query-specific flags after the missing index path is detected.

### Assumptions
- **ASM-001-1**: A missing path failure should identify the caller-selected path at the user-observable diagnostic boundary. - *Why*: CAP-002 requires a clear loading failure, and the selected path is the actionable input the caller can correct. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-002 - Reject Unreadable Index Inputs

### References
- goal spec: README.md#what-is-scip-search - Requires loading the caller-provided index file before answering a query.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-002---load-the-caller-selected-scip-index - Covers unreadable index input failures before query execution.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#general-information - Excludes custom index formats and index generation.

### User Story
**As a** CLI developer implementing the runtime boundary, **I want to** reject selected index paths that cannot be opened as readable files, **so that** query commands never run against absent or unusable index input.

### Acceptance Criteria
- AC-002-1: Given a documented query command is invoked with `--index <index-path>` and `<index-path>` identifies a directory, when index path validation runs, then the process reports an index loading failure before query traversal starts.
- AC-002-2: Given a documented query command is invoked with `--index <index-path>` and the process cannot open `<index-path>` for reading, when index path validation runs, then the process reports an index loading failure before query traversal starts.
- AC-002-2b: Given the selected path is present but becomes unreadable before loading completes, when the runtime attempts to open it, then the process reports an index loading failure rather than continuing to query execution.

### Depends on:
Implementation ordering:
- Story document cli-shared-index-flag.md - The runtime must accept a syntactic `--index` value before filesystem validation can apply to it.

Run time coupling:
- Interface I-003-001 - CLI process contract

### Out of Scope
- Parsing file contents or classifying invalid SCIP protobuf data.
- Defining exact diagnostic text, numeric exit code taxonomy, or JSON error schema.
- Permission repair, index generation, index discovery, or retry loops.

### Assumptions
- **ASM-002-1**: A directory selected through `--index` is an unusable index input even if the process can inspect the directory. - *Why*: README.md describes loading a SCIP index file, not a directory of indexes. - Confidence: HIGH

### Open Questions
- None.
