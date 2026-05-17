package traversal

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/scip-code/scip/bindings/go/scip"
	"google.golang.org/protobuf/proto"

	runtimecontract "scip-search/internal/runtime"
	"scip-search/internal/scipindex"
)

func TestNewViewBuildsFromLoadedRuntimeIndexAndPreservesTraversalFacts(t *testing.T) {
	t.Parallel()

	index := traversalTestIndex()
	loaded := loadThroughRuntimeBoundary(t, index)

	view := NewView(loaded)

	metadata := view.Metadata()
	if metadata.ToolName != "scip-search-test" ||
		metadata.ProjectRoot != "file:///workspace" ||
		metadata.TextDocumentEncoding != scip.TextEncoding_UTF8 {
		t.Fatalf("metadata = %+v, want tool, project root, and encoding preserved", metadata)
	}

	documents := view.Documents()
	if len(documents) != 2 {
		t.Fatalf("document count = %d, want 2", len(documents))
	}

	first := documents[0]
	if first.RelativePath != "cmd/main.go" ||
		first.Language != "go" ||
		first.PositionEncoding != scip.PositionEncoding_UTF8CodeUnitOffsetFromLineStart {
		t.Fatalf("first document = %+v, want path, language, and position encoding preserved", first)
	}
	if len(first.Symbols) != 1 {
		t.Fatalf("first document symbol count = %d, want 1", len(first.Symbols))
	}
	assertSymbolFact(t, first.Symbols[0], SymbolSourceDocument, "cmd/main.go")

	if len(first.Occurrences) != 2 {
		t.Fatalf("first document occurrence count = %d, want 2", len(first.Occurrences))
	}
	definition := first.Occurrences[0]
	if definition.DocumentPath != "cmd/main.go" ||
		definition.PositionEncoding != scip.PositionEncoding_UTF8CodeUnitOffsetFromLineStart {
		t.Fatalf("definition occurrence context = %+v, want containing document path and encoding", definition)
	}
	if !slices.Equal(definition.Range, []int32{3, 4, 12}) {
		t.Fatalf("same-line range = %v, want exact SCIP range", definition.Range)
	}
	if !definition.HasEnclosingRange {
		t.Fatal("definition occurrence has enclosing range = false, want true")
	}
	if !slices.Equal(definition.EnclosingRange, []int32{2, 0, 5, 1}) {
		t.Fatalf("enclosing range = %v, want exact SCIP enclosing range", definition.EnclosingRange)
	}
	wantRoles := int32(scip.SymbolRole_Definition | scip.SymbolRole_ReadAccess)
	if definition.SymbolRoles != wantRoles {
		t.Fatalf("symbol roles = %d, want raw bitset %d", definition.SymbolRoles, wantRoles)
	}
	if !slices.Equal(definition.OverrideDocumentation, []string{"override docs for Alpha at this range"}) {
		t.Fatalf("override documentation = %v, want preserved occurrence docs", definition.OverrideDocumentation)
	}

	reference := first.Occurrences[1]
	if !slices.Equal(reference.Range, []int32{7, 1, 8, 9}) {
		t.Fatalf("multi-position range = %v, want exact SCIP range", reference.Range)
	}
	if reference.HasEnclosingRange {
		t.Fatalf("reference occurrence has enclosing range = true, want false")
	}
	if reference.EnclosingRange != nil {
		t.Fatalf("absent enclosing range = %v, want nil", reference.EnclosingRange)
	}
	if reference.SymbolRoles != int32(scip.SymbolRole_ReadAccess) {
		t.Fatalf("reference roles = %d, want read-access bitset", reference.SymbolRoles)
	}
	if len(reference.OverrideDocumentation) != 0 {
		t.Fatalf("reference override docs = %v, want empty", reference.OverrideDocumentation)
	}

	occurrences := view.Occurrences()
	if len(occurrences) != 2 {
		t.Fatalf("view occurrence inventory length = %d, want 2", len(occurrences))
	}
	if occurrences[0].DocumentPath != "cmd/main.go" || occurrences[1].DocumentPath != "cmd/main.go" {
		t.Fatalf("view occurrence document paths = %q, %q; want containing document path", occurrences[0].DocumentPath, occurrences[1].DocumentPath)
	}

	second := documents[1]
	if second.RelativePath != "pkg/empty.go" || len(second.Symbols) != 0 || len(second.Occurrences) != 0 {
		t.Fatalf("second document = %+v, want empty inventories preserved", second)
	}

	externalSymbols := view.ExternalSymbols()
	if len(externalSymbols) != 1 {
		t.Fatalf("external symbol count = %d, want 1", len(externalSymbols))
	}
	assertSymbolFact(t, externalSymbols[0], SymbolSourceExternal, "")
}

func TestNewViewExposesEverySymbolAndOccurrenceInventoryFact(t *testing.T) {
	t.Parallel()

	localAlpha := "scip-go gomod example.com/project . cmd/Alpha()."
	localBeta := "scip-go gomod example.com/project . pkg/Beta#"
	externalGamma := "scip-go gomod example.com/dependency v1.2.3 dep/Gamma#"
	externalDelta := "scip-go gomod example.com/dependency v1.2.3 dep/Delta()."
	view := NewView(runtimecontract.LoadedIndex{
		Index: &scip.Index{
			Documents: []*scip.Document{
				{
					RelativePath: "cmd/main.go",
					Language:     "go",
					Symbols: []*scip.SymbolInformation{
						symbolInformation(localAlpha, scip.SymbolInformation_Function),
						symbolInformation(localBeta, scip.SymbolInformation_Class),
					},
					Occurrences: []*scip.Occurrence{
						{Symbol: localAlpha, Range: []int32{1, 2, 9}},
						{Symbol: externalGamma, Range: []int32{3, 4, 10}},
					},
				},
				{
					RelativePath: "pkg/worker.go",
					Language:     "go",
					Occurrences: []*scip.Occurrence{
						{Symbol: localBeta, Range: []int32{5, 1, 6}},
					},
				},
			},
			ExternalSymbols: []*scip.SymbolInformation{
				symbolInformation(externalGamma, scip.SymbolInformation_Interface),
				symbolInformation(externalDelta, scip.SymbolInformation_Method),
			},
		},
	})

	documents := view.Documents()
	if got := collectSymbols(documents[0].Symbols); !slices.Equal(got, []string{localAlpha, localBeta}) {
		t.Fatalf("first document symbols = %v, want every document-level symbol", got)
	}
	for _, symbol := range documents[0].Symbols {
		if symbol.Source != SymbolSourceDocument || symbol.DocumentPath != "cmd/main.go" {
			t.Fatalf("document symbol source/path = %q/%q, want document/cmd/main.go", symbol.Source, symbol.DocumentPath)
		}
	}
	if len(documents[1].Symbols) != 0 {
		t.Fatalf("second document symbols = %v, want empty collection", documents[1].Symbols)
	}

	externalSymbols := view.ExternalSymbols()
	if got := collectSymbols(externalSymbols); !slices.Equal(got, []string{externalGamma, externalDelta}) {
		t.Fatalf("external symbols = %v, want every index-level external symbol", got)
	}
	for _, symbol := range externalSymbols {
		if symbol.Source != SymbolSourceExternal || symbol.DocumentPath != "" {
			t.Fatalf("external symbol source/path = %q/%q, want external/empty path", symbol.Source, symbol.DocumentPath)
		}
	}

	occurrences := view.Occurrences()
	if got := collectOccurrenceSymbols(occurrences); !slices.Equal(got, []string{localAlpha, externalGamma, localBeta}) {
		t.Fatalf("occurrence symbols = %v, want every referenced SCIP symbol", got)
	}
	if occurrences[0].DocumentPath != "cmd/main.go" ||
		occurrences[1].DocumentPath != "cmd/main.go" ||
		occurrences[2].DocumentPath != "pkg/worker.go" {
		t.Fatalf("occurrence document paths = %q, %q, %q; want containing documents", occurrences[0].DocumentPath, occurrences[1].DocumentPath, occurrences[2].DocumentPath)
	}
	if got := collectOccurrenceSymbols(view.Occurrences()); !slices.Equal(got, collectOccurrenceSymbols(occurrences)) {
		t.Fatalf("repeated occurrence enumeration = %v, want deterministic %v", got, collectOccurrenceSymbols(occurrences))
	}
}

func TestNewViewKeepsEmptyInventoriesAsData(t *testing.T) {
	t.Parallel()

	view := NewView(runtimecontract.LoadedIndex{
		Path: "not-opened-by-traversal.scip",
		Index: &scip.Index{
			Metadata: &scip.Metadata{},
		},
	})

	if len(view.Documents()) != 0 {
		t.Fatalf("empty document inventory = %v, want empty", view.Documents())
	}
	if len(view.ExternalSymbols()) != 0 {
		t.Fatalf("empty external symbol inventory = %v, want empty", view.ExternalSymbols())
	}
	if len(view.Occurrences()) != 0 {
		t.Fatalf("empty occurrence inventory = %v, want empty", view.Occurrences())
	}
}

func TestNewViewPreservesPresentEmptyEnclosingRangeWhenBindingExposesIt(t *testing.T) {
	t.Parallel()

	view := NewView(runtimecontract.LoadedIndex{
		Index: &scip.Index{
			Documents: []*scip.Document{
				{
					RelativePath: "generated.go",
					Occurrences: []*scip.Occurrence{
						{
							Range:          []int32{1, 2, 3},
							EnclosingRange: []int32{},
						},
					},
				},
			},
		},
	})

	occurrence := view.Documents()[0].Occurrences[0]
	if !occurrence.HasEnclosingRange {
		t.Fatal("present empty enclosing range reported absent")
	}
	if occurrence.EnclosingRange == nil {
		t.Fatal("present empty enclosing range = nil, want empty slice")
	}
	if len(occurrence.EnclosingRange) != 0 {
		t.Fatalf("present empty enclosing range = %v, want empty slice", occurrence.EnclosingRange)
	}
}

func TestNewViewEnumerationIsDeterministicAndReturnsCopies(t *testing.T) {
	t.Parallel()

	view := NewView(runtimecontract.LoadedIndex{Index: traversalTestIndex()})

	firstDocuments := view.Documents()
	secondDocuments := view.Documents()
	if !slices.EqualFunc(firstDocuments, secondDocuments, func(left Document, right Document) bool {
		return left.RelativePath == right.RelativePath &&
			len(left.Symbols) == len(right.Symbols) &&
			len(left.Occurrences) == len(right.Occurrences)
	}) {
		t.Fatalf("repeated document enumeration differs: first=%v second=%v", firstDocuments, secondDocuments)
	}

	firstDocuments[0].RelativePath = "mutated.go"
	firstDocuments[0].Occurrences[0].Range[0] = 99
	metadata := view.Metadata()
	metadata.ToolArguments[0] = "mutated"
	if got := view.Documents()[0].RelativePath; got != "cmd/main.go" {
		t.Fatalf("view document path after caller mutation = %q, want original", got)
	}
	if got := view.Documents()[0].Occurrences[0].Range[0]; got != 3 {
		t.Fatalf("view occurrence range after caller mutation = %d, want original", got)
	}
	if got := view.Metadata().ToolArguments[0]; got != "scip-go -o index.scip" {
		t.Fatalf("view metadata arguments after caller mutation = %q, want original", got)
	}
}

func assertSymbolFact(t *testing.T, symbol Symbol, source SymbolSource, documentPath string) {
	t.Helper()

	if symbol.Symbol != "scip-go gomod example.com/project . cmd/Alpha()." {
		t.Fatalf("symbol = %q, want full SCIP symbol", symbol.Symbol)
	}
	if symbol.Source != source || symbol.DocumentPath != documentPath ||
		symbol.Kind != scip.SymbolInformation_Function || symbol.DisplayName != "Alpha" {
		t.Fatalf("symbol metadata = %+v, want source, document path, kind, and display name", symbol)
	}
	if !slices.Equal(symbol.Documentation, []string{"Alpha docs"}) {
		t.Fatalf("documentation = %v, want preserved docs", symbol.Documentation)
	}
	if symbol.SignatureDocumentation == nil {
		t.Fatal("signature documentation = nil, want SCIP document")
	}
	if symbol.SignatureDocumentation.GetLanguage() != "go" {
		t.Fatalf("signature language = %q, want go", symbol.SignatureDocumentation.GetLanguage())
	}
	if symbol.SignatureDocumentation.GetText() != "func Alpha() string" {
		t.Fatalf("signature text = %q, want preserved signature text", symbol.SignatureDocumentation.GetText())
	}
	if symbol.EnclosingSymbol != "scip-go gomod example.com/project . cmd/" {
		t.Fatalf("enclosing symbol = %q, want preserved metadata", symbol.EnclosingSymbol)
	}
}

func symbolInformation(symbol string, kind scip.SymbolInformation_Kind) *scip.SymbolInformation {
	return &scip.SymbolInformation{
		Symbol:      symbol,
		Kind:        kind,
		DisplayName: symbol,
	}
}

func collectSymbols(symbols []Symbol) []string {
	collected := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		collected = append(collected, symbol.Symbol)
	}
	return collected
}

func collectOccurrenceSymbols(occurrences []Occurrence) []string {
	collected := make([]string, 0, len(occurrences))
	for _, occurrence := range occurrences {
		collected = append(collected, occurrence.Symbol)
	}
	return collected
}

func traversalTestIndex() *scip.Index {
	symbol := "scip-go gomod example.com/project . cmd/Alpha()."
	symbolInfo := &scip.SymbolInformation{
		Symbol:          symbol,
		Kind:            scip.SymbolInformation_Function,
		DisplayName:     "Alpha",
		Documentation:   []string{"Alpha docs"},
		EnclosingSymbol: "scip-go gomod example.com/project . cmd/",
		SignatureDocumentation: &scip.Document{
			Language: "go",
			Text:     "func Alpha() string",
		},
	}

	return &scip.Index{
		Metadata: &scip.Metadata{
			ToolInfo: &scip.ToolInfo{
				Name:    "scip-search-test",
				Version: "v1",
				Arguments: []string{
					"scip-go -o index.scip",
				},
			},
			ProjectRoot:          "file:///workspace",
			TextDocumentEncoding: scip.TextEncoding_UTF8,
		},
		Documents: []*scip.Document{
			{
				RelativePath:     "cmd/main.go",
				Language:         "go",
				PositionEncoding: scip.PositionEncoding_UTF8CodeUnitOffsetFromLineStart,
				Symbols:          []*scip.SymbolInformation{symbolInfo},
				Occurrences: []*scip.Occurrence{
					{
						Range:                 []int32{3, 4, 12},
						EnclosingRange:        []int32{2, 0, 5, 1},
						Symbol:                symbol,
						SymbolRoles:           int32(scip.SymbolRole_Definition | scip.SymbolRole_ReadAccess),
						OverrideDocumentation: []string{"override docs for Alpha at this range"},
					},
					{
						Range:       []int32{7, 1, 8, 9},
						Symbol:      symbol,
						SymbolRoles: int32(scip.SymbolRole_ReadAccess),
					},
				},
			},
			{
				RelativePath:     "pkg/empty.go",
				Language:         "go",
				PositionEncoding: scip.PositionEncoding_UTF16CodeUnitOffsetFromLineStart,
			},
		},
		ExternalSymbols: []*scip.SymbolInformation{symbolInfo},
	}
}

func loadThroughRuntimeBoundary(t *testing.T, index *scip.Index) runtimecontract.LoadedIndex {
	t.Helper()

	indexBytes, err := proto.Marshal(index)
	if err != nil {
		t.Fatalf("proto.Marshal(test index) error = %v", err)
	}

	indexPath := filepath.Join(t.TempDir(), "index.scip")
	if err := os.WriteFile(indexPath, indexBytes, 0o600); err != nil {
		t.Fatalf("os.WriteFile(index) error = %v", err)
	}

	loaded, err := scipindex.NewLoader().LoadIndex(indexPath)
	if err != nil {
		t.Fatalf("LoadIndex(test index) error = %v", err)
	}

	return loaded
}
