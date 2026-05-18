package discovery_test

import (
	"bytes"
	"errors"
	"reflect"
	"slices"
	"testing"

	"scip-search/internal/cli"
	"scip-search/internal/query/discovery"
	runtimecontract "scip-search/internal/runtime"
	"scip-search/internal/traversal"
)

const (
	discoveryTraversalDependencyPackageKey = "scip-go gomod example.com/dependency v1.2.3"
	discoveryTraversalFixturePackageKey    = "scip-go gomod example.com/fixture ."
)

func TestPackagesCommandGoldenJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		prefix          string
		goldenFile      string
		wantPackageKeys []string
	}{
		{
			name:       "all packages",
			goldenFile: "packages-all.json",
			wantPackageKeys: []string{
				discoverySchemeOnlyPackageKey,
				discoveryManagerOnlyPackageKey,
				discoveryTraversalDependencyPackageKey,
				discoveryDescriptorOnlyPackageKey,
				discoveryTraversalFixturePackageKey,
				discoveryDocumentPathOnlyPackageKey,
				discoverySymbolNameOnlyPackageKey,
				discoveryVersionOnlyPackageKey,
				discoveryLizaPackageKey,
				discoveryScipSearchPackageKey,
				discoveryBindingsPackageKey,
			},
		},
		{
			name:       "organization prefix",
			prefix:     "github.com/liza-mas/",
			goldenFile: "packages-liza-mas.json",
			wantPackageKeys: []string{
				discoveryLizaPackageKey,
				discoveryScipSearchPackageKey,
			},
		},
		{
			name:       "exact package-name prefix",
			prefix:     "github.com/liza-mas/scip-search",
			goldenFile: "packages-scip-search.json",
			wantPackageKeys: []string{
				discoveryScipSearchPackageKey,
			},
		},
		{
			name:            "no match prefix",
			prefix:          "github.com/no-match/",
			goldenFile:      "packages-no-match.json",
			wantPackageKeys: []string{},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			stdout := runPackagesCommand(t, test.prefix)
			got := decodeJSONValue(t, stdout)
			want := readGoldenJSONValue(t, test.goldenFile)

			if !reflect.DeepEqual(got, want) {
				t.Fatalf("packages JSON value = %#v, want golden %#v", got, want)
			}
			assertGoldenPackagesPayload(t, got, test.wantPackageKeys)
		})
	}
}

func runPackagesCommand(t *testing.T, prefix string) []byte {
	t.Helper()

	fixture := loadDiscoveryFixture(t)
	runtime := cli.NewProductionRuntime(map[string]cli.Handler{
		"packages": packagesCommandHandler{},
	})
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	args := []string{"packages", "--index", fixture.IndexPath}
	if prefix != "" {
		args = append(args, "--prefix", prefix)
	}

	status := runtime.Run(args, &stdout, &stderr)
	if status != runtimecontract.StatusOK {
		t.Fatalf("packages command status = %d, want %d; stderr = %q", status, runtimecontract.StatusOK, stderr.String())
	}
	if stderr.String() != "" {
		t.Fatalf("packages command stderr = %q, want empty", stderr.String())
	}
	if stdout.Len() == 0 {
		t.Fatal("packages command stdout is empty, want JSON payload")
	}

	return stdout.Bytes()
}

type packagesCommandHandler struct{}

func (packagesCommandHandler) Handle(loadedIndex any, args []string) (any, error) {
	loaded, ok := loadedIndex.(runtimecontract.LoadedIndex)
	if !ok {
		return nil, errors.New("packages handler received non-SCIP loaded index")
	}
	prefix, err := parsePackagePrefixArg(args)
	if err != nil {
		return nil, err
	}

	return discovery.Packages(traversal.NewView(loaded), prefix)
}

func parsePackagePrefixArg(args []string) (string, error) {
	var prefix string
	for position := 0; position < len(args); position++ {
		if args[position] != "--prefix" {
			return "", errors.New("packages only accepts --prefix")
		}
		if position+1 >= len(args) || args[position+1] == "" {
			return "", errors.New("--prefix requires a value")
		}
		if prefix != "" {
			return "", errors.New("--prefix can only be provided once")
		}
		prefix = args[position+1]
		position++
	}

	return prefix, nil
}

func assertGoldenPackagesPayload(t *testing.T, payload any, wantPackageKeys []string) {
	t.Helper()

	object, ok := payload.(map[string]any)
	if !ok {
		t.Fatalf("payload = %T, want top-level object", payload)
	}
	packagesValue, ok := object["packages"]
	if !ok {
		t.Fatalf("payload = %#v, want top-level packages collection", payload)
	}
	packages, ok := packagesValue.([]any)
	if !ok {
		t.Fatalf("packages = %T, want array", packagesValue)
	}

	gotPackageKeys := collectGoldenPackageKeys(t, packages)
	if !slices.Equal(gotPackageKeys, wantPackageKeys) {
		t.Fatalf("package keys = %v, want %v", gotPackageKeys, wantPackageKeys)
	}
	if !slices.IsSorted(gotPackageKeys) {
		t.Fatalf("package keys = %v, want stable ascending order by exact packageKey", gotPackageKeys)
	}
	if len(wantPackageKeys) == 0 && len(packages) != 0 {
		t.Fatalf("packages = %#v, want explicit empty collection", packages)
	}

	for _, entry := range packages {
		assertRequiredPackageFields(t, entry)
	}
}

func collectGoldenPackageKeys(t *testing.T, packages []any) []string {
	t.Helper()

	collected := make([]string, 0, len(packages))
	for _, entry := range packages {
		object, ok := entry.(map[string]any)
		if !ok {
			t.Fatalf("package entry = %T, want object", entry)
		}
		packageKey, ok := object["packageKey"].(string)
		if !ok || packageKey == "" {
			t.Fatalf("package entry = %#v, want non-empty packageKey string", object)
		}
		collected = append(collected, packageKey)
	}

	return collected
}

func assertRequiredPackageFields(t *testing.T, entry any) {
	t.Helper()

	object := entry.(map[string]any)
	for _, field := range []string{
		"scheme",
		"packageManager",
		"packageName",
		"packageVersion",
		"packageKey",
	} {
		value, ok := object[field].(string)
		if !ok || value == "" {
			t.Fatalf("package entry = %#v, want required non-empty string field %q", object, field)
		}
	}
}
