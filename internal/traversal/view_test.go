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
		metadata.ToolVersion != "v1" ||
		metadata.ProjectRoot != "file:///workspace" ||
		metadata.TextDocumentEncoding != scip.TextEncoding_UTF8 {
		t.Fatalf("metadata = %+v, want tool, project root, and encoding preserved", metadata)
	}
	if metadata.ProtocolVersion != scip.ProtocolVersion_UnspecifiedProtocolVersion {
		t.Fatalf("protocol version = %v, want preserved SCIP protocol version", metadata.ProtocolVersion)
	}
	if !slices.Equal(metadata.ToolArguments, []string{"scip-go -o index.scip"}) {
		t.Fatalf("tool arguments = %v, want preserved metadata arguments", metadata.ToolArguments)
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
		definition.DocumentLanguage != "go" ||
		definition.PositionEncoding != scip.PositionEncoding_UTF8CodeUnitOffsetFromLineStart {
		t.Fatalf("definition occurrence context = %+v, want containing document path, language, and encoding", definition)
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

func TestNewViewPreservesHoverMetadataWithoutMergingDocumentation(t *testing.T) {
	t.Parallel()

	localSymbol := "scip-go gomod example.com/project . cmd/Documented()."
	externalSymbol := "scip-go gomod example.com/dependency v1.2.3 dep/External#"
	undocumentedSymbol := "scip-go gomod example.com/project . cmd/Undocumented()."
	view := NewView(runtimecontract.LoadedIndex{
		Index: &scip.Index{
			Documents: []*scip.Document{
				{
					RelativePath: "cmd/main.go",
					Language:     "go",
					Symbols: []*scip.SymbolInformation{
						{
							Symbol:          localSymbol,
							Kind:            scip.SymbolInformation_Method,
							DisplayName:     "Documented",
							Documentation:   []string{"**symbol** docs", "more symbol docs"},
							EnclosingSymbol: "scip-go gomod example.com/project . cmd/",
							SignatureDocumentation: &scip.Document{
								Language: "go",
								Text:     "func Documented() string",
							},
						},
						{
							Symbol:      undocumentedSymbol,
							Kind:        scip.SymbolInformation_Function,
							DisplayName: "Undocumented",
						},
					},
					Occurrences: []*scip.Occurrence{
						{
							Range:                 []int32{4, 2, 12},
							Symbol:                localSymbol,
							OverrideDocumentation: []string{"range-specific docs"},
						},
						{
							Range:  []int32{9, 3, 15},
							Symbol: localSymbol,
						},
					},
				},
			},
			ExternalSymbols: []*scip.SymbolInformation{
				{
					Symbol:          externalSymbol,
					Kind:            scip.SymbolInformation_Interface,
					DisplayName:     "External",
					Documentation:   []string{"external docs"},
					EnclosingSymbol: "scip-go gomod example.com/dependency v1.2.3 dep/",
					SignatureDocumentation: &scip.Document{
						Language: "go",
						Text:     "type External interface{}",
					},
				},
			},
		},
	})

	documents := view.Documents()
	if len(documents) != 1 {
		t.Fatalf("document count = %d, want 1", len(documents))
	}
	documentSymbols := documents[0].Symbols
	if len(documentSymbols) != 2 {
		t.Fatalf("document symbol count = %d, want documented and undocumented symbols", len(documentSymbols))
	}
	documented := documentSymbols[0]
	if documented.Symbol != localSymbol ||
		documented.Source != SymbolSourceDocument ||
		documented.DocumentPath != "cmd/main.go" ||
		documented.Kind != scip.SymbolInformation_Method ||
		documented.DisplayName != "Documented" {
		t.Fatalf("document symbol metadata = %+v, want full SCIP symbol, source, kind, display name, and path", documented)
	}
	if !slices.Equal(documented.Documentation, []string{"**symbol** docs", "more symbol docs"}) {
		t.Fatalf("document symbol docs = %v, want exact SCIP documentation entries", documented.Documentation)
	}
	if documented.SignatureDocumentation == nil ||
		documented.SignatureDocumentation.GetLanguage() != "go" ||
		documented.SignatureDocumentation.GetText() != "func Documented() string" {
		t.Fatalf("document signature documentation = %+v, want structured SCIP document", documented.SignatureDocumentation)
	}
	if documented.EnclosingSymbol != "scip-go gomod example.com/project . cmd/" {
		t.Fatalf("document enclosing symbol = %q, want exact SCIP enclosing symbol", documented.EnclosingSymbol)
	}

	undocumented := documentSymbols[1]
	if len(undocumented.Documentation) != 0 ||
		undocumented.SignatureDocumentation != nil ||
		undocumented.EnclosingSymbol != "" {
		t.Fatalf("undocumented symbol hover metadata = %+v, want absent docs preserved as absent", undocumented)
	}

	externalSymbols := view.ExternalSymbols()
	if len(externalSymbols) != 1 {
		t.Fatalf("external symbol count = %d, want 1", len(externalSymbols))
	}
	external := externalSymbols[0]
	if external.Symbol != externalSymbol ||
		external.Source != SymbolSourceExternal ||
		external.DocumentPath != "" ||
		external.Kind != scip.SymbolInformation_Interface ||
		external.DisplayName != "External" {
		t.Fatalf("external symbol metadata = %+v, want full SCIP symbol, source, kind, display name, and no document path", external)
	}
	if !slices.Equal(external.Documentation, []string{"external docs"}) {
		t.Fatalf("external docs = %v, want exact SCIP documentation entries", external.Documentation)
	}
	if external.SignatureDocumentation == nil ||
		external.SignatureDocumentation.GetLanguage() != "go" ||
		external.SignatureDocumentation.GetText() != "type External interface{}" {
		t.Fatalf("external signature documentation = %+v, want structured SCIP document", external.SignatureDocumentation)
	}
	if external.EnclosingSymbol != "scip-go gomod example.com/dependency v1.2.3 dep/" {
		t.Fatalf("external enclosing symbol = %q, want exact SCIP enclosing symbol", external.EnclosingSymbol)
	}

	occurrences := documents[0].Occurrences
	if len(occurrences) != 2 {
		t.Fatalf("occurrence count = %d, want present and absent override documentation cases", len(occurrences))
	}
	if !slices.Equal(occurrences[0].OverrideDocumentation, []string{"range-specific docs"}) {
		t.Fatalf("override docs = %v, want occurrence-level documentation preserved", occurrences[0].OverrideDocumentation)
	}
	if len(occurrences[1].OverrideDocumentation) != 0 {
		t.Fatalf("absent override docs = %v, want no symbol docs copied onto occurrence", occurrences[1].OverrideDocumentation)
	}
	if !slices.Equal(documented.Documentation, []string{"**symbol** docs", "more symbol docs"}) {
		t.Fatalf("symbol docs after occurrence inspection = %v, want symbol-level docs distinguishable", documented.Documentation)
	}
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

func TestOccurrencesForSymbolReturnsExactCrossDocumentMatchesWithMetadata(t *testing.T) {
	t.Parallel()

	target := "scip-go gomod example.com/project . shared/Target()."
	other := "scip-go gomod example.com/project . shared/TargetExtra()."
	view := NewView(runtimecontract.LoadedIndex{
		Index: &scip.Index{
			Documents: []*scip.Document{
				{
					RelativePath:     "pkg/first.go",
					Language:         "go",
					PositionEncoding: scip.PositionEncoding_UTF8CodeUnitOffsetFromLineStart,
					Occurrences: []*scip.Occurrence{
						{
							Range:                 []int32{2, 4, 10},
							EnclosingRange:        []int32{1, 0, 5, 1},
							Symbol:                target,
							SymbolRoles:           int32(scip.SymbolRole_Definition | scip.SymbolRole_ReadAccess),
							OverrideDocumentation: []string{"first occurrence docs"},
						},
						{
							Range:       []int32{8, 1, 9},
							Symbol:      other,
							SymbolRoles: int32(scip.SymbolRole_ReadAccess),
						},
					},
				},
				{
					RelativePath:     "pkg/second.go",
					Language:         "go",
					PositionEncoding: scip.PositionEncoding_UTF16CodeUnitOffsetFromLineStart,
					Occurrences: []*scip.Occurrence{
						{
							Range:                 []int32{12, 3, 14, 8},
							Symbol:                target,
							SymbolRoles:           int32(scip.SymbolRole_WriteAccess | scip.SymbolRole_Generated | scip.SymbolRole_Test),
							OverrideDocumentation: []string{"second occurrence docs"},
						},
					},
				},
			},
		},
	})

	matches := view.OccurrencesForSymbol(target)
	if len(matches) != 2 {
		t.Fatalf("target occurrence count = %d, want both cross-document matches", len(matches))
	}

	first := matches[0]
	if first.DocumentPath != "pkg/first.go" ||
		first.DocumentLanguage != "go" ||
		first.PositionEncoding != scip.PositionEncoding_UTF8CodeUnitOffsetFromLineStart ||
		first.Symbol != target {
		t.Fatalf("first occurrence context = %+v, want exact containing document and symbol", first)
	}
	if !slices.Equal(first.Range, []int32{2, 4, 10}) {
		t.Fatalf("first occurrence range = %v, want exact SCIP range", first.Range)
	}
	if !first.HasEnclosingRange || !slices.Equal(first.EnclosingRange, []int32{1, 0, 5, 1}) {
		t.Fatalf("first enclosing range = present:%v value:%v, want exact SCIP enclosing range", first.HasEnclosingRange, first.EnclosingRange)
	}
	wantFirstRoles := int32(scip.SymbolRole_Definition | scip.SymbolRole_ReadAccess)
	if first.SymbolRoles != wantFirstRoles {
		t.Fatalf("first symbol roles = %d, want raw bitset %d", first.SymbolRoles, wantFirstRoles)
	}
	if !slices.Equal(first.OverrideDocumentation, []string{"first occurrence docs"}) {
		t.Fatalf("first override docs = %v, want exact occurrence docs", first.OverrideDocumentation)
	}

	second := matches[1]
	if second.DocumentPath != "pkg/second.go" ||
		second.DocumentLanguage != "go" ||
		second.PositionEncoding != scip.PositionEncoding_UTF16CodeUnitOffsetFromLineStart ||
		second.Symbol != target {
		t.Fatalf("second occurrence context = %+v, want exact containing document and symbol", second)
	}
	if !slices.Equal(second.Range, []int32{12, 3, 14, 8}) {
		t.Fatalf("second occurrence range = %v, want exact SCIP range", second.Range)
	}
	if second.HasEnclosingRange || second.EnclosingRange != nil {
		t.Fatalf("second enclosing range = present:%v value:%v, want absent", second.HasEnclosingRange, second.EnclosingRange)
	}
	wantSecondRoles := int32(scip.SymbolRole_WriteAccess | scip.SymbolRole_Generated | scip.SymbolRole_Test)
	if second.SymbolRoles != wantSecondRoles {
		t.Fatalf("second symbol roles = %d, want raw bitset %d", second.SymbolRoles, wantSecondRoles)
	}
	if !slices.Equal(second.OverrideDocumentation, []string{"second occurrence docs"}) {
		t.Fatalf("second override docs = %v, want exact occurrence docs", second.OverrideDocumentation)
	}

	for _, occurrence := range matches {
		if occurrence.Symbol == other {
			t.Fatalf("target lookup returned other symbol occurrence: %+v", occurrence)
		}
	}
	if got := view.OccurrencesForSymbol("Target"); len(got) != 0 {
		t.Fatalf("partial symbol lookup = %+v, want empty exact-match result", got)
	}
	if got := view.OccurrencesForSymbol("scip-go gomod example.com/project . shared/Missing()."); len(got) != 0 {
		t.Fatalf("absent symbol lookup = %+v, want empty traversal result", got)
	}

	matches[0].Range[0] = 99
	matches[0].OverrideDocumentation[0] = "mutated"
	again := view.OccurrencesForSymbol(target)
	if again[0].Range[0] != 2 || again[0].OverrideDocumentation[0] != "first occurrence docs" {
		t.Fatalf("mutating lookup result changed stored occurrence: %+v", again[0])
	}
}

func TestNewViewPreservesOccurrenceLocationFactsWithoutSourceText(t *testing.T) {
	t.Parallel()

	sameLineSymbol := "scip-go gomod example.com/project . missing/SameLine()."
	multiPositionSymbol := "scip-go gomod example.com/project . empty/MultiPosition()."
	view := NewView(runtimecontract.LoadedIndex{
		Path: filepath.Join(t.TempDir(), "index.scip"),
		Index: &scip.Index{
			Documents: []*scip.Document{
				{
					RelativePath:     "missing/source.go",
					Language:         "go",
					PositionEncoding: scip.PositionEncoding_UTF8CodeUnitOffsetFromLineStart,
					Occurrences: []*scip.Occurrence{
						{
							Range:       []int32{11, 2, 9},
							Symbol:      sameLineSymbol,
							SymbolRoles: int32(scip.SymbolRole_ReadAccess),
						},
					},
				},
				{
					RelativePath:     "empty/source.go",
					Language:         "go",
					Text:             "",
					PositionEncoding: scip.PositionEncoding_UTF16CodeUnitOffsetFromLineStart,
					Occurrences: []*scip.Occurrence{
						{
							Range:          []int32{21, 1, 23, 4},
							EnclosingRange: []int32{20, 0, 24, 0},
							Symbol:         multiPositionSymbol,
							SymbolRoles:    int32(scip.SymbolRole_Definition | scip.SymbolRole_WriteAccess),
						},
					},
				},
			},
		},
	})

	documents := view.Documents()
	if len(documents) != 2 {
		t.Fatalf("document count = %d, want 2", len(documents))
	}
	if len(documents[0].Occurrences) != 1 {
		t.Fatalf("missing-text occurrence count = %d, want 1", len(documents[0].Occurrences))
	}
	if len(documents[1].Occurrences) != 1 {
		t.Fatalf("empty-text occurrence count = %d, want 1", len(documents[1].Occurrences))
	}

	missingTextOccurrence := documents[0].Occurrences[0]
	if missingTextOccurrence.DocumentPath != "missing/source.go" {
		t.Fatalf("missing-text occurrence document path = %q, want SCIP relative path", missingTextOccurrence.DocumentPath)
	}
	if missingTextOccurrence.PositionEncoding != scip.PositionEncoding_UTF8CodeUnitOffsetFromLineStart {
		t.Fatalf("missing-text occurrence position encoding = %v, want SCIP document encoding", missingTextOccurrence.PositionEncoding)
	}
	if !slices.Equal(missingTextOccurrence.Range, []int32{11, 2, 9}) {
		t.Fatalf("missing-text occurrence range = %v, want exact 3-integer SCIP range", missingTextOccurrence.Range)
	}
	if missingTextOccurrence.HasEnclosingRange || missingTextOccurrence.EnclosingRange != nil {
		t.Fatalf("missing-text enclosing range = present:%v value:%v, want absent", missingTextOccurrence.HasEnclosingRange, missingTextOccurrence.EnclosingRange)
	}
	if missingTextOccurrence.SymbolRoles != int32(scip.SymbolRole_ReadAccess) {
		t.Fatalf("missing-text occurrence roles = %d, want raw read-access bitset", missingTextOccurrence.SymbolRoles)
	}

	emptyTextOccurrence := documents[1].Occurrences[0]
	if emptyTextOccurrence.DocumentPath != "empty/source.go" {
		t.Fatalf("empty-text occurrence document path = %q, want SCIP relative path", emptyTextOccurrence.DocumentPath)
	}
	if emptyTextOccurrence.PositionEncoding != scip.PositionEncoding_UTF16CodeUnitOffsetFromLineStart {
		t.Fatalf("empty-text occurrence position encoding = %v, want SCIP document encoding", emptyTextOccurrence.PositionEncoding)
	}
	if !slices.Equal(emptyTextOccurrence.Range, []int32{21, 1, 23, 4}) {
		t.Fatalf("empty-text occurrence range = %v, want exact 4-integer SCIP range", emptyTextOccurrence.Range)
	}
	if !emptyTextOccurrence.HasEnclosingRange || !slices.Equal(emptyTextOccurrence.EnclosingRange, []int32{20, 0, 24, 0}) {
		t.Fatalf("empty-text enclosing range = present:%v value:%v, want exact SCIP enclosing range", emptyTextOccurrence.HasEnclosingRange, emptyTextOccurrence.EnclosingRange)
	}
	wantRoles := int32(scip.SymbolRole_Definition | scip.SymbolRole_WriteAccess)
	if emptyTextOccurrence.SymbolRoles != wantRoles {
		t.Fatalf("empty-text occurrence roles = %d, want raw bitset %d", emptyTextOccurrence.SymbolRoles, wantRoles)
	}

	if got := collectOccurrenceSymbols(view.Occurrences()); !slices.Equal(got, []string{sameLineSymbol, multiPositionSymbol}) {
		t.Fatalf("view occurrence symbols = %v, want SCIP occurrence symbols without filtering", got)
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

func TestNewViewKeepsEmptyDocumentCollectionsEnumerable(t *testing.T) {
	t.Parallel()

	view := NewView(runtimecontract.LoadedIndex{
		Index: &scip.Index{
			Documents: []*scip.Document{
				{
					RelativePath:     "pkg/empty.go",
					Language:         "go",
					PositionEncoding: scip.PositionEncoding_UTF8CodeUnitOffsetFromLineStart,
				},
			},
		},
	})

	documents := view.Documents()
	if len(documents) != 1 {
		t.Fatalf("document count = %d, want 1", len(documents))
	}
	if documents[0].Symbols == nil {
		t.Fatal("empty document symbols = nil, want enumerable empty collection")
	}
	if len(documents[0].Symbols) != 0 {
		t.Fatalf("empty document symbols = %v, want empty", documents[0].Symbols)
	}
	if documents[0].Occurrences == nil {
		t.Fatal("empty document occurrences = nil, want enumerable empty collection")
	}
	if len(documents[0].Occurrences) != 0 {
		t.Fatalf("empty document occurrences = %v, want empty", documents[0].Occurrences)
	}
}

func TestNewViewUsesOnlyLoadedIndexData(t *testing.T) {
	t.Parallel()

	view := NewView(runtimecontract.LoadedIndex{
		Path: "/path/that/traversal/must/not/open/index.scip",
		Index: &scip.Index{
			Documents: []*scip.Document{
				{
					RelativePath:     "missing/source/file.go",
					Language:         "go",
					PositionEncoding: scip.PositionEncoding_UTF8CodeUnitOffsetFromLineStart,
					Occurrences: []*scip.Occurrence{
						{
							Range:       []int32{1, 2, 3},
							Symbol:      "scip-go gomod example.com/project . missing/Symbol.",
							SymbolRoles: int32(scip.SymbolRole_ReadAccess),
						},
					},
				},
			},
		},
	})

	documents := view.Documents()
	if len(documents) != 1 {
		t.Fatalf("document count = %d, want traversal over loaded data without opening paths", len(documents))
	}
	if documents[0].RelativePath != "missing/source/file.go" {
		t.Fatalf("document path = %q, want loaded SCIP relative path", documents[0].RelativePath)
	}
	if len(documents[0].Occurrences) != 1 {
		t.Fatalf("occurrence count = %d, want loaded occurrence without reading source files", len(documents[0].Occurrences))
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

func TestNewViewExposesRelationshipsOwnedByExactSymbol(t *testing.T) {
	t.Parallel()

	localOwner := "scip-go gomod example.com/project . cmd/Owner#"
	externalOwner := "scip-go gomod example.com/dependency v1.2.3 dep/ExternalOwner#"
	otherOwner := "scip-go gomod example.com/project . cmd/OtherOwner#"
	emptyOwner := "scip-go gomod example.com/project . cmd/EmptyOwner#"
	referenceTarget := "scip-go gomod example.com/project . cmd/ReferenceTarget#"
	implementationTarget := "scip-go gomod example.com/project . cmd/ImplementationTarget#"
	typeDefinitionTarget := "scip-go gomod example.com/dependency v1.2.3 dep/TypeTarget#"
	multiFlagTarget := "scip-go gomod example.com/dependency v1.2.3 dep/MultiFlagTarget#"
	otherTarget := "scip-go gomod example.com/project . cmd/OtherTarget#"

	view := NewView(runtimecontract.LoadedIndex{
		Index: &scip.Index{
			Documents: []*scip.Document{
				{
					RelativePath: "cmd/owners.go",
					Symbols: []*scip.SymbolInformation{
						{
							Symbol: localOwner,
							Relationships: []*scip.Relationship{
								{Symbol: referenceTarget, IsReference: true},
								{Symbol: implementationTarget, IsImplementation: true},
							},
						},
						{
							Symbol: emptyOwner,
						},
						{
							Symbol: otherOwner,
							Relationships: []*scip.Relationship{
								{Symbol: otherTarget, IsDefinition: true},
							},
						},
					},
				},
			},
			ExternalSymbols: []*scip.SymbolInformation{
				{
					Symbol: externalOwner,
					Relationships: []*scip.Relationship{
						{Symbol: typeDefinitionTarget, IsTypeDefinition: true},
						{
							Symbol:           multiFlagTarget,
							IsReference:      true,
							IsImplementation: true,
							IsTypeDefinition: true,
							IsDefinition:     true,
						},
					},
				},
			},
		},
	})

	localRelationships := view.RelationshipsOwnedBy(localOwner)
	if len(localRelationships) != 2 {
		t.Fatalf("local owner relationship count = %d, want 2", len(localRelationships))
	}
	assertRelationshipFact(t, localRelationships, Relationship{
		SourceSymbol: localOwner,
		TargetSymbol: referenceTarget,
		IsReference:  true,
	})
	assertRelationshipFact(t, localRelationships, Relationship{
		SourceSymbol:     localOwner,
		TargetSymbol:     implementationTarget,
		IsImplementation: true,
	})
	assertNoRelationshipTarget(t, localRelationships, otherTarget)

	externalRelationships := view.RelationshipsOwnedBy(externalOwner)
	if len(externalRelationships) != 2 {
		t.Fatalf("external owner relationship count = %d, want 2", len(externalRelationships))
	}
	assertRelationshipFact(t, externalRelationships, Relationship{
		SourceSymbol:     externalOwner,
		TargetSymbol:     typeDefinitionTarget,
		IsTypeDefinition: true,
	})
	assertRelationshipFact(t, externalRelationships, Relationship{
		SourceSymbol:     externalOwner,
		TargetSymbol:     multiFlagTarget,
		IsReference:      true,
		IsImplementation: true,
		IsTypeDefinition: true,
		IsDefinition:     true,
	})

	if got := view.RelationshipsOwnedBy(emptyOwner); len(got) != 0 {
		t.Fatalf("empty owner relationships = %+v, want empty result", got)
	}
	if got := view.RelationshipsOwnedBy("scip-go gomod example.com/project . cmd/Missing#"); len(got) != 0 {
		t.Fatalf("absent owner relationships = %+v, want empty result", got)
	}
}

func TestNewViewExposesRelationshipsTargetingExactSymbol(t *testing.T) {
	t.Parallel()

	localOwner := "scip-go gomod example.com/project . cmd/LocalOwner#"
	externalOwner := "scip-go gomod example.com/dependency v1.2.3 dep/ExternalOwner#"
	otherOwner := "scip-go gomod example.com/project . cmd/OtherOwner#"
	target := "scip-go gomod example.com/project . cmd/SharedTarget#"
	otherTarget := "scip-go gomod example.com/project . cmd/OtherTarget#"
	missingTarget := "scip-go gomod example.com/project . cmd/MissingTarget#"

	view := NewView(runtimecontract.LoadedIndex{
		Index: &scip.Index{
			Documents: []*scip.Document{
				{
					RelativePath: "cmd/owners.go",
					Symbols: []*scip.SymbolInformation{
						{
							Symbol: localOwner,
							Relationships: []*scip.Relationship{
								{Symbol: target, IsReference: true},
								{Symbol: otherTarget, IsDefinition: true},
							},
						},
						{
							Symbol: otherOwner,
							Relationships: []*scip.Relationship{
								{Symbol: otherTarget, IsImplementation: true},
							},
						},
					},
				},
			},
			ExternalSymbols: []*scip.SymbolInformation{
				{
					Symbol: externalOwner,
					Relationships: []*scip.Relationship{
						{
							Symbol:           target,
							IsReference:      true,
							IsImplementation: true,
							IsTypeDefinition: true,
							IsDefinition:     true,
						},
					},
				},
			},
		},
	})

	targetRelationships := view.RelationshipsTargeting(target)
	if len(targetRelationships) != 2 {
		t.Fatalf("target relationship count = %d, want 2", len(targetRelationships))
	}
	assertRelationshipFact(t, targetRelationships, Relationship{
		SourceSymbol: localOwner,
		TargetSymbol: target,
		IsReference:  true,
	})
	assertRelationshipFact(t, targetRelationships, Relationship{
		SourceSymbol:     externalOwner,
		TargetSymbol:     target,
		IsReference:      true,
		IsImplementation: true,
		IsTypeDefinition: true,
		IsDefinition:     true,
	})
	assertNoRelationshipTarget(t, targetRelationships, otherTarget)
	assertNoRelationshipSource(t, targetRelationships, target)

	if got := view.RelationshipsTargeting(missingTarget); len(got) != 0 {
		t.Fatalf("absent target relationships = %+v, want empty result", got)
	}
	if got := view.RelationshipsOwnedBy(target); len(got) != 0 {
		t.Fatalf("target used as owner relationships = %+v, want no synthesized reverse owner relationships", got)
	}
}

func TestRelationshipOwnerLookupReturnsCopies(t *testing.T) {
	t.Parallel()

	owner := "scip-go gomod example.com/project . cmd/Owner#"
	target := "scip-go gomod example.com/project . cmd/Target#"
	view := NewView(runtimecontract.LoadedIndex{
		Index: &scip.Index{
			Documents: []*scip.Document{
				{
					Symbols: []*scip.SymbolInformation{
						{
							Symbol: owner,
							Relationships: []*scip.Relationship{
								{Symbol: target, IsReference: true},
							},
						},
					},
				},
			},
		},
	})

	relationships := view.RelationshipsOwnedBy(owner)
	if len(relationships) != 1 {
		t.Fatalf("relationship count = %d, want 1", len(relationships))
	}
	relationships[0].TargetSymbol = "mutated"
	relationships[0].IsReference = false

	again := view.RelationshipsOwnedBy(owner)
	assertRelationshipFact(t, again, Relationship{
		SourceSymbol: owner,
		TargetSymbol: target,
		IsReference:  true,
	})
}

func TestRelationshipTargetLookupReturnsCopies(t *testing.T) {
	t.Parallel()

	owner := "scip-go gomod example.com/project . cmd/Owner#"
	target := "scip-go gomod example.com/project . cmd/Target#"
	view := NewView(runtimecontract.LoadedIndex{
		Index: &scip.Index{
			Documents: []*scip.Document{
				{
					Symbols: []*scip.SymbolInformation{
						{
							Symbol: owner,
							Relationships: []*scip.Relationship{
								{Symbol: target, IsReference: true},
							},
						},
					},
				},
			},
		},
	})

	relationships := view.RelationshipsTargeting(target)
	if len(relationships) != 1 {
		t.Fatalf("relationship count = %d, want 1", len(relationships))
	}
	relationships[0].SourceSymbol = "mutated"
	relationships[0].IsReference = false

	again := view.RelationshipsTargeting(target)
	assertRelationshipFact(t, again, Relationship{
		SourceSymbol: owner,
		TargetSymbol: target,
		IsReference:  true,
	})
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

func assertRelationshipFact(t *testing.T, relationships []Relationship, want Relationship) {
	t.Helper()

	for _, relationship := range relationships {
		if relationship == want {
			return
		}
	}

	t.Fatalf("relationships = %+v, want relationship %+v", relationships, want)
}

func assertNoRelationshipTarget(t *testing.T, relationships []Relationship, target string) {
	t.Helper()

	for _, relationship := range relationships {
		if relationship.TargetSymbol == target {
			t.Fatalf("relationships = %+v, want no relationship targeting %q", relationships, target)
		}
	}
}

func assertNoRelationshipSource(t *testing.T, relationships []Relationship, source string) {
	t.Helper()

	for _, relationship := range relationships {
		if relationship.SourceSymbol == source {
			t.Fatalf("relationships = %+v, want no relationship sourced by %q", relationships, source)
		}
	}
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
