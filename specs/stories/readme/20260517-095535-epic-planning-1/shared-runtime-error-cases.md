# User Stories: Shared Runtime Error Cases

Status: review

## Goal
Shared invocation and index-loading error cases are covered across documented `scip-search` commands without redefining query-specific validation or result schemas.

## Parent Epic
specs/epics/readme/20260517-095535-epic-planning-1.md - Capability CAP-003

## Context
CAP-001 and CAP-002 define the shared failure sources. This document ensures those sources are exercised consistently across the command surface through the CAP-003 stream and status contract.

## Personas
- **Automation agent**: A coding or planning agent running terminal commands inside concurrent worktrees, needing stable machine-readable output and deterministic failures.
- **CLI developer**: A Go developer implementing and testing `scip-search`, needing a bounded runtime contract that separates CLI behavior from query traversal.

## General information

Applies to: shared invocation and index-loading runtime error cases across documented commands.

### References
- goal spec: README.md#what-is-scip-search - Lists documented query command forms and requires structured JSON stdout plus process exit.
- parent epic: specs/epics/readme/20260517-095535-epic-planning-1.md#capability-cap-003---report-machine-readable-results-and-actionable-failures - Requires shared invocation and index-loading failures to follow the runtime stream/status contract.
- dependency story: specs/stories/readme/20260517-095535-epic-planning-1/cli-command-routing-and-usage.md - Defines missing and unsupported command failures.
- dependency story: specs/stories/readme/20260517-095535-epic-planning-1/cli-shared-index-flag.md - Defines shared `--index` flag failures across commands.
- dependency story: specs/stories/readme/20260517-095535-epic-planning-1/index-path-validation.md - Defines nonexistent and unreadable selected index failures.
- dependency story: specs/stories/readme/20260517-095535-epic-planning-1/scip-index-loading-boundary.md - Defines invalid SCIP input failures.
- sibling CAP-003 story: specs/stories/readme/20260517-095535-epic-planning-1/stderr-exit-error-contract.md - Defines the shared failure stream and status behavior these cases must satisfy.
- task: epic-planning-1-us-writing-2 - Requires shared invocation and index-loading error cases to be covered across commands.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/ - Existing CAP-001 and CAP-002 story documents were read to avoid changing their failure ownership.

### Non-Functional Requirements
- NFR-000-1: Shared runtime error coverage must preserve stdout as a success-only JSON channel and stderr as the runtime failure diagnostic channel.
- NFR-000-2: Coverage must apply to `symbols`, `references`, `implementations`, and `packages` without requiring query-specific traversal behavior.
- NFR-000-3: Error case coverage must not generate indexes, update indexes, compile source, type-check source, start a daemon, start a watcher, or require interactive correction.

### Related External Components
- Component C-001 - SCIP index file: The on-disk input selected by the caller with `--index`.
- Component C-003 - Calling process environment: The shell, script, or agent invoking `scip-search` and observing stdout, stderr, and process status.

### Interfaces
- I-003-001 - CLI process contract (Interface 001 of Component C-003): The command-line invocation, stdout/stderr streams, and process exit status exposed by `scip-search`.

### Out of Scope
- Query-specific validation for `--name`, `--symbol`, `--prefix`, ambiguous symbols, empty query results, or result ordering.
- Exact result schemas for successful commands.
- Exact stderr wording, pretty-printing, progress output, configurable logging, human UI output, distribution documentation, install verification, and version output.

### Assumptions
- **ASM-000-1**: "Covered across commands" means the shared runtime contract is observable for each documented command where the failure can occur before query traversal. - *Why*: The task requires cross-command coverage while excluding query-specific behavior. - Confidence: HIGH
- **ASM-000-2**: Root-level missing and unsupported command cases are covered once because no documented query command has been selected yet. - *Why*: Those cases occur before command routing can identify `symbols`, `references`, `implementations`, or `packages`. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Cover Shared Invocation Failures

### References
- dependency story: specs/stories/readme/20260517-095535-epic-planning-1/cli-command-routing-and-usage.md - Defines missing and unsupported command usage failures.
- dependency story: specs/stories/readme/20260517-095535-epic-planning-1/cli-shared-index-flag.md - Defines missing `--index` and missing index value usage failures.
- sibling CAP-003 story: specs/stories/readme/20260517-095535-epic-planning-1/stderr-exit-error-contract.md - Defines stderr-only diagnostics and documented nonzero status for failures.

### User Story
**As a** CLI developer validating the shared runtime, **I want to** cover invocation failures through the common stderr and status contract, **so that** invalid command shapes fail consistently before any query-specific work starts.

### Acceptance Criteria
- AC-001-1: Given `scip-search` is invoked without a query command, when the runtime reports the failure, then the failure satisfies the shared stderr-only diagnostic contract and exits with usage-failure status `2`.
- AC-001-2: Given `scip-search` is invoked with an unsupported command name, when the runtime reports the failure, then the failure satisfies the shared stderr-only diagnostic contract and exits with usage-failure status `2`.
- AC-001-3: Given `symbols` is invoked without `--index` or with `--index` but no value, when the runtime reports the failure, then the failure satisfies the shared stderr-only diagnostic contract and exits with usage-failure status `2`.
- AC-001-4: Given `references` is invoked without `--index` or with `--index` but no value, when the runtime reports the failure, then the failure satisfies the shared stderr-only diagnostic contract and exits with usage-failure status `2`.
- AC-001-5: Given `implementations` is invoked without `--index` or with `--index` but no value, when the runtime reports the failure, then the failure satisfies the shared stderr-only diagnostic contract and exits with usage-failure status `2`.
- AC-001-6: Given `packages` is invoked without `--index` or with `--index` but no value, when the runtime reports the failure, then the failure satisfies the shared stderr-only diagnostic contract and exits with usage-failure status `2`.

### Depends on:
Implementation ordering:
- Story document cli-command-routing-and-usage.md - Missing and unsupported command failures must be defined before CAP-003 coverage can be tested.
- Story document cli-shared-index-flag.md - Shared index flag failures must be defined before CAP-003 coverage can be tested.
- Story document stderr-exit-error-contract.md - The shared failure stream and status contract must exist before specific cases can assert it.

Run time coupling:
- Interface I-003-001 - CLI process contract

### Out of Scope
- Query-specific missing flags such as `--name`, `--symbol`, or `--prefix`.
- Suggestions for unsupported command names.
- Human-readable help output formatting.

### Assumptions
- **ASM-001-1**: Missing query-specific flags are not part of shared invocation failure coverage unless a later query story classifies them under the same runtime failure contract. - *Why*: The task explicitly excludes query-specific result schemas and prior CAP-001 stories only define shared invocation inputs. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-002 - Cover Shared Index-Loading Failures Across Commands

### References
- dependency story: specs/stories/readme/20260517-095535-epic-planning-1/index-path-validation.md - Defines nonexistent and unreadable selected index failures.
- dependency story: specs/stories/readme/20260517-095535-epic-planning-1/scip-index-loading-boundary.md - Defines invalid SCIP input failures.
- sibling CAP-003 story: specs/stories/readme/20260517-095535-epic-planning-1/stderr-exit-error-contract.md - Defines stderr-only diagnostics and documented nonzero status for failures.

### User Story
**As an** automation agent running any documented query command, **I want to** see the same stream and status behavior for shared index-loading failures, **so that** index problems are not mistaken for command-specific empty results.

### Acceptance Criteria
- AC-002-1: Given `symbols` is invoked with a nonexistent, unreadable, directory, or invalid SCIP selected index path, when the runtime reports the failure, then the failure satisfies the shared stderr-only diagnostic contract and exits with index-loading-failure status `3` before query traversal starts.
- AC-002-2: Given `references` is invoked with a nonexistent, unreadable, directory, or invalid SCIP selected index path, when the runtime reports the failure, then the failure satisfies the shared stderr-only diagnostic contract and exits with index-loading-failure status `3` before query traversal starts.
- AC-002-3: Given `implementations` is invoked with a nonexistent, unreadable, directory, or invalid SCIP selected index path, when the runtime reports the failure, then the failure satisfies the shared stderr-only diagnostic contract and exits with index-loading-failure status `3` before query traversal starts.
- AC-002-4: Given `packages` is invoked with a nonexistent, unreadable, directory, or invalid SCIP selected index path, when the runtime reports the failure, then the failure satisfies the shared stderr-only diagnostic contract and exits with index-loading-failure status `3` before query traversal starts.
- AC-002-5: Given any documented query command fails during shared index loading, when the process exits, then the command does not emit a successful empty-result JSON payload.

### Depends on:
Implementation ordering:
- Story document cli-shared-index-flag.md - The runtime must establish the selected index path before loading failures can be tested.
- Story document index-path-validation.md - Path-level loading failures must be defined before CAP-003 coverage can assert stream/status behavior.
- Story document scip-index-loading-boundary.md - Invalid SCIP loading failures must be defined before CAP-003 coverage can assert stream/status behavior.
- Story document stderr-exit-error-contract.md - The shared failure stream and status contract must exist before specific cases can assert it.

Run time coupling:
- Interface I-003-001 - CLI process contract

### Out of Scope
- Query traversal failures after a valid index has loaded.
- Empty result semantics for valid queries.
- Recovering partial results from malformed SCIP input.

### Assumptions
- **ASM-002-1**: Directory inputs can be covered with unreadable or invalid selected index cases as long as the observable failure class is documented and nonzero. - *Why*: CAP-002 allowed directory and permission failures to share the loading-failure class. - Confidence: MEDIUM

### Open Questions
- None.
