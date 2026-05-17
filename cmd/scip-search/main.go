package main

import (
	"errors"
	"io"
	"os"

	"scip-search/internal/cli"
	runtimecontract "scip-search/internal/runtime"
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
	cliRuntime := cli.NewRuntimeWithBuildIdentity(unimplementedLoader{}, map[string]cli.Handler{
		"symbols":         unimplementedHandler{},
		"references":      unimplementedHandler{},
		"implementations": unimplementedHandler{},
		"packages":        unimplementedHandler{},
	}, buildIdentity)

	return cliRuntime.Run(args, stdout, stderr)
}

type unimplementedLoader struct{}

func (unimplementedLoader) Load(_ string) (any, error) {
	return nil, errors.New("index loading is not implemented")
}

type unimplementedHandler struct{}

func (unimplementedHandler) Handle(_ any, _ []string) (any, error) {
	return nil, errors.New("query execution is not implemented")
}
