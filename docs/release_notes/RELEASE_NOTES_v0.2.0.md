# scip-search v0.2.0

`scip-search` v0.2.0 adds multi-root SCIP index aggregation so agents can query repositories with several same-language roots through one standard SCIP protobuf index.

This release also exposes selected-index project-root metadata in path-bearing query outputs, making compact paths resolvable for both single-root and aggregate indexes.

---

## Highlights

- New `aggregate-index` command builds one standard SCIP index from multiple same-language input indexes
- Aggregate document paths are rewritten into one repository-root-relative path space under the supplied `--project-root`
- Path-bearing query outputs now expose the loaded index project root once per output
- Aggregate validation rejects ambiguous or corrupt outputs before replacing the requested output file
- External symbol handling preserves repeated local symbols and de-duplicates non-local external symbols by SCIP symbol identity
- README installation guidance now points Python users at `npm install -g @sourcegraph/scip-python`

---

## New Command: aggregate-index

`aggregate-index` is a pre-query artifact generation command. It reads two or more already-generated same-language SCIP protobuf indexes and writes one aggregate SCIP protobuf index:

```bash
scip-search aggregate-index \
  --project-root /home/me/Workspace/the-repo \
  --root apps/api --index apps/api/index.scip \
  --root services/some-service --index services/some-service/index.scip \
  --out python.scip
```

The output is a normal SCIP index. Existing query commands still read exactly one caller-selected `--index`, so the query workflow stays unchanged after aggregation:

```bash
scip-search symbols --index python.scip --name Handler
scip-search references --index python.scip --name Handler --one-line
```

`--project-root` accepts an absolute filesystem path or `file://` URI. Each `--root` value is a slash-separated path relative to that aggregate project root; `.` identifies the project root itself.

---

## Output Metadata

Path-bearing one-line outputs now start with:

```text
# project_root=<scip-project-root-uri>
```

Path-bearing Markdown outputs start with:

```text
Project root: <scip-project-root-uri>
```

Path-bearing JSON query payloads include top-level `project_root`. `graph-export` includes the selected index project root at `inputs.scip_index.project_root`.

`--location-only` output remains path-only and intentionally omits the project-root header. Use one-line, JSON, Markdown, or graph-export output when the selected index project root is not already known.

---

## Validation and Safety

Aggregation fails without replacing the output file when it detects:

- fewer than two completed `--root <root> --index <path>` input pairs
- invalid `--project-root`, invalid `--root`, or malformed flag ordering
- unreadable or invalid input SCIP files
- duplicate aggregate document paths
- source-root mappings that conflict with comparable input `metadata.project_root` values
- mixed input indexer families, such as combining `scip-python` and `scip-typescript`
- duplicate non-local definition symbols that resolve to different aggregate document paths
- output paths that resolve to one of the input index paths

Malformed `aggregate-index` command lines exit `2`; unreadable or invalid input SCIP files exit `3`; aggregation validation failures exit `4`.

---

## Upgrade Notes

- Consumers that parse one-line path-bearing output should handle the new leading `# project_root=...` metadata line.
- Consumers that parse path-bearing JSON query payloads should tolerate the new top-level `project_root` field.
- `graph-export` consumers should resolve node paths with `inputs.scip_index.project_root` plus each node `document_path`.
- `packages --json` is unchanged because package results do not contain document paths.

---

## Known Limitations

- Aggregation is same-language only; build one aggregate index per language.
- `aggregate-index` does not run language indexers, discover roots, infer imports, rewrite symbols, or repair mismatched package identities.
- Cross-root references are available only when input indexes already use matching SCIP symbol identities.
- Aggregate indexes do not preserve per-input source provenance or section query results by original root.
- Query commands still read one explicit `--index`; they do not aggregate at query time.

---

## Documentation

- **README** (`README.md`): updated CLI syntax, runtime contract, multi-root aggregation workflow, project-root metadata, and SCIP Python installation guidance
- **Multi-root PRD** (`specs/goals/20260610-multiroots.md`): requirements and acceptance criteria for aggregation, path metadata, validation, and non-goals
- **Release validation** (`docs/release-validation.md`): unchanged distribution validation checklist
