# User Stories: CLI Command Routing and Usage Failures

Status: draft

## Goal
Every documented `scip-search` query command is recognized consistently, and missing or unknown commands fail as usage errors before query-specific work begins.

## Parent Epic
specs/epics/readme/20260517-095535-epic-planning-1.md - Capability CAP-001

## Context
`scip-search` is a one-shot CLI with four documented query commands. This story document defines the command surface and usage-failure boundary only, so query traversal and result contracts can be specified by sibling story documents.

## Personas
- **Automation agent**: A coding or planning agent running terminal commands inside concurrent worktrees, needing stable machine-readable output and deterministic failures.
- **CLI developer**: A Go developer implementing and testing `scip-search`, needing a bounded runtime contract that separates CLI behavior from query traversal.

## General information

Applies to: command names and invocation-shape failures for `symbols`, `references`, `implementations`, and `packages`.

### References
- goal spec: README.md#what-is-scip-search, lines 56-80 - Defines `scip-search` as a thin binary and lists the four query command forms.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#personas, lines 11-14 - Defines the automation agent and CLI developer personas.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#general-information, lines 15-57 - Defines epic-wide runtime constraints, components, interfaces, assumptions, and out-of-scope boundaries.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-001---invoke-scip-search-commands-consistently, lines 59-93 - Defines the command routing, shared flag, usage failure, and one-shot lifecycle capability.
- consistency check: specs/stories/readme/ - No existing story documents were present in this domain when checked.

### Non-Functional Requirements
- NFR-000-1: Command routing must not perform compilation, type-checking, daemon startup, watch loops, incremental updates, or query traversal before the selected query command is identified.
- NFR-000-2: Usage failures must be deterministic for automation: the same invalid invocation shape produces the same failure category each time.

### Related External Components
Summary of all the external components referenced by this document:
- Component C-003 - Calling process environment: The shell, script, or agent that invokes `scip-search` and observes stdout, stderr, and exit status.

### Interfaces
Summary of all the external interfaces referenced by this document:
- I-003-001 - CLI process contract (Interface 001 of Component C-003): The command-line invocation, stdout/stderr streams, and process exit status exposed by `scip-search`.

### Out of Scope
- Shared `--index` flag validation beyond recognizing that valid query invocations use the shared index flag.
- Query-specific flags such as `--name`, `--symbol`, and `--prefix`.
- SCIP index loading, SCIP traversal, result filtering, and result schemas.
- Installation, release packaging, version command behavior, daemon/watch mode, MCP server behavior, UI, graph visualization, semantic similarity, vector storage, and custom index formats.
- Exact stderr wording, stdout JSON schema, and numeric exit-code taxonomy beyond routing failures going through the shared runtime failure path.

### Assumptions
- **ASM-000-1**: The command list for this scope is limited to `symbols`, `references`, `implementations`, and `packages`. - *Why*: The goal spec and CAP-001 list only those query commands, while version/install behavior is assigned to distribution scope. - Confidence: HIGH
- **ASM-000-2**: A usage failure means the invocation is rejected before index loading or query traversal is attempted. - *Why*: CAP-001 scopes invalid invocation shape separately from index loading and query behavior. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Recognize Documented Query Commands

### References
- goal spec: README.md#what-is-scip-search, lines 58-72 - Lists the documented command forms for `symbols`, `references`, `implementations`, and `packages`.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-001---invoke-scip-search-commands-consistently, lines 59-93 - Requires the documented query commands to share a command surface.

### User Story
**As a** CLI developer, **I want to** route only the documented query command names, **so that** each query capability can be implemented behind a stable CLI entrypoint without redefining command discovery.

### Acceptance Criteria
- AC-001-1: Given an invocation beginning with `scip-search symbols`, when the runtime parses the command name, then `symbols` is accepted as a recognized query command.
- AC-001-2: Given an invocation beginning with `scip-search references`, when the runtime parses the command name, then `references` is accepted as a recognized query command.
- AC-001-3: Given an invocation beginning with `scip-search implementations`, when the runtime parses the command name, then `implementations` is accepted as a recognized query command.
- AC-001-4: Given an invocation beginning with `scip-search packages`, when the runtime parses the command name, then `packages` is accepted as a recognized query command.
- AC-001-5: Given a recognized query command and required shared runtime flags, when the invocation is parsed, then the caller does not receive a missing-command or unknown-command usage failure.

### Depends on:
Run time coupling:
- Interface I-003-001 - CLI process contract

### Out of Scope
- Query-specific flag validation and query-specific result contents.
- Loading or traversing a SCIP index.
- Defining output for successful query execution.

### Assumptions
- None.

### Open Questions
- None.

---

## Story ST-002 - Reject Missing or Unknown Commands as Usage Failures

### References
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-001---invoke-scip-search-commands-consistently, lines 59-93 - Requires invalid invocation shape to be rejected before query traversal.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#general-information, lines 15-57 - Defines the CLI process contract and shared runtime failure boundary.

### User Story
**As an** automation agent, **I want to** receive a deterministic usage failure when I omit or mistype the query command, **so that** scripts can distinguish invocation mistakes from query results.

### Acceptance Criteria
- AC-002-1: Given `scip-search` is invoked without a query command, when the runtime parses the invocation, then it fails through the shared usage-failure path.
- AC-002-2: Given `scip-search` is invoked with a command name other than `symbols`, `references`, `implementations`, or `packages`, when the runtime parses the invocation, then it fails through the shared usage-failure path.
- AC-002-3: Given an invocation fails because the command is missing or unknown, when the runtime handles the failure, then the caller receives a usage failure rather than an index-loading failure or query result.
- AC-002-4: Given an invocation fails because the command is missing or unknown, when the runtime handles the failure, then control returns to the calling process as a completed one-shot invocation.

### Depends on:
Run time coupling:
- Interface I-003-001 - CLI process contract

### Out of Scope
- Exact diagnostic wording, stdout/stderr stream contract, and numeric exit status details.
- Usage failures caused by missing `--index` values or query-specific flags.

### Assumptions
- None.

### Open Questions
- None.
