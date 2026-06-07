package graph

import (
	"cmp"
	"fmt"
	"path"
	"slices"
	"strings"

	"github.com/scip-code/scip/bindings/go/scip"

	"scip-search/internal/query/oneline"
	"scip-search/internal/traversal"
)

const (
	UnavailableNoDefinition     = "definition is not indexed"
	UnavailableNoEnclosingRange = "definition has no enclosing range in this index"
)

type Payload struct {
	Symbol              string         `json:"symbol"`
	Definition          *Location      `json:"definition,omitempty"`
	Incoming            []Occurrence   `json:"incoming"`
	Outgoing            []Occurrence   `json:"outgoing"`
	OutgoingUnavailable string         `json:"outgoingUnavailable,omitempty"`
	Relationships       []Relationship `json:"relationships"`
}

type ImpactPayload struct {
	Symbol              string         `json:"symbol"`
	Definition          *Location      `json:"definition,omitempty"`
	Review              []Occurrence   `json:"review"`
	Dependencies        []Occurrence   `json:"dependencies"`
	OutgoingUnavailable string         `json:"outgoingUnavailable,omitempty"`
	Relationships       []Relationship `json:"relationships"`
	Tests               []TestHint     `json:"tests"`
}

type Location struct {
	DocumentPath string  `json:"documentPath"`
	Range        []int32 `json:"range"`
}

type Occurrence struct {
	Symbol       string  `json:"symbol"`
	DocumentPath string  `json:"documentPath"`
	Range        []int32 `json:"range"`
	Roles        int32   `json:"roles"`
}

type Relationship struct {
	SourceSymbol     string `json:"sourceSymbol"`
	TargetSymbol     string `json:"targetSymbol"`
	Direction        string `json:"direction"`
	IsReference      bool   `json:"isReference,omitempty"`
	IsImplementation bool   `json:"isImplementation,omitempty"`
	IsTypeDefinition bool   `json:"isTypeDefinition,omitempty"`
	IsDefinition     bool   `json:"isDefinition,omitempty"`
}

type TestHint struct {
	Symbol       string   `json:"symbol"`
	DocumentPath string   `json:"documentPath"`
	Range        []int32  `json:"range"`
	Reasons      []string `json:"reasons"`
}

type occurrenceKey struct {
	symbol       string
	documentPath string
	roles        int32
	rangeLength  int
	scipRange    [4]int32
}

type relationshipKey struct {
	source           string
	target           string
	direction        string
	isReference      bool
	isImplementation bool
	isTypeDefinition bool
	isDefinition     bool
}

func Query(view traversal.View, symbol string) Payload {
	definitionOccurrence, hasDefinition := firstDefinition(view, symbol)
	incoming := incomingOccurrences(view, symbol)
	outgoing, unavailable := outgoingOccurrences(view, symbol, definitionOccurrence, hasDefinition)

	payload := Payload{
		Symbol:              symbol,
		Incoming:            nonNilOccurrences(incoming),
		Outgoing:            nonNilOccurrences(outgoing),
		OutgoingUnavailable: unavailable,
		Relationships:       nonNilRelationships(relationships(view, symbol)),
	}
	if hasDefinition {
		payload.Definition = &Location{
			DocumentPath: definitionOccurrence.DocumentPath,
			Range:        slices.Clone(definitionOccurrence.Range),
		}
	}

	return payload
}

func Impact(view traversal.View, symbol string) ImpactPayload {
	payload := Query(view, symbol)
	tests := testHints(view, payload)

	return ImpactPayload{
		Symbol:              payload.Symbol,
		Definition:          cloneLocation(payload.Definition),
		Review:              nonNilOccurrences(payload.Incoming),
		Dependencies:        nonNilOccurrences(payload.Outgoing),
		OutgoingUnavailable: payload.OutgoingUnavailable,
		Relationships:       nonNilRelationships(payload.Relationships),
		Tests:               nonNilTestHints(tests),
	}
}

func Callers(payload Payload) Payload {
	return Payload{
		Symbol:        payload.Symbol,
		Definition:    cloneLocation(payload.Definition),
		Incoming:      nonNilOccurrences(payload.Incoming),
		Relationships: nonNilRelationships(incomingRelationships(payload.Relationships)),
	}
}

func Callees(payload Payload) Payload {
	return Payload{
		Symbol:              payload.Symbol,
		Definition:          cloneLocation(payload.Definition),
		Outgoing:            nonNilOccurrences(payload.Outgoing),
		OutgoingUnavailable: payload.OutgoingUnavailable,
		Relationships:       nonNilRelationships(outgoingRelationships(payload.Relationships)),
	}
}

func OneLine(payload Payload) string {
	var builder strings.Builder
	writeDefinitionLine(&builder, payload.Symbol, payload.Definition)
	for _, occurrence := range payload.Incoming {
		writeOccurrenceLine(&builder, occurrence, "incoming")
	}
	for _, occurrence := range payload.Outgoing {
		writeOccurrenceLine(&builder, occurrence, "outgoing")
	}
	if payload.OutgoingUnavailable != "" {
		fmt.Fprintf(&builder, "?:0:0 symbol=%s; direction=outgoing; unavailable=%s\n",
			oneline.Quote(payload.Symbol),
			oneline.Quote(payload.OutgoingUnavailable),
		)
	}
	for _, relationship := range payload.Relationships {
		writeRelationshipLine(&builder, relationship)
	}
	return builder.String()
}

func ImpactOneLine(payload ImpactPayload) string {
	var builder strings.Builder
	writeDefinitionLine(&builder, payload.Symbol, payload.Definition)
	for _, occurrence := range payload.Review {
		writeOccurrenceLine(&builder, occurrence, "review")
	}
	for _, occurrence := range payload.Dependencies {
		writeOccurrenceLine(&builder, occurrence, "dependency")
	}
	if payload.OutgoingUnavailable != "" {
		fmt.Fprintf(&builder, "?:0:0 symbol=%s; section=dependencies; unavailable=%s\n",
			oneline.Quote(payload.Symbol),
			oneline.Quote(payload.OutgoingUnavailable),
		)
	}
	for _, relationship := range payload.Relationships {
		writeRelationshipLine(&builder, relationship)
	}
	for _, hint := range payload.Tests {
		pathValue, line, column := oneline.Location(hint.DocumentPath, hint.Range)
		fmt.Fprintf(&builder, "%s:%d:%d symbol=%s; section=tests; reasons=%s\n",
			pathValue,
			line,
			column,
			oneline.Quote(hint.Symbol),
			oneline.Quote(strings.Join(hint.Reasons, ",")),
		)
	}
	return builder.String()
}

func Markdown(payload Payload) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "[SCIP Static Graph]\nSymbol: %s\n", payload.Symbol)
	if payload.Definition != nil {
		pathValue, line, column := oneline.Location(payload.Definition.DocumentPath, payload.Definition.Range)
		fmt.Fprintf(&builder, "Defined: %s:%d:%d\n", pathValue, line, column)
	} else {
		builder.WriteString("Defined: unavailable\n")
	}
	writeMarkdownOccurrences(&builder, "Incoming", payload.Incoming)
	if payload.OutgoingUnavailable != "" {
		fmt.Fprintf(&builder, "\nOutgoing:\n- unavailable: %s\n", payload.OutgoingUnavailable)
	} else {
		writeMarkdownOccurrences(&builder, "Outgoing", payload.Outgoing)
	}
	writeMarkdownRelationships(&builder, payload.Relationships)
	return builder.String()
}

func ImpactMarkdown(payload ImpactPayload) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "[SCIP Static Impact]\nSymbol: %s\n", payload.Symbol)
	if payload.Definition != nil {
		pathValue, line, column := oneline.Location(payload.Definition.DocumentPath, payload.Definition.Range)
		fmt.Fprintf(&builder, "Defined: %s:%d:%d\n", pathValue, line, column)
	} else {
		builder.WriteString("Defined: unavailable\n")
	}
	writeMarkdownOccurrences(&builder, "Review", payload.Review)
	if payload.OutgoingUnavailable != "" {
		fmt.Fprintf(&builder, "\nDependencies:\n- unavailable: %s\n", payload.OutgoingUnavailable)
	} else {
		writeMarkdownOccurrences(&builder, "Dependencies", payload.Dependencies)
	}
	writeMarkdownRelationships(&builder, payload.Relationships)
	builder.WriteString("\nTests:\n")
	if len(payload.Tests) == 0 {
		builder.WriteString("- none\n")
	} else {
		for _, hint := range payload.Tests {
			pathValue, line, column := oneline.Location(hint.DocumentPath, hint.Range)
			fmt.Fprintf(&builder, "- %s %s:%d:%d (%s)\n", shortSymbol(hint.Symbol), pathValue, line, column, strings.Join(hint.Reasons, ","))
		}
	}
	return builder.String()
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

func incomingOccurrences(view traversal.View, symbol string) []Occurrence {
	candidates := map[string]struct{}{symbol: {}}
	for _, relationship := range view.RelationshipsOwnedBy(symbol) {
		if relationship.IsReference && relationship.TargetSymbol != "" {
			candidates[relationship.TargetSymbol] = struct{}{}
		}
	}
	for _, relationship := range view.RelationshipsTargeting(symbol) {
		if relationship.IsReference && relationship.SourceSymbol != "" {
			candidates[relationship.SourceSymbol] = struct{}{}
		}
	}

	var results []Occurrence
	seen := map[occurrenceKey]struct{}{}
	for candidate := range candidates {
		for _, occurrence := range view.OccurrencesForSymbol(candidate) {
			if isDefinition(occurrence) {
				continue
			}
			result := occurrenceResult(occurrence)
			key := keyOccurrence(result)
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			results = append(results, result)
		}
	}
	slices.SortFunc(results, compareOccurrences)
	return results
}

func outgoingOccurrences(
	view traversal.View,
	symbol string,
	definition traversal.Occurrence,
	hasDefinition bool,
) ([]Occurrence, string) {
	if !hasDefinition {
		return []Occurrence{}, UnavailableNoDefinition
	}
	if !definition.HasEnclosingRange || len(definition.EnclosingRange) == 0 {
		return []Occurrence{}, UnavailableNoEnclosingRange
	}

	var results []Occurrence
	seen := map[occurrenceKey]struct{}{}
	for _, occurrence := range view.Occurrences() {
		if occurrence.DocumentPath != definition.DocumentPath || occurrence.Symbol == "" || isDefinition(occurrence) {
			continue
		}
		if !rangeWithin(occurrence.Range, definition.EnclosingRange) {
			continue
		}
		result := occurrenceResult(occurrence)
		key := keyOccurrence(result)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		results = append(results, result)
	}
	slices.SortFunc(results, compareOccurrences)
	return results, ""
}

func relationships(view traversal.View, symbol string) []Relationship {
	var results []Relationship
	seen := map[relationshipKey]struct{}{}
	for _, relationship := range view.RelationshipsOwnedBy(symbol) {
		result := relationshipResult(relationship, "outgoing")
		key := keyRelationship(result)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		results = append(results, result)
	}
	for _, relationship := range view.RelationshipsTargeting(symbol) {
		result := relationshipResult(relationship, "incoming")
		key := keyRelationship(result)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		results = append(results, result)
	}
	slices.SortFunc(results, compareRelationships)
	return results
}

func testHints(view traversal.View, payload Payload) []TestHint {
	occurrences := make([]Occurrence, 0)
	for _, occurrence := range view.OccurrencesForSymbol(payload.Symbol) {
		occurrences = append(occurrences, occurrenceResult(occurrence))
	}
	occurrences = append(occurrences, payload.Incoming...)
	occurrences = append(occurrences, payload.Outgoing...)

	hints := map[occurrenceKey]TestHint{}
	for _, occurrence := range occurrences {
		reasons := testReasons(occurrence)
		if len(reasons) == 0 {
			continue
		}
		key := keyOccurrence(occurrence)
		hint, exists := hints[key]
		if !exists {
			hint = TestHint{
				Symbol:       occurrence.Symbol,
				DocumentPath: occurrence.DocumentPath,
				Range:        slices.Clone(occurrence.Range),
			}
		}
		hint.Reasons = appendUniqueStrings(hint.Reasons, reasons...)
		hints[key] = hint
	}

	results := make([]TestHint, 0, len(hints))
	for _, hint := range hints {
		slices.Sort(hint.Reasons)
		results = append(results, hint)
	}
	slices.SortFunc(results, compareTestHints)
	return results
}

func testReasons(occurrence Occurrence) []string {
	var reasons []string
	if occurrence.Roles&int32(scip.SymbolRole_Test) != 0 {
		reasons = append(reasons, "testRole")
	}
	if isTestPath(occurrence.DocumentPath) {
		reasons = append(reasons, "testPath")
	}
	return reasons
}

func isTestPath(documentPath string) bool {
	base := path.Base(documentPath)
	return strings.Contains(documentPath, "/test/") ||
		strings.Contains(documentPath, "/tests/") ||
		strings.HasPrefix(base, "test_") ||
		strings.HasSuffix(base, "_test.go") ||
		strings.HasSuffix(base, "_test.py") ||
		strings.HasSuffix(base, ".test.ts") ||
		strings.HasSuffix(base, ".test.tsx") ||
		strings.HasSuffix(base, ".spec.ts") ||
		strings.HasSuffix(base, ".spec.tsx") ||
		strings.HasSuffix(base, ".test.js") ||
		strings.HasSuffix(base, ".spec.js")
}

func writeDefinitionLine(builder *strings.Builder, symbol string, definition *Location) {
	if definition == nil {
		return
	}
	pathValue, line, column := oneline.Location(definition.DocumentPath, definition.Range)
	fmt.Fprintf(builder, "%s:%d:%d symbol=%s; kind=definition\n", pathValue, line, column, oneline.Quote(symbol))
}

func writeOccurrenceLine(builder *strings.Builder, occurrence Occurrence, direction string) {
	pathValue, line, column := oneline.Location(occurrence.DocumentPath, occurrence.Range)
	fmt.Fprintf(builder, "%s:%d:%d symbol=%s; direction=%s; roles=%d\n",
		pathValue,
		line,
		column,
		oneline.Quote(occurrence.Symbol),
		direction,
		occurrence.Roles,
	)
}

func writeRelationshipLine(builder *strings.Builder, relationship Relationship) {
	fmt.Fprintf(builder, "?:0:0 symbol=%s; relationship=%s; direction=%s\n",
		oneline.Quote(relationshipSymbol(relationship)),
		oneline.Quote(relationshipKinds(relationship)),
		relationship.Direction,
	)
}

func writeMarkdownOccurrences(builder *strings.Builder, title string, occurrences []Occurrence) {
	fmt.Fprintf(builder, "\n%s:\n", title)
	if len(occurrences) == 0 {
		builder.WriteString("- none\n")
		return
	}
	for _, occurrence := range occurrences {
		pathValue, line, column := oneline.Location(occurrence.DocumentPath, occurrence.Range)
		fmt.Fprintf(builder, "- %s %s:%d:%d\n", shortSymbol(occurrence.Symbol), pathValue, line, column)
	}
}

func writeMarkdownRelationships(builder *strings.Builder, relationships []Relationship) {
	builder.WriteString("\nRelationships:\n")
	if len(relationships) == 0 {
		builder.WriteString("- none\n")
		return
	}
	for _, relationship := range relationships {
		fmt.Fprintf(builder, "- %s %s -> %s\n",
			relationship.Direction,
			relationshipKinds(relationship),
			shortSymbol(relationshipSymbol(relationship)),
		)
	}
}

func relationshipKinds(relationship Relationship) string {
	var kinds []string
	if relationship.IsReference {
		kinds = append(kinds, "reference")
	}
	if relationship.IsImplementation {
		kinds = append(kinds, "implementation")
	}
	if relationship.IsTypeDefinition {
		kinds = append(kinds, "type-definition")
	}
	if relationship.IsDefinition {
		kinds = append(kinds, "definition")
	}
	if len(kinds) == 0 {
		return "unknown"
	}
	return strings.Join(kinds, ",")
}

func relationshipSymbol(relationship Relationship) string {
	if relationship.Direction == "incoming" {
		return relationship.SourceSymbol
	}
	return relationship.TargetSymbol
}

func incomingRelationships(relationships []Relationship) []Relationship {
	return filterRelationships(relationships, "incoming")
}

func outgoingRelationships(relationships []Relationship) []Relationship {
	return filterRelationships(relationships, "outgoing")
}

func filterRelationships(relationships []Relationship, direction string) []Relationship {
	results := make([]Relationship, 0)
	for _, relationship := range relationships {
		if relationship.Direction == direction {
			results = append(results, relationship)
		}
	}
	return results
}

func relationshipResult(relationship traversal.Relationship, direction string) Relationship {
	return Relationship{
		SourceSymbol:     relationship.SourceSymbol,
		TargetSymbol:     relationship.TargetSymbol,
		Direction:        direction,
		IsReference:      relationship.IsReference,
		IsImplementation: relationship.IsImplementation,
		IsTypeDefinition: relationship.IsTypeDefinition,
		IsDefinition:     relationship.IsDefinition,
	}
}

func occurrenceResult(occurrence traversal.Occurrence) Occurrence {
	return Occurrence{
		Symbol:       occurrence.Symbol,
		DocumentPath: occurrence.DocumentPath,
		Range:        slices.Clone(occurrence.Range),
		Roles:        occurrence.SymbolRoles,
	}
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

func isDefinition(occurrence traversal.Occurrence) bool {
	return occurrence.SymbolRoles&int32(scip.SymbolRole_Definition) != 0
}

func compareTraversalOccurrences(left traversal.Occurrence, right traversal.Occurrence) int {
	if byDocument := cmp.Compare(left.DocumentPath, right.DocumentPath); byDocument != 0 {
		return byDocument
	}
	return slices.Compare(left.Range, right.Range)
}

func compareOccurrences(left Occurrence, right Occurrence) int {
	if byDocument := cmp.Compare(left.DocumentPath, right.DocumentPath); byDocument != 0 {
		return byDocument
	}
	if byRange := slices.Compare(left.Range, right.Range); byRange != 0 {
		return byRange
	}
	return cmp.Compare(left.Symbol, right.Symbol)
}

func compareRelationships(left Relationship, right Relationship) int {
	if byDirection := cmp.Compare(left.Direction, right.Direction); byDirection != 0 {
		return byDirection
	}
	if bySource := cmp.Compare(left.SourceSymbol, right.SourceSymbol); bySource != 0 {
		return bySource
	}
	if byTarget := cmp.Compare(left.TargetSymbol, right.TargetSymbol); byTarget != 0 {
		return byTarget
	}
	return cmp.Compare(relationshipKinds(left), relationshipKinds(right))
}

func compareTestHints(left TestHint, right TestHint) int {
	if byDocument := cmp.Compare(left.DocumentPath, right.DocumentPath); byDocument != 0 {
		return byDocument
	}
	if byRange := slices.Compare(left.Range, right.Range); byRange != 0 {
		return byRange
	}
	return cmp.Compare(left.Symbol, right.Symbol)
}

func keyOccurrence(occurrence Occurrence) occurrenceKey {
	key := occurrenceKey{
		symbol:       occurrence.Symbol,
		documentPath: occurrence.DocumentPath,
		roles:        occurrence.Roles,
		rangeLength:  len(occurrence.Range),
	}
	for index, value := range occurrence.Range {
		if index >= len(key.scipRange) {
			break
		}
		key.scipRange[index] = value
	}
	return key
}

func keyRelationship(relationship Relationship) relationshipKey {
	return relationshipKey{
		source:           relationship.SourceSymbol,
		target:           relationship.TargetSymbol,
		direction:        relationship.Direction,
		isReference:      relationship.IsReference,
		isImplementation: relationship.IsImplementation,
		isTypeDefinition: relationship.IsTypeDefinition,
		isDefinition:     relationship.IsDefinition,
	}
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

func shortSymbol(symbol string) string {
	if index := strings.LastIndex(symbol, " "); index >= 0 && index+1 < len(symbol) {
		return symbol[index+1:]
	}
	return symbol
}

func cloneLocation(location *Location) *Location {
	if location == nil {
		return nil
	}
	return &Location{
		DocumentPath: location.DocumentPath,
		Range:        slices.Clone(location.Range),
	}
}

func appendUniqueStrings(values []string, additions ...string) []string {
	for _, addition := range additions {
		if !slices.Contains(values, addition) {
			values = append(values, addition)
		}
	}
	return values
}

func nonNilOccurrences(occurrences []Occurrence) []Occurrence {
	if occurrences == nil {
		return []Occurrence{}
	}
	return slices.Clone(occurrences)
}

func nonNilRelationships(relationships []Relationship) []Relationship {
	if relationships == nil {
		return []Relationship{}
	}
	return slices.Clone(relationships)
}

func nonNilTestHints(hints []TestHint) []TestHint {
	if hints == nil {
		return []TestHint{}
	}
	return slices.Clone(hints)
}
