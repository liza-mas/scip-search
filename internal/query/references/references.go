package references

import (
	"sort"

	"github.com/scip-code/scip/bindings/go/scip"

	"scip-search/internal/traversal"
)

type Payload struct {
	Symbol     string      `json:"symbol"`
	References []Reference `json:"references"`
}

type Reference struct {
	Symbol       string  `json:"symbol"`
	DocumentPath string  `json:"documentPath"`
	Range        []int32 `json:"range"`
	Roles        int32   `json:"roles"`
}

func Query(view traversal.View, symbol string) Payload {
	candidates := referenceCandidateSymbols(view, symbol)
	results := make([]Reference, 0)

	for candidate := range candidates {
		for _, occurrence := range view.OccurrencesForSymbol(candidate) {
			if isDefinition(occurrence) {
				continue
			}
			results = append(results, Reference{
				Symbol:       occurrence.Symbol,
				DocumentPath: occurrence.DocumentPath,
				Range:        append([]int32(nil), occurrence.Range...),
				Roles:        occurrence.SymbolRoles,
			})
		}
	}

	sort.SliceStable(results, func(left, right int) bool {
		return referenceLess(results[left], results[right])
	})

	return Payload{
		Symbol:     symbol,
		References: results,
	}
}

func referenceCandidateSymbols(view traversal.View, symbol string) map[string]struct{} {
	seen := map[string]struct{}{symbol: {}}
	queue := []string{symbol}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, relationship := range view.RelationshipsOwnedBy(current) {
			if !relationship.IsReference {
				continue
			}
			if addCandidate(seen, relationship.TargetSymbol) {
				queue = append(queue, relationship.TargetSymbol)
			}
		}
		for _, relationship := range view.RelationshipsTargeting(current) {
			if !relationship.IsReference {
				continue
			}
			if addCandidate(seen, relationship.SourceSymbol) {
				queue = append(queue, relationship.SourceSymbol)
			}
		}
	}

	return seen
}

func addCandidate(seen map[string]struct{}, symbol string) bool {
	if symbol == "" {
		return false
	}
	if _, exists := seen[symbol]; exists {
		return false
	}
	seen[symbol] = struct{}{}
	return true
}

func isDefinition(occurrence traversal.Occurrence) bool {
	return occurrence.SymbolRoles&int32(scip.SymbolRole_Definition) != 0
}

func referenceLess(left Reference, right Reference) bool {
	if left.DocumentPath != right.DocumentPath {
		return left.DocumentPath < right.DocumentPath
	}
	leftStartLine, leftStartCharacter, leftEndLine, leftEndCharacter := rangeOrderValues(left.Range)
	rightStartLine, rightStartCharacter, rightEndLine, rightEndCharacter := rangeOrderValues(right.Range)
	if leftStartLine != rightStartLine {
		return leftStartLine < rightStartLine
	}
	if leftStartCharacter != rightStartCharacter {
		return leftStartCharacter < rightStartCharacter
	}
	if leftEndLine != rightEndLine {
		return leftEndLine < rightEndLine
	}
	if leftEndCharacter != rightEndCharacter {
		return leftEndCharacter < rightEndCharacter
	}
	return left.Symbol < right.Symbol
}

func rangeOrderValues(scipRange []int32) (int32, int32, int32, int32) {
	if len(scipRange) >= 4 {
		return scipRange[0], scipRange[1], scipRange[2], scipRange[3]
	}
	if len(scipRange) >= 3 {
		return scipRange[0], scipRange[1], scipRange[0], scipRange[2]
	}
	if len(scipRange) >= 2 {
		return scipRange[0], scipRange[1], scipRange[0], scipRange[1]
	}
	if len(scipRange) == 1 {
		return scipRange[0], 0, scipRange[0], 0
	}
	return 0, 0, 0, 0
}
