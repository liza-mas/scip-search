# User Stories: Implementation Relationship Selection

Status: review

## Goal
`implementations --symbol` returns only implementation symbols from incoming SCIP implementation relationships that point at the caller's exact queried symbol.

## Parent Epic
specs/epics/readme/20260517-141006-epic-planning-4.md - Capability CAP-003

## Context
The shared exact-symbol stories define how `implementations --symbol` accepts and preserves the caller-provided symbol. This document defines the command-specific successful selection rule: implementation results come from implementer-side SCIP relationships whose implementation target is the queried exact symbol.

## Personas
- **Automation Agent**: An AI or script-driven caller running `scip-search` in a terminal or sandbox, needing exact implementation results it can use for later code-navigation or editing workflows.
- **CLI Maintainer**: A Go developer maintaining `scip-search`, needing bounded implementation semantics that stay aligned with SCIP relationship data rather than inferred hierarchy analysis.

## General information

Applies to: successful implementation result selection for exact `implementations --symbol` queries.

### References
- goal spec: README.md#what-is-scip-search - Documents the `scip-search implementations --index <index-path> --symbol <scip-symbol>` command form and structured JSON output.
- goal spec: README.md#complementary-existing-tool - Frames `implementations` as the query that answers what implements a symbol.
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#personas - Defines the Automation Agent and CLI Maintainer personas for symbol-based query stories.
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#general-information - Defines shared NFRs, interfaces, out-of-scope boundaries, and implementation relationship assumptions for this epic.
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#capability-cap-003---return-implementation-relationship-results - Requires incoming SCIP implementation relationships, implementer symbols, exclusion of outgoing implemented targets, definition locations when available, and stable successful payloads.
- official source: https://raw.githubusercontent.com/scip-code/scip/main/scip.proto - Defines SCIP symbol relationships and `Relationship.is_implementation`.
- dependency story: specs/stories/readme/epic-planning-4/CAP-001-01-exact-symbol-input.md - Defines literal exact `implementations --symbol` input and top-level queried-symbol preservation.
- dependency story: specs/stories/readme/epic-planning-4/CAP-001-02-missing-symbol-results.md - Defines successful empty `implementations` payloads for missing exact symbols and no-match outcomes.
- consistency check: specs/stories/readme/20260517-134857-epic-planning-3/CAP-002-03-package-result-json-shape.md - Provides precedent for query-specific JSON fields under the shared success contract.

### Non-Functional Requirements
- NFR-000-1: Implementation selection must be deterministic for the same loaded SCIP index and exact queried symbol.
- NFR-000-2: Implementation selection must preserve traversal-provided SCIP symbol strings exactly and must not rewrite or normalize implementation symbols.
- NFR-000-3: Implementation selection must not synthesize relationships from names, packages, occurrence proximity, source text, type-definition traversal, semantic similarity, or package dependency data.
- NFR-000-4: These stories must not redefine shared command routing, `--index` loading, stdout/stderr stream rules, process status, shared runtime failures, symbol discovery, package discovery, reference occurrence selection, traversal view construction, or fixture coverage.

### Related External Components
- Component C-001 - SCIP symbol identifiers: Exact full symbol strings supplied through `--symbol` and returned in implementation result payloads.
- Component C-003 - SCIP relationship data: Symbol relationship edges used to select incoming implementation results.
- Component C-004 - Calling process environment: The shell, script, or agent invoking `scip-search` and parsing successful JSON stdout under the shared runtime contract.

### Interfaces
- I-001-002 - Implementation query contract (Interface 002 of Component C-001): The `implementations --symbol` query accepts an exact full SCIP symbol and returns implementation symbols derived from incoming SCIP implementation relationships.
- I-003-001 - Implementation relationship source (Interface 001 of Component C-003): Traversal data exposes SCIP implementation relationships including the implementer symbol and the implemented target symbol.
- I-004-001 - CLI process contract (Interface 001 of Component C-004): The shared process contract that exposes successful query payloads as parseable JSON on stdout.

### Out of Scope
- Exact-symbol input parsing, partial/fuzzy/regex/semantic symbol matching, and symbol discovery.
- Shared invocation errors, index-loading errors, stderr diagnostics, stdout stream purity, and process status taxonomy.
- Reference occurrence results, type-definition traversal, full hierarchy graphs, package dependency graphs, semantic similarity, or source-file reads.
- Synthesizing implementation relationships that are absent from the SCIP index.
- Defining implementation definition-location field names, range field names, golden fixtures, or fixture generation.

### Assumptions
- **ASM-000-1**: An implementation relationship result can be described to callers as incoming when the relationship belongs to an implementation symbol and its implemented target symbol equals the queried exact symbol. - *Why*: CAP-003 defines implementer-side relationships pointing at the queried symbol as the basis for "what implements this symbol?" results. - Confidence: HIGH
- **ASM-000-2**: The implementation query result uses the top-level `implementations` collection defined by CAP-001 missing-symbol stories for both empty and non-empty successful results. - *Why*: CAP-001 established the command-specific collection name, and CAP-003 fills in the non-empty entry semantics. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Select Incoming Implementation Relationships

### References
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#capability-cap-003---return-implementation-relationship-results - Requires implementer symbols whose SCIP implementation relationships point at the queried exact symbol.
- official source: https://raw.githubusercontent.com/scip-code/scip/main/scip.proto - Defines SCIP symbol relationships and `Relationship.is_implementation`.
- dependency story: specs/stories/readme/epic-planning-4/CAP-001-01-exact-symbol-input.md - Defines the queried exact symbol used as the traversal starting point.

### User Story
**As an** Automation Agent asking what implements an exact SCIP symbol, **I want to** receive implementation results selected from incoming SCIP implementation relationships, **so that** I can navigate to concrete implementers without deriving type relationships myself.

### Acceptance Criteria
- AC-001-1: Given the loaded index contains an implementation symbol with a SCIP relationship marked as an implementation of the queried exact symbol, when the caller runs `implementations --symbol <scip-symbol>` for that target symbol, then the successful payload includes an entry for that implementation symbol in the `implementations` collection.
- AC-001-2: Given multiple implementation symbols each have SCIP implementation relationships that point at the queried exact symbol, when the caller runs `implementations --symbol` for that symbol, then the successful payload includes one implementation entry for each distinct implementation symbol.
- AC-001-3: Given an implementation relationship points at a symbol other than the queried exact symbol, when the caller runs `implementations --symbol` for the queried symbol, then that relationship does not produce an implementation entry.
- AC-001-4: Given an indexed symbol appears in occurrences or packages but has no incoming SCIP implementation relationship pointing at the queried exact symbol, when the caller runs `implementations --symbol`, then that symbol is not returned as an implementation result.
- AC-001-5: Given an implementation entry is returned, when the caller inspects it, then the entry identifies the incoming implementation relationship basis that made the implementation symbol a result.

### Depends on:
Implementation ordering:
- Story document CAP-001-01-exact-symbol-input.md - The implementation query must already treat `--symbol` as a literal exact symbol and preserve it in successful payloads.

Run time coupling:
- I-001-002 - Implementation query contract
- I-003-001 - Implementation relationship source
- I-004-001 - CLI process contract

### Out of Scope
- Resolving definition locations or ranges for the returned implementation symbols.
- Defining exact JSON field names for the relationship basis.
- Returning reference occurrences, type definitions, package dependencies, semantic matches, or symbols discovered by partial-name matching.

### Assumptions
- **ASM-001-1**: If duplicate relationship records identify the same implementation symbol and queried target, the caller observes one implementation result entry for that implementation symbol. - *Why*: CAP-003 requires implementation symbols as results, while stable automation payloads should not expose duplicate relationship storage as duplicate user-facing implementers. - Confidence: MEDIUM

### Open Questions
- None.

---

## Story ST-002 - Exclude Outgoing and Synthesized Implementation Results

### References
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#capability-cap-003---return-implementation-relationship-results - Excludes outgoing implemented targets from the queried symbol as implementation results.
- parent epic: specs/epics/readme/20260517-141006-epic-planning-4.md#out-of-scope - Excludes synthesized relationships, type hierarchy graphs, semantic similarity, source-file reads, and adjacent discovery capabilities.

### User Story
**As a** CLI Maintainer defining implementation-query boundaries, **I want to** exclude outgoing and synthesized relationships from `implementations --symbol` results, **so that** the command answers "what implements this symbol?" without expanding into hierarchy analysis or discovery behavior.

### Acceptance Criteria
- AC-002-1: Given the queried symbol has an outgoing SCIP implementation relationship to a different implemented target, when the caller runs `implementations --symbol` for the queried symbol, then that outgoing target is not returned as an implementation result.
- AC-002-2: Given the queried symbol is itself an implementer of another symbol, when the caller runs `implementations --symbol` for the queried symbol, then the payload includes only symbols that implement the queried symbol and not the symbols that the queried symbol implements.
- AC-002-3: Given two symbols have similar names, shared packages, matching descriptors, inheritance-like names, or nearby source locations but no incoming SCIP implementation relationship pointing at the queried exact symbol, when the caller runs `implementations --symbol`, then those symbols are not returned as implementation results.
- AC-002-4: Given the index contains reference relationships or reference occurrences involving the queried symbol, when the caller runs `implementations --symbol`, then those references do not produce implementation entries unless a SCIP implementation relationship also qualifies the implementation symbol.

### Depends on:
Implementation ordering:
- Story ST-001 - Select Incoming Implementation Relationships

Run time coupling:
- I-001-002 - Implementation query contract
- I-003-001 - Implementation relationship source

### Out of Scope
- Providing a separate "what does this implementation implement?" query.
- Producing full inheritance graphs, reverse dependency graphs, reference result lists, type-definition traversals, or semantic-similarity rankings.
- Synthesizing missing SCIP relationships from source files or language-specific analysis.

### Assumptions
- **ASM-002-1**: Outgoing implementation relationships from the queried symbol may exist in the loaded SCIP data but are explanatory context, not implementation results for this command. - *Why*: CAP-003 explicitly distinguishes incoming implementers from outgoing implemented targets. - Confidence: HIGH

### Open Questions
- None.
