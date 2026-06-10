# scip-search v0.1.0

First public release of `scip-search`: a fast, one-shot CLI for querying worktree-local SCIP indexes with output formats designed for coding agents.

---

## Why scip-search?

Agentic coding depends on concurrent worktrees, but most code-indexing solutions assume a global index or editor-managed state. That breaks down when several agents are changing related code in separate checkouts and each one needs answers from its own local state.

`scip-search` was built around worktree-local indexes: point a query at the specific SCIP file for the checkout you are working in, ask a narrow question, and get stable text or JSON output that reflects that index. It also keeps the agent navigation loop explicit, replacing repeated grep/read passes with symbol-aware source locations that can be fed into the next shell read.

The result is a small CLI that helps agents move from symbol discovery to precise source locations in a few commands, without compiling the project, starting a daemon, or mutating the worktree.

---

## Highlights

- Worktree-local index selection via explicit `--index` paths for concurrent agent workflows
- Agent-oriented symbol, reference, implementation, package, graph, caller, callee, impact, and graph-export queries over SCIP indexes
- Grep-style one-line defaults for low-token shell workflows, with JSON and Markdown modes where structured or richer output is useful
- Repeated `--name` and `--symbol` inputs with de-duplicated results for batch lookup
- Static graph and impact views for review boundaries, dependency hints, and test hints
- Installer and source-build workflows for macOS/Linux distribution validation

---

## Core Features

**SCIP Index Queries**
- `symbols` finds indexed symbols by literal partial name
- `references` finds occurrences for exact SCIP symbols or symbols resolved by name
- `implementations` follows SCIP implementation relationships
- `packages` lists package identities, with optional prefix filtering

**Static Graph Views**
- `graph` shows incoming, outgoing, and relationship edges for selected symbols
- `callers` focuses on incoming dependents
- `callees` focuses on outgoing dependencies
- `impact` groups static review, dependency, and test hints
- `graph-export` emits a factual JSON graph artifact and can write it to a file

**Agent-Friendly Output**
- One-line text is the default query output
- `--json` returns structured payloads for automation
- `--nested-json` compacts symbol results by package identity
- `--markdown` provides readable graph and impact summaries
- `--location-only` emits bare source locations for exact-symbol reference and implementation queries

**Runtime Contract**
- Query commands read only the caller-supplied `--index` file
- No default index discovery, compilation, type-checking, daemon, cache, or watch mode
- Usage failures, index loading failures, and runtime failures use distinct exit codes
- Successful query commands write selected output to stdout and keep stderr empty

---

## Tooling

CLI commands exposed by `scip-search`:

| Command | Purpose |
|---------|---------|
| `symbols` | Find symbols by literal partial name |
| `references` | Find references to exact symbols or symbols resolved by name |
| `implementations` | Find implementations from SCIP relationship edges |
| `packages` | List package identities in an index |
| `graph` | Show static incoming, outgoing, and relationship edges |
| `callers` | Show incoming dependents |
| `callees` | Show outgoing dependencies |
| `impact` | Show static review, dependency, and test hints |
| `graph-export` | Export a factual SCIP symbol graph as JSON |

Distribution and validation helpers:

| File | Purpose |
|------|---------|
| `install.sh` | Install the latest release, an explicit release, or a source branch |
| `Makefile` | Build, test, install, and run distribution validation workflows |
| `scripts/validate-distribution.sh` | Validate supported install paths |
| `docs/release-validation.md` | Maintainer checklist for release distribution workflows |

---

## Documentation & Specs

- **README** (`README.md`): CLI purpose, SCIP background, usage examples, runtime contract, installation, complementary tools, and explicit non-goals
- **Release validation** (`docs/release-validation.md`): commands for latest-release, explicit-release, custom-directory, branch-source, and local-clone installs
- **Specifications** (`specs/`): planning and story documents covering command routing, runtime behavior, traversal fixtures, graph/impact, graph export, and release/install workflows

---

## Getting Started

```bash
# Install the latest release
curl -fsSL https://raw.githubusercontent.com/liza-mas/scip-search/main/install.sh | bash
scip-search --version

# Generate a Go SCIP index from a Go module
scip-go index --output go.scip

# Find symbols, then inspect semantic references
scip-search symbols --index go.scip --name Handler
scip-search references --index go.scip --name Handler --one-line
```

**Requirements**: a pre-built SCIP index for query commands. Source installs require Go and make. SCIP language indexers, such as `scip-go`, are installed separately from `scip-search`.

---

## Known Limitations (v0.1.0)

- **Index-only**: Reads existing SCIP indexes; does not generate, update, discover, cache, or watch indexes
- **No daemon or UI**: One process per query, terminal-first output
- **No semantic/vector search**: Symbol-aware SCIP queries only; no embeddings or similarity search
- **Static graph hints**: Graph, caller, callee, and impact output are SCIP-derived hints, not complete runtime call graphs
- **No MCP server**: Designed as a CLI primitive rather than a long-running integration service

---

## What's Next

- Multi-root SCIP index aggregation: produce one standard per-language SCIP index from multiple per-root indexes so existing `--index` queries can operate across a multi-root repository
- Project-root metadata in query output so aggregate indexes remain easy to resolve without grouping normal results by original input root

---

## Philosophy

`scip-search` optimizes for precise, auditable code-navigation loops. It does not try to replace `rg`, `ast-grep`, editors, or language servers; it gives agents a narrow SCIP-backed primitive for the query class where symbol identity matters.

The intended loop is simple:

```bash
scip-search symbols --index <path> --name Foo
scip-search references --index <path> --symbol '<exact-symbol>' --location-only
nl -ba <result-path> | sed -n '<first-line>,<last-line>p'
```
