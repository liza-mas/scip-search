# User Stories: Package Query Fixtures

Status: review

## Goal
`scip-search` maintainers can validate `packages` behavior with deterministic SCIP fixtures and package-specific golden cases for all-package listing, literal prefix filtering, de-duplication, empty results, and stable package ordering.

## Parent Epic
`specs/epics/readme/20260517-134857-epic-planning-3.md` - Capability CAP-003, "Validate discovery queries with fixtures"

## Context
Package discovery must be stable enough for automation to compare package inventories and filtered package lists across repeated runs. This document defines the package-query fixture coverage needed to prove CAP-002 package identity, de-duplication, prefix filtering, empty-result, and payload stories without absorbing reference, implementation, runtime, or traversal-fixture scope.

## Personas
- **Automation Agent**: an AI or script-driven caller running `scip-search` in a terminal or sandbox, needing deterministic JSON results and exact package identities for query planning.
- **CLI Maintainer**: a Go developer maintaining `scip-search`, needing bounded query semantics and fixtures that keep symbol and package behavior stable across SCIP indexes.

## General information

Applies to: package-query fixture and golden-case expectations for successful `packages` discovery with and without `--prefix`.

### References
- goal spec: `README.md#scip-symbol-format` - Defines SCIP package identity components before descriptors.
- goal spec: `README.md#what-is-scip-search` - Documents the `scip-search packages --index <index-path> [--prefix <prefix>]` command form and structured JSON stdout.
- parent epic: `specs/epics/readme/20260517-134857-epic-planning-3.md#general-information` - Requires deterministic, small query-specific fixtures that use the same SCIP loading and traversal path as commands.
- parent epic: `specs/epics/readme/20260517-134857-epic-planning-3.md#capability-cap-003---validate-discovery-queries-with-fixtures` - Requires package fixtures for shared and non-shared prefixes, empty results, stable ordering, and package identities.
- sibling story: `specs/stories/readme/20260517-134857-epic-planning-3/CAP-002-01-package-inventory.md` - Defines all-package listing, package identity de-duplication, and stable package ordering.
- sibling story: `specs/stories/readme/20260517-134857-epic-planning-3/CAP-002-02-package-prefix-filtering.md` - Defines literal package-name prefix filtering and empty successful filtered results.
- sibling story: `specs/stories/readme/20260517-134857-epic-planning-3/CAP-002-03-package-result-json-shape.md` - Defines the successful `packages` payload fields and package key expectations.
- consistency scan: `specs/stories/readme/20260517-134857-epic-planning-3/` - Existing CAP-001 and CAP-002 story documents establish query-specific fixture expectations while leaving shared runtime and traversal fixtures out of scope.

### Non-Functional Requirements
- NFR-000-1: Package-query fixtures must be deterministic and small enough for normal CLI test runs.
- NFR-000-2: Package-query validation must exercise the same SCIP loading and traversal path used by the `packages` command.
- NFR-000-3: Package golden cases must preserve package identity components exactly as derived from fixture SCIP symbols.
- NFR-000-4: Package fixture stories must not redefine shared command invocation, index loading, stdout/stderr placement, process status, or raw traversal fixture construction.

### Related External Components
- Component C-001 - SCIP symbol identifiers: Human-readable symbol strings containing scheme, package manager, package name, package version, and descriptors.
- Component C-002 - SCIP traversal view: The downstream query input that exposes indexed symbol data from which package identities can be discovered.
- Component C-003 - Query fixture set: Deterministic SCIP test data and expected JSON cases used by maintainers to validate symbol and package queries.

### Interfaces
- I-001-002 - Package discovery query contract (Interface 002 of Component C-001): The `packages` query returns package identities from the index, optionally filtered by package-name prefix.

### Out of Scope
- Symbol query fixture expectations, symbol partial matching, and full symbol string match context.
- Shared traversal fixtures for documents, occurrences, relationships, hover metadata, and source range preservation.
- Reference and implementation query fixtures, exact-symbol lookup behavior, missing exact-symbol behavior, and relationship traversal cases.
- Large real-world repository fixtures, performance benchmarks, external indexer installation, ctags fallback fixtures, and generated indexer workflows.
- Shared runtime fixtures for command routing, `--index`, stdout/stderr failures, exit status, and malformed index loading.

### Assumptions
- **ASM-000-1**: Package-query fixtures may reuse shared traversal fixture data as long as the query-specific cases own only command inputs and expected package-query payloads. - *Why*: CAP-003 explicitly depends on epic-planning-2 traversal fixtures while owning discovery-query behavior. - Confidence: HIGH
- **ASM-000-2**: Package golden cases assert the same package identity fields named by CAP-002 package result stories. - *Why*: CAP-003 validates behavior from CAP-002 rather than defining an alternate package payload. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Provide Package Fixture Data for Inventory and Prefix Cases

### References
- parent epic: `specs/epics/readme/20260517-134857-epic-planning-3.md#capability-cap-003---validate-discovery-queries-with-fixtures` - Requires fixtures with packages that share and do not share a prefix, empty successful results, stable ordering, and package identities.
- sibling story: `specs/stories/readme/20260517-134857-epic-planning-3/CAP-002-01-package-inventory.md#story-st-001---list-all-indexed-package-identities` - Defines all-package listing, identity de-duplication, distinct identities, and stable order.
- sibling story: `specs/stories/readme/20260517-134857-epic-planning-3/CAP-002-02-package-prefix-filtering.md#story-st-001---filter-packages-by-literal-package-name-prefix` - Defines literal prefix filtering by package name only.

### User Story
**As a** CLI Maintainer, **I want to** maintain deterministic package fixture data with duplicate and distinct package identities, **so that** `packages` validation can prove inventory listing, de-duplication, and prefix filtering without relying on external indexers or large repositories.

### Acceptance Criteria
- AC-001-1: Given the package-query fixture data, when maintainers inspect the indexed symbols used by the fixture, then it includes multiple symbols that share the same scheme, package manager, package name, and package version.
- AC-001-2: Given the package-query fixture data, when maintainers inspect the indexed symbols used by the fixture, then it includes at least two distinct package identities.
- AC-001-3: Given the package-query fixture data, when maintainers inspect package names in the fixture, then at least two package names share a prefix that can be used by a filtered `packages --prefix <prefix>` case.
- AC-001-4: Given the package-query fixture data, when maintainers inspect package names in the fixture, then at least one package name does not share that prefix and should be excluded by the filtered case.
- AC-001-5: Given the package-query fixture data, when maintainers inspect planned prefix inputs, then it includes at least one accepted prefix value that matches no package name.
- AC-001-6: Given the package-query fixture data is used by validation, when maintainers run the fixture-backed package cases repeatedly, then the fixture inputs do not depend on an external indexer installation or mutable repository state.

### Depends on:
Implementation ordering:
- Story document `CAP-002-01-package-inventory.md` - Package identity and de-duplication behavior must be defined before fixture cases can assert them.
- Story document `CAP-002-02-package-prefix-filtering.md` - Prefix filtering behavior must be defined before fixture cases can assert it.

Run time coupling:
- I-001-002 - Package discovery query contract

### Out of Scope
- Defining symbol-query fixture data.
- Creating shared traversal fixture coverage for source ranges, hover text, relationships, or occurrence lookup behavior.
- Testing dependency graphs, registry lookups, version resolution, language filters, descriptor filters, regex filters, fuzzy filters, semantic filters, or cross-index behavior.

### Assumptions
- **ASM-001-1**: Fixture package identities may be synthetic as long as their SCIP package components are valid and deterministic. - *Why*: CAP-003 excludes large real-world fixtures and external indexer installation while requiring deterministic coverage. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-002 - Assert Package Query Golden Cases

### References
- parent epic: `specs/epics/readme/20260517-134857-epic-planning-3.md#capability-cap-003---validate-discovery-queries-with-fixtures` - Requires golden JSON cases proving package identities, prefix filtering, ordering, and empty collections.
- sibling story: `specs/stories/readme/20260517-134857-epic-planning-3/CAP-002-01-package-inventory.md` - Defines unfiltered package listing, de-duplication, and stable package ordering.
- sibling story: `specs/stories/readme/20260517-134857-epic-planning-3/CAP-002-02-package-prefix-filtering.md` - Defines literal package-name prefix filtering and successful empty filtered results.
- sibling story: `specs/stories/readme/20260517-134857-epic-planning-3/CAP-002-03-package-result-json-shape.md` - Defines `packages` collection shape and package identity fields.

### User Story
**As a** CLI Maintainer, **I want to** compare `packages` results against package-specific golden JSON cases, **so that** future changes preserve all-package listing, literal prefix filtering, de-duplication, empty results, and deterministic ordering.

### Acceptance Criteria
- AC-002-1: Given the package-query fixture data contains multiple symbols from the same package identity, when maintainers run the unfiltered inventory golden case, then the expected JSON contains that package identity once in the `packages` collection.
- AC-002-2: Given the package-query fixture data contains multiple distinct package identities, when maintainers run the unfiltered inventory golden case, then the expected JSON contains every distinct package identity.
- AC-002-3: Given the package-query fixture data contains package names that begin with the supplied prefix and package names that do not, when maintainers run the prefix golden case, then the expected JSON includes only package identities whose package name begins with that prefix.
- AC-002-4: Given the package-query fixture data contains no package name that begins with the supplied prefix, when maintainers run the empty filtered golden case, then the expected JSON contains an empty `packages` collection.
- AC-002-5: Given a package-query golden case returns more than one package identity, when maintainers compare the expected JSON to actual output, then the expected `packages` order is stable and follows the observable package ordering defined by CAP-002.
- AC-002-6: Given a non-empty package-query golden case is evaluated, when maintainers inspect each expected result entry, then each entry includes the package result fields required by the CAP-002 payload story and does not include symbol descriptors, references, implementations, or dependency graph data.

### Depends on:
Implementation ordering:
- Story ST-001 - Provide Package Fixture Data for Inventory and Prefix Cases
- Story document `CAP-002-03-package-result-json-shape.md` - Package JSON fields must be defined before golden package entries can assert them.

Run time coupling:
- I-001-002 - Package discovery query contract

### Out of Scope
- Golden cases for `symbols`, `references`, or `implementations`.
- Shared runtime error golden cases for bad flags, missing index paths, unreadable indexes, malformed SCIP files, stderr diagnostics, or exit statuses.
- Query result enrichment from registry metadata, dependency graph data, source files, hover text, or relationship traversal.

### Assumptions
- **ASM-002-1**: Package golden JSON cases assert query-specific payload contents and rely on the shared runtime story for stdout-only success behavior. - *Why*: CAP-003 coordinates with epic-planning-1 but does not own shared stream contracts. - Confidence: HIGH

### Open Questions
- None.
