package install_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBranchInstallBuildsRequestedBranchAndInstallsSourceExecutable(t *testing.T) {
	t.Parallel()

	harness := newBranchInstallHarness(t)
	branch := "branch-install-test"
	harness.createSourceRepo(t, branch)
	sourceRevision := harness.commitBranchMarker(t, branch)
	harness.addReleaseFallbackArtifact(t, "v9.9.9")

	result := harness.run(t, map[string]string{
		"BRANCH":  branch,
		"VERSION": "v9.9.9",
	})

	result.requireSuccess(t)
	result.requireBranchSuccessOutput(t, harness.installedPath(), branch, sourceRevision)
	harness.requireInstalledSourceBuild(t, branch, sourceRevision)
	harness.requireSourceTmpClean(t)
}

func TestBranchInstallRequiresGoBeforeClaimingSuccess(t *testing.T) {
	t.Parallel()

	harness := newBranchInstallHarness(t)

	result := harness.run(t, map[string]string{
		"BRANCH": "branch-install-test",
		"PATH":   filepath.Join(t.TempDir(), "empty-path"),
	})

	result.requireFailure(t)
	result.requireDiagnostic(t, "Go is required")
	harness.requireNotInstalled(t)
}

func TestBranchInstallRequiresMakeBeforeClaimingSuccess(t *testing.T) {
	t.Parallel()

	harness := newBranchInstallHarness(t)
	goShim := harness.writeTool(t, "go", "#!/bin/sh\nexit 0\n")

	result := harness.run(t, map[string]string{
		"BRANCH": "branch-install-test",
		"PATH":   filepath.Dir(goShim),
	})

	result.requireFailure(t)
	result.requireDiagnostic(t, "make is required")
	harness.requireNotInstalled(t)
}

func TestBranchInstallUnavailableBranchFailsWithoutReleaseFallback(t *testing.T) {
	t.Parallel()

	harness := newBranchInstallHarness(t)
	harness.createSourceRepo(t, "available-branch")
	harness.addReleaseFallbackArtifact(t, "v9.9.9")

	result := harness.run(t, map[string]string{
		"BRANCH":  "missing-branch",
		"VERSION": "v9.9.9",
	})

	result.requireFailure(t)
	result.requireDiagnostic(t, "BRANCH=missing-branch")
	harness.requireNotInstalled(t)
}

func TestBranchInstallBuildFailureFailsWithoutSuccessClaim(t *testing.T) {
	t.Parallel()

	harness := newBranchInstallHarness(t)
	branch := "build-fails"
	harness.createSourceRepo(t, branch)
	harness.replaceMakefile(t, branch, "#!/usr/bin/make -f\ninstall:\n\t@printf 'controlled build failure\\n' >&2\n\t@exit 73\n")

	result := harness.run(t, map[string]string{
		"BRANCH": branch,
	})

	result.requireFailure(t)
	result.requireDiagnostic(t, "source install failed")
	result.requireDiagnostic(t, "BRANCH="+branch)
	harness.requireNotInstalled(t)
}

func TestBranchInstallInstallFailureFailsWithoutSuccessClaim(t *testing.T) {
	t.Parallel()

	harness := newBranchInstallHarness(t)
	branch := "install-dir-fails"
	harness.createSourceRepo(t, branch)
	harness.installDir = filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(harness.installDir, []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("create unusable install dir sentinel: %v", err)
	}

	result := harness.run(t, map[string]string{
		"BRANCH": branch,
	})

	result.requireFailure(t)
	result.requireDiagnostic(t, "source install failed")
	result.requireNoSuccessOrVerificationClaim(t)
}

type branchInstallHarness struct {
	repoRoot      string
	sourceRepo    string
	installDir    string
	sourceTmpRoot string
	cacheDir      string
	toolDir       string
	gitPath       string
}

func newBranchInstallHarness(t *testing.T) *branchInstallHarness {
	t.Helper()

	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}
	gitPath, err := exec.LookPath("git")
	if err != nil {
		t.Fatalf("git is required to validate branch install: %v", err)
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Fatalf("go is required to validate branch install: %v", err)
	}
	if _, err := exec.LookPath("make"); err != nil {
		t.Fatalf("make is required to validate branch install: %v", err)
	}

	root := t.TempDir()
	toolDir := filepath.Join(root, "tools")
	for _, dir := range []string{toolDir, filepath.Join(root, "source-tmp")} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("create test directory %s: %v", dir, err)
		}
	}

	return &branchInstallHarness{
		repoRoot:      repoRoot,
		sourceRepo:    filepath.Join(root, "source-repo"),
		installDir:    filepath.Join(root, "install"),
		sourceTmpRoot: filepath.Join(root, "source-tmp"),
		cacheDir:      filepath.Join(root, "go-cache"),
		toolDir:       toolDir,
		gitPath:       gitPath,
	}
}

func (h *branchInstallHarness) createSourceRepo(t *testing.T, branch string) {
	t.Helper()

	runCommand(t, h.gitPath, "clone", "--quiet", h.repoRoot, h.sourceRepo)
	runCommand(t, h.gitPath, "-C", h.sourceRepo, "checkout", "--quiet", "-B", branch)
}

func (h *branchInstallHarness) commitBranchMarker(t *testing.T, branch string) string {
	t.Helper()

	runCommand(t, h.gitPath, "-C", h.sourceRepo, "checkout", "--quiet", branch)
	markerPath := filepath.Join(h.sourceRepo, "branch-source-marker.txt")
	if err := os.WriteFile(markerPath, []byte("controlled source branch: "+branch+"\n"), 0o644); err != nil {
		t.Fatalf("write controlled branch marker: %v", err)
	}
	runCommand(t, h.gitPath, "-C", h.sourceRepo, "add", "branch-source-marker.txt")
	runCommand(t, h.gitPath, "-C", h.sourceRepo, "-c", "user.name=Branch Test", "-c", "user.email=branch-test@example.invalid", "commit", "--quiet", "-m", "test: controlled branch source")

	cmd := exec.Command(h.gitPath, "-C", h.sourceRepo, "rev-parse", "--short", "HEAD")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("read controlled branch revision: %v\nstdout:\n%s\nstderr:\n%s", err, stdout.String(), stderr.String())
	}
	return strings.TrimSpace(stdout.String())
}

func (h *branchInstallHarness) replaceMakefile(t *testing.T, branch, content string) {
	t.Helper()

	runCommand(t, h.gitPath, "-C", h.sourceRepo, "checkout", "--quiet", branch)
	if err := os.WriteFile(filepath.Join(h.sourceRepo, "Makefile"), []byte(content), 0o644); err != nil {
		t.Fatalf("write controlled Makefile: %v", err)
	}
	runCommand(t, h.gitPath, "-C", h.sourceRepo, "add", "Makefile")
	runCommand(t, h.gitPath, "-C", h.sourceRepo, "-c", "user.name=Branch Test", "-c", "user.email=branch-test@example.invalid", "commit", "--quiet", "-m", "test: controlled makefile")
}

func (h *branchInstallHarness) addReleaseFallbackArtifact(t *testing.T, version string) {
	t.Helper()

	artifact := filepath.Join(h.toolDir, "release-scip-search")
	content := "#!/bin/sh\nif [ \"$1\" = \"--version\" ]; then\nprintf 'scip-search release " + version + "\\n'\nexit 0\nfi\nexit 64\n"
	if err := os.WriteFile(artifact, []byte(content), 0o755); err != nil {
		t.Fatalf("write release fallback artifact: %v", err)
	}
	metadata := strings.Join([]string{version, "2026-01-01T00:00:00Z", "linux", "amd64", artifact}, "\t") + "\n"
	if err := os.WriteFile(filepath.Join(h.toolDir, "releases.tsv"), []byte(metadata), 0o644); err != nil {
		t.Fatalf("write release fallback metadata: %v", err)
	}
}

func (h *branchInstallHarness) writeTool(t *testing.T, name, content string) string {
	t.Helper()

	path := filepath.Join(h.toolDir, name)
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("write tool shim %s: %v", path, err)
	}
	return path
}

func (h *branchInstallHarness) run(t *testing.T, env map[string]string) commandResult {
	t.Helper()

	cmd := exec.Command("bash", installerPath)
	cmd.Dir = filepath.Join("..", "..", "tests", "install")
	cmd.Env = append(cleanInstallerEnv(),
		"HOME="+filepath.Join(t.TempDir(), "home"),
		"GOPATH="+filepath.Join(h.cacheDir, "gopath"),
		"GOMODCACHE="+filepath.Join(h.cacheDir, "gopath", "pkg", "mod"),
		"GOCACHE="+filepath.Join(h.cacheDir, "build"),
		"GOFLAGS=-modcacherw",
		"INSTALL_DIR="+h.installDir,
		"PATH="+h.toolDir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"SCIP_SEARCH_INSTALL_OS=linux",
		"SCIP_SEARCH_INSTALL_ARCH=amd64",
		"SCIP_SEARCH_RELEASES_FILE="+filepath.Join(h.toolDir, "releases.tsv"),
		"SCIP_SEARCH_SOURCE_REPO="+h.sourceRepo,
		"SCIP_SEARCH_SOURCE_TMPDIR="+h.sourceTmpRoot,
	)
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

func (h *branchInstallHarness) installedPath() string {
	return filepath.Join(h.installDir, "scip-search")
}

func (h *branchInstallHarness) requireInstalledSourceBuild(t *testing.T, branch, sourceRevision string) {
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
		t.Fatalf("%s --version failed outside source checkout: %v\nstdout:\n%s\nstderr:\n%s", h.installedPath(), err, stdout.String(), stderr.String())
	}
	for _, want := range []string{"scip-search", "source", "branch:" + branch, sourceRevision} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("--version output = %q, want substring %q", stdout.String(), want)
		}
	}
	if strings.Contains(stdout.String(), "release") {
		t.Fatalf("--version output = %q, branch source install must not claim release provenance", stdout.String())
	}
	if stderr.Len() > 0 {
		t.Fatalf("--version wrote stderr outside source checkout:\n%s", stderr.String())
	}
}

func (h *branchInstallHarness) requireNotInstalled(t *testing.T) {
	t.Helper()

	if _, err := os.Stat(h.installedPath()); err == nil {
		t.Fatalf("expected no installed scip-search at %s", h.installedPath())
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat installed scip-search: %v", err)
	}
}

func (h *branchInstallHarness) requireSourceTmpClean(t *testing.T) {
	t.Helper()

	entries, err := os.ReadDir(h.sourceTmpRoot)
	if err != nil {
		t.Fatalf("read source temp root: %v", err)
	}
	if len(entries) != 0 {
		var names []string
		for _, entry := range entries {
			names = append(names, entry.Name())
		}
		t.Fatalf("source temporary checkout was not cleaned up: %s", strings.Join(names, ", "))
	}
}

func (r commandResult) requireBranchSuccessOutput(t *testing.T, installedPath, branch, sourceRevision string) {
	t.Helper()

	if !strings.Contains(r.stdout, installedPath) {
		t.Fatalf("success output did not identify installed path %q:\nstdout:\n%s", installedPath, r.stdout)
	}
	if !strings.Contains(r.stdout, "source") || !strings.Contains(r.stdout, branch) {
		t.Fatalf("success output did not identify branch source provenance for %q:\nstdout:\n%s", branch, r.stdout)
	}
	if !strings.Contains(r.stdout, sourceRevision) {
		t.Fatalf("success output did not identify controlled source revision %q:\nstdout:\n%s", sourceRevision, r.stdout)
	}
}

func runCommand(t *testing.T, name string, args ...string) {
	t.Helper()

	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("%s %s failed: %v\nstdout:\n%s\nstderr:\n%s", name, strings.Join(args, " "), err, stdout.String(), stderr.String())
	}
}
