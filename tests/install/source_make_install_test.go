package install_test

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestMakeInstallBuildsCurrentCheckoutAndInstallsExecutableSourceBuild(t *testing.T) {
	t.Parallel()

	harness := newMakeInstallHarness(t)

	result := harness.run(t, nil)

	result.requireSuccess(t)
	result.requireSuccessOutput(t, harness.installedPath())
	harness.requireForbiddenToolsUnused(t)
	harness.requireInstalledSourceBuild(t, "make-install-test", "cafebabe")
}

func TestMakeInstallFailsWhenGoIsMissingWithoutSuccessClaim(t *testing.T) {
	t.Parallel()

	harness := newMakeInstallHarness(t)
	harness.linkHostTool(t, harness.toolDir, "make", harness.makePath)

	result := harness.run(t, map[string]string{
		"PATH": harness.toolDir,
	})

	result.requireFailure(t)
	result.requireDiagnostic(t, "Go is required")
	result.requireNoSuccessClaim(t)
	harness.requireNotInstalled(t)
	harness.requireForbiddenToolsUnused(t)
}

func TestMakeInstallFailsWhenMakeIsMissingWithoutSuccessClaim(t *testing.T) {
	t.Parallel()

	harness := newMakeInstallHarness(t)
	goOnlyPath := filepath.Join(t.TempDir(), "go-only-path")
	if err := os.MkdirAll(goOnlyPath, 0o755); err != nil {
		t.Fatalf("create go-only PATH directory: %v", err)
	}
	harness.linkHostTool(t, goOnlyPath, "go", harness.goPath)

	result := harness.run(t, map[string]string{
		"PATH": goOnlyPath,
	})

	result.requireFailure(t)
	result.requireDiagnostic(t, "make")
	result.requireNoSuccessClaim(t)
	harness.requireNotInstalled(t)
	harness.requireForbiddenToolsUnused(t)
}

func TestMakeInstallFailsWhenSourceBuildFailsWithoutSuccessClaim(t *testing.T) {
	t.Parallel()

	harness := newMakeInstallHarness(t)
	harness.linkHostTool(t, harness.toolDir, "make", harness.makePath)
	fakeGo := harness.writeTool(t, "go", `#!/bin/sh
printf '%s\n' "$*" >> "$SCIP_SEARCH_TEST_SOURCE_BUILD_LOG"
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
	harness.requireForbiddenToolsUnused(t)
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
	harness.requireNotInstalled(t)
	harness.requireForbiddenToolsUnused(t)
}

type makeInstallHarness struct {
	sourceDir    string
	shellPath    string
	makePath     string
	goPath       string
	installDir   string
	buildDir     string
	cacheDir     string
	toolDir      string
	toolLogPath  string
	buildLogPath string
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
	goPath, err := exec.LookPath("go")
	if err != nil {
		t.Fatalf("go is required to validate source make install: %v", err)
	}
	gitPath, err := exec.LookPath("git")
	if err != nil {
		t.Fatalf("git is required to create controlled source checkouts: %v", err)
	}
	shellPath, err := exec.LookPath("sh")
	if err != nil {
		t.Fatalf("sh is required to validate the README make install workflow: %v", err)
	}

	root := t.TempDir()
	toolDir := filepath.Join(root, "tools")
	if err := os.MkdirAll(toolDir, 0o755); err != nil {
		t.Fatalf("create tool dir: %v", err)
	}
	sourceDir := filepath.Join(root, "source")

	harness := &makeInstallHarness{
		sourceDir:    sourceDir,
		shellPath:    shellPath,
		makePath:     makePath,
		goPath:       goPath,
		installDir:   filepath.Join(root, "install"),
		buildDir:     filepath.Join(root, "build"),
		cacheDir:     filepath.Join(root, "go-cache"),
		toolDir:      toolDir,
		toolLogPath:  filepath.Join(root, "tool.log"),
		buildLogPath: filepath.Join(root, "source-build.log"),
	}
	runCommand(t, gitPath, "clone", "--quiet", repoRoot, sourceDir)

	for _, name := range []string{"curl", "wget", "scip-go", "scip-typescript", "scip-python", "rust-analyzer", "scip-search"} {
		harness.writeTool(t, name, `#!/bin/sh
printf '%s\n' "$0 $*" >> "$SCIP_SEARCH_TEST_TOOL_LOG"
exit 97
`)
	}

	return harness
}

func (h *makeInstallHarness) run(t *testing.T, env map[string]string) makeInstallResult {
	t.Helper()

	cmd := exec.Command(h.shellPath, "-c", `make -C "$SCIP_SEARCH_TEST_SOURCE_DIR" install INSTALL_DIR="$SCIP_SEARCH_TEST_INSTALL_DIR" BUILD_DIR="$SCIP_SEARCH_TEST_BUILD_DIR" SOURCE_REF=make-install-test SOURCE_REVISION=cafebabe`)
	cmd.Dir = t.TempDir()
	cmd.Env = append(cleanMakeInstallEnv(),
		"HOME="+filepath.Join(t.TempDir(), "home"),
		"GOPATH="+filepath.Join(h.cacheDir, "gopath"),
		"GOMODCACHE="+filepath.Join(h.cacheDir, "gopath", "pkg", "mod"),
		"GOCACHE="+filepath.Join(h.cacheDir, "build"),
		"GOFLAGS=-modcacherw",
		"PATH="+h.toolDir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"SCIP_SEARCH_TEST_SOURCE_DIR="+h.sourceDir,
		"SCIP_SEARCH_TEST_INSTALL_DIR="+h.installDir,
		"SCIP_SEARCH_TEST_BUILD_DIR="+h.buildDir,
		"SCIP_SEARCH_TEST_TOOL_LOG="+h.toolLogPath,
		"SCIP_SEARCH_TEST_SOURCE_BUILD_LOG="+h.buildLogPath,
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

func (h *makeInstallHarness) linkHostTool(t *testing.T, dir, name, hostPath string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.Symlink(hostPath, path); err != nil {
		t.Fatalf("link host tool %s to %s: %v", name, path, err)
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
	} else if !os.IsNotExist(err) && !errors.Is(err, syscall.ENOTDIR) {
		t.Fatalf("stat installed scip-search: %v", err)
	}
}

func (h *makeInstallHarness) requireForbiddenToolsUnused(t *testing.T) {
	t.Helper()

	if data, err := os.ReadFile(h.toolLogPath); err == nil {
		t.Fatalf("make install invoked forbidden release, indexer, or query tooling:\n%s", data)
	} else if !os.IsNotExist(err) {
		t.Fatalf("read forbidden tool log: %v", err)
	}
}

func (h *makeInstallHarness) requireToolLog(t *testing.T, want string) {
	t.Helper()

	data, err := os.ReadFile(h.buildLogPath)
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
		"GOCACHE":                   true,
		"GOFLAGS":                   true,
		"GOMODCACHE":                true,
		"GOPATH":                    true,
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
	harness.requireForbiddenToolsUnused(t)
}
