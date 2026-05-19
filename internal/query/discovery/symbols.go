package discovery

import (
	"fmt"
	"sort"
	"strings"

	"github.com/scip-code/scip/bindings/go/scip"

	"scip-search/internal/scipmodel"
	"scip-search/internal/traversal"
)

type MatchSource string

const (
	MatchSourceDisplayName MatchSource = "displayName"
	MatchSourceDescriptor  MatchSource = "descriptor"
)

type SymbolsPayload struct {
	Symbols []SymbolResult `json:"symbols"`
}

type SymbolPackagesPayload struct {
	Packages []SymbolPackageResult `json:"packages"`
}

type SymbolPackageResult struct {
	Scheme         string                `json:"scheme"`
	PackageManager string                `json:"packageManager"`
	PackageName    string                `json:"packageName"`
	PackageVersion string                `json:"packageVersion"`
	PackageKey     string                `json:"packageKey"`
	Symbols        []CompactSymbolResult `json:"symbols"`
}

type CompactSymbolResult struct {
	Descriptor  string      `json:"descriptor"`
	MatchText   string      `json:"matchText"`
	MatchSource MatchSource `json:"matchSource"`
	Definition  *Definition `json:"definition,omitempty"`
}

type SymbolResult struct {
	Symbol         string      `json:"symbol"`
	Scheme         string      `json:"scheme"`
	PackageManager string      `json:"packageManager"`
	PackageName    string      `json:"packageName"`
	PackageVersion string      `json:"packageVersion"`
	MatchText      string      `json:"matchText"`
	MatchSource    MatchSource `json:"matchSource"`
	Definition     *Definition `json:"definition,omitempty"`
}

type Definition struct {
	DocumentPath string  `json:"documentPath"`
	Range        []int32 `json:"range"`
}

func SymbolsByName(view traversal.View, name string) (SymbolPackagesPayload, error) {
	return SymbolsByNames(view, []string{name})
}

func SymbolsByNames(view traversal.View, names []string) (SymbolPackagesPayload, error) {
	flat, err := FlatSymbolsByNames(view, names)
	if err != nil {
		return SymbolPackagesPayload{}, err
	}

	packagesByKey := map[string]int{}
	packages := make([]SymbolPackageResult, 0)
	for _, symbol := range flat.Symbols {
		identity, err := scipmodel.ParseIdentity(symbol.Symbol)
		if err != nil {
			return SymbolPackagesPayload{}, err
		}

		packageKey := identity.PackageKey()
		packageIndex, exists := packagesByKey[packageKey]
		if !exists {
			packagesByKey[packageKey] = len(packages)
			packageIndex = len(packages)
			packages = append(packages, SymbolPackageResult{
				Scheme:         symbol.Scheme,
				PackageManager: symbol.PackageManager,
				PackageName:    symbol.PackageName,
				PackageVersion: symbol.PackageVersion,
				PackageKey:     packageKey,
				Symbols:        make([]CompactSymbolResult, 0),
			})
		}

		packages[packageIndex].Symbols = append(packages[packageIndex].Symbols, CompactSymbolResult{
			Descriptor:  identity.Descriptor,
			MatchText:   symbol.MatchText,
			MatchSource: symbol.MatchSource,
			Definition:  cloneDefinition(symbol.Definition),
		})
	}

	return SymbolPackagesPayload{Packages: packages}, nil
}

func FlatSymbolsByName(view traversal.View, name string) (SymbolsPayload, error) {
	return FlatSymbolsByNames(view, []string{name})
}

func FlatSymbolsByNames(view traversal.View, names []string) (SymbolsPayload, error) {
	definitions := definitionsBySymbol(view.Occurrences())
	results := make([]SymbolResult, 0)

	for _, symbol := range symbolsInView(view) {
		if scipmodel.IsLocalSymbol(symbol.Symbol) {
			continue
		}
		identity, err := scipmodel.ParseIdentity(symbol.Symbol)
		if err != nil {
			return SymbolsPayload{}, err
		}

		matchText, matchSource, matched := matchAnyName(identity, symbol.DisplayName, names)
		if !matched {
			continue
		}

		result := SymbolResult{
			Symbol:         identity.Symbol,
			Scheme:         identity.Scheme,
			PackageManager: identity.PackageManager,
			PackageName:    identity.PackageName,
			PackageVersion: identity.PackageVersion,
			MatchText:      matchText,
			MatchSource:    matchSource,
			Definition:     cloneDefinition(definitions[identity.Symbol]),
		}
		results = append(results, result)
	}

	sort.SliceStable(results, func(left, right int) bool {
		return results[left].Symbol < results[right].Symbol
	})

	return SymbolsPayload{Symbols: results}, nil
}

func OneLineSymbolsByName(view traversal.View, name string) (string, error) {
	return OneLineSymbolsByNames(view, []string{name})
}

func OneLineSymbolsByNames(view traversal.View, names []string) (string, error) {
	flat, err := FlatSymbolsByNames(view, names)
	if err != nil {
		return "", err
	}
	if len(flat.Symbols) == 0 {
		return "", nil
	}

	var builder strings.Builder
	for _, symbol := range flat.Symbols {
		identity, err := scipmodel.ParseIdentity(symbol.Symbol)
		if err != nil {
			return "", err
		}
		path, line, column := oneLineLocation(symbol.Definition)
		fmt.Fprintf(
			&builder,
			"%s:%d:%d:%s %s match=%s text=%s\n",
			path,
			line,
			column,
			identity.PackageKey(),
			identity.Descriptor,
			symbol.MatchSource,
			escapeOneLineValue(symbol.MatchText),
		)
	}

	return builder.String(), nil
}

func oneLineLocation(definition *Definition) (string, int32, int32) {
	if definition == nil {
		return "?", 0, 0
	}
	if len(definition.Range) < 2 {
		return definition.DocumentPath, 0, 0
	}

	return definition.DocumentPath, definition.Range[0] + 1, definition.Range[1] + 1
}

func escapeOneLineValue(value string) string {
	var builder strings.Builder
	for _, char := range value {
		switch char {
		case '\\':
			builder.WriteString(`\\`)
		case '\n':
			builder.WriteString(`\n`)
		case '\r':
			builder.WriteString(`\r`)
		case '\t':
			builder.WriteString(`\t`)
		default:
			builder.WriteRune(char)
		}
	}

	return builder.String()
}

func symbolsInView(view traversal.View) []traversal.Symbol {
	documents := view.Documents()
	symbols := make([]traversal.Symbol, 0)
	for _, document := range documents {
		symbols = append(symbols, document.Symbols...)
	}
	symbols = append(symbols, view.ExternalSymbols()...)
	return symbols
}

func matchAnyName(identity scipmodel.Identity, displayName string, names []string) (string, MatchSource, bool) {
	for _, name := range names {
		matchText, matchSource, matched := matchSymbol(identity, displayName, name)
		if matched {
			return matchText, matchSource, true
		}
	}
	return "", "", false
}

func matchSymbol(identity scipmodel.Identity, displayName string, name string) (string, MatchSource, bool) {
	if strings.Contains(displayName, name) {
		return identity.MatchText(displayName), MatchSourceDisplayName, true
	}
	if strings.Contains(identity.Descriptor, name) {
		return identity.Descriptor, MatchSourceDescriptor, true
	}
	return "", "", false
}

func definitionsBySymbol(occurrences []traversal.Occurrence) map[string]*Definition {
	definitions := map[string]*Definition{}
	for _, occurrence := range occurrences {
		if occurrence.Symbol == "" || !isDefinition(occurrence) {
			continue
		}
		if _, exists := definitions[occurrence.Symbol]; exists {
			continue
		}
		definitions[occurrence.Symbol] = &Definition{
			DocumentPath: occurrence.DocumentPath,
			Range:        append([]int32(nil), occurrence.Range...),
		}
	}
	return definitions
}

func cloneDefinition(definition *Definition) *Definition {
	if definition == nil {
		return nil
	}
	return &Definition{
		DocumentPath: definition.DocumentPath,
		Range:        append([]int32(nil), definition.Range...),
	}
}

func isDefinition(occurrence traversal.Occurrence) bool {
	return occurrence.SymbolRoles&int32(scip.SymbolRole_Definition) != 0
}
