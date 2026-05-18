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
	name, outputMode, err := parseSymbolArgs(args)
	if err != nil {
		return nil, err
	}
	view := traversal.NewView(loaded)
	switch outputMode {
	case symbolsOutputOneLine:
		text, err := discovery.OneLineSymbolsByName(view, name)
		if err != nil {
			return nil, err
		}
		return runtimecontract.RawOutput{Text: text}, nil
	case symbolsOutputJSON:
		return discovery.FlatSymbolsByName(view, name)
	case symbolsOutputNestedJSON:
		return discovery.SymbolsByName(view, name)
	default:
		return nil, errors.New("unknown symbols output mode")
	}
}

type symbolsOutputMode int

const (
	symbolsOutputOneLine symbolsOutputMode = iota
	symbolsOutputNestedJSON
	symbolsOutputJSON
)

func parseSymbolArgs(args []string) (string, symbolsOutputMode, error) {
	if duplicateFlag(args, "--name") {
		return "", symbolsOutputOneLine, errors.New("--name can only be provided once")
	}
	if duplicateFlag(args, "--one-line") {
		return "", symbolsOutputOneLine, errors.New("--one-line can only be provided once")
	}
	if duplicateFlag(args, "--nested-json") {
		return "", symbolsOutputOneLine, errors.New("--nested-json can only be provided once")
	}
	if duplicateFlag(args, "--json") {
		return "", symbolsOutputOneLine, errors.New("--json can only be provided once")
	}

	var name string
	hasName := false
	outputMode := symbolsOutputOneLine
	hasOutputMode := false
	for position := 0; position < len(args); position++ {
		arg := args[position]
		switch arg {
		case "--name":
			if position+1 >= len(args) || isMissingQueryValue(args[position+1]) {
				return "", symbolsOutputOneLine, errors.New("--name requires a value")
			}
			name = args[position+1]
			hasName = true
			position++
		case "--one-line":
			if hasOutputMode {
				return "", symbolsOutputOneLine, errors.New("symbols output flags are mutually exclusive")
			}
			outputMode = symbolsOutputOneLine
			hasOutputMode = true
		case "--nested-json":
			if hasOutputMode {
				return "", symbolsOutputOneLine, errors.New("symbols output flags are mutually exclusive")
			}
			outputMode = symbolsOutputNestedJSON
			hasOutputMode = true
		case "--json":
			if hasOutputMode {
				return "", symbolsOutputOneLine, errors.New("symbols output flags are mutually exclusive")
			}
			outputMode = symbolsOutputJSON
			hasOutputMode = true
		case "--flat":
			return "", symbolsOutputOneLine, errors.New("--flat was renamed to --json")
		default:
			return "", symbolsOutputOneLine, errors.New("symbols only accepts --name, --one-line, --nested-json, and --json")
		}
	}

	if !hasName {
		return "", symbolsOutputOneLine, errors.New("missing --name")
	}

	return name, outputMode, nil
}

type oneLineJSONOutputMode int

const (
	outputOneLine oneLineJSONOutputMode = iota
	outputJSON
)

type packagesHandler struct{}

func (packagesHandler) Handle(loadedIndex any, args []string) (any, error) {
	loaded, ok := loadedIndex.(runtimecontract.LoadedIndex)
	if !ok {
		return nil, errors.New("packages handler received non-SCIP loaded index")
	}
	prefix, outputMode, err := parsePackagePrefixArgs(args)
	if err != nil {
		return nil, err
	}

	payload, err := discovery.Packages(traversal.NewView(loaded), prefix)
	if err != nil {
		return nil, err
	}
	if outputMode == outputJSON {
		return payload, nil
	}

	return runtimecontract.RawOutput{Text: discovery.OneLinePackages(payload)}, nil
}

type implementationsHandler struct{}

func (implementationsHandler) Handle(loadedIndex any, args []string) (any, error) {
	loaded, ok := loadedIndex.(runtimecontract.LoadedIndex)
	if !ok {
		return nil, errors.New("implementations handler received non-SCIP loaded index")
	}
	symbol, outputMode, err := parseExactSymbolArgs(args, "implementations")
	if err != nil {
		return nil, err
	}

	payload, err := implementations.Implementations(traversal.NewView(loaded), symbol)
	if err != nil {
		return nil, err
	}
	if outputMode == outputJSON {
		return payload, nil
	}

	return runtimecontract.RawOutput{Text: implementations.OneLine(payload)}, nil
}

type referencesHandler struct{}

func (referencesHandler) Handle(loadedIndex any, args []string) (any, error) {
	loaded, ok := loadedIndex.(runtimecontract.LoadedIndex)
	if !ok {
		return nil, errors.New("references handler received non-SCIP loaded index")
	}
	symbol, outputMode, err := parseExactSymbolArgs(args, "references")
	if err != nil {
		return nil, err
	}

	payload := references.Query(traversal.NewView(loaded), symbol)
	if outputMode == outputJSON {
		return payload, nil
	}

	return runtimecontract.RawOutput{Text: references.OneLine(payload)}, nil
}

func parsePackagePrefixArgs(args []string) (string, oneLineJSONOutputMode, error) {
	if duplicateFlag(args, "--prefix") {
		return "", outputOneLine, errors.New("--prefix can only be provided once")
	}
	if duplicateFlag(args, "--one-line") {
		return "", outputOneLine, errors.New("--one-line can only be provided once")
	}
	if duplicateFlag(args, "--json") {
		return "", outputOneLine, errors.New("--json can only be provided once")
	}

	var prefix string
	outputMode := outputOneLine
	hasOutputMode := false
	for position := 0; position < len(args); position++ {
		arg := args[position]
		switch arg {
		case "--prefix":
			if position+1 >= len(args) || isMissingQueryValue(args[position+1]) {
				return "", outputOneLine, errors.New("--prefix requires a value")
			}
			prefix = args[position+1]
			position++
		case "--one-line":
			if hasOutputMode {
				return "", outputOneLine, errors.New("packages output flags are mutually exclusive")
			}
			outputMode = outputOneLine
			hasOutputMode = true
		case "--json":
			if hasOutputMode {
				return "", outputOneLine, errors.New("packages output flags are mutually exclusive")
			}
			outputMode = outputJSON
			hasOutputMode = true
		default:
			return "", outputOneLine, errors.New("packages only accepts --prefix, --one-line, and --json")
		}
	}

	return prefix, outputMode, nil
}

func parseExactSymbolArgs(args []string, command string) (string, oneLineJSONOutputMode, error) {
	return parseRequiredQueryValueWithOutput(args, "--symbol", command)
}

func parseRequiredQueryValueWithOutput(args []string, flag string, command string) (string, oneLineJSONOutputMode, error) {
	if len(args) == 0 {
		return "", outputOneLine, errors.New("missing " + flag)
	}
	if duplicateFlag(args, flag) {
		return "", outputOneLine, errors.New(flag + " can only be provided once")
	}
	if duplicateFlag(args, "--one-line") {
		return "", outputOneLine, errors.New("--one-line can only be provided once")
	}
	if duplicateFlag(args, "--json") {
		return "", outputOneLine, errors.New("--json can only be provided once")
	}

	var value string
	hasValue := false
	outputMode := outputOneLine
	hasOutputMode := false
	for position := 0; position < len(args); position++ {
		arg := args[position]
		switch arg {
		case flag:
			if position+1 >= len(args) || isMissingQueryValue(args[position+1]) {
				return "", outputOneLine, errors.New(flag + " requires a value")
			}
			value = args[position+1]
			hasValue = true
			position++
		case "--one-line":
			if hasOutputMode {
				return "", outputOneLine, errors.New(command + " output flags are mutually exclusive")
			}
			outputMode = outputOneLine
			hasOutputMode = true
		case "--json":
			if hasOutputMode {
				return "", outputOneLine, errors.New(command + " output flags are mutually exclusive")
			}
			outputMode = outputJSON
			hasOutputMode = true
		default:
			return "", outputOneLine, errors.New(command + " only accepts " + flag + ", --one-line, and --json")
		}
	}

	if !hasValue {
		return "", outputOneLine, errors.New("missing " + flag)
	}

	return value, outputMode, nil
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
