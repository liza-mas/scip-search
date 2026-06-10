package aggregate

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/scip-code/scip/bindings/go/scip"
	"google.golang.org/protobuf/proto"

	"scip-search/internal/scipindex"
	"scip-search/internal/version"
)

const ProducerName = "scip-search aggregate-index"

type Pair struct {
	Root      string
	IndexPath string
}

type Options struct {
	ProjectRoot string
	OutPath     string
	Pairs       []Pair
}

type Result struct {
	DocumentCount       int
	ExternalSymbolCount int
}

type ValidationError struct {
	message string
}

func NewValidationError(message string) ValidationError {
	return ValidationError{message: message}
}

func (err ValidationError) Error() string {
	return err.message
}

func IsValidationError(err error) bool {
	var validation ValidationError
	return errors.As(err, &validation)
}

func Run(options Options, buildIdentity version.BuildIdentity) (Result, error) {
	index, result, err := Build(options, buildIdentity)
	if err != nil {
		return Result{}, err
	}
	indexBytes, err := proto.MarshalOptions{Deterministic: true}.Marshal(index)
	if err != nil {
		return Result{}, err
	}
	if err := writeFileAtomically(options.OutPath, indexBytes); err != nil {
		return Result{}, err
	}
	return result, nil
}

func Build(options Options, buildIdentity version.BuildIdentity) (*scip.Index, Result, error) {
	projectRoot, projectRootURI, err := normalizeProjectRoot(options.ProjectRoot)
	if err != nil {
		return nil, Result{}, err
	}
	if strings.TrimSpace(options.OutPath) == "" {
		return nil, Result{}, NewValidationError("missing --out")
	}
	if len(options.Pairs) < 2 {
		return nil, Result{}, NewValidationError("aggregate-index requires at least two --root/--index input pairs")
	}
	if err := rejectOutputInputCollision(options.OutPath, options.Pairs); err != nil {
		return nil, Result{}, err
	}

	loader := scipindex.NewLoader()
	documents := make([]*scip.Document, 0)
	externalSymbols := make([]*scip.SymbolInformation, 0)
	documentPaths := map[string]struct{}{}
	definitionLocations := map[string]string{}
	externalBySymbol := map[string]struct{}{}
	var protocolVersion scip.ProtocolVersion
	var textEncoding scip.TextEncoding
	var indexerFamily string

	for inputIndex, pair := range options.Pairs {
		sourceRoot, err := cleanSourceRoot(pair.Root)
		if err != nil {
			return nil, Result{}, err
		}
		loaded, err := loader.LoadIndex(pair.IndexPath)
		if err != nil {
			return nil, Result{}, err
		}
		index := loaded.Index
		metadata := index.GetMetadata()
		if err := validateRootMapping(projectRoot, sourceRoot, metadata.GetProjectRoot()); err != nil {
			return nil, Result{}, err
		}
		protocolVersion, err = mergeProtocolVersion(protocolVersion, metadata.GetVersion(), inputIndex)
		if err != nil {
			return nil, Result{}, err
		}
		textEncoding, err = mergeTextEncoding(textEncoding, metadata.GetTextDocumentEncoding(), inputIndex)
		if err != nil {
			return nil, Result{}, err
		}
		indexerFamily, err = mergeIndexerFamily(indexerFamily, metadata.GetToolInfo().GetName(), inputIndex)
		if err != nil {
			return nil, Result{}, err
		}

		for _, document := range index.GetDocuments() {
			documentCopy := proto.Clone(document).(*scip.Document)
			aggregatePath, err := aggregateDocumentPath(sourceRoot, document.GetRelativePath())
			if err != nil {
				return nil, Result{}, err
			}
			if _, exists := documentPaths[aggregatePath]; exists {
				return nil, Result{}, NewValidationError(fmt.Sprintf("duplicate aggregate document path %q", aggregatePath))
			}
			documentPaths[aggregatePath] = struct{}{}
			documentCopy.RelativePath = aggregatePath
			documents = append(documents, documentCopy)
			for _, symbolInfo := range documentCopy.GetSymbols() {
				if err := recordDefinitionLocation(definitionLocations, symbolInfo.GetSymbol(), aggregatePath); err != nil {
					return nil, Result{}, err
				}
			}
			for _, occurrence := range documentCopy.GetOccurrences() {
				if occurrence.GetSymbol() == "" || !isDefinition(occurrence) {
					continue
				}
				if err := recordDefinitionLocation(definitionLocations, occurrence.GetSymbol(), aggregatePath); err != nil {
					return nil, Result{}, err
				}
			}
		}

		for _, external := range index.GetExternalSymbols() {
			externalCopy := proto.Clone(external).(*scip.SymbolInformation)
			symbol := externalCopy.GetSymbol()
			if scip.IsLocalSymbol(symbol) {
				externalSymbols = append(externalSymbols, externalCopy)
				continue
			}
			if _, exists := externalBySymbol[symbol]; exists {
				continue
			}
			externalBySymbol[symbol] = struct{}{}
			externalSymbols = append(externalSymbols, externalCopy)
		}
	}

	return &scip.Index{
			Metadata: &scip.Metadata{
				Version:              protocolVersion,
				ProjectRoot:          projectRootURI,
				TextDocumentEncoding: textEncoding,
				ToolInfo: &scip.ToolInfo{
					Name:    ProducerName,
					Version: version.Format(buildIdentity),
				},
			},
			Documents:       documents,
			ExternalSymbols: externalSymbols,
		}, Result{
			DocumentCount:       len(documents),
			ExternalSymbolCount: len(externalSymbols),
		}, nil
}

func normalizeProjectRoot(value string) (string, string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", "", NewValidationError("missing --project-root")
	}
	if strings.HasPrefix(value, "file://") {
		parsed, err := url.Parse(value)
		if err != nil || parsed.Scheme != "file" || parsed.Path == "" {
			return "", "", NewValidationError("--project-root must be an absolute filesystem path or file:// URI")
		}
		cleaned := filepath.Clean(parsed.Path)
		if !filepath.IsAbs(cleaned) {
			return "", "", NewValidationError("--project-root file:// URI must identify an absolute path")
		}
		return cleaned, fileURI(cleaned), nil
	}
	if !filepath.IsAbs(value) {
		return "", "", NewValidationError("--project-root must be absolute or file://")
	}
	cleaned := filepath.Clean(value)
	return cleaned, fileURI(cleaned), nil
}

func cleanSourceRoot(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", NewValidationError("--root values cannot be empty")
	}
	value = filepath.ToSlash(value)
	if strings.HasPrefix(value, "/") {
		return "", NewValidationError("--root values must be relative to --project-root")
	}
	cleaned := path.Clean(value)
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", NewValidationError("--root values must not escape --project-root")
	}
	if cleaned == "/" {
		return "", NewValidationError("--root values must be relative to --project-root")
	}
	if cleaned == "." {
		return ".", nil
	}
	return cleaned, nil
}

func aggregateDocumentPath(root string, documentPath string) (string, error) {
	documentPath = filepath.ToSlash(documentPath)
	if strings.HasPrefix(documentPath, "/") {
		return "", NewValidationError(fmt.Sprintf("document path %q escapes aggregate project root", documentPath))
	}
	joined := documentPath
	if root != "." {
		joined = root + "/" + documentPath
	}
	cleaned := path.Clean(joined)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.HasPrefix(cleaned, "/") {
		return "", NewValidationError(fmt.Sprintf("document path %q escapes aggregate project root", documentPath))
	}
	return cleaned, nil
}

func validateRootMapping(projectRoot string, root string, inputProjectRoot string) error {
	inputRoot, comparable := comparableFileProjectRoot(inputProjectRoot)
	if !comparable {
		return nil
	}
	relative, err := filepath.Rel(projectRoot, inputRoot)
	if err != nil {
		return NewValidationError(fmt.Sprintf("input project root %q is not comparable to aggregate project root", inputProjectRoot))
	}
	relative = filepath.ToSlash(filepath.Clean(relative))
	if relative == ".." || strings.HasPrefix(relative, "../") {
		return NewValidationError(fmt.Sprintf("input project root %q escapes aggregate project root", inputProjectRoot))
	}
	if relative == "" {
		relative = "."
	}
	if relative != root {
		return NewValidationError(fmt.Sprintf("root mapping mismatch: --root %q does not match input project root relative path %q", root, relative))
	}
	return nil
}

func comparableFileProjectRoot(value string) (string, bool) {
	if value == "" || !strings.HasPrefix(value, "file://") {
		return "", false
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "file" || parsed.Path == "" {
		return "", false
	}
	cleaned := filepath.Clean(parsed.Path)
	if !filepath.IsAbs(cleaned) {
		return "", false
	}
	return cleaned, true
}

func mergeProtocolVersion(current scip.ProtocolVersion, next scip.ProtocolVersion, inputIndex int) (scip.ProtocolVersion, error) {
	if next == scip.ProtocolVersion_UnspecifiedProtocolVersion {
		return current, nil
	}
	if current == scip.ProtocolVersion_UnspecifiedProtocolVersion {
		return next, nil
	}
	if current != next {
		return current, NewValidationError(fmt.Sprintf("metadata version mismatch at input %d", inputIndex+1))
	}
	return current, nil
}

func mergeTextEncoding(current scip.TextEncoding, next scip.TextEncoding, inputIndex int) (scip.TextEncoding, error) {
	if next == scip.TextEncoding_UnspecifiedTextEncoding {
		return current, nil
	}
	if current == scip.TextEncoding_UnspecifiedTextEncoding {
		return next, nil
	}
	if current != next {
		return current, NewValidationError(fmt.Sprintf("metadata text_document_encoding mismatch at input %d", inputIndex+1))
	}
	return current, nil
}

func mergeIndexerFamily(current string, toolName string, inputIndex int) (string, error) {
	family := indexerFamily(toolName)
	if family == "" {
		return current, nil
	}
	if current == "" {
		return family, nil
	}
	if current != family {
		return current, NewValidationError(fmt.Sprintf("mixed indexer families: %s and %s at input %d", current, family, inputIndex+1))
	}
	return current, nil
}

func indexerFamily(toolName string) string {
	normalized := strings.ToLower(strings.TrimSpace(toolName))
	switch {
	case normalized == "":
		return ""
	case strings.Contains(normalized, "scip-typescript"):
		return "scip-typescript"
	case strings.Contains(normalized, "scip-python"):
		return "scip-python"
	case strings.Contains(normalized, "scip-go"):
		return "scip-go"
	default:
		return strings.Fields(normalized)[0]
	}
}

func rejectOutputInputCollision(outPath string, pairs []Pair) error {
	outAbs, err := filepath.Abs(outPath)
	if err != nil {
		return NewValidationError(fmt.Sprintf("normalize --out path: %v", err))
	}
	for _, pair := range pairs {
		inputAbs, err := filepath.Abs(pair.IndexPath)
		if err != nil {
			return NewValidationError(fmt.Sprintf("normalize input index path %q: %v", pair.IndexPath, err))
		}
		if outAbs == inputAbs {
			return NewValidationError("--out must not resolve to the same path as an input --index")
		}
	}
	return nil
}

func recordDefinitionLocation(locations map[string]string, symbol string, aggregatePath string) error {
	if symbol == "" {
		return nil
	}
	if scip.IsLocalSymbol(symbol) {
		return nil
	}
	if previous, exists := locations[symbol]; exists && previous != aggregatePath {
		return NewValidationError(fmt.Sprintf("symbol collision for %q between %q and %q", symbol, previous, aggregatePath))
	}
	locations[symbol] = aggregatePath
	return nil
}

func isDefinition(occurrence *scip.Occurrence) bool {
	return occurrence.GetSymbolRoles()&int32(scip.SymbolRole_Definition) != 0
}

func fileURI(pathValue string) string {
	return (&url.URL{Scheme: "file", Path: filepath.ToSlash(filepath.Clean(pathValue))}).String()
}

func writeFileAtomically(outPath string, contents []byte) error {
	directory := filepath.Dir(outPath)
	tempFile, err := os.CreateTemp(directory, ".scip-search-aggregate-*")
	if err != nil {
		return err
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	if _, err := tempFile.Write(contents); err != nil {
		_ = tempFile.Close()
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}
	return os.Rename(tempPath, outPath)
}
