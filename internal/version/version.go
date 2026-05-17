package version

import (
	"fmt"
	"strings"
)

// ReleaseIdentity, SourceRef, and SourceRevision are build metadata inputs.
// Release/source tooling can set them with Go linker variables.
var (
	ReleaseIdentity string
	SourceRef       = "unknown"
	SourceRevision  = "unknown"
)

// BuildIdentity is the offline identity reported by scip-search --version.
type BuildIdentity struct {
	Release        string
	SourceRef      string
	SourceRevision string
}

// Current returns release metadata when supplied, otherwise source provenance.
func Current() BuildIdentity {
	release := strings.TrimSpace(ReleaseIdentity)
	if release != "" {
		return BuildIdentity{Release: release}
	}

	return BuildIdentity{
		SourceRef:      fallback(SourceRef),
		SourceRevision: fallback(SourceRevision),
	}
}

// Format renders a human-readable build identity outside the query JSON path.
func Format(identity BuildIdentity) string {
	release := strings.TrimSpace(identity.Release)
	if release != "" {
		return fmt.Sprintf("scip-search release %s", release)
	}

	return fmt.Sprintf(
		"scip-search source ref=%s revision=%s",
		fallback(identity.SourceRef),
		fallback(identity.SourceRevision),
	)
}

func fallback(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "unknown"
	}

	return trimmed
}
