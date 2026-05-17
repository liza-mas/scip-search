package cli

import (
	"bytes"
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

type recordingLoader struct {
	calls int
}

func (loader *recordingLoader) Load(string) (any, error) {
	loader.calls++

	return struct{}{}, nil
}

type recordingHandler struct {
	name  string
	calls int
}

func (handler *recordingHandler) Handle(any, []string) (any, error) {
	handler.calls++

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
