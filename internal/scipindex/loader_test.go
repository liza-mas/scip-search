package scipindex

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/scip-code/scip/bindings/go/scip"
	"google.golang.org/protobuf/proto"

	runtimecontract "scip-search/internal/runtime"
)

func TestLoadIndexRejectsMissingPathAsIndexLoadFailureWithoutOpeningFallbacks(t *testing.T) {
	t.Parallel()

	selectedPath := filepath.Join(t.TempDir(), "missing.scip")
	loader := newLoaderWithOpenFile(func(path string) (io.ReadCloser, error) {
		t.Fatalf("openFile(%q) called for missing selected path", path)
		return nil, nil
	})

	_, err := loader.LoadIndex(selectedPath)

	assertIndexLoadFailure(t, err)
}

func TestLoadIndexRejectsDirectoryAsIndexLoadFailureWithoutReadingIt(t *testing.T) {
	t.Parallel()

	selectedPath := t.TempDir()
	loader := newLoaderWithOpenFile(func(path string) (io.ReadCloser, error) {
		t.Fatalf("openFile(%q) called for directory selected path", path)
		return nil, nil
	})

	_, err := loader.LoadIndex(selectedPath)

	assertIndexLoadFailure(t, err)
}

func TestLoadIndexRejectsInjectedUnreadableOpenErrorAsIndexLoadFailure(t *testing.T) {
	t.Parallel()

	selectedBytes := []byte("ignored")
	selectedPath := writeTempFile(t, selectedBytes)
	openErr := errors.New("injected open failure")
	var opened []string
	loader := newLoaderWithOpenFile(func(path string) (io.ReadCloser, error) {
		opened = append(opened, path)
		return nil, openErr
	})

	_, err := loader.LoadIndex(selectedPath)

	assertIndexLoadFailure(t, err)
	if !errors.Is(err, openErr) {
		t.Fatalf("LoadIndex() error does not wrap injected open error: %v", err)
	}
	if !slices.Equal(opened, []string{selectedPath}) {
		t.Fatalf("opened paths = %v, want only caller-selected path %q", opened, selectedPath)
	}
	assertFileBytes(t, selectedPath, selectedBytes)
}

func TestLoadIndexRejectsReadableInvalidBytesAsIndexLoadFailure(t *testing.T) {
	t.Parallel()

	selectedBytes := []byte("not a SCIP protobuf index")
	selectedPath := writeTempFile(t, selectedBytes)
	var opened []string
	loader := newLoaderWithOpenFile(func(path string) (io.ReadCloser, error) {
		opened = append(opened, path)
		return os.Open(path)
	})

	_, err := loader.LoadIndex(selectedPath)

	assertIndexLoadFailure(t, err)
	if !slices.Equal(opened, []string{selectedPath}) {
		t.Fatalf("opened paths = %v, want only caller-selected path %q", opened, selectedPath)
	}
	assertFileBytes(t, selectedPath, selectedBytes)
}

func TestLoadIndexParsesValidSCIPBytesWithSelectedPathPreserved(t *testing.T) {
	t.Parallel()

	wantIndex := &scip.Index{
		Metadata: &scip.Metadata{
			ToolInfo: &scip.ToolInfo{
				Name:    "scip-search-test",
				Version: "v1",
			},
			ProjectRoot:          "file:///workspace",
			TextDocumentEncoding: scip.TextEncoding_UTF8,
		},
		Documents: []*scip.Document{
			{
				RelativePath: "cmd/main.go",
			},
		},
	}
	indexBytes, err := proto.Marshal(wantIndex)
	if err != nil {
		t.Fatalf("proto.Marshal(valid SCIP index) error = %v", err)
	}
	selectedPath := writeTempFile(t, indexBytes)
	var opened []string
	loader := newLoaderWithOpenFile(func(path string) (io.ReadCloser, error) {
		opened = append(opened, path)
		return os.Open(path)
	})

	loaded, err := loader.LoadIndex(selectedPath)

	if err != nil {
		t.Fatalf("LoadIndex(valid SCIP) error = %v", err)
	}
	if loaded.Path != selectedPath {
		t.Fatalf("loaded path = %q, want caller-selected path %q", loaded.Path, selectedPath)
	}
	if loaded.Fingerprint != fingerprintForTest(indexBytes) {
		t.Fatalf("fingerprint = %q, want stable hash of selected index bytes", loaded.Fingerprint)
	}
	if loaded.Index == nil {
		t.Fatal("loaded index is nil, want official SCIP index data")
	}
	if got := loaded.Index.GetMetadata().GetToolInfo().GetName(); got != "scip-search-test" {
		t.Fatalf("loaded index tool name = %q, want %q", got, "scip-search-test")
	}
	if got := loaded.Index.GetDocuments()[0].GetRelativePath(); got != "cmd/main.go" {
		t.Fatalf("loaded document path = %q, want %q", got, "cmd/main.go")
	}
	if !slices.Equal(opened, []string{selectedPath}) {
		t.Fatalf("opened paths = %v, want only caller-selected path %q", opened, selectedPath)
	}
	assertFileBytes(t, selectedPath, indexBytes)
}

func fingerprintForTest(contents []byte) string {
	sum := sha256.Sum256(contents)
	return fmt.Sprintf("sha256:%x", sum)
}

func assertIndexLoadFailure(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fatal("LoadIndex() error = nil, want index-loading failure")
	}

	var failure runtimecontract.Failure
	if !errors.As(err, &failure) {
		t.Fatalf("LoadIndex() error = %T %[1]v, want runtime failure class", err)
	}
	if failure.Status() != runtimecontract.StatusIndexLoad {
		t.Fatalf("failure status = %d, want %d", failure.Status(), runtimecontract.StatusIndexLoad)
	}
}

func writeTempFile(t *testing.T, contents []byte) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "selected.scip")
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v", path, err)
	}

	return path
}

func assertFileBytes(t *testing.T, path string, want []byte) {
	t.Helper()

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", path, err)
	}
	if !slices.Equal(got, want) {
		t.Fatalf("file bytes after LoadIndex() = %v, want unchanged %v", got, want)
	}
}
