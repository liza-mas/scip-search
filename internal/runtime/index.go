package runtime

import "github.com/scip-code/scip/bindings/go/scip"

// LoadedIndex is the shared runtime context produced from the caller-selected SCIP file.
type LoadedIndex struct {
	Path        string
	Fingerprint string
	Index       *scip.Index
}
