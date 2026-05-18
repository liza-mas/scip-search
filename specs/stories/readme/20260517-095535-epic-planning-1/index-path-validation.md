# User Stories: Index Path Validation Failures

Status: review

## Goal
`scip-search` rejects caller-selected index paths that do not identify a readable SCIP input file before query-specific traversal begins.

## Parent Epic
specs/epics/readme/20260517-095535-epic-planning-1.md - Capability CAP-002

## Context
The goal spec requires every query command to load the SCIP index file selected with `--index`. This document covers failures that occur after the shared `--index` flag has supplied a path but before the runtime can hand input to the official SCIP loading boundary.

## Personas
- **Automation agent**: A coding or planning agent running terminal commands inside concurrent worktrees, needing stable machine-readable output and deterministic failures.
- **CLI developer**: A Go developer implementing and testing `scip-search`, needing a bounded runtime contract that separates CLI behavior from query traversal.

## General information

Applies to: caller-provided index paths for documented query commands after shared invocation validation succeeds.

### References
- goal spec: README.md#what-is-scip-search - Requires `scip-search` to load the index file at the path provided as an argument before answering a query.
- goal spec: README.md#out-of-scope - Excludes daemon/watch mode, custom index formats, index generation, UI, MCP, and related adjacent capabilities.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-002---load-the-caller-selected-scip-index - Defines caller-provided index path validation and clear loading failures.
- task: epic-planning-1-us-writing-1 - Includes caller-provided index path validation; excludes index generation, watching, caching, incremental updates, custom formats, ctags fallback behavior, and query traversal.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/cli-shared-index-flag.md - Defines the prior invocation boundary where `--index <index-path>` is accepted as the selected path.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/cli-command-routing-and-usage.md and specs/stories/readme/20260517-095535-epic-planning-1/cli-one-shot-process-lifecycle.md - Define command routing and one-shot failure behavior this document must preserve.

### Non-Functional Requirements
- NFR-000-1: Index path validation must not compile code, type-check code, generate indexes, update indexes, start a watcher, start a daemon, or search for default index locations.
- NFR-000-2: Index path validation failures must be reported before query-specific SCIP traversal starts.
- NFR-000-3: Validation must preserve automation-friendly process behavior: diagnostics use the shared failure channel and successful stdout remains reserved for query JSON owned by sibling stories.

### Related External Components
- Component C-001 - SCIP index file: The on-disk input selected by the caller with `--index`.
- Component C-003 - Calling process environment: The shell, script, or agent invoking `scip-search` and observing stdout, stderr, and process status.

### Interfaces
- I-003-001 - CLI process contract (Interface 001 of Component C-003): The command-line invocation, stdout/stderr streams, and process exit status exposed by `scip-search`.

### Out of Scope
- Missing `--index` flag or `--index` without a value; those are CAP-001 usage failures.
- Parsing readable files as SCIP data; that belongs to scip-index-loading-boundary.md.
- Query-specific flags, query traversal, result filtering, success JSON schemas, exact stderr text, exact numeric exit code taxonomy, install/version behavior, ctags fallback behavior, custom index formats, caching, watching, generating, or updating indexes.

### Assumptions
- **ASM-000-1**: A selected index path that does not exist is an index-loading failure rather than a usage failure. - *Why*: CAP-001 already accepted the invocation shape, while CAP-002 owns missing caller-provided SCIP inputs. - Confidence: HIGH
- **ASM-000-2**: A selected path that cannot be opened for reading must fail without trying alternative paths. - *Why*: The epic requires explicit caller-selected indexes to support concurrent worktrees. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Reject a Nonexistent Selected Index Path

### References
- goal spec: README.md#what-is-scip-search - Requires loading the index file at the path provided as an argument.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-002---load-the-caller-selected-scip-index - Covers distinguishing missing index input from successful loading.

### User Story
**As an** automation agent running `scip-search` in concurrent worktrees, **I want to** receive a deterministic failure when the selected index path does not exist, **so that** I can fix the index location before relying on query results.

### Acceptance Criteria
- AC-001-1: Given a documented query command is invoked with `--index <index-path>` and `<index-path>` does not identify an existing filesystem entry, when index path validation runs, then the process reports an index-loading failure before query-specific traversal starts.
- AC-001-2: Given the selected index path is missing, when the failure is reported, then the runtime does not search default locations, infer a repository-local index, generate a new index, or update any existing index.
- AC-001-3: Given the selected index path is missing, when the process exits, then the selected command does not produce a successful query result.

### Depends on:
Implementation ordering:
- Story document cli-shared-index-flag.md - The runtime must first accept `--index <index-path>` as the selected index path.

Run time coupling:
- Interface I-003-001 - CLI process contract

### Out of Scope
- Missing `--index` flag or missing flag value handling.
- Exact diagnostic wording or exact numeric exit code.
- Query traversal behavior after a successful load.

### Assumptions
- **ASM-001-1**: "Missing SCIP input" includes a caller-selected path that does not exist, not an omitted flag. - *Why*: The sibling CAP-001 story already owns omitted flag usage failures. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-002 - Reject an Unreadable Selected Index Path

### References
- goal spec: README.md#what-is-scip-search - Requires loading the caller-provided index file before answering a query.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-002---load-the-caller-selected-scip-index - Covers clear loading failures for unreadable index input.

### User Story
**As a** CLI developer implementing the runtime boundary, **I want to** reject selected index paths that cannot be read as input files, **so that** the command fails clearly before any query-specific behavior depends on absent index data.

### Acceptance Criteria
- AC-002-1: Given a documented query command is invoked with `--index <index-path>` and `<index-path>` exists but cannot be opened for reading by the current process, when index path validation runs, then the process reports an index-loading failure before query-specific traversal starts.
- AC-002-2: Given the selected index path identifies a directory or other non-file input that cannot be loaded as a SCIP index file, when index path validation runs, then the process reports an index-loading failure before query-specific traversal starts.
- AC-002-3: Given the selected index path cannot be read, when the failure is reported, then the runtime does not generate, rewrite, delete, or otherwise mutate that path.

### Depends on:
Implementation ordering:
- Story document cli-shared-index-flag.md - The runtime must first accept `--index <index-path>` as the selected index path.

Run time coupling:
- Interface I-003-001 - CLI process contract

### Out of Scope
- Distinguishing every operating-system-specific permission or file-type error in acceptance criteria.
- Parsing readable but invalid SCIP content.
- Exact diagnostic wording or exact numeric exit code.

### Assumptions
- **ASM-002-1**: Directory and permission failures may share the same documented loading-failure class. - *Why*: The source requires clear loading failure behavior but does not define separate failure taxonomies for path errors. - Confidence: MEDIUM

### Open Questions
- None.
