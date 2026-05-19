package cli

import (
	"fmt"
	"io"
	"strings"

	runtimecontract "scip-search/internal/runtime"
	"scip-search/internal/scipindex"
	"scip-search/internal/version"
)

var documentedCommands = []string{
	"symbols",
	"references",
	"implementations",
	"packages",
}

const helpText = `Description:
  Search a pre-built SCIP index for symbols, references, implementations, and packages.

Usage:
  scip-search --help
  scip-search --version
  scip-search symbols --index <index-path> --name <name> [--name <name>]... [--one-line|--nested-json|--json]
  scip-search references --index <index-path> [--symbol <scip-symbol>]... [--name <name>]... [--one-line|--json]
  scip-search implementations --index <index-path> [--symbol <scip-symbol>]... [--name <name>]... [--one-line|--json]
  scip-search packages --index <index-path> [--prefix <prefix>] [--one-line|--json]

Commands:
  symbols          Find symbols by literal partial name.
  references       Find references to exact SCIP symbols or symbols found by name.
  implementations  Find implementations of exact SCIP symbols or symbols found by name.
  packages         List package identities in an index.

Output:
  --one-line     Grep-style text output; default for all query commands.
  --json         Structured JSON output.
  --nested-json  Compact package-grouped JSON output for symbols only.

One-line formats:
  symbols          <path>:<line>:<column>:<packageKey> <descriptor> match=<source> text=<text>
  references       <path>:<line>:<column>:<referenced-symbol> roles=<roles>
  implementations  <path>:<line>:<column>:<implementation-symbol>
  packages         <packageKey>

Notes:
  symbols accepts repeated --name; references and implementations accept repeated --name and --symbol.
  Repeated results are de-duplicated.
  references and implementations require --symbol, --name, or both.
  Reads an existing SCIP index; does not generate, update, or discover indexes.

Exit codes:
  0  success
  1  unexpected runtime error
  2  usage error
  3  index loading error

Examples:
  scip-search symbols --index go.scip --name Handler
  scip-search references --index go.scip --name Handler --one-line
`

// Loader is the boundary that later runtime stages use after command selection.
type Loader interface {
	Load(indexPath string) (any, error)
}

// Handler is the selected query boundary for one documented command.
type Handler interface {
	Handle(loadedIndex any, args []string) (any, error)
}

// Runtime owns shared CLI command routing and invocation-shape failures.
type Runtime struct {
	loader        Loader
	registry      map[string]Handler
	buildIdentity version.BuildIdentity
}

// RoutedCommand is the documented command selected from the shared registry.
type RoutedCommand struct {
	Name    string
	Handler Handler
}

// NewRuntime builds a runtime with exactly the documented query command names.
func NewRuntime(loader Loader, handlers map[string]Handler) Runtime {
	return NewRuntimeWithBuildIdentity(loader, handlers, version.Current())
}

// NewProductionRuntime builds a runtime backed by the production SCIP index loader.
func NewProductionRuntime(handlers map[string]Handler) Runtime {
	return NewProductionRuntimeWithBuildIdentity(handlers, version.Current())
}

// NewProductionRuntimeWithBuildIdentity builds a production-loader runtime with explicit build provenance.
func NewProductionRuntimeWithBuildIdentity(handlers map[string]Handler, buildIdentity version.BuildIdentity) Runtime {
	return NewRuntimeWithBuildIdentity(scipindex.NewLoader(), handlers, buildIdentity)
}

// NewRuntimeWithBuildIdentity builds a runtime with explicit offline build provenance.
func NewRuntimeWithBuildIdentity(loader Loader, handlers map[string]Handler, buildIdentity version.BuildIdentity) Runtime {
	registry := make(map[string]Handler, len(documentedCommands))
	for _, name := range documentedCommands {
		handler, ok := handlers[name]
		if ok && handler != nil {
			registry[name] = handler
		}
	}

	return Runtime{
		loader:        loader,
		registry:      registry,
		buildIdentity: buildIdentity,
	}
}

// Route selects a documented query command and leaves later flags opaque.
func (rt Runtime) Route(args []string) (RoutedCommand, []string, bool) {
	if len(args) == 0 {
		return RoutedCommand{}, nil, false
	}

	commandName := args[0]
	handler, ok := rt.registry[commandName]
	if !ok {
		return RoutedCommand{}, nil, false
	}

	return RoutedCommand{
		Name:    commandName,
		Handler: handler,
	}, args[1:], true
}

// Run validates shared invocation shape, loads the selected index, and executes one handler.
func (rt Runtime) Run(args []string, stdout io.Writer, stderr io.Writer) runtimecontract.Status {
	if len(args) > 0 && args[0] == "--help" {
		if _, err := fmt.Fprint(stdout, helpText); err != nil {
			return runtimecontract.WriteDiagnostic(stderr, runtimecontract.UsageFailure(err.Error()))
		}

		return runtimecontract.StatusOK
	}

	if len(args) > 0 && args[0] == "--version" {
		if _, err := fmt.Fprintln(stdout, version.Format(rt.buildIdentity)); err != nil {
			return runtimecontract.WriteDiagnostic(stderr, runtimecontract.UsageFailure(err.Error()))
		}

		return runtimecontract.StatusOK
	}

	routed, commandArgs, ok := rt.Route(args)
	if !ok && len(args) == 0 {
		return runtimecontract.WriteDiagnostic(stderr, runtimecontract.UsageFailure("missing command"))
	}
	if !ok {
		return runtimecontract.WriteDiagnostic(
			stderr,
			runtimecontract.UsageFailure(fmt.Sprintf("unsupported command: %s", args[0])),
		)
	}

	indexPath, handlerArgs, diagnostic, ok := parseSharedArgs(commandArgs)
	if !ok {
		return runtimecontract.WriteDiagnostic(stderr, runtimecontract.UsageFailure(diagnostic))
	}

	loadedIndex, err := rt.loader.Load(indexPath)
	if err != nil {
		return runtimecontract.WriteDiagnostic(stderr, runtimecontract.IndexLoadFailure(err.Error()))
	}

	result, err := routed.Handler.Handle(loadedIndex, handlerArgs)
	if err != nil {
		return runtimecontract.WriteDiagnostic(stderr, runtimecontract.UsageFailure(err.Error()))
	}

	status, err := writeSuccess(stdout, result)
	if err != nil {
		return runtimecontract.WriteDiagnostic(stderr, runtimecontract.UsageFailure(err.Error()))
	}

	return status
}

func writeSuccess(stdout io.Writer, result any) (runtimecontract.Status, error) {
	if raw, ok := result.(runtimecontract.RawOutput); ok {
		return runtimecontract.WriteRawSuccess(stdout, raw.Text)
	}

	return runtimecontract.WriteJSONSuccess(stdout, result)
}

func parseSharedArgs(args []string) (string, []string, string, bool) {
	var indexPath string
	hasIndex := false
	handlerArgs := make([]string, 0, len(args))

	for position := 0; position < len(args); position++ {
		arg := args[position]
		if arg == "--index" {
			if position+1 >= len(args) || strings.HasPrefix(args[position+1], "--") {
				return "", nil, "--index requires a value", false
			}

			indexPath = args[position+1]
			hasIndex = true
			position++
			continue
		}

		if isObviousAdditionalCommand(args, position) {
			return "", nil, fmt.Sprintf("additional command token is not supported: %s", arg), false
		}

		handlerArgs = append(handlerArgs, arg)
	}

	if !hasIndex {
		return "", nil, "missing --index", false
	}

	return indexPath, handlerArgs, "", true
}

func isObviousAdditionalCommand(args []string, position int) bool {
	return isDocumentedCommand(args[position]) && (position == 0 || !strings.HasPrefix(args[position-1], "--"))
}

func isDocumentedCommand(arg string) bool {
	for _, command := range documentedCommands {
		if arg == command {
			return true
		}
	}

	return false
}
