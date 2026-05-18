package implementations_test

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"scip-search/internal/query/implementations"
	"scip-search/internal/traversal/traversaltest"
)

func TestImplementationsGoldenJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		symbol     string
		goldenFile string
	}{
		{
			name:       "incoming implementation with definition",
			symbol:     traversaltest.ImplSymbol,
			goldenFile: "implementations-impl.json",
		},
		{
			name:       "outgoing implementation is not returned",
			symbol:     traversaltest.AlphaSymbol,
			goldenFile: "implementations-alpha.json",
		},
		{
			name:       "missing exact symbol",
			symbol:     implementationAbsentSymbol,
			goldenFile: "implementations-missing.json",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			fixture := loadImplementationFixture(t)
			payload, err := implementations.Implementations(fixture.View, test.symbol)
			if err != nil {
				t.Fatalf("Implementations(%q) error = %v", test.symbol, err)
			}

			got := decodeJSONValue(t, mustMarshalJSON(t, payload))
			want := readGoldenJSONValue(t, test.goldenFile)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("implementations JSON value = %#v, want golden %#v", got, want)
			}
		})
	}
}

func mustMarshalJSON(t *testing.T, payload any) []byte {
	t.Helper()

	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal(%T) error = %v", payload, err)
	}
	return encoded
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
