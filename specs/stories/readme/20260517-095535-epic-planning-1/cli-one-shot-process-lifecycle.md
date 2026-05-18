# User Stories: One-Shot Process Lifecycle

Status: review

## Goal
Each valid `scip-search` query invocation performs one bounded command run and exits without daemon, watch, interactive, or multi-query lifecycle behavior.

## Parent Epic
specs/epics/readme/20260517-095535-epic-planning-1.md - Capability CAP-001

## Context
`scip-search` is described as a thin Go binary that loads a pre-built index, answers one query, prints structured JSON, and exits. This document captures the lifecycle portion of that shared runtime contract without specifying SCIP traversal or result schemas.

## Personas
- **Automation agent**: A coding or planning agent running terminal commands inside concurrent worktrees, needing stable machine-readable output and deterministic failures.
- **CLI developer**: A Go developer implementing and testing `scip-search`, needing a bounded runtime contract that separates CLI behavior from query traversal.

## General information

Applies to: one-shot lifecycle behavior for documented query commands.

### References
- goal spec: README.md#what-is-scip-search - Defines `scip-search` as loading an index, answering one query, printing structured JSON to stdout, and exiting.
- goal spec: README.md#out-of-scope - Excludes daemon mode, watch mode, UI, MCP server behavior, and related adjacent capabilities.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-001---invoke-scip-search-commands-consistently - Includes deterministic one-shot process behavior before query-specific traversal.
- task: epic-planning-1-us-writing-0 - Includes one-shot process lifecycle; excludes SCIP traversal, result filtering, and result schemas.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/cli-command-routing-and-usage.md and specs/stories/readme/20260517-095535-epic-planning-1/cli-shared-index-flag.md - Written in this task to define the valid invocation boundary this lifecycle applies to.

### Non-Functional Requirements
- NFR-000-1: The runtime must execute as a one-shot CLI process, not as a daemon, watcher, REPL, server, or interactive UI.
- NFR-000-2: The runtime must not compile code, type-check code, generate indexes, update indexes, or start incremental background work as part of lifecycle handling.

### Related External Components
- Component C-001 - SCIP index file: The on-disk input selected by the caller with `--index`.
- Component C-003 - Calling process environment: The shell, script, or agent invoking `scip-search` and observing stdout, stderr, and process status.

### Interfaces
- I-003-001 - CLI process contract (Interface 001 of Component C-003): The command-line invocation, stdout/stderr streams, and process exit status exposed by `scip-search`.

### Out of Scope
- Query traversal behavior, result filtering, result schemas, JSON field names, and successful query payload contents.
- Exact stderr text, exact numeric exit code taxonomy, progress logging configuration, version output, install behavior, and release packaging.
- Daemon, watch, REPL, interactive prompt, MCP server, UI, graph visualization, semantic similarity, vector storage, and custom index format behavior.

### Assumptions
- **ASM-000-1**: "One-shot" means one process invocation handles one selected query command and terminates after that command reaches its success or failure outcome. - *Why*: README.md says the binary answers a query and exits, and explicitly excludes daemon and watch mode. - Confidence: HIGH
- **ASM-000-2**: This lifecycle story may require the process to terminate after a valid invocation without specifying the successful result schema. - *Why*: Result schemas are owned by query-specific sibling epics. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Run One Command Per Process Invocation

### References
- goal spec: README.md#what-is-scip-search - Describes a thin binary that answers a query and exits.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-001---invoke-scip-search-commands-consistently - Defines one-shot process behavior for documented commands.

### User Story
**As an** automation agent running `scip-search` from scripts, **I want to** receive one complete process result for one selected query command, **so that** each terminal invocation is deterministic and does not require session management.

### Acceptance Criteria
- AC-001-1: Given a documented query command is invoked with the required shared invocation inputs, when the runtime begins command execution, then the process handles only that selected query command for the current invocation.
- AC-001-2: Given a documented query command reaches the query execution boundary, when the command completes or fails through a shared runtime path, then the `scip-search` process exits instead of waiting for more commands.
- AC-001-2b: Given additional command names are provided after the selected command's invocation inputs, when the runtime validates the invocation shape, then those extra command names are not treated as a second query to execute in the same process.

### Depends on:
Implementation ordering:
- Story document cli-command-routing-and-usage.md - The runtime must identify a selected query command before lifecycle rules can apply to it.
- Story document cli-shared-index-flag.md - The runtime must define the shared valid invocation boundary before valid one-shot behavior can be tested.

Run time coupling:
- Interface I-003-001 - CLI process contract

### Out of Scope
- Defining successful query result content.
- Executing multiple query commands in one `scip-search` process.
- Defining a batch mode.

### Assumptions
- **ASM-001-1**: Multi-query invocation is outside this capability unless a future spec explicitly introduces batch behavior. - *Why*: README.md shows one command per invocation and lists no batch interface. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-002 - Avoid Interactive and Background Runtime Modes

### References
- goal spec: README.md#out-of-scope - Excludes daemon and watch mode, as well as UI and MCP server behavior.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-001---invoke-scip-search-commands-consistently - Limits this capability to deterministic one-shot process behavior.

### User Story
**As a** CLI developer implementing the runtime contract, **I want to** avoid interactive prompts and background lifecycle modes, **so that** the CLI remains a predictable one-shot tool for automation callers.

### Acceptance Criteria
- AC-002-1: Given a valid documented query command is invoked, when the runtime processes the command, then it does not prompt the caller for interactive input before reaching success or failure.
- AC-002-2: Given a valid documented query command is invoked, when the runtime processes the command, then it does not start a daemon, server, watch loop, or background indexing process.
- AC-002-3: Given a usage failure occurs before query traversal, when the runtime reports the failure, then the process terminates without waiting for corrective interactive input.

### Depends on:
Implementation ordering:
- Story document cli-command-routing-and-usage.md - Usage failures must be defined before their non-interactive lifecycle can be tested.

Run time coupling:
- Interface I-003-001 - CLI process contract

### Out of Scope
- Defining timeout behavior for long-running index loading or traversal.
- Installing or invoking language-specific SCIP indexers.
- Any server, daemon, watch, MCP, or UI mode.

### Assumptions
- **ASM-002-1**: The runtime does not ask the caller to supply missing values interactively. - *Why*: The product is specified as a command that receives arguments, answers one query, prints, and exits. - Confidence: HIGH

### Open Questions
- None.
