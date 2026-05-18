# User Stories: Symbol Query Fixtures

Status: review

## Goal
Maintainers can validate `scip-search symbols --name <name>` against a small deterministic SCIP fixture matrix that proves literal partial matching, ambiguous names, empty successful results, one-line output, JSON package grouping, exact full symbol reconstruction, and stable ordering.

## Parent Epic
`specs/epics/readme/20260517-134857-epic-planning-3.md` - Capability CAP-003, "Validate discovery queries with fixtures"

## Context
Symbol discovery behavior is defined by CAP-001. This document defines only the query-specific fixture and golden-case expectations that keep that behavior stable. Shared command invocation, index loading, stdout/stderr behavior, and raw traversal fixture construction remain owned by sibling epics.

## Personas
- **CLI Maintainer**: a Go developer maintaining `scip-search`, needing bounded query fixtures that catch regressions in symbol discovery behavior.
- **Automation Agent**: an AI or script-driven caller running `scip-search` in a terminal or sandbox, needing deterministic symbol results and exact full SCIP symbol strings for follow-up commands.

## General information

Applies to: symbol discovery fixture coverage for successful `symbols --name` queries.

### References
- goal spec: `README.md#scip-symbol-format` - Defines full SCIP symbol structure and partial-name discovery requirements.
- goal spec: `README.md#what-is-scip-search` - Documents `scip-search symbols --index <index-path> --name <name>` and symbol output modes.
- parent epic: `specs/epics/readme/20260517-134857-epic-planning-3.md#general-information` - Requires deterministic ordering, explicit empty collections, and small query-specific fixtures.
- parent epic: `specs/epics/readme/20260517-134857-epic-planning-3.md#capability-cap-003---validate-discovery-queries-with-fixtures` - Defines discovery-query fixture and golden JSON validation scope.
- consistency check: `specs/stories/readme/20260517-134857-epic-planning-3/cap-001-symbol-name-discovery.md` - Defines symbol matching, payload fields, and existing symbol-query fixture expectations that this document must preserve.
- consistency check: `specs/stories/readme/20260517-134857-epic-planning-3/CAP-002-01-package-inventory.md`, `CAP-002-02-package-prefix-filtering.md`, and `CAP-002-03-package-result-json-shape.md` - Confirms package query behavior is sibling scope.

### Non-Functional Requirements
- NFR-000-1: Symbol fixture data must be deterministic and small enough to run in normal CLI validation.
- NFR-000-2: Fixture-backed symbol golden cases must preserve exact package keys and descriptors from the fixture so full SCIP symbol strings can be reconstructed with no normalization.
- NFR-000-3: Symbol fixture validation must exercise the same SCIP loading and traversal path used by the `symbols` command.
- NFR-000-4: Symbol fixture stories must not redefine shared runtime failures, raw SCIP traversal construction, package query behavior, reference behavior, or implementation behavior.

### Related External Components
- Component C-001 - SCIP symbol identifiers: Human-readable symbol strings containing scheme, package manager, package name, package version, and descriptors.
- Component C-002 - SCIP traversal view: The shared traversal input that exposes symbol inventories and definition context to query behavior.
- Component C-003 - Query fixture set: Deterministic SCIP test data, expected one-line cases, and expected JSON cases used by maintainers to validate discovery queries.

### Interfaces
- I-001-001 - Symbol discovery query contract (Interface 001 of Component C-001): The default `symbols --name` query accepts literal partial text and returns one grep-style line per result. `symbols --name --nested-json` returns a successful JSON payload with a `packages` collection of package identities containing nested matched symbol descriptors.

### Out of Scope
- Package discovery fixture cases, package JSON shape, package prefix filtering, and package de-duplication.
- Shared command routing, `--index` handling, stdout/stderr stream rules, exit status taxonomy, invalid invocation, unreadable index, and malformed SCIP errors.
- Raw SCIP fixture generation mechanics, traversal fixture coverage for documents, occurrences, relationships, ranges, and hover metadata.
- `references --symbol`, `implementations --symbol`, exact-symbol missing behavior, reference occurrences, implementation relationships, and location/range JSON for those commands.
- Large real-world fixtures, performance benchmarks, external indexer installation, ctags fallback fixtures, fuzzy search, regex search, semantic search, and cross-index search.

### Assumptions
- **ASM-000-1**: The symbol query fixture may use artificial but valid SCIP symbols rather than a fixture generated from a real repository. - *Why*: CAP-003 excludes external indexer installation and requires deterministic fixtures, while epic-planning-2 owns raw traversal fixture mechanics. - Confidence: HIGH
- **ASM-000-2**: Symbol fixture golden cases assert observable result fields from CAP-001 instead of asserting internal traversal data structures. - *Why*: The maintainer-visible contract is the command result payload, not the internal query implementation. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Cover Literal and Ambiguous Symbol Matches

### References
- parent epic: `specs/epics/readme/20260517-134857-epic-planning-3.md#capability-cap-003---validate-discovery-queries-with-fixtures` - Requires exact-looking partial matches, partial substring matches, overlapping names, and stable output ordering.
- consistency check: `specs/stories/readme/20260517-134857-epic-planning-3/cap-001-symbol-name-discovery.md#story-st-001---match-partial-symbol-names-deterministically` - Defines literal partial matching and successful ambiguity.
- consistency check: `specs/stories/readme/20260517-134857-epic-planning-3/cap-001-symbol-name-discovery.md#story-st-002---return-reconstructable-scip-symbols-with-match-context` - Defines exact full symbol reconstruction and match context fields.

### User Story
**As a** CLI Maintainer, **I want to** validate symbol discovery with deterministic fixture symbols that produce exact-looking, partial, and ambiguous matches, **so that** future changes cannot silently alter `symbols --name` matching or ordering.

### Acceptance Criteria
- AC-001-1: Given the symbol fixture contains at least the full symbols `scip-go gomod github.com/liza-mas/liza . supervisor/Supervisor#`, `scip-go gomod github.com/liza-mas/liza . supervisor/SupervisorConfig#`, `scip-go gomod github.com/liza-mas/liza . supervisor/Run().`, and `scip-go gomod github.com/liza-mas/liza . agent/SupervisorAgent#`, when maintainers run the fixture-backed symbol query cases, then the fixture provides deterministic data for exact-looking partial, substring, and overlapping-name matches.
- AC-001-2: Given the query case `symbols --name Supervisor`, when maintainers evaluate the default result, then the successful one-line output includes every fixture symbol descriptor whose display or descriptor text contains `Supervisor`, including multiple matches rather than an ambiguity failure.
- AC-001-3: Given the query case `symbols --name Run`, when maintainers evaluate the default result, then the successful one-line output includes the matching `supervisor/Run().` descriptor under its package identity and does not include non-matching `Supervisor` fixture symbols solely because they share a package.
- AC-001-4: Given any non-empty symbol fixture result, when maintainers inspect result entries, then each one-line entry asserts the exact descriptor, package key, match context, and definition location prefix required by the CAP-001 symbol payload contract, and JSON mode cases assert package identity fields.
- AC-001-5: Given a symbol query produces more than one match, when maintainers compare the golden result order across repeated validation runs, then reconstructed full symbols appear in stable ascending order by the observable full symbol value.

### Depends on:
Implementation ordering:
- `cap-001-symbol-name-discovery.md` - Symbol matching and symbol payload contracts must be defined before fixture golden cases can validate them.

Run time coupling:
- Interface I-001-001 - Symbol discovery query contract

### Out of Scope
- Choosing a best symbol candidate from ambiguous matches.
- Adding relevance scoring, fuzzy matching, regex matching, semantic matching, or case-folded matching.
- Validating package discovery, references, implementations, raw traversal fixture construction, or shared runtime failures.

### Assumptions
- **ASM-001-1**: The fixture's exact-looking query can use a descriptor name such as `Supervisor` without requiring exact-only search semantics. - *Why*: CAP-001 defines `--name` as partial literal matching, so a query equal to one descriptor still returns all literal matches. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-002 - Cover Empty Successful Symbol Results

### References
- parent epic: `specs/epics/readme/20260517-134857-epic-planning-3.md#general-information` - Requires successful no-match discovery queries to return explicit empty collections.
- parent epic: `specs/epics/readme/20260517-134857-epic-planning-3.md#capability-cap-003---validate-discovery-queries-with-fixtures` - Requires successful empty result fixture cases.
- consistency check: `specs/stories/readme/20260517-134857-epic-planning-3/cap-001-symbol-name-discovery.md#story-st-003---validate-symbol-query-fixtures` - Requires symbol-query validation of unmatched partial names.

### User Story
**As a** CLI Maintainer, **I want to** validate no-match symbol discovery with a deterministic golden case, **so that** successful absence of matching symbols remains distinguishable from invocation or index-loading failures.

### Acceptance Criteria
- AC-002-1: Given the symbol fixture contains no symbol whose display or descriptor text contains `DoesNotExist`, when maintainers run the query case `symbols --name DoesNotExist`, then the default successful result is empty stdout and the `--nested-json` golden result is a successful payload with an empty `packages` collection.
- AC-002-2: Given the no-match symbol query case is evaluated, when maintainers inspect the golden result, then it does not assert stderr diagnostics, nonzero process status, shared runtime error fields, or suggestions for nearby symbols.
- AC-002-3: Given the same no-match symbol query case is evaluated repeatedly against the same fixture, when maintainers compare the default output and golden JSON, then empty stdout and the empty `packages` collection remain stable and explicit.

### Depends on:
Implementation ordering:
- Story ST-001 - Cover Literal and Ambiguous Symbol Matches

Run time coupling:
- Interface I-001-001 - Symbol discovery query contract

### Out of Scope
- Shared runtime failure validation for missing indexes, unreadable indexes, malformed indexes, missing flags, or unsupported commands.
- Suggesting alternate symbol names or fallback search modes.

### Assumptions
- **ASM-002-1**: The no-match symbol JSON golden case uses the same top-level `packages` collection as non-empty `--nested-json` symbol results. - *Why*: The parent epic requires explicit empty result collections for successful no-match JSON cases, and CAP-001 defines one-line empty output separately as empty stdout. - Confidence: HIGH

### Open Questions
- None.
