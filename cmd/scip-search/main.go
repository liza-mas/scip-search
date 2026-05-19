package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"scip-search/internal/cli"
	"scip-search/internal/query/discovery"
	"scip-search/internal/query/implementations"
	"scip-search/internal/query/oneline"
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
	names, outputMode, err := parseSymbolArgs(args)
	if err != nil {
		return nil, err
	}
	view := traversal.NewView(loaded)
	switch outputMode {
	case symbolsOutputOneLine:
		text, err := discovery.OneLineSymbolsByNames(view, names)
		if err != nil {
			return nil, err
		}
		return runtimecontract.RawOutput{Text: text}, nil
	case symbolsOutputJSON:
		return discovery.FlatSymbolsByNames(view, names)
	case symbolsOutputNestedJSON:
		return discovery.SymbolsByNames(view, names)
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

func parseSymbolArgs(args []string) ([]string, symbolsOutputMode, error) {
	if duplicateFlag(args, "--one-line") {
		return nil, symbolsOutputOneLine, errors.New("--one-line can only be provided once")
	}
	if duplicateFlag(args, "--nested-json") {
		return nil, symbolsOutputOneLine, errors.New("--nested-json can only be provided once")
	}
	if duplicateFlag(args, "--json") {
		return nil, symbolsOutputOneLine, errors.New("--json can only be provided once")
	}

	var names []string
	outputMode := symbolsOutputOneLine
	hasOutputMode := false
	for position := 0; position < len(args); position++ {
		arg := args[position]
		switch arg {
		case "--name":
			if position+1 >= len(args) || isMissingQueryValue(args[position+1]) {
				return nil, symbolsOutputOneLine, errors.New("--name requires a value")
			}
			names = append(names, args[position+1])
			position++
		case "--one-line":
			if hasOutputMode {
				return nil, symbolsOutputOneLine, errors.New("symbols output flags are mutually exclusive")
			}
			outputMode = symbolsOutputOneLine
			hasOutputMode = true
		case "--nested-json":
			if hasOutputMode {
				return nil, symbolsOutputOneLine, errors.New("symbols output flags are mutually exclusive")
			}
			outputMode = symbolsOutputNestedJSON
			hasOutputMode = true
		case "--json":
			if hasOutputMode {
				return nil, symbolsOutputOneLine, errors.New("symbols output flags are mutually exclusive")
			}
			outputMode = symbolsOutputJSON
			hasOutputMode = true
		case "--flat":
			return nil, symbolsOutputOneLine, errors.New("--flat was renamed to --json")
		default:
			return nil, symbolsOutputOneLine, errors.New("symbols only accepts --name, --one-line, --nested-json, and --json")
		}
	}

	if len(names) == 0 {
		return nil, symbolsOutputOneLine, errors.New("missing --name")
	}

	return names, outputMode, nil
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
	symbols, usesName, outputMode, err := parseAndResolveSymbolSet(traversal.NewView(loaded), args, "implementations")
	if err != nil {
		return nil, err
	}

	if !usesName && len(symbols) == 1 {
		payload, err := implementations.Implementations(traversal.NewView(loaded), symbols[0])
		if err != nil {
			return nil, err
		}
		if outputMode == outputJSON {
			return payload, nil
		}

		return runtimecontract.RawOutput{Text: implementations.OneLine(payload)}, nil
	}

	payload, err := implementationQueries(traversal.NewView(loaded), symbols)
	if err != nil {
		return nil, err
	}
	if outputMode == outputJSON {
		return payload, nil
	}

	return runtimecontract.RawOutput{Text: oneLineImplementationQueries(payload)}, nil
}

type referencesHandler struct{}

func (referencesHandler) Handle(loadedIndex any, args []string) (any, error) {
	loaded, ok := loadedIndex.(runtimecontract.LoadedIndex)
	if !ok {
		return nil, errors.New("references handler received non-SCIP loaded index")
	}
	symbols, usesName, outputMode, err := parseAndResolveSymbolSet(traversal.NewView(loaded), args, "references")
	if err != nil {
		return nil, err
	}

	if !usesName && len(symbols) == 1 {
		payload := references.Query(traversal.NewView(loaded), symbols[0])
		if outputMode == outputJSON {
			return payload, nil
		}

		return runtimecontract.RawOutput{Text: references.OneLine(payload)}, nil
	}

	payload := referenceQueries(traversal.NewView(loaded), symbols)
	if outputMode == outputJSON {
		return payload, nil
	}

	return runtimecontract.RawOutput{Text: oneLineReferenceQueries(payload)}, nil
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

type symbolSetArgs struct {
	symbols []string
	names   []string
}

type referenceQueriesPayload struct {
	Symbols []string             `json:"symbols"`
	Queries []references.Payload `json:"queries"`
}

type implementationQueriesPayload struct {
	Symbols []string                  `json:"symbols"`
	Queries []implementations.Payload `json:"queries"`
}

type referenceOutputKey struct {
	symbol       string
	documentPath string
	roles        int32
	rangeLength  int
	scipRange    [4]int32
}

func parseAndResolveSymbolSet(
	view traversal.View,
	args []string,
	command string,
) ([]string, bool, oneLineJSONOutputMode, error) {
	queryArgs, outputMode, err := parseSymbolSetArgs(args, command)
	if err != nil {
		return nil, false, outputOneLine, err
	}

	symbols, err := resolveSymbolSet(view, queryArgs)
	if err != nil {
		return nil, false, outputOneLine, err
	}

	return symbols, len(queryArgs.names) > 0, outputMode, nil
}

func parseSymbolSetArgs(args []string, command string) (symbolSetArgs, oneLineJSONOutputMode, error) {
	if duplicateFlag(args, "--one-line") {
		return symbolSetArgs{}, outputOneLine, errors.New("--one-line can only be provided once")
	}
	if duplicateFlag(args, "--json") {
		return symbolSetArgs{}, outputOneLine, errors.New("--json can only be provided once")
	}

	var queryArgs symbolSetArgs
	outputMode := outputOneLine
	hasOutputMode := false
	for position := 0; position < len(args); position++ {
		arg := args[position]
		switch arg {
		case "--symbol":
			if position+1 >= len(args) || isMissingQueryValue(args[position+1]) {
				return symbolSetArgs{}, outputOneLine, errors.New("--symbol requires a value")
			}
			queryArgs.symbols = append(queryArgs.symbols, args[position+1])
			position++
		case "--name":
			if position+1 >= len(args) || isMissingQueryValue(args[position+1]) {
				return symbolSetArgs{}, outputOneLine, errors.New("--name requires a value")
			}
			queryArgs.names = append(queryArgs.names, args[position+1])
			position++
		case "--one-line":
			if hasOutputMode {
				return symbolSetArgs{}, outputOneLine, errors.New(command + " output flags are mutually exclusive")
			}
			outputMode = outputOneLine
			hasOutputMode = true
		case "--json":
			if hasOutputMode {
				return symbolSetArgs{}, outputOneLine, errors.New(command + " output flags are mutually exclusive")
			}
			outputMode = outputJSON
			hasOutputMode = true
		default:
			return symbolSetArgs{}, outputOneLine, errors.New(command + " only accepts --symbol, --name, --one-line, and --json")
		}
	}

	if len(queryArgs.symbols) == 0 && len(queryArgs.names) == 0 {
		return symbolSetArgs{}, outputOneLine, errors.New("missing --symbol or --name")
	}

	return queryArgs, outputMode, nil
}

func resolveSymbolSet(view traversal.View, queryArgs symbolSetArgs) ([]string, error) {
	symbols := make([]string, 0, 1)
	symbols = append(symbols, queryArgs.symbols...)
	for _, name := range queryArgs.names {
		discovered, err := discovery.FlatSymbolsByName(view, name)
		if err != nil {
			return nil, err
		}
		for _, symbol := range discovered.Symbols {
			symbols = append(symbols, symbol.Symbol)
		}
	}

	return uniqueSortedStrings(symbols), nil
}

func uniqueSortedStrings(values []string) []string {
	slices.Sort(values)
	return slices.Compact(values)
}

func referenceQueries(view traversal.View, symbols []string) referenceQueriesPayload {
	queries := make([]references.Payload, 0, len(symbols))
	for _, symbol := range symbols {
		queries = append(queries, references.Query(view, symbol))
	}
	return referenceQueriesPayload{
		Symbols: symbols,
		Queries: queries,
	}
}

func implementationQueries(view traversal.View, symbols []string) (implementationQueriesPayload, error) {
	queries := make([]implementations.Payload, 0, len(symbols))
	for _, symbol := range symbols {
		payload, err := implementations.Implementations(view, symbol)
		if err != nil {
			return implementationQueriesPayload{}, err
		}
		queries = append(queries, payload)
	}
	return implementationQueriesPayload{
		Symbols: symbols,
		Queries: queries,
	}, nil
}

func oneLineReferenceQueries(payload referenceQueriesPayload) string {
	var builder strings.Builder
	seen := map[referenceOutputKey]struct{}{}
	for _, query := range payload.Queries {
		for _, reference := range query.References {
			key := keyReferenceOutput(reference)
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}

			path, line, column := oneline.Location(reference.DocumentPath, reference.Range)
			fmt.Fprintf(
				&builder,
				"%s:%d:%d:%s roles=%d\n",
				path,
				line,
				column,
				reference.Symbol,
				reference.Roles,
			)
		}
	}
	return builder.String()
}

func keyReferenceOutput(reference references.Reference) referenceOutputKey {
	key := referenceOutputKey{
		symbol:       reference.Symbol,
		documentPath: reference.DocumentPath,
		roles:        reference.Roles,
		rangeLength:  len(reference.Range),
	}
	for index, value := range reference.Range {
		if index >= len(key.scipRange) {
			break
		}
		key.scipRange[index] = value
	}
	return key
}

func oneLineImplementationQueries(payload implementationQueriesPayload) string {
	var builder strings.Builder
	for _, query := range payload.Queries {
		builder.WriteString(implementations.OneLine(query))
	}
	return builder.String()
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
