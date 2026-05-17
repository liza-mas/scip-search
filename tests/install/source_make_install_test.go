package install_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMakeInstallBuildsCurrentCheckoutAndInstallsExecutableSourceBuild(t *testing.T) {
	t.Parallel()

	harness := newMakeInstallHarness(t)

	result := harness.run(t, nil)

	result.requireSuccess(t)
	result.requireSuccessOutput(t, harness.installedPath())
	harness.requireForbiddenReleaseToolsUnused(t)
	harness.requireInstalledSourceBuild(t, "make-install-test", "cafebabe")
}

func TestMakeInstallFailsWhenGoIsMissingWithoutSuccessClaim(t *testing.T) {
	t.Parallel()

	harness := newMakeInstallHarness(t)
	emptyPath := filepath.Join(t.TempDir(), "empty-path")
	if err := os.MkdirAll(emptyPath, 0o755); err != nil {
		t.Fatalf("create empty PATH directory: %v", err)
	}

	result := harness.run(t, map[string]string{
		"PATH": emptyPath,
	})

	result.requireFailure(t)
	result.requireDiagnostic(t, "Go is required")
	result.requireNoSuccessClaim(t)
	harness.requireNotInstalled(t)
}

func TestMakeInstallFailsWhenSourceBuildFailsWithoutSuccessClaim(t *testing.T) {
	t.Parallel()

	harness := newMakeInstallHarness(t)
	fakeGo := harness.writeTool(t, "go", `#!/bin/sh
printf '%s\n' "$*" >> "$SCIP_SEARCH_TEST_TOOL_LOG"
exit 73
`)

	result := harness.run(t, map[string]string{
		"PATH": filepath.Dir(fakeGo) + string(os.PathListSeparator) + os.Getenv("PATH"),
	})

	result.requireFailure(t)
	result.requireDiagnostic(t, "source build failed")
	result.requireNoSuccessClaim(t)
	harness.requireNotInstalled(t)
	harness.requireToolLog(t, "build")
	harness.requireToolLog(t, "./cmd/scip-search")
}

func TestMakeInstallFailsWhenInstallPathIsUnusableWithoutSuccessClaim(t *testing.T) {
	t.Parallel()

	harness := newMakeInstallHarness(t)
	unusableInstallDir := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(unusableInstallDir, []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("create unusable install dir sentinel: %v", err)
	}
	harness.installDir = unusableInstallDir

	result := harness.run(t, nil)

	result.requireFailure(t)
	result.requireDiagnostic(t, "INSTALL_DIR")
	result.requireNoSuccessClaim(t)
}

type makeInstallHarness struct {
	repoRoot    string
	makePath    string
	installDir  string
	buildDir    string
	toolDir     string
	toolLogPath string
}

func newMakeInstallHarness(t *testing.T) *makeInstallHarness {
	t.Helper()

	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}
	makePath, err := exec.LookPath("make")
	if err != nil {
		t.Fatalf("make is required to validate make install: %v", err)
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Fatalf("go is required to validate source make install: %v", err)
	}

	root := t.TempDir()
	toolDir := filepath.Join(root, "tools")
	if err := os.MkdirAll(toolDir, 0o755); err != nil {
		t.Fatalf("create tool dir: %v", err)
	}

	harness := &makeInstallHarness{
		repoRoot:    repoRoot,
		makePath:    makePath,
		installDir:  filepath.Join(root, "install"),
		buildDir:    filepath.Join(root, "build"),
		toolDir:     toolDir,
		toolLogPath: filepath.Join(root, "tool.log"),
	}
	for _, name := range []string{"curl", "wget"} {
		harness.writeTool(t, name, `#!/bin/sh
printf '%s\n' "$0 $*" >> "$SCIP_SEARCH_TEST_TOOL_LOG"
exit 97
`)
	}

	return harness
}

func (h *makeInstallHarness) run(t *testing.T, env map[string]string) makeInstallResult {
	t.Helper()

	cmd := exec.Command(h.makePath, "-C", h.repoRoot, "install",
		"INSTALL_DIR="+h.installDir,
		"BUILD_DIR="+h.buildDir,
		"SOURCE_REF=make-install-test",
		"SOURCE_REVISION=cafebabe",
	)
	cmd.Env = append(cleanMakeInstallEnv(),
		"HOME="+filepath.Join(t.TempDir(), "home"),
		"PATH="+h.toolDir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"SCIP_SEARCH_TEST_TOOL_LOG="+h.toolLogPath,
		"VERSION=v9.9.9",
		"SCIP_SEARCH_RELEASES_FILE="+filepath.Join(t.TempDir(), "releases.tsv"),
	)
	for key, value := range env {
		cmd.Env = append(cmd.Env, key+"="+value)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	return makeInstallResult{
		err:    err,
		stdout: stdout.String(),
		stderr: stderr.String(),
	}
}

func (h *makeInstallHarness) installedPath() string {
	return filepath.Join(h.installDir, "scip-search")
}

func (h *makeInstallHarness) writeTool(t *testing.T, name string, content string) string {
	t.Helper()

	path := filepath.Join(h.toolDir, name)
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("write tool shim %s: %v", path, err)
	}
	return path
}

func (h *makeInstallHarness) requireInstalledSourceBuild(t *testing.T, wantRef, wantRevision string) {
	t.Helper()

	info, err := os.Stat(h.installedPath())
	if err != nil {
		t.Fatalf("expected installed scip-search at %s: %v", h.installedPath(), err)
	}
	if info.Mode()&0o111 == 0 {
		t.Fatalf("installed scip-search is not executable: mode %s", info.Mode())
	}

	cmd := exec.Command(h.installedPath(), "--version")
	cmd.Dir = t.TempDir()
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("installed scip-search --version failed outside clone: %v\nstdout:\n%s\nstderr:\n%s", err, stdout.String(), stderr.String())
	}
	for _, want := range []string{"scip-search", "source", wantRef, wantRevision} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("--version output = %q, want substring %q", stdout.String(), want)
		}
	}
	if strings.Contains(stdout.String(), "release") {
		t.Fatalf("--version output = %q, source install must not claim release provenance", stdout.String())
	}
	if stderr.Len() > 0 {
		t.Fatalf("--version wrote stderr outside clone:\n%s", stderr.String())
	}
}

func (h *makeInstallHarness) requireNotInstalled(t *testing.T) {
	t.Helper()

	if _, err := os.Stat(h.installedPath()); err == nil {
		t.Fatalf("expected no installed scip-search at %s", h.installedPath())
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat installed scip-search: %v", err)
	}
}

func (h *makeInstallHarness) requireForbiddenReleaseToolsUnused(t *testing.T) {
	t.Helper()

	if data, err := os.ReadFile(h.toolLogPath); err == nil {
		t.Fatalf("make install invoked release artifact tooling:\n%s", data)
	} else if !os.IsNotExist(err) {
		t.Fatalf("read release tool log: %v", err)
	}
}

func (h *makeInstallHarness) requireToolLog(t *testing.T, want string) {
	t.Helper()

	data, err := os.ReadFile(h.toolLogPath)
	if err != nil {
		t.Fatalf("read tool log: %v", err)
	}
	if !strings.Contains(string(data), want) {
		t.Fatalf("tool log = %q, want substring %q", string(data), want)
	}
}

type makeInstallResult struct {
	err    error
	stdout string
	stderr string
}

func (r makeInstallResult) requireSuccess(t *testing.T) {
	t.Helper()

	if r.err != nil {
		t.Fatalf("make install failed: %v\nstdout:\n%s\nstderr:\n%s", r.err, r.stdout, r.stderr)
	}
	if !strings.Contains(r.stdout, "Installed scip-search") {
		t.Fatalf("success output did not claim source installation:\nstdout:\n%s", r.stdout)
	}
}

func (r makeInstallResult) requireFailure(t *testing.T) {
	t.Helper()

	if r.err == nil {
		t.Fatalf("make install unexpectedly succeeded\nstdout:\n%s\nstderr:\n%s", r.stdout, r.stderr)
	}
}

func (r makeInstallResult) requireNoSuccessClaim(t *testing.T) {
	t.Helper()

	if strings.Contains(r.stdout, "Installed scip-search") || strings.Contains(r.stderr, "Installed scip-search") {
		t.Fatalf("failure output claimed success\nstdout:\n%s\nstderr:\n%s", r.stdout, r.stderr)
	}
	if strings.Contains(r.stdout, "scip-search --version") || strings.Contains(r.stderr, "scip-search --version") {
		t.Fatalf("failure output instructed verification of a missing binary\nstdout:\n%s\nstderr:\n%s", r.stdout, r.stderr)
	}
}

func (r makeInstallResult) requireDiagnostic(t *testing.T, want string) {
	t.Helper()

	if !strings.Contains(r.stderr, want) {
		t.Fatalf("diagnostic %q missing\nstdout:\n%s\nstderr:\n%s", want, r.stdout, r.stderr)
	}
}

func (r makeInstallResult) requireSuccessOutput(t *testing.T, installedPath string) {
	t.Helper()

	if !strings.Contains(r.stdout, installedPath) {
		t.Fatalf("success output did not identify installed path %q:\nstdout:\n%s", installedPath, r.stdout)
	}
	if !strings.Contains(r.stdout, "source") {
		t.Fatalf("success output did not identify source install:\nstdout:\n%s", r.stdout)
	}
}

func cleanMakeInstallEnv() []string {
	blocked := map[string]bool{
		"BUILD_DIR":                 true,
		"GO":                        true,
		"HOME":                      true,
		"INSTALL_DIR":               true,
		"SCIP_SEARCH_RELEASES_FILE": true,
		"SCIP_SEARCH_RELEASES_URL":  true,
		"SCIP_SEARCH_TEST_TOOL_LOG": true,
		"SOURCE_REF":                true,
		"SOURCE_REVISION":           true,
		"VERSION":                   true,
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

func TestMakeInstallRequiresNoReleaseArtifactLookup(t *testing.T) {
	t.Parallel()

	harness := newMakeInstallHarness(t)
	result := harness.run(t, map[string]string{
		"SCIP_SEARCH_RELEASES_URL": "https://example.invalid/releases.tsv",
	})

	result.requireSuccess(t)
	harness.requireForbiddenReleaseToolsUnused(t)
}
