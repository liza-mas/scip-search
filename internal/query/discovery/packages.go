package discovery

import (
	"sort"
	"strings"

	"scip-search/internal/scipmodel"
	"scip-search/internal/traversal"
)

type PackagesPayload struct {
	Packages []PackageResult `json:"packages"`
}

type PackageResult struct {
	Scheme         string `json:"scheme"`
	PackageManager string `json:"packageManager"`
	PackageName    string `json:"packageName"`
	PackageVersion string `json:"packageVersion"`
	PackageKey     string `json:"packageKey"`
}

func Packages(view traversal.View, prefix string) (PackagesPayload, error) {
	byPackageKey := map[string]PackageResult{}

	for _, symbol := range symbolsInView(view) {
		identity, err := scipmodel.ParseIdentity(symbol.Symbol)
		if err != nil {
			return PackagesPayload{}, err
		}
		if prefix != "" && !strings.HasPrefix(identity.PackageName, prefix) {
			continue
		}

		packageKey := identity.PackageKey()
		if _, exists := byPackageKey[packageKey]; exists {
			continue
		}
		byPackageKey[packageKey] = PackageResult{
			Scheme:         identity.Scheme,
			PackageManager: identity.PackageManager,
			PackageName:    identity.PackageName,
			PackageVersion: identity.PackageVersion,
			PackageKey:     packageKey,
		}
	}

	results := make([]PackageResult, 0, len(byPackageKey))
	for _, result := range byPackageKey {
		results = append(results, result)
	}
	sort.SliceStable(results, func(left, right int) bool {
		return results[left].PackageKey < results[right].PackageKey
	})

	return PackagesPayload{Packages: results}, nil
}
