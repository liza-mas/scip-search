package aggregate

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/scip-code/scip/bindings/go/scip"
	"google.golang.org/protobuf/proto"

	"scip-search/internal/version"
)

const (
	targetSymbol   = "scip-typescript npm example 1.0.0 src/Target#"
	sharedExternal = "scip-typescript npm dependency 1.0.0 dep/External#"
)

func TestBuildRewritesPathsAndMetadataIntoAggregateProjectRoot(t *testing.T) {
	t.Parallel()

	repoRoot := filepath.Join(t.TempDir(), "repo")
	inputOne := writeIndex(t, singleDocumentIndex(
		"file://"+filepath.ToSlash(filepath.Join(repoRoot, "apps", "web", "src")),
		"App.tsx",
		targetSymbol,
		sharedExternal,
	))
	inputTwo := writeIndex(t, singleDocumentIndex(
		"file://"+filepath.ToSlash(filepath.Join(repoRoot, "services", "api")),
		"../api.py",
		"scip-typescript npm example 1.0.0 api/API#",
		sharedExternal,
	))

	index, result, err := Build(Options{
		ProjectRoot: repoRoot + string(os.PathSeparator),
		OutPath:     filepath.Join(t.TempDir(), "aggregate.scip"),
		Pairs: []Pair{
			{Root: "apps/web/src", IndexPath: inputOne},
			{Root: "services/api", IndexPath: inputTwo},
		},
	}, version.BuildIdentity{Release: "v1.2.3"})

	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if index.GetMetadata().GetProjectRoot() != "file://"+filepath.ToSlash(repoRoot) {
		t.Fatalf("project_root = %q, want aggregate repo root", index.GetMetadata().GetProjectRoot())
	}
	if index.GetMetadata().GetToolInfo().GetName() != ProducerName {
		t.Fatalf("tool name = %q, want aggregate producer", index.GetMetadata().GetToolInfo().GetName())
	}
	if got, want := documentPaths(index), []string{"apps/web/src/App.tsx", "services/api.py"}; !slices.Equal(got, want) {
		t.Fatalf("document paths = %v, want %v", got, want)
	}
	if result.DocumentCount != 2 || result.ExternalSymbolCount != 1 {
		t.Fatalf("result = %+v, want 2 documents and 1 deduped external", result)
	}
}

func TestBuildRewritesSingleIndexIntoAggregateProjectRoot(t *testing.T) {
	t.Parallel()

	repoRoot := filepath.Join(t.TempDir(), "repo")
	input := writeIndex(t, singleDocumentIndex(
		"file://"+filepath.ToSlash(filepath.Join(repoRoot, "apps", "web", "src")),
		"../vite.config.ts",
		targetSymbol,
		sharedExternal,
	))

	index, result, err := Build(Options{
		ProjectRoot: repoRoot,
		OutPath:     filepath.Join(t.TempDir(), "aggregate.scip"),
		Pairs: []Pair{
			{Root: "apps/web/src", IndexPath: input},
		},
	}, version.BuildIdentity{})

	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if index.GetMetadata().GetProjectRoot() != "file://"+filepath.ToSlash(repoRoot) {
		t.Fatalf("project_root = %q, want aggregate repo root", index.GetMetadata().GetProjectRoot())
	}
	if got, want := documentPaths(index), []string{"apps/web/vite.config.ts"}; !slices.Equal(got, want) {
		t.Fatalf("document paths = %v, want %v", got, want)
	}
	if result.DocumentCount != 1 || result.ExternalSymbolCount != 1 {
		t.Fatalf("result = %+v, want 1 document and 1 external", result)
	}
}

func TestBuildAllowsRepeatedLocalSymbolsInDifferentDocuments(t *testing.T) {
	t.Parallel()

	repoRoot := filepath.Join(t.TempDir(), "repo")
	first := writeIndex(t, singleDocumentIndex("", "supabase.ts", "local 4", ""))
	second := writeIndex(t, singleDocumentIndex("", "api.ts", "local 4", ""))

	index, result, err := Build(aggregateOptions(repoRoot, first, second, "apps/web/src/lib", "services/design-diagnosis/web/src/lib"), version.BuildIdentity{})

	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if got, want := documentPaths(index), []string{"apps/web/src/lib/supabase.ts", "services/design-diagnosis/web/src/lib/api.ts"}; !slices.Equal(got, want) {
		t.Fatalf("document paths = %v, want %v", got, want)
	}
	if result.DocumentCount != 2 {
		t.Fatalf("document count = %d, want 2", result.DocumentCount)
	}
}

func TestBuildAllowsRepeatedNonLocalDefinitionsWithinOneInputIndex(t *testing.T) {
	t.Parallel()

	repoRoot := filepath.Join(t.TempDir(), "repo")
	repeatedSymbol := "scip-go gomod github.com/example/project . `github.com/example/project/cmd`/"
	index := indexWithTool("scip-go", "cluster.go", repeatedSymbol)
	secondDocument := proto.Clone(index.Documents[0]).(*scip.Document)
	secondDocument.RelativePath = "create.go"
	index.Documents = append(index.Documents, secondDocument)
	input := writeIndex(t, index)

	aggregate, result, err := Build(Options{
		ProjectRoot: repoRoot,
		OutPath:     filepath.Join(t.TempDir(), "aggregate.scip"),
		Pairs: []Pair{
			{Root: "services/cli/cmd", IndexPath: input},
		},
	}, version.BuildIdentity{})

	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if got, want := documentPaths(aggregate), []string{"services/cli/cmd/cluster.go", "services/cli/cmd/create.go"}; !slices.Equal(got, want) {
		t.Fatalf("document paths = %v, want %v", got, want)
	}
	if result.DocumentCount != 2 {
		t.Fatalf("document count = %d, want 2", result.DocumentCount)
	}
}

func TestBuildDropsDocumentsThatEscapeAggregateProjectRoot(t *testing.T) {
	t.Parallel()

	repoRoot := filepath.Join(t.TempDir(), "repo")
	index := indexWithTool("scip-go", "../external/util.go", "scip-go gomod example.com/project . external/Keep#")
	escapedDocument := proto.Clone(index.Documents[0]).(*scip.Document)
	escapedDocument.RelativePath = "../../../../../.cache/go-build/generated.go"
	escapedDocument.Symbols[0].Symbol = "scip-go gomod example.com/project . cache/Drop#"
	escapedDocument.Occurrences[0].Symbol = escapedDocument.Symbols[0].Symbol
	index.Documents = append(index.Documents, escapedDocument)
	input := writeIndex(t, index)

	aggregate, result, err := Build(Options{
		ProjectRoot: repoRoot,
		OutPath:     filepath.Join(t.TempDir(), "aggregate.scip"),
		Pairs: []Pair{
			{Root: "apps/web/src", IndexPath: input},
		},
	}, version.BuildIdentity{})

	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if got, want := documentPaths(aggregate), []string{"apps/web/external/util.go"}; !slices.Equal(got, want) {
		t.Fatalf("document paths = %v, want %v", got, want)
	}
	if result.DocumentCount != 1 || len(result.DroppedDocuments) != 1 {
		t.Fatalf("result = %+v, want 1 kept document and 1 dropped document", result)
	}
	if got, want := result.DroppedDocuments, []DroppedDocument{{
		InputIndex:   0,
		RelativePath: "../../../../../.cache/go-build/generated.go",
		Reason:       `document path "../../../../../.cache/go-build/generated.go" escapes aggregate project root`,
	}}; !slices.Equal(got, want) {
		t.Fatalf("dropped documents = %+v, want %+v", got, want)
	}
}

func TestBuildDeduplicatesNonLocalExternalSymbolsByIdentity(t *testing.T) {
	t.Parallel()

	repoRoot := filepath.Join(t.TempDir(), "repo")
	first := writeIndex(t, indexWithExternalDocumentation("AppFoo", "```python\nclass range(__start: SupportsIndex)\n```"))
	second := writeIndex(t, indexWithExternalDocumentation("DiagnosisFoo", "```python\nclass range(__stop: SupportsIndex)\n```"))

	index, result, err := Build(aggregateOptions(repoRoot, first, second, "apps/api", "services/design-diagnosis"), version.BuildIdentity{})

	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if result.ExternalSymbolCount != 1 {
		t.Fatalf("external symbol count = %d, want 1", result.ExternalSymbolCount)
	}
	if got := index.GetExternalSymbols()[0].GetDocumentation(); !slices.Equal(got, []string{"```python\nclass range(__start: SupportsIndex)\n```"}) {
		t.Fatalf("external documentation = %q, want first input record", got)
	}
}

func TestBuildPreservesRepeatedLocalExternalSymbols(t *testing.T) {
	t.Parallel()

	repoRoot := filepath.Join(t.TempDir(), "repo")
	first := writeIndex(t, indexWithLocalExternal("FirstLocal"))
	second := writeIndex(t, indexWithLocalExternal("SecondLocal"))

	index, result, err := Build(aggregateOptions(repoRoot, first, second, "apps/api", "services/design-diagnosis"), version.BuildIdentity{})

	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if result.ExternalSymbolCount != 2 {
		t.Fatalf("external symbol count = %d, want 2", result.ExternalSymbolCount)
	}
	if got := externalDisplayNames(index); !slices.Equal(got, []string{"FirstLocal", "SecondLocal"}) {
		t.Fatalf("external display names = %v, want both local records preserved", got)
	}
}

func TestBuildRejectsInvalidAggregates(t *testing.T) {
	t.Parallel()

	repoRoot := filepath.Join(t.TempDir(), "repo")
	tests := []struct {
		name    string
		options func(t *testing.T) Options
		want    string
	}{
		{
			name: "duplicate document paths",
			options: func(t *testing.T) Options {
				first := writeIndex(t, singleDocumentIndex("", "main.ts", targetSymbol, ""))
				second := writeIndex(t, singleDocumentIndex("", "main.ts", "scip-typescript npm example 1.0.0 src/Other#", ""))
				return aggregateOptions(repoRoot, first, second, ".", ".")
			},
			want: "duplicate aggregate document path",
		},
		{
			name: "root mapping mismatch",
			options: func(t *testing.T) Options {
				first := writeIndex(t, singleDocumentIndex("file://"+filepath.ToSlash(filepath.Join(repoRoot, "apps", "web")), "main.ts", targetSymbol, ""))
				second := writeIndex(t, singleDocumentIndex("", "other.ts", "scip-typescript npm example 1.0.0 src/Other#", ""))
				return aggregateOptions(repoRoot, first, second, "apps/api", "services/api")
			},
			want: "root mapping mismatch",
		},
		{
			name: "symbol collision",
			options: func(t *testing.T) Options {
				first := writeIndex(t, singleDocumentIndex("", "main.ts", targetSymbol, ""))
				second := writeIndex(t, singleDocumentIndex("", "other.ts", targetSymbol, ""))
				return aggregateOptions(repoRoot, first, second, "apps/a", "apps/b")
			},
			want: "symbol collision",
		},
		{
			name: "document symbol collision without definition occurrence",
			options: func(t *testing.T) Options {
				firstIndex := singleDocumentIndex("", "main.ts", targetSymbol, "")
				firstIndex.Documents[0].Occurrences = nil
				secondIndex := singleDocumentIndex("", "other.ts", targetSymbol, "")
				secondIndex.Documents[0].Occurrences = nil
				first := writeIndex(t, firstIndex)
				second := writeIndex(t, secondIndex)
				return aggregateOptions(repoRoot, first, second, "apps/a", "apps/b")
			},
			want: "symbol collision",
		},
		{
			name: "mixed indexer families",
			options: func(t *testing.T) Options {
				first := writeIndex(t, indexWithTool("scip-python", "main.py", "scip-python pip example 1.0.0 main/Foo#"))
				second := writeIndex(t, indexWithTool("scip-typescript", "main.ts", "scip-typescript npm example 1.0.0 main/Foo#"))
				return aggregateOptions(repoRoot, first, second, "apps/a", "apps/b")
			},
			want: "mixed indexer families",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, _, err := Build(test.options(t), version.BuildIdentity{})

			if err == nil || !IsValidationError(err) || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Build() error = %v, want validation containing %q", err, test.want)
			}
		})
	}
}

func TestRunLeavesExistingOutputUnchangedOnValidationFailure(t *testing.T) {
	t.Parallel()

	repoRoot := filepath.Join(t.TempDir(), "repo")
	first := writeIndex(t, singleDocumentIndex("", "main.ts", targetSymbol, ""))
	second := writeIndex(t, singleDocumentIndex("", "main.ts", "scip-typescript npm example 1.0.0 src/Other#", ""))
	outPath := filepath.Join(t.TempDir(), "aggregate.scip")
	original := []byte("keep me")
	if err := os.WriteFile(outPath, original, 0o600); err != nil {
		t.Fatalf("write existing output: %v", err)
	}

	_, err := Run(aggregateOptions(repoRoot, first, second, ".", "."), version.BuildIdentity{})

	if err == nil || !IsValidationError(err) {
		t.Fatalf("Run() error = %v, want validation failure", err)
	}
	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !slices.Equal(got, original) {
		t.Fatalf("output bytes = %q, want unchanged %q", got, original)
	}
}

func aggregateOptions(repoRoot string, first string, second string, firstRoot string, secondRoot string) Options {
	return Options{
		ProjectRoot: repoRoot,
		OutPath:     filepath.Join(filepath.Dir(first), "aggregate.scip"),
		Pairs: []Pair{
			{Root: firstRoot, IndexPath: first},
			{Root: secondRoot, IndexPath: second},
		},
	}
}

func singleDocumentIndex(projectRoot string, relativePath string, symbol string, externalSymbol string) *scip.Index {
	index := indexWithTool("scip-typescript", relativePath, symbol)
	index.Metadata.ProjectRoot = projectRoot
	if externalSymbol != "" {
		index.ExternalSymbols = []*scip.SymbolInformation{{Symbol: externalSymbol, DisplayName: "External"}}
	}
	return index
}

func indexWithExternalDocumentation(displayName string, documentation string) *scip.Index {
	index := indexWithTool("scip-python", "main.py", "scip-python python example 1.0.0 main/"+displayName+"#")
	index.ExternalSymbols = []*scip.SymbolInformation{{
		Symbol:        "scip-python python python-stdlib 3.11 builtins/range#",
		Documentation: []string{documentation},
	}}
	return index
}

func indexWithLocalExternal(displayName string) *scip.Index {
	index := indexWithTool("scip-python", "main.py", "scip-python python example 1.0.0 main/"+displayName+"#")
	index.ExternalSymbols = []*scip.SymbolInformation{{
		Symbol:      "local 4",
		DisplayName: displayName,
	}}
	return index
}

func indexWithTool(toolName string, relativePath string, symbol string) *scip.Index {
	return &scip.Index{
		Metadata: &scip.Metadata{
			Version:              scip.ProtocolVersion_UnspecifiedProtocolVersion,
			TextDocumentEncoding: scip.TextEncoding_UTF8,
			ToolInfo:             &scip.ToolInfo{Name: toolName, Version: "test"},
		},
		Documents: []*scip.Document{
			{
				RelativePath: relativePath,
				Language:     "typescript",
				Symbols:      []*scip.SymbolInformation{{Symbol: symbol, DisplayName: "Target"}},
				Occurrences: []*scip.Occurrence{
					{
						Symbol:      symbol,
						Range:       []int32{1, 2, 3},
						SymbolRoles: int32(scip.SymbolRole_Definition),
					},
				},
			},
		},
	}
}

func writeIndex(t *testing.T, index *scip.Index) string {
	t.Helper()

	payload, err := proto.Marshal(index)
	if err != nil {
		t.Fatalf("marshal index: %v", err)
	}
	path := filepath.Join(t.TempDir(), "index.scip")
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatalf("write index: %v", err)
	}
	return path
}

func documentPaths(index *scip.Index) []string {
	paths := make([]string, 0, len(index.GetDocuments()))
	for _, document := range index.GetDocuments() {
		paths = append(paths, document.GetRelativePath())
	}
	return paths
}

func externalDisplayNames(index *scip.Index) []string {
	names := make([]string, 0, len(index.GetExternalSymbols()))
	for _, symbol := range index.GetExternalSymbols() {
		names = append(names, symbol.GetDisplayName())
	}
	return names
}
