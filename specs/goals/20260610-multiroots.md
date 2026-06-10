# PRD: Multi-Root SCIP Index Aggregation

Status: draft

## Goal

Produce one valid SCIP index per language from multiple per-root SCIP indexes so existing `scip-search --index` queries can operate across a multi-root repository.

## Context

Some repositories contain multiple Python or TypeScript roots. Existing SCIP indexers index one root at a time, which leaves callers with several same-language indexes for one repository. `scip-search` already consumes one caller-selected SCIP protobuf index per invocation, so the aggregation capability should produce a standard SCIP index file that normalizes many input root-relative path spaces into one aggregate project-root-relative path space.

## General information

Applies to: same-language aggregation of already-generated SCIP protobuf indexes.

### References

- pairing discussion: 2026-06-10 multi-root index question - User requested one index per language across roots such as `apps/api`, `services/some-service`, `apps/web`, `infra/cdk`, and `services/some-service/web`.
- pairing discussion: 2026-06-10 project-root output question - User asked how output project-root metadata should work for aggregated multi-root indexes and whether results should be sectioned per root.
- README: `README.md:130-225` - Documents the one-shot `--index` runtime contract, official SCIP bindings, and no hidden query-time index discovery.
- source: `internal/scipindex/loader.go:42-74` - Current loader decodes one caller-selected SCIP protobuf file.
- source: `internal/traversal/view.go:22-59` - Traversal reads documents, external symbols, occurrences, and relationships from the loaded index.
- source: `internal/traversal/facts.go:5-27` - Traversal exposes metadata project root and document relative paths.
- official SCIP Go bindings: `scip.Index` and `scip.Metadata` docs - `Index` contains metadata, documents, and external symbols; metadata has one `project_root`.

### Non-Functional Requirements

- NFR-000-1: The aggregate output MUST be a standard SCIP protobuf index, not a custom wrapper format.
- NFR-000-2: Aggregation MUST be deterministic for the same ordered inputs.
- NFR-000-3: Aggregation MUST NOT parse source files, run language indexers, infer imports, or rewrite symbol identities.
- NFR-000-4: Existing query commands MUST continue to read only the caller-supplied `--index` file.
- NFR-000-5: Aggregate indexes MUST expose one aggregate `metadata.project_root`; query consumers MUST NOT need per-input-root sections to resolve document paths.
- NFR-000-6: Project-root exposure in query and graph-export outputs is in scope for this capability because aggregate-root-relative paths cannot be resolved safely without the loaded index root.

### Related External Components

- Component C-001 - Input SCIP index: A single-root SCIP protobuf file produced by a language indexer.
- Component C-002 - Aggregate SCIP index: The merged SCIP protobuf file consumed by existing `scip-search` queries.
- Component C-003 - Source root mapping: Caller-supplied association between an input index and its repository-relative root path.

### Interfaces

- I-001 - Aggregation CLI: A new non-query `scip-search aggregate-index` command.
- I-002 - Query output path metadata: Existing query commands expose the loaded SCIP index `metadata.project_root` once when output contains document paths.

#### Aggregation CLI Syntax

```bash
scip-search aggregate-index \
  --project-root /home/me/Workspace/the-repo \
  --root apps/api --index apps/api/index.scip \
  --root services/some-service --index services/some-service/index.scip \
  --root apps/web --index apps/web/index.scip \
  --out python.scip
```

`aggregate-index` is not a query command. It reads multiple caller-selected SCIP indexes and writes one aggregate SCIP protobuf index.

`aggregate-index` uses the shared process status conventions where they fit: success exits `0`, malformed command-line shape exits `2`, and an unreadable or invalid input SCIP file exits `3`. Aggregation validation failures after all inputs are readable SCIP data, such as duplicate document paths, root-mapping mismatches, mixed languages, metadata incompatibility, symbol collisions, or output path collisions, exit `4`.

### Out of Scope

- Running or installing SCIP indexers.
- Discovering roots automatically.
- Cross-language aggregation.
- Symbol rewriting to repair indexer package identity.
- Inferring cross-root references that are not already represented by matching SCIP symbols.
- Adding per-result absolute paths to query outputs.
- Changing existing query commands to accept multiple `--index` values.
- Sectioning normal query results or graph-export nodes by original input root.
- Persisting per-input source provenance inside the aggregate SCIP protobuf.
- Watch mode, daemon mode, caching, or incremental updates.

### Assumptions

- **ASM-000-1**: v1 aggregates indexes for one language at a time. - *Why*: The user asked for one index per language, and mixed-language symbol/package identity introduces separate compatibility concerns. - Confidence: HIGH
- **ASM-000-2**: Caller-supplied root mappings are required. - *Why*: SCIP metadata has one project root, while each input index's document paths are relative to its own root. - Confidence: HIGH
- **ASM-000-3**: Per-input roots are useful as provenance metadata, not as the primary path-resolution contract. - *Why*: Rewriting aggregate document paths to be relative to one aggregate project root gives clients one resolution rule for single-root and multi-root indexes. - Confidence: HIGH
- **ASM-000-4**: v1 does not preserve per-input provenance in the aggregate SCIP file. - *Why*: Standard SCIP metadata has one project root and no source-list field; encoding structured provenance into tool arguments would make command-line text a data contract. - Confidence: HIGH

### Open Questions

- None for v1.

---

## Feature FT-001 - Build an Aggregate SCIP Index

### References

- pairing discussion: 2026-06-10 multi-root index question.
- README: `README.md:221-225`.
- source: `internal/scipindex/loader.go:42-74`.

### Functional Requirements

- FR-001-1: The aggregation interface MUST accept a repository project root, an output path, and two or more input pairs where each pair has an input SCIP index path and a source root path.
- FR-001-1a: The aggregation interface MUST be exposed as `scip-search aggregate-index`.
- FR-001-1b: `aggregate-index` MUST accept `--project-root <path-or-file-uri>`, `--out <output-scip-path>`, and repeated adjacent `--root <repo-relative-root> --index <input-scip-path>` pairs.
- FR-001-1c: `aggregate-index` MUST reject an unpaired `--root` or `--index`, including a sequence where two `--root` flags or two `--index` flags appear without completing a pair.
- FR-001-1d: `--root` values MUST be slash-separated paths relative to the aggregate project root; `.` is allowed for the aggregate project root itself.
- FR-001-1e: `aggregate-index` MUST reject `--root` values that are empty strings, absolute paths, or paths that escape the aggregate project root after path cleaning.
- FR-001-1f: `aggregate-index` MUST treat trailing slashes and redundant `.` path segments in `--root` values as equivalent after path cleaning.
- FR-001-1f2: `aggregate-index` MUST reject invocations with fewer than two completed `--root <root> --index <path>` input pairs.
- FR-001-1g: `--project-root` MUST be accepted only as an absolute filesystem path or a `file://` URI.
- FR-001-1h: `--project-root` MUST be cleaned, converted to a SCIP `file://` URI for aggregate metadata and output metadata, and emitted without a trailing slash except for filesystem root.
- FR-001-1i: Relative `--project-root` values MUST be rejected.
- FR-001-2: For each input document, the aggregate document `relative_path` MUST be rewritten by joining the cleaned input source root with the input document `relative_path`, cleaning the joined path, and emitting the result as a slash-separated path relative to the aggregate project root.
- FR-001-2a: Joined document paths containing `..` segments MUST be normalized rather than rejected when the cleaned result remains inside the aggregate project root.
- FR-001-2b: `aggregate-index` MUST reject any cleaned joined document path that escapes the aggregate project root.
- FR-001-3: The aggregate metadata `project_root` MUST identify the aggregate repository root, not any individual input root.
- FR-001-3a: The aggregate MUST NOT preserve multiple path-resolution roots in document paths; it MUST convert every document path into the aggregate project-root-relative path space.
- FR-001-3b: When the aggregate `--project-root` and an input `metadata.project_root` are comparable `file://` URIs, the cleaned `--root` mapping for that input MUST equal the cleaned relative path from the aggregate project root to the input project root.
- FR-001-3c: When an input `metadata.project_root` is empty, non-file, or otherwise not comparable to the aggregate project root, `aggregate-index` MUST treat the explicit `--root` value as caller-provided mapping data.
- FR-001-4: The aggregate MUST preserve document symbols, occurrences, relationships, ranges, language fields, and position encodings without semantic rewriting.
- FR-001-5: The aggregate MUST preserve SCIP symbol strings exactly as emitted by the input indexers.
- FR-001-6: The aggregate MUST include external symbols from all inputs, deduplicated by exact SCIP symbol string.
- FR-001-7: The aggregate SCIP index MUST NOT claim or encode per-input root provenance unless a future spec defines a structured standard-SCIP-compatible storage contract.
- FR-001-8: The aggregate metadata `tool_info.name` MUST identify `scip-search aggregate-index` as the producer of the aggregate index.
- FR-001-8a: The aggregate metadata `tool_info.version` MUST identify the current `scip-search` binary version string.

### Acceptance Criteria

- AC-001-1: Given two valid same-language input indexes rooted at `apps/api` and `services/some-service`, when aggregation succeeds, then the output index contains documents whose paths are prefixed with `apps/api/` and `services/some-service/`.
- AC-001-1a: Given the command `scip-search aggregate-index --project-root /home/me/Workspace/the-repo --root apps/api --index apps/api/index.scip --root services/some-service --index services/some-service/index.scip --out python.scip`, when both input indexes are valid, then the command writes `python.scip` as a valid SCIP protobuf index.
- AC-001-1b: Given an input index with `project_root=file:///home/me/Workspace/the-repo/apps/web/src` and `document_path=App.tsx`, when it is aggregated with `--project-root /home/me/Workspace/the-repo --root apps/web/src`, then the aggregate uses `project_root=file:///home/me/Workspace/the-repo` and `document_path=apps/web/src/App.tsx`.
- AC-001-1c: Given an input index with `project_root=file:///home/me/Workspace/the-repo/apps/web/src` and `document_path=../vite.config.ts`, when it is aggregated with `--project-root /home/me/Workspace/the-repo --root apps/web/src`, then the aggregate emits `document_path=apps/web/vite.config.ts`.
- AC-001-1d: Given an input document path would clean to a path outside the aggregate project root, when aggregation runs, then it fails before writing output.
- AC-001-1e: Given an input index has `metadata.project_root=file:///home/me/Workspace/the-repo/apps/web/src`, when aggregation is invoked with `--project-root /home/me/Workspace/the-repo --root apps/web`, then it fails with a root-mapping diagnostic.
- AC-001-1f: Given `--project-root /home/me/Workspace/the-repo/`, when aggregation succeeds, then aggregate metadata and output metadata use `file:///home/me/Workspace/the-repo`.
- AC-001-1g: Given `aggregate-index` is invoked with fewer than two completed input pairs, when invocation validation runs, then it fails before reading or writing any SCIP files.
- AC-001-2: Given the aggregate output is passed to `scip-search packages --index <output>`, when the command runs, then it loads through the existing SCIP loader rather than a custom aggregate reader.
- AC-001-3: Given an input symbol string appears in an occurrence or relationship, when aggregation succeeds, then that symbol string is byte-for-byte unchanged in the output index.

---

## Feature FT-002 - Reject Ambiguous or Corrupt Aggregates

### References

- source: `internal/traversal/view.go:28-49`.
- source: `internal/traversal/facts.go:5-12` - Metadata fields used for compatibility checks.
- source: `internal/traversal/facts.go:21-27` - Document fields used for path and language-family validation.

### Functional Requirements

- FR-002-1: Aggregation MUST reject duplicate aggregate document paths.
- FR-002-2: Aggregation MUST reject source root mappings that make an output document path escape the aggregate project root.
- FR-002-2a: Aggregation MUST reject an input whose `metadata.project_root` is a comparable `file://` URI and whose cleaned relative path from the aggregate project root does not equal the cleaned `--root` value, with a root-mapping diagnostic.
- FR-002-3: Aggregation MUST reject mixed non-empty SCIP indexer families across all inputs.
- FR-002-3a: Aggregation MUST reject incompatible input metadata when one aggregate metadata value cannot represent all inputs.
- FR-002-3b: Input `metadata.version` values MUST match when specified.
- FR-002-3c: Input `metadata.text_document_encoding` values MUST match when specified.
- FR-002-3d: Input `metadata.project_root` values are not required to match because the aggregate metadata `project_root` is supplied by `--project-root`.
- FR-002-3e: Input `metadata.tool_info` values are not required to match because the aggregate metadata `tool_info` identifies `scip-search aggregate-index` as the producer.
- FR-002-3f: The input SCIP indexer family is derived from non-empty `metadata.tool_info.name` values, normalized to the SCIP-producing indexer family such as `scip-typescript`, `scip-python`, or `scip-go`.
- FR-002-3g: Mixed non-empty `Document.language` values within one indexer family MUST NOT be rejected solely for differing labels.
- FR-002-4: Aggregation MUST reject duplicate local definition symbols that resolve to different aggregate document paths.
- FR-002-5: Aggregation MUST reject conflicting duplicate external symbol records for the same SCIP symbol string.
- FR-002-5a: Two duplicate external symbol records conflict when any field other than the SCIP symbol string differs after deterministic protobuf comparison. Identical duplicate records are deduplicated.
- FR-002-6: On failure, aggregation MUST leave no partial output at the requested output path.
- FR-002-7: Aggregation MUST reject an `--out` path that resolves to the same path as any input `--index` after path normalization.

### Acceptance Criteria

- AC-002-1: Given two inputs both contain `src/main.ts` and both are mapped to the same source root, when aggregation runs, then it fails with a duplicate document path diagnostic.
- AC-002-2: Given an input would produce `../outside.py` or an absolute aggregate-relative document path, when aggregation runs, then it fails before writing output.
- AC-002-3: Given two inputs define the same SCIP symbol in different aggregate document paths, when aggregation runs, then it fails with a symbol collision diagnostic.
- AC-002-4: Given aggregation fails for any validation error and the requested output path did not exist before invocation, when the process exits, then the requested output path remains absent.
- AC-002-4a: Given aggregation fails for any validation error and the requested output path existed before invocation, when the process exits, then the requested output path contents are unchanged.
- AC-002-5: Given one input was produced by `scip-python` and another was produced by `scip-typescript`, when aggregation runs, then it fails with a mixed-indexer-family diagnostic.
- AC-002-6: Given `--out python.scip` resolves to the same filesystem path as one input `--index python.scip`, when aggregation runs, then it fails before opening the output for writing.
- AC-002-7: Given all inputs were produced by `scip-typescript` and contain a mix of TypeScript, TSX, JavaScript, or declaration-file document language labels, when aggregation runs, then differing document language labels alone do not cause rejection.

---

## Feature FT-003 - Preserve Query-Time Simplicity

### References

- README: `README.md:173-225`.
- source: `internal/runtime/index.go:5-10`.

### Functional Requirements

- FR-003-1: Existing query commands MUST NOT accept multiple `--index` values as part of this capability.
- FR-003-2: Existing query commands MUST NOT discover or aggregate root indexes at query time.
- FR-003-3: Documentation MUST explain that aggregation is a pre-query artifact generation step.
- FR-003-4: Documentation MUST explain that cross-root references are available only when input indexes already use matching SCIP symbol identities.
- FR-003-5: Query outputs over aggregate indexes MUST use the same result row and payload shapes as query outputs over normal single-root indexes, plus the shared project-root metadata defined in FT-004.
- FR-003-6: Query outputs MUST NOT be grouped into sections per original input root.

### Acceptance Criteria

- AC-003-1: Given an aggregate index was produced earlier, when a query command runs with `--index <aggregate>`, then the command follows the same one-shot load/query/exit lifecycle as any other SCIP index.
- AC-003-2: Given a caller wants fresh aggregate results, when source roots change, then the caller must rerun the language indexers and the aggregation step before querying.
- AC-003-3: Given documentation for the feature is read, when a user expects the aggregate step to repair mismatched symbols across roots, then the docs state that this is out of scope.
- AC-003-4: Given an aggregate index contains documents from `apps/web/src` and `services/some-service`, when `symbols --one-line` returns matches from both inputs, then stdout uses one `project_root` header and compact aggregate-relative paths for all rows.

---

## Feature FT-004 - Expose Project Root for Path Resolution

FT-004 applies to all successful path-bearing query outputs, not only outputs over aggregate indexes. It is additive metadata for JSON and graph-export payloads. It changes agent-facing text and Markdown output by adding one project-root metadata line before path rows or path-bearing sections.

### References

- pairing discussion: 2026-06-10 project-root output question.
- source: `internal/traversal/facts.go:5-11`.
- source: `internal/query/graphexport/export.go:38-45`.

### Functional Requirements

- FR-004-1: Successful one-line agent-facing text outputs that contain document paths MUST include a first line `# project_root=<scip-project-root-uri>`.
- FR-004-1a: Successful Markdown outputs that contain document paths MUST include `Project root: <scip-project-root-uri>` as the first line, before the existing Markdown heading.
- FR-004-2: The project root header MUST appear once per command output, not once per result row or once per original input root.
- FR-004-3: One-line result rows MUST continue to use compact SCIP document paths relative to the loaded index `metadata.project_root`.
- FR-004-4: `--location-only` outputs MUST remain path-only rows and MUST NOT include the project root header.
- FR-004-4a: Documentation MUST state that `--location-only` is only sufficient for path resolution when the caller already knows the selected index `project_root`; otherwise the caller must use a non-location text output, JSON output, or graph-export metadata to obtain it.
- FR-004-5: Successful query JSON outputs that contain document paths MUST include top-level `project_root` with the loaded index `metadata.project_root` URI while preserving existing query payload fields.
- FR-004-5a: In v1, `project_root` is required for `symbols --json`, `symbols --nested-json`, `references --json`, `implementations --json`, `graph --json`, `callers --json`, `callees --json`, and `impact --json`.
- FR-004-5b: In v1, `packages --json` MUST NOT add `project_root` because package results contain no document paths.
- FR-004-6: `graph-export` JSON MUST include the loaded index project root at `inputs.scip_index.project_root`.
- FR-004-7: Consumers MUST resolve graph-export node paths with `inputs.scip_index.project_root + node.document_path`, not source sections.

### Acceptance Criteria

- AC-004-1: Given a single-root index with `metadata.project_root=file:///home/me/Workspace/the-repo/apps/web/src`, when `symbols --one-line` returns `App.tsx`, then stdout starts with that project root and keeps `App.tsx` compact.
- AC-004-2: Given an aggregate index with `metadata.project_root=file:///home/me/Workspace/the-repo`, when `symbols --one-line` returns documents from multiple source roots, then stdout has one project root header and rows such as `apps/web/src/App.tsx:35:10`.
- AC-004-3: Given `graph-export` runs on an aggregate index, then `inputs.scip_index.project_root` is the aggregate root and every node `document_path` is aggregate-root-relative.
- AC-004-4: Given `references --location-only` runs on an index with a sub-root `metadata.project_root`, then stdout contains only location rows and no project-root header.
- AC-004-5: Given `graph --markdown` returns a definition path, then the Markdown output includes the loaded index project root once and leaves the definition path compact.
