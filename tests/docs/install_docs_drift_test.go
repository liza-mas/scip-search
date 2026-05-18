package docs_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	readmePath            = "README.md"
	releaseValidationPath = "docs/release-validation.md"
	installScriptURL      = "https://raw.githubusercontent.com/liza-mas/scip-search/main/install.sh"
)

func TestInstallDocumentationIncludesSupportedDistributionCommandSurface(t *testing.T) {
	t.Parallel()

	docs := loadDistributionDocs(t)

	for _, requirement := range []docRequirement{
		{
			name:     "README latest release install",
			file:     readmePath,
			contains: []string{"curl -fsSL " + installScriptURL + " | bash", "scip-search --version"},
		},
		{
			name:     "README explicit release install",
			file:     readmePath,
			contains: []string{"curl -fsSL " + installScriptURL + " | VERSION=<release> bash", "scip-search --version"},
		},
		{
			name:     "README branch source install",
			file:     readmePath,
			contains: []string{"curl -fsSL " + installScriptURL + " | BRANCH=<branch> bash", "scip-search --version"},
		},
		{
			name:     "README custom install directory",
			file:     readmePath,
			contains: []string{"curl -fsSL " + installScriptURL + " | INSTALL_DIR=<directory> bash", "<directory>/scip-search --version"},
		},
		{
			name:     "README local clone source install",
			file:     readmePath,
			contains: []string{"git clone https://github.com/liza-mas/scip-search.git", "make install", "scip-search --version"},
		},
		{
			name:     "README source build prerequisites",
			file:     readmePath,
			contains: []string{"caller-provided Go and make"},
		},
		{
			name:     "README separates language indexer prerequisites",
			file:     readmePath,
			contains: []string{"Language indexers are separate tools", "are not installed by the `scip-search` installer"},
		},
		{
			name:     "maintainer latest release validation",
			file:     releaseValidationPath,
			contains: []string{"curl -fsSL " + installScriptURL + " | bash", "scip-search --version"},
		},
		{
			name:     "maintainer explicit release validation",
			file:     releaseValidationPath,
			contains: []string{"curl -fsSL " + installScriptURL + " | VERSION=<release> bash", "scip-search --version"},
		},
		{
			name:     "maintainer custom install directory validation",
			file:     releaseValidationPath,
			contains: []string{"curl -fsSL " + installScriptURL + " | INSTALL_DIR=<directory> bash", "<directory>/scip-search --version"},
		},
		{
			name:     "maintainer branch source validation",
			file:     releaseValidationPath,
			contains: []string{"curl -fsSL " + installScriptURL + " | BRANCH=<branch> bash", "scip-search --version"},
		},
		{
			name:     "maintainer local clone validation",
			file:     releaseValidationPath,
			contains: []string{"git clone https://github.com/liza-mas/scip-search.git", "make install", "scip-search --version"},
		},
		{
			name:     "maintainer distribution-only success proof",
			file:     releaseValidationPath,
			contains: []string{"distribution packaging only", "executable succeeds with `--version`"},
		},
	} {
		requireDocContains(t, docs, requirement)
	}
}

func TestInstallDocumentationRejectsUnsupportedDistributionValidationSurfaces(t *testing.T) {
	t.Parallel()

	docs := loadDistributionDocs(t)

	for _, forbidden := range []forbiddenSurface{
		{
			name:       "package manager installation path",
			substrings: []string{"brew install scip-search", "apt install scip-search", "npm install -g scip-search", "pip install scip-search"},
		},
		{
			name:       "signing or notarization validation",
			substrings: []string{"codesign", "notarytool", "gpg --verify", "cosign verify"},
		},
		{
			name:       "hosted release setup validation",
			substrings: []string{"gh release create", "gh release upload"},
		},
		{
			name:       "query command validation",
			substrings: []string{"scip-search symbols --index", "scip-search references --index", "scip-search implementations --index", "scip-search packages --index"},
		},
		{
			name:       "real language indexer validation",
			substrings: []string{"scip-go -o", "scip-go --module-root", "scip-python", "scip-typescript"},
		},
		{
			name:       "traversal fixture or query golden validation",
			substrings: []string{"go test ./internal/traversal", "go test ./internal/query", "testdata/golden"},
		},
	} {
		requireDistributionDocsForbid(t, docs, forbidden)
	}
}

type distributionDocs struct {
	files map[string]string
}

type docRequirement struct {
	name     string
	file     string
	contains []string
}

type forbiddenSurface struct {
	name       string
	substrings []string
}

func loadDistributionDocs(t *testing.T) distributionDocs {
	t.Helper()

	root := repoRoot(t)
	files := make(map[string]string, 2)
	for _, path := range []string{readmePath, releaseValidationPath} {
		data, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Fatalf("distribution documentation mismatch: read %s: %v", path, err)
		}
		content := string(data)
		if path == readmePath {
			content = markdownSection(t, content, "## Installation")
		}
		files[path] = content
	}

	return distributionDocs{files: files}
}

func repoRoot(t *testing.T) string {
	t.Helper()

	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("distribution documentation mismatch: resolve repo root: %v", err)
	}

	return root
}

func requireDocContains(t *testing.T, docs distributionDocs, requirement docRequirement) {
	t.Helper()

	content := docs.files[requirement.file]
	for _, expected := range requirement.contains {
		if !strings.Contains(content, expected) {
			t.Fatalf("distribution documentation mismatch: %s in %s is missing %q", requirement.name, requirement.file, expected)
		}
	}
}

func requireDistributionDocsForbid(t *testing.T, docs distributionDocs, forbidden forbiddenSurface) {
	t.Helper()

	for path, content := range docs.files {
		for _, substring := range forbidden.substrings {
			if strings.Contains(content, substring) {
				t.Fatalf("distribution documentation mismatch: %s contains unsupported %s surface %q", path, forbidden.name, substring)
			}
		}
	}
}

func markdownSection(t *testing.T, content, heading string) string {
	t.Helper()

	start := strings.Index(content, heading)
	if start == -1 {
		t.Fatalf("distribution documentation mismatch: missing section %q", heading)
	}

	rest := content[start+len(heading):]
	nextHeading := strings.Index(rest, "\n## ")
	if nextHeading == -1 {
		return content[start:]
	}

	return content[start : start+len(heading)+nextHeading]
}
