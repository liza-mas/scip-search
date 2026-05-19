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
	escapedMatchSymbol        = "scip-go gomod github.com/liza-mas/liza . supervisor/Escaped#"
	shortRangeSymbol          = "scip-go gomod github.com/liza-mas/liza . supervisor/ShortRange#"
	runSymbol                 = "scip-go gomod github.com/liza-mas/liza . supervisor/Run()."
)

func TestSymbolsByNameReturnsEverySupervisorMatchInStableSymbolOrder(t *testing.T) {
	t.Parallel()

	result, err := discovery.FlatSymbolsByName(discoveryFixtureView(), "Supervisor")
	if err != nil {
		t.Fatalf("FlatSymbolsByName() error = %v", err)
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

	result, err := discovery.FlatSymbolsByName(discoveryFixtureView(), "Run")
	if err != nil {
		t.Fatalf("FlatSymbolsByName() error = %v", err)
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

func TestSymbolsByNamesReturnsUnionInStableSymbolOrder(t *testing.T) {
	t.Parallel()

	result, err := discovery.FlatSymbolsByNames(discoveryFixtureView(), []string{"Supervisor", "Run"})
	if err != nil {
		t.Fatalf("FlatSymbolsByNames() error = %v", err)
	}

	gotSymbols := collectResultSymbols(result.Symbols)
	wantSymbols := []string{
		agentSupervisorSymbol,
		runSymbol,
		supervisorSymbol,
		supervisorConfigSymbol,
	}
	if !slices.Equal(gotSymbols, wantSymbols) {
		t.Fatalf("symbols = %v, want sorted union matches %v", gotSymbols, wantSymbols)
	}
}

func TestSymbolsByNamesUsesFirstMatchingNameForMatchContext(t *testing.T) {
	t.Parallel()

	result, err := discovery.FlatSymbolsByNames(discoveryFixtureView(), []string{"supervisor/", "Supervisor"})
	if err != nil {
		t.Fatalf("FlatSymbolsByNames() error = %v", err)
	}

	var got discovery.SymbolResult
	found := false
	for _, symbol := range result.Symbols {
		if symbol.Symbol == supervisorSymbol {
			got = symbol
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("symbols = %v, want %q", collectResultSymbols(result.Symbols), supervisorSymbol)
	}
	if got.MatchSource != discovery.MatchSourceDescriptor || got.MatchText != "supervisor/Supervisor#" {
		t.Fatalf("match context = (%s, %q), want first matching name descriptor context", got.MatchSource, got.MatchText)
	}
}

func TestSymbolsByNameReturnsExplicitEmptyCollectionForMissingName(t *testing.T) {
	t.Parallel()

	result, err := discovery.FlatSymbolsByName(discoveryFixtureView(), "DoesNotExist")
	if err != nil {
		t.Fatalf("FlatSymbolsByName() error = %v", err)
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

	result, err := discovery.FlatSymbolsByName(discoveryFixtureView(), "[.*]")
	if err != nil {
		t.Fatalf("FlatSymbolsByName() error = %v", err)
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

func TestSymbolsByNameGroupsMatchesByPackageWithDescriptors(t *testing.T) {
	t.Parallel()

	result, err := discovery.SymbolsByName(discoveryFixtureView(), "Supervisor")
	if err != nil {
		t.Fatalf("SymbolsByName() error = %v", err)
	}

	if result.Packages == nil {
		t.Fatal("Packages = nil, want explicit collection")
	}
	if len(result.Packages) != 1 {
		t.Fatalf("Packages = %+v, want one package group", result.Packages)
	}

	pkg := result.Packages[0]
	if pkg.Scheme != "scip-go" ||
		pkg.PackageManager != "gomod" ||
		pkg.PackageName != "github.com/liza-mas/liza" ||
		pkg.PackageVersion != "." ||
		pkg.PackageKey != "scip-go gomod github.com/liza-mas/liza ." {
		t.Fatalf("package = %+v, want liza package identity", pkg)
	}

	gotDescriptors := collectCompactDescriptors(pkg.Symbols)
	wantDescriptors := []string{
		"agent/SupervisorAgent#",
		"supervisor/Supervisor#",
		"supervisor/SupervisorConfig#",
	}
	if !slices.Equal(gotDescriptors, wantDescriptors) {
		t.Fatalf("descriptors = %v, want %v", gotDescriptors, wantDescriptors)
	}
	if got := pkg.PackageKey + " " + pkg.Symbols[1].Descriptor; got != supervisorSymbol {
		t.Fatalf("reconstructed symbol = %q, want %q", got, supervisorSymbol)
	}
	if pkg.Symbols[1].Definition == nil {
		t.Fatal("definition = nil, want supervisor definition")
	}
	if pkg.Symbols[2].Definition != nil {
		t.Fatalf("definition = %+v, want omitted for symbol without definition", pkg.Symbols[2].Definition)
	}
}

func TestSymbolsByNameReturnsExplicitEmptyPackageCollectionForMissingName(t *testing.T) {
	t.Parallel()

	result, err := discovery.SymbolsByName(discoveryFixtureView(), "DoesNotExist")
	if err != nil {
		t.Fatalf("SymbolsByName() error = %v", err)
	}
	if result.Packages == nil {
		t.Fatal("Packages = nil, want explicit empty collection")
	}
	if len(result.Packages) != 0 {
		t.Fatalf("Packages = %+v, want empty collection", result.Packages)
	}

	payload, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal(SymbolPackagesPayload) error = %v", err)
	}
	if got, want := string(payload), `{"packages":[]}`; got != want {
		t.Fatalf("JSON payload = %s, want %s", got, want)
	}
}

func TestOneLineSymbolsByNameFormatsStableGrepStyleLines(t *testing.T) {
	t.Parallel()

	got, err := discovery.OneLineSymbolsByName(discoveryFixtureView(), "Supervisor")
	if err != nil {
		t.Fatalf("OneLineSymbolsByName() error = %v", err)
	}

	want := strings.Join([]string{
		"?:0:0:scip-go gomod github.com/liza-mas/liza . agent/SupervisorAgent# match=displayName text=SupervisorAgent",
		"supervisor/supervisor.go:11:6:scip-go gomod github.com/liza-mas/liza . supervisor/Supervisor# match=displayName text=Supervisor",
		"?:0:0:scip-go gomod github.com/liza-mas/liza . supervisor/SupervisorConfig# match=descriptor text=supervisor/SupervisorConfig#",
		"",
	}, "\n")
	if got != want {
		t.Fatalf("one-line output = %q, want %q", got, want)
	}
}

func TestOneLineSymbolsByNameReturnsEmptyOutputForMissingName(t *testing.T) {
	t.Parallel()

	got, err := discovery.OneLineSymbolsByName(discoveryFixtureView(), "DoesNotExist")
	if err != nil {
		t.Fatalf("OneLineSymbolsByName() error = %v", err)
	}
	if got != "" {
		t.Fatalf("one-line output = %q, want empty stdout", got)
	}
}

func TestOneLineSymbolsByNameEscapesMatchTextControls(t *testing.T) {
	t.Parallel()

	got, err := discovery.OneLineSymbolsByName(discoveryFixtureView(), "Escaped")
	if err != nil {
		t.Fatalf("OneLineSymbolsByName() error = %v", err)
	}

	want := "supervisor/supervisor.go:31:3:scip-go gomod github.com/liza-mas/liza . supervisor/Escaped# match=displayName text=Escaped\\\\Name\\nWith\\rTab\\tEnd\n"
	if got != want {
		t.Fatalf("one-line output = %q, want %q", got, want)
	}
}

func TestOneLineSymbolsByNameFallsBackForShortDefinitionRange(t *testing.T) {
	t.Parallel()

	got, err := discovery.OneLineSymbolsByName(discoveryFixtureView(), "ShortRange")
	if err != nil {
		t.Fatalf("OneLineSymbolsByName() error = %v", err)
	}

	want := "supervisor/supervisor.go:0:0:scip-go gomod github.com/liza-mas/liza . supervisor/ShortRange# match=displayName text=ShortRange\n"
	if got != want {
		t.Fatalf("one-line output = %q, want %q", got, want)
	}
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
						{Symbol: escapedMatchSymbol, DisplayName: "Escaped\\Name\nWith\rTab\tEnd"},
						{Symbol: shortRangeSymbol, DisplayName: "ShortRange"},
					},
					Occurrences: []*scip.Occurrence{
						{
							Symbol:      supervisorSymbol,
							Range:       []int32{10, 5, 15},
							SymbolRoles: int32(scip.SymbolRole_Definition),
						},
						{
							Symbol:      escapedMatchSymbol,
							Range:       []int32{30, 2, 7},
							SymbolRoles: int32(scip.SymbolRole_Definition),
						},
						{
							Symbol:      shortRangeSymbol,
							Range:       []int32{40},
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

func collectCompactDescriptors(symbols []discovery.CompactSymbolResult) []string {
	collected := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		collected = append(collected, symbol.Descriptor)
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
