package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"slices"
	"testing"

	runtimecontract "scip-search/internal/runtime"
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
}

func (loader *recordingLoader) Load(indexPath string) (any, error) {
	loader.calls++
	loader.paths = append(loader.paths, indexPath)

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
