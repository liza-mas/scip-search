package traversaltest

import (
	"slices"
	"testing"

	"github.com/scip-code/scip/bindings/go/scip"

	"scip-search/internal/traversal"
)

func TestSharedFixtureLoadsThroughTraversalView(t *testing.T) {
	t.Parallel()

	fixture := LoadSharedFixture(t)

	if fixture.LoadedIndex.Path == "" {
		t.Fatal("fixture loaded index path is empty, want production loader path")
	}

	documents := fixture.View.Documents()
	if len(documents) < 2 {
		t.Fatalf("fixture document count = %d, want at least 2 traversal coverage documents", len(documents))
	}

	alpha := fixture.DocumentByPath(t, "cmd/alpha.go")
	beta := fixture.DocumentByPath(t, "pkg/beta.go")
	if alpha.RelativePath == beta.RelativePath {
		t.Fatalf("fixture document paths are not distinct: %q", alpha.RelativePath)
	}
	assertDocumentMetadata(t, alpha, "go", scip.PositionEncoding_UTF8CodeUnitOffsetFromLineStart)
	assertDocumentMetadata(t, beta, "go", scip.PositionEncoding_UTF16CodeUnitOffsetFromLineStart)

	localSymbol := fixture.SymbolByName(t, AlphaSymbol)
	if localSymbol.Source != traversal.SymbolSourceDocument || localSymbol.DocumentPath != "cmd/alpha.go" {
		t.Fatalf("local symbol source = %q path = %q, want document symbol in cmd/alpha.go", localSymbol.Source, localSymbol.DocumentPath)
	}
	if localSymbol.Kind != scip.SymbolInformation_Function ||
		localSymbol.DisplayName != "Alpha" ||
		!slices.Equal(localSymbol.Documentation, []string{"Alpha renders fixture data."}) ||
		localSymbol.SignatureDocumentation.GetText() != "func Alpha() string" {
		t.Fatalf("local symbol hover metadata = %+v, want kind, display name, docs, and signature docs", localSymbol)
	}

	externalSymbol := fixture.ExternalSymbolByName(t, ExternalSymbol)
	if externalSymbol.Source != traversal.SymbolSourceExternal {
		t.Fatalf("external symbol source = %q, want external", externalSymbol.Source)
	}
}

func TestSharedFixtureCoversOccurrencesRangesRolesAndHover(t *testing.T) {
	t.Parallel()

	fixture := LoadSharedFixture(t)

	alphaOccurrences := fixture.OccurrencesBySymbol(t, AlphaSymbol)
	if len(alphaOccurrences) != 2 {
		t.Fatalf("alpha occurrence count = %d, want 2", len(alphaOccurrences))
	}
	betaOccurrences := fixture.OccurrencesBySymbol(t, BetaSymbol)
	if len(betaOccurrences) != 1 {
		t.Fatalf("beta occurrence count = %d, want 1", len(betaOccurrences))
	}

	definition := fixture.OccurrenceByRange(t, AlphaSymbol, []int32{2, 6, 11})
	if !slices.Equal(definition.EnclosingRange, []int32{1, 0, 5, 1}) || !definition.HasEnclosingRange {
		t.Fatalf("definition enclosing range = present:%v value:%v, want present SCIP enclosing range", definition.HasEnclosingRange, definition.EnclosingRange)
	}
	if definition.SymbolRoles != int32(scip.SymbolRole_Definition|scip.SymbolRole_WriteAccess) {
		t.Fatalf("definition roles = %d, want definition plus write-access bitset", definition.SymbolRoles)
	}
	if !slices.Equal(definition.OverrideDocumentation, []string{"Alpha occurrence override documentation."}) {
		t.Fatalf("definition override docs = %v, want present override documentation", definition.OverrideDocumentation)
	}

	reference := fixture.OccurrenceByRange(t, AlphaSymbol, []int32{8, 1, 8, 15})
	if reference.HasEnclosingRange || reference.EnclosingRange != nil {
		t.Fatalf("reference enclosing range = present:%v value:%v, want absent", reference.HasEnclosingRange, reference.EnclosingRange)
	}
	if reference.SymbolRoles != int32(scip.SymbolRole_ReadAccess) {
		t.Fatalf("reference roles = %d, want non-definition read-access role", reference.SymbolRoles)
	}
	if reference.OverrideDocumentation != nil {
		t.Fatalf("reference override docs = %v, want absent override documentation", reference.OverrideDocumentation)
	}

	multiPosition := fixture.OccurrenceByRange(t, BetaSymbol, []int32{12, 4, 14, 2})
	if !multiPosition.HasEnclosingRange || !slices.Equal(multiPosition.EnclosingRange, []int32{11, 0, 15, 0}) {
		t.Fatalf("multi-position enclosing range = present:%v value:%v, want present multi-line enclosing range", multiPosition.HasEnclosingRange, multiPosition.EnclosingRange)
	}
}

func TestSharedFixtureCoversRelationshipFacts(t *testing.T) {
	t.Parallel()

	fixture := LoadSharedFixture(t)

	owned := fixture.RelationshipsOwnedBy(t, AlphaSymbol)
	assertRelationship(t, owned, traversal.Relationship{
		SourceSymbol: AlphaSymbol,
		TargetSymbol: BetaSymbol,
		IsReference:  true,
	})
	assertRelationship(t, owned, traversal.Relationship{
		SourceSymbol:     AlphaSymbol,
		TargetSymbol:     ImplSymbol,
		IsImplementation: true,
	})
	assertRelationship(t, owned, traversal.Relationship{
		SourceSymbol: AlphaSymbol,
		TargetSymbol: DefinitionSymbol,
		IsDefinition: true,
	})
	assertRelationship(t, owned, traversal.Relationship{
		SourceSymbol:     AlphaSymbol,
		TargetSymbol:     TypeDefinitionSymbol,
		IsTypeDefinition: true,
	})

	multiFlag := fixture.RelationshipByTarget(t, ExternalSymbol, MultiFlagTargetSymbol)
	if !multiFlag.IsReference || !multiFlag.IsImplementation || !multiFlag.IsDefinition || !multiFlag.IsTypeDefinition {
		t.Fatalf("multi-flag relationship = %+v, want all edge-kind flags preserved together", multiFlag)
	}

	targetingBeta := fixture.RelationshipsTargeting(t, BetaSymbol)
	assertRelationship(t, targetingBeta, traversal.Relationship{
		SourceSymbol: AlphaSymbol,
		TargetSymbol: BetaSymbol,
		IsReference:  true,
	})
}

func assertDocumentMetadata(t *testing.T, document traversal.Document, language string, encoding scip.PositionEncoding) {
	t.Helper()

	if document.Language != language || document.PositionEncoding != encoding {
		t.Fatalf("document %q metadata = language:%q encoding:%v, want %q %v", document.RelativePath, document.Language, document.PositionEncoding, language, encoding)
	}
}

func assertRelationship(t *testing.T, relationships []traversal.Relationship, want traversal.Relationship) {
	t.Helper()

	for _, relationship := range relationships {
		if relationship == want {
			return
		}
	}
	t.Fatalf("relationships = %+v, want %+v", relationships, want)
}
