package runtime

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// RawOutput is a successful result that has already been rendered for stdout.
type RawOutput struct {
	Text string
}

// JSONFileOutput is a successful result that should be rendered as JSON to a file.
type JSONFileOutput struct {
	Path  string
	Value any
}

// WriteDiagnostic writes a shared runtime failure diagnostic to stderr.
func WriteDiagnostic(stderr io.Writer, failure Failure) Status {
	fmt.Fprintln(stderr, failure.Message())

	return failure.Status()
}

// WriteRawSuccess writes pre-rendered successful output to stdout.
func WriteRawSuccess(stdout io.Writer, text string) (Status, error) {
	if _, err := fmt.Fprint(stdout, text); err != nil {
		return StatusOK, err
	}

	return StatusOK, nil
}

// WriteJSONSuccess writes one compact JSON value plus a trailing newline to stdout.
func WriteJSONSuccess(stdout io.Writer, result any) (Status, error) {
	if err := json.NewEncoder(stdout).Encode(result); err != nil {
		return StatusOK, err
	}

	return StatusOK, nil
}

// WriteJSONFileSuccess writes one compact JSON value plus a trailing newline to path.
func WriteJSONFileSuccess(path string, result any) (Status, error) {
	file, err := os.Create(path)
	if err != nil {
		return StatusOK, err
	}

	if err := json.NewEncoder(file).Encode(result); err != nil {
		_ = file.Close()
		return StatusOK, err
	}

	if err := file.Close(); err != nil {
		return StatusOK, err
	}

	return StatusOK, nil
}
