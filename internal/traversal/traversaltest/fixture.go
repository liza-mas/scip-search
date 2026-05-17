package traversaltest

import (
	_ "embed"
	"encoding/base64"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	runtimecontract "scip-search/internal/runtime"
	"scip-search/internal/scipindex"
	"scip-search/internal/traversal"
)

const (
	AlphaSymbol            = "scip-go gomod example.com/fixture . cmd/Alpha()."
	BetaSymbol             = "scip-go gomod example.com/fixture . pkg/Beta#"
	ExternalSymbol         = "scip-go gomod example.com/dependency v1.2.3 dep/External#"
	ImplSymbol             = "scip-go gomod example.com/fixture . pkg/Impl#"
	DefinitionSymbol       = "scip-go gomod example.com/fixture . pkg/Definition#"
	TypeDefinitionSymbol   = "scip-go gomod example.com/fixture . pkg/TypeDefinition#"
	MultiFlagTargetSymbol  = "scip-go gomod example.com/dependency v1.2.3 dep/MultiFlag#"
	sharedFixtureIndexFile = "shared-fixture.scip"
)

//go:embed testdata/shared-fixture.scip.b64
var sharedFixtureBase64 string

type Fixture struct {
	LoadedIndex runtimecontract.LoadedIndex
	View        traversal.View
}

func LoadSharedFixture(t testing.TB) Fixture {
	t.Helper()

	payload, err := base64.StdEncoding.DecodeString(strings.TrimSpace(sharedFixtureBase64))
	if err != nil {
		t.Fatalf("decode shared traversal fixture: %v", err)
	}

	indexPath := filepath.Join(t.TempDir(), sharedFixtureIndexFile)
	if err := os.WriteFile(indexPath, payload, 0o600); err != nil {
		t.Fatalf("write shared traversal fixture: %v", err)
	}

	loaded, err := scipindex.NewLoader().LoadIndex(indexPath)
	if err != nil {
		t.Fatalf("load shared traversal fixture through official binding path: %v", err)
	}

	return Fixture{
		LoadedIndex: loaded,
		View:        traversal.NewView(loaded),
	}
}

func (fixture Fixture) DocumentByPath(t testing.TB, relativePath string) traversal.Document {
	t.Helper()

	for _, document := range fixture.View.Documents() {
		if document.RelativePath == relativePath {
			return document
		}
	}
	t.Fatalf("shared traversal fixture missing document category path %q", relativePath)
	return traversal.Document{}
}

func (fixture Fixture) SymbolByName(t testing.TB, symbolName string) traversal.Symbol {
	t.Helper()

	for _, document := range fixture.View.Documents() {
		for _, symbol := range document.Symbols {
			if symbol.Symbol == symbolName {
				return symbol
			}
		}
	}
	t.Fatalf("shared traversal fixture missing document symbol category %q", symbolName)
	return traversal.Symbol{}
}

func (fixture Fixture) ExternalSymbolByName(t testing.TB, symbolName string) traversal.Symbol {
	t.Helper()

	for _, symbol := range fixture.View.ExternalSymbols() {
		if symbol.Symbol == symbolName {
			return symbol
		}
	}
	t.Fatalf("shared traversal fixture missing external symbol category %q", symbolName)
	return traversal.Symbol{}
}

func (fixture Fixture) OccurrencesBySymbol(t testing.TB, symbolName string) []traversal.Occurrence {
	t.Helper()

	var occurrences []traversal.Occurrence
	for _, occurrence := range fixture.View.Occurrences() {
		if occurrence.Symbol == symbolName {
			occurrences = append(occurrences, occurrence)
		}
	}
	if len(occurrences) == 0 {
		t.Fatalf("shared traversal fixture missing occurrence category for symbol %q", symbolName)
	}
	return occurrences
}

func (fixture Fixture) OccurrenceByRange(t testing.TB, symbolName string, wantRange []int32) traversal.Occurrence {
	t.Helper()

	for _, occurrence := range fixture.OccurrencesBySymbol(t, symbolName) {
		if slices.Equal(occurrence.Range, wantRange) {
			return occurrence
		}
	}
	t.Fatalf("shared traversal fixture missing range category for symbol %q range %v", symbolName, wantRange)
	return traversal.Occurrence{}
}

func (fixture Fixture) RelationshipsOwnedBy(t testing.TB, sourceSymbol string) []traversal.Relationship {
	t.Helper()

	relationships := fixture.View.RelationshipsOwnedBy(sourceSymbol)
	if len(relationships) == 0 {
		t.Fatalf("shared traversal fixture missing relationship owner category for symbol %q", sourceSymbol)
	}
	return relationships
}

func (fixture Fixture) RelationshipsTargeting(t testing.TB, targetSymbol string) []traversal.Relationship {
	t.Helper()

	relationships := fixture.View.RelationshipsTargeting(targetSymbol)
	if len(relationships) == 0 {
		t.Fatalf("shared traversal fixture missing relationship target category for symbol %q", targetSymbol)
	}
	return relationships
}

func (fixture Fixture) RelationshipByTarget(
	t testing.TB,
	sourceSymbol string,
	targetSymbol string,
) traversal.Relationship {
	t.Helper()

	for _, relationship := range fixture.RelationshipsOwnedBy(t, sourceSymbol) {
		if relationship.TargetSymbol == targetSymbol {
			return relationship
		}
	}
	t.Fatalf("shared traversal fixture missing relationship edge category from %q to %q", sourceSymbol, targetSymbol)
	return traversal.Relationship{}
}
