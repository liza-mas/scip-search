package install_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleaseInstallDefaultsToHomeLocalBin(t *testing.T) {
	t.Parallel()

	harness := newReleaseHarness(t)
	harness.addArtifact("v1.2.0", "linux", "amd64", "2026-02-01T00:00:00Z")
	home := filepath.Join(t.TempDir(), "home")

	result := harness.runWithInstallDir(t, "", map[string]string{
		"HOME":                     home,
		"SCIP_SEARCH_INSTALL_OS":   "linux",
		"SCIP_SEARCH_INSTALL_ARCH": "amd64",
	})

	result.requireSuccess(t)
	installedPath := filepath.Join(home, ".local", "bin", "scip-search")
	requireInstalledReleaseAt(t, installedPath, "v1.2.0")
	result.requireSuccessOutput(t, installedPath, "v1.2.0")
	harness.requireNoToolWorkflow(t)
}

func TestReleaseInstallUsesCustomInstallDirAndVerifiesRequestedVersion(t *testing.T) {
	t.Parallel()

	harness := newReleaseHarness(t)
	harness.addArtifact("v1.0.0", "linux", "amd64", "2026-01-01T00:00:00Z")
	harness.addArtifact("v1.2.0", "linux", "amd64", "2026-02-01T00:00:00Z")

	result := harness.run(t, map[string]string{
		"SCIP_SEARCH_INSTALL_OS":   "linux",
		"SCIP_SEARCH_INSTALL_ARCH": "amd64",
		"VERSION":                  "v1.0.0",
	})

	result.requireSuccess(t)
	installedPath := filepath.Join(harness.installDir, "scip-search")
	requireInstalledReleaseAt(t, installedPath, "v1.0.0")
	result.requireSuccessOutput(t, installedPath, "v1.0.0")
	harness.requireNoToolWorkflow(t)
}

func TestUnusableInstallDirFailsWithoutSuccessOrVerificationClaim(t *testing.T) {
	t.Parallel()

	harness := newReleaseHarness(t)
	harness.addArtifact("v1.2.0", "linux", "amd64", "2026-02-01T00:00:00Z")
	unusableInstallDir := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(unusableInstallDir, []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("create unusable install dir sentinel: %v", err)
	}

	result := harness.runWithInstallDir(t, unusableInstallDir, map[string]string{
		"SCIP_SEARCH_INSTALL_OS":   "linux",
		"SCIP_SEARCH_INSTALL_ARCH": "amd64",
	})

	result.requireFailure(t)
	result.requireDiagnostic(t, "INSTALL_DIR")
	result.requireDiagnostic(t, unusableInstallDir)
	harness.requireNotInstalled(t)
	harness.requireNoToolWorkflow(t)
}

func TestInstalledBinaryVersionMismatchFailsWithoutSuccessClaim(t *testing.T) {
	t.Parallel()

	harness := newReleaseHarness(t)
	harness.addArtifactWithVersionOutput("v1.2.0", "linux", "amd64", "2026-02-01T00:00:00Z", "scip-search release v9.9.9", 0)

	result := harness.run(t, map[string]string{
		"SCIP_SEARCH_INSTALL_OS":   "linux",
		"SCIP_SEARCH_INSTALL_ARCH": "amd64",
	})

	result.requireFailure(t)
	result.requireDiagnostic(t, "v1.2.0")
	result.requireDiagnostic(t, filepath.Join(harness.installDir, "scip-search"))
	harness.requireNoToolWorkflow(t)
}

func (r commandResult) requireSuccessOutput(t *testing.T, installedPath, release string) {
	t.Helper()

	if !strings.Contains(r.stdout, installedPath) {
		t.Fatalf("success output did not identify installed path %q:\nstdout:\n%s", installedPath, r.stdout)
	}
	if !strings.Contains(r.stdout, release) {
		t.Fatalf("success output did not identify release %q:\nstdout:\n%s", release, r.stdout)
	}
}
