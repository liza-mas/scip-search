package traversal_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/scip-code/scip/bindings/go/scip"

	"scip-search/internal/traversal"
	"scip-search/internal/traversal/traversaltest"
)

const absentFixtureSymbol = "scip-go gomod example.com/fixture . missing/Absent#"

func TestSharedFixtureCoverageValidationCoversTraversalCategories(t *testing.T) {
	t.Parallel()

	fixture := traversaltest.LoadSharedFixture(t)

	if missing := missingTraversalCategories(fixture.View); len(missing) != 0 {
		t.Fatalf("shared traversal fixture missing categories: %s", strings.Join(missing, ", "))
	}
}

func TestSharedFixtureCoverageReportsMissingTraversalCategories(t *testing.T) {
	t.Parallel()

	missing := missingTraversalCategories(traversal.View{})

	for _, category := range []string{
		"document path coverage",
		"document metadata coverage",
		"document symbol coverage",
		"external symbol coverage",
		"occurrence lookup coverage",
		"relationship owner coverage",
		"relationship target coverage",
		"relationship edge-kind coverage",
		"range coverage",
		"enclosing range coverage",
		"role coverage",
		"symbol hover coverage",
		"occurrence hover coverage",
	} {
		if !slices.Contains(missing, category) {
			t.Fatalf("missing categories = %v, want category-named failure for %q", missing, category)
		}
	}
}

func missingTraversalCategories(view traversal.View) []string {
	var missing []string
	documents := view.Documents()
	externalSymbols := view.ExternalSymbols()
	occurrences := view.Occurrences()

	if !hasDocumentPathCoverage(documents) {
		missing = append(missing, "document path coverage")
	}
	if !hasDocumentMetadataCoverage(documents) {
		missing = append(missing, "document metadata coverage")
	}
	if !hasDocumentSymbolCoverage(documents) {
		missing = append(missing, "document symbol coverage")
	}
	if !hasExternalSymbolCoverage(externalSymbols) {
		missing = append(missing, "external symbol coverage")
	}
	if !hasOccurrenceLookupCoverage(view) {
		missing = append(missing, "occurrence lookup coverage")
	}
	if !hasAbsentLookupCoverage(view) {
		missing = append(missing, "absent lookup coverage")
	}
	if !hasRelationshipOwnerCoverage(view) {
		missing = append(missing, "relationship owner coverage")
	}
	if !hasRelationshipTargetCoverage(view) {
		missing = append(missing, "relationship target coverage")
	}
	if !hasRelationshipEdgeKindCoverage(view) {
		missing = append(missing, "relationship edge-kind coverage")
	}
	if !hasRangeCoverage(occurrences) {
		missing = append(missing, "range coverage")
	}
	if !hasEnclosingRangeCoverage(occurrences) {
		missing = append(missing, "enclosing range coverage")
	}
	if !hasRoleCoverage(occurrences) {
		missing = append(missing, "role coverage")
	}
	if !hasSymbolHoverCoverage(documents) {
		missing = append(missing, "symbol hover coverage")
	}
	if !hasOccurrenceHoverCoverage(occurrences) {
		missing = append(missing, "occurrence hover coverage")
	}
	return missing
}

func hasDocumentPathCoverage(documents []traversal.Document) bool {
	paths := map[string]bool{}
	for _, document := range documents {
		if document.RelativePath != "" {
			paths[document.RelativePath] = true
		}
	}
	return len(paths) >= 2
}

func hasDocumentMetadataCoverage(documents []traversal.Document) bool {
	for _, document := range documents {
		if document.Language == "" || document.PositionEncoding == scip.PositionEncoding_UnspecifiedPositionEncoding {
			return false
		}
	}
	return len(documents) >= 2
}

func hasDocumentSymbolCoverage(documents []traversal.Document) bool {
	for _, document := range documents {
		for _, symbol := range document.Symbols {
			if symbol.Symbol != "" &&
				symbol.Source == traversal.SymbolSourceDocument &&
				symbol.DocumentPath == document.RelativePath {
				return true
			}
		}
	}
	return false
}

func hasExternalSymbolCoverage(symbols []traversal.Symbol) bool {
	for _, symbol := range symbols {
		if symbol.Symbol != "" && symbol.Source == traversal.SymbolSourceExternal {
			return true
		}
	}
	return false
}

func hasOccurrenceLookupCoverage(view traversal.View) bool {
	alphaOccurrences := view.OccurrencesForSymbol(traversaltest.AlphaSymbol)
	betaOccurrences := view.OccurrencesForSymbol(traversaltest.BetaSymbol)
	if len(alphaOccurrences) == 0 || len(betaOccurrences) == 0 {
		return false
	}
	for _, occurrence := range alphaOccurrences {
		if occurrence.Symbol != traversaltest.AlphaSymbol || occurrence.DocumentPath == "" {
			return false
		}
	}
	for _, occurrence := range betaOccurrences {
		if occurrence.Symbol != traversaltest.BetaSymbol || occurrence.DocumentPath == "" {
			return false
		}
	}
	return true
}

func hasAbsentLookupCoverage(view traversal.View) bool {
	return len(view.OccurrencesForSymbol(absentFixtureSymbol)) == 0 &&
		len(view.RelationshipsOwnedBy(absentFixtureSymbol)) == 0 &&
		len(view.RelationshipsTargeting(absentFixtureSymbol)) == 0
}

func hasRelationshipOwnerCoverage(view traversal.View) bool {
	for _, relationship := range view.RelationshipsOwnedBy(traversaltest.AlphaSymbol) {
		if relationship.SourceSymbol == traversaltest.AlphaSymbol &&
			relationship.TargetSymbol == traversaltest.BetaSymbol &&
			relationship.IsReference {
			return true
		}
	}
	return false
}

func hasRelationshipTargetCoverage(view traversal.View) bool {
	for _, relationship := range view.RelationshipsTargeting(traversaltest.MultiFlagTargetSymbol) {
		if relationship.SourceSymbol == traversaltest.ExternalSymbol &&
			relationship.TargetSymbol == traversaltest.MultiFlagTargetSymbol &&
			relationship.IsReference &&
			relationship.IsImplementation &&
			relationship.IsDefinition &&
			relationship.IsTypeDefinition {
			return true
		}
	}
	return false
}

func hasRelationshipEdgeKindCoverage(view traversal.View) bool {
	var hasReference, hasImplementation, hasDefinition, hasTypeDefinition, hasMultiFlag bool
	for _, relationship := range append(
		view.RelationshipsOwnedBy(traversaltest.AlphaSymbol),
		view.RelationshipsOwnedBy(traversaltest.ExternalSymbol)...,
	) {
		hasReference = hasReference || relationship.IsReference
		hasImplementation = hasImplementation || relationship.IsImplementation
		hasDefinition = hasDefinition || relationship.IsDefinition
		hasTypeDefinition = hasTypeDefinition || relationship.IsTypeDefinition
		hasMultiFlag = hasMultiFlag || relationship.IsReference &&
			relationship.IsImplementation &&
			relationship.IsDefinition &&
			relationship.IsTypeDefinition
	}
	return hasReference && hasImplementation && hasDefinition && hasTypeDefinition && hasMultiFlag
}

func hasRangeCoverage(occurrences []traversal.Occurrence) bool {
	var hasSameLine, hasMultiPosition bool
	for _, occurrence := range occurrences {
		hasSameLine = hasSameLine || len(occurrence.Range) == 3
		hasMultiPosition = hasMultiPosition || len(occurrence.Range) == 4
	}
	return hasSameLine && hasMultiPosition
}

func hasEnclosingRangeCoverage(occurrences []traversal.Occurrence) bool {
	var hasPresent, hasAbsent bool
	for _, occurrence := range occurrences {
		hasPresent = hasPresent || occurrence.HasEnclosingRange
		hasAbsent = hasAbsent || !occurrence.HasEnclosingRange
	}
	return hasPresent && hasAbsent
}

func hasRoleCoverage(occurrences []traversal.Occurrence) bool {
	var hasDefinition, hasNonDefinition, hasMultiBit bool
	for _, occurrence := range occurrences {
		hasDefinition = hasDefinition || occurrence.SymbolRoles&int32(scip.SymbolRole_Definition) != 0
		hasNonDefinition = hasNonDefinition || occurrence.SymbolRoles&int32(scip.SymbolRole_Definition) == 0
		hasMultiBit = hasMultiBit || occurrence.SymbolRoles != 0 &&
			occurrence.SymbolRoles&(occurrence.SymbolRoles-1) != 0
	}
	return hasDefinition && hasNonDefinition && hasMultiBit
}

func hasSymbolHoverCoverage(documents []traversal.Document) bool {
	for _, document := range documents {
		for _, symbol := range document.Symbols {
			if symbol.Kind != scip.SymbolInformation_UnspecifiedKind &&
				symbol.DisplayName != "" &&
				len(symbol.Documentation) > 0 &&
				symbol.SignatureDocumentation != nil {
				return true
			}
		}
	}
	return false
}

func hasOccurrenceHoverCoverage(occurrences []traversal.Occurrence) bool {
	var hasPresent, hasAbsent bool
	for _, occurrence := range occurrences {
		hasPresent = hasPresent || len(occurrence.OverrideDocumentation) > 0
		hasAbsent = hasAbsent || len(occurrence.OverrideDocumentation) == 0
	}
	return hasPresent && hasAbsent
}
