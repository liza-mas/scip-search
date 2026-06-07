package graph_test

import (
	"encoding/json"
	"fmt"
	"slices"
	"testing"

	"github.com/scip-code/scip/bindings/go/scip"

	"scip-search/internal/query/graph"
	runtimecontract "scip-search/internal/runtime"
	"scip-search/internal/traversal"
)

const (
	targetSymbol      = "scip-go gomod example.com/project . pkg/Target()."
	dependencySymbol  = "scip-go gomod example.com/project . pkg/Dependency()."
	dependentSymbol   = "scip-go gomod example.com/project . pkg/Dependent()."
	implementationSym = "scip-go gomod example.com/project . pkg/Implementation#"
	typeDefinitionSym = "scip-go gomod example.com/project . pkg/TargetType#"
	outsideSymbol     = "scip-go gomod example.com/project . pkg/Outside()."
	testRoleSymbol    = "scip-go gomod example.com/project . pkg/TestRole()."
	testPathSymbol    = "scip-go gomod example.com/project . pkg/TestPath()."
	plainUntestedSym  = "scip-go gomod example.com/project . pkg/PlainUntested()."
	noRangeSymbol     = "scip-go gomod example.com/project . pkg/NoRange()."
	missingSymbol     = "scip-go gomod example.com/project . pkg/Missing()."
)

func TestQueryReturnsIncomingOutgoingRelationshipsAndDefinition(t *testing.T) {
	t.Parallel()

	payload := graph.Query(graphFixtureView(), targetSymbol)

	if payload.Symbol != targetSymbol {
		t.Fatalf("symbol = %q, want %q", payload.Symbol, targetSymbol)
	}
	if payload.Definition == nil ||
		payload.Definition.DocumentPath != "pkg/target.go" ||
		!slices.Equal(payload.Definition.Range, []int32{2, 5, 11}) {
		t.Fatalf("definition = %+v, want target definition location", payload.Definition)
	}
	gotIncoming := collectOccurrenceSymbols(payload.Incoming)
	wantIncoming := []string{targetSymbol, dependentSymbol, dependencySymbol, targetSymbol, targetSymbol}
	if !slices.Equal(gotIncoming, wantIncoming) {
		t.Fatalf("incoming symbols = %v, want %v", gotIncoming, wantIncoming)
	}
	gotOutgoing := collectOccurrenceSymbols(payload.Outgoing)
	wantOutgoing := []string{dependencySymbol, targetSymbol, testRoleSymbol}
	if !slices.Equal(gotOutgoing, wantOutgoing) {
		t.Fatalf("outgoing symbols = %v, want contained non-definition dependencies %v", gotOutgoing, wantOutgoing)
	}
	if slices.Contains(gotOutgoing, outsideSymbol) {
		t.Fatalf("outgoing symbols = %v, want outside occurrence excluded", gotOutgoing)
	}
	if payload.OutgoingUnavailable != "" {
		t.Fatalf("outgoing unavailable = %q, want empty", payload.OutgoingUnavailable)
	}
	if gotRelationships := collectRelationshipKinds(payload.Relationships); !slices.Equal(gotRelationships, []string{
		"incoming:reference",
		"outgoing:reference",
		"outgoing:implementation",
		"outgoing:type-definition",
	}) {
		t.Fatalf("relationships = %v, want stable relationship facts", gotRelationships)
	}
}

func TestQueryReportsUnavailableOutgoingWhenDefinitionOrEnclosingRangeIsMissing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		symbol string
		want   string
	}{
		{
			name:   "missing definition",
			symbol: missingSymbol,
			want:   graph.UnavailableNoDefinition,
		},
		{
			name:   "missing enclosing range",
			symbol: noRangeSymbol,
			want:   graph.UnavailableNoEnclosingRange,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			payload := graph.Query(graphFixtureView(), test.symbol)

			if payload.OutgoingUnavailable != test.want {
				t.Fatalf("outgoing unavailable = %q, want %q", payload.OutgoingUnavailable, test.want)
			}
			if len(payload.Outgoing) != 0 {
				t.Fatalf("outgoing = %+v, want empty when unavailable", payload.Outgoing)
			}
		})
	}
}

func TestImpactReturnsReviewDependencyAndTestHints(t *testing.T) {
	t.Parallel()

	payload := graph.Impact(graphFixtureView(), targetSymbol)

	if got := collectOccurrenceSymbols(payload.Review); !slices.Equal(got, []string{targetSymbol, dependentSymbol, dependencySymbol, targetSymbol, targetSymbol}) {
		t.Fatalf("review symbols = %v, want incoming dependents", got)
	}
	if got := collectOccurrenceSymbols(payload.Dependencies); !slices.Equal(got, []string{dependencySymbol, targetSymbol, testRoleSymbol}) {
		t.Fatalf("dependency symbols = %v, want outgoing dependencies", got)
	}
	gotHints := collectTestHints(payload.Tests)
	wantHints := []string{
		"pkg/target.go:[7 2 10]:testRole",
		"pkg/target_test.go:[5 1 7]:testPath",
	}
	if !slices.Equal(gotHints, wantHints) {
		t.Fatalf("test hints = %v, want %v", gotHints, wantHints)
	}
}

func TestImpactReturnsExplicitEmptyTestsWhenNoSignalExists(t *testing.T) {
	t.Parallel()

	payload := graph.Impact(graphFixtureView(), plainUntestedSym)

	if payload.Tests == nil {
		t.Fatal("tests = nil, want explicit empty collection")
	}
	if len(payload.Tests) != 0 {
		t.Fatalf("tests = %+v, want empty without test role or path signals", payload.Tests)
	}
}

func TestJSONPayloadsKeepExplicitEmptyCollectionsAndUnavailableReason(t *testing.T) {
	t.Parallel()

	payload := graph.Query(graphFixtureView(), missingSymbol)
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal(Query) error = %v", err)
	}
	var object map[string]any
	if err := json.Unmarshal(raw, &object); err != nil {
		t.Fatalf("json.Unmarshal(Query) error = %v", err)
	}
	if object["outgoingUnavailable"] != graph.UnavailableNoDefinition {
		t.Fatalf("outgoingUnavailable = %#v, want unavailable reason", object["outgoingUnavailable"])
	}
	for _, field := range []string{"incoming", "outgoing", "relationships"} {
		if _, ok := object[field].([]any); !ok {
			t.Fatalf("field %s = %#v, want explicit JSON array", field, object[field])
		}
	}
}

func graphFixtureView() traversal.View {
	return traversal.NewView(runtimecontract.LoadedIndex{
		Index: &scip.Index{
			Documents: []*scip.Document{
				{
					RelativePath: "pkg/target.go",
					Language:     "go",
					Symbols: []*scip.SymbolInformation{
						{
							Symbol:      targetSymbol,
							DisplayName: "Target",
							Kind:        scip.SymbolInformation_Function,
							Relationships: []*scip.Relationship{
								{Symbol: dependencySymbol, IsReference: true},
								{Symbol: implementationSym, IsImplementation: true},
								{Symbol: typeDefinitionSym, IsTypeDefinition: true},
							},
						},
						{
							Symbol:      noRangeSymbol,
							DisplayName: "NoRange",
							Kind:        scip.SymbolInformation_Function,
						},
					},
					Occurrences: []*scip.Occurrence{
						{
							Symbol:         targetSymbol,
							Range:          []int32{2, 5, 11},
							EnclosingRange: []int32{2, 0, 9, 1},
							SymbolRoles:    int32(scip.SymbolRole_Definition),
						},
						{
							Symbol:      dependencySymbol,
							Range:       []int32{4, 2, 12},
							SymbolRoles: int32(scip.SymbolRole_ReadAccess),
						},
						{
							Symbol:      targetSymbol,
							Range:       []int32{5, 2, 8},
							SymbolRoles: int32(scip.SymbolRole_ReadAccess),
						},
						{
							Symbol:      testRoleSymbol,
							Range:       []int32{7, 2, 10},
							SymbolRoles: int32(scip.SymbolRole_ReadAccess | scip.SymbolRole_Test),
						},
						{
							Symbol:      outsideSymbol,
							Range:       []int32{20, 1, 9},
							SymbolRoles: int32(scip.SymbolRole_ReadAccess),
						},
						{
							Symbol:      noRangeSymbol,
							Range:       []int32{30, 5, 12},
							SymbolRoles: int32(scip.SymbolRole_Definition),
						},
						{
							Symbol:         plainUntestedSym,
							Range:          []int32{40, 5, 16},
							EnclosingRange: []int32{40, 0, 42, 1},
							SymbolRoles:    int32(scip.SymbolRole_Definition),
						},
					},
				},
				{
					RelativePath: "pkg/dependent.go",
					Language:     "go",
					Symbols: []*scip.SymbolInformation{
						{
							Symbol:      dependentSymbol,
							DisplayName: "Dependent",
							Kind:        scip.SymbolInformation_Function,
							Relationships: []*scip.Relationship{
								{Symbol: targetSymbol, IsReference: true},
							},
						},
					},
					Occurrences: []*scip.Occurrence{
						{
							Symbol:      dependentSymbol,
							Range:       []int32{2, 5, 14},
							SymbolRoles: int32(scip.SymbolRole_Definition),
						},
						{
							Symbol:      targetSymbol,
							Range:       []int32{6, 1, 7},
							SymbolRoles: int32(scip.SymbolRole_ReadAccess),
						},
						{
							Symbol:      dependentSymbol,
							Range:       []int32{8, 1, 10},
							SymbolRoles: int32(scip.SymbolRole_ReadAccess),
						},
					},
				},
				{
					RelativePath: "pkg/target_test.go",
					Language:     "go",
					Occurrences: []*scip.Occurrence{
						{
							Symbol:      targetSymbol,
							Range:       []int32{5, 1, 7},
							SymbolRoles: int32(scip.SymbolRole_ReadAccess),
						},
					},
				},
			},
		},
	})
}

func collectOccurrenceSymbols(occurrences []graph.Occurrence) []string {
	symbols := make([]string, 0, len(occurrences))
	for _, occurrence := range occurrences {
		symbols = append(symbols, occurrence.Symbol)
	}
	return symbols
}

func collectRelationshipKinds(relationships []graph.Relationship) []string {
	kinds := make([]string, 0, len(relationships))
	for _, relationship := range relationships {
		switch {
		case relationship.IsReference:
			kinds = append(kinds, relationship.Direction+":reference")
		case relationship.IsImplementation:
			kinds = append(kinds, relationship.Direction+":implementation")
		case relationship.IsTypeDefinition:
			kinds = append(kinds, relationship.Direction+":type-definition")
		case relationship.IsDefinition:
			kinds = append(kinds, relationship.Direction+":definition")
		}
	}
	return kinds
}

func collectTestHints(hints []graph.TestHint) []string {
	collected := make([]string, 0, len(hints))
	for _, hint := range hints {
		collected = append(collected, hint.DocumentPath+":"+formatRange(hint.Range)+":"+hint.Reasons[0])
	}
	return collected
}

func formatRange(scipRange []int32) string {
	return fmt.Sprintf("%v", scipRange)
}
