package cli

import (
	"fmt"
	"io"

	runtimecontract "scip-search/internal/runtime"
)

var documentedCommands = []string{
	"symbols",
	"references",
	"implementations",
	"packages",
}

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
	loader   Loader
	registry map[string]Handler
}

// RoutedCommand is the documented command selected from the shared registry.
type RoutedCommand struct {
	Name    string
	Handler Handler
}

// NewRuntime builds a runtime with exactly the documented query command names.
func NewRuntime(loader Loader, handlers map[string]Handler) Runtime {
	registry := make(map[string]Handler, len(documentedCommands))
	for _, name := range documentedCommands {
		handler, ok := handlers[name]
		if ok && handler != nil {
			registry[name] = handler
		}
	}

	return Runtime{
		loader:   loader,
		registry: registry,
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

// Run reports command routing usage failures before loader or handler work.
func (rt Runtime) Run(args []string, stdout io.Writer, stderr io.Writer) runtimecontract.Status {
	_, _, ok := rt.Route(args)
	if ok {
		return runtimecontract.StatusOK
	}

	if len(args) == 0 {
		return runtimecontract.WriteDiagnostic(stderr, runtimecontract.UsageFailure("missing command"))
	}

	return runtimecontract.WriteDiagnostic(
		stderr,
		runtimecontract.UsageFailure(fmt.Sprintf("unsupported command: %s", args[0])),
	)
}
