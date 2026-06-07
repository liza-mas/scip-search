package graphexport

import (
	"cmp"
	"slices"
	"strings"
	"time"

	"github.com/scip-code/scip/bindings/go/scip"

	runtimecontract "scip-search/internal/runtime"
	"scip-search/internal/scipmodel"
	"scip-search/internal/traversal"
	"scip-search/internal/version"
)

const SchemaVersion = "scip.graph-export.v1"

type Filters struct {
	Symbols         []string
	PackagePrefixes []string
}

type Payload struct {
	SchemaVersion string    `json:"schema_version"`
	Generator     Generator `json:"generator"`
	GeneratedAt   string    `json:"generated_at"`
	Inputs        Inputs    `json:"inputs"`
	Nodes         []Node    `json:"nodes"`
	Edges         []Edge    `json:"edges"`
}

type Generator struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Inputs struct {
	SCIPIndex SCIPIndexInput `json:"scip_index"`
}

type SCIPIndexInput struct {
	Path        string `json:"path"`
	Fingerprint string `json:"fingerprint"`
}

type Node struct {
	ID                  string    `json:"id"`
	DisplayName         string    `json:"display_name,omitempty"`
	Kind                string    `json:"kind,omitempty"`
	Package             string    `json:"package,omitempty"`
	DocumentPath        string    `json:"document_path,omitempty"`
	Location            *Location `json:"location,omitempty"`
	Roles               int32     `json:"roles,omitempty"`
	External            *bool     `json:"external,omitempty"`
	SymbolInfoAvailable bool      `json:"symbol_info_available"`
}

type Location struct {
	Range []int32 `json:"range"`
}

type Edge struct {
	Source          string   `json:"source"`
	Target          string   `json:"target"`
	Type            string   `json:"type"`
	Provenance      string   `json:"provenance"`
	OccurrenceCount int      `json:"occurrence_count"`
	Weight          *float64 `json:"weight,omitempty"`
}

type edgeKey struct {
	source     string
	target     string
	edgeType   string
	provenance string
}

type documentOccurrences struct {
	definitions []traversal.Occurrence
	references  []traversal.Occurrence
}

func Export(
	loaded runtimecontract.LoadedIndex,
	buildIdentity version.BuildIdentity,
	filters Filters,
	now func() time.Time,
) Payload {
	return ExportView(loaded, traversal.NewView(loaded), buildIdentity, filters, now)
}

func ExportView(
	loaded runtimecontract.LoadedIndex,
	view traversal.View,
	buildIdentity version.BuildIdentity,
	filters Filters,
	now func() time.Time,
) Payload {
	if now == nil {
		now = time.Now
	}

	nodesByID := knownNodes(view)
	edgesByKey := collectEdges(view)
	addMissingEndpointNodes(nodesByID, edgesByKey)

	selected := selectedNodeIDs(nodesByID, filters)
	nodes := filteredNodes(nodesByID, selected)
	edges := filteredEdges(edgesByKey, selected)

	return Payload{
		SchemaVersion: SchemaVersion,
		Generator: Generator{
			Name:    "scip-search",
			Version: version.Format(buildIdentity),
		},
		GeneratedAt: now().UTC().Format(time.RFC3339),
		Inputs: Inputs{
			SCIPIndex: SCIPIndexInput{
				Path:        loaded.Path,
				Fingerprint: loaded.Fingerprint,
			},
		},
		Nodes: nodes,
		Edges: edges,
	}
}

func knownNodes(view traversal.View) map[string]Node {
	nodes := map[string]Node{}
	for _, document := range view.Documents() {
		for _, symbol := range document.Symbols {
			external := false
			nodes[symbol.Symbol] = nodeFromSymbol(view, symbol, &external)
		}
	}
	for _, symbol := range view.ExternalSymbols() {
		external := true
		nodes[symbol.Symbol] = nodeFromSymbol(view, symbol, &external)
	}
	return nodes
}

func nodeFromSymbol(view traversal.View, symbol traversal.Symbol, external *bool) Node {
	node := Node{
		ID:                  symbol.Symbol,
		DisplayName:         symbol.DisplayName,
		Kind:                kindName(symbol.Kind),
		DocumentPath:        symbol.DocumentPath,
		External:            external,
		SymbolInfoAvailable: true,
	}
	if identity, err := scipmodel.ParseIdentity(symbol.Symbol); err == nil {
		node.Package = identity.PackageKey()
	}
	if definition, ok := firstDefinition(view, symbol.Symbol); ok {
		node.Location = &Location{Range: slices.Clone(definition.Range)}
		node.Roles = definition.SymbolRoles
	}
	return node
}

func collectEdges(view traversal.View) map[edgeKey]int {
	edges := map[edgeKey]int{}
	for _, document := range view.Documents() {
		for _, symbol := range document.Symbols {
			addRelationshipEdges(edges, view, symbol.Symbol)
		}
	}
	for _, symbol := range view.ExternalSymbols() {
		addRelationshipEdges(edges, view, symbol.Symbol)
	}
	addContainedDependencyEdges(edges, view)
	return edges
}

func addRelationshipEdges(edges map[edgeKey]int, view traversal.View, source string) {
	for _, relationship := range view.RelationshipsOwnedBy(source) {
		if relationship.TargetSymbol == "" {
			continue
		}
		key := edgeKey{
			source:     relationship.SourceSymbol,
			target:     relationship.TargetSymbol,
			edgeType:   relationshipType(relationship),
			provenance: "scip_relationship",
		}
		edges[key]++
	}
}

func addContainedDependencyEdges(edges map[edgeKey]int, view traversal.View) {
	occurrencesByDocument := map[string]documentOccurrences{}
	for _, occurrence := range view.Occurrences() {
		if occurrence.Symbol == "" {
			continue
		}
		group := occurrencesByDocument[occurrence.DocumentPath]
		if isDefinition(occurrence) {
			if occurrence.HasEnclosingRange && len(occurrence.EnclosingRange) > 0 {
				group.definitions = append(group.definitions, occurrence)
			}
		} else {
			group.references = append(group.references, occurrence)
		}
		occurrencesByDocument[occurrence.DocumentPath] = group
	}

	for _, group := range occurrencesByDocument {
		for _, definition := range group.definitions {
			for _, occurrence := range group.references {
				if !rangeWithin(occurrence.Range, definition.EnclosingRange) {
					continue
				}
				key := edgeKey{
					source:     definition.Symbol,
					target:     occurrence.Symbol,
					edgeType:   "dependency",
					provenance: "contained_dependency",
				}
				edges[key]++
			}
		}
	}
}

func addMissingEndpointNodes(nodes map[string]Node, edges map[edgeKey]int) {
	for key := range edges {
		if _, ok := nodes[key.source]; !ok {
			nodes[key.source] = Node{ID: key.source, SymbolInfoAvailable: false}
		}
		if _, ok := nodes[key.target]; !ok {
			nodes[key.target] = Node{ID: key.target, SymbolInfoAvailable: false}
		}
	}
}

func selectedNodeIDs(nodes map[string]Node, filters Filters) map[string]bool {
	if len(filters.Symbols) == 0 && len(filters.PackagePrefixes) == 0 {
		selected := make(map[string]bool, len(nodes))
		for id := range nodes {
			selected[id] = true
		}
		return selected
	}

	requested := map[string]bool{}
	for _, symbol := range filters.Symbols {
		requested[symbol] = true
	}

	selected := map[string]bool{}
	for id, node := range nodes {
		if len(requested) > 0 && !requested[id] {
			continue
		}
		if len(filters.PackagePrefixes) > 0 && !matchesAnyPackagePrefix(node.Package, filters.PackagePrefixes) {
			continue
		}
		selected[id] = true
	}
	return selected
}

func filteredNodes(nodes map[string]Node, selected map[string]bool) []Node {
	results := make([]Node, 0, len(selected))
	for id := range selected {
		results = append(results, nodes[id])
	}
	slices.SortFunc(results, compareNodes)
	return results
}

func filteredEdges(edges map[edgeKey]int, selected map[string]bool) []Edge {
	results := make([]Edge, 0, len(edges))
	for key, count := range edges {
		if !selected[key.source] || !selected[key.target] {
			continue
		}
		results = append(results, Edge{
			Source:          key.source,
			Target:          key.target,
			Type:            key.edgeType,
			Provenance:      key.provenance,
			OccurrenceCount: count,
		})
	}
	slices.SortFunc(results, compareEdges)
	return results
}

func matchesAnyPackagePrefix(packageKey string, prefixes []string) bool {
	if packageKey == "" {
		return false
	}
	packageName := packageNameFromKey(packageKey)
	for _, prefix := range prefixes {
		if strings.HasPrefix(packageName, prefix) || strings.HasPrefix(packageKey, prefix) {
			return true
		}
	}
	return false
}

func packageNameFromKey(packageKey string) string {
	fields := strings.Fields(packageKey)
	if len(fields) < 3 {
		return ""
	}
	return fields[2]
}

func firstDefinition(view traversal.View, symbol string) (traversal.Occurrence, bool) {
	definitions := make([]traversal.Occurrence, 0)
	for _, occurrence := range view.OccurrencesForSymbol(symbol) {
		if isDefinition(occurrence) {
			definitions = append(definitions, occurrence)
		}
	}
	slices.SortFunc(definitions, compareTraversalOccurrences)
	if len(definitions) == 0 {
		return traversal.Occurrence{}, false
	}
	return definitions[0], true
}

func relationshipType(relationship traversal.Relationship) string {
	switch {
	case relationship.IsReference:
		return "reference"
	case relationship.IsImplementation:
		return "implementation"
	case relationship.IsTypeDefinition:
		return "type-definition"
	case relationship.IsDefinition:
		return "definition"
	default:
		return "relationship"
	}
}

func kindName(kind scip.SymbolInformation_Kind) string {
	name := kind.String()
	name = strings.TrimPrefix(name, "SymbolInformation_")
	name = strings.ReplaceAll(name, "_", "-")
	return strings.ToLower(name)
}

func isDefinition(occurrence traversal.Occurrence) bool {
	return occurrence.SymbolRoles&int32(scip.SymbolRole_Definition) != 0
}

func rangeWithin(scipRange []int32, enclosingRange []int32) bool {
	startLine, startCharacter, endLine, endCharacter := rangeValues(scipRange)
	enclosingStartLine, enclosingStartCharacter, enclosingEndLine, enclosingEndCharacter := rangeValues(enclosingRange)
	if startLine < enclosingStartLine || endLine > enclosingEndLine {
		return false
	}
	if startLine == enclosingStartLine && startCharacter < enclosingStartCharacter {
		return false
	}
	if endLine == enclosingEndLine && endCharacter > enclosingEndCharacter {
		return false
	}
	return true
}

func rangeValues(scipRange []int32) (int32, int32, int32, int32) {
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

func compareTraversalOccurrences(left traversal.Occurrence, right traversal.Occurrence) int {
	if byDocument := cmp.Compare(left.DocumentPath, right.DocumentPath); byDocument != 0 {
		return byDocument
	}
	return slices.Compare(left.Range, right.Range)
}

func compareNodes(left Node, right Node) int {
	return cmp.Compare(left.ID, right.ID)
}

func compareEdges(left Edge, right Edge) int {
	if bySource := cmp.Compare(left.Source, right.Source); bySource != 0 {
		return bySource
	}
	if byTarget := cmp.Compare(left.Target, right.Target); byTarget != 0 {
		return byTarget
	}
	if byType := cmp.Compare(left.Type, right.Type); byType != 0 {
		return byType
	}
	return cmp.Compare(left.Provenance, right.Provenance)
}
