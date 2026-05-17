package discovery_test

import (
	_ "embed"
	"encoding/base64"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"scip-search/internal/query/discovery"
	runtimecontract "scip-search/internal/runtime"
	"scip-search/internal/scipindex"
	"scip-search/internal/traversal"
)

const (
	discoverySupervisorSymbol       = "scip-go gomod github.com/liza-mas/liza . supervisor/Supervisor#"
	discoverySupervisorConfigSymbol = "scip-go gomod github.com/liza-mas/liza . supervisor/SupervisorConfig#"
	discoveryRunSymbol              = "scip-go gomod github.com/liza-mas/liza . supervisor/Run()."
	discoverySupervisorAgentSymbol  = "scip-go gomod github.com/liza-mas/liza . agent/SupervisorAgent#"
	discoveryLizaPackageKey         = "scip-go gomod github.com/liza-mas/liza ."
	discoveryScipSearchPackageKey   = "scip-go gomod github.com/liza-mas/scip-search ."
	discoveryBindingsPackageKey     = "scip-go gomod github.com/sourcegraph/scip-bindings ."
	discoveryFixtureIndexFile       = "discovery.scip"
	maxDiscoveryFixtureBytes        = 2048
)

//go:embed testdata/discovery.scip.b64
var discoveryFixtureBase64 string

func TestDiscoveryFixtureLoadsThroughSharedIndexAndTraversalPath(t *testing.T) {
	t.Parallel()

	fixture := loadDiscoveryFixture(t)

	if fixture.IndexPath == "" {
		t.Fatal("discovery fixture index path is empty, want temp SCIP file path")
	}
	if fixture.LoadedIndex.Path != fixture.IndexPath {
		t.Fatalf("loaded index path = %q, want fixture path %q", fixture.LoadedIndex.Path, fixture.IndexPath)
	}
	if len(fixture.View.Documents()) == 0 {
		t.Fatal("discovery fixture traversal documents are empty")
	}
	if fixture.SizeBytes > maxDiscoveryFixtureBytes {
		t.Fatalf("fixture size = %d bytes, want <= %d bytes", fixture.SizeBytes, maxDiscoveryFixtureBytes)
	}
}

func TestDiscoveryFixtureExposesSymbolsForDiscoveryCases(t *testing.T) {
	t.Parallel()

	fixture := loadDiscoveryFixture(t)
	got := collectFixtureResultSymbols(t, fixture, "Supervisor")
	want := []string{
		discoverySupervisorAgentSymbol,
		discoverySupervisorSymbol,
		discoverySupervisorConfigSymbol,
	}
	if !slices.Equal(got, want) {
		t.Fatalf("Supervisor symbols = %v, want %v", got, want)
	}

	runSymbols := collectFixtureResultSymbols(t, fixture, "Run")
	if !slices.Equal(runSymbols, []string{discoveryRunSymbol}) {
		t.Fatalf("Run symbols = %v, want only %q", runSymbols, discoveryRunSymbol)
	}

	missingSymbols := collectFixtureResultSymbols(t, fixture, "DoesNotExist")
	if missingSymbols == nil || len(missingSymbols) != 0 {
		t.Fatalf("DoesNotExist symbols = %v, want explicit empty collection", missingSymbols)
	}
}

func TestDiscoveryFixtureExposesPackageIdentitiesAndPrefixNegativeData(t *testing.T) {
	t.Parallel()

	fixture := loadDiscoveryFixture(t)
	allPackages, err := discovery.Packages(fixture.View, "")
	if err != nil {
		t.Fatalf("Packages(discovery fixture) error = %v", err)
	}
	gotKeys := collectPackageKeys(allPackages.Packages)
	for _, wantKey := range []string{
		discoveryLizaPackageKey,
		discoveryScipSearchPackageKey,
		discoveryBindingsPackageKey,
	} {
		if !slices.Contains(gotKeys, wantKey) {
			t.Fatalf("package keys = %v, want %q", gotKeys, wantKey)
		}
	}
	if countPackageKey(gotKeys, discoveryLizaPackageKey) != 1 {
		t.Fatalf("package keys = %v, want duplicated liza symbols de-duplicated to one package", gotKeys)
	}

	lizaMasPackages, err := discovery.Packages(fixture.View, "liza-mas")
	if err != nil {
		t.Fatalf("Packages(liza-mas) error = %v", err)
	}
	if len(lizaMasPackages.Packages) != 0 {
		t.Fatalf("liza-mas prefix packages = %+v, want descriptor-only data not to match packageName", lizaMasPackages.Packages)
	}

	organizationPackages, err := discovery.Packages(fixture.View, "github.com/liza-mas/")
	if err != nil {
		t.Fatalf("Packages(github.com/liza-mas/) error = %v", err)
	}
	if got, want := collectPackageKeys(organizationPackages.Packages), []string{
		discoveryLizaPackageKey,
		discoveryScipSearchPackageKey,
	}; !slices.Equal(got, want) {
		t.Fatalf("github.com/liza-mas/ package keys = %v, want %v", got, want)
	}

	scipSearchPackages, err := discovery.Packages(fixture.View, "github.com/liza-mas/scip-search")
	if err != nil {
		t.Fatalf("Packages(github.com/liza-mas/scip-search) error = %v", err)
	}
	if got, want := collectPackageKeys(scipSearchPackages.Packages), []string{discoveryScipSearchPackageKey}; !slices.Equal(got, want) {
		t.Fatalf("github.com/liza-mas/scip-search package keys = %v, want %v", got, want)
	}

	noMatchPackages, err := discovery.Packages(fixture.View, "github.com/no-match/")
	if err != nil {
		t.Fatalf("Packages(github.com/no-match/) error = %v", err)
	}
	if noMatchPackages.Packages == nil || len(noMatchPackages.Packages) != 0 {
		t.Fatalf("github.com/no-match/ packages = %+v, want explicit empty collection", noMatchPackages.Packages)
	}

	hasDescriptorCase := false
	for _, document := range fixture.View.Documents() {
		for _, symbol := range document.Symbols {
			hasDescriptorCase = hasDescriptorCase || strings.Contains(symbol.Symbol, " . internal/liza-mas/")
		}
	}
	if !hasDescriptorCase {
		t.Fatal("discovery fixture missing non-package-name liza-mas descriptor case")
	}
}

type discoveryFixture struct {
	IndexPath   string
	LoadedIndex runtimecontract.LoadedIndex
	View        traversal.View
	SizeBytes   int
}

func loadDiscoveryFixture(t testing.TB) discoveryFixture {
	t.Helper()

	indexBytes, err := base64.StdEncoding.DecodeString(strings.TrimSpace(discoveryFixtureBase64))
	if err != nil {
		t.Fatalf("decode discovery fixture: %v", err)
	}

	indexPath := filepath.Join(t.TempDir(), discoveryFixtureIndexFile)
	if err := os.WriteFile(indexPath, indexBytes, 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v", indexPath, err)
	}

	loaded, err := scipindex.NewLoader().LoadIndex(indexPath)
	if err != nil {
		t.Fatalf("load discovery fixture through official binding path: %v", err)
	}

	return discoveryFixture{
		IndexPath:   indexPath,
		LoadedIndex: loaded,
		View:        traversal.NewView(loaded),
		SizeBytes:   len(indexBytes),
	}
}

func collectFixtureResultSymbols(t *testing.T, fixture discoveryFixture, name string) []string {
	t.Helper()

	payload, err := discovery.SymbolsByName(fixture.View, name)
	if err != nil {
		t.Fatalf("SymbolsByName(%q) error = %v", name, err)
	}
	symbols := make([]string, 0, len(payload.Symbols))
	for _, symbol := range payload.Symbols {
		symbols = append(symbols, symbol.Symbol)
	}
	return symbols
}

func countPackageKey(keys []string, want string) int {
	count := 0
	for _, key := range keys {
		if key == want {
			count++
		}
	}
	return count
}
