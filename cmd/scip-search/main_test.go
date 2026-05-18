package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/scip-code/scip/bindings/go/scip"
	"google.golang.org/protobuf/proto"

	"scip-search/internal/cli"
	runtimecontract "scip-search/internal/runtime"
	"scip-search/internal/traversal/traversaltest"
	"scip-search/internal/version"
)

func TestRunSharedInvocationFailuresUseCLIStreamsAndStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "no command",
			args: nil,
		},
		{
			name: "unsupported command",
			args: []string{"search", "--index", "repo.scip", "--name", "Supervisor"},
		},
		{
			name: "symbols missing index",
			args: []string{"symbols", "--name", "Supervisor"},
		},
		{
			name: "references missing index",
			args: []string{"references", "--symbol", "scip-go gomod example . pkg/Foo#"},
		},
		{
			name: "implementations missing index",
			args: []string{"implementations", "--symbol", "scip-go gomod example . pkg/Foo#"},
		},
		{
			name: "packages missing index",
			args: []string{"packages", "--prefix", "github.com/example"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var stdout bytes.Buffer
			var stderr bytes.Buffer

			status := run(test.args, &stdout, &stderr)

			if status != runtimecontract.StatusUsage {
				t.Fatalf("run(%v) status = %d, want %d", test.args, status, runtimecontract.StatusUsage)
			}
			if stdout.String() != "" {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
			if stderr.String() == "" {
				t.Fatal("stderr is empty, want usage diagnostic")
			}
		})
	}
}

func TestRunHelpUsesSharedRuntimeBeforeQueryValidation(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	status := run([]string{"--help"}, &stdout, &stderr)

	if status != runtimecontract.StatusOK {
		t.Fatalf("run(--help) status = %d, want %d", status, runtimecontract.StatusOK)
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	for _, want := range []string{
		"Usage:",
		"scip-search symbols --index <index-path> --name <name>",
		"scip-search references --index <index-path> --symbol <scip-symbol>",
		"scip-search implementations --index <index-path> --symbol <scip-symbol>",
		"scip-search packages --index <index-path> [--prefix <prefix>]",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want substring %q", stdout.String(), want)
		}
	}
	if strings.HasPrefix(strings.TrimSpace(stdout.String()), "{") {
		t.Fatalf("stdout = %q, want non-JSON help output", stdout.String())
	}
}

func TestRunVersionUsesBuildIdentityBeforeQueryValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		identity version.BuildIdentity
		want     []string
	}{
		{
			name: "release",
			identity: version.BuildIdentity{
				Release: "v2.0.0",
			},
			want: []string{"scip-search", "release", "v2.0.0"},
		},
		{
			name: "source",
			identity: version.BuildIdentity{
				SourceRef:      "local",
				SourceRevision: "abc999",
			},
			want: []string{"scip-search", "source", "local", "abc999"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var stdout bytes.Buffer
			var stderr bytes.Buffer

			status := runWithBuildIdentity([]string{"--version"}, &stdout, &stderr, test.identity)

			if status != runtimecontract.StatusOK {
				t.Fatalf("runWithBuildIdentity(--version) status = %d, want %d", status, runtimecontract.StatusOK)
			}
			if stdout.String() == "" {
				t.Fatal("stdout is empty, want version output")
			}
			for _, want := range test.want {
				if !strings.Contains(stdout.String(), want) {
					t.Fatalf("stdout = %q, want substring %q", stdout.String(), want)
				}
			}
			if stderr.String() != "" {
				t.Fatalf("stderr = %q, want empty", stderr.String())
			}
			if strings.HasPrefix(strings.TrimSpace(stdout.String()), "{") {
				t.Fatalf("stdout = %q, want non-JSON version output", stdout.String())
			}
		})
	}
}

func TestRunProductionIndexLoadingFailuresAcrossDocumentedCommands(t *testing.T) {
	t.Parallel()

	failures := []struct {
		name       string
		indexPath  func(t *testing.T) string
		assertPath func(t *testing.T, path string)
	}{
		{
			name: "nonexistent",
			indexPath: func(t *testing.T) string {
				t.Helper()

				return filepath.Join(t.TempDir(), "missing.scip")
			},
			assertPath: func(t *testing.T, path string) {
				t.Helper()

				if _, err := os.Stat(path); !os.IsNotExist(err) {
					t.Fatalf("os.Stat(%q) error = %v, want nonexistent selected index", path, err)
				}
			},
		},
		{
			name: "directory",
			indexPath: func(t *testing.T) string {
				t.Helper()

				return t.TempDir()
			},
			assertPath: func(t *testing.T, path string) {
				t.Helper()

				info, err := os.Stat(path)
				if err != nil {
					t.Fatalf("os.Stat(%q) error = %v", path, err)
				}
				if !info.IsDir() {
					t.Fatalf("selected index path %q is not a directory after run", path)
				}
			},
		},
		{
			name: "invalid",
			indexPath: func(t *testing.T) string {
				t.Helper()

				return writeSelectedIndexBytes(t, []byte("not a SCIP protobuf"))
			},
			assertPath: func(t *testing.T, path string) {
				t.Helper()

				assertSelectedIndexBytes(t, path, []byte("not a SCIP protobuf"))
			},
		},
	}

	for _, command := range documentedQueryCommands() {
		for _, failure := range failures {
			t.Run(command+"/"+failure.name, func(t *testing.T) {
				t.Parallel()

				indexPath := failure.indexPath(t)
				var stdout bytes.Buffer
				var stderr bytes.Buffer

				status := run(documentedCommandArgs(command, indexPath), &stdout, &stderr)

				assertIndexLoadFailure(t, status, &stdout, &stderr)
				failure.assertPath(t, indexPath)
			})
		}
	}
}

func TestRunDeterministicUnreadableOpenFailureAcrossDocumentedCommands(t *testing.T) {
	t.Parallel()

	for _, command := range documentedQueryCommands() {
		t.Run(command, func(t *testing.T) {
			t.Parallel()

			indexPath := writeSelectedIndexBytes(t, []byte("kept unreadable by injected loader failure"))
			loader := &openFailureLoader{}
			handlers := newBoundaryHandlers()
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			status := runWithTestRuntime(documentedCommandArgs(command, indexPath), loader, handlers, &stdout, &stderr)

			assertIndexLoadFailure(t, status, &stdout, &stderr)
			if got, want := loader.paths, []string{indexPath}; !slices.Equal(got, want) {
				t.Fatalf("loader paths = %v, want only selected index %q", got, indexPath)
			}
			assertNoBoundaryHandlerCalls(t, handlers)
			assertSelectedIndexBytes(t, indexPath, []byte("kept unreadable by injected loader failure"))
		})
	}
}

func TestRunValidGeneratedSCIPLoadsReachQueryBoundaryAcrossDocumentedCommands(t *testing.T) {
	t.Parallel()

	for _, command := range documentedQueryCommands() {
		t.Run(command, func(t *testing.T) {
			t.Parallel()

			indexPath := writeValidSelectedSCIPIndex(t, command)
			handlers := newBoundaryHandlers()
			selectedHandler := handlers[command].(*boundaryHandler)
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			status := runWithProductionLoader(documentedCommandArgs(command, indexPath), handlers, &stdout, &stderr)

			if status != runtimecontract.StatusOK {
				t.Fatalf("runWithProductionLoader(%s) status = %d, want %d", command, status, runtimecontract.StatusOK)
			}
			if stderr.String() != "" {
				t.Fatalf("stderr = %q, want empty", stderr.String())
			}
			if selectedHandler.calls != 1 {
				t.Fatalf("%s handler calls = %d, want 1", command, selectedHandler.calls)
			}
			loaded, ok := selectedHandler.loaded.(runtimecontract.LoadedIndex)
			if !ok {
				t.Fatalf("%s handler loaded context = %T, want runtime LoadedIndex", command, selectedHandler.loaded)
			}
			if loaded.Path != indexPath {
				t.Fatalf("%s loaded path = %q, want selected index %q", command, loaded.Path, indexPath)
			}
			if got := loaded.Index.GetMetadata().GetToolInfo().GetName(); got != command {
				t.Fatalf("%s loaded SCIP tool name = %q, want %q", command, got, command)
			}
			if got, want := selectedHandler.args, documentedQueryArgs(command); !slices.Equal(got, want) {
				t.Fatalf("%s handler args = %v, want %v", command, got, want)
			}
			assertOtherBoundaryHandlersNotCalled(t, handlers, command)
			assertSingleJSONCommand(t, stdout.Bytes(), command)
		})
	}
}

func TestRunProductionSymbolsCommandUsesDiscoveryImplementation(t *testing.T) {
	t.Parallel()

	fixture := traversaltest.LoadSharedFixture(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	status := run([]string{"symbols", "--index", fixture.IndexPath, "--name", "Supervisor"}, &stdout, &stderr)

	if status != runtimecontract.StatusOK {
		t.Fatalf("symbols status = %d, want %d; stderr = %q", status, runtimecontract.StatusOK, stderr.String())
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	assertProductionJSONMatchesGolden(t, stdout.Bytes(), "symbols-supervisor.json")
}

func TestRunProductionPackagesCommandUsesDiscoveryImplementation(t *testing.T) {
	t.Parallel()

	fixture := traversaltest.LoadSharedFixture(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	status := run([]string{"packages", "--index", fixture.IndexPath}, &stdout, &stderr)

	if status != runtimecontract.StatusOK {
		t.Fatalf("packages status = %d, want %d; stderr = %q", status, runtimecontract.StatusOK, stderr.String())
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	assertProductionJSONMatchesGolden(t, stdout.Bytes(), "packages-all.json")
}

func TestRunProductionPackagesCommandAcceptsPrefix(t *testing.T) {
	t.Parallel()

	fixture := traversaltest.LoadSharedFixture(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	status := run([]string{"packages", "--index", fixture.IndexPath, "--prefix", "github.com/liza-mas/"}, &stdout, &stderr)

	if status != runtimecontract.StatusOK {
		t.Fatalf("packages prefix status = %d, want %d; stderr = %q", status, runtimecontract.StatusOK, stderr.String())
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	assertProductionJSONMatchesGolden(t, stdout.Bytes(), "packages-liza-mas.json")
}

func TestRunProductionImplementationsCommandUsesImplementationQuery(t *testing.T) {
	t.Parallel()

	fixture := traversaltest.LoadSharedFixture(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	status := run([]string{"implementations", "--index", fixture.IndexPath, "--symbol", traversaltest.ImplSymbol}, &stdout, &stderr)

	if status != runtimecontract.StatusOK {
		t.Fatalf("implementations status = %d, want %d; stderr = %q", status, runtimecontract.StatusOK, stderr.String())
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	assertProductionImplementationsJSONMatchesGolden(t, stdout.Bytes(), "implementations-impl.json")
}

func TestRunProductionReferencesCommandUsesReferencesImplementation(t *testing.T) {
	t.Parallel()

	fixture := traversaltest.LoadSharedFixture(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	status := run([]string{"references", "--index", fixture.IndexPath, "--symbol", traversaltest.AlphaSymbol}, &stdout, &stderr)

	if status != runtimecontract.StatusOK {
		t.Fatalf("references status = %d, want %d; stderr = %q", status, runtimecontract.StatusOK, stderr.String())
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	assertProductionJSONMatchesReferenceGolden(t, stdout.Bytes(), "references-alpha.json")
}

func TestRunProductionReferencesCommandIncludesIncomingReferenceRelationships(t *testing.T) {
	t.Parallel()

	fixture := traversaltest.LoadSharedFixture(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	status := run([]string{"references", "--index", fixture.IndexPath, "--symbol", traversaltest.BetaSymbol}, &stdout, &stderr)

	if status != runtimecontract.StatusOK {
		t.Fatalf("references incoming status = %d, want %d; stderr = %q", status, runtimecontract.StatusOK, stderr.String())
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	assertProductionJSONMatchesReferenceGolden(t, stdout.Bytes(), "references-beta.json")
}

func TestRunProductionImplementationsEmptyResultPreservesQueriedSymbol(t *testing.T) {
	t.Parallel()

	fixture := traversaltest.LoadSharedFixture(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	status := run([]string{"implementations", "--index", fixture.IndexPath, "--symbol", traversaltest.AlphaSymbol}, &stdout, &stderr)

	if status != runtimecontract.StatusOK {
		t.Fatalf("implementations empty status = %d, want %d; stderr = %q", status, runtimecontract.StatusOK, stderr.String())
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	assertProductionImplementationsJSONMatchesGolden(t, stdout.Bytes(), "implementations-alpha.json")
}

func TestRunProductionReferencesCommandReturnsEmptyResultsForAbsentExactSymbol(t *testing.T) {
	t.Parallel()

	fixture := traversaltest.LoadSharedFixture(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	status := run([]string{
		"references",
		"--index",
		fixture.IndexPath,
		"--symbol",
		"scip-go gomod example.com/fixture . missing/Absent#",
	}, &stdout, &stderr)

	if status != runtimecontract.StatusOK {
		t.Fatalf("references absent status = %d, want %d; stderr = %q", status, runtimecontract.StatusOK, stderr.String())
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	assertProductionJSONMatchesReferenceGolden(t, stdout.Bytes(), "references-absent.json")
}

func TestRunProductionQuerySpecificArgumentUsageFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		command string
		args    []string
	}{
		{
			name:    "symbols missing name",
			command: "symbols",
			args:    nil,
		},
		{
			name:    "symbols missing name value",
			command: "symbols",
			args:    []string{"--name"},
		},
		{
			name:    "symbols empty name",
			command: "symbols",
			args:    []string{"--name", ""},
		},
		{
			name:    "symbols flag-shaped name",
			command: "symbols",
			args:    []string{"--name", "--unknown"},
		},
		{
			name:    "symbols duplicate name",
			command: "symbols",
			args:    []string{"--name", "Supervisor", "--name", "Run"},
		},
		{
			name:    "symbols unknown trailing flag",
			command: "symbols",
			args:    []string{"--name", "Supervisor", "--unknown"},
		},
		{
			name:    "symbols stray positional before name",
			command: "symbols",
			args:    []string{"stray", "--name", "Supervisor"},
		},
		{
			name:    "symbols stray positional after name",
			command: "symbols",
			args:    []string{"--name", "Supervisor", "stray"},
		},
		{
			name:    "symbols wrong query flag",
			command: "symbols",
			args:    []string{"--symbol", traversaltest.AlphaSymbol},
		},
		{
			name:    "references missing symbol",
			command: "references",
			args:    nil,
		},
		{
			name:    "references missing symbol value",
			command: "references",
			args:    []string{"--symbol"},
		},
		{
			name:    "references empty symbol",
			command: "references",
			args:    []string{"--symbol", ""},
		},
		{
			name:    "references flag-shaped symbol",
			command: "references",
			args:    []string{"--symbol", "--unknown"},
		},
		{
			name:    "references duplicate symbol",
			command: "references",
			args:    []string{"--symbol", traversaltest.AlphaSymbol, "--symbol", traversaltest.BetaSymbol},
		},
		{
			name:    "references unknown trailing flag",
			command: "references",
			args:    []string{"--symbol", traversaltest.AlphaSymbol, "--unknown"},
		},
		{
			name:    "references stray positional before symbol",
			command: "references",
			args:    []string{"stray", "--symbol", traversaltest.AlphaSymbol},
		},
		{
			name:    "references stray positional after symbol",
			command: "references",
			args:    []string{"--symbol", traversaltest.AlphaSymbol, "stray"},
		},
		{
			name:    "references wrong query flag",
			command: "references",
			args:    []string{"--name", "Supervisor"},
		},
		{
			name:    "implementations missing symbol",
			command: "implementations",
			args:    nil,
		},
		{
			name:    "implementations missing symbol value",
			command: "implementations",
			args:    []string{"--symbol"},
		},
		{
			name:    "implementations empty symbol",
			command: "implementations",
			args:    []string{"--symbol", ""},
		},
		{
			name:    "implementations flag-shaped symbol",
			command: "implementations",
			args:    []string{"--symbol", "--unknown"},
		},
		{
			name:    "implementations duplicate symbol",
			command: "implementations",
			args:    []string{"--symbol", traversaltest.AlphaSymbol, "--symbol", traversaltest.ImplSymbol},
		},
		{
			name:    "implementations unknown trailing flag",
			command: "implementations",
			args:    []string{"--symbol", traversaltest.ImplSymbol, "--unknown"},
		},
		{
			name:    "implementations stray positional before symbol",
			command: "implementations",
			args:    []string{"stray", "--symbol", traversaltest.ImplSymbol},
		},
		{
			name:    "implementations stray positional after symbol",
			command: "implementations",
			args:    []string{"--symbol", traversaltest.ImplSymbol, "stray"},
		},
		{
			name:    "implementations wrong query flag",
			command: "implementations",
			args:    []string{"--name", "Supervisor"},
		},
		{
			name:    "packages stray positional",
			command: "packages",
			args:    []string{"stray"},
		},
		{
			name:    "packages missing prefix value",
			command: "packages",
			args:    []string{"--prefix"},
		},
		{
			name:    "packages empty prefix",
			command: "packages",
			args:    []string{"--prefix", ""},
		},
		{
			name:    "packages flag-shaped prefix",
			command: "packages",
			args:    []string{"--prefix", "--name"},
		},
		{
			name:    "packages duplicate prefix",
			command: "packages",
			args:    []string{"--prefix", "github.com/liza-mas/", "--prefix", "github.com/sourcegraph/"},
		},
		{
			name:    "packages unknown trailing flag",
			command: "packages",
			args:    []string{"--prefix", "github.com/liza-mas/", "--unknown"},
		},
		{
			name:    "packages stray positional before prefix",
			command: "packages",
			args:    []string{"stray", "--prefix", "github.com/liza-mas/"},
		},
		{
			name:    "packages stray positional after prefix",
			command: "packages",
			args:    []string{"--prefix", "github.com/liza-mas/", "stray"},
		},
		{
			name:    "packages wrong query flag",
			command: "packages",
			args:    []string{"--name", "Supervisor"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			fixture := traversaltest.LoadSharedFixture(t)
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			args := []string{test.command, "--index", fixture.IndexPath}
			args = append(args, test.args...)

			status := run(args, &stdout, &stderr)

			if status != runtimecontract.StatusUsage {
				t.Fatalf("%s invalid args status = %d, want %d; stdout = %q; stderr = %q", test.command, status, runtimecontract.StatusUsage, stdout.String(), stderr.String())
			}
			if stdout.String() != "" {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
			if stderr.String() == "" {
				t.Fatal("stderr is empty, want usage diagnostic")
			}
		})
	}
}

func TestRunProductionSymbolsMissingNameRemainsUsageFailure(t *testing.T) {
	t.Parallel()

	fixture := traversaltest.LoadSharedFixture(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	status := run([]string{"symbols", "--index", fixture.IndexPath}, &stdout, &stderr)

	if status != runtimecontract.StatusUsage {
		t.Fatalf("symbols missing --name status = %d, want %d", status, runtimecontract.StatusUsage)
	}
	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "missing --name") {
		t.Fatalf("stderr = %q, want missing --name diagnostic", stderr.String())
	}
}

func TestRunProductionReferencesSymbolUsageFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "missing symbol",
			args: nil,
			want: "missing --symbol",
		},
		{
			name: "empty symbol",
			args: []string{"--symbol", ""},
			want: "--symbol requires a value",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			fixture := traversaltest.LoadSharedFixture(t)
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			args := []string{"references", "--index", fixture.IndexPath}
			args = append(args, test.args...)

			status := run(args, &stdout, &stderr)

			if status != runtimecontract.StatusUsage {
				t.Fatalf("references invalid args status = %d, want %d", status, runtimecontract.StatusUsage)
			}
			if stdout.String() != "" {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
			if !strings.Contains(stderr.String(), test.want) {
				t.Fatalf("stderr = %q, want diagnostic containing %q", stderr.String(), test.want)
			}
		})
	}
}

func TestRunProductionPackagesInvalidFlagsRemainUsageFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "unknown flag",
			args: []string{"--name", "Supervisor"},
			want: "packages only accepts --prefix",
		},
		{
			name: "prefix missing value",
			args: []string{"--prefix"},
			want: "--prefix requires a value",
		},
		{
			name: "duplicate prefix",
			args: []string{"--prefix", "github.com/liza-mas/", "--prefix", "github.com/sourcegraph/"},
			want: "--prefix can only be provided once",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			fixture := traversaltest.LoadSharedFixture(t)
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			args := []string{"packages", "--index", fixture.IndexPath}
			args = append(args, test.args...)

			status := run(args, &stdout, &stderr)

			if status != runtimecontract.StatusUsage {
				t.Fatalf("packages invalid args status = %d, want %d", status, runtimecontract.StatusUsage)
			}
			if stdout.String() != "" {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
			if !strings.Contains(stderr.String(), test.want) {
				t.Fatalf("stderr = %q, want diagnostic containing %q", stderr.String(), test.want)
			}
		})
	}
}

func TestRunProductionImplementationsMissingSymbolRemainsUsageFailure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "missing symbol",
			args: nil,
			want: "missing --symbol",
		},
		{
			name: "empty symbol",
			args: []string{"--symbol", ""},
			want: "--symbol requires a value",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			fixture := traversaltest.LoadSharedFixture(t)
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			args := []string{"implementations", "--index", fixture.IndexPath}
			args = append(args, test.args...)

			status := run(args, &stdout, &stderr)

			if status != runtimecontract.StatusUsage {
				t.Fatalf("implementations invalid args status = %d, want %d", status, runtimecontract.StatusUsage)
			}
			if stdout.String() != "" {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
			if !strings.Contains(stderr.String(), test.want) {
				t.Fatalf("stderr = %q, want diagnostic containing %q", stderr.String(), test.want)
			}
		})
	}
}

type openFailureLoader struct {
	paths []string
}

func (loader *openFailureLoader) Load(indexPath string) (any, error) {
	loader.paths = append(loader.paths, indexPath)

	return nil, errors.New("selected index cannot be opened: deterministic loader-open failure")
}

type boundaryHandler struct {
	command string
	calls   int
	loaded  any
	args    []string
}

func (handler *boundaryHandler) Handle(loadedIndex any, args []string) (any, error) {
	handler.calls++
	handler.loaded = loadedIndex
	handler.args = slices.Clone(args)

	loaded := loadedIndex.(runtimecontract.LoadedIndex)

	return map[string]string{
		"command": handler.command,
		"tool":    loaded.Index.GetMetadata().GetToolInfo().GetName(),
	}, nil
}

func newBoundaryHandlers() map[string]cli.Handler {
	handlers := make(map[string]cli.Handler, len(documentedQueryCommands()))
	for _, command := range documentedQueryCommands() {
		handlers[command] = &boundaryHandler{command: command}
	}

	return handlers
}

func runWithTestRuntime(
	args []string,
	loader cli.Loader,
	handlers map[string]cli.Handler,
	stdout io.Writer,
	stderr io.Writer,
) runtimecontract.Status {
	return cli.NewRuntime(loader, handlers).Run(args, stdout, stderr)
}

func runWithProductionLoader(
	args []string,
	handlers map[string]cli.Handler,
	stdout io.Writer,
	stderr io.Writer,
) runtimecontract.Status {
	return cli.NewProductionRuntime(handlers).Run(args, stdout, stderr)
}

func documentedQueryCommands() []string {
	return []string{"symbols", "references", "implementations", "packages"}
}

func documentedCommandArgs(command string, indexPath string) []string {
	args := []string{command, "--index", indexPath}

	return append(args, documentedQueryArgs(command)...)
}

func documentedQueryArgs(command string) []string {
	switch command {
	case "symbols":
		return []string{"--name", "Supervisor"}
	case "references", "implementations":
		return []string{"--symbol", "scip-go gomod example.com/repo . pkg/Foo#"}
	case "packages":
		return []string{"--prefix", "example.com"}
	default:
		return nil
	}
}

func assertIndexLoadFailure(
	t *testing.T,
	status runtimecontract.Status,
	stdout *bytes.Buffer,
	stderr *bytes.Buffer,
) {
	t.Helper()

	if status != runtimecontract.StatusIndexLoad {
		t.Fatalf("status = %d, want %d", status, runtimecontract.StatusIndexLoad)
	}
	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if stderr.String() == "" {
		t.Fatal("stderr is empty, want index-loading diagnostic")
	}
}

func assertNoBoundaryHandlerCalls(t *testing.T, handlers map[string]cli.Handler) {
	t.Helper()

	for command, handler := range handlers {
		recorder := handler.(*boundaryHandler)
		if recorder.calls != 0 {
			t.Fatalf("%s handler calls = %d, want 0", command, recorder.calls)
		}
	}
}

func assertOtherBoundaryHandlersNotCalled(t *testing.T, handlers map[string]cli.Handler, selectedCommand string) {
	t.Helper()

	for command, handler := range handlers {
		if command == selectedCommand {
			continue
		}

		recorder := handler.(*boundaryHandler)
		if recorder.calls != 0 {
			t.Fatalf("%s handler calls = %d, want 0", command, recorder.calls)
		}
	}
}

func assertSingleJSONCommand(t *testing.T, output []byte, command string) {
	t.Helper()

	decoder := json.NewDecoder(bytes.NewReader(output))
	var got map[string]string
	if err := decoder.Decode(&got); err != nil {
		t.Fatalf("stdout JSON decode failed: %v; output = %q", err, output)
	}
	if got["command"] != command {
		t.Fatalf("stdout JSON command = %q, want %q", got["command"], command)
	}
	if got["tool"] != command {
		t.Fatalf("stdout JSON tool = %q, want %q", got["tool"], command)
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		t.Fatalf("stdout contains extra JSON or non-JSON content after first value: %v", err)
	}
}

func assertProductionJSONMatchesGolden(t *testing.T, output []byte, goldenFile string) {
	t.Helper()

	got := decodeJSONValue(t, output)
	want := readDiscoveryGoldenJSONValue(t, goldenFile)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("production JSON value = %#v, want golden %#v", got, want)
	}
}

func assertProductionJSONMatchesReferenceGolden(t *testing.T, output []byte, goldenFile string) {
	t.Helper()

	got := decodeJSONValue(t, output)
	want := readReferenceGoldenJSONValue(t, goldenFile)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("production JSON value = %#v, want golden %#v", got, want)
	}
}

func readDiscoveryGoldenJSONValue(t *testing.T, goldenFile string) any {
	t.Helper()

	payload, err := os.ReadFile(filepath.Join("..", "..", "internal", "query", "discovery", "testdata", "golden", goldenFile))
	if err != nil {
		t.Fatalf("read discovery golden file %q: %v", goldenFile, err)
	}

	return decodeJSONValue(t, payload)
}

func assertProductionImplementationsJSONMatchesGolden(t *testing.T, output []byte, goldenFile string) {
	t.Helper()

	got := decodeJSONValue(t, output)
	want := readImplementationsGoldenJSONValue(t, goldenFile)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("production implementations JSON value = %#v, want golden %#v", got, want)
	}
}

func readImplementationsGoldenJSONValue(t *testing.T, goldenFile string) any {
	t.Helper()

	payload, err := os.ReadFile(filepath.Join("..", "..", "internal", "query", "implementations", "testdata", "golden", goldenFile))
	if err != nil {
		t.Fatalf("read implementations golden file %q: %v", goldenFile, err)
	}

	return decodeJSONValue(t, payload)
}

func readReferenceGoldenJSONValue(t *testing.T, goldenFile string) any {
	t.Helper()

	payload, err := os.ReadFile(filepath.Join("..", "..", "internal", "query", "references", "testdata", "golden", goldenFile))
	if err != nil {
		t.Fatalf("read references golden file %q: %v", goldenFile, err)
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

func writeSelectedIndexBytes(t *testing.T, contents []byte) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "selected.scip")
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v", path, err)
	}

	return path
}

func writeValidSelectedSCIPIndex(t *testing.T, toolName string) string {
	t.Helper()

	indexBytes, err := proto.Marshal(&scip.Index{
		Metadata: &scip.Metadata{
			ToolInfo: &scip.ToolInfo{
				Name:    toolName,
				Version: "v1",
			},
			TextDocumentEncoding: scip.TextEncoding_UTF8,
		},
	})
	if err != nil {
		t.Fatalf("proto.Marshal(valid SCIP index) error = %v", err)
	}

	return writeSelectedIndexBytes(t, indexBytes)
}

func assertSelectedIndexBytes(t *testing.T, path string, want []byte) {
	t.Helper()

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", path, err)
	}
	if !slices.Equal(got, want) {
		t.Fatalf("selected index bytes after run = %v, want unchanged %v", got, want)
	}
}
