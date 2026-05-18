package implementations_test

import (
	"reflect"
	"testing"

	"github.com/scip-code/scip/bindings/go/scip"

	"scip-search/internal/query/implementations"
	runtimecontract "scip-search/internal/runtime"
	"scip-search/internal/traversal"
)

const (
	queryInterfaceSymbol     = "scip-go gomod example.com/project . api/Doer#"
	firstImplementation      = "scip-go gomod example.com/project . impl/AlphaDoer#"
	secondImplementation     = "scip-go gomod example.com/project . impl/BetaDoer#"
	outgoingTargetSymbol     = "scip-go gomod example.com/project . api/Other#"
	referenceOnlySymbol      = "scip-go gomod example.com/project . impl/ReferenceOnly#"
	externalImplementation   = "scip-go gomod example.com/dependency v1.2.3 dep/ExternalDoer#"
	missingDefinitionTarget  = "scip-go gomod example.com/project . api/MissingDefinition#"
	absentImplementationSink = "scip-go gomod example.com/project . api/Absent#"
)

func TestImplementationsSelectsOnlyIncomingImplementationRelationships(t *testing.T) {
	t.Parallel()

	payload, err := implementations.Implementations(implementationFixtureView(), queryInterfaceSymbol)
	if err != nil {
		t.Fatalf("Implementations(%q) error = %v", queryInterfaceSymbol, err)
	}

	want := implementations.Payload{
		Symbol: queryInterfaceSymbol,
		Implementations: []implementations.Result{
			{
				ImplementationSymbol: firstImplementation,
				Relationship: implementations.Relationship{
					SourceSymbol:     firstImplementation,
					TargetSymbol:     queryInterfaceSymbol,
					IsImplementation: true,
				},
				DocumentPath: "impl/alpha.go",
				Range:        []int32{4, 5, 14},
			},
			{
				ImplementationSymbol: secondImplementation,
				Relationship: implementations.Relationship{
					SourceSymbol:     secondImplementation,
					TargetSymbol:     queryInterfaceSymbol,
					IsImplementation: true,
				},
				DocumentPath: "impl/beta.go",
				Range:        []int32{8, 2, 11},
			},
		},
	}
	if !reflect.DeepEqual(payload, want) {
		t.Fatalf("Implementations(%q) = %#v, want %#v", queryInterfaceSymbol, payload, want)
	}
}

func TestImplementationsDoNotReturnOutgoingRelationshipsFromQueriedSymbol(t *testing.T) {
	t.Parallel()

	payload, err := implementations.Implementations(implementationFixtureView(), firstImplementation)
	if err != nil {
		t.Fatalf("Implementations(%q) error = %v", firstImplementation, err)
	}

	if payload.Symbol != firstImplementation {
		t.Fatalf("payload symbol = %q, want queried symbol %q", payload.Symbol, firstImplementation)
	}
	if len(payload.Implementations) != 0 {
		t.Fatalf("implementations = %#v, want no outgoing targets returned as implementers", payload.Implementations)
	}
}

func TestImplementationsOmitLocationWhenDefinitionOccurrenceIsUnavailable(t *testing.T) {
	t.Parallel()

	payload, err := implementations.Implementations(implementationFixtureView(), missingDefinitionTarget)
	if err != nil {
		t.Fatalf("Implementations(%q) error = %v", missingDefinitionTarget, err)
	}

	want := implementations.Payload{
		Symbol: missingDefinitionTarget,
		Implementations: []implementations.Result{
			{
				ImplementationSymbol: externalImplementation,
				Relationship: implementations.Relationship{
					SourceSymbol:     externalImplementation,
					TargetSymbol:     missingDefinitionTarget,
					IsReference:      true,
					IsImplementation: true,
					IsTypeDefinition: true,
					IsDefinition:     true,
				},
			},
		},
	}
	if !reflect.DeepEqual(payload, want) {
		t.Fatalf("Implementations(%q) = %#v, want %#v", missingDefinitionTarget, payload, want)
	}
}

func TestImplementationsReturnEmptyCollectionForAbsentExactSymbol(t *testing.T) {
	t.Parallel()

	payload, err := implementations.Implementations(implementationFixtureView(), absentImplementationSink)
	if err != nil {
		t.Fatalf("Implementations(%q) error = %v", absentImplementationSink, err)
	}

	if payload.Symbol != absentImplementationSink {
		t.Fatalf("payload symbol = %q, want queried symbol %q", payload.Symbol, absentImplementationSink)
	}
	if payload.Implementations == nil || len(payload.Implementations) != 0 {
		t.Fatalf("implementations = %#v, want explicit empty collection", payload.Implementations)
	}
}

func implementationFixtureView() traversal.View {
	return traversal.NewView(runtimecontract.LoadedIndex{
		Index: &scip.Index{
			Documents: []*scip.Document{
				{
					RelativePath: "impl/alpha.go",
					Symbols: []*scip.SymbolInformation{
						{
							Symbol: firstImplementation,
							Relationships: []*scip.Relationship{
								{Symbol: queryInterfaceSymbol, IsImplementation: true},
								{Symbol: queryInterfaceSymbol, IsImplementation: true},
								{Symbol: outgoingTargetSymbol, IsImplementation: true},
							},
						},
						{
							Symbol: referenceOnlySymbol,
							Relationships: []*scip.Relationship{
								{Symbol: queryInterfaceSymbol, IsReference: true},
							},
						},
					},
					Occurrences: []*scip.Occurrence{
						{
							Symbol:      firstImplementation,
							Range:       []int32{4, 5, 14},
							SymbolRoles: int32(scip.SymbolRole_Definition),
						},
						{
							Symbol:      firstImplementation,
							Range:       []int32{12, 1, 9},
							SymbolRoles: int32(scip.SymbolRole_ReadAccess),
						},
					},
				},
				{
					RelativePath: "impl/beta.go",
					Symbols: []*scip.SymbolInformation{
						{
							Symbol: secondImplementation,
							Relationships: []*scip.Relationship{
								{Symbol: queryInterfaceSymbol, IsImplementation: true},
							},
						},
					},
					Occurrences: []*scip.Occurrence{
						{
							Symbol:      secondImplementation,
							Range:       []int32{8, 2, 11},
							SymbolRoles: int32(scip.SymbolRole_Definition),
						},
					},
				},
			},
			ExternalSymbols: []*scip.SymbolInformation{
				{
					Symbol: externalImplementation,
					Relationships: []*scip.Relationship{
						{
							Symbol:           missingDefinitionTarget,
							IsReference:      true,
							IsImplementation: true,
							IsDefinition:     true,
							IsTypeDefinition: true,
						},
					},
				},
			},
		},
	})
}
