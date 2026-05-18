# User Stories: Package Result JSON Shape

Status: review

## Goal
Successful package discovery results expose a stable `packages` payload whose entries identify each package by SCIP identity components and an exact package key.

## Parent Epic
specs/epics/readme/20260517-134857-epic-planning-3.md - Capability CAP-002

## Context
The shared runtime contract owns JSON-only stdout on success. This document defines only the package query payload that appears inside that successful JSON result, so automation can compare package identities without parsing full symbol descriptors.

## Personas
- **Automation Agent**: An AI or script-driven caller parsing `scip-search` output, needing stable package fields for comparison and follow-up query planning.
- **CLI Maintainer**: A Go developer maintaining `scip-search`, needing a bounded query-specific JSON contract and golden cases for package results.

## General information

Applies to: successful unfiltered and filtered package query payloads.

### References
- goal spec: README.md#scip-symbol-format - Defines SCIP symbol identity components, including scheme, package manager, package name, and package version.
- goal spec: README.md#what-is-scip-search - Requires `scip-search` to print structured JSON for query results.
- parent epic: specs/epics/readme/20260517-134857-epic-planning-3.md#capability-cap-002---list-indexed-packages-with-prefix-filtering - Requires a `packages` collection with identity components and an exact package key.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/stdout-json-success-contract.md - Owns stdout stream purity and leaves query-specific result fields to this document.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/stderr-exit-error-contract.md - Owns shared runtime failures, which this document does not redefine.

### Non-Functional Requirements
- NFR-000-1: Package result JSON must be stable enough for automation to compare package identities across repeated runs over the same index.
- NFR-000-2: Package result JSON must expose package identity fields without requiring callers to parse descriptors from full SCIP symbols.
- NFR-000-3: Package result stories must not redefine shared stream placement, diagnostic behavior, or process status.

### Related External Components
- Component C-001 - SCIP symbol identifiers: Human-readable symbol strings containing scheme, package manager, package name, package version, and descriptors.
- Component C-003 - Calling process environment: The shell, script, or agent that parses successful JSON stdout under the shared runtime contract.

### Interfaces
- I-001-002 - Package discovery query contract (Interface 002 of Component C-001): The `packages` query returns package identities from the index, optionally filtered by package-name prefix.
- I-003-001 - CLI process contract (Interface 001 of Component C-003): The shared successful runtime contract that provides a parseable JSON value on stdout.

### Out of Scope
- Defining stdout stream purity, stderr diagnostics, exit status, missing command, unsupported command, missing `--index`, unreadable index, or invalid SCIP behavior.
- Defining result schemas for `symbols`, `references`, or `implementations`.
- Returning symbol descriptors, occurrences, source ranges, hover text, reference lists, implementation lists, dependency data, registry data, or resolved versions.
- Pretty printing, alternate output formats, progress messages, or human-only display output.

### Assumptions
- **ASM-000-1**: The package query result is a JSON object with a top-level `packages` collection. - *Why*: The parent epic explicitly names a `packages` collection and the shared runtime story permits query-specific payload schemas. - Confidence: HIGH
- **ASM-000-2**: Package entry fields are named `scheme`, `packageManager`, `packageName`, `packageVersion`, and `packageKey`. - *Why*: These names expose the README's SCIP identity components while keeping package-specific JSON readable for automation. - Confidence: MEDIUM
- **ASM-000-3**: `packageKey` is the exact SCIP package prefix made from scheme, package manager, package name, and package version, without descriptors. - *Why*: The capability requires an exact package key suitable for display or comparison, and descriptors identify symbols within packages rather than package identities. - Confidence: MEDIUM

### Open Questions
- None.

---

## Story ST-001 - Expose Package Identity Components in JSON

### References
- parent epic: specs/epics/readme/20260517-134857-epic-planning-3.md#capability-cap-002---list-indexed-packages-with-prefix-filtering - Requires package result entries to expose identity components and an exact package key.
- goal spec: README.md#scip-symbol-format - Defines SCIP package identity components before descriptors.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/stdout-json-success-contract.md - Confirms this story owns query-specific fields only.

### User Story
**As an** Automation Agent parsing successful package discovery output, **I want to** receive package entries with explicit SCIP package identity components, **so that** I can compare packages without reparsing full symbol strings.

### Acceptance Criteria
- AC-001-1: Given a successful package query returns one or more package identities, when the caller parses the query-specific JSON payload, then the top-level payload contains a `packages` collection.
- AC-001-2: Given a package identity appears in the `packages` collection, when the caller inspects that package entry, then it exposes `scheme`, `packageManager`, `packageName`, `packageVersion`, and `packageKey`.
- AC-001-3: Given a package entry is derived from a SCIP symbol, when the caller compares its identity fields to that symbol, then the fields preserve the symbol's scheme, package manager, package name, and package version exactly.
- AC-001-4: Given a package entry includes `packageKey`, when the caller compares keys across entries, then equal keys identify the same package identity and different keys identify distinct package identities.
- AC-001-5: Given package-query golden JSON cases are evaluated for unfiltered, filtered, and empty successful results, when the caller inspects them, then each case uses the same top-level `packages` collection shape.

### Depends on:
Implementation ordering:
- Story document CAP-002-01-package-inventory.md - Package identity semantics must be defined before the JSON entry fields can be validated.

Run time coupling:
- I-001-002 - Package discovery query contract
- I-003-001 - CLI process contract

### Out of Scope
- Shared successful stdout stream rules, stderr behavior, process status, runtime failures, or command routing.
- Returning full symbol descriptors, symbol locations, references, implementations, dependency graph edges, registry metadata, or resolved package versions.
- Supporting alternate field names, alternate output formats, or human-readable package tables.

### Assumptions
- **ASM-001-1**: The JSON shape remains the same for unfiltered and prefix-filtered successful package results. - *Why*: Both commands return the same kind of package identities, differing only in which entries are included. - Confidence: HIGH

### Open Questions
- None.
