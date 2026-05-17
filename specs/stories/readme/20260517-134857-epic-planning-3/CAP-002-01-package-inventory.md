# User Stories: Package Inventory Listing

Status: review

## Goal
`scip-search packages` returns one deterministic entry for each indexed package identity discoverable from the selected SCIP index.

## Parent Epic
specs/epics/readme/20260517-134857-epic-planning-3.md - Capability CAP-002

## Context
Package discovery lets automation inspect which SCIP packages are present before choosing later symbol queries. This document covers the unfiltered `packages` behavior and de-duplication of package identities that appear across many symbols in the same index.

## Personas
- **Automation Agent**: An AI or script-driven caller running `scip-search` in a terminal or sandbox, needing deterministic JSON results that can be compared across runs.
- **CLI Maintainer**: A Go developer maintaining `scip-search`, needing bounded package query semantics and fixture expectations that keep package discovery stable across SCIP indexes.

## General information

Applies to: unfiltered package inventory returned by a successful `packages` query.

### References
- goal spec: README.md#scip-symbol-format - Defines SCIP symbol package components: scheme, package manager, package name, package version, and descriptors.
- goal spec: README.md#what-is-scip-search - Documents `scip-search packages --index <index-path> [--prefix <prefix>]`.
- goal spec: README.md#complementary-existing-tool - Frames `packages` as the query used to answer what packages exist.
- parent epic: specs/epics/readme/20260517-134857-epic-planning-3.md#capability-cap-002---list-indexed-packages-with-prefix-filtering - Defines all-package listing, de-duplication, and deterministic package identities.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/stdout-json-success-contract.md - Confirms this document must not redefine shared success stream behavior.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/stderr-exit-error-contract.md - Confirms this document must not redefine shared runtime failure behavior.

### Non-Functional Requirements
- NFR-000-1: Package inventory ordering must be deterministic for automation and stable across repeated runs over the same index.
- NFR-000-2: Package inventory behavior must not redefine command routing, `--index` handling, stdout/stderr stream rules, process status, or shared runtime failures.

### Related External Components
- Component C-001 - SCIP symbol identifiers: Human-readable symbol strings containing scheme, package manager, package name, package version, and descriptors.
- Component C-002 - SCIP traversal view: The downstream query input that exposes indexed symbol data from which package identities can be discovered.

### Interfaces
- I-001-002 - Package discovery query contract (Interface 002 of Component C-001): The `packages` query returns package identities from the index.

### Out of Scope
- `--prefix` filtering behavior beyond preserving compatibility with the same package identity model.
- Querying package registries, dependency graphs, module graphs, transitive dependencies, or version resolution.
- Filtering by descriptor text, symbol name, document path, language, package manager, package version, or regular expression.
- Raw SCIP parsing, traversal view construction, and fixture file construction outside the package-query expectations stated here.
- Shared invocation failures, index-loading failures, stdout/stderr rules, and process exit behavior.

### Assumptions
- **ASM-000-1**: A package identity is the tuple of SCIP symbol components before descriptors: scheme, package manager, package name, and package version. - *Why*: The capability says packages are derived from SCIP symbol prefixes and the README defines those components before descriptors. - Confidence: HIGH
- **ASM-000-2**: Stable package ordering means ascending order by the exact package key exposed in package results. - *Why*: The capability requires stable ordering and an exact package key suitable for display or comparison by automation. - Confidence: MEDIUM

### Open Questions
- None.

---

## Story ST-001 - List All Indexed Package Identities

### References
- parent epic: specs/epics/readme/20260517-134857-epic-planning-3.md#capability-cap-002---list-indexed-packages-with-prefix-filtering - Requires `packages` without `--prefix` to return all indexed package identities.
- goal spec: README.md#scip-symbol-format - Defines the SCIP package identity components that appear before descriptors.
- goal spec: README.md#what-is-scip-search - Documents the unfiltered `packages` command form.

### User Story
**As an** Automation Agent inspecting a SCIP index in a worktree, **I want to** run `scip-search packages` without a prefix filter, **so that** I can see every package identity represented in the index before choosing follow-up symbol queries.

### Acceptance Criteria
- AC-001-1: Given a selected SCIP index contains symbols from one or more package identities, when a successful `packages` query is run without `--prefix`, then the result includes every distinct package identity discoverable from those symbols.
- AC-001-2: Given multiple symbols share the same scheme, package manager, package name, and package version, when package identities are listed, then that package identity appears once in the result.
- AC-001-3: Given two symbols differ by scheme, package manager, package name, or package version, when package identities are listed, then they are treated as distinct package identities.
- AC-001-4: Given the same selected index is queried repeatedly, when package identities are listed, then the result order is the same across runs.
- AC-001-5: Given the package-query fixture contains several symbols from the same package and at least one symbol from another package, when the unfiltered golden case is evaluated, then it proves both all-package listing and de-duplication.

### Depends on:
Run time coupling:
- I-001-002 - Package discovery query contract

### Out of Scope
- Prefix matching rules for `--prefix`.
- Defining stdout stream purity, stderr diagnostics, exit status, missing `--index`, unreadable index, or invalid SCIP failure behavior.
- Returning symbol descriptors, document paths, references, implementations, dependency edges, registry metadata, or resolved package versions.

### Assumptions
- **ASM-001-1**: De-duplication occurs by full package identity, not by package name alone. - *Why*: SCIP package identities include scheme, package manager, name, and version; collapsing by name alone would lose user-visible package distinctions. - Confidence: HIGH

### Open Questions
- None.
