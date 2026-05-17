# Code Plan: Package Discovery Query

Status: draft

## Source Context

Based on:
- `README.md#scip-symbol-format`
- `README.md#what-is-scip-search`
- `specs/epics/readme/20260517-134857-epic-planning-3.md`
- `specs/stories/readme/20260517-134857-epic-planning-3/CAP-002-01-package-inventory.md`
- `specs/stories/readme/20260517-134857-epic-planning-3/CAP-002-02-package-prefix-filtering.md`
- `specs/stories/readme/20260517-134857-epic-planning-3/CAP-002-03-package-result-json-shape.md`
- `specs/stories/readme/20260517-134857-epic-planning-3/CAP-003-02-package-query-fixtures.md`
- `specs/stories/readme/20260517-134857-epic-planning-3/CAP-003-03-golden-json-validation.md`
- Architecture reference `specs/arch-plan/readme/20260517-170650-epic-planning-3-architecture.md` read from merge commit `9d8fb6318ef527b937d651e29a6b1a44a62953bb` because the file is referenced by blackboard state but absent from this worktree checkout.
- Prior plan `specs/plans/readme/20260517-184518-epic-planning-3-architecture-code-planning-0.md`
- Prior plan `specs/plans/readme/20260517-190917-epic-planning-3-architecture-code-planning-1-replacement.md`
- Active dependency task `epic-planning-3-architecture-code-planning-1-replacement-coding-0`

## Planning Boundary

This plan covers only successful `packages` and `packages --prefix <prefix>` discovery behavior over an already-loaded shared traversal view.

Out of scope: CLI command routing, shared `--index` handling, shared index loading, stdout/stderr/status behavior, runtime failures, raw SCIP protobuf parsing, traversal construction, symbol discovery, reference and implementation queries, source-file reads, package registry lookups, dependency graph traversal, version resolution, descriptor filters, regex, glob, fuzzy, semantic, case-folded, package-manager, package-version, scheme, document-path, or cross-index filtering.

## Architectural Direction

Add package discovery behavior in `internal/query/discovery` as pure query logic over traversal-provided symbol facts and the `internal/scipmodel` identity helpers. The query should collect package identities from full SCIP symbols exposed by traversal, derive exact identity fields through `internal/scipmodel`, de-duplicate by full package identity using `packageKey`, optionally filter only by `packageName` with literal case-sensitive prefix matching, sort by `packageKey`, and return a typed success payload with an explicit `packages` collection.

The query package must not own process streams, CLI parsing, index loading, SCIP bytes, traversal construction, source-file reads, package registry access, dependency graph traversal, symbol discovery, reference discovery, or implementation discovery. If the downstream checkout lacks the shared traversal view or the `internal/scipmodel` helpers, the coder should mark blocked rather than implementing raw SCIP parsing or broadening this task.

Architecture trade-off: keep package query code colocated with symbol discovery in `internal/query/discovery` so discovery payloads share traversal and identity conventions, but keep package files separate (`packages.go`, `packages_test.go`) so the task remains one observable behavior and avoids editing symbol-query files from the prior plan.

## Planned Coding Tasks

### Task 1 - Implement package discovery query behavior

**desc:** `scip-search` users can run `packages` with or without `--prefix` against a loaded traversal view and receive a successful `packages` payload containing de-duplicated package identities filtered only by literal package-name prefix when provided.

**done_when:** Unit and fixture-backed query tests in `internal/query/discovery` pass for unfiltered `packages`, `packages --prefix github.com/liza-mas/`, `packages --prefix github.com/liza-mas/scip-search`, and `packages --prefix github.com/no-match/`, asserting the top-level `packages` collection is present, unfiltered results include every distinct full package identity once, duplicate symbols from the same package identity do not duplicate package entries, entries are sorted ascending by `packageKey`, each non-empty entry exposes exact `scheme`, `packageManager`, `packageName`, `packageVersion`, and `packageKey`, prefix filtering is literal and case-sensitive over `packageName` only, prefix text in scheme, package manager, package version, descriptor, symbol name, or document path does not cause a match, no-match prefixes return an explicit empty `packages` collection, and no CLI routing, `--index` handling, shared stdout/stderr failure behavior, raw SCIP parsing, traversal construction, symbol discovery, reference, implementation, source-file read, registry, dependency graph, version resolution, regex, glob, fuzzy, semantic, or cross-index behavior is introduced.

**scope:** In scope: `internal/query/discovery` package query implementation, successful package result payload value types, package candidate extraction from traversal-provided full SCIP symbol facts, package identity fields supplied by `internal/scipmodel`, de-duplication by full package identity, stable ascending sort by `packageKey`, optional literal case-sensitive prefix filtering over `packageName` only, and colocated tests using deterministic traversal fixtures or fakes. Out of scope: CLI command routing, shared `--index` handling, shared index loading, stdout/stderr/status behavior, raw SCIP protobuf parsing or traversal construction, symbol discovery, reference and implementation queries, source-file reads, package registry or dependency graph behavior, version resolution beyond the SCIP package version component, descriptor, scheme, package-manager, package-version, document-path, regex, glob, fuzzy, semantic, case-folded, or cross-index filtering, and final e2e/golden coverage owned by `epic-planning-3-architecture-code-planning-3-replacement`.

**spec_ref:** README.md#scip-symbol-format

**Planned files:**
- `internal/query/discovery/packages.go`
- `internal/query/discovery/packages_test.go`

**Implementation notes:**
- Consume traversal-provided full SCIP symbol facts through the shared traversal interfaces; do not inspect raw SCIP protobufs, index paths, or source files in this package.
- Use `internal/scipmodel.ParseIdentity` to preserve exact identity components and `Identity.PackageKey()` to build the comparison and output key.
- Include symbols from whichever traversal symbol inventories represent indexed package-bearing symbols, including external symbols if the traversal interface exposes them as package-bearing symbol facts.
- Skip or surface parse failures consistently with the traversal/discovery package's existing conventions if malformed symbol facts appear; do not synthesize package identities from partial symbols.
- De-duplicate by `packageKey`, not by `packageName` alone, so packages that differ by scheme, package manager, package version, or package name remain distinct.
- Apply the optional prefix after candidate identity derivation and de-duplication. Filtering before de-duplication is acceptable only if tests prove the same visible output and the filter is still applied to `packageName` only.
- Treat prefix matching as `strings.HasPrefix(identity.PackageName, prefix)`. Prefix characters such as `.`, `/`, `*`, `[`, `]`, `(`, `)`, `#`, and `-` must remain ordinary literal characters.
- Return a payload object with `packages` set to an empty slice for successful no-match cases; do not return nil payloads, errors, suggestions, stderr diagnostics, or alternate field names.
- Sort the final `packages` slice by exact `packageKey` ascending after de-duplication and filtering.
- Validation command for the task: `go test ./internal/query/discovery`.

## Dependency Plan

Task 1 depends on the discovery identity model implementation from `epic-planning-3-architecture-code-planning-0-coding-0-replacement`.

Task 1 depends on `epic-planning-3-architecture-code-planning-1-replacement-coding-0` because the architecture Scope 3 package query plan depends on Scope 2 symbol query, and both scopes are expected to create or extend `internal/query/discovery` package structure and successful payload conventions. This dependency serializes likely shared package setup while keeping package behavior in separate package-specific files.

Task 1 also requires the shared traversal view planned by `epic-planning-2-architecture-code-planning-0`. Concrete traversal coding task IDs are not available in this planning context; if the downstream coder's checkout does not contain traversal interfaces exposing package-bearing symbol facts, the coder should mark blocked on the traversal foundation rather than implementing raw SCIP parsing or traversal construction here.

The separate discovery fixture and golden JSON plan `epic-planning-3-architecture-code-planning-3-replacement` owns final e2e/golden coverage after symbol and package query implementations exist.

## Shared-File Audit

Task 1 is the only planned task in this plan and owns all files listed in its scope. No sibling `depends_on` edge is required inside this output array.

Task 1 must not edit CLI runtime files, shared index-loading files, traversal construction files, `internal/scipmodel` identity helper files, symbol discovery files, reference or implementation query files, README files, or final golden JSON fixture files. Those are dependency or sibling scopes.

Cross-plan shared-file risk is handled by `task_depends_on` on `epic-planning-3-architecture-code-planning-1-replacement-coding-0`, because that symbol discovery implementation owns earlier `internal/query/discovery` package structure and this package task should extend it after it lands.

## Test Impact

Task 1 must add colocated tests in `internal/query/discovery/packages_test.go` for all-package listing, full package identity de-duplication, distinct identities that differ by scheme, package manager, package name, or version, stable `packageKey` ordering, literal case-sensitive prefix matching, exclusion when prefix text appears outside `packageName`, exact payload fields, and explicit empty `packages` results.

End-to-end/golden validation through the normal CLI loading path is covered by sibling task `epic-planning-3-architecture-code-planning-3-replacement`, because this plan excludes CLI routing, `--index`, stdout/stderr behavior, and final golden files.

## Doc Impact

No user-facing documentation update is required in this plan. README already documents the `packages --index <index-path> [--prefix <prefix>]` command and SCIP package identity format; this task implements that existing documented behavior without changing public contract text. Query-specific specs are already present in the epic and story documents listed above.

## Spec Compliance Matrix

| # | Requirement | Source | Task(s) | Status |
|---|-------------|--------|---------|--------|
| 1 | Unfiltered `packages` returns every distinct package identity discoverable from indexed SCIP symbols. | `CAP-002-01-package-inventory.md` AC-001-1; architecture Scope 3 | Task 1 | Covered |
| 2 | Multiple symbols with the same scheme, package manager, package name, and package version produce one package entry. | `CAP-002-01-package-inventory.md` AC-001-2; CAP-003-02 AC-001-2 | Task 1 | Covered |
| 3 | Symbols that differ by scheme, package manager, package name, or package version are distinct package identities. | `CAP-002-01-package-inventory.md` AC-001-3; CAP-003-02 AC-001-3 | Task 1 | Covered |
| 4 | Package result ordering is deterministic and stable across repeated runs. | `CAP-002-01-package-inventory.md` NFR-000-1 and AC-001-4; parent epic NFR-000-3 | Task 1 | Covered |
| 5 | Package entries are sorted ascending by observable `packageKey`. | `CAP-002-01-package-inventory.md` ASM-000-2; CAP-003-02 AC-001-5; assigned done_when | Task 1 | Covered |
| 6 | `packages --prefix <prefix>` includes package identities whose `packageName` begins with the requested literal prefix. | `CAP-002-02-package-prefix-filtering.md` AC-001-1; architecture Scope 3 | Task 1 | Covered |
| 7 | Prefix filtering excludes package identities whose `packageName` does not begin with the requested prefix. | `CAP-002-02-package-prefix-filtering.md` AC-001-2 | Task 1 | Covered |
| 8 | Prefix text outside `packageName` does not cause a match. | `CAP-002-02-package-prefix-filtering.md` AC-001-3; assigned done_when | Task 1 | Covered |
| 9 | Prefix filtering is literal and case-sensitive, not regex, glob, fuzzy, semantic, or case-folded matching. | `CAP-002-02-package-prefix-filtering.md` NFR-000-1, ASM-000-1, and ASM-001-1; parent epic ASM-000-3 | Task 1 | Covered |
| 10 | Filtered package result ordering follows the same package ordering as unfiltered package listing. | `CAP-002-02-package-prefix-filtering.md` AC-001-4; CAP-003-02 AC-002-5 | Task 1 | Covered |
| 11 | No-match package prefixes return a successful payload with an explicit empty `packages` collection. | `CAP-002-02-package-prefix-filtering.md` AC-002-1 through AC-002-3; parent epic NFR-000-3 | Task 1 | Covered |
| 12 | Successful package payloads expose a top-level `packages` collection for non-empty and empty results. | `CAP-002-03-package-result-json-shape.md` AC-001-1 and AC-001-5; assigned done_when | Task 1 | Covered |
| 13 | Each package entry exposes exact `scheme`, `packageManager`, `packageName`, `packageVersion`, and `packageKey` fields. | `CAP-002-03-package-result-json-shape.md` AC-001-2 and AC-001-3; assigned done_when | Task 1 | Covered |
| 14 | `packageKey` identifies full package identity and is derived from scheme, package manager, package name, and package version without descriptors. | `CAP-002-03-package-result-json-shape.md` ASM-000-3 and AC-001-4; prior plan `20260517-184518-epic-planning-3-architecture-code-planning-0.md` | Task 1 | Covered |
| 15 | Package fixture coverage includes repeated package identities, `github.com/liza-mas/liza`, `github.com/liza-mas/scip-search`, and `github.com/sourcegraph/scip-bindings`. | `CAP-003-02-package-query-fixtures.md` AC-001-1 | Task 1 | Covered |
| 16 | Package fixture coverage includes `packages --prefix github.com/liza-mas/`, `packages --prefix github.com/liza-mas/scip-search`, and `packages --prefix github.com/no-match/`. | `CAP-003-02-package-query-fixtures.md` AC-002-1, AC-002-2, and AC-002-4; assigned done_when | Task 1 | Covered |
| 17 | Package fixture coverage proves a prefix in descriptor or other non-package-name data does not match. | `CAP-003-02-package-query-fixtures.md` AC-002-3; assigned done_when | Task 1 | Covered |
| 18 | Implementation stays over shared traversal facts and does not own CLI routing, `--index`, shared failures, raw SCIP parsing, traversal construction, symbol discovery, references, implementations, registry, dependency graph, version resolution, descriptor filters, source-file reads, or unsupported match modes. | Assigned scope; parent epic Out of Scope; architecture Constraints and Scope 3 | Task 1 | Covered |
| E2E | e2e test coverage for new behavior | Cross-cutting | Sibling task `epic-planning-3-architecture-code-planning-3-replacement`: final discovery fixture and golden JSON validation through normal runtime path | Covered |
| DOC | Documentation updates for changed behavior | Cross-cutting | N/A: README already documents the package command and SCIP package identity format; this plan implements existing user-facing specs without changing the public contract text | N/A |

## Pre-Submit Validation Checklist

- Re-read this plan and verify the output JSON fields are character-identical to Task 1 `desc`, `done_when`, `scope`, and `spec_ref`.
- Run `jq . specs/plans/readme/20260517-191918-epic-planning-3-architecture-code-planning-2-replacement-output.json`.
- Search this plan for `Task 1`, `epic-planning-3-architecture-code-planning-1-replacement-coding-0`, `epic-planning-3-architecture-code-planning-0-coding-0-replacement`, and `epic-planning-3-architecture-code-planning-3-replacement` references and confirm the referenced responsibility, dependency, or exclusion is explicitly stated.
- Run pre-commit on the plan and output JSON files.
- Commit only the plan and output JSON artifacts.
