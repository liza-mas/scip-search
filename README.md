# SCIP Search

## What is SCIP

[SCIP](https://github.com/scip-code/scip) is a language-agnostic Protobuf format for code intelligence, developed by Sourcegraph. It powers Go to Definition, Find References, and Find Implementations across languages. It replaces the older LSIF format.

Key properties relevant to this use case:

- Binary Protobuf — compact on disk, fast to load
- Human-readable symbol identifiers (not opaque numeric IDs)
- Standardised schema covering symbols, occurrences, references, relationships, and hover documentation
- Rich Go bindings available for building consumers
- Produced by mature, actively maintained indexers for the different languages

### Existing Indexers

The appropriate indexers per language need to be installed in the environment.

| Language   | Indexer            | Install                                      | Symbols | References | Implementations |
|------------|--------------------|----------------------------------------------|---------|------------|-----------------|
| Go         | `scip-go`          | `go install github.com/scip-code/scip-go/cmd/scip-go@latest` | ✓ | ✓ | ✓ |
| TypeScript | `scip-typescript`  | `npm install -g @sourcegraph/scip-typescript` | ✓ | ✓ | ✓ |
| Python     | `scip-python`      | `pip install scip-python`                    | ✓ | partial | — |
| Java/Kotlin/Scala | `scip-java` | via Gradle/Maven plugin                      | ✓ | ✓ | ✓ |
| Rust       | `rust-analyzer`    | ships with rustup                            | ✓ | ✓ | ✓ |

As the indexers produce Protobuf files, it is possible for multiple indexes to coexist at different locations — essential to support concurrent worktrees.

**Known limitation:** `scip-go` expects a Go module. Run it from the directory containing `go.mod`, or pass `--module-root`.

Generate a Go SCIP index:
```bash
scip-go -o /path/to/go.scip
scip-go --module-root /path/to/repo -o /path/to/go.scip
```

### SCIP Symbol Format

SCIP uses human-readable string identifiers for symbols. The format is:

```
<scheme> <package-manager> <package-name> <package-version> <descriptors>
```

Example Go symbols:
```
scip-go gomod github.com/liza-mas/liza . supervisor/Supervisor#
scip-go gomod github.com/liza-mas/liza . supervisor/Run().
scip-go gomod github.com/liza-mas/liza . agent/Doer#
```

`scip-search` resolves partial name queries (e.g. `--name Supervisor`) to full SCIP symbols and returns them alongside results, so agents can use the full symbol string in subsequent `references` or `implementations` calls.

---

## What is SCIP Search

**`scip-search`** is a thin Go binary that:

1. Loads a SCIP index file at the path provided with `--index`
2. Answers a query using the SCIP Go bindings
3. Prints structured JSON to stdout
4. Exits

Cold start is milliseconds — loads a pre-built binary index, performs no compilation or type-checking.

```bash
scip-search symbols --index <index-path> --name <name>
scip-search references --index <index-path> --symbol <scip-symbol>
scip-search implementations --index <index-path> --symbol <scip-symbol>
scip-search packages --index <index-path> [--prefix <prefix>]
```

Examples:
```bash
scip-search symbols --index /path/to/go.scip --name Supervisor
scip-search references --index /path/to/go.scip --symbol 'scip-go gomod github.com/liza-mas/liza . supervisor/Run().'
scip-search implementations --index /path/to/go.scip --symbol 'scip-go gomod github.com/liza-mas/liza . agent/Doer#'
scip-search packages --index /path/to/go.scip
```

### Runtime Contract

All query commands require `--index <index-path>`.

When a query command succeeds, `scip-search` writes exactly one JSON value to stdout, writes nothing to stderr, and exits with status `0`.

Shared invocation failures, including a missing query command, an unsupported query command, or a missing `--index`, are usage failures. They leave stdout empty, write a human-readable diagnostic to stderr, and exit with status `2`.

After the shared runtime accepts an index path, failures while opening or loading that path belong to the subsequent index-loading boundary. This shared invocation contract does not define loader diagnostics, query result fields, or traversal behavior.

### Language Support

Uses the official SCIP bindings (e.g. `github.com/scip-code/scip/bindings/go/scip`) for parsing and traversal.

`scip-search` reads the SCIP output directly.

### Out of scope

- No daemon, no watch mode, no incremental updates
- No embedding, no semantic similarity, no vector store
- No UI, no graph visualization
- No MCP server
- No custom index format — SCIP is the format

Planned: For ctags fallback languages, `scip-search` reads a thin JSON wrapper (same query interface, reduced capability — definitions only).

---

## Complementary Existing Tool

**[ast-grep](https://github.com/ast-grep/ast-grep)** (not built here)

A structural search tool that scans the worktree on demand using language-aware pattern matching.
Covers the query class that `scip-search` cannot: the caller knows a structure but not a name.

```bash
ast-grep --pattern 'func $F(ctx context.Context, $$$) error' --lang go .
ast-grep --pattern 'return nil, errors.New($$$)' --lang go .
```

Handles multiple languages natively by file extension — no per-language invocation needed.

| Question | Tool |
|---|---|
| Where is `Supervisor` defined? | `scip-search symbols` |
| What implements `Doer`? | `scip-search implementations` |
| What calls `blackboard.Write`? | `scip-search references` |
| What packages exist? | `scip-search packages` |
| Find all functions returning unwrapped errors | `ast-grep` |
| Find all struct literals missing field `Timeout` | `ast-grep` |

---

## Installation

### Prerequisites

Install the appropriate language indexers listed in [Existing Indexers](#existing-indexers) before generating SCIP indexes for `scip-search`.

**Quick install (latest release, macOS/Linux):**

```bash
curl -fsSL https://raw.githubusercontent.com/liza-mas/scip-search/main/install.sh | bash
```

**Options:**

```bash
# Specific version
curl -fsSL https://raw.githubusercontent.com/liza-mas/scip-search/main/install.sh | VERSION=v1.0.0 bash

# Build from a branch (requires Go and make)
curl -fsSL https://raw.githubusercontent.com/liza-mas/scip-search/main/install.sh | BRANCH=main bash

# Custom directory
curl -fsSL https://raw.githubusercontent.com/liza-mas/scip-search/main/install.sh | INSTALL_DIR=~/.local/bin bash
```

**From a local clone:**

```bash
git clone https://github.com/liza-mas/scip-search.git && cd scip-search
make install
```

**Verify:**

```bash
scip-search --version
```
