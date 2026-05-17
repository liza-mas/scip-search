# User Stories: CLI One-Shot Process Lifecycle

Status: draft

## Goal
Valid `scip-search` query invocations run as non-interactive one-shot processes that return control to the caller without daemon, watch, or traversal behavior being defined in this capability.

## Parent Epic
specs/epics/readme/20260517-095535-epic-planning-1.md - Capability CAP-001

## Context
The product is a thin Go binary that loads an index, answers one query, prints output, and exits. This story document defines the lifecycle shape shared by all query commands while leaving index loading, query traversal, and JSON result schemas to sibling capabilities.

## Personas
- **Automation agent**: A coding or planning agent running terminal commands inside concurrent worktrees, needing stable machine-readable output and deterministic failures.
- **CLI developer**: A Go developer implementing and testing `scip-search`, needing a bounded runtime contract that separates CLI behavior from query traversal.

## General information

Applies to: process lifecycle for recognized query commands with complete shared invocation shape.

### References
- goal spec: README.md#what-is-scip-search, lines 56-80 - Defines `scip-search` as a thin binary that loads a SCIP index, answers one query, prints structured JSON, and exits.
- goal spec: README.md#out-of-scope, lines 88-96 - Excludes daemon, watch mode, MCP server behavior, UI, graph visualization, semantic similarity, vector storage, and custom index formats.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#personas, lines 11-14 - Defines the automation agent and CLI developer personas.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#general-information, lines 15-57 - Defines one-shot runtime constraints and the CLI process interface.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-001---invoke-scip-search-commands-consistently, lines 59-93 - Defines one-shot process behavior before query-specific traversal runs.
- consistency check: specs/stories/readme/ - No existing story documents were present in this domain when checked.

### Non-Functional Requirements
- NFR-000-1: The runtime must perform no daemon startup, watch loop, interactive prompt, incremental update work, compilation, or type-checking as part of the shared lifecycle.
- NFR-000-2: Each invocation must be independent from previous invocations from the caller's perspective.

### Related External Components
Summary of all the external components referenced by this document:
- Component C-003 - Calling process environment: The shell, script, or agent that invokes `scip-search` and observes stdout, stderr, and exit status.

### Interfaces
Summary of all the external interfaces referenced by this document:
- I-003-001 - CLI process contract (Interface 001 of Component C-003): The command-line invocation, stdout/stderr streams, and process exit status exposed by `scip-search`.

### Out of Scope
- SCIP index path validation, index loading, and SCIP parse errors.
- SCIP document, occurrence, symbol, relationship, range, hover traversal, result filtering, and result schemas.
- Query-specific behavior for `symbols`, `packages`, `references`, and `implementations`.
- Installation, release packaging, version command behavior, and user-facing distribution docs.
- Exact successful JSON payload fields, exact diagnostic wording, and numeric exit-code taxonomy.

### Assumptions
- **ASM-000-1**: In this document, a complete shared invocation means a recognized query command with `--index <index-path>` present; query-specific flag completeness is owned by query-specific story documents. - *Why*: CAP-001 owns command routing and shared flags, while sibling epics own query behavior. - Confidence: HIGH
- **ASM-000-2**: One-shot lifecycle can be specified without defining successful query payloads. - *Why*: The epic explicitly separates shared runtime behavior from query-specific result schemas. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Complete One Invocation and Exit

### References
- goal spec: README.md#what-is-scip-search, lines 56-80 - Defines the binary as loading an index, answering one query, printing structured JSON, and exiting.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-001---invoke-scip-search-commands-consistently, lines 59-93 - Requires deterministic one-shot process behavior before query-specific traversal runs.

### User Story
**As an** automation agent, **I want to** run a documented query command as a single completed process invocation, **so that** scripts and agents can sequence `scip-search` calls without managing a daemon or interactive session.

### Acceptance Criteria
- AC-001-1: Given a recognized query command with complete shared invocation shape, when the calling process runs `scip-search`, then the invocation returns control to the caller after handling that single command.
- AC-001-2: Given a recognized query command with complete shared invocation shape, when the calling process runs `scip-search`, then the runtime does not enter an interactive prompt, daemon mode, watch mode, or long-running session.
- AC-001-3: Given two recognized query commands are invoked as separate processes, when the calling process runs them in sequence, then each invocation behaves as an independent one-shot command from the caller's perspective.

### Depends on:
Implementation ordering:
- Story document specs/stories/readme/cli-command-routing-and-usage.md - Recognized query commands must exist before one-shot behavior can be validated for them.
- Story document specs/stories/readme/cli-shared-index-flag.md - Complete shared invocation shape requires the shared `--index <index-path>` flag.

Run time coupling:
- Interface I-003-001 - CLI process contract

### Out of Scope
- Successful query result fields and filtering behavior.
- Index loading success or failure behavior.
- Exact stdout/stderr contents and numeric exit statuses.

### Assumptions
- None.

### Open Questions
- None.

---

## Story ST-002 - Avoid Ambient Long-Running Runtime Modes

### References
- goal spec: README.md#out-of-scope, lines 88-96 - Explicitly excludes daemon, watch mode, MCP server behavior, UI, graph visualization, semantic similarity, vector storage, and custom index formats.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#general-information, lines 15-57 - Defines the runtime as reading pre-built SCIP output directly and performing no daemon or watch work.

### User Story
**As a** CLI developer, **I want to** keep the shared runtime free of ambient long-running modes, **so that** query stories can rely on a small CLI lifecycle contract instead of a service lifecycle.

### Acceptance Criteria
- AC-002-1: Given any documented query command, when the runtime handles the invocation, then it does not start a daemon, server, watch loop, or interactive UI.
- AC-002-2: Given any documented query command, when the runtime handles the invocation, then it does not require stdin interaction to complete the shared lifecycle.
- AC-002-3: Given any documented query command, when the runtime handles the invocation, then it does not generate, update, or cache SCIP indexes as part of command lifecycle behavior.

### Depends on:
Implementation ordering:
- Story document specs/stories/readme/cli-command-routing-and-usage.md - The documented command surface must exist before lifecycle exclusions can be validated for each command.

Run time coupling:
- Interface I-003-001 - CLI process contract

### Out of Scope
- Performance budgets or timing thresholds.
- Index file parsing and query traversal.
- Distribution commands and install verification behavior.

### Assumptions
- None.

### Open Questions
- None.
