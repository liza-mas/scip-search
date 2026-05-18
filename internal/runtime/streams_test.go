package runtime

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"
)

func TestStatusConstants(t *testing.T) {
	t.Parallel()

	if StatusOK != 0 {
		t.Fatalf("StatusOK = %d, want 0", StatusOK)
	}
	if StatusUsage != 2 {
		t.Fatalf("StatusUsage = %d, want 2", StatusUsage)
	}
	if StatusIndexLoad != 3 {
		t.Fatalf("StatusIndexLoad = %d, want 3", StatusIndexLoad)
	}
}

func TestWriteDiagnosticWritesUsageFailureToStderrOnly(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	status := WriteDiagnostic(&stderr, UsageFailure("missing command"))

	if status != StatusUsage {
		t.Fatalf("WriteDiagnostic() status = %d, want %d", status, StatusUsage)
	}
	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if stderr.String() != "missing command\n" {
		t.Fatalf("stderr = %q, want diagnostic with trailing newline", stderr.String())
	}
}

func TestWriteDiagnosticWritesIndexLoadFailureToStderrOnly(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	status := WriteDiagnostic(&stderr, IndexLoadFailure("load index: invalid SCIP data"))

	if status != StatusIndexLoad {
		t.Fatalf("WriteDiagnostic() status = %d, want %d", status, StatusIndexLoad)
	}
	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if stderr.String() != "load index: invalid SCIP data\n" {
		t.Fatalf("stderr = %q, want diagnostic with trailing newline", stderr.String())
	}
}

func TestWriteJSONSuccessEmitsOneParseableValueAndNewline(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	result := map[string]any{
		"items": []any{
			map[string]any{"name": "first"},
		},
		"ok": true,
	}

	status, err := WriteJSONSuccess(&stdout, result)
	if err != nil {
		t.Fatalf("WriteJSONSuccess() error = %v", err)
	}

	if status != StatusOK {
		t.Fatalf("WriteJSONSuccess() status = %d, want %d", status, StatusOK)
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if stdout.String() != "{\"items\":[{\"name\":\"first\"}],\"ok\":true}\n" {
		t.Fatalf("stdout = %q, want exactly one compact JSON value plus newline", stdout.String())
	}

	decoder := json.NewDecoder(bytes.NewReader(stdout.Bytes()))
	var decoded map[string]any
	if err := decoder.Decode(&decoded); err != nil {
		t.Fatalf("json.Decode(stdout) error = %v", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		t.Fatalf("stdout contains more than one JSON value: decode err = %v", err)
	}
}

func TestWriteRawSuccessWritesTextExactly(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	text := "first line\nsecond line\n"

	status, err := WriteRawSuccess(&stdout, text)
	if err != nil {
		t.Fatalf("WriteRawSuccess() error = %v", err)
	}

	if status != StatusOK {
		t.Fatalf("WriteRawSuccess() status = %d, want %d", status, StatusOK)
	}
	if stdout.String() != text {
		t.Fatalf("stdout = %q, want exact raw text %q", stdout.String(), text)
	}
}

func TestWriteJSONSuccessAcceptsAnonymousStructResults(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	result := struct {
		Count int    `json:"count"`
		Label string `json:"label"`
	}{
		Count: 2,
		Label: "sample",
	}

	status, err := WriteJSONSuccess(&stdout, result)
	if err != nil {
		t.Fatalf("WriteJSONSuccess() error = %v", err)
	}

	if status != StatusOK {
		t.Fatalf("WriteJSONSuccess() status = %d, want %d", status, StatusOK)
	}
	if stdout.String() != "{\"count\":2,\"label\":\"sample\"}\n" {
		t.Fatalf("stdout = %q, want anonymous struct JSON plus newline", stdout.String())
	}
}
