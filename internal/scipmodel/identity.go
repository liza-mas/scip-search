package scipmodel

import (
	"errors"
	"strings"
)

// Identity exposes exact SCIP symbol identity fields used by discovery queries.
type Identity struct {
	Symbol         string
	Scheme         string
	PackageManager string
	PackageName    string
	PackageVersion string
	Descriptor     string
}

// ParseIdentity splits a full SCIP symbol into package identity and descriptor text.
func ParseIdentity(symbol string) (Identity, error) {
	scheme, rest, ok := cutSymbolComponent(symbol)
	if !ok {
		return Identity{}, errors.New("parse SCIP symbol identity: missing scheme")
	}

	packageManager, rest, ok := cutSymbolComponent(rest)
	if !ok {
		return Identity{}, errors.New("parse SCIP symbol identity: missing package manager")
	}

	packageName, rest, ok := cutSymbolComponent(rest)
	if !ok {
		return Identity{}, errors.New("parse SCIP symbol identity: missing package name")
	}

	packageVersion, descriptor, ok := cutSymbolComponent(rest)
	if !ok {
		return Identity{}, errors.New("parse SCIP symbol identity: missing package version or descriptor")
	}

	return Identity{
		Symbol:         symbol,
		Scheme:         scheme,
		PackageManager: packageManager,
		PackageName:    packageName,
		PackageVersion: packageVersion,
		Descriptor:     descriptor,
	}, nil
}

// PackageKey returns the exact SCIP package prefix without descriptor text.
func (identity Identity) PackageKey() string {
	return strings.Join([]string{
		identity.Scheme,
		identity.PackageManager,
		identity.PackageName,
		identity.PackageVersion,
	}, " ")
}

// MatchText returns displayName when present, otherwise the symbol descriptor.
func (identity Identity) MatchText(displayName string) string {
	if displayName != "" {
		return displayName
	}

	return identity.Descriptor
}

func IsLocalSymbol(symbol string) bool {
	return strings.HasPrefix(symbol, "local ")
}

func cutSymbolComponent(symbol string) (string, string, bool) {
	component, rest, found := strings.Cut(symbol, " ")
	if !found || component == "" || rest == "" {
		return "", "", false
	}

	return component, rest, true
}
