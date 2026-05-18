package main

import (
	"errors"
	"io"
	"os"

	"scip-search/internal/cli"
	"scip-search/internal/query/discovery"
	runtimecontract "scip-search/internal/runtime"
	"scip-search/internal/traversal"
	"scip-search/internal/version"
)

func main() {
	os.Exit(int(run(os.Args[1:], os.Stdout, os.Stderr)))
}

func run(args []string, stdout io.Writer, stderr io.Writer) runtimecontract.Status {
	return runWithBuildIdentity(args, stdout, stderr, version.Current())
}

func runWithBuildIdentity(
	args []string,
	stdout io.Writer,
	stderr io.Writer,
	buildIdentity version.BuildIdentity,
) runtimecontract.Status {
	cliRuntime := cli.NewProductionRuntimeWithBuildIdentity(map[string]cli.Handler{
		"symbols":         symbolsHandler{},
		"references":      unimplementedHandler{},
		"implementations": unimplementedHandler{},
		"packages":        packagesHandler{},
	}, buildIdentity)

	return cliRuntime.Run(args, stdout, stderr)
}

type symbolsHandler struct{}

func (symbolsHandler) Handle(loadedIndex any, args []string) (any, error) {
	loaded, ok := loadedIndex.(runtimecontract.LoadedIndex)
	if !ok {
		return nil, errors.New("symbols handler received non-SCIP loaded index")
	}
	name, err := parseSymbolNameArg(args)
	if err != nil {
		return nil, err
	}

	return discovery.SymbolsByName(traversal.NewView(loaded), name)
}

type packagesHandler struct{}

func (packagesHandler) Handle(loadedIndex any, args []string) (any, error) {
	loaded, ok := loadedIndex.(runtimecontract.LoadedIndex)
	if !ok {
		return nil, errors.New("packages handler received non-SCIP loaded index")
	}
	prefix, err := parsePackagePrefixArg(args)
	if err != nil {
		return nil, err
	}

	return discovery.Packages(traversal.NewView(loaded), prefix)
}

type unimplementedHandler struct{}

func (unimplementedHandler) Handle(_ any, _ []string) (any, error) {
	return nil, errors.New("query execution is not implemented")
}

func parseSymbolNameArg(args []string) (string, error) {
	for position := 0; position < len(args); position++ {
		if args[position] != "--name" {
			continue
		}
		if position+1 >= len(args) || args[position+1] == "" {
			return "", errors.New("--name requires a value")
		}

		return args[position+1], nil
	}

	return "", errors.New("missing --name")
}

func parsePackagePrefixArg(args []string) (string, error) {
	var prefix string
	for position := 0; position < len(args); position++ {
		if args[position] != "--prefix" {
			return "", errors.New("packages only accepts --prefix")
		}
		if position+1 >= len(args) || args[position+1] == "" {
			return "", errors.New("--prefix requires a value")
		}
		if prefix != "" {
			return "", errors.New("--prefix can only be provided once")
		}
		prefix = args[position+1]
		position++
	}

	return prefix, nil
}
