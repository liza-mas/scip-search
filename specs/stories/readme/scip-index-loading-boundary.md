# User Stories: SCIP Index Loading Boundary

Status: review

## Goal
`scip-search` loads a readable caller-selected SCIP index through the official SCIP Go boundary, or reports invalid SCIP input before query-specific traversal starts.

## Parent Epic
specs/epics/readme/20260517-095535-epic-planning-1.md - Capability CAP-002

## Context
After shared command and `--index` validation succeeds, the runtime must read the caller-selected SCIP output directly and expose a loaded index boundary to query-specific stories. This document covers the official SCIP loading boundary only; it does not define traversal views, query filters, or result schemas.

## Personas
- **Automation agent**: A coding or planning agent running terminal commands inside concurrent worktrees, needing stable machine-readable output and deterministic failures.
- **CLI developer**: A Go developer implementing and testing `scip-search`, needing a bounded runtime contract that separates CLI behavior from query traversal.

## General information

Applies to: parsing readable caller-selected SCIP files and exposing the loaded boundary to query execution.

### References
- goal spec: README.md#what-is-scip-search - Requires loading the index file at the provided path, answering one query, printing structured JSON, and exiting.
- goal spec: README.md#language-support - Requires use of the official SCIP bindings for parsing and traversal and direct reads of SCIP output.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-002---load-the-caller-selected-scip-index - Defines successful loading and clear loading failures before query execution.
- dependency story: specs/stories/readme/20260517-095535-epic-planning-1/cli-shared-index-flag.md - Defines how the selected index path is supplied to the loading boundary.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/cli-one-shot-process-lifecycle.md - Confirms the loading boundary participates in one bounded process invocation.

### Non-Functional Requirements
- NFR-000-1: The runtime must use the official SCIP Go bindings as the SCIP parse boundary instead of implementing a custom SCIP file format.
- NFR-000-2: The runtime must read pre-built SCIP output directly and must not compile, type-check, generate, update, cache, or watch indexes while loading.
- NFR-000-3: Loading success must occur before query-specific traversal begins.

### Related External Components
- Component C-001 - SCIP index file: The on-disk input selected by the caller with `--index`.
- Component C-002 - SCIP Go bindings: The official parsing boundary used by the runtime to read SCIP output.
- Component C-003 - Calling process environment: The shell, script, or agent invoking `scip-search` and observing stdout, stderr, and process status.

### Interfaces
- I-002-001 - SCIP loading boundary (Interface 001 of Component C-002): The runtime boundary that accepts readable SCIP input and returns either loaded SCIP index data for query execution or a loading failure.
- I-003-001 - CLI process contract (Interface 001 of Component C-003): The command-line invocation, stdout/stderr streams, and process exit status exposed by `scip-search`.

### Out of Scope
- Filesystem path existence and readability failures covered by specs/stories/readme/index-path-validation.md.
- Traversing documents, occurrences, symbols, relationships, ranges, or hover data from the loaded index.
- Query-specific behavior for `symbols`, `packages`, `references`, or `implementations`.
- Exact stderr wording, exact numeric exit code taxonomy, and JSON error schema owned by the shared runtime error contract.
- Generating indexes, invoking language indexers, watching, caching, incremental updates, custom formats, and ctags fallback behavior.

### Assumptions
- **ASM-000-1**: A readable file that the official SCIP Go boundary cannot load is treated as invalid SCIP input. - *Why*: CAP-002 requires invalid SCIP inputs to fail clearly, and README.md names SCIP as the only in-scope format. - Confidence: HIGH
- **ASM-000-2**: The loaded boundary is a runtime contract for query stories, not a user-facing result payload by itself. - *Why*: CAP-002 says successful loading exposes a boundary to query stories, while sibling epics own traversal and result schemas. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Load a Valid SCIP Index Before Query Execution

### References
- goal spec: README.md#what-is-scip-search - Requires loading the index at the provided path before answering a query.
- goal spec: README.md#language-support - Requires official SCIP bindings and direct SCIP output reads.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-002---load-the-caller-selected-scip-index - Covers handing a loaded SCIP index to query-specific behavior.

### User Story
**As a** CLI developer implementing query commands, **I want to** receive a loaded SCIP index boundary after the selected index file is valid, **so that** query-specific code can run without owning file loading or SCIP parsing.

### Acceptance Criteria
- AC-001-1: Given a documented query command is invoked with `--index <index-path>` and `<index-path>` identifies a readable valid SCIP index file, when the runtime reaches the SCIP loading boundary, then the index is loaded before query-specific traversal begins.
- AC-001-2: Given the selected SCIP index loads successfully, when the runtime hands control to the selected query command, then the query command receives access to the loaded SCIP index boundary for that invocation.
- AC-001-3: Given one invocation loads a selected SCIP index successfully, when the process exits, then that loaded index boundary is not reused as implicit state for a later invocation.

### Depends on:
Implementation ordering:
- Story document cli-shared-index-flag.md - The runtime must define the caller-selected index path before loading can use it.
- Story document index-path-validation.md - The runtime must reject missing and unreadable path inputs before successful SCIP parsing can be tested reliably.

Run time coupling:
- Interface I-002-001 - SCIP loading boundary
- Interface I-003-001 - CLI process contract

### Out of Scope
- Defining the shape of traversal views or successful JSON query payloads.
- Loading more than one index in a single invocation.
- Caching loaded indexes across invocations.

### Assumptions
- **ASM-001-1**: A successfully loaded SCIP index is scoped to the current one-shot process invocation. - *Why*: CAP-001 defines one-shot process behavior, and CAP-002 excludes caching. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-002 - Reject Invalid SCIP Input at the Official Parse Boundary

### References
- goal spec: README.md#what-is-scip-search - Defines SCIP as the pre-built binary index input used before answering a query.
- goal spec: README.md#language-support - Requires the official SCIP bindings for parsing and direct SCIP output reads.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-002---load-the-caller-selected-scip-index - Covers invalid SCIP loading failures before query execution.

### User Story
**As an** automation agent invoking `scip-search`, **I want to** receive a clear failure when my readable index file is not valid SCIP input, **so that** I do not mistake a loading problem for an empty or successful query result.

### Acceptance Criteria
- AC-002-1: Given a documented query command is invoked with `--index <index-path>` and `<index-path>` identifies a readable file that the official SCIP loading boundary rejects as invalid SCIP input, when loading runs, then the process reports an index loading failure before query traversal starts.
- AC-002-1b: Given invalid SCIP input is detected after some bytes have been read, when the official SCIP loading boundary rejects the input, then the runtime still reports a loading failure instead of handing partial data to query execution.
- AC-002-2: Given the selected file is valid data in a custom or fallback format other than SCIP, when loading runs for the in-scope runtime, then the process rejects it as invalid SCIP input before query traversal starts.

### Depends on:
Implementation ordering:
- Story document cli-shared-index-flag.md - The runtime must define the caller-selected index path before loading can use it.
- Story document index-path-validation.md - The runtime must reach the parse boundary only after filesystem-level path validation succeeds.

Run time coupling:
- Interface I-002-001 - SCIP loading boundary
- Interface I-003-001 - CLI process contract

### Out of Scope
- Validating ctags fallback wrappers or any custom index format.
- Recovering partial SCIP data after parse failure.
- Defining exact diagnostic text, numeric exit code taxonomy, or JSON error schema.

### Assumptions
- **ASM-002-1**: Invalid SCIP input is reported as an index loading failure rather than as a successful empty result. - *Why*: The runtime has not reached a loaded query context, so query execution has no valid input. - Confidence: HIGH

### Open Questions
- None.
