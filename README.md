# SCIP Search

## A Code Exploration CLI Optimized for Coding Agents

Usage example:
1. Find relevant symbols
```bash
❯ scip-search symbols --index go.scip --name SymbolSource
internal/traversal/facts.go:14:6 symbol="scip-go gomod scip-search 552d21c479c8 `scip-search/internal/traversal`/SymbolSource#"; match=displayName; text=SymbolSource
internal/traversal/facts.go:17:2 symbol="scip-go gomod scip-search 552d21c479c8 `scip-search/internal/traversal`/SymbolSourceDocument."; match=displayName; text=SymbolSourceDocument
internal/traversal/facts.go:18:2 symbol="scip-go gomod scip-search 552d21c479c8 `scip-search/internal/traversal`/SymbolSourceExternal."; match=displayName; text=SymbolSourceExternal
```
2. Locate the symbols
```bash
❯ scip-search references --index go.scip --symbol 'scip-go gomod scip-search 552d21c479c8 `scip-search/internal/traversal`/SymbolSource#' --location-only
internal/traversal/facts.go:17:23
internal/traversal/facts.go:18:23
internal/traversal/facts.go:31:25
internal/traversal/view.go:130:57
internal/traversal/view_test.go:780:59
```
3. Focus read
```bash
❯ nl -ba internal/traversal/facts.go | sed -n '17,31p'
17          SymbolSourceDocument SymbolSource = "document"
18          SymbolSourceExternal SymbolSource = "external"
19  )
20
21  type Document struct {
22          RelativePath     string
23          Language         string
24          PositionEncoding scip.PositionEncoding
25          Symbols          []Symbol
26          Occurrences      []Occurrence
27  }
28
29  type Symbol struct {
30          Symbol                 string
31          Source                 SymbolSource
```

The pattern:
```
scip-search symbols --index <path> --name Foo --name Bar
scip-search references --index <path> --symbol '<exact-foo>' --symbol '<exact-bar>' --location-only
nl -ba <result-path> | sed -n '<first-line>,<last-line>p'
```

That's the full loop: three Bash calls replacing the 5-10 grep/read round-trips agents typically need.

Not only `scip-search` is token efficient but it also support concurrent indexes for local worktrees.

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
| Python     | `scip-python`      | `npm install -g @sourcegraph/scip-python`    | ✓ | partial | — |
| Java/Kotlin/Scala | `scip-java` | via Gradle/Maven plugin                      | ✓ | ✓ | ✓ |
| Rust       | `rust-analyzer`    | ships with rustup                            | ✓ | ✓ | ✓ |

As the indexers produce Protobuf files, it is possible for multiple indexes to coexist at different locations — essential to support concurrent worktrees.

**Known limitation:** `scip-go` expects a Go module. Run it from the directory containing `go.mod`, or pass `--module-root`.

Generate a Go SCIP index:
```bash
scip-go index --output /path/to/go.scip
scip-go index --module-root /path/to/repo --output /path/to/go.scip
```

### SCIP Symbol Format

SCIP uses human-readable string identifiers for symbols. The format is:

```
<scheme> <packageManager> <packageName> <packageVersion> <descriptor>
```

`<scheme> <packageManager> <packageName> <packageVersion>` forms the `packageKey`.

Example Go symbols:
```
scip-go gomod scip-search 8ae7b309d177 `scip-search/internal/traversal`/SymbolSource#
scip-go gomod scip-search 8ae7b309d177 `scip-search/internal/traversal`/SymbolSourceDocument.
scip-go gomod scip-search 8ae7b309d177 `scip-search/internal/cli`/Handler#
```

`scip-search` resolves partial name queries (e.g. `--name SymbolSource`) to full SCIP symbols.
The default `symbols` response is one line per match.
`symbols --nested-json` groups matching descriptors by package to reduce repeated package identity text.
A full SCIP symbol can be reconstructed as `<packageKey> <descriptor>` from one-line or nested JSON output and used in
subsequent `references` or `implementations` calls.
The package version comes from the indexed checkout and varies by commit.

In default one-line `symbols` output, the value to pass to `references --symbol`
or `implementations --symbol` is the JSON string after `symbol=`. Decode that
JSON string to recover the exact SCIP symbol:

```text
internal/traversal/facts.go:14:6 symbol="scip-go gomod scip-search 552d21c479c8 `scip-search/internal/traversal`/SymbolSource#"; match=displayName; text=SymbolSource
```

Use the decoded value as the exact symbol:

```text
scip-go gomod scip-search 552d21c479c8 `scip-search/internal/traversal`/SymbolSource#
```

Automation should prefer `symbols --json` to ease parsing.
The full SCIP symbol for `references --symbol` or `implementations --symbol` is `<scheme> <packageManager> <packageName> <packageVersion> <descriptor>`.

---

## What is SCIP Search

**`scip-search`** is a thin Go binary that:

1. Loads a SCIP index file at the path provided with `--index`
2. Answers a query using the official SCIP bindings for Go
3. Prints the selected successful output format to stdout
4. Exits

Cold start is milliseconds — loads a pre-built binary index, performs no compilation or type-checking.

```bash
scip-search --help
scip-search --version
scip-search symbols --index <index-path> --name <name> [--name <name>]... [--one-line|--nested-json|--json]
scip-search references --index <index-path> [--symbol <scip-symbol>]... [--name <name>]... [--one-line|--json|--location-only]
scip-search implementations --index <index-path> [--symbol <scip-symbol>]... [--name <name>]... [--one-line|--json|--location-only]
scip-search packages --index <index-path> [--prefix <prefix>] [--one-line|--json]
scip-search graph --index <index-path> [--symbol <scip-symbol>]... [--name <name>]... [--one-line|--json|--markdown]
scip-search callers --index <index-path> [--symbol <scip-symbol>]... [--name <name>]... [--one-line|--json|--markdown]
scip-search callees --index <index-path> [--symbol <scip-symbol>]... [--name <name>]... [--one-line|--json|--markdown]
scip-search impact --index <index-path> [--symbol <scip-symbol>]... [--name <name>]... [--one-line|--json|--markdown]
scip-search graph-export --index <index-path> [--symbol <scip-symbol>]... [--name <name>]... [--package-prefix <prefix>]... [-o <path>]
scip-search aggregate-index --project-root <path-or-file-uri> --root <repo-relative-root> --index <input-scip-path> [--root <repo-relative-root> --index <input-scip-path>]... --out <output-scip-path>
```

Examples:
```bash
scip-search symbols --index /path/to/go.scip --name SymbolSource
scip-search symbols --index /path/to/go.scip --name SymbolSource --nested-json
scip-search symbols --index /path/to/go.scip --name SymbolSource --json
scip-search references --index /path/to/go.scip --symbol 'scip-go gomod scip-search 8ae7b309d177 `scip-search/internal/traversal`/SymbolSource#'
scip-search references --index /path/to/go.scip --name SymbolSource
scip-search references --index /path/to/go.scip --symbol 'scip-go gomod scip-search 8ae7b309d177 `scip-search/internal/traversal`/SymbolSource#' --json
scip-search implementations --index /path/to/go.scip --symbol 'scip-go gomod scip-search 8ae7b309d177 `scip-search/internal/cli`/Handler#'
scip-search implementations --index /path/to/go.scip --name Handler
scip-search packages --index /path/to/go.scip
scip-search graph --index /path/to/go.scip --name Handler
scip-search callers --index /path/to/go.scip --name Handler
scip-search callees --index /path/to/go.scip --name Handler
scip-search impact --index /path/to/go.scip --name Handler --markdown
scip-search graph-export --index /path/to/go.scip --package-prefix scip-search -o graph.json
scip-search aggregate-index --project-root /repo --root apps/web --index apps/web/index.scip --root services/api --index services/api/index.scip --out typescript.scip
```

### Runtime Contract

All query commands require `--index <index-path>`.

`symbols` accepts one or more `--name <name>` values. `references`, `implementations`, `graph`, `callers`, `callees`, and `impact` require at least one symbol source: `--symbol <scip-symbol>`, `--name <name>`, or both. `graph-export` accepts optional `--symbol`, `--name`, and `--package-prefix` filters plus optional `-o <path>` file output; with no filters it exports every factual graph node and edge it can derive from the selected SCIP index. `--name` and `--symbol` can be repeated; repeated resolved symbols are de-duplicated. When multiple `symbols --name` values match the same symbol, the result appears once and its `matchText` / `matchSource` come from the first matching `--name` value in CLI order.

`scip-search --help` and `scip-search --version` are global commands. They do not require `--index`, write human-readable text to stdout, and exit with status `0`.

When a query command succeeds, `scip-search` writes the selected output format to stdout, writes nothing to stderr, and exits with status `0`. By default, query commands write one-line text output. Successful path-bearing one-line outputs start with one metadata line:

```text
# project_root=<scip-project-root-uri>
```

`symbols --nested-json`, `symbols --json`, `references --json`, `implementations --json`, `graph --json`, `callers --json`, `callees --json`, and `impact --json` write exactly one JSON value to stdout. Path-bearing query JSON payloads include top-level `project_root` with the selected index metadata project root. `packages --json` does not add `project_root` because package results do not contain document paths. `graph-export` writes exactly one JSON value to stdout unless `-o <path>` is provided; with `-o`, it writes that JSON value to the selected file and leaves stdout empty. It includes the selected index project root at `inputs.scip_index.project_root`. It does not have text or Markdown output modes. `graph --markdown`, `callers --markdown`, `callees --markdown`, and `impact --markdown` write compact multi-line Markdown text for agent reading; path-bearing Markdown output starts with `Project root: <scip-project-root-uri>`.

By default, `symbols --name` returns one grep-style line per matched symbol:

```text
<path>:<line>:<column> symbol="<packageKey> <descriptor>"; match=<matchSource>; text=<matchText>
```

Default reference and implementation output also use one source-location-prefixed line per result:

```text
<path>:<line>:<column> symbol="<referenced-symbol>"; roles=<roles>
<path>:<line>:<column> symbol="<implementation-symbol>"
```

Default graph and impact output stays one physical line per fact:

```text
<path>:<line>:<column> symbol="<symbol>"; direction=<incoming|outgoing>; roles=<roles>
<path>:<line>:<column> symbol="<symbol>"; section=<review|dependency|tests>; ...
?:0:0 symbol="<symbol>"; relationship="<relationship-kinds>"; direction=<incoming|outgoing>
?:0:0 symbol="<symbol>"; direction=outgoing; unavailable="<reason>"
```

`--location-only` selects source locations without metadata for exact-symbol reference and implementation queries:

```text
<path>:<line>:<column>
```

`--location-only` output intentionally omits the project-root header. It is sufficient for path resolution only when the caller already knows the selected index `metadata.project_root`; otherwise use one-line, JSON, Markdown, or graph-export output to obtain the project root.

`--one-line` explicitly selects the default one-line output. For `symbols`, `--nested-json` returns the compact package-grouped payload, while `--json` returns one self-contained JSON entry per symbol with `scheme`, `packageManager`, `packageName`, and `packageVersion` repeated on every symbol result. For `references`, `implementations`, `packages`, `graph`, `callers`, `callees`, and `impact`, `--json` selects the structured JSON payload. `references`, `implementations`, `graph`, `callers`, `callees`, `impact`, and `graph-export` accept `--symbol`, `--name`, or both. `--name` is resolved through the same literal symbol-name discovery used by `symbols --name`; when `--name` is present, query JSON output contains `symbols` and per-symbol `queries` so multiple resolved symbols can be represented in one JSON value. `graph-export --package-prefix` matches SCIP package names, and filtered graph exports omit edges unless both endpoints remain selected. `--location-only` for `references` and `implementations` requires `--symbol` and cannot be used with `--name`.

`graph`, `callers`, `callees`, and `impact` are static SCIP-derived views. They are useful for code-review boundaries, but they are not complete runtime call graphs. Outgoing dependencies are collected from non-definition occurrences inside the target definition occurrence's SCIP enclosing range. If the definition or enclosing range is absent, the output reports an explicit unavailable reason instead of guessing. `impact` test hints come from SCIP `SymbolRole_Test` occurrences and fixed test-like path patterns in indexed occurrences; no source files or Git diffs are read.

`graph-export` emits the factual graph artifact schema `scip.graph-export.v1`. The artifact includes generator metadata, UTC generation time, the caller-supplied index path, the selected index project root, a `sha256:` fingerprint of the consumed index bytes, explicit node arrays, and explicit edge arrays. Consumers resolve graph-export node paths with `inputs.scip_index.project_root` plus each node `document_path`. Edges use factual provenance such as `scip_relationship` and `contained_dependency`; standalone occurrence references are not exported as edges unless a factual source symbol can be derived from SCIP ranges. Every emitted edge endpoint has a matching node. If a node is present only to close an edge endpoint and no SCIP symbol information exists for it, unknown optional fields such as `external`, package, kind, document path, and location are omitted instead of guessed.

In one-line output, `line` and `column` are the SCIP range start offsets plus 1, not source-file-normalized editor columns. `scip-search` does not read source files to render one-line output. Symbols or implementations without a definition location render as `?:0:0`, which is common for external symbols. Only the `path:line:column` prefix is stable colon-delimited location data; metadata is labeled text after the location prefix. The `symbol=` value is a JSON string because SCIP symbols can contain semicolons; fields after that quoted value are separated by semicolons. `symbols` match text escapes `\`, newline, carriage return, and tab as `\\`, `\n`, `\r`, and `\t` so each result stays on one physical line. `packages` one-line output writes one package key per line.

Shared invocation failures, including a missing query command, an unsupported query command, or a missing `--index`, are usage failures. They leave stdout empty, write a human-readable diagnostic to stderr, and exit with status `2`.

After the shared runtime accepts an index path, every documented query command reads only the caller-supplied `--index` file for the current invocation. `scip-search` does not search for a default index, update or delete the selected index, cache it for later runs, watch it for changes, compile or type-check source code, or parse a custom index format.

Index-loading failures happen before the selected query handler runs. A nonexistent path, unreadable file, directory path, or readable file that is not valid SCIP input leaves stdout empty, writes a human-readable diagnostic to stderr, and exits with status `3`.

Valid SCIP input from any supported SCIP-producing language is loaded through the official SCIP bindings for Go before the selected query handler runs. This loading boundary does not define query result fields or traversal behavior.

### Multi-root aggregation

`aggregate-index` is a pre-query artifact generation command. It reads two or more already-generated same-language SCIP protobuf indexes and writes one standard SCIP protobuf index. Query commands still accept exactly one `--index`; they do not discover roots or aggregate at query time.

Each input pair supplies the repository-relative source root that the input index was generated from:

```bash
scip-search aggregate-index \
  --project-root /home/me/Workspace/the-repo \
  --root apps/api --index apps/api/index.scip \
  --root services/some-service --index services/some-service/index.scip \
  --out python.scip
```

The aggregate output metadata uses the supplied aggregate `--project-root`, and every document path is rewritten into that project-root-relative path space, for example `apps/api/server.py` and `services/some-service/worker.py`. If an input index has a comparable `file://` metadata project root, the supplied `--root` must match its relative path from the aggregate project root. `--project-root` accepts an absolute filesystem path or `file://` URI. `--root` values are slash-separated paths relative to the aggregate project root; `.` means the aggregate project root itself.

Aggregation preserves SCIP symbol strings exactly. Cross-root references are available only when the input indexes already use matching SCIP symbol identities. `aggregate-index` does not run language indexers, discover roots, infer imports, rewrite symbols, or repair mismatched package identities.

Malformed `aggregate-index` command lines exit `2`. Unreadable or invalid input SCIP files exit `3`. Aggregation validation failures after all inputs are readable SCIP data, such as duplicate output document paths, root-mapping mismatches, mixed indexer families, symbol collisions, conflicting external symbol records, or an output path that resolves to an input path, exit `4` and leave any existing output file unchanged.

### Language Support

Uses the official SCIP bindings for Go (e.g. `github.com/scip-code/scip/bindings/go/scip`) to parse language-agnostic SCIP indexes.

`scip-search` reads the SCIP output directly.

### Out of scope

- No daemon, no watch mode, no incremental updates
- No embedding, no semantic similarity, no vector store
- No UI, no graph visualization
- No MCP server
- No custom index format — SCIP is the format
- No complete runtime call graph; graph and impact commands are static SCIP-derived hints
- No `impact --diff`; query commands remain index-only

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

**[ripgrep](https://github.com/BurntSushi/ripgrep)** (not built here)

A fast raw text and file discovery tool that scans the current worktree.
Covers the query class that `scip-search` intentionally avoids: filenames,
comments, string literals, testdata, golden files, generated artifacts, and any
text that may not be an indexed symbol.

```bash
rg -n "LoadSharedFixture|DisplayName" internal cmd -g'*.go'
rg --files | rg 'traversaltest|fixture|golden'
```

`rg` always reflects the files on disk. `scip-search` reflects the supplied SCIP
index, which may be older than the worktree until the index is regenerated.
Because `rg` is purely textual, it can return comments, strings, and other
incidental matches; `scip-search` only returns matches that SCIP indexed as
symbols.

Use the tools by question shape:

- Use `scip-search` when the index should decide what a symbol is: definitions,
  references, implementations, and package identities. It filters out many
  accidental text matches and can follow SCIP relationships.
- Use `ast-grep` when syntax matters more than symbol identity: function
  signatures, call forms, struct literals, return statements, and other
  language-aware patterns.
- Use `rg` when raw worktree discovery is the goal: filenames, arbitrary text,
  comments, strings, fixtures, and files outside the SCIP index.

For example, `rg -n "LoadSharedFixture|DisplayName" internal cmd -g'*.go'`
finds every textual occurrence. `scip-search symbols --name LoadSharedFixture`
finds the indexed function definition, and `scip-search references --name
LoadSharedFixture` finds semantic call sites. Conversely,
`rg --files | rg 'traversaltest|fixture|golden'` is a filesystem query; SCIP
does not model arbitrary filenames or JSON golden files.

| Question | Tool |
|---|---|
| Where is `SymbolSource` defined? | `scip-search symbols` |
| What implements `Handler`? | `scip-search implementations` |
| What references `SymbolSource`? | `scip-search references` |
| What packages exist? | `scip-search packages` |
| Find all functions returning unwrapped errors | `ast-grep` |
| Find all struct literals missing field `Timeout` | `ast-grep` |
| Find every textual mention of `DisplayName` | `rg` |
| Find fixture or golden files by path | `rg --files \| rg 'fixture\|golden'` |

---

## Installation

### Prerequisites

Installing `scip-search` only installs the `scip-search` CLI.

Language indexers are separate tools used to generate SCIP indexes before
running query commands. Install the appropriate indexers listed in
[Existing Indexers](#existing-indexers) when you need to create SCIP data; they
are not installed by the `scip-search` installer.

**Quick install (latest release, macOS/Linux):**

```bash
curl -fsSL https://raw.githubusercontent.com/liza-mas/scip-search/main/install.sh | bash
scip-search --version
```

**Options:**

```bash
# Explicit release
curl -fsSL https://raw.githubusercontent.com/liza-mas/scip-search/main/install.sh | VERSION=<release> bash
scip-search --version

# Build from a branch with caller-provided Go and make
curl -fsSL https://raw.githubusercontent.com/liza-mas/scip-search/main/install.sh | BRANCH=<branch> bash
scip-search --version

# Custom install directory
curl -fsSL https://raw.githubusercontent.com/liza-mas/scip-search/main/install.sh | INSTALL_DIR=<directory> bash
<directory>/scip-search --version
```

**From a local clone:**

```bash
git clone https://github.com/liza-mas/scip-search.git
cd scip-search
make install
scip-search --version
```

Local clone installs also require caller-provided Go and make. Use
`INSTALL_DIR=<directory> make install` to install from a local clone into a
custom directory, then verify with `<directory>/scip-search --version`.

Release notes are available in [`docs/release_notes/`](docs/release_notes/).
