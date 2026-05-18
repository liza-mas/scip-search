package install_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const installerPath = "../../install.sh"

func TestLatestReleaseSelectionChoosesNewestSupportedArtifact(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		goos     string
		goarch   string
		expected string
	}{
		{name: "linux amd64", goos: "linux", goarch: "amd64", expected: "v1.2.0"},
		{name: "darwin arm64", goos: "darwin", goarch: "arm64", expected: "v1.1.0"},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			harness := newReleaseHarness(t)
			harness.addArtifact("v1.0.0", tc.goos, tc.goarch, "2026-01-01T00:00:00Z")
			harness.addArtifact(tc.expected, tc.goos, tc.goarch, "2026-02-01T00:00:00Z")
			harness.addArtifact("v9.9.9", "linux", "arm64", "2026-03-01T00:00:00Z")

			result := harness.run(t, map[string]string{
				"SCIP_SEARCH_INSTALL_OS":   tc.goos,
				"SCIP_SEARCH_INSTALL_ARCH": tc.goarch,
			})

			result.requireSuccess(t)
			harness.requireInstalledRelease(t, tc.expected)
			harness.requireNoToolWorkflow(t)
		})
	}
}

func TestVersionSelectsExactlyRequestedRelease(t *testing.T) {
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
	harness.requireInstalledRelease(t, "v1.0.0")
	if strings.Contains(result.stdout, "v1.2.0") {
		t.Fatalf("explicit VERSION install reported latest release:\nstdout:\n%s", result.stdout)
	}
	harness.requireNoToolWorkflow(t)
}

func TestUnsupportedPlatformFailsWithPlatformDiagnostic(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		goos   string
		goarch string
	}{
		{name: "unsupported os", goos: "freebsd", goarch: "amd64"},
		{name: "unsupported architecture", goos: "linux", goarch: "sparc64"},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			harness := newReleaseHarness(t)
			harness.addArtifact("v1.0.0", "linux", "amd64", "2026-01-01T00:00:00Z")

			result := harness.run(t, map[string]string{
				"SCIP_SEARCH_INSTALL_OS":   tc.goos,
				"SCIP_SEARCH_INSTALL_ARCH": tc.goarch,
			})

			result.requireFailure(t)
			result.requireDiagnostic(t, "platform "+tc.goos+"/"+tc.goarch)
			harness.requireNotInstalled(t)
			harness.requireNoToolWorkflow(t)
		})
	}
}

func TestUnavailableLatestFailsWithPlatformDiagnostic(t *testing.T) {
	t.Parallel()

	harness := newReleaseHarness(t)
	harness.addArtifact("v1.0.0", "darwin", "arm64", "2026-01-01T00:00:00Z")

	result := harness.run(t, map[string]string{
		"SCIP_SEARCH_INSTALL_OS":   "linux",
		"SCIP_SEARCH_INSTALL_ARCH": "amd64",
	})

	result.requireFailure(t)
	result.requireDiagnostic(t, "platform linux/amd64")
	harness.requireNotInstalled(t)
	harness.requireNoToolWorkflow(t)
}

func TestUnavailableVersionFailsWithVersionDiagnostic(t *testing.T) {
	t.Parallel()

	harness := newReleaseHarness(t)
	harness.addArtifact("v1.0.0", "linux", "amd64", "2026-01-01T00:00:00Z")

	result := harness.run(t, map[string]string{
		"SCIP_SEARCH_INSTALL_OS":   "linux",
		"SCIP_SEARCH_INSTALL_ARCH": "amd64",
		"VERSION":                  "v9.9.9",
	})

	result.requireFailure(t)
	result.requireDiagnostic(t, "VERSION=v9.9.9")
	harness.requireNotInstalled(t)
	harness.requireNoToolWorkflow(t)
}

type releaseHarness struct {
	metadataPath string
	artifactDir  string
	installDir   string
	binDir       string
	toolLog      string
	metadata     bytes.Buffer
}

func newReleaseHarness(t *testing.T) *releaseHarness {
	t.Helper()

	root := t.TempDir()
	harness := &releaseHarness{
		metadataPath: filepath.Join(root, "releases.tsv"),
		artifactDir:  filepath.Join(root, "artifacts"),
		installDir:   filepath.Join(root, "bin"),
		binDir:       filepath.Join(root, "path"),
		toolLog:      filepath.Join(root, "tool-workflow.log"),
	}

	for _, dir := range []string{harness.artifactDir, harness.installDir, harness.binDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("create test directory %s: %v", dir, err)
		}
	}

	for _, name := range []string{"go", "make", "git", "scip-go", "scip-typescript", "scip-search"} {
		path := filepath.Join(harness.binDir, name)
		script := fmt.Sprintf("#!/bin/sh\necho %s >> %s\nexit 97\n", name, harness.toolLog)
		if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
			t.Fatalf("write forbidden tool shim %s: %v", name, err)
		}
	}

	t.Cleanup(func() {
		if t.Failed() {
			return
		}
		if _, err := os.Stat(harness.toolLog); err == nil {
			t.Fatalf("installer invoked a forbidden tool workflow; log at %s", harness.toolLog)
		}
	})

	return harness
}

func (h *releaseHarness) addArtifact(version, goos, goarch, publishedAt string) {
	h.addArtifactWithVersionOutput(version, goos, goarch, publishedAt, "scip-search release "+version, 0)
}

func (h *releaseHarness) addArtifactWithVersionOutput(version, goos, goarch, publishedAt, versionOutput string, exitCode int) {
	artifact := filepath.Join(h.artifactDir, fmt.Sprintf("scip-search-%s-%s-%s", version, goos, goarch))
	content := fmt.Sprintf("#!/bin/sh\nif [ \"$1\" = \"--version\" ]; then\n\tprintf '%%s\\n' %q\n\texit %d\nfi\nprintf 'unexpected arguments: %%s\\n' \"$*\" >&2\nexit 64\n", versionOutput, exitCode)
	if err := os.WriteFile(artifact, []byte(content), 0o755); err != nil {
		panic(err)
	}
	fmt.Fprintf(&h.metadata, "%s\t%s\t%s\t%s\t%s\n", version, publishedAt, goos, goarch, artifact)
	if err := os.WriteFile(h.metadataPath, h.metadata.Bytes(), 0o644); err != nil {
		panic(err)
	}
}

func (h *releaseHarness) run(t *testing.T, env map[string]string) commandResult {
	t.Helper()

	return h.runWithInstallDir(t, h.installDir, env)
}

func (h *releaseHarness) runWithInstallDir(t *testing.T, installDir string, env map[string]string) commandResult {
	t.Helper()

	cmd := exec.Command("bash", installerPath)
	cmd.Dir = filepath.Join("..", "..", "tests", "install")
	cmd.Env = append(cleanInstallerEnv(),
		"PATH="+h.binDir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"SCIP_SEARCH_RELEASES_FILE="+h.metadataPath,
	)
	if installDir != "" {
		cmd.Env = append(cmd.Env, "INSTALL_DIR="+installDir)
	}
	for key, value := range env {
		cmd.Env = append(cmd.Env, key+"="+value)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	return commandResult{
		err:    err,
		stdout: stdout.String(),
		stderr: stderr.String(),
	}
}

func cleanInstallerEnv() []string {
	blocked := map[string]bool{
		"BRANCH":                   true,
		"HOME":                     true,
		"INSTALL_DIR":              true,
		"SCIP_SEARCH_INSTALL_ARCH": true,
		"SCIP_SEARCH_INSTALL_OS":   true,
		"SCIP_SEARCH_RELEASES_FILE": true,
		"SCIP_SEARCH_RELEASES_URL": true,
		"VERSION":                  true,
	}
	var env []string
	for _, entry := range os.Environ() {
		key, _, found := strings.Cut(entry, "=")
		if found && blocked[key] {
			continue
		}
		env = append(env, entry)
	}
	return env
}

func (h *releaseHarness) requireInstalledRelease(t *testing.T, version string) {
	t.Helper()

	path := filepath.Join(h.installDir, "scip-search")
	requireInstalledReleaseAt(t, path, version)
}

func (h *releaseHarness) requireNotInstalled(t *testing.T) {
	t.Helper()

	path := filepath.Join(h.installDir, "scip-search")
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("expected no installed scip-search at %s", path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat installed scip-search: %v", err)
	}
}

func (h *releaseHarness) requireNoToolWorkflow(t *testing.T) {
	t.Helper()

	if data, err := os.ReadFile(h.toolLog); err == nil {
		t.Fatalf("installer invoked forbidden workflow tools:\n%s", data)
	} else if !os.IsNotExist(err) {
		t.Fatalf("read forbidden tool log: %v", err)
	}
}

type commandResult struct {
	err    error
	stdout string
	stderr string
}

func (r commandResult) requireSuccess(t *testing.T) {
	t.Helper()

	if r.err != nil {
		t.Fatalf("installer failed: %v\nstdout:\n%s\nstderr:\n%s", r.err, r.stdout, r.stderr)
	}
	if !strings.Contains(r.stdout, "Installed scip-search") {
		t.Fatalf("success output did not identify installation:\nstdout:\n%s", r.stdout)
	}
}

func (r commandResult) requireFailure(t *testing.T) {
	t.Helper()

	if r.err == nil {
		t.Fatalf("installer unexpectedly succeeded\nstdout:\n%s\nstderr:\n%s", r.stdout, r.stderr)
	}
	r.requireNoSuccessOrVerificationClaim(t)
}

func (r commandResult) requireNoSuccessOrVerificationClaim(t *testing.T) {
	t.Helper()

	if strings.Contains(r.stdout, "Installed scip-search") {
		t.Fatalf("failure output claimed success:\nstdout:\n%s", r.stdout)
	}
	if strings.Contains(r.stdout, "scip-search --version") || strings.Contains(r.stderr, "scip-search --version") {
		t.Fatalf("failure output told caller to verify a missing binary\nstdout:\n%s\nstderr:\n%s", r.stdout, r.stderr)
	}
}

func (r commandResult) requireDiagnostic(t *testing.T, want string) {
	t.Helper()

	if !strings.Contains(r.stderr, want) {
		t.Fatalf("diagnostic %q missing\nstdout:\n%s\nstderr:\n%s", want, r.stdout, r.stderr)
	}
}

func requireInstalledReleaseAt(t *testing.T, path, version string) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected installed scip-search at %s: %v", path, err)
	}
	if info.Mode()&0o111 == 0 {
		t.Fatalf("installed scip-search is not executable: mode %s", info.Mode())
	}

	cmd := exec.Command(path, "--version")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("%s --version failed: %v\nstdout:\n%s\nstderr:\n%s", path, err, stdout.String(), stderr.String())
	}
	if got := stdout.String(); !strings.Contains(got, version) {
		t.Fatalf("%s --version output %q does not identify release %s", path, got, version)
	}
}
