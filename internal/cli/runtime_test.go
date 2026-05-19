package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/scip-code/scip/bindings/go/scip"
	"google.golang.org/protobuf/proto"

	runtimecontract "scip-search/internal/runtime"
	"scip-search/internal/version"
)

func TestDocumentedCommandsRouteThroughSharedRegistry(t *testing.T) {
	t.Parallel()

	handlers := map[string]Handler{
		"symbols":         &recordingHandler{name: "symbols"},
		"references":      &recordingHandler{name: "references"},
		"implementations": &recordingHandler{name: "implementations"},
		"packages":        &recordingHandler{name: "packages"},
	}
	cliRuntime := NewRuntime(&recordingLoader{}, handlers)

	for _, command := range []string{"symbols", "references", "implementations", "packages"} {
		t.Run(command, func(t *testing.T) {
			t.Parallel()

			selected, remainingArgs, ok := cliRuntime.Route([]string{command, "--index", "repo.scip"})
			if !ok {
				t.Fatalf("Route(%q) was rejected, want recognized command", command)
			}
			if selected.Name != command {
				t.Fatalf("selected command = %q, want %q", selected.Name, command)
			}
			if selected.Handler != handlers[command] {
				t.Fatalf("selected handler for %q did not come from command registry", command)
			}
			if got, want := remainingArgs, []string{"--index", "repo.scip"}; !slices.Equal(got, want) {
				t.Fatalf("remaining args = %v, want %v", got, want)
			}
		})
	}
}

func TestRunWithoutCommandReturnsUsageBeforeLoaderOrHandlers(t *testing.T) {
	t.Parallel()

	loader := &recordingLoader{}
	handlers := newRecordingHandlers()
	cliRuntime := NewRuntime(loader, handlers)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	status := cliRuntime.Run(nil, &stdout, &stderr)

	if status != runtimecontract.StatusUsage {
		t.Fatalf("Run() status = %d, want %d", status, runtimecontract.StatusUsage)
	}
	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if stderr.String() == "" {
		t.Fatal("stderr is empty, want usage diagnostic")
	}
	if loader.calls != 0 {
		t.Fatalf("loader calls = %d, want 0", loader.calls)
	}
	assertNoHandlerCalls(t, handlers)
}

func TestRunWithUnsupportedCommandReturnsUsageBeforeLoaderOrHandlers(t *testing.T) {
	t.Parallel()

	loader := &recordingLoader{}
	handlers := newRecordingHandlers()
	cliRuntime := NewRuntime(loader, handlers)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	status := cliRuntime.Run([]string{"search", "--index", "repo.scip", "--name", "Supervisor"}, &stdout, &stderr)

	if status != runtimecontract.StatusUsage {
		t.Fatalf("Run() status = %d, want %d", status, runtimecontract.StatusUsage)
	}
	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if stderr.String() == "" {
		t.Fatal("stderr is empty, want usage diagnostic")
	}
	if loader.calls != 0 {
		t.Fatalf("loader calls = %d, want 0", loader.calls)
	}
	assertNoHandlerCalls(t, handlers)
}

func TestRunHelpBypassesQueryValidationLoaderAndHandlers(t *testing.T) {
	t.Parallel()

	loader := &recordingLoader{}
	handlers := newRecordingHandlers()
	cliRuntime := NewRuntime(loader, handlers)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	status := cliRuntime.Run([]string{"--help"}, &stdout, &stderr)

	if status != runtimecontract.StatusOK {
		t.Fatalf("Run(--help) status = %d, want %d", status, runtimecontract.StatusOK)
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	for _, want := range []string{
		"Description:",
		"Search a pre-built SCIP index for symbols, references, implementations, and packages.",
		"Usage:",
		"scip-search symbols --index <index-path> --name <name> [--name <name>]... [--one-line|--nested-json|--json]",
		"scip-search references --index <index-path> [--symbol <scip-symbol>]... [--name <name>]... [--one-line|--json|--location-only]",
		"scip-search implementations --index <index-path> [--symbol <scip-symbol>]... [--name <name>]... [--one-line|--json|--location-only]",
		"scip-search packages --index <index-path> [--prefix <prefix>] [--one-line|--json]",
		"Output:",
		"--one-line     Grep-style text output; default for all query commands.",
		"--location-only  Location-only text output for exact-symbol references and implementations.",
		"--nested-json  Compact package-grouped JSON output for symbols only.",
		"One-line formats:",
		"symbols          <path>:<line>:<column>:<packageKey> <descriptor> match=<source> text=<text>",
		"references       <path>:<line>:<column>:<referenced-symbol> roles=<roles>",
		"implementations  <path>:<line>:<column>:<implementation-symbol>",
		"location-only    <path>:<line>:<column>",
		"symbols accepts repeated --name; references and implementations accept repeated --name and --symbol.",
		"--location-only for references and implementations requires --symbol and cannot be used with --name.",
		"Repeated results are de-duplicated.",
		"references and implementations require --symbol, --name, or both.",
		"Reads an existing SCIP index; does not generate, update, or discover indexes.",
		"Exit codes:",
		"2  usage error",
		"Examples:",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want substring %q", stdout.String(), want)
		}
	}
	if loader.calls != 0 {
		t.Fatalf("loader calls = %d, want 0", loader.calls)
	}
	assertNoHandlerCalls(t, handlers)
	assertNotQueryJSON(t, stdout.String())
}

func TestRunVersionBypassesQueryValidationLoaderAndHandlers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		identity version.BuildIdentity
		want     []string
	}{
		{
			name: "release build",
			identity: version.BuildIdentity{
				Release: "v9.8.7",
			},
			want: []string{"scip-search", "release", "v9.8.7"},
		},
		{
			name: "source build",
			identity: version.BuildIdentity{
				SourceRef:      "main",
				SourceRevision: "abc123",
			},
			want: []string{"scip-search", "source", "main", "abc123"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			loader := &recordingLoader{}
			handlers := newRecordingHandlers()
			cliRuntime := NewRuntimeWithBuildIdentity(loader, handlers, test.identity)
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			status := cliRuntime.Run([]string{"--version"}, &stdout, &stderr)

			if status != runtimecontract.StatusOK {
				t.Fatalf("Run(--version) status = %d, want %d", status, runtimecontract.StatusOK)
			}
			if stdout.String() == "" {
				t.Fatal("stdout is empty, want version identity")
			}
			for _, want := range test.want {
				if !strings.Contains(stdout.String(), want) {
					t.Fatalf("stdout = %q, want substring %q", stdout.String(), want)
				}
			}
			if stderr.String() != "" {
				t.Fatalf("stderr = %q, want empty", stderr.String())
			}
			if loader.calls != 0 {
				t.Fatalf("loader calls = %d, want 0", loader.calls)
			}
			assertNoHandlerCalls(t, handlers)
			assertNotQueryJSON(t, stdout.String())
		})
	}
}

func TestRunVersionDistinguishesReleaseFromSourceBuilds(t *testing.T) {
	t.Parallel()

	releaseRuntime := NewRuntimeWithBuildIdentity(&recordingLoader{}, newRecordingHandlers(), version.BuildIdentity{
		Release: "v1.0.0",
	})
	sourceRuntime := NewRuntimeWithBuildIdentity(&recordingLoader{}, newRecordingHandlers(), version.BuildIdentity{
		SourceRef:      "main",
		SourceRevision: "def456",
	})
	var releaseStdout bytes.Buffer
	var sourceStdout bytes.Buffer

	releaseStatus := releaseRuntime.Run([]string{"--version"}, &releaseStdout, io.Discard)
	sourceStatus := sourceRuntime.Run([]string{"--version"}, &sourceStdout, io.Discard)

	if releaseStatus != runtimecontract.StatusOK {
		t.Fatalf("release --version status = %d, want %d", releaseStatus, runtimecontract.StatusOK)
	}
	if sourceStatus != runtimecontract.StatusOK {
		t.Fatalf("source --version status = %d, want %d", sourceStatus, runtimecontract.StatusOK)
	}
	if releaseStdout.String() == sourceStdout.String() {
		t.Fatalf("release and source version output both = %q, want distinct output", releaseStdout.String())
	}
}

func TestRunRequiresIndexForEveryDocumentedCommand(t *testing.T) {
	t.Parallel()

	for _, command := range []string{"symbols", "references", "implementations", "packages"} {
		t.Run(command, func(t *testing.T) {
			t.Parallel()

			loader := &recordingLoader{}
			handlers := newRecordingHandlers()
			cliRuntime := NewRuntime(loader, handlers)
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			status := cliRuntime.Run([]string{command, "--name", "Supervisor"}, &stdout, &stderr)

			assertUsageFailureBeforeLoaderAndHandlers(t, status, &stdout, &stderr, loader, handlers)
		})
	}
}

func TestRunRejectsIndexWithoutValueForEveryDocumentedCommand(t *testing.T) {
	t.Parallel()

	for _, command := range []string{"symbols", "references", "implementations", "packages"} {
		t.Run(command, func(t *testing.T) {
			t.Parallel()

			loader := &recordingLoader{}
			handlers := newRecordingHandlers()
			cliRuntime := NewRuntime(loader, handlers)
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			status := cliRuntime.Run([]string{command, "--index"}, &stdout, &stderr)

			assertUsageFailureBeforeLoaderAndHandlers(t, status, &stdout, &stderr, loader, handlers)
		})
	}
}

func TestRunLoaderFailuresReturnStatusIndexLoadBeforeHandlers(t *testing.T) {
	t.Parallel()

	failures := []struct {
		name         string
		diagnostic   string
		selectedPath func(t *testing.T) string
		assertPath   func(t *testing.T, path string)
	}{
		{
			name:       "nonexistent",
			diagnostic: "selected index does not exist",
			selectedPath: func(t *testing.T) string {
				t.Helper()

				return filepath.Join(t.TempDir(), "missing.scip")
			},
			assertPath: func(t *testing.T, path string) {
				t.Helper()

				if _, err := os.Stat(path); !os.IsNotExist(err) {
					t.Fatalf("selected missing path stat error = %v, want not exist", err)
				}
			},
		},
		{
			name:       "unreadable",
			diagnostic: "selected index cannot be opened",
			selectedPath: func(t *testing.T) string {
				t.Helper()

				return writeSelectedIndexBytes(t, []byte("unreadable sentinel"))
			},
			assertPath: func(t *testing.T, path string) {
				t.Helper()

				assertSelectedIndexBytes(t, path, []byte("unreadable sentinel"))
			},
		},
		{
			name:       "directory",
			diagnostic: "selected index is a directory",
			selectedPath: func(t *testing.T) string {
				t.Helper()

				return t.TempDir()
			},
			assertPath: func(t *testing.T, path string) {
				t.Helper()

				info, err := os.Stat(path)
				if err != nil {
					t.Fatalf("selected directory stat error = %v", err)
				}
				if !info.IsDir() {
					t.Fatalf("selected path %q is not a directory after Run()", path)
				}
			},
		},
		{
			name:       "invalid_scip",
			diagnostic: "selected index is not valid SCIP",
			selectedPath: func(t *testing.T) string {
				t.Helper()

				return writeSelectedIndexBytes(t, []byte("not a SCIP protobuf"))
			},
			assertPath: func(t *testing.T, path string) {
				t.Helper()

				assertSelectedIndexBytes(t, path, []byte("not a SCIP protobuf"))
			},
		},
	}

	for _, command := range []string{"symbols", "references", "implementations", "packages"} {
		for _, failure := range failures {
			t.Run(command+"/"+failure.name, func(t *testing.T) {
				t.Parallel()

				selectedPath := failure.selectedPath(t)
				loader := &recordingLoader{err: errors.New(failure.diagnostic)}
				handlers := newRecordingHandlers()
				cliRuntime := NewRuntime(loader, handlers)
				var stdout bytes.Buffer
				var stderr bytes.Buffer

				status := cliRuntime.Run(
					[]string{command, "--index", selectedPath, "--opaque-query-arg", "value"},
					&stdout,
					&stderr,
				)

				assertIndexLoadFailureBeforeHandlers(t, status, &stdout, &stderr, loader, handlers, selectedPath)
				if !strings.Contains(stderr.String(), failure.diagnostic) {
					t.Fatalf("stderr = %q, want loader diagnostic substring %q", stderr.String(), failure.diagnostic)
				}
				failure.assertPath(t, selectedPath)
			})
		}
	}
}

func TestProductionRuntimeReportsRealLoaderFailuresBeforeHandlers(t *testing.T) {
	t.Parallel()

	failures := []struct {
		name         string
		selectedPath func(t *testing.T) string
		assertPath   func(t *testing.T, path string)
	}{
		{
			name: "nonexistent",
			selectedPath: func(t *testing.T) string {
				t.Helper()

				return filepath.Join(t.TempDir(), "missing.scip")
			},
			assertPath: func(t *testing.T, path string) {
				t.Helper()

				if _, err := os.Stat(path); !os.IsNotExist(err) {
					t.Fatalf("selected missing path stat error = %v, want not exist", err)
				}
			},
		},
		{
			name: "directory",
			selectedPath: func(t *testing.T) string {
				t.Helper()

				return t.TempDir()
			},
			assertPath: func(t *testing.T, path string) {
				t.Helper()

				info, err := os.Stat(path)
				if err != nil {
					t.Fatalf("selected directory stat error = %v", err)
				}
				if !info.IsDir() {
					t.Fatalf("selected path %q is not a directory after Run()", path)
				}
			},
		},
		{
			name: "invalid_scip",
			selectedPath: func(t *testing.T) string {
				t.Helper()

				return writeSelectedIndexBytes(t, []byte("not a SCIP protobuf"))
			},
			assertPath: func(t *testing.T, path string) {
				t.Helper()

				assertSelectedIndexBytes(t, path, []byte("not a SCIP protobuf"))
			},
		},
	}

	for _, command := range []string{"symbols", "references", "implementations", "packages"} {
		for _, failure := range failures {
			t.Run(command+"/"+failure.name, func(t *testing.T) {
				t.Parallel()

				selectedPath := failure.selectedPath(t)
				handlers := newRecordingHandlers()
				cliRuntime := NewProductionRuntime(handlers)
				var stdout bytes.Buffer
				var stderr bytes.Buffer

				status := cliRuntime.Run([]string{command, "--index", selectedPath}, &stdout, &stderr)

				assertProductionIndexLoadFailureBeforeHandlers(t, status, &stdout, &stderr, handlers)
				failure.assertPath(t, selectedPath)
			})
		}
	}
}

func TestProductionRuntimeLoadsSelectedSCIPBeforeSelectedHandler(t *testing.T) {
	t.Parallel()

	for _, command := range []string{"symbols", "references", "implementations", "packages"} {
		t.Run(command, func(t *testing.T) {
			t.Parallel()

			selectedPath, selectedBytes := writeValidSelectedSCIPIndex(t, command)
			handlers := newRecordingHandlers()
			selected := handlers[command].(*recordingHandler)
			selected.result = map[string]string{"command": command}
			cliRuntime := NewProductionRuntime(handlers)
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			status := cliRuntime.Run(
				[]string{command, "--index", selectedPath, "--opaque-query-arg", "value"},
				&stdout,
				&stderr,
			)

			if status != runtimecontract.StatusOK {
				t.Fatalf("Run() status = %d, want %d", status, runtimecontract.StatusOK)
			}
			if stderr.String() != "" {
				t.Fatalf("stderr = %q, want empty", stderr.String())
			}
			if selected.calls != 1 {
				t.Fatalf("%s handler calls = %d, want 1", command, selected.calls)
			}
			loaded, ok := selected.loaded.(runtimecontract.LoadedIndex)
			if !ok {
				t.Fatalf("%s handler loaded context = %T, want runtime LoadedIndex", command, selected.loaded)
			}
			if loaded.Path != selectedPath {
				t.Fatalf("%s loaded path = %q, want caller-selected path %q", command, loaded.Path, selectedPath)
			}
			if loaded.Index.GetMetadata().GetToolInfo().GetName() != command {
				t.Fatalf("%s loaded index tool name = %q, want %q", command, loaded.Index.GetMetadata().GetToolInfo().GetName(), command)
			}
			if got, want := selected.args, []string{"--opaque-query-arg", "value"}; !slices.Equal(got, want) {
				t.Fatalf("%s handler args = %v, want %v", command, got, want)
			}
			assertOtherHandlersNotCalled(t, handlers, command)
			assertSingleJSONValue(t, stdout.Bytes(), map[string]string{"command": command})
			assertSelectedIndexBytes(t, selectedPath, selectedBytes)
		})
	}
}

func TestRunWithIndexLoadsAndExecutesOnlySelectedHandler(t *testing.T) {
	t.Parallel()

	for _, command := range []string{"symbols", "references", "implementations", "packages"} {
		t.Run(command, func(t *testing.T) {
			t.Parallel()

			loaded := &loadedContext{id: command}
			loader := &recordingLoader{loaded: loaded}
			handlers := newRecordingHandlers()
			selected := handlers[command].(*recordingHandler)
			selected.result = map[string]string{"command": command}
			cliRuntime := NewRuntime(loader, handlers)
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			status := cliRuntime.Run(
				[]string{command, "--name", "references", "--index", "/tmp/repo.scip", "--limit", "10"},
				&stdout,
				&stderr,
			)

			if status != runtimecontract.StatusOK {
				t.Fatalf("Run() status = %d, want %d", status, runtimecontract.StatusOK)
			}
			if stderr.String() != "" {
				t.Fatalf("stderr = %q, want empty", stderr.String())
			}
			if got, want := loader.paths, []string{"/tmp/repo.scip"}; !slices.Equal(got, want) {
				t.Fatalf("loader paths = %v, want %v", got, want)
			}
			if selected.calls != 1 {
				t.Fatalf("%s handler calls = %d, want 1", command, selected.calls)
			}
			if selected.loaded != loaded {
				t.Fatalf("%s handler loaded context = %#v, want %#v", command, selected.loaded, loaded)
			}
			if got, want := selected.args, []string{"--name", "references", "--limit", "10"}; !slices.Equal(got, want) {
				t.Fatalf("%s handler args = %v, want %v", command, got, want)
			}
			assertOtherHandlersNotCalled(t, handlers, command)
			assertSingleJSONValue(t, stdout.Bytes(), map[string]string{"command": command})
		})
	}
}

func TestRunWritesRawOutputWithoutJSONEncoding(t *testing.T) {
	t.Parallel()

	loader := &recordingLoader{loaded: &loadedContext{id: "symbols"}}
	handlers := newRecordingHandlers()
	selected := handlers["symbols"].(*recordingHandler)
	selected.result = runtimecontract.RawOutput{Text: "path.go:1:1:symbol\n"}
	cliRuntime := NewRuntime(loader, handlers)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	status := cliRuntime.Run([]string{"symbols", "--index", "/tmp/repo.scip", "--name", "Symbol"}, &stdout, &stderr)

	if status != runtimecontract.StatusOK {
		t.Fatalf("Run() status = %d, want %d", status, runtimecontract.StatusOK)
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if stdout.String() != "path.go:1:1:symbol\n" {
		t.Fatalf("stdout = %q, want raw output without JSON encoding", stdout.String())
	}
}

func TestRunRejectsSecondDocumentedCommandTokenBeforeLoaderOrHandlers(t *testing.T) {
	t.Parallel()

	loader := &recordingLoader{}
	handlers := newRecordingHandlers()
	cliRuntime := NewRuntime(loader, handlers)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	status := cliRuntime.Run(
		[]string{"symbols", "--index", "repo.scip", "references", "--symbol", "scip-go gomod example . pkg/Foo#"},
		&stdout,
		&stderr,
	)

	assertUsageFailureBeforeLoaderAndHandlers(t, status, &stdout, &stderr, loader, handlers)
}

type loadedContext struct {
	id string
}

type recordingLoader struct {
	calls  int
	paths  []string
	loaded any
	err    error
}

func (loader *recordingLoader) Load(indexPath string) (any, error) {
	loader.calls++
	loader.paths = append(loader.paths, indexPath)

	if loader.err != nil {
		return nil, loader.err
	}

	return loader.loaded, nil
}

type recordingHandler struct {
	name   string
	calls  int
	loaded any
	args   []string
	result any
}

func (handler *recordingHandler) Handle(loadedIndex any, args []string) (any, error) {
	handler.calls++
	handler.loaded = loadedIndex
	handler.args = slices.Clone(args)

	if handler.result != nil {
		return handler.result, nil
	}

	return map[string]string{"command": handler.name}, nil
}

func newRecordingHandlers() map[string]Handler {
	return map[string]Handler{
		"symbols":         &recordingHandler{name: "symbols"},
		"references":      &recordingHandler{name: "references"},
		"implementations": &recordingHandler{name: "implementations"},
		"packages":        &recordingHandler{name: "packages"},
	}
}

func assertNoHandlerCalls(t *testing.T, handlers map[string]Handler) {
	t.Helper()

	for command, handler := range handlers {
		recorder := handler.(*recordingHandler)
		if recorder.calls != 0 {
			t.Fatalf("%s handler calls = %d, want 0", command, recorder.calls)
		}
	}
}

func assertOtherHandlersNotCalled(t *testing.T, handlers map[string]Handler, selectedCommand string) {
	t.Helper()

	for command, handler := range handlers {
		if command == selectedCommand {
			continue
		}

		recorder := handler.(*recordingHandler)
		if recorder.calls != 0 {
			t.Fatalf("%s handler calls = %d, want 0", command, recorder.calls)
		}
	}
}

func assertUsageFailureBeforeLoaderAndHandlers(
	t *testing.T,
	status runtimecontract.Status,
	stdout *bytes.Buffer,
	stderr *bytes.Buffer,
	loader *recordingLoader,
	handlers map[string]Handler,
) {
	t.Helper()

	if status != runtimecontract.StatusUsage {
		t.Fatalf("Run() status = %d, want %d", status, runtimecontract.StatusUsage)
	}
	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if stderr.String() == "" {
		t.Fatal("stderr is empty, want usage diagnostic")
	}
	if loader.calls != 0 {
		t.Fatalf("loader calls = %d, want 0", loader.calls)
	}
	assertNoHandlerCalls(t, handlers)
}

func assertIndexLoadFailureBeforeHandlers(
	t *testing.T,
	status runtimecontract.Status,
	stdout *bytes.Buffer,
	stderr *bytes.Buffer,
	loader *recordingLoader,
	handlers map[string]Handler,
	selectedPath string,
) {
	t.Helper()

	if status != runtimecontract.StatusIndexLoad {
		t.Fatalf("Run() status = %d, want %d", status, runtimecontract.StatusIndexLoad)
	}
	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if stderr.String() == "" {
		t.Fatal("stderr is empty, want index-loading diagnostic")
	}
	if got, want := loader.paths, []string{selectedPath}; !slices.Equal(got, want) {
		t.Fatalf("loader paths = %v, want only caller-selected path %q", got, selectedPath)
	}
	assertNoHandlerCalls(t, handlers)
}

func assertProductionIndexLoadFailureBeforeHandlers(
	t *testing.T,
	status runtimecontract.Status,
	stdout *bytes.Buffer,
	stderr *bytes.Buffer,
	handlers map[string]Handler,
) {
	t.Helper()

	if status != runtimecontract.StatusIndexLoad {
		t.Fatalf("Run() status = %d, want %d", status, runtimecontract.StatusIndexLoad)
	}
	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if stderr.String() == "" {
		t.Fatal("stderr is empty, want production loader diagnostic")
	}
	assertNoHandlerCalls(t, handlers)
}

func assertSingleJSONValue(t *testing.T, output []byte, want map[string]string) {
	t.Helper()

	decoder := json.NewDecoder(bytes.NewReader(output))
	var got map[string]string
	if err := decoder.Decode(&got); err != nil {
		t.Fatalf("stdout JSON decode failed: %v; output = %q", err, output)
	}
	if got["command"] != want["command"] {
		t.Fatalf("stdout JSON command = %q, want %q", got["command"], want["command"])
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		t.Fatalf("stdout contains extra JSON or non-JSON content after first value: %v", err)
	}
}

func assertNotQueryJSON(t *testing.T, output string) {
	t.Helper()

	var decoded map[string]any
	if err := json.Unmarshal([]byte(output), &decoded); err == nil {
		t.Fatalf("stdout = %q, want human-readable version output instead of query JSON", output)
	}
}

func writeSelectedIndexBytes(t *testing.T, contents []byte) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "selected.scip")
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v", path, err)
	}

	return path
}

func writeValidSelectedSCIPIndex(t *testing.T, toolName string) (string, []byte) {
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

	return writeSelectedIndexBytes(t, indexBytes), indexBytes
}

func assertSelectedIndexBytes(t *testing.T, path string, want []byte) {
	t.Helper()

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", path, err)
	}
	if !slices.Equal(got, want) {
		t.Fatalf("selected index bytes after Run() = %v, want unchanged %v", got, want)
	}
}
