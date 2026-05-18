package implementations

import (
	"cmp"
	"slices"

	"github.com/scip-code/scip/bindings/go/scip"

	"scip-search/internal/traversal"
)

type Payload struct {
	Symbol          string   `json:"symbol"`
	Implementations []Result `json:"implementations"`
}

type Result struct {
	ImplementationSymbol string       `json:"implementationSymbol"`
	Relationship         Relationship `json:"relationship"`
	DocumentPath         string       `json:"documentPath,omitempty"`
	Range                []int32      `json:"range,omitempty"`
}

type Relationship struct {
	SourceSymbol     string `json:"sourceSymbol"`
	TargetSymbol     string `json:"targetSymbol"`
	IsReference      bool   `json:"isReference,omitempty"`
	IsImplementation bool   `json:"isImplementation"`
	IsTypeDefinition bool   `json:"isTypeDefinition,omitempty"`
	IsDefinition     bool   `json:"isDefinition,omitempty"`
}

func Implementations(view traversal.View, symbol string) (Payload, error) {
	definitions := definitionsBySymbol(view.Occurrences())
	resultsBySymbol := map[string]Result{}

	for _, relationship := range view.RelationshipsTargeting(symbol) {
		if !relationship.IsImplementation {
			continue
		}
		if _, exists := resultsBySymbol[relationship.SourceSymbol]; exists {
			continue
		}

		result := Result{
			ImplementationSymbol: relationship.SourceSymbol,
			Relationship: Relationship{
				SourceSymbol:     relationship.SourceSymbol,
				TargetSymbol:     relationship.TargetSymbol,
				IsReference:      relationship.IsReference,
				IsImplementation: relationship.IsImplementation,
				IsTypeDefinition: relationship.IsTypeDefinition,
				IsDefinition:     relationship.IsDefinition,
			},
		}
		if definition, ok := definitions[relationship.SourceSymbol]; ok {
			result.DocumentPath = definition.DocumentPath
			result.Range = slices.Clone(definition.Range)
		}
		resultsBySymbol[relationship.SourceSymbol] = result
	}

	results := make([]Result, 0, len(resultsBySymbol))
	for _, result := range resultsBySymbol {
		results = append(results, result)
	}
	slices.SortFunc(results, compareResults)

	return Payload{Symbol: symbol, Implementations: results}, nil
}

type definition struct {
	DocumentPath string
	Range        []int32
}

func definitionsBySymbol(occurrences []traversal.Occurrence) map[string]definition {
	definitions := map[string]definition{}
	for _, occurrence := range occurrences {
		if occurrence.Symbol == "" || !isDefinition(occurrence) {
			continue
		}
		if _, exists := definitions[occurrence.Symbol]; exists {
			continue
		}
		definitions[occurrence.Symbol] = definition{
			DocumentPath: occurrence.DocumentPath,
			Range:        slices.Clone(occurrence.Range),
		}
	}
	return definitions
}

func isDefinition(occurrence traversal.Occurrence) bool {
	return occurrence.SymbolRoles&int32(scip.SymbolRole_Definition) != 0
}

func compareResults(left Result, right Result) int {
	if bySymbol := cmp.Compare(left.ImplementationSymbol, right.ImplementationSymbol); bySymbol != 0 {
		return bySymbol
	}
	if byDocument := cmp.Compare(left.DocumentPath, right.DocumentPath); byDocument != 0 {
		return byDocument
	}
	return slices.Compare(left.Range, right.Range)
}
