# User Stories: Command Routing and Usage Failures

Status: review

## Goal
`scip-search` recognizes the documented query command names and rejects missing or unsupported invocation shapes before query traversal begins.

## Parent Epic
specs/epics/readme/20260517-095535-epic-planning-1.md - Capability CAP-001

## Context
This document covers the command surface for the shared CLI runtime. It gives Coder agents a bounded contract for root invocation, documented query subcommands, and usage failures while leaving query-specific arguments and traversal behavior to sibling story sets.

## Personas
- **Automation agent**: A coding or planning agent running terminal commands inside concurrent worktrees, needing stable machine-readable output and deterministic failures.
- **CLI developer**: A Go developer implementing and testing `scip-search`, needing a bounded runtime contract that separates CLI behavior from query traversal.

## General information

Applies to: command routing and shared usage failures for all stories in this document.

### References
- goal spec: README.md#what-is-scip-search - Lists the one-shot binary behavior and the documented `symbols`, `references`, `implementations`, and `packages` command forms.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-001---invoke-scip-search-commands-consistently - Defines command routing, invalid invocation handling, and shared process behavior before traversal.
- task: epic-planning-1-us-writing-0 - Limits this story work to command routing, shared flags, usage failures, and one-shot lifecycle.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/ - Directory did not exist before this task, so no adjacent story documents were available to compare.

### Non-Functional Requirements
- NFR-000-1: The CLI runtime must not perform daemon startup, watch mode, incremental update work, compilation, or type-checking while routing commands.
- NFR-000-2: Usage failures must occur before query-specific SCIP traversal begins.

### Related External Components
- Component C-003 - Calling process environment: The shell, script, or agent invoking `scip-search` and observing stdout, stderr, and process status.

### Interfaces
- I-003-001 - CLI process contract (Interface 001 of Component C-003): The command-line invocation, stdout/stderr streams, and process exit status exposed by `scip-search`.

### Out of Scope
- Query-specific flags such as `--name`, `--symbol`, and `--prefix`, except where they appear as opaque arguments owned by sibling query stories.
- SCIP index loading success or failure behavior.
- SCIP traversal, result filtering, result schemas, JSON success payload fields, version output, install behavior, release packaging, daemon mode, watch mode, MCP behavior, UI behavior, graph visualization, semantic similarity, vector storage, and custom index formats.

### Assumptions
- **ASM-000-1**: The documented query commands for this capability are exactly `symbols`, `references`, `implementations`, and `packages`. - *Why*: README.md lists these four command forms, and the epic assigns version/install behavior elsewhere. - Confidence: HIGH
- **ASM-000-2**: A command can be considered routed successfully when it reaches the shared runtime boundary for the selected query command without being rejected as a missing or unknown command. - *Why*: Query traversal and result schemas are explicitly out of scope for this capability. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Route Documented Query Commands

### References
- goal spec: README.md#what-is-scip-search - Lists the four documented `scip-search` query command forms.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-001---invoke-scip-search-commands-consistently - Defines the shared command surface.

### User Story
**As a** automation agent running terminal commands in a worktree, **I want to** invoke any documented `scip-search` query command by name, **so that** I can rely on a stable command surface before query-specific behavior is implemented.

### Acceptance Criteria
- AC-001-1: Given the `symbols` command is invoked with the shared required runtime inputs, when command routing runs, then the invocation is accepted as the `symbols` query command rather than rejected as a missing or unknown command.
- AC-001-2: Given the `references` command is invoked with the shared required runtime inputs, when command routing runs, then the invocation is accepted as the `references` query command rather than rejected as a missing or unknown command.
- AC-001-3: Given the `implementations` command is invoked with the shared required runtime inputs, when command routing runs, then the invocation is accepted as the `implementations` query command rather than rejected as a missing or unknown command.
- AC-001-4: Given the `packages` command is invoked with the shared required runtime inputs, when command routing runs, then the invocation is accepted as the `packages` query command rather than rejected as a missing or unknown command.

### Depends on:
Run time coupling:
- Interface I-003-001 - CLI process contract

### Out of Scope
- Defining the required query-specific inputs for each command.
- Returning query results or placeholder result schemas.
- Loading, parsing, or traversing the SCIP index.

### Assumptions
- **ASM-001-1**: "Shared required runtime inputs" means runtime-level inputs common to documented query commands, not query-specific filters. - *Why*: The task scope includes shared flags and excludes query-specific flags. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-002 - Reject Missing or Unsupported Commands

### References
- goal spec: README.md#what-is-scip-search - Documents command forms that require a query command after `scip-search`.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-001---invoke-scip-search-commands-consistently - Includes invalid invocation shape handling.

### User Story
**As a** CLI developer implementing `scip-search`, **I want to** reject invocations that omit the query command or use an unsupported command name, **so that** invalid command shapes fail through the shared usage path before any query work starts.

### Acceptance Criteria
- AC-002-1: Given `scip-search` is invoked without a query command, when the runtime validates the invocation shape, then the process reports a usage failure and exits without starting query traversal.
- AC-002-2: Given `scip-search` is invoked with a command name other than `symbols`, `references`, `implementations`, or `packages`, when the runtime validates the invocation shape, then the process reports a usage failure and exits without starting query traversal.
- AC-002-2b: Given an unsupported command name is provided together with otherwise plausible flags, when the runtime validates the invocation shape, then the unsupported command is still reported as a usage failure instead of being treated as a query-specific argument.

### Depends on:
Run time coupling:
- Interface I-003-001 - CLI process contract

### Out of Scope
- Exact stderr text, exact numeric exit code taxonomy, or JSON error schema.
- Validating query-specific flag presence or semantics.
- Suggesting similar command names.

### Assumptions
- **ASM-002-1**: Unsupported command names should fail as usage failures even when they appear with known shared flags. - *Why*: The command name is the first runtime routing boundary exposed to callers. - Confidence: HIGH

### Open Questions
- None.
