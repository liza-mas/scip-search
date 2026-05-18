# User Stories: Package Query Fixtures

Status: review

## Goal
Maintainers can validate `scip-search packages` and `scip-search packages --prefix <prefix>` against deterministic SCIP fixtures that prove package listing, package-name prefix filtering, de-duplication, empty successful results, and stable ordering.

## Parent Epic
`specs/epics/readme/20260517-134857-epic-planning-3.md` - Capability CAP-003, "Validate discovery queries with fixtures"

## Context
Package discovery behavior is defined by CAP-002. This document defines only the query-specific fixture and golden-case expectations that keep package inventory and prefix filtering stable. Shared runtime behavior and raw SCIP traversal fixture construction remain owned by sibling epics.

## Personas
- **CLI Maintainer**: a Go developer maintaining `scip-search`, needing bounded package query fixtures that catch regressions in package discovery behavior.
- **Automation Agent**: an AI or script-driven caller running `scip-search` in a terminal or sandbox, needing deterministic package JSON results for planning symbol queries.

## General information

Applies to: package discovery fixture coverage for successful unfiltered and prefix-filtered `packages` queries.

### References
- goal spec: `README.md#scip-symbol-format` - Defines SCIP package identity components within full symbol identifiers.
- goal spec: `README.md#what-is-scip-search` - Documents `scip-search packages --index <index-path> [--prefix <prefix>]`.
- parent epic: `specs/epics/readme/20260517-134857-epic-planning-3.md#general-information` - Requires deterministic ordering, explicit empty collections, and small query-specific fixtures.
- parent epic: `specs/epics/readme/20260517-134857-epic-planning-3.md#capability-cap-003---validate-discovery-queries-with-fixtures` - Defines package fixture and golden JSON validation scope.
- consistency check: `specs/stories/readme/20260517-134857-epic-planning-3/CAP-002-01-package-inventory.md` - Defines all-package listing and package identity de-duplication.
- consistency check: `specs/stories/readme/20260517-134857-epic-planning-3/CAP-002-02-package-prefix-filtering.md` - Defines literal package-name prefix filtering and empty package results.
- consistency check: `specs/stories/readme/20260517-134857-epic-planning-3/CAP-002-03-package-result-json-shape.md` - Defines package JSON identity fields and `packageKey`.
- consistency check: `specs/stories/readme/20260517-134857-epic-planning-3/cap-001-symbol-name-discovery.md` - Confirms symbol discovery fixture behavior is sibling scope.

### Non-Functional Requirements
- NFR-000-1: Package fixture data must be deterministic and small enough to run in normal CLI validation.
- NFR-000-2: Package fixture golden cases must assert package identities through observable package result fields, not internal traversal structures.
- NFR-000-3: Package fixture validation must exercise the same SCIP loading and traversal path used by the `packages` command.
- NFR-000-4: Package fixture stories must not redefine shared runtime failures, raw SCIP traversal construction, symbol query behavior, reference behavior, or implementation behavior.

### Related External Components
- Component C-001 - SCIP symbol identifiers: Human-readable symbol strings containing scheme, package manager, package name, package version, and descriptors.
- Component C-002 - SCIP traversal view: The shared traversal input that exposes indexed symbol data from which package identities are derived.
- Component C-003 - Query fixture set: Deterministic SCIP test data and expected JSON cases used by maintainers to validate discovery queries.

### Interfaces
- I-001-002 - Package discovery query contract (Interface 002 of Component C-001): The `packages` query returns package identities from the index, optionally filtered by package-name prefix.

### Out of Scope
- Symbol discovery fixture cases, symbol JSON shape, symbol ambiguity, and symbol empty results.
- Shared command routing, `--index` handling, stdout/stderr stream rules, exit status taxonomy, invalid invocation, unreadable index, and malformed SCIP errors.
- Raw SCIP fixture generation mechanics, traversal fixture coverage for documents, occurrences, relationships, ranges, and hover metadata.
- Reference and implementation query fixtures, exact-symbol missing behavior, relationship traversal cases, and location/range JSON.
- Large real-world fixtures, performance benchmarks, external indexer installation, ctags fallback fixtures, dependency graph analysis, registry lookups, and version resolution.

### Assumptions
- **ASM-000-1**: The package query fixture may reuse the same deterministic SCIP fixture as symbol query validation if the expected cases remain separated by command. - *Why*: CAP-003 allows query-specific fixtures to extend shared traversal fixtures, and separation is required at the golden-case level rather than necessarily at the physical fixture file level. - Confidence: HIGH
- **ASM-000-2**: Package fixture golden cases assert `scheme`, `packageManager`, `packageName`, `packageVersion`, and `packageKey` because CAP-002 defines those observable package result fields. - *Why*: Fixture validation should prove the existing package JSON contract rather than introduce new fields. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Cover Package Listing and De-duplication

### References
- parent epic: `specs/epics/readme/20260517-134857-epic-planning-3.md#capability-cap-003---validate-discovery-queries-with-fixtures` - Requires package query fixtures that cover packages sharing and not sharing prefixes.
- consistency check: `specs/stories/readme/20260517-134857-epic-planning-3/CAP-002-01-package-inventory.md#story-st-001---list-all-indexed-package-identities` - Defines all-package listing and de-duplication.
- consistency check: `specs/stories/readme/20260517-134857-epic-planning-3/CAP-002-03-package-result-json-shape.md#story-st-001---expose-package-identity-components-in-json` - Defines package identity JSON fields.

### User Story
**As a** CLI Maintainer, **I want to** validate unfiltered package listing with deterministic fixture symbols from repeated and distinct package identities, **so that** future changes cannot duplicate package entries or drop indexed packages.

### Acceptance Criteria
- AC-001-1: Given the package fixture contains at least two symbols from package identity `scip-go gomod github.com/liza-mas/liza .`, at least one symbol from `scip-go gomod github.com/liza-mas/scip-search .`, and at least one symbol from `scip-go gomod github.com/sourcegraph/scip-bindings .`, when maintainers run the unfiltered `packages` fixture case, then the fixture provides deterministic data for repeated and distinct package identities.
- AC-001-2: Given the unfiltered `packages` golden result is evaluated, when maintainers inspect the `packages` collection, then each distinct fixture package identity appears exactly once even when multiple symbols share that package identity.
- AC-001-3: Given two fixture package identities differ by package name, package manager, scheme, or package version, when maintainers inspect the unfiltered golden result, then those identities appear as separate package entries.
- AC-001-4: Given any non-empty package golden result, when maintainers inspect result entries, then each entry asserts `scheme`, `packageManager`, `packageName`, `packageVersion`, and `packageKey` according to the CAP-002 package JSON contract.
- AC-001-5: Given the unfiltered package query returns more than one package identity, when maintainers compare the golden result order across repeated validation runs, then entries appear in stable ascending order by observable `packageKey`.

### Depends on:
Implementation ordering:
- `CAP-002-01-package-inventory.md` - Package identity and de-duplication semantics must be defined before fixture golden cases can validate them.
- `CAP-002-03-package-result-json-shape.md` - Package result fields must be defined before golden JSON can assert them.

Run time coupling:
- Interface I-001-002 - Package discovery query contract

### Out of Scope
- Prefix filtering, empty prefix results, symbol discovery, reference discovery, implementation discovery, and raw traversal fixture construction.
- Dependency graph expansion, registry metadata, resolved versions beyond the SCIP package version component, or package-manager-specific behavior.

### Assumptions
- **ASM-001-1**: Multiple fixture symbols from the same package identity are sufficient to prove de-duplication without requiring duplicate serialized package records. - *Why*: SCIP symbols carry package identity in their symbol string, and CAP-002 derives packages from indexed symbols. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-002 - Cover Package Prefix Filtering and Empty Results

### References
- parent epic: `specs/epics/readme/20260517-134857-epic-planning-3.md#capability-cap-003---validate-discovery-queries-with-fixtures` - Requires package prefix filtering and successful empty result cases.
- consistency check: `specs/stories/readme/20260517-134857-epic-planning-3/CAP-002-02-package-prefix-filtering.md#story-st-001---filter-packages-by-literal-package-name-prefix` - Defines literal package-name prefix filtering.
- consistency check: `specs/stories/readme/20260517-134857-epic-planning-3/CAP-002-02-package-prefix-filtering.md#story-st-002---return-successful-empty-package-results` - Defines empty successful package results.

### User Story
**As a** CLI Maintainer, **I want to** validate package prefix filtering with deterministic matching, non-matching, and empty-result cases, **so that** package filtering remains literal, package-name scoped, and stable.

### Acceptance Criteria
- AC-002-1: Given the package fixture includes package names `github.com/liza-mas/liza`, `github.com/liza-mas/scip-search`, and `github.com/sourcegraph/scip-bindings`, when maintainers evaluate the golden case for `packages --prefix github.com/liza-mas/`, then the `packages` collection includes only the two `github.com/liza-mas/` package identities.
- AC-002-2: Given the same fixture, when maintainers evaluate the golden case for `packages --prefix github.com/liza-mas/scip-search`, then the `packages` collection includes only the exact package identity whose package name starts with that literal prefix.
- AC-002-3: Given the fixture contains a symbol whose descriptor text contains `liza-mas` in a non-package component but whose package name does not start with `github.com/liza-mas/`, when maintainers evaluate the prefix golden case, then that package identity is not included solely because the prefix text appears outside the package-name component.
- AC-002-4: Given the fixture contains no package name that starts with `github.com/no-match/`, when maintainers evaluate `packages --prefix github.com/no-match/`, then the successful golden result contains an explicit empty `packages` collection.
- AC-002-5: Given a filtered package query returns more than one package identity, when maintainers compare the golden result order across repeated validation runs, then entries use the same stable package ordering as the unfiltered package fixture case.

### Depends on:
Implementation ordering:
- Story ST-001 - Cover Package Listing and De-duplication
- `CAP-002-02-package-prefix-filtering.md` - Prefix filtering and empty-result behavior must be defined before fixture golden cases can validate them.

Run time coupling:
- Interface I-001-002 - Package discovery query contract

### Out of Scope
- Filtering by scheme, package manager, package version, descriptor, symbol name, document path, regex, glob, fuzzy matching, semantic matching, or case folding.
- Suggestions for nearby package names or alternative prefixes.
- Shared runtime failures for malformed command syntax, missing index paths, unreadable indexes, or malformed SCIP indexes.

### Assumptions
- **ASM-002-1**: Prefix fixture cases use syntactically valid prefix values and leave malformed flag validation to shared CLI stories. - *Why*: CAP-003 owns successful query-specific fixture coverage, while shared invocation errors are out of scope. - Confidence: HIGH

### Open Questions
- None.
