package runtime

import (
	"encoding/json"
	"fmt"
	"io"
)

// WriteDiagnostic writes a shared runtime failure diagnostic to stderr.
func WriteDiagnostic(stderr io.Writer, failure Failure) Status {
	fmt.Fprintln(stderr, failure.Message())

	return failure.Status()
}

// WriteJSONSuccess writes one compact JSON value plus a trailing newline to stdout.
func WriteJSONSuccess(stdout io.Writer, result any) (Status, error) {
	if err := json.NewEncoder(stdout).Encode(result); err != nil {
		return StatusOK, err
	}

	return StatusOK, nil
}
