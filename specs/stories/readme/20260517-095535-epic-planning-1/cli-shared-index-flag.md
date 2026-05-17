# User Stories: Shared Index Flag Contract

Status: review

## Goal
Every documented `scip-search` query command requires the shared `--index` flag and handles missing or unusable flag shape before query-specific traversal begins.

## Parent Epic
specs/epics/readme/20260517-095535-epic-planning-1.md - Capability CAP-001

## Context
The goal spec shows each query command receiving an explicit SCIP index path through `--index`. This document covers the invocation-level flag contract only: the flag is required and its supplied value is passed to the shared index-loading boundary. Actual file loading outcomes belong to the sibling index-loading story set.

## Personas
- **Automation agent**: A coding or planning agent running terminal commands inside concurrent worktrees, needing stable machine-readable output and deterministic failures.
- **CLI developer**: A Go developer implementing and testing `scip-search`, needing a bounded runtime contract that separates CLI behavior from query traversal.

## General information

Applies to: the shared `--index` flag for all documented query commands.

### References
- goal spec: README.md#what-is-scip-search - Shows `--index <index-path>` in the documented command forms for `symbols`, `references`, `implementations`, and `packages`.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-001---invoke-scip-search-commands-consistently - Includes shared flag behavior in the command invocation contract.
- task: epic-planning-1-us-writing-0 - Includes shared `--index` flag behavior and usage failures; excludes index loading and query traversal.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/cli-command-routing-and-usage.md - Written in this task to define the command names that share this flag.

### Non-Functional Requirements
- NFR-000-1: The CLI runtime must support concurrent worktrees by requiring a caller-selected index path rather than assuming a global or repository-local index.
- NFR-000-2: The shared `--index` contract must be consistent across `symbols`, `references`, `implementations`, and `packages`.

### Related External Components
- Component C-001 - SCIP index file: The on-disk input selected by the caller with `--index`.
- Component C-003 - Calling process environment: The shell, script, or agent invoking `scip-search` and observing stdout, stderr, and process status.

### Interfaces
- I-003-001 - CLI process contract (Interface 001 of Component C-003): The command-line invocation, stdout/stderr streams, and process exit status exposed by `scip-search`.

### Out of Scope
- Whether the supplied index path exists, is readable, or contains valid SCIP data.
- Loading the SCIP file, parsing with SCIP Go bindings, caching, watching, generating, or updating indexes.
- Query-specific flags, query traversal, result schemas, JSON success payload fields, version output, install behavior, and release packaging.

### Assumptions
- **ASM-000-1**: The only shared required flag in this capability is `--index`. - *Why*: README.md shows `--index` on every query command, while all other shown flags are query-specific. - Confidence: HIGH
- **ASM-000-2**: A syntactically supplied `--index` value is enough for this capability to hand off to the index-loading boundary. - *Why*: CAP-002 owns missing, unreadable, and invalid SCIP input outcomes after invocation shape is accepted. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Require Index Flag on Every Query Command

### References
- goal spec: README.md#what-is-scip-search - Shows every documented query command using `--index <index-path>`.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-001---invoke-scip-search-commands-consistently - Defines shared flag behavior for documented query commands.

### User Story
**As a** automation agent running commands against multiple worktrees, **I want to** provide the SCIP index path with the same `--index` flag on every query command, **so that** each invocation targets the intended pre-built index explicitly.

### Acceptance Criteria
- AC-001-1: Given `symbols` is invoked without `--index`, when the runtime validates shared invocation inputs, then the process reports a usage failure before query traversal starts.
- AC-001-2: Given `references` is invoked without `--index`, when the runtime validates shared invocation inputs, then the process reports a usage failure before query traversal starts.
- AC-001-3: Given `implementations` is invoked without `--index`, when the runtime validates shared invocation inputs, then the process reports a usage failure before query traversal starts.
- AC-001-4: Given `packages` is invoked without `--index`, when the runtime validates shared invocation inputs, then the process reports a usage failure before query traversal starts.

### Depends on:
Implementation ordering:
- Story document cli-command-routing-and-usage.md - The documented query command names must exist before the shared flag can be applied consistently across them.

Run time coupling:
- Interface I-003-001 - CLI process contract

### Out of Scope
- Validating whether the provided path points to a real or valid SCIP file.
- Defining query-specific flag requirements for each command.

### Assumptions
- **ASM-001-1**: Missing `--index` is a runtime usage failure rather than an index-loading failure. - *Why*: No index path has been selected, so the runtime cannot reach the file-loading boundary. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-002 - Accept an Explicit Index Path Value

### References
- goal spec: README.md#what-is-scip-search - Describes loading a SCIP index file at the path provided by argument and shows `--index <index-path>`.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-001---invoke-scip-search-commands-consistently - Bounds shared invocation behavior before query-specific traversal.

### User Story
**As a** CLI developer implementing the runtime contract, **I want to** treat a supplied `--index` value as the caller-selected index path for the current command, **so that** the later index-loading boundary receives the same path the caller provided.

### Acceptance Criteria
- AC-002-1: Given any documented query command is invoked with `--index <index-path>`, when shared invocation validation completes, then the supplied `<index-path>` is accepted as the selected index path for that one command invocation.
- AC-002-1b: Given `--index` is present but no value is provided, when shared invocation validation runs, then the process reports a usage failure before index loading or query traversal starts.
- AC-002-2: Given one documented query command is invoked with `--index <index-path>`, when the process exits, then the selected index path does not persist into any later `scip-search` invocation.

### Depends on:
Implementation ordering:
- Story document cli-command-routing-and-usage.md - The documented query command names must exist before the shared flag can be applied consistently across them.

Run time coupling:
- Interface I-003-001 - CLI process contract

### Out of Scope
- Path normalization, existence checks, file permissions, file parsing, or invalid SCIP data handling.
- Supporting default index locations.
- Supporting multiple index paths in one invocation.

### Assumptions
- **ASM-002-1**: The runtime does not define a default index path when `--index` is omitted. - *Why*: README.md shows an explicit index path on every command and the epic emphasizes concurrent worktree support. - Confidence: HIGH

### Open Questions
- None.
