package scipindex

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	"github.com/scip-code/scip/bindings/go/scip"
	"google.golang.org/protobuf/proto"

	runtimecontract "scip-search/internal/runtime"
)

// Loader reads the caller-selected SCIP index file and decodes it with official bindings.
type Loader struct {
	openFile func(string) (io.ReadCloser, error)
}

var defaultOpenFile = func(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

// NewLoader creates the production SCIP index loader.
func NewLoader() Loader {
	return newLoaderWithOpenFile(defaultOpenFile)
}

func newLoaderWithOpenFile(openFile func(string) (io.ReadCloser, error)) Loader {
	if openFile == nil {
		openFile = defaultOpenFile
	}

	return Loader{openFile: openFile}
}

// Load satisfies the shared CLI loader interface without exposing query traversal details.
func (loader Loader) Load(indexPath string) (any, error) {
	return loader.LoadIndex(indexPath)
}

// LoadIndex validates and decodes only the caller-selected SCIP index file.
func (loader Loader) LoadIndex(indexPath string) (runtimecontract.LoadedIndex, error) {
	fileInfo, err := os.Stat(indexPath)
	if err != nil {
		return runtimecontract.LoadedIndex{}, indexLoadError(indexPath, "stat selected index", err)
	}
	if fileInfo.IsDir() {
		return runtimecontract.LoadedIndex{}, runtimecontract.IndexLoadFailure(
			fmt.Sprintf("load index %q: selected path is a directory", indexPath),
		)
	}

	indexFile, err := loader.openFile(indexPath)
	if err != nil {
		return runtimecontract.LoadedIndex{}, indexLoadError(indexPath, "open selected index", err)
	}
	defer indexFile.Close()

	indexBytes, err := io.ReadAll(indexFile)
	if err != nil {
		return runtimecontract.LoadedIndex{}, indexLoadError(indexPath, "read selected index", err)
	}

	index := &scip.Index{}
	if err := proto.Unmarshal(indexBytes, index); err != nil {
		return runtimecontract.LoadedIndex{}, indexLoadError(indexPath, "parse SCIP protobuf", err)
	}

	return runtimecontract.LoadedIndex{
		Path:        indexPath,
		Fingerprint: fingerprint(indexBytes),
		Index:       index,
	}, nil
}

func fingerprint(contents []byte) string {
	sum := sha256.Sum256(contents)
	return fmt.Sprintf("sha256:%x", sum)
}

func indexLoadError(indexPath string, operation string, cause error) error {
	return fmt.Errorf(
		"%w: %w",
		runtimecontract.IndexLoadFailure(fmt.Sprintf("load index %q: %s", indexPath, operation)),
		cause,
	)
}
