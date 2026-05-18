# User Stories: Package Prefix Filtering

Status: review

## Goal
`scip-search packages --prefix <prefix>` returns only package identities whose package name begins with the literal requested prefix, including successful empty results.

## Parent Epic
specs/epics/readme/20260517-134857-epic-planning-3.md - Capability CAP-002

## Context
Prefix filtering lets automation narrow package discovery while preserving deterministic success behavior. This document covers package-name prefix matching, empty successful results, and ordering of filtered package lists.

## Personas
- **Automation Agent**: An AI or script-driven caller running `scip-search` in a terminal or sandbox, needing deterministic filtered JSON results without interpreting missing matches as runtime failures.
- **CLI Maintainer**: A Go developer maintaining `scip-search`, needing precise filter semantics that can be validated with small package-query fixtures.

## General information

Applies to: successful `packages` queries that include `--prefix <prefix>`.

### References
- goal spec: README.md#scip-symbol-format - Defines the package name component of SCIP symbols.
- goal spec: README.md#what-is-scip-search - Documents `scip-search packages --index <index-path> [--prefix <prefix>]`.
- parent epic: specs/epics/readme/20260517-134857-epic-planning-3.md#capability-cap-002---list-indexed-packages-with-prefix-filtering - Defines literal package-name prefix filtering and empty results.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/stdout-json-success-contract.md - Confirms successful empty results belong in query-specific JSON, not shared runtime errors.
- consistency check: specs/stories/readme/20260517-095535-epic-planning-1/stderr-exit-error-contract.md - Confirms this document must not redefine shared runtime failure behavior.

### Non-Functional Requirements
- NFR-000-1: Prefix filtering must be deterministic and literal so automation can compare filtered results without fuzzy matching ambiguity.
- NFR-000-2: Successful no-match filtered queries must remain successful query results, not shared runtime failures.

### Related External Components
- Component C-001 - SCIP symbol identifiers: Human-readable symbol strings containing package names.
- Component C-002 - SCIP traversal view: The downstream query input that exposes indexed symbol data from which package identities can be discovered.

### Interfaces
- I-001-002 - Package discovery query contract (Interface 002 of Component C-001): The `packages` query returns package identities from the index, optionally filtered by package-name prefix.

### Out of Scope
- Filtering by scheme, package manager, package version, descriptors, symbol names, document paths, language, regular expression, glob, fuzzy text, or semantic similarity.
- Dependency graph analysis, package registry lookups, module graph analysis, and version resolution.
- Shared invocation failures, index-loading failures, stdout/stderr rules, and process exit behavior.
- Raw SCIP traversal construction or package fixture generation mechanics.

### Assumptions
- **ASM-000-1**: Prefix matching is case-sensitive. - *Why*: The parent epic assumes literal, case-sensitive matching unless a source-backed reason says otherwise, and SCIP package names preserve spelling. - Confidence: MEDIUM
- **ASM-000-2**: Prefix filtering is applied to de-duplicated package identities and yields the same visible result as filtering before de-duplication. - *Why*: The user-visible contract is a package identity list, while duplicate symbols are not package results. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Filter Packages by Literal Package Name Prefix

### References
- parent epic: specs/epics/readme/20260517-134857-epic-planning-3.md#capability-cap-002---list-indexed-packages-with-prefix-filtering - Requires `--prefix` to match package-name prefixes.
- goal spec: README.md#what-is-scip-search - Documents the optional `--prefix <prefix>` form.

### User Story
**As an** Automation Agent narrowing package discovery in a terminal workflow, **I want to** filter `packages` by a literal package-name prefix, **so that** I can inspect only the indexed packages relevant to a planned symbol query.

### Acceptance Criteria
- AC-001-1: Given a selected SCIP index contains package identities with package names that begin with the requested literal prefix, when a successful `packages --prefix <prefix>` query is run, then the result includes those matching package identities.
- AC-001-2: Given a selected SCIP index contains package identities whose package names do not begin with the requested literal prefix, when the filtered query is run, then those package identities are excluded from the result.
- AC-001-3: Given the requested prefix text appears only in a package manager, package version, scheme, descriptor, symbol name, or document path, when the filtered query is run, then that text alone does not cause a package identity to match.
- AC-001-4: Given matching package identities are returned for a filtered query, when the caller compares result order across repeated runs on the same index, then the order is stable and follows the same package ordering used by unfiltered package listing.
- AC-001-5: Given the package-query fixture contains at least one matching package name, one non-matching package name, and one package where the prefix appears outside the package-name component, when the filtered golden case is evaluated, then it proves literal package-name prefix filtering.

### Depends on:
Implementation ordering:
- Story document CAP-002-01-package-inventory.md - The package identity model and unfiltered ordering must be defined before filtered package results can preserve them.

Run time coupling:
- I-001-002 - Package discovery query contract

### Out of Scope
- Regular expression, glob, fuzzy, case-folded, semantic, descriptor, version, package-manager, or language filters.
- Returning why excluded packages did not match.
- Defining query-specific validation errors for malformed option syntax beyond the shared command-line contracts.

### Assumptions
- **ASM-001-1**: A prefix value that is syntactically accepted by the shared CLI is treated as literal text, not as a pattern language. - *Why*: The capability says literal prefix filtering and excludes regular expressions and non-package filters. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-002 - Return Successful Empty Package Results

### References
- parent epic: specs/epics/readme/20260517-134857-epic-planning-3.md#capability-cap-002---list-indexed-packages-with-prefix-filtering - Requires empty-result behavior for package prefix filtering.
- parent epic: specs/epics/readme/20260517-134857-epic-planning-3.md#general-information - Assumes successful no-match discovery queries return empty result collections rather than shared runtime errors.

### User Story
**As an** Automation Agent probing an unfamiliar index, **I want to** receive a successful empty package result when no package names match my prefix, **so that** I can distinguish absence of matches from invocation or index-loading failures.

### Acceptance Criteria
- AC-002-1: Given the selected SCIP index has no package names that begin with the requested prefix, when a successful `packages --prefix <prefix>` query is run, then the result contains an empty `packages` collection.
- AC-002-2: Given a filtered query returns no package matches, when the caller evaluates the query-specific payload, then the payload represents an empty successful result rather than an error result.
- AC-002-3: Given the package-query fixture includes a prefix that matches no package names, when the empty-result golden case is evaluated, then it proves that no-match prefix filtering returns `packages` as an empty collection.

### Depends on:
Implementation ordering:
- Story ST-001 - Filter Packages by Literal Package Name Prefix

Run time coupling:
- I-001-002 - Package discovery query contract

### Out of Scope
- Shared nonzero failure status, stderr diagnostics, or stdout suppression for actual runtime failures.
- Suggestions for similar package names or alternative prefixes.
- Treating no-match results as missing-index, invalid-index, or unsupported-command failures.

### Assumptions
- **ASM-002-1**: Empty successful results use the same top-level `packages` collection as non-empty successful results. - *Why*: The parent epic requires explicit empty result collections for successful no-match cases. - Confidence: HIGH

### Open Questions
- None.
