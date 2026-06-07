# SCIP Search - Graph Export

## Goal

Expose the complete static symbol graph contained in a SCIP index for downstream analysis tools.

The export is intended for machine consumption and enables:

- clustering
- impact analysis
- architectural analysis
- process discovery
- visualization

This feature does not perform analysis itself.

## Philosophy

SCIP Search remains a factual tool.

It exports graph facts.

It does not infer:

- business capabilities
- execution flows
- architectural intent

Those concerns belong to higher-level tools.

## Inputs

Required:

- SCIP index

Optional:

- symbol filters
- package filters

## Artifact Metadata

Every export includes:

- `schema_version`: `"scip.graph-export.v1"`
- `generator`: tool name and version string
- `generated_at`: UTC RFC3339 timestamp
- `inputs.scip_index.path`: caller-supplied index path
- `inputs.scip_index.fingerprint`: stable fingerprint of the consumed index bytes

The export command reads only the explicitly supplied SCIP index.

## Outputs

### Nodes

Nodes represent symbols.

Each node contains:

- stable symbol identifier
- display name
- symbol kind
- package
- document path
- location

Locations preserve SCIP range coordinates. They are not source-normalized editor
columns unless the index already uses that encoding.

### Edges

Edges represent relationships.

Examples:

- reference
- implementation
- type-definition
- dependency
- import

Edge provenance must be explicit.

Supported v1 provenance values:

- `scip_relationship`: relationship facts present in the SCIP index
- `occurrence_reference`: non-definition symbol occurrences
- `contained_dependency`: non-definition occurrences inside an indexed enclosing range

`import` is a future provenance value unless the SCIP index exposes it as a
stable fact without source parsing.

Each edge contains:

- source node
- target node
- edge type
- occurrence count
- optional weight

`weight` is reserved for downstream weighted graph artifacts. It is a normalized
`0.0` to `1.0` value. The factual SCIP export does not compute semantic weights
by default. When absent, consumers may derive weight from factual fields such as
edge type, provenance, and occurrence count.

## JSON Schema

The v1 JSON shape is:

```json
{
  "schema_version": "scip.graph-export.v1",
  "generator": {
    "name": "scip-search",
    "version": "..."
  },
  "generated_at": "2026-06-07T00:00:00Z",
  "inputs": {
    "scip_index": {
      "path": "go.scip",
      "fingerprint": "sha256:..."
    }
  },
  "nodes": [
    {
      "id": "scip-go gomod example.com/project cmd/liza/Main().",
      "display_name": "Main",
      "kind": "function",
      "package": "scip-go gomod example.com/project",
      "document_path": "cmd/liza/main.go",
      "location": {
        "range": [10, 0, 10, 4]
      },
      "roles": 1,
      "external": false
    }
  ],
  "edges": [
    {
      "source": "scip-go gomod example.com/project cmd/liza/Main().",
      "target": "scip-go gomod example.com/project internal/commands/Run().",
      "type": "reference",
      "provenance": "occurrence_reference",
      "occurrence_count": 3,
      "weight": 0.82
    }
  ]
}
```

Arrays are present even when empty. Unknown optional facts are omitted, not
serialized as guessed values.

## Identity

The stable node identifier is the full SCIP symbol string.

Document paths are repository-relative paths as recorded by the SCIP index.
Consumers that join this export with Stacklit must use document paths as the
primary cross-artifact key and full SCIP symbols as the symbol key.

## Traversal Consistency

Graph export and live `graph` / `impact` query commands must use the same SCIP
traversal semantics.

The same SCIP index and symbol facts must not produce contradictory definition,
relationship, incoming-reference, outgoing-contained-dependency, or unavailable
reason results between:

- exported graph artifacts
- query-time `graph` output
- query-time `impact` output

An implementation may satisfy this by sharing one traversal core or by
maintaining conformance tests that prove the export and live commands agree on
the same fixture indexes.

## Error Behavior

If the SCIP index cannot be read or decoded, the command fails and emits no
partial artifact.

If a symbol lacks a definition, location, or enclosing range, the export keeps
the node and records the missing field by omission. It must not infer missing
facts from source files.

## Output Formats

### JSON

Human-readable.

Suitable for debugging.

### Binary

Future / out of scope for v1.

No binary format is part of this contract until a concrete consumer requires
one.

## Constraints

- No source file parsing.
- No repository scanning.
- No hidden index discovery.
- Only consumes the explicitly supplied SCIP index.

## Success Criteria

Given a valid SCIP index, the export command writes one JSON artifact that:

- contains `schema_version`, `generator`, `generated_at`, and input fingerprint metadata
- contains every exported node with a stable SCIP symbol identifier
- contains only edges whose provenance is explicit
- can be consumed by downstream tools without re-reading the SCIP index
- is produced without reading source files or discovering hidden indexes
