package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"time"

	"scip-search/internal/aggregate"
	"scip-search/internal/cli"
	"scip-search/internal/query/discovery"
	graphquery "scip-search/internal/query/graph"
	"scip-search/internal/query/graphexport"
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
	if len(args) > 0 && args[0] == "aggregate-index" {
		return runAggregateIndex(args[1:], stderr, buildIdentity)
	}

	cliRuntime := cli.NewProductionRuntimeWithBuildIdentity(map[string]cli.Handler{
		"symbols":         symbolsHandler{},
		"references":      referencesHandler{},
		"implementations": implementationsHandler{},
		"packages":        packagesHandler{},
		"graph":           graphHandler{kind: graphCommandGraph},
		"callers":         graphHandler{kind: graphCommandCallers},
		"callees":         graphHandler{kind: graphCommandCallees},
		"impact":          graphHandler{kind: graphCommandImpact},
		"graph-export":    graphExportHandler{buildIdentity: buildIdentity},
	}, buildIdentity)

	return cliRuntime.Run(args, stdout, stderr)
}

func runAggregateIndex(args []string, stderr io.Writer, buildIdentity version.BuildIdentity) runtimecontract.Status {
	options, err := aggregate.ParseArgs(args)
	if err != nil {
		return runtimecontract.WriteDiagnostic(stderr, runtimecontract.UsageFailure(err.Error()))
	}
	if _, err := aggregate.Run(options, buildIdentity); err != nil {
		var failure runtimecontract.Failure
		if errors.As(err, &failure) && failure.Status() == runtimecontract.StatusIndexLoad {
			return runtimecontract.WriteDiagnostic(stderr, runtimecontract.IndexLoadFailure(err.Error()))
		}
		if aggregate.IsValidationError(err) {
			return runtimecontract.WriteDiagnostic(stderr, runtimecontract.ValidationFailure(err.Error()))
		}
		return runtimecontract.WriteDiagnostic(stderr, runtimecontract.UsageFailure(err.Error()))
	}
	return runtimecontract.StatusOK
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
		return runtimecontract.RawOutput{Text: oneLineWithProjectRoot(view, text)}, nil
	case symbolsOutputJSON:
		payload, err := discovery.FlatSymbolsByNames(view, names)
		if err != nil {
			return nil, err
		}
		return withProjectRoot(view, payload), nil
	case symbolsOutputNestedJSON:
		payload, err := discovery.SymbolsByNames(view, names)
		if err != nil {
			return nil, err
		}
		return withProjectRoot(view, payload), nil
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

type queryOutputMode int

const (
	outputOneLine queryOutputMode = iota
	outputJSON
	outputLocationOnly
	outputMarkdown
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
			return withProjectRoot(traversal.NewView(loaded), payload), nil
		}
		if outputMode == outputLocationOnly {
			return runtimecontract.RawOutput{Text: implementations.LocationOnly(payload)}, nil
		}

		return runtimecontract.RawOutput{Text: oneLineWithProjectRoot(traversal.NewView(loaded), implementations.OneLine(payload))}, nil
	}

	payload, err := implementationQueries(traversal.NewView(loaded), symbols)
	if err != nil {
		return nil, err
	}
	if outputMode == outputJSON {
		return withProjectRoot(traversal.NewView(loaded), payload), nil
	}
	if outputMode == outputLocationOnly {
		return runtimecontract.RawOutput{Text: locationOnlyImplementationQueries(payload)}, nil
	}

	return runtimecontract.RawOutput{Text: oneLineWithProjectRoot(traversal.NewView(loaded), oneLineImplementationQueries(payload))}, nil
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
			return withProjectRoot(traversal.NewView(loaded), payload), nil
		}
		if outputMode == outputLocationOnly {
			return runtimecontract.RawOutput{Text: references.LocationOnly(payload)}, nil
		}

		return runtimecontract.RawOutput{Text: oneLineWithProjectRoot(traversal.NewView(loaded), references.OneLine(payload))}, nil
	}

	payload := referenceQueries(traversal.NewView(loaded), symbols)
	if outputMode == outputJSON {
		return withProjectRoot(traversal.NewView(loaded), payload), nil
	}
	if outputMode == outputLocationOnly {
		return runtimecontract.RawOutput{Text: locationOnlyReferenceQueries(payload)}, nil
	}

	return runtimecontract.RawOutput{Text: oneLineWithProjectRoot(traversal.NewView(loaded), oneLineReferenceQueries(payload))}, nil
}

func parsePackagePrefixArgs(args []string) (string, queryOutputMode, error) {
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

type graphQueriesPayload struct {
	Symbols []string             `json:"symbols"`
	Queries []graphquery.Payload `json:"queries"`
}

type impactQueriesPayload struct {
	Symbols []string                   `json:"symbols"`
	Queries []graphquery.ImpactPayload `json:"queries"`
}

type referenceOutputKey struct {
	symbol       string
	documentPath string
	roles        int32
	rangeLength  int
	scipRange    [4]int32
}

type graphCommandKind int

const (
	graphCommandGraph graphCommandKind = iota
	graphCommandCallers
	graphCommandCallees
	graphCommandImpact
)

type graphHandler struct {
	kind graphCommandKind
}

type graphExportHandler struct {
	buildIdentity version.BuildIdentity
}

func (handler graphExportHandler) Handle(loadedIndex any, args []string) (any, error) {
	loaded, ok := loadedIndex.(runtimecontract.LoadedIndex)
	if !ok {
		return nil, errors.New("graph-export handler received non-SCIP loaded index")
	}
	view := traversal.NewView(loaded)
	filters, outputPath, err := parseGraphExportArgs(view, args)
	if err != nil {
		return nil, err
	}
	payload := graphexport.ExportView(loaded, view, handler.buildIdentity, filters, time.Now)
	if outputPath != "" {
		return runtimecontract.JSONFileOutput{Path: outputPath, Value: payload}, nil
	}
	return payload, nil
}

func (handler graphHandler) Handle(loadedIndex any, args []string) (any, error) {
	loaded, ok := loadedIndex.(runtimecontract.LoadedIndex)
	if !ok {
		return nil, errors.New(handler.commandName() + " handler received non-SCIP loaded index")
	}
	view := traversal.NewView(loaded)
	symbols, usesName, outputMode, err := parseAndResolveGraphSymbolSet(view, args, handler.commandName())
	if err != nil {
		return nil, err
	}

	if handler.kind == graphCommandImpact {
		if !usesName && len(symbols) == 1 {
			payload := graphquery.Impact(view, symbols[0])
			return graphImpactOutput(view, payload, outputMode)
		}
		payload := impactQueries(view, symbols)
		return graphImpactQueriesOutput(view, payload, outputMode)
	}

	if !usesName && len(symbols) == 1 {
		payload := graphPayloadForKind(view, symbols[0], handler.kind)
		return graphOutput(view, payload, outputMode)
	}
	payload := graphQueries(view, symbols, handler.kind)
	return graphQueriesOutput(view, payload, outputMode)
}

func (handler graphHandler) commandName() string {
	switch handler.kind {
	case graphCommandCallers:
		return "callers"
	case graphCommandCallees:
		return "callees"
	case graphCommandImpact:
		return "impact"
	default:
		return "graph"
	}
}

func parseAndResolveGraphSymbolSet(
	view traversal.View,
	args []string,
	command string,
) ([]string, bool, queryOutputMode, error) {
	queryArgs, outputMode, err := parseGraphSymbolSetArgs(args, command)
	if err != nil {
		return nil, false, outputOneLine, err
	}
	symbols, err := resolveSymbolSet(view, queryArgs)
	if err != nil {
		return nil, false, outputOneLine, err
	}
	return symbols, len(queryArgs.names) > 0, outputMode, nil
}

func parseGraphSymbolSetArgs(args []string, command string) (symbolSetArgs, queryOutputMode, error) {
	if duplicateFlag(args, "--one-line") {
		return symbolSetArgs{}, outputOneLine, errors.New("--one-line can only be provided once")
	}
	if duplicateFlag(args, "--json") {
		return symbolSetArgs{}, outputOneLine, errors.New("--json can only be provided once")
	}
	if duplicateFlag(args, "--markdown") {
		return symbolSetArgs{}, outputOneLine, errors.New("--markdown can only be provided once")
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
		case "--markdown":
			if hasOutputMode {
				return symbolSetArgs{}, outputOneLine, errors.New(command + " output flags are mutually exclusive")
			}
			outputMode = outputMarkdown
			hasOutputMode = true
		default:
			return symbolSetArgs{}, outputOneLine, errors.New(command + " only accepts --symbol, --name, --one-line, --json, and --markdown")
		}
	}
	if len(queryArgs.symbols) == 0 && len(queryArgs.names) == 0 {
		return symbolSetArgs{}, outputOneLine, errors.New("missing --symbol or --name")
	}
	return queryArgs, outputMode, nil
}

func parseGraphExportArgs(view traversal.View, args []string) (graphexport.Filters, string, error) {
	if duplicateFlag(args, "-o") {
		return graphexport.Filters{}, "", errors.New("-o can only be provided once")
	}

	var queryArgs symbolSetArgs
	var packagePrefixes []string
	var outputPath string
	for position := 0; position < len(args); position++ {
		arg := args[position]
		switch arg {
		case "--symbol":
			if position+1 >= len(args) || isMissingGraphExportValue(args[position+1]) {
				return graphexport.Filters{}, "", errors.New("--symbol requires a value")
			}
			queryArgs.symbols = append(queryArgs.symbols, args[position+1])
			position++
		case "--name":
			if position+1 >= len(args) || isMissingGraphExportValue(args[position+1]) {
				return graphexport.Filters{}, "", errors.New("--name requires a value")
			}
			queryArgs.names = append(queryArgs.names, args[position+1])
			position++
		case "--package-prefix":
			if position+1 >= len(args) || isMissingGraphExportValue(args[position+1]) {
				return graphexport.Filters{}, "", errors.New("--package-prefix requires a value")
			}
			packagePrefixes = append(packagePrefixes, args[position+1])
			position++
		case "-o":
			if position+1 >= len(args) || isMissingGraphExportValue(args[position+1]) {
				return graphexport.Filters{}, "", errors.New("-o requires a value")
			}
			outputPath = args[position+1]
			position++
		case "--one-line", "--json", "--markdown", "--location-only":
			return graphexport.Filters{}, "", errors.New(arg + " is not supported for graph-export")
		default:
			return graphexport.Filters{}, "", errors.New("graph-export only accepts --symbol, --name, --package-prefix, and -o")
		}
	}

	symbols, err := resolveSymbolSet(view, queryArgs)
	if err != nil {
		return graphexport.Filters{}, "", err
	}
	return graphexport.Filters{
		Symbols:         symbols,
		PackagePrefixes: uniqueSortedStrings(packagePrefixes),
	}, outputPath, nil
}

func isMissingGraphExportValue(value string) bool {
	return isMissingQueryValue(value) || value == "-o"
}

func graphPayloadForKind(view traversal.View, symbol string, kind graphCommandKind) graphquery.Payload {
	payload := graphquery.Query(view, symbol)
	switch kind {
	case graphCommandCallers:
		return graphquery.Callers(payload)
	case graphCommandCallees:
		return graphquery.Callees(payload)
	default:
		return payload
	}
}

func graphQueries(view traversal.View, symbols []string, kind graphCommandKind) graphQueriesPayload {
	queries := make([]graphquery.Payload, 0, len(symbols))
	for _, symbol := range symbols {
		queries = append(queries, graphPayloadForKind(view, symbol, kind))
	}
	return graphQueriesPayload{Symbols: symbols, Queries: queries}
}

func impactQueries(view traversal.View, symbols []string) impactQueriesPayload {
	queries := make([]graphquery.ImpactPayload, 0, len(symbols))
	for _, symbol := range symbols {
		queries = append(queries, graphquery.Impact(view, symbol))
	}
	return impactQueriesPayload{Symbols: symbols, Queries: queries}
}

func graphOutput(view traversal.View, payload graphquery.Payload, outputMode queryOutputMode) (any, error) {
	switch outputMode {
	case outputJSON:
		return withProjectRoot(view, payload), nil
	case outputMarkdown:
		return runtimecontract.RawOutput{Text: markdownWithProjectRoot(view, graphquery.Markdown(payload))}, nil
	default:
		return runtimecontract.RawOutput{Text: oneLineWithProjectRoot(view, graphquery.OneLine(payload))}, nil
	}
}

func graphQueriesOutput(view traversal.View, payload graphQueriesPayload, outputMode queryOutputMode) (any, error) {
	switch outputMode {
	case outputJSON:
		return withProjectRoot(view, payload), nil
	case outputMarkdown:
		var builder strings.Builder
		for index, query := range payload.Queries {
			if index > 0 {
				builder.WriteString("\n")
			}
			builder.WriteString(graphquery.Markdown(query))
		}
		return runtimecontract.RawOutput{Text: markdownWithProjectRoot(view, builder.String())}, nil
	default:
		var builder strings.Builder
		seen := map[string]struct{}{}
		for _, query := range payload.Queries {
			for _, line := range strings.Split(graphquery.OneLine(query), "\n") {
				if line == "" {
					continue
				}
				if _, exists := seen[line]; exists {
					continue
				}
				seen[line] = struct{}{}
				builder.WriteString(line)
				builder.WriteString("\n")
			}
		}
		return runtimecontract.RawOutput{Text: oneLineWithProjectRoot(view, builder.String())}, nil
	}
}

func graphImpactOutput(view traversal.View, payload graphquery.ImpactPayload, outputMode queryOutputMode) (any, error) {
	switch outputMode {
	case outputJSON:
		return withProjectRoot(view, payload), nil
	case outputMarkdown:
		return runtimecontract.RawOutput{Text: markdownWithProjectRoot(view, graphquery.ImpactMarkdown(payload))}, nil
	default:
		return runtimecontract.RawOutput{Text: oneLineWithProjectRoot(view, graphquery.ImpactOneLine(payload))}, nil
	}
}

func graphImpactQueriesOutput(view traversal.View, payload impactQueriesPayload, outputMode queryOutputMode) (any, error) {
	switch outputMode {
	case outputJSON:
		return withProjectRoot(view, payload), nil
	case outputMarkdown:
		var builder strings.Builder
		for index, query := range payload.Queries {
			if index > 0 {
				builder.WriteString("\n")
			}
			builder.WriteString(graphquery.ImpactMarkdown(query))
		}
		return runtimecontract.RawOutput{Text: markdownWithProjectRoot(view, builder.String())}, nil
	default:
		var builder strings.Builder
		seen := map[string]struct{}{}
		for _, query := range payload.Queries {
			for _, line := range strings.Split(graphquery.ImpactOneLine(query), "\n") {
				if line == "" {
					continue
				}
				if _, exists := seen[line]; exists {
					continue
				}
				seen[line] = struct{}{}
				builder.WriteString(line)
				builder.WriteString("\n")
			}
		}
		return runtimecontract.RawOutput{Text: oneLineWithProjectRoot(view, builder.String())}, nil
	}
}

type projectRootPayload struct {
	projectRoot string
	payload     any
}

func withProjectRoot(view traversal.View, payload any) projectRootPayload {
	return projectRootPayload{projectRoot: view.Metadata().ProjectRoot, payload: payload}
}

func (payload projectRootPayload) MarshalJSON() ([]byte, error) {
	rawPayload, err := json.Marshal(payload.payload)
	if err != nil {
		return nil, err
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(rawPayload, &object); err != nil {
		return nil, err
	}
	projectRoot, err := json.Marshal(payload.projectRoot)
	if err != nil {
		return nil, err
	}
	object["project_root"] = projectRoot
	return json.Marshal(object)
}

func oneLineWithProjectRoot(view traversal.View, text string) string {
	if text == "" {
		return ""
	}
	return fmt.Sprintf("# project_root=%s\n%s", view.Metadata().ProjectRoot, text)
}

func markdownWithProjectRoot(view traversal.View, text string) string {
	if text == "" {
		return ""
	}
	return fmt.Sprintf("Project root: %s\n%s", view.Metadata().ProjectRoot, text)
}

func parseAndResolveSymbolSet(
	view traversal.View,
	args []string,
	command string,
) ([]string, bool, queryOutputMode, error) {
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

func parseSymbolSetArgs(args []string, command string) (symbolSetArgs, queryOutputMode, error) {
	if duplicateFlag(args, "--one-line") {
		return symbolSetArgs{}, outputOneLine, errors.New("--one-line can only be provided once")
	}
	if duplicateFlag(args, "--json") {
		return symbolSetArgs{}, outputOneLine, errors.New("--json can only be provided once")
	}
	if duplicateFlag(args, "--location-only") {
		return symbolSetArgs{}, outputOneLine, errors.New("--location-only can only be provided once")
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
		case "--location-only":
			if hasOutputMode {
				return symbolSetArgs{}, outputOneLine, errors.New(command + " output flags are mutually exclusive")
			}
			outputMode = outputLocationOnly
			hasOutputMode = true
		default:
			return symbolSetArgs{}, outputOneLine, errors.New(command + " only accepts --symbol, --name, --one-line, --json, and --location-only")
		}
	}

	if outputMode == outputLocationOnly && len(queryArgs.names) > 0 {
		return symbolSetArgs{}, outputOneLine, errors.New("--location-only cannot be used with --name")
	}
	if outputMode == outputLocationOnly && len(queryArgs.symbols) == 0 {
		return symbolSetArgs{}, outputOneLine, errors.New("--location-only requires --symbol")
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
				"%s:%d:%d symbol=%s; roles=%d\n",
				path,
				line,
				column,
				oneline.Quote(reference.Symbol),
				reference.Roles,
			)
		}
	}
	return builder.String()
}

func locationOnlyReferenceQueries(payload referenceQueriesPayload) string {
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
			fmt.Fprintf(&builder, "%s:%d:%d\n", path, line, column)
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

func locationOnlyImplementationQueries(payload implementationQueriesPayload) string {
	var builder strings.Builder
	for _, query := range payload.Queries {
		builder.WriteString(implementations.LocationOnly(query))
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
