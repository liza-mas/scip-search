package runtime

import (
	"encoding/json"
	"fmt"
	"io"
)

// RawOutput is a successful result that has already been rendered for stdout.
type RawOutput struct {
	Text string
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
