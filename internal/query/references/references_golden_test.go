package references_test

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
	"scip-search/internal/query/references"
	runtimecontract "scip-search/internal/runtime"
	"scip-search/internal/traversal"
	"scip-search/internal/traversal/traversaltest"
)

func TestReferencesCommandGoldenJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		symbol      string
		goldenFile  string
		wantSymbols []string
	}{
		{
			name:       "alpha exact and related references",
			symbol:     traversaltest.AlphaSymbol,
			goldenFile: "references-alpha.json",
			wantSymbols: []string{
				traversaltest.AlphaSymbol,
				traversaltest.BetaSymbol,
			},
		},
		{
			name:        "absent exact symbol",
			symbol:      "scip-go gomod example.com/fixture . missing/Absent#",
			goldenFile:  "references-absent.json",
			wantSymbols: []string{},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			stdout := runReferencesCommand(t, test.symbol)
			got := decodeJSONValue(t, stdout)
			want := readGoldenJSONValue(t, test.goldenFile)

			if !reflect.DeepEqual(got, want) {
				t.Fatalf("references JSON value = %#v, want golden %#v", got, want)
			}
			assertGoldenReferencesPayload(t, got, test.symbol, test.wantSymbols)
		})
	}
}

func runReferencesCommand(t *testing.T, symbol string) []byte {
	t.Helper()

	fixture := traversaltest.LoadSharedFixture(t)
	runtime := cli.NewProductionRuntime(map[string]cli.Handler{
		"references": referencesCommandHandler{},
	})
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	status := runtime.Run([]string{"references", "--index", fixture.IndexPath, "--symbol", symbol}, &stdout, &stderr)
	if status != runtimecontract.StatusOK {
		t.Fatalf("references command status = %d, want %d; stderr = %q", status, runtimecontract.StatusOK, stderr.String())
	}
	if stderr.String() != "" {
		t.Fatalf("references command stderr = %q, want empty", stderr.String())
	}
	if stdout.Len() == 0 {
		t.Fatal("references command stdout is empty, want JSON payload")
	}

	return stdout.Bytes()
}

type referencesCommandHandler struct{}

func (referencesCommandHandler) Handle(loadedIndex any, args []string) (any, error) {
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

func parseExactSymbolArg(args []string) (string, error) {
	for position := 0; position < len(args); position++ {
		if args[position] != "--symbol" {
			continue
		}
		if position+1 >= len(args) || args[position+1] == "" {
			return "", errors.New("--symbol requires a value")
		}

		return args[position+1], nil
	}

	return "", errors.New("missing --symbol")
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

func assertGoldenReferencesPayload(t *testing.T, payload any, wantSymbol string, wantSymbols []string) {
	t.Helper()

	object, ok := payload.(map[string]any)
	if !ok {
		t.Fatalf("payload = %T, want top-level object", payload)
	}
	if got, ok := object["symbol"].(string); !ok || got != wantSymbol {
		t.Fatalf("payload = %#v, want top-level symbol %q", payload, wantSymbol)
	}
	referencesValue, ok := object["references"]
	if !ok {
		t.Fatalf("payload = %#v, want top-level references collection", payload)
	}
	referenceEntries, ok := referencesValue.([]any)
	if !ok {
		t.Fatalf("references = %T, want array", referencesValue)
	}

	gotSymbols := collectGoldenReferenceSymbols(t, referenceEntries)
	if !slices.Equal(gotSymbols, wantSymbols) {
		t.Fatalf("reference symbols = %v, want %v", gotSymbols, wantSymbols)
	}
	if len(wantSymbols) == 0 && len(referenceEntries) != 0 {
		t.Fatalf("references = %#v, want explicit empty collection", referenceEntries)
	}

	for _, entry := range referenceEntries {
		assertRequiredReferenceFields(t, entry)
	}
}

func collectGoldenReferenceSymbols(t *testing.T, references []any) []string {
	t.Helper()

	collected := make([]string, 0, len(references))
	for _, entry := range references {
		object, ok := entry.(map[string]any)
		if !ok {
			t.Fatalf("reference entry = %T, want object", entry)
		}
		symbol, ok := object["symbol"].(string)
		if !ok || symbol == "" {
			t.Fatalf("reference entry = %#v, want non-empty symbol string", object)
		}
		collected = append(collected, symbol)
	}

	return collected
}

func assertRequiredReferenceFields(t *testing.T, entry any) {
	t.Helper()

	object := entry.(map[string]any)
	for _, field := range []string{"symbol", "documentPath"} {
		value, ok := object[field].(string)
		if !ok || value == "" {
			t.Fatalf("reference entry = %#v, want required non-empty string field %q", object, field)
		}
	}
	rangeValue, ok := object["range"].([]any)
	if !ok || len(rangeValue) == 0 {
		t.Fatalf("reference entry = %#v, want SCIP range array", object)
	}
	if _, ok := object["roles"].(float64); !ok {
		t.Fatalf("reference entry = %#v, want numeric roles field", object)
	}
}
