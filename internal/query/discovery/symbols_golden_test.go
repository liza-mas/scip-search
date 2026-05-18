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

			stdout := runSymbolsCommand(t, test.queryName)
			got := decodeJSONValue(t, stdout)
			want := readGoldenJSONValue(t, test.goldenFile)

			if !reflect.DeepEqual(got, want) {
				t.Fatalf("symbols JSON value = %#v, want golden %#v", got, want)
			}
			assertGoldenSymbolsPayload(t, got, test.wantSymbols)
		})
	}
}

func runSymbolsCommand(t *testing.T, name string) []byte {
	t.Helper()

	fixture := loadDiscoveryFixture(t)
	runtime := cli.NewProductionRuntime(map[string]cli.Handler{
		"symbols": symbolsCommandHandler{},
	})
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	status := runtime.Run([]string{"symbols", "--index", fixture.IndexPath, "--name", name}, &stdout, &stderr)
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
	name, err := parseSymbolNameArg(args)
	if err != nil {
		return nil, err
	}

	return discovery.SymbolsByName(traversal.NewView(loaded), name)
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

func assertGoldenSymbolsPayload(t *testing.T, payload any, wantSymbols []string) {
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

	gotSymbols := collectGoldenSymbols(t, symbols)
	if !slices.Equal(gotSymbols, wantSymbols) {
		t.Fatalf("symbols = %v, want %v", gotSymbols, wantSymbols)
	}
	if !slices.IsSorted(gotSymbols) {
		t.Fatalf("symbols = %v, want stable ascending order by exact symbol", gotSymbols)
	}
	if len(wantSymbols) == 0 && len(symbols) != 0 {
		t.Fatalf("symbols = %#v, want explicit empty collection", symbols)
	}

	for _, entry := range symbols {
		assertRequiredSymbolFields(t, entry)
	}
}

func collectGoldenSymbols(t *testing.T, symbols []any) []string {
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

func assertRequiredSymbolFields(t *testing.T, entry any) {
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
