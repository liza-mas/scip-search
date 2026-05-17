# User Stories: SCIP Index Loading Boundary

Status: review

## Goal
`scip-search` loads a readable caller-selected SCIP index through the official SCIP boundary, rejects invalid SCIP input before query traversal, and exposes a loaded query context to query-specific stories.

## Parent Epic
specs/epics/readme/20260517-095535-epic-planning-1.md - Capability CAP-002

## Context
The goal spec describes `scip-search` as a thin Go binary that reads pre-built SCIP output directly, uses official SCIP Go bindings, answers one query, prints structured JSON, and exits. This document defines the loading boundary after the selected index path is readable and before query-specific traversal begins.

## Personas
- **Automation agent**: A coding or planning agent running terminal commands inside concurrent worktrees, needing stable machine-readable output and deterministic failures.
- **CLI developer**: A Go developer implementing and testing `scip-search`, needing a bounded runtime contract that separates CLI behavior from query traversal.

## General information

Applies to: readable caller-selected index inputs for documented query commands.

### References
- goal spec: README.md#what-is-scip-search - Defines `scip-search` as loading the provided index file, answering a query, printing structured JSON to stdout, and exiting.
- goal spec: README.md#language-support - Requires official SCIP bindings for parsing and traversal and says `scip-search` reads SCIP output directly.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-002---load-the-caller-selected-scip-index - Defines successful loading, invalid SCIP failures, and the loaded index handoff boundary.
- task: epic-planning-1-us-writing-1 - Includes the official SCIP loading boundary and excludes compiling, type-checking, generating, updating, caching, watching, custom formats, ctags fallback behavior, and query traversal.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/index-path-validation.md - Written for this capability to define path-level failures before parsing.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/cli-shared-index-flag.md and specs/stories/readme/20260517-095535-epic-planning-1/cli-one-shot-process-lifecycle.md - Define the selected path and one-shot process boundaries inherited by this document.

### Non-Functional Requirements
- NFR-000-1: The loading boundary must use official SCIP data semantics and must not introduce a custom index format.
- NFR-000-2: The runtime must read pre-built SCIP output directly and must not compile code, type-check code, generate indexes, update indexes, start a watcher, start a daemon, or cache indexes as part of this capability.
- NFR-000-3: Successful index loading must not itself define query result schemas; successful stdout remains owned by query-specific story documents.

### Related External Components
- Component C-001 - SCIP index file: The on-disk input selected by the caller with `--index`.
- Component C-002 - SCIP Go bindings: The official parsing boundary used by the runtime to read SCIP output.
- Component C-003 - Calling process environment: The shell, script, or agent invoking `scip-search` and observing stdout, stderr, and process status.

### Interfaces
- I-002-001 - SCIP loading boundary (Interface 001 of Component C-002): The runtime boundary that converts the readable caller-selected SCIP file into either loaded SCIP index data for the selected query command or a loading failure before traversal.
- I-003-001 - CLI process contract (Interface 001 of Component C-003): The command-line invocation, stdout/stderr streams, and process exit status exposed by `scip-search`.

### Out of Scope
- Filesystem existence and readability checks covered by index-path-validation.md.
- Traversing documents, occurrences, symbols, relationships, ranges, or hover data.
- Query-specific behavior for `symbols`, `packages`, `references`, or `implementations`.
- Success JSON result fields, filtering, ordering, exact stderr text, exact numeric exit code taxonomy, installing indexers, invoking indexers, generating indexes, updating indexes, caching, watching, ctags fallback behavior, and custom index formats.

### Assumptions
- **ASM-000-1**: A readable file that cannot be interpreted as SCIP data is an index-loading failure, not a query failure. - *Why*: Query-specific traversal must not begin until the runtime has loaded a SCIP index. - Confidence: HIGH
- **ASM-000-2**: A successful load only establishes a boundary object or context for the selected query command; it does not require any standalone success output. - *Why*: The source says the command prints query JSON, while sibling epics own result schemas. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Load a Readable SCIP Index for the Selected Command

### References
- goal spec: README.md#what-is-scip-search - Requires loading the index file at the path provided as an argument before answering a query.
- goal spec: README.md#language-support - Requires official SCIP bindings and direct SCIP output reads.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-002---load-the-caller-selected-scip-index - Covers handing a loaded SCIP index to query-specific behavior.

### User Story
**As a** CLI developer implementing query commands, **I want to** receive a loaded SCIP index boundary after the caller-selected file is accepted, **so that** query-specific stories can traverse SCIP data without owning file parsing.

### Acceptance Criteria
- AC-001-1: Given a documented query command is invoked with `--index <index-path>` and `<index-path>` identifies a readable valid SCIP index file, when the loading boundary runs, then the selected command reaches its query execution boundary with loaded SCIP index data available.
- AC-001-2: Given a readable valid SCIP index file is loaded, when the selected command reaches its query execution boundary, then the runtime preserves the caller-selected index as the only index input for that invocation.
- AC-001-3: Given a readable valid SCIP index file is loaded, when loading completes, then the runtime does not compile source code, type-check source code, generate an index, update an index, start a watcher, start a daemon, or require a custom index format.
- AC-001-4: Given loading succeeds, when the selected query command later produces its own result, then that query-specific result behavior is governed by the selected query story rather than by this loading story.

### Depends on:
Implementation ordering:
- Story document cli-shared-index-flag.md - The runtime must first establish the selected `--index` path.
- Story document index-path-validation.md - The runtime must first reject missing or unreadable selected paths before this readable-file loading boundary can be tested.

Run time coupling:
- Interface I-002-001 - SCIP loading boundary
- Interface I-003-001 - CLI process contract

### Out of Scope
- Defining query result fields, query traversal, or query-specific filtering.
- Supporting multiple indexes in one invocation.
- Persisting the loaded index beyond the current process invocation.

### Assumptions
- **ASM-001-1**: A "loaded query context" means the selected command has access to loaded SCIP index data at its execution boundary; the exact internal type is implementation-owned. - *Why*: The story must expose behavior without prescribing Go implementation details. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-002 - Reject Readable but Invalid SCIP Input

### References
- goal spec: README.md#language-support - Requires official SCIP bindings and direct SCIP output reads.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-002---load-the-caller-selected-scip-index - Covers invalid SCIP inputs and clear loading failures.

### User Story
**As an** automation agent running `scip-search` from scripts, **I want to** receive a clear failure when the selected readable file is not valid SCIP input, **so that** downstream automation does not mistake invalid index data for an empty query result.

### Acceptance Criteria
- AC-002-1: Given a documented query command is invoked with `--index <index-path>` and `<index-path>` identifies a readable file that is not valid SCIP input, when the loading boundary runs, then the process reports an index-loading failure before query-specific traversal starts.
- AC-002-2: Given a readable invalid SCIP input file is selected, when the failure is reported, then the selected command does not produce a successful query result or an empty-result success solely because parsing failed.
- AC-002-3: Given a readable invalid SCIP input file is selected, when the loading boundary rejects it, then the runtime does not attempt custom-format parsing or ctags fallback behavior.
- AC-002-4: Given invalid SCIP input is rejected, when the process exits, then the input file remains caller-owned and unmodified.

### Depends on:
Implementation ordering:
- Story document cli-shared-index-flag.md - The runtime must first establish the selected `--index` path.
- Story document index-path-validation.md - The runtime must first distinguish readable files from missing or unreadable selected paths.

Run time coupling:
- Interface I-002-001 - SCIP loading boundary
- Interface I-003-001 - CLI process contract

### Out of Scope
- Enumerating every parse error shape from the official SCIP bindings.
- Exact diagnostic wording or exact numeric exit code.
- Recovering partial results from malformed SCIP input.

### Assumptions
- **ASM-002-1**: Invalid SCIP input must be reported as a failure instead of as a successful query with zero matches. - *Why*: The capability says loading failure must occur before query execution. - Confidence: HIGH

### Open Questions
- None.
