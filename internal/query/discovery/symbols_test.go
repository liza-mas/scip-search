package discovery_test

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/scip-code/scip/bindings/go/scip"

	"scip-search/internal/query/discovery"
	runtimecontract "scip-search/internal/runtime"
	"scip-search/internal/traversal"
)

const (
	agentSupervisorSymbol     = "scip-go gomod github.com/liza-mas/liza . agent/SupervisorAgent#"
	lowercaseSupervisorSymbol = "scip-go gomod github.com/liza-mas/liza . supervisor/lowercase#"
	patternCharacterSymbol    = "scip-go gomod github.com/liza-mas/liza . patterns/Name[.*]()."
	supervisorConfigSymbol    = "scip-go gomod github.com/liza-mas/liza . supervisor/SupervisorConfig#"
	supervisorSymbol          = "scip-go gomod github.com/liza-mas/liza . supervisor/Supervisor#"
	runSymbol                 = "scip-go gomod github.com/liza-mas/liza . supervisor/Run()."
)

func TestSymbolsByNameReturnsEverySupervisorMatchInStableSymbolOrder(t *testing.T) {
	t.Parallel()

	result, err := discovery.SymbolsByName(discoveryFixtureView(), "Supervisor")
	if err != nil {
		t.Fatalf("SymbolsByName() error = %v", err)
	}

	gotSymbols := collectResultSymbols(result.Symbols)
	wantSymbols := []string{
		agentSupervisorSymbol,
		supervisorSymbol,
		supervisorConfigSymbol,
	}
	if !slices.Equal(gotSymbols, wantSymbols) {
		t.Fatalf("symbols = %v, want sorted exact matches %v", gotSymbols, wantSymbols)
	}

	assertSymbolResult(t, result.Symbols[0], discovery.SymbolResult{
		Symbol:         agentSupervisorSymbol,
		Scheme:         "scip-go",
		PackageManager: "gomod",
		PackageName:    "github.com/liza-mas/liza",
		PackageVersion: ".",
		MatchText:      "SupervisorAgent",
		MatchSource:    discovery.MatchSourceDisplayName,
	})
	assertSymbolResult(t, result.Symbols[1], discovery.SymbolResult{
		Symbol:         supervisorSymbol,
		Scheme:         "scip-go",
		PackageManager: "gomod",
		PackageName:    "github.com/liza-mas/liza",
		PackageVersion: ".",
		MatchText:      "Supervisor",
		MatchSource:    discovery.MatchSourceDisplayName,
		Definition: &discovery.Definition{
			DocumentPath: "supervisor/supervisor.go",
			Range:        []int32{10, 5, 15},
		},
	})
	assertSymbolResult(t, result.Symbols[2], discovery.SymbolResult{
		Symbol:         supervisorConfigSymbol,
		Scheme:         "scip-go",
		PackageManager: "gomod",
		PackageName:    "github.com/liza-mas/liza",
		PackageVersion: ".",
		MatchText:      "supervisor/SupervisorConfig#",
		MatchSource:    discovery.MatchSourceDescriptor,
	})

	withoutDefinition, err := json.Marshal(result.Symbols[2])
	if err != nil {
		t.Fatalf("json.Marshal(symbol without definition) error = %v", err)
	}
	if strings.Contains(string(withoutDefinition), "definition") {
		t.Fatalf("symbol without definition JSON = %s, want definition omitted", withoutDefinition)
	}
}

func TestSymbolsByNameReturnsOnlyRunMatches(t *testing.T) {
	t.Parallel()

	result, err := discovery.SymbolsByName(discoveryFixtureView(), "Run")
	if err != nil {
		t.Fatalf("SymbolsByName() error = %v", err)
	}

	if got, want := collectResultSymbols(result.Symbols), []string{runSymbol}; !slices.Equal(got, want) {
		t.Fatalf("symbols = %v, want only literal Run match %v", got, want)
	}
	assertSymbolResult(t, result.Symbols[0], discovery.SymbolResult{
		Symbol:         runSymbol,
		Scheme:         "scip-go",
		PackageManager: "gomod",
		PackageName:    "github.com/liza-mas/liza",
		PackageVersion: ".",
		MatchText:      "Run",
		MatchSource:    discovery.MatchSourceDisplayName,
		Definition: &discovery.Definition{
			DocumentPath: "supervisor/run.go",
			Range:        []int32{20, 1, 8},
		},
	})
}

func TestSymbolsByNameReturnsExplicitEmptyCollectionForMissingName(t *testing.T) {
	t.Parallel()

	result, err := discovery.SymbolsByName(discoveryFixtureView(), "DoesNotExist")
	if err != nil {
		t.Fatalf("SymbolsByName() error = %v", err)
	}
	if result.Symbols == nil {
		t.Fatal("Symbols = nil, want explicit empty collection")
	}
	if len(result.Symbols) != 0 {
		t.Fatalf("Symbols = %+v, want empty collection", result.Symbols)
	}

	payload, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal(SymbolsResult) error = %v", err)
	}
	if got, want := string(payload), `{"symbols":[]}`; got != want {
		t.Fatalf("JSON payload = %s, want %s", got, want)
	}
}

func TestSymbolsByNameTreatsPatternCharactersLiterally(t *testing.T) {
	t.Parallel()

	result, err := discovery.SymbolsByName(discoveryFixtureView(), "[.*]")
	if err != nil {
		t.Fatalf("SymbolsByName() error = %v", err)
	}

	if got, want := collectResultSymbols(result.Symbols), []string{patternCharacterSymbol}; !slices.Equal(got, want) {
		t.Fatalf("symbols = %v, want only literal pattern-character match %v", got, want)
	}
	assertSymbolResult(t, result.Symbols[0], discovery.SymbolResult{
		Symbol:         patternCharacterSymbol,
		Scheme:         "scip-go",
		PackageManager: "gomod",
		PackageName:    "github.com/liza-mas/liza",
		PackageVersion: ".",
		MatchText:      "patterns/Name[.*]().",
		MatchSource:    discovery.MatchSourceDescriptor,
	})
}

func discoveryFixtureView() traversal.View {
	return traversal.NewView(runtimecontract.LoadedIndex{
		Index: &scip.Index{
			Documents: []*scip.Document{
				{
					RelativePath: "supervisor/supervisor.go",
					Symbols: []*scip.SymbolInformation{
						{Symbol: "local 0", DisplayName: "SupervisorLocal"},
						{Symbol: supervisorConfigSymbol, DisplayName: "Config"},
						{Symbol: lowercaseSupervisorSymbol, DisplayName: "supervisor"},
						{Symbol: supervisorSymbol, DisplayName: "Supervisor"},
					},
					Occurrences: []*scip.Occurrence{
						{
							Symbol:      supervisorSymbol,
							Range:       []int32{10, 5, 15},
							SymbolRoles: int32(scip.SymbolRole_Definition),
						},
					},
				},
				{
					RelativePath: "supervisor/run.go",
					Symbols: []*scip.SymbolInformation{
						{Symbol: runSymbol, DisplayName: "Run"},
						{Symbol: agentSupervisorSymbol, DisplayName: "SupervisorAgent"},
						{Symbol: patternCharacterSymbol, DisplayName: "Name"},
					},
					Occurrences: []*scip.Occurrence{
						{
							Symbol:      runSymbol,
							Range:       []int32{20, 1, 8},
							SymbolRoles: int32(scip.SymbolRole_Definition),
						},
					},
				},
			},
		},
	})
}

func collectResultSymbols(symbols []discovery.SymbolResult) []string {
	collected := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		collected = append(collected, symbol.Symbol)
	}
	return collected
}

func assertSymbolResult(t *testing.T, got discovery.SymbolResult, want discovery.SymbolResult) {
	t.Helper()

	if got.Symbol != want.Symbol ||
		got.Scheme != want.Scheme ||
		got.PackageManager != want.PackageManager ||
		got.PackageName != want.PackageName ||
		got.PackageVersion != want.PackageVersion ||
		got.MatchText != want.MatchText ||
		got.MatchSource != want.MatchSource {
		t.Fatalf("symbol result = %+v, want %+v", got, want)
	}
	if want.Definition == nil {
		if got.Definition != nil {
			t.Fatalf("definition = %+v, want omitted", got.Definition)
		}
		return
	}
	if got.Definition == nil {
		t.Fatalf("definition = nil, want %+v", want.Definition)
	}
	if got.Definition.DocumentPath != want.Definition.DocumentPath ||
		!slices.Equal(got.Definition.Range, want.Definition.Range) {
		t.Fatalf("definition = %+v, want %+v", got.Definition, want.Definition)
	}
}
