package references

import (
	"fmt"
	"sort"
	"strings"

	"github.com/scip-code/scip/bindings/go/scip"

	"scip-search/internal/query/oneline"
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

func OneLine(payload Payload) string {
	if len(payload.References) == 0 {
		return ""
	}

	var builder strings.Builder
	for _, reference := range payload.References {
		path, line, column := oneline.Location(reference.DocumentPath, reference.Range)
		fmt.Fprintf(
			&builder,
			"%s:%d:%d:%s roles=%d\n",
			path,
			line,
			column,
			reference.Symbol,
			reference.Roles,
		)
	}

	return builder.String()
}

func referenceCandidateSymbols(view traversal.View, symbol string) map[string]struct{} {
	seen := map[string]struct{}{symbol: {}}

	for _, relationship := range view.RelationshipsOwnedBy(symbol) {
		if !relationship.IsReference || relationship.TargetSymbol == "" {
			continue
		}
		seen[relationship.TargetSymbol] = struct{}{}
	}
	for _, relationship := range view.RelationshipsTargeting(symbol) {
		if !relationship.IsReference || relationship.SourceSymbol == "" {
			continue
		}
		seen[relationship.SourceSymbol] = struct{}{}
	}

	return seen
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
