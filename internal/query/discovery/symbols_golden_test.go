package discovery_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"scip-search/internal/cli"
	"scip-search/internal/query/discovery"
	runtimecontract "scip-search/internal/runtime"
	"scip-search/internal/traversal"
)

func TestSymbolsCommandGoldenJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		queryName   string
		goldenFile  string
		wantSymbols []string
	}{
		{
			name:       "supervisor matches",
			queryName:  "Supervisor",
			goldenFile: "symbols-supervisor.json",
			wantSymbols: []string{
				discoverySupervisorAgentSymbol,
				discoverySupervisorSymbol,
				discoverySupervisorConfigSymbol,
			},
		},
		{
			name:        "run match",
			queryName:   "Run",
			goldenFile:  "symbols-run.json",
			wantSymbols: []string{discoveryRunSymbol},
		},
		{
			name:        "missing name",
			queryName:   "DoesNotExist",
			goldenFile:  "symbols-does-not-exist.json",
			wantSymbols: []string{},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			stdout := runSymbolsCommand(t, test.queryName, "--nested-json")
			got := decodeJSONValue(t, stdout)
			want := readGoldenJSONValue(t, test.goldenFile)

			if !reflect.DeepEqual(got, want) {
				t.Fatalf("symbols JSON value = %#v, want golden %#v", got, want)
			}
			assertGoldenCompactSymbolsPayload(t, got, test.wantSymbols)
		})
	}
}

func TestSymbolsJSONCommandPreservesSelfContainedPayload(t *testing.T) {
	t.Parallel()

	stdout := runSymbolsCommand(t, "Supervisor", "--json")
	got := decodeJSONValue(t, stdout)

	assertGoldenFlatSymbolsPayload(t, got, []string{
		discoverySupervisorAgentSymbol,
		discoverySupervisorSymbol,
		discoverySupervisorConfigSymbol,
	})
}

func runSymbolsCommand(t *testing.T, name string, extraArgs ...string) []byte {
	t.Helper()

	fixture := loadDiscoveryFixture(t)
	runtime := cli.NewProductionRuntime(map[string]cli.Handler{
		"symbols": symbolsCommandHandler{},
	})
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	args := []string{"symbols", "--index", fixture.IndexPath, "--name", name}
	args = append(args, extraArgs...)
	status := runtime.Run(args, &stdout, &stderr)
	if status != runtimecontract.StatusOK {
		t.Fatalf("symbols command status = %d, want %d; stderr = %q", status, runtimecontract.StatusOK, stderr.String())
	}
	if stderr.String() != "" {
		t.Fatalf("symbols command stderr = %q, want empty", stderr.String())
	}
	if stdout.Len() == 0 {
		t.Fatal("symbols command stdout is empty, want JSON payload")
	}

	return stdout.Bytes()
}

type symbolsCommandHandler struct{}

func (symbolsCommandHandler) Handle(loadedIndex any, args []string) (any, error) {
	loaded, ok := loadedIndex.(runtimecontract.LoadedIndex)
	if !ok {
		return nil, errors.New("symbols handler received non-SCIP loaded index")
	}
	name, outputMode, err := parseSymbolArgs(args)
	if err != nil {
		return nil, err
	}
	if outputMode == symbolsOutputJSON {
		return discovery.FlatSymbolsByName(traversal.NewView(loaded), name)
	}

	return discovery.SymbolsByName(traversal.NewView(loaded), name)
}

type symbolsOutputMode int

// Golden tests exercise only JSON symbol output modes; one-line formatting is
// covered in symbols_test.go and the production CLI tests.
const (
	symbolsOutputNestedJSON symbolsOutputMode = iota
	symbolsOutputJSON
)

func parseSymbolArgs(args []string) (string, symbolsOutputMode, error) {
	name := ""
	hasName := false
	outputMode := symbolsOutputNestedJSON
	for position := 0; position < len(args); position++ {
		switch args[position] {
		case "--nested-json":
			outputMode = symbolsOutputNestedJSON
		case "--json":
			outputMode = symbolsOutputJSON
		case "--name":
			if position+1 >= len(args) || args[position+1] == "" {
				return "", symbolsOutputNestedJSON, errors.New("--name requires a value")
			}
			name = args[position+1]
			hasName = true
			position++
		default:
			return "", symbolsOutputNestedJSON, errors.New("symbols only accepts --name, --nested-json, and --json")

		}
	}

	if !hasName {
		return "", symbolsOutputNestedJSON, errors.New("missing --name")
	}

	return name, outputMode, nil
}

func readGoldenJSONValue(t *testing.T, goldenFile string) any {
	t.Helper()

	payload, err := os.ReadFile(filepath.Join("testdata", "golden", goldenFile))
	if err != nil {
		t.Fatalf("read golden file %q: %v", goldenFile, err)
	}

	return decodeJSONValue(t, payload)
}

func decodeJSONValue(t *testing.T, payload []byte) any {
	t.Helper()

	decoder := json.NewDecoder(bytes.NewReader(payload))
	var decoded any
	if err := decoder.Decode(&decoded); err != nil {
		t.Fatalf("decode JSON payload %q: %v", payload, err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		t.Fatalf("JSON payload %q contains extra content: %v", payload, err)
	}

	return decoded
}

func assertGoldenCompactSymbolsPayload(t *testing.T, payload any, wantSymbols []string) {
	t.Helper()

	object, ok := payload.(map[string]any)
	if !ok {
		t.Fatalf("payload = %T, want top-level object", payload)
	}
	packagesValue, ok := object["packages"]
	if !ok {
		t.Fatalf("payload = %#v, want top-level packages collection", payload)
	}
	packages, ok := packagesValue.([]any)
	if !ok {
		t.Fatalf("packages = %T, want array", packagesValue)
	}

	gotSymbols := collectGoldenCompactSymbols(t, packages)
	if !slices.Equal(gotSymbols, wantSymbols) {
		t.Fatalf("symbols = %v, want %v", gotSymbols, wantSymbols)
	}
	if !slices.IsSorted(gotSymbols) {
		t.Fatalf("symbols = %v, want stable ascending order by exact symbol", gotSymbols)
	}
	if len(wantSymbols) == 0 && len(packages) != 0 {
		t.Fatalf("packages = %#v, want explicit empty collection", packages)
	}

	for _, entry := range packages {
		assertRequiredCompactPackageFields(t, entry)
	}
}

func assertGoldenFlatSymbolsPayload(t *testing.T, payload any, wantSymbols []string) {
	t.Helper()

	object, ok := payload.(map[string]any)
	if !ok {
		t.Fatalf("payload = %T, want top-level object", payload)
	}
	symbolsValue, ok := object["symbols"]
	if !ok {
		t.Fatalf("payload = %#v, want top-level symbols collection", payload)
	}
	symbols, ok := symbolsValue.([]any)
	if !ok {
		t.Fatalf("symbols = %T, want array", symbolsValue)
	}

	gotSymbols := collectGoldenFlatSymbols(t, symbols)
	if !slices.Equal(gotSymbols, wantSymbols) {
		t.Fatalf("symbols = %v, want %v", gotSymbols, wantSymbols)
	}
	for _, entry := range symbols {
		assertRequiredFlatSymbolFields(t, entry)
	}
}

func collectGoldenCompactSymbols(t *testing.T, packages []any) []string {
	t.Helper()

	collected := make([]string, 0)
	for _, entry := range packages {
		object, ok := entry.(map[string]any)
		if !ok {
			t.Fatalf("package entry = %T, want object", entry)
		}
		packageKey, ok := object["packageKey"].(string)
		if !ok || packageKey == "" {
			t.Fatalf("package entry = %#v, want non-empty packageKey string", object)
		}
		symbols, ok := object["symbols"].([]any)
		if !ok {
			t.Fatalf("package entry = %#v, want symbols array", object)
		}
		for _, symbolEntry := range symbols {
			symbolObject, ok := symbolEntry.(map[string]any)
			if !ok {
				t.Fatalf("symbol entry = %T, want object", symbolEntry)
			}
			descriptor, ok := symbolObject["descriptor"].(string)
			if !ok || descriptor == "" {
				t.Fatalf("symbol entry = %#v, want non-empty descriptor string", symbolObject)
			}
			collected = append(collected, packageKey+" "+descriptor)
		}
	}

	return collected
}

func collectGoldenFlatSymbols(t *testing.T, symbols []any) []string {
	t.Helper()

	collected := make([]string, 0, len(symbols))
	for _, entry := range symbols {
		object, ok := entry.(map[string]any)
		if !ok {
			t.Fatalf("symbol entry = %T, want object", entry)
		}
		symbol, ok := object["symbol"].(string)
		if !ok || symbol == "" {
			t.Fatalf("symbol entry = %#v, want non-empty symbol string", object)
		}
		collected = append(collected, symbol)
	}

	return collected
}

func assertRequiredCompactPackageFields(t *testing.T, entry any) {
	t.Helper()

	object := entry.(map[string]any)
	for _, field := range []string{
		"scheme",
		"packageManager",
		"packageName",
		"packageVersion",
		"packageKey",
	} {
		value, ok := object[field].(string)
		if !ok || value == "" {
			t.Fatalf("package entry = %#v, want required non-empty string field %q", object, field)
		}
	}
	symbols, ok := object["symbols"].([]any)
	if !ok {
		t.Fatalf("package entry = %#v, want symbols array", object)
	}
	for _, symbol := range symbols {
		assertRequiredCompactSymbolFields(t, symbol)
	}
}

func assertRequiredCompactSymbolFields(t *testing.T, entry any) {
	t.Helper()

	object := entry.(map[string]any)
	for _, field := range []string{
		"descriptor",
		"matchText",
		"matchSource",
	} {
		value, ok := object[field].(string)
		if !ok || value == "" {
			t.Fatalf("symbol entry = %#v, want required non-empty string field %q", object, field)
		}
	}
	assertOptionalDefinition(t, object)
}

func assertRequiredFlatSymbolFields(t *testing.T, entry any) {
	t.Helper()

	object := entry.(map[string]any)
	for _, field := range []string{
		"symbol",
		"scheme",
		"packageManager",
		"packageName",
		"packageVersion",
		"matchText",
		"matchSource",
	} {
		value, ok := object[field].(string)
		if !ok || value == "" {
			t.Fatalf("symbol entry = %#v, want required non-empty string field %q", object, field)
		}
	}

	assertOptionalDefinition(t, object)
}

func assertOptionalDefinition(t *testing.T, object map[string]any) {
	t.Helper()

	definition, hasDefinition := object["definition"]
	if !hasDefinition {
		return
	}
	definitionObject, ok := definition.(map[string]any)
	if !ok {
		t.Fatalf("definition = %T, want object", definition)
	}
	if documentPath, ok := definitionObject["documentPath"].(string); !ok || documentPath == "" {
		t.Fatalf("definition = %#v, want documentPath", definitionObject)
	}
	rangeValue, ok := definitionObject["range"].([]any)
	if !ok || len(rangeValue) == 0 {
		t.Fatalf("definition = %#v, want SCIP range array", definitionObject)
	}
}
