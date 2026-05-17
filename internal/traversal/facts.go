package traversal

import "github.com/scip-code/scip/bindings/go/scip"

type Metadata struct {
	ProtocolVersion      scip.ProtocolVersion
	ToolName             string
	ToolVersion          string
	ToolArguments        []string
	ProjectRoot          string
	TextDocumentEncoding scip.TextEncoding
}

type SymbolSource string

const (
	SymbolSourceDocument SymbolSource = "document"
	SymbolSourceExternal SymbolSource = "external"
)

type Document struct {
	RelativePath     string
	Language         string
	PositionEncoding scip.PositionEncoding
	Symbols          []Symbol
	Occurrences      []Occurrence
}

type Symbol struct {
	Symbol                 string
	Source                 SymbolSource
	DocumentPath           string
	Kind                   scip.SymbolInformation_Kind
	DisplayName            string
	Documentation          []string
	SignatureDocumentation *scip.Document
	EnclosingSymbol        string
}

type Occurrence struct {
	DocumentPath          string
	DocumentLanguage      string
	PositionEncoding      scip.PositionEncoding
	Range                 []int32
	HasEnclosingRange     bool
	EnclosingRange        []int32
	Symbol                string
	SymbolRoles           int32
	OverrideDocumentation []string
}
