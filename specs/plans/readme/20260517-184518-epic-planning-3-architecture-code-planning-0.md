# Code Plan: SCIP Discovery Identity Model

Status: draft

## Source Context

Based on:
- `README.md#scip-symbol-format`
- `specs/epics/readme/20260517-134857-epic-planning-3.md`
- `specs/stories/readme/20260517-134857-epic-planning-3/cap-001-symbol-name-discovery.md`
- `specs/stories/readme/20260517-134857-epic-planning-3/CAP-002-01-package-inventory.md`
- `specs/stories/readme/20260517-134857-epic-planning-3/CAP-002-02-package-prefix-filtering.md`
- `specs/stories/readme/20260517-134857-epic-planning-3/CAP-002-03-package-result-json-shape.md`
- Architecture reference `specs/arch-plan/readme/20260517-170650-epic-planning-3-architecture.md#scope-1-scip-discovery-identity-model` read from merge commit `9d8fb6318ef527b937d651e29a6b1a44a62953bb` because the file is referenced by blackboard state but absent from this worktree checkout.

## Planning Boundary

This plan covers only reusable SCIP symbol identity extraction and match-text/package-key helpers for discovery queries.

Out of scope: CLI command routing, shared `--index` handling, shared index loading, traversal construction, final JSON emission, symbol result selection, package result filtering, reference or implementation behavior, source-file reads, package registry lookups, dependency graph traversal, and fuzzy, regex, semantic, glob, case-folded, or cross-index matching.

## Architectural Direction

Create a small `internal/scipmodel` package that owns symbol-format interpretation for discovery code. The package should be pure Go logic over already-loaded symbol strings and optional traversal display names. It should not import CLI, runtime, traversal, JSON output, or official SCIP protobuf types.

The parser should split a full SCIP symbol into the first four space-delimited identity components and the remaining descriptor text:

```text
<scheme> <package-manager> <package-name> <package-version> <descriptors>
```

The original full symbol string must be stored and returned unchanged. `packageKey` should be the exact package prefix made from `scheme`, `packageManager`, `packageName`, and `packageVersion`, without descriptors. Match text should prefer a non-empty traversal display name when provided and fall back to descriptor text from the symbol string.

## Planned Coding Tasks

### Task 1 - Implement SCIP discovery identity helpers

**desc:** `scip-search` maintainers can use a shared Go SCIP discovery identity model that parses fixture full SCIP symbols into exact symbol, package identity fields, descriptor/display match text, and package key without rewriting the symbol string.

**done_when:** Unit tests in `internal/scipmodel` pass with fixture symbols including `scip-go gomod github.com/liza-mas/liza . supervisor/Supervisor#`, `scip-go gomod github.com/liza-mas/liza . supervisor/Run().`, and `scip-go gomod github.com/liza-mas/scip-search . internal/query/Search().`, asserting the exact original `symbol` string is preserved, `scheme`, `packageManager`, `packageName`, `packageVersion`, descriptor text, display-name-preferred match text, descriptor-fallback match text, and `packageKey` are derived exactly, and no CLI routing, traversal construction, query filtering, or JSON emission code is introduced.

**scope:** In scope: `internal/scipmodel` identity value types, parsing helpers, package-key helper, match-text helper, and colocated unit tests over deterministic SCIP fixture symbols. Out of scope: CLI command routing, shared index loading, traversal construction, symbol-query result selection, package-query filtering, final JSON structs or emission, source-file reads, reference or implementation lookup, package registry behavior, dependency graph behavior, and fuzzy, regex, glob, semantic, case-folded, or cross-index matching.

**spec_ref:** README.md#scip-symbol-format

**Planned files:**
- `internal/scipmodel/identity.go`
- `internal/scipmodel/identity_test.go`

**Implementation notes:**
- Keep the package independent of traversal and CLI packages so symbol and package discovery can both import it without circular dependencies.
- Treat descriptor text as the exact suffix after the fourth space-delimited component; do not normalize punctuation, casing, slashes, `#`, `().`, or other descriptor markers.
- Return an explicit parse failure for strings with fewer than five symbol components, but do not synthesize or rewrite a replacement symbol string.
- Tests should use table-driven cases for class, method, and function-like descriptors and for display-name override versus descriptor fallback.
- Validation command for the task: `go test ./internal/scipmodel`.

## Dependency Plan

Task 1 has no sibling dependencies in this plan. It is intentionally designed as the first discovery building block used by later symbol and package query tasks.

If the downstream coder's checkout does not yet contain the project Go module baseline, the coder should mark the implementation task blocked on the module baseline rather than broadening this task to own `go.mod` or CLI runtime setup.

## Shared-File Audit

Task 1 is the only planned task and owns all files listed in its scope. No `depends_on` edge is required for shared files inside this plan.

The task must not edit CLI runtime, traversal, query result, package discovery, symbol discovery, reference, implementation, README, or fixture golden files. Those are sibling scopes.

## Test Impact

Task 1 must add unit tests in `internal/scipmodel/identity_test.go` that exercise the changed helper behavior directly. No e2e test is required in this plan because no user-visible CLI command behavior is introduced.

## Doc Impact

No user-facing documentation update is required. The behavior is internal helper code for later discovery tasks; README command behavior is unchanged by this task.

## Spec Compliance Matrix

| # | Requirement | Source | Task(s) | Status |
|---|-------------|--------|---------|--------|
| 1 | Preserve the exact full SCIP symbol string so downstream symbol-based commands can receive it unchanged. | `README.md#scip-symbol-format`; parent epic NFR-000-1; architecture Scope 1 | Task 1 | Covered |
| 2 | Derive `scheme` from fixture SCIP symbols. | `README.md#scip-symbol-format`; architecture Scope 1 done_when | Task 1 | Covered |
| 3 | Derive `packageManager` from fixture SCIP symbols. | `README.md#scip-symbol-format`; architecture Scope 1 done_when | Task 1 | Covered |
| 4 | Derive `packageName` from fixture SCIP symbols. | `README.md#scip-symbol-format`; package inventory ASM-000-1 | Task 1 | Covered |
| 5 | Derive `packageVersion` from fixture SCIP symbols, preserving values such as `.` exactly. | `README.md#scip-symbol-format`; package result JSON shape AC-001-3 | Task 1 | Covered |
| 6 | Expose descriptor text from the descriptors portion of the full SCIP symbol for discovery match context. | parent epic CAP-001 description; symbol story AC-002-2; architecture Scope 1 | Task 1 | Covered |
| 7 | Expose display-name-preferred match text while retaining descriptor fallback for symbols without a display name. | symbol story ASM-001-1; architecture SCIP Identity Model key decision | Task 1 | Covered |
| 8 | Derive `packageKey` from scheme, package manager, package name, and package version only, without descriptors. | package result JSON shape ASM-000-3; architecture SCIP Identity Model key decision | Task 1 | Covered |
| 9 | Keep helper logic reusable by both symbol and package discovery without owning query matching, filtering, traversal construction, or JSON emission. | assigned scope; architecture Components and Interfaces | Task 1 | Covered |
| 10 | Validate behavior with deterministic fixture SCIP symbols. | assigned done_when; CAP-003 fixture requirements for exact symbol preservation and package identity fields | Task 1 | Covered |
| E2E | e2e test coverage for new behavior | Cross-cutting | N/A: internal helper package only; no CLI command behavior changes in this task | N/A |
| DOC | Documentation updates for changed behavior | Cross-cutting | N/A: internal helper package only; README behavior is unchanged and later query tasks own user-visible command docs if needed | N/A |

## Pre-Submit Validation Checklist

- Re-read this plan and verify the output JSON fields are character-identical to Task 1 `desc`, `done_when`, `scope`, and `spec_ref`.
- Run `jq . specs/plans/readme/20260517-184518-epic-planning-3-architecture-code-planning-0-output.json`.
- Search this plan for `Task 1` references and confirm each responsibility is stated by the task.
- Run pre-commit on the plan and output JSON files.
- Commit only the plan and output JSON artifacts.
