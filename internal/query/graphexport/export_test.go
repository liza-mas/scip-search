package graphexport_test

import (
	"encoding/json"
	"slices"
	"testing"
	"time"

	"github.com/scip-code/scip/bindings/go/scip"

	"scip-search/internal/query/graphexport"
	runtimecontract "scip-search/internal/runtime"
	"scip-search/internal/version"
)

const (
	targetSymbol     = "scip-go gomod example.com/project . pkg/Target()."
	dependencySymbol = "scip-go gomod example.com/project . pkg/Dependency()."
	externalSymbol   = "scip-go gomod example.com/external v1.0.0 ext/External()."
	missingSymbol    = "scip-go gomod example.com/missing v1.0.0 missing/Missing()."
)

func TestExportReturnsMetadataNodesAndFactualEdges(t *testing.T) {
	t.Parallel()

	payload := graphexport.Export(
		graphExportFixture(),
		version.BuildIdentity{Release: "v1.2.3"},
		graphexport.Filters{},
		fixedClock,
	)

	if payload.SchemaVersion != graphexport.SchemaVersion {
		t.Fatalf("schema version = %q, want %q", payload.SchemaVersion, graphexport.SchemaVersion)
	}
	if payload.Generator.Name != "scip-search" || payload.Generator.Version != "scip-search release v1.2.3" {
		t.Fatalf("generator = %+v, want scip-search release metadata", payload.Generator)
	}
	if payload.GeneratedAt != "2026-06-07T11:00:00Z" {
		t.Fatalf("generated_at = %q, want fixed UTC RFC3339 timestamp", payload.GeneratedAt)
	}
	if payload.Inputs.SCIPIndex.Path != "fixture.scip" || payload.Inputs.SCIPIndex.Fingerprint != "sha256:fixture" {
		t.Fatalf("input metadata = %+v, want selected path and fingerprint", payload.Inputs.SCIPIndex)
	}

	assertNodeClosure(t, payload)

	target := nodeByID(t, payload, targetSymbol)
	if target.DisplayName != "Target" || target.Kind != "function" || target.Package != "scip-go gomod example.com/project ." {
		t.Fatalf("target node = %+v, want known SCIP symbol facts", target)
	}
	if target.External == nil || *target.External {
		t.Fatalf("target external = %v, want known document symbol false", target.External)
	}
	if !target.SymbolInfoAvailable {
		t.Fatalf("target symbol_info_available = false, want true")
	}
	if target.DocumentPath != "pkg/target.go" {
		t.Fatalf("target document path = %q, want indexed document path", target.DocumentPath)
	}
	if target.Location == nil || !slices.Equal(target.Location.Range, []int32{2, 5, 11}) {
		t.Fatalf("target location = %+v, want definition range", target.Location)
	}

	knownExternal := nodeByID(t, payload, externalSymbol)
	if knownExternal.External == nil || !*knownExternal.External {
		t.Fatalf("external node external = %v, want known external symbol true", knownExternal.External)
	}
	if !knownExternal.SymbolInfoAvailable {
		t.Fatalf("external symbol_info_available = false, want true")
	}

	minimal := nodeByID(t, payload, missingSymbol)
	if minimal.SymbolInfoAvailable {
		t.Fatalf("minimal symbol_info_available = true, want false for endpoint without symbol inventory")
	}
	rawNode := marshalNode(t, minimal)
	if _, exists := rawNode["external"]; exists {
		t.Fatalf("minimal node JSON = %#v, must omit unknown external fact", rawNode)
	}

	assertEdge(t, payload, targetSymbol, dependencySymbol, "dependency", "contained_dependency", 2)
	assertEdge(t, payload, targetSymbol, externalSymbol, "reference", "scip_relationship", 1)
	assertEdge(t, payload, targetSymbol, missingSymbol, "implementation", "scip_relationship", 1)
}

func TestExportFiltersKeepOnlySelfContainedSelectedSubgraph(t *testing.T) {
	t.Parallel()

	payload := graphexport.Export(
		graphExportFixture(),
		version.BuildIdentity{},
		graphexport.Filters{PackagePrefixes: []string{"scip-go gomod example.com/project"}},
		fixedClock,
	)

	assertNodeClosure(t, payload)
	if hasNode(payload, externalSymbol) {
		t.Fatalf("external node included despite package-prefix filter")
	}
	if hasNode(payload, missingSymbol) {
		t.Fatalf("minimal missing node included despite package-prefix filter")
	}
	if hasEdge(payload, targetSymbol, externalSymbol, "reference", "scip_relationship") {
		t.Fatalf("edge to filtered-out external endpoint was emitted")
	}
	if !hasEdge(payload, targetSymbol, dependencySymbol, "dependency", "contained_dependency") {
		t.Fatalf("contained dependency within selected package was not emitted")
	}
}

func TestExportOmitsWeightByDefaultAndKeepsExplicitArrays(t *testing.T) {
	t.Parallel()

	payload := graphexport.Export(graphExportFixture(), version.BuildIdentity{}, graphexport.Filters{}, fixedClock)
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal(Payload) error = %v", err)
	}
	var object map[string]any
	if err := json.Unmarshal(raw, &object); err != nil {
		t.Fatalf("json.Unmarshal(Payload) error = %v", err)
	}
	if _, ok := object["nodes"].([]any); !ok {
		t.Fatalf("nodes = %#v, want explicit JSON array", object["nodes"])
	}
	edges, ok := object["edges"].([]any)
	if !ok {
		t.Fatalf("edges = %#v, want explicit JSON array", object["edges"])
	}
	for _, rawEdge := range edges {
		edge := rawEdge.(map[string]any)
		if _, exists := edge["weight"]; exists {
			t.Fatalf("edge JSON = %#v, weight must be omitted by default", edge)
		}
	}
}

func graphExportFixture() runtimecontract.LoadedIndex {
	return runtimecontract.LoadedIndex{
		Path:        "fixture.scip",
		Fingerprint: "sha256:fixture",
		Index: &scip.Index{
			Documents: []*scip.Document{
				{
					RelativePath: "pkg/target.go",
					Language:     "go",
					Symbols: []*scip.SymbolInformation{
						{
							Symbol:      targetSymbol,
							DisplayName: "Target",
							Kind:        scip.SymbolInformation_Function,
							Relationships: []*scip.Relationship{
								{Symbol: externalSymbol, IsReference: true},
								{Symbol: missingSymbol, IsImplementation: true},
							},
						},
						{
							Symbol:      dependencySymbol,
							DisplayName: "Dependency",
							Kind:        scip.SymbolInformation_Function,
						},
					},
					Occurrences: []*scip.Occurrence{
						{
							Symbol:         targetSymbol,
							Range:          []int32{2, 5, 11},
							EnclosingRange: []int32{2, 0, 8, 1},
							SymbolRoles:    int32(scip.SymbolRole_Definition),
						},
						{
							Symbol:      dependencySymbol,
							Range:       []int32{4, 2, 12},
							SymbolRoles: int32(scip.SymbolRole_ReadAccess),
						},
						{
							Symbol:      dependencySymbol,
							Range:       []int32{5, 2, 12},
							SymbolRoles: int32(scip.SymbolRole_ReadAccess),
						},
					},
				},
				{
					RelativePath: "pkg/other.go",
					Language:     "go",
					Occurrences: []*scip.Occurrence{
						{
							Symbol:      dependencySymbol,
							Range:       []int32{4, 2, 12},
							SymbolRoles: int32(scip.SymbolRole_ReadAccess),
						},
					},
				},
			},
			ExternalSymbols: []*scip.SymbolInformation{
				{
					Symbol:      externalSymbol,
					DisplayName: "External",
					Kind:        scip.SymbolInformation_Function,
				},
			},
		},
	}
}

func fixedClock() time.Time {
	return time.Date(2026, 6, 7, 13, 0, 0, 0, time.FixedZone("CEST", 2*60*60))
}

func nodeByID(t *testing.T, payload graphexport.Payload, id string) graphexport.Node {
	t.Helper()
	for _, node := range payload.Nodes {
		if node.ID == id {
			return node
		}
	}
	t.Fatalf("node %q not found in %+v", id, payload.Nodes)
	return graphexport.Node{}
}

func assertNodeClosure(t *testing.T, payload graphexport.Payload) {
	t.Helper()
	nodes := map[string]bool{}
	for _, node := range payload.Nodes {
		nodes[node.ID] = true
	}
	for _, edge := range payload.Edges {
		if !nodes[edge.Source] {
			t.Fatalf("edge source %q has no node", edge.Source)
		}
		if !nodes[edge.Target] {
			t.Fatalf("edge target %q has no node", edge.Target)
		}
	}
}

func assertEdge(
	t *testing.T,
	payload graphexport.Payload,
	source string,
	target string,
	edgeType string,
	provenance string,
	count int,
) {
	t.Helper()
	for _, edge := range payload.Edges {
		if edge.Source == source && edge.Target == target && edge.Type == edgeType && edge.Provenance == provenance {
			if edge.OccurrenceCount != count {
				t.Fatalf("edge %+v occurrence_count = %d, want %d", edge, edge.OccurrenceCount, count)
			}
			return
		}
	}
	t.Fatalf("edge %s -> %s type=%s provenance=%s not found in %+v", source, target, edgeType, provenance, payload.Edges)
}

func hasNode(payload graphexport.Payload, id string) bool {
	for _, node := range payload.Nodes {
		if node.ID == id {
			return true
		}
	}
	return false
}

func hasEdge(payload graphexport.Payload, source string, target string, edgeType string, provenance string) bool {
	for _, edge := range payload.Edges {
		if edge.Source == source && edge.Target == target && edge.Type == edgeType && edge.Provenance == provenance {
			return true
		}
	}
	return false
}

func marshalNode(t *testing.T, node graphexport.Node) map[string]any {
	t.Helper()
	raw, err := json.Marshal(node)
	if err != nil {
		t.Fatalf("json.Marshal(Node) error = %v", err)
	}
	var object map[string]any
	if err := json.Unmarshal(raw, &object); err != nil {
		t.Fatalf("json.Unmarshal(Node) error = %v", err)
	}
	return object
}
