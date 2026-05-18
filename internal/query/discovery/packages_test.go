package discovery_test

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/scip-code/scip/bindings/go/scip"

	"scip-search/internal/query/discovery"
	runtimecontract "scip-search/internal/runtime"
	"scip-search/internal/traversal"
)

const (
	lizaPackageKey              = "scip-go gomod github.com/liza-mas/liza ."
	scipSearchPackageKey        = "scip-go gomod github.com/liza-mas/scip-search ."
	scipSearchNPMKey            = "scip-go npm github.com/liza-mas/scip-search ."
	scipSearchOtherSchemeKey    = "scip-typescript npm github.com/liza-mas/scip-search ."
	scipSearchVersionPackageKey = "scip-go gomod github.com/liza-mas/scip-search v1.2.3"
	sourcegraphPackageKey       = "scip-go gomod github.com/sourcegraph/scip-bindings v0.6.0"
)

func TestPackagesReturnsDistinctIdentitiesInPackageKeyOrder(t *testing.T) {
	t.Parallel()

	result, err := discovery.Packages(discoveryPackagesFixtureView(), "")
	if err != nil {
		t.Fatalf("Packages() error = %v", err)
	}
	if result.Packages == nil {
		t.Fatal("Packages = nil, want explicit collection")
	}

	want := []discovery.PackageResult{
		{
			Scheme:         "other-scheme",
			PackageManager: "gomod",
			PackageName:    "example.com/scheme-only",
			PackageVersion: ".",
			PackageKey:     "other-scheme gomod example.com/scheme-only .",
		},
		{
			Scheme:         "scip-go",
			PackageManager: "github.com/liza-mas/scip-search",
			PackageName:    "example.com/manager-only",
			PackageVersion: ".",
			PackageKey:     "scip-go github.com/liza-mas/scip-search example.com/manager-only .",
		},
		{
			Scheme:         "scip-go",
			PackageManager: "gomod",
			PackageName:    "example.com/descriptor-only",
			PackageVersion: ".",
			PackageKey:     "scip-go gomod example.com/descriptor-only .",
		},
		{
			Scheme:         "scip-go",
			PackageManager: "gomod",
			PackageName:    "example.com/path-only",
			PackageVersion: ".",
			PackageKey:     "scip-go gomod example.com/path-only .",
		},
		{
			Scheme:         "scip-go",
			PackageManager: "gomod",
			PackageName:    "example.com/symbol-name-only",
			PackageVersion: ".",
			PackageKey:     "scip-go gomod example.com/symbol-name-only .",
		},
		{
			Scheme:         "scip-go",
			PackageManager: "gomod",
			PackageName:    "example.com/version-only",
			PackageVersion: "github.com/liza-mas/scip-search",
			PackageKey:     "scip-go gomod example.com/version-only github.com/liza-mas/scip-search",
		},
		{
			Scheme:         "scip-go",
			PackageManager: "gomod",
			PackageName:    "github.com/liza-mas/liza",
			PackageVersion: ".",
			PackageKey:     lizaPackageKey,
		},
		{
			Scheme:         "scip-go",
			PackageManager: "gomod",
			PackageName:    "github.com/liza-mas/scip-search",
			PackageVersion: ".",
			PackageKey:     scipSearchPackageKey,
		},
		{
			Scheme:         "scip-go",
			PackageManager: "gomod",
			PackageName:    "github.com/liza-mas/scip-search",
			PackageVersion: "v1.2.3",
			PackageKey:     scipSearchVersionPackageKey,
		},
		{
			Scheme:         "scip-go",
			PackageManager: "gomod",
			PackageName:    "github.com/sourcegraph/scip-bindings",
			PackageVersion: "v0.6.0",
			PackageKey:     sourcegraphPackageKey,
		},
		{
			Scheme:         "scip-go",
			PackageManager: "npm",
			PackageName:    "github.com/liza-mas/scip-search",
			PackageVersion: ".",
			PackageKey:     scipSearchNPMKey,
		},
		{
			Scheme:         "scip-typescript",
			PackageManager: "npm",
			PackageName:    "github.com/liza-mas/scip-search",
			PackageVersion: ".",
			PackageKey:     scipSearchOtherSchemeKey,
		},
	}
	assertPackages(t, result.Packages, want)

	payload, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal(PackagesPayload) error = %v", err)
	}
	if !slices.Contains(collectPackageJSONFields(t, payload), "packages") {
		t.Fatalf("JSON payload = %s, want top-level packages collection", payload)
	}
}

func TestPackagesFiltersByLiteralPackageNamePrefixOnly(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		prefix   string
		wantKeys []string
	}{
		{
			name:   "organization prefix",
			prefix: "github.com/liza-mas/",
			wantKeys: []string{
				lizaPackageKey,
				scipSearchPackageKey,
				scipSearchVersionPackageKey,
				scipSearchNPMKey,
				scipSearchOtherSchemeKey,
			},
		},
		{
			name:   "exact package-name prefix",
			prefix: "github.com/liza-mas/scip-search",
			wantKeys: []string{
				scipSearchPackageKey,
				scipSearchVersionPackageKey,
				scipSearchNPMKey,
				scipSearchOtherSchemeKey,
			},
		},
		{
			name:     "no match",
			prefix:   "github.com/no-match/",
			wantKeys: []string{},
		},
		{
			name:     "case mismatch",
			prefix:   "Github.com/liza-mas/",
			wantKeys: []string{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := discovery.Packages(discoveryPackagesFixtureView(), tt.prefix)
			if err != nil {
				t.Fatalf("Packages() error = %v", err)
			}

			if gotKeys := collectPackageKeys(result.Packages); !slices.Equal(gotKeys, tt.wantKeys) {
				t.Fatalf("package keys = %v, want literal packageName prefix matches %v", gotKeys, tt.wantKeys)
			}
			if len(tt.wantKeys) == 0 && result.Packages == nil {
				t.Fatal("Packages = nil, want explicit empty collection")
			}
			for _, pkg := range result.Packages {
				if pkg.PackageName == "example.com/descriptor-only" ||
					pkg.PackageName == "example.com/path-only" ||
					pkg.PackageName == "example.com/symbol-name-only" ||
					pkg.PackageName == "example.com/scheme-only" ||
					pkg.PackageName == "example.com/manager-only" ||
					pkg.PackageName == "example.com/version-only" {
					t.Fatalf("package = %+v matched prefix %q outside packageName", pkg, tt.prefix)
				}
			}
		})
	}
}

func TestPackagesReturnsExplicitEmptyCollectionForNoMatchPrefix(t *testing.T) {
	t.Parallel()

	result, err := discovery.Packages(discoveryPackagesFixtureView(), "github.com/no-match/")
	if err != nil {
		t.Fatalf("Packages() error = %v", err)
	}
	if result.Packages == nil {
		t.Fatal("Packages = nil, want explicit empty collection")
	}
	if len(result.Packages) != 0 {
		t.Fatalf("Packages = %+v, want empty collection", result.Packages)
	}

	payload, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal(PackagesPayload) error = %v", err)
	}
	if got, want := string(payload), `{"packages":[]}`; got != want {
		t.Fatalf("JSON payload = %s, want %s", got, want)
	}
}

func discoveryPackagesFixtureView() traversal.View {
	return traversal.NewView(runtimecontract.LoadedIndex{
		Index: &scip.Index{
			Documents: []*scip.Document{
				{
					RelativePath: "github.com/liza-mas/scip-search/path-only.go",
					Symbols: []*scip.SymbolInformation{
						{Symbol: "local 0", DisplayName: "LocalSymbol"},
						{Symbol: "scip-go gomod github.com/liza-mas/scip-search . internal/query/Packages().", DisplayName: "Packages"},
						{Symbol: "scip-go gomod github.com/liza-mas/scip-search . internal/query/Duplicate#"},
						{Symbol: "scip-go gomod github.com/liza-mas/liza . supervisor/Supervisor#"},
						{Symbol: "scip-go gomod example.com/path-only . internal/query/PathOnly#"},
					},
				},
				{
					RelativePath: "internal/query/packages.go",
					Symbols: []*scip.SymbolInformation{
						{Symbol: "scip-go gomod github.com/sourcegraph/scip-bindings v0.6.0 scip/Index#"},
						{Symbol: "scip-go npm github.com/liza-mas/scip-search . src/index.ts"},
						{Symbol: "scip-typescript npm github.com/liza-mas/scip-search . src/index.ts"},
						{Symbol: "scip-go gomod github.com/liza-mas/scip-search v1.2.3 internal/query/Packages()."},
						{Symbol: "other-scheme gomod example.com/scheme-only . internal/query/SchemeOnly#"},
						{Symbol: "scip-go github.com/liza-mas/scip-search example.com/manager-only . internal/query/ManagerOnly#"},
						{Symbol: "scip-go gomod example.com/version-only github.com/liza-mas/scip-search internal/query/VersionOnly#"},
						{Symbol: "scip-go gomod example.com/descriptor-only . github.com/liza-mas/scip-search/DescriptorOnly#"},
						{Symbol: "scip-go gomod example.com/symbol-name-only . internal/query/github.com/liza-mas/scip-search#"},
					},
				},
			},
		},
	})
}

func collectPackageKeys(packages []discovery.PackageResult) []string {
	collected := make([]string, 0, len(packages))
	for _, pkg := range packages {
		collected = append(collected, pkg.PackageKey)
	}
	return collected
}

func assertPackages(t *testing.T, got []discovery.PackageResult, want []discovery.PackageResult) {
	t.Helper()

	if !slices.IsSorted(collectPackageKeys(got)) {
		t.Fatalf("package keys = %v, want ascending order", collectPackageKeys(got))
	}
	if !slices.Equal(got, want) {
		t.Fatalf("packages = %+v, want %+v", got, want)
	}
}

func collectPackageJSONFields(t *testing.T, payload []byte) []string {
	t.Helper()

	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("json.Unmarshal(PackagesPayload) error = %v", err)
	}

	fields := make([]string, 0, len(decoded))
	for field := range decoded {
		fields = append(fields, field)
	}
	return fields
}
