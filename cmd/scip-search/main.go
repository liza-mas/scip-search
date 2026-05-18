package main

import (
	"errors"
	"io"
	"os"
	"strings"

	"scip-search/internal/cli"
	"scip-search/internal/query/discovery"
	"scip-search/internal/query/implementations"
	"scip-search/internal/query/references"
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
		"references":      referencesHandler{},
		"implementations": implementationsHandler{},
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

type implementationsHandler struct{}

func (implementationsHandler) Handle(loadedIndex any, args []string) (any, error) {
	loaded, ok := loadedIndex.(runtimecontract.LoadedIndex)
	if !ok {
		return nil, errors.New("implementations handler received non-SCIP loaded index")
	}
	symbol, err := parseExactSymbolArg(args)
	if err != nil {
		return nil, err
	}

	return implementations.Implementations(traversal.NewView(loaded), symbol)
}

type referencesHandler struct{}

func (referencesHandler) Handle(loadedIndex any, args []string) (any, error) {
	loaded, ok := loadedIndex.(runtimecontract.LoadedIndex)
	if !ok {
		return nil, errors.New("references handler received non-SCIP loaded index")
	}
	symbol, err := parseExactSymbolArg(args)
	if err != nil {
		return nil, err
	}

	return references.Query(traversal.NewView(loaded), symbol), nil
}

func parseSymbolNameArg(args []string) (string, error) {
	return parseRequiredQueryValue(args, "--name")
}

func parseExactSymbolArg(args []string) (string, error) {
	return parseRequiredQueryValue(args, "--symbol")
}

func parsePackagePrefixArg(args []string) (string, error) {
	if len(args) == 0 {
		return "", nil
	}
	if duplicateFlag(args, "--prefix") {
		return "", errors.New("--prefix can only be provided once")
	}
	if args[0] == "--prefix" && (len(args) == 1 || isMissingQueryValue(args[1])) {
		return "", errors.New("--prefix requires a value")
	}
	if len(args) != 2 {
		return "", errors.New("packages only accepts --prefix")
	}
	if args[0] != "--prefix" {
		return "", errors.New("packages only accepts --prefix")
	}
	if isMissingQueryValue(args[1]) {
		return "", errors.New("--prefix requires a value")
	}

	return args[1], nil
}

func parseRequiredQueryValue(args []string, flag string) (string, error) {
	if len(args) == 0 {
		return "", errors.New("missing " + flag)
	}
	if args[0] == flag && (len(args) == 1 || isMissingQueryValue(args[1])) {
		return "", errors.New(flag + " requires a value")
	}
	if len(args) != 2 || args[0] != flag || duplicateFlag(args, flag) {
		return "", errors.New(flag + " accepts exactly one value")
	}

	return args[1], nil
}

func duplicateFlag(args []string, flag string) bool {
	count := 0
	for _, arg := range args {
		if arg == flag {
			count++
		}
	}

	return count > 1
}

func isMissingQueryValue(value string) bool {
	return value == "" || strings.HasPrefix(value, "--")
}
