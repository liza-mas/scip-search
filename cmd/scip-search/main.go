package main

import (
	"errors"
	"io"
	"os"

	"scip-search/internal/cli"
	runtimecontract "scip-search/internal/runtime"
)

func main() {
	os.Exit(int(run(os.Args[1:], os.Stdout, os.Stderr)))
}

func run(args []string, stdout io.Writer, stderr io.Writer) runtimecontract.Status {
	cliRuntime := cli.NewRuntime(unimplementedLoader{}, map[string]cli.Handler{
		"symbols":         unimplementedHandler{},
		"references":      unimplementedHandler{},
		"implementations": unimplementedHandler{},
		"packages":        unimplementedHandler{},
	})

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
