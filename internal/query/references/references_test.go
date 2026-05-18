package references_test

import (
	"encoding/json"
	"fmt"
	"slices"
	"testing"

	"github.com/scip-code/scip/bindings/go/scip"

	"scip-search/internal/query/references"
	runtimecontract "scip-search/internal/runtime"
	"scip-search/internal/traversal"
)

const (
	querySymbol      = "scip-go gomod example.com/project . pkg/Query#"
	relatedSymbol    = "scip-go gomod example.com/project . pkg/Related#"
	incomingSymbol   = "scip-go gomod example.com/project . pkg/Incoming#"
	transitiveSymbol = "scip-go gomod example.com/project . pkg/Transitive#"
	unrelatedSymbol  = "scip-go gomod example.com/project . pkg/Unrelated#"
)

func TestQueryReturnsExactAndDirectReferenceRelatedNonDefinitionOccurrences(t *testing.T) {
	t.Parallel()

	result := references.Query(referenceFixtureView(), querySymbol)

	if result.Symbol != querySymbol {
		t.Fatalf("symbol = %q, want queried symbol %q", result.Symbol, querySymbol)
	}
	gotSymbols := collectReferenceSymbols(result.References)
	wantSymbols := []string{
		querySymbol,
		incomingSymbol,
		relatedSymbol,
	}
	if !slices.Equal(gotSymbols, wantSymbols) {
		t.Fatalf("reference symbols = %v, want exact and direct related symbols %v", gotSymbols, wantSymbols)
	}
	for _, reference := range result.References {
		if reference.Roles&int32(scip.SymbolRole_Definition) != 0 {
			t.Fatalf("reference = %+v, want no definition roles", reference)
		}
		if reference.DocumentPath == "" || len(reference.Range) == 0 {
			t.Fatalf("reference = %+v, want traversal path and range", reference)
		}
	}
}

func TestQueryReturnsStableSourceOrder(t *testing.T) {
	t.Parallel()

	result := references.Query(referenceFixtureView(), querySymbol)
	gotLocations := collectReferenceLocations(result.References)
	wantLocations := []string{
		"cmd/query.go:[8 1 8 15]:" + querySymbol,
		"pkg/incoming.go:[9 0 8]:" + incomingSymbol,
		"pkg/related.go:[10 2 14]:" + relatedSymbol,
	}
	if !slices.Equal(gotLocations, wantLocations) {
		t.Fatalf("reference order = %v, want %v", gotLocations, wantLocations)
	}
}

func TestQueryReturnsExplicitEmptyCollectionForMissingOrDefinitionOnlySymbol(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		symbol string
	}{
		{
			name:   "absent",
			symbol: "scip-go gomod example.com/project . missing/Absent#",
		},
		{
			name:   "definition only",
			symbol: "scip-go gomod example.com/project . pkg/DefinitionOnly#",
		},
		{
			name:   "substring does not match",
			symbol: "Query",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := references.Query(referenceFixtureView(), test.symbol)

			if result.Symbol != test.symbol {
				t.Fatalf("symbol = %q, want queried symbol %q", result.Symbol, test.symbol)
			}
			if result.References == nil {
				t.Fatal("References = nil, want explicit empty collection")
			}
			if len(result.References) != 0 {
				t.Fatalf("References = %+v, want empty collection", result.References)
			}

			payload, err := json.Marshal(result)
			if err != nil {
				t.Fatalf("json.Marshal(Query result) error = %v", err)
			}
			want := `{"symbol":"` + test.symbol + `","references":[]}`
			if string(payload) != want {
				t.Fatalf("JSON payload = %s, want %s", payload, want)
			}
		})
	}
}

func TestOneLineFormatsReferenceLocationsAndRoles(t *testing.T) {
	t.Parallel()

	payload := references.Payload{
		References: []references.Reference{
			{
				Symbol:       querySymbol,
				DocumentPath: "cmd/query.go",
				Range:        []int32{8, 1, 8, 15},
				Roles:        int32(scip.SymbolRole_ReadAccess),
			},
			{
				Symbol:       incomingSymbol,
				DocumentPath: "pkg/incoming.go",
				Range:        []int32{9},
				Roles:        int32(scip.SymbolRole_WriteAccess),
			},
			{
				Symbol: relatedSymbol,
				Range:  []int32{10, 2, 14},
				Roles:  int32(scip.SymbolRole_ReadAccess),
			},
		},
	}

	want := "cmd/query.go:9:2:scip-go gomod example.com/project . pkg/Query# roles=8\n" +
		"pkg/incoming.go:0:0:scip-go gomod example.com/project . pkg/Incoming# roles=4\n" +
		"?:0:0:scip-go gomod example.com/project . pkg/Related# roles=8\n"
	if got := references.OneLine(payload); got != want {
		t.Fatalf("OneLine() = %q, want %q", got, want)
	}
}

func TestOneLineReturnsEmptyOutputForEmptyReferencePayload(t *testing.T) {
	t.Parallel()

	if got := references.OneLine(references.Payload{References: []references.Reference{}}); got != "" {
		t.Fatalf("OneLine(empty) = %q, want empty", got)
	}
}

func referenceFixtureView() traversal.View {
	return traversal.NewView(runtimecontract.LoadedIndex{
		Index: &scip.Index{
			Documents: []*scip.Document{
				{
					RelativePath: "cmd/query.go",
					Symbols: []*scip.SymbolInformation{
						{
							Symbol: querySymbol,
							Relationships: []*scip.Relationship{
								{Symbol: relatedSymbol, IsReference: true},
								{Symbol: relatedSymbol, IsReference: true},
								{Symbol: incomingSymbol, IsReference: true},
							},
						},
					},
					Occurrences: []*scip.Occurrence{
						{
							Symbol:      querySymbol,
							Range:       []int32{2, 6, 11},
							SymbolRoles: int32(scip.SymbolRole_Definition | scip.SymbolRole_WriteAccess),
						},
						{
							Symbol:      querySymbol,
							Range:       []int32{8, 1, 8, 15},
							SymbolRoles: int32(scip.SymbolRole_ReadAccess),
						},
						{
							Symbol:      unrelatedSymbol,
							Range:       []int32{3, 0, 9},
							SymbolRoles: int32(scip.SymbolRole_ReadAccess),
						},
					},
				},
				{
					RelativePath: "pkg/incoming.go",
					Symbols: []*scip.SymbolInformation{
						{
							Symbol: incomingSymbol,
							Relationships: []*scip.Relationship{
								{Symbol: querySymbol, IsReference: true},
							},
						},
					},
					Occurrences: []*scip.Occurrence{
						{
							Symbol:      incomingSymbol,
							Range:       []int32{9, 0, 8},
							SymbolRoles: int32(scip.SymbolRole_ReadAccess),
						},
					},
				},
				{
					RelativePath: "pkg/related.go",
					Symbols: []*scip.SymbolInformation{
						{
							Symbol: relatedSymbol,
							Relationships: []*scip.Relationship{
								{Symbol: transitiveSymbol, IsReference: true},
							},
						},
					},
					Occurrences: []*scip.Occurrence{
						{
							Symbol:      relatedSymbol,
							Range:       []int32{10, 2, 14},
							SymbolRoles: int32(scip.SymbolRole_ReadAccess),
						},
						{
							Symbol:      transitiveSymbol,
							Range:       []int32{10, 2, 14},
							SymbolRoles: int32(scip.SymbolRole_ReadAccess),
						},
						{
							Symbol:      relatedSymbol,
							Range:       []int32{11, 2, 15},
							SymbolRoles: int32(scip.SymbolRole_Definition),
						},
						{
							Symbol:      "scip-go gomod example.com/project . pkg/DefinitionOnly#",
							Range:       []int32{20, 2, 15},
							SymbolRoles: int32(scip.SymbolRole_Definition),
						},
					},
				},
			},
		},
	})
}

func collectReferenceSymbols(results []references.Reference) []string {
	symbols := make([]string, 0, len(results))
	for _, result := range results {
		symbols = append(symbols, result.Symbol)
	}
	return symbols
}

func collectReferenceLocations(results []references.Reference) []string {
	locations := make([]string, 0, len(results))
	for _, result := range results {
		locations = append(locations, result.DocumentPath+":"+int32SliceString(result.Range)+":"+result.Symbol)
	}
	return locations
}

func int32SliceString(values []int32) string {
	return fmt.Sprint(values)
}
