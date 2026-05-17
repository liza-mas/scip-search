package traversal

import (
	"slices"

	"github.com/scip-code/scip/bindings/go/scip"
	"google.golang.org/protobuf/proto"

	runtimecontract "scip-search/internal/runtime"
)

type View struct {
	metadata              Metadata
	documents             []Document
	externalSymbols       []Symbol
	occurrences           []Occurrence
	relationshipsByOwner  map[string][]Relationship
	relationshipsByTarget map[string][]Relationship
}

func NewView(loaded runtimecontract.LoadedIndex) View {
	index := loaded.Index
	if index == nil {
		return View{}
	}

	documents := make([]Document, 0, len(index.GetDocuments()))
	var occurrences []Occurrence
	relationshipsByOwner := map[string][]Relationship{}
	relationshipsByTarget := map[string][]Relationship{}
	for _, document := range index.GetDocuments() {
		documentFact := buildDocument(document)
		occurrences = append(occurrences, documentFact.Occurrences...)
		documents = append(documents, documentFact)
		for _, symbol := range document.GetSymbols() {
			appendRelationships(relationshipsByOwner, relationshipsByTarget, symbol)
		}
	}

	externalSymbols := make([]Symbol, 0, len(index.GetExternalSymbols()))
	for _, symbol := range index.GetExternalSymbols() {
		externalSymbols = append(externalSymbols, buildSymbol(symbol, SymbolSourceExternal, ""))
		appendRelationships(relationshipsByOwner, relationshipsByTarget, symbol)
	}

	return View{
		metadata:              buildMetadata(index.GetMetadata()),
		documents:             documents,
		externalSymbols:       externalSymbols,
		occurrences:           occurrences,
		relationshipsByOwner:  relationshipsByOwner,
		relationshipsByTarget: relationshipsByTarget,
	}
}

func (view View) Metadata() Metadata {
	return Metadata{
		ProtocolVersion:      view.metadata.ProtocolVersion,
		ToolName:             view.metadata.ToolName,
		ToolVersion:          view.metadata.ToolVersion,
		ToolArguments:        slices.Clone(view.metadata.ToolArguments),
		ProjectRoot:          view.metadata.ProjectRoot,
		TextDocumentEncoding: view.metadata.TextDocumentEncoding,
	}
}

func (view View) Documents() []Document {
	return cloneDocuments(view.documents)
}

func (view View) ExternalSymbols() []Symbol {
	return cloneSymbols(view.externalSymbols)
}

func (view View) Occurrences() []Occurrence {
	return cloneOccurrences(view.occurrences)
}

func (view View) RelationshipsOwnedBy(sourceSymbol string) []Relationship {
	return cloneRelationships(view.relationshipsByOwner[sourceSymbol])
}

func (view View) RelationshipsTargeting(targetSymbol string) []Relationship {
	return cloneRelationships(view.relationshipsByTarget[targetSymbol])
}

func buildMetadata(metadata *scip.Metadata) Metadata {
	if metadata == nil {
		return Metadata{}
	}

	toolInfo := metadata.GetToolInfo()
	return Metadata{
		ProtocolVersion:      metadata.GetVersion(),
		ToolName:             toolInfo.GetName(),
		ToolVersion:          toolInfo.GetVersion(),
		ToolArguments:        slices.Clone(toolInfo.GetArguments()),
		ProjectRoot:          metadata.GetProjectRoot(),
		TextDocumentEncoding: metadata.GetTextDocumentEncoding(),
	}
}

func buildDocument(document *scip.Document) Document {
	fact := Document{
		RelativePath:     document.GetRelativePath(),
		Language:         document.GetLanguage(),
		PositionEncoding: document.GetPositionEncoding(),
	}

	for _, symbol := range document.GetSymbols() {
		fact.Symbols = append(fact.Symbols, buildSymbol(symbol, SymbolSourceDocument, fact.RelativePath))
	}
	for _, occurrence := range document.GetOccurrences() {
		fact.Occurrences = append(fact.Occurrences, buildOccurrence(occurrence, fact))
	}

	return fact
}

func buildSymbol(symbol *scip.SymbolInformation, source SymbolSource, documentPath string) Symbol {
	return Symbol{
		Symbol:                 symbol.GetSymbol(),
		Source:                 source,
		DocumentPath:           documentPath,
		Kind:                   symbol.GetKind(),
		DisplayName:            symbol.GetDisplayName(),
		Documentation:          slices.Clone(symbol.GetDocumentation()),
		SignatureDocumentation: cloneSignatureDocumentation(symbol.GetSignatureDocumentation()),
		EnclosingSymbol:        symbol.GetEnclosingSymbol(),
	}
}

func buildOccurrence(occurrence *scip.Occurrence, document Document) Occurrence {
	hasEnclosingRange := occurrence != nil && occurrence.EnclosingRange != nil
	enclosingRange := occurrence.GetEnclosingRange()
	return Occurrence{
		DocumentPath:          document.RelativePath,
		DocumentLanguage:      document.Language,
		PositionEncoding:      document.PositionEncoding,
		Range:                 slices.Clone(occurrence.GetRange()),
		HasEnclosingRange:     hasEnclosingRange,
		EnclosingRange:        slices.Clone(enclosingRange),
		Symbol:                occurrence.GetSymbol(),
		SymbolRoles:           occurrence.GetSymbolRoles(),
		OverrideDocumentation: slices.Clone(occurrence.GetOverrideDocumentation()),
	}
}

func appendRelationships(
	relationshipsByOwner map[string][]Relationship,
	relationshipsByTarget map[string][]Relationship,
	symbol *scip.SymbolInformation,
) {
	ownerSymbol := symbol.GetSymbol()
	for _, relationship := range symbol.GetRelationships() {
		fact := Relationship{
			SourceSymbol:     ownerSymbol,
			TargetSymbol:     relationship.GetSymbol(),
			IsReference:      relationship.GetIsReference(),
			IsImplementation: relationship.GetIsImplementation(),
			IsTypeDefinition: relationship.GetIsTypeDefinition(),
			IsDefinition:     relationship.GetIsDefinition(),
		}
		relationshipsByOwner[ownerSymbol] = append(relationshipsByOwner[ownerSymbol], fact)
		relationshipsByTarget[fact.TargetSymbol] = append(relationshipsByTarget[fact.TargetSymbol], fact)
	}
}

func cloneDocuments(documents []Document) []Document {
	if documents == nil {
		return nil
	}

	cloned := make([]Document, len(documents))
	for index, document := range documents {
		cloned[index] = document
		cloned[index].Symbols = cloneSymbols(document.Symbols)
		cloned[index].Occurrences = cloneOccurrences(document.Occurrences)
	}

	return cloned
}

func cloneSymbols(symbols []Symbol) []Symbol {
	if symbols == nil {
		return nil
	}

	cloned := make([]Symbol, len(symbols))
	for index, symbol := range symbols {
		cloned[index] = symbol
		cloned[index].Documentation = slices.Clone(symbol.Documentation)
		cloned[index].SignatureDocumentation = cloneSignatureDocumentation(symbol.SignatureDocumentation)
	}

	return cloned
}

func cloneOccurrences(occurrences []Occurrence) []Occurrence {
	if occurrences == nil {
		return nil
	}

	cloned := make([]Occurrence, len(occurrences))
	for index, occurrence := range occurrences {
		cloned[index] = occurrence
		cloned[index].Range = slices.Clone(occurrence.Range)
		cloned[index].EnclosingRange = slices.Clone(occurrence.EnclosingRange)
		cloned[index].OverrideDocumentation = slices.Clone(occurrence.OverrideDocumentation)
	}

	return cloned
}

func cloneRelationships(relationships []Relationship) []Relationship {
	return slices.Clone(relationships)
}

func cloneSignatureDocumentation(document *scip.Document) *scip.Document {
	if document == nil {
		return nil
	}

	return proto.Clone(document).(*scip.Document)
}
