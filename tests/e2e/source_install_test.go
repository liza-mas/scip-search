package e2e_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestSourceBranchInstallEndToEndUsesControlledSourceRepository(t *testing.T) {
	t.Parallel()

	harness := newSourceBranchE2EHarness(t)
	branch := "source-e2e-branch"
	wantRevision := harness.createControlledBranch(t, branch)
	harness.addReleaseFallbackArtifact(t, "v9.9.9")

	result := harness.runInstaller(t, map[string]string{
		"BRANCH":  branch,
		"VERSION": "v9.9.9",
	})

	result.requireSuccess(t)
	result.requireSourceBranchSuccessOutput(t, harness.installedPath(), branch, wantRevision)
	harness.requireExecutableSourceBranch(t, branch, wantRevision)
	harness.requireNoReleaseFallback(t)
}

func TestSourceBranchInstallEndToEndUnavailableBranchIsActionable(t *testing.T) {
	t.Parallel()

	harness := newSourceBranchE2EHarness(t)
	harness.createControlledBranch(t, "available-source-branch")
	harness.addReleaseFallbackArtifact(t, "v9.9.9")

	result := harness.runInstaller(t, map[string]string{
		"BRANCH":  "missing-source-branch",
		"VERSION": "v9.9.9",
	})

	result.requireFailure(t)
	result.requireDiagnostic(t, "BRANCH=missing-source-branch")
	harness.requireNoInstalledBinary(t)
	harness.requireNoReleaseFallback(t)
}

func TestSourceBranchInstallEndToEndMissingGoFailsBeforeProvisioning(t *testing.T) {
	t.Parallel()

	harness := newSourceBranchE2EHarness(t)
	harness.addReleaseFallbackArtifact(t, "v9.9.9")

	result := harness.runInstaller(t, map[string]string{
		"BRANCH":  "source-missing-go",
		"PATH":    harness.pathDir,
		"VERSION": "v9.9.9",
	})

	result.requireFailure(t)
	result.requireDiagnostic(t, "Go is required")
	result.requireNoProvisioningOrOutOfScopeDiagnostics(t)
	harness.requireNoInstalledBinary(t)
	harness.requireNoReleaseFallback(t)
}

func TestSourceBranchInstallEndToEndMissingMakeFailsBeforeProvisioning(t *testing.T) {
	t.Parallel()

	harness := newSourceBranchE2EHarness(t)
	harness.addReleaseFallbackArtifact(t, "v9.9.9")
	harness.writeTool(t, "go", "#!/bin/sh\nexit 0\n")

	result := harness.runInstaller(t, map[string]string{
		"BRANCH":  "source-missing-make",
		"PATH":    harness.pathDir,
		"VERSION": "v9.9.9",
	})

	result.requireFailure(t)
	result.requireDiagnostic(t, "make is required")
	result.requireNoProvisioningOrOutOfScopeDiagnostics(t)
	harness.requireNoInstalledBinary(t)
	harness.requireNoReleaseFallback(t)
}

func TestSourceBranchInstallEndToEndBuildFailureDoesNotFallback(t *testing.T) {
	t.Parallel()

	harness := newSourceBranchE2EHarness(t)
	branch := "source-build-fails"
	harness.createControlledBranch(t, branch)
	harness.replaceControlledMakefile(t, branch, "install:\n\t@printf 'controlled source build failure\\n' >&2\n\t@exit 73\n")
	harness.addReleaseFallbackArtifact(t, "v9.9.9")

	result := harness.runInstaller(t, map[string]string{
		"BRANCH":  branch,
		"VERSION": "v9.9.9",
	})

	result.requireFailure(t)
	result.requireDiagnostic(t, "source install failed")
	result.requireDiagnostic(t, "BRANCH="+branch)
	result.requireNoProvisioningOrOutOfScopeDiagnostics(t)
	harness.requireNoInstalledBinary(t)
	harness.requireNoReleaseFallback(t)
}

func TestSourceBranchInstallEndToEndInstallFailureDoesNotFallback(t *testing.T) {
	t.Parallel()

	harness := newSourceBranchE2EHarness(t)
	branch := "source-install-fails"
	harness.createControlledBranch(t, branch)
	unusableInstallDir := filepath.Join(harness.root, "not-a-directory")
	if err := os.WriteFile(unusableInstallDir, []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("create unusable install dir sentinel: %v", err)
	}
	harness.installDir = unusableInstallDir
	harness.addReleaseFallbackArtifact(t, "v9.9.9")

	result := harness.runInstaller(t, map[string]string{
		"BRANCH":  branch,
		"VERSION": "v9.9.9",
	})

	result.requireFailure(t)
	result.requireDiagnostic(t, "source install failed")
	result.requireDiagnostic(t, "BRANCH="+branch)
	result.requireNoProvisioningOrOutOfScopeDiagnostics(t)
	harness.requireNoInstalledBinary(t)
	harness.requireNoReleaseFallback(t)
}

type sourceBranchE2EHarness struct {
	root           string
	repoRoot       string
	sourceRepo     string
	installDir     string
	sourceTmpRoot  string
	pathDir        string
	releaseLog     string
	metadataPath   string
	metadataBuffer bytes.Buffer
	gitPath        string
}

func newSourceBranchE2EHarness(t *testing.T) *sourceBranchE2EHarness {
	t.Helper()

	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}
	gitPath, err := exec.LookPath("git")
	if err != nil {
		t.Fatalf("git is required for source branch e2e validation: %v", err)
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Fatalf("go is required for source branch e2e validation: %v", err)
	}
	if _, err := exec.LookPath("make"); err != nil {
		t.Fatalf("make is required for source branch e2e validation: %v", err)
	}

	root := t.TempDir()
	h := &sourceBranchE2EHarness{
		root:          root,
		repoRoot:      repoRoot,
		sourceRepo:    filepath.Join(root, "source-repo"),
		installDir:    filepath.Join(root, "install"),
		sourceTmpRoot: filepath.Join(root, "source-tmp"),
		pathDir:       filepath.Join(root, "path"),
		releaseLog:    filepath.Join(root, "release-fallback.log"),
		metadataPath:  filepath.Join(root, "releases.tsv"),
		gitPath:       gitPath,
	}

	for _, dir := range []string{h.installDir, h.sourceTmpRoot, h.pathDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("create %s: %v", dir, err)
		}
	}
	for _, name := range []string{"curl", "wget", "scip-go", "scip-typescript", "scip-python", "rust-analyzer", "scip-search"} {
		content := fmt.Sprintf("#!/bin/sh\nprintf 'unexpected tool %%s\\n' %q >> %q\nexit 97\n", name, h.releaseLog)
		if err := os.WriteFile(filepath.Join(h.pathDir, name), []byte(content), 0o755); err != nil {
			t.Fatalf("write forbidden tool shim %s: %v", name, err)
		}
	}

	return h
}

func (h *sourceBranchE2EHarness) createControlledBranch(t *testing.T, branch string) string {
	t.Helper()

	runE2ECommand(t, h.gitPath, "clone", "--quiet", h.repoRoot, h.sourceRepo)
	runE2ECommand(t, h.gitPath, "-C", h.sourceRepo, "checkout", "--quiet", "-B", branch)
	marker := filepath.Join(h.sourceRepo, "controlled-source-branch.txt")
	if err := os.WriteFile(marker, []byte(branch+"\n"), 0o644); err != nil {
		t.Fatalf("write controlled branch marker: %v", err)
	}
	runE2ECommand(t, h.gitPath, "-C", h.sourceRepo, "add", "controlled-source-branch.txt")
	runE2ECommand(t, h.gitPath, "-C", h.sourceRepo, "-c", "user.name=Source E2E", "-c", "user.email=source-e2e@example.invalid", "commit", "--quiet", "-m", "test: controlled source branch")
	return commandOutput(t, h.gitPath, "-C", h.sourceRepo, "rev-parse", "--short", "HEAD")
}

func (h *sourceBranchE2EHarness) replaceControlledMakefile(t *testing.T, branch, content string) {
	t.Helper()

	runE2ECommand(t, h.gitPath, "-C", h.sourceRepo, "checkout", "--quiet", branch)
	if err := os.WriteFile(filepath.Join(h.sourceRepo, "Makefile"), []byte(content), 0o644); err != nil {
		t.Fatalf("write controlled Makefile: %v", err)
	}
	runE2ECommand(t, h.gitPath, "-C", h.sourceRepo, "add", "Makefile")
	runE2ECommand(t, h.gitPath, "-C", h.sourceRepo, "-c", "user.name=Source E2E", "-c", "user.email=source-e2e@example.invalid", "commit", "--quiet", "-m", "test: controlled source failure")
}

func (h *sourceBranchE2EHarness) addReleaseFallbackArtifact(t *testing.T, version string) {
	t.Helper()

	artifact := filepath.Join(h.root, "release-scip-search")
	content := fmt.Sprintf(`#!/bin/sh
printf 'release fallback invoked: %%s\n' "$*" >> %q
if [ "$1" = "--version" ] && [ "$#" -eq 1 ]; then
	printf 'scip-search release %s\n'
	exit 0
fi
exit 64
`, h.releaseLog, version)
	if err := os.WriteFile(artifact, []byte(content), 0o755); err != nil {
		t.Fatalf("write release fallback artifact: %v", err)
	}
	fmt.Fprintf(&h.metadataBuffer, "%s\t2026-01-01T00:00:00Z\tlinux\tamd64\t%s\n", version, artifact)
	if err := os.WriteFile(h.metadataPath, h.metadataBuffer.Bytes(), 0o644); err != nil {
		t.Fatalf("write release metadata: %v", err)
	}
}

func (h *sourceBranchE2EHarness) writeTool(t *testing.T, name, content string) string {
	t.Helper()

	path := filepath.Join(h.pathDir, name)
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("write tool shim %s: %v", path, err)
	}
	return path
}

func (h *sourceBranchE2EHarness) runInstaller(t *testing.T, env map[string]string) installerResult {
	t.Helper()

	scriptPath, err := filepath.Abs(filepath.Join("..", "..", "install.sh"))
	if err != nil {
		t.Fatalf("resolve installer path: %v", err)
	}
	cmd := exec.Command("sh", scriptPath)
	cmd.Dir = filepath.Join(h.root, "outside-source-checkout")
	if err := os.MkdirAll(cmd.Dir, 0o755); err != nil {
		t.Fatalf("create installer working directory: %v", err)
	}
	cmd.Env = append(cleanSourceBranchE2EEnv(),
		"HOME="+filepath.Join(h.root, "home"),
		"GOPATH="+filepath.Join(h.root, "go-cache", "gopath"),
		"GOMODCACHE="+filepath.Join(h.root, "go-cache", "gopath", "pkg", "mod"),
		"GOCACHE="+filepath.Join(h.root, "go-cache", "build"),
		"GOFLAGS=-modcacherw",
		"INSTALL_DIR="+h.installDir,
		"PATH="+h.pathDir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"SCIP_SEARCH_INSTALL_OS=linux",
		"SCIP_SEARCH_INSTALL_ARCH=amd64",
		"SCIP_SEARCH_RELEASES_FILE="+h.metadataPath,
		"SCIP_SEARCH_SOURCE_REPO="+h.sourceRepo,
		"SCIP_SEARCH_SOURCE_TMPDIR="+h.sourceTmpRoot,
	)
	for key, value := range env {
		cmd.Env = append(cmd.Env, key+"="+value)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()

	return installerResult{
		err:    err,
		stdout: stdout.String(),
		stderr: stderr.String(),
	}
}

func cleanSourceBranchE2EEnv() []string {
	blocked := map[string]bool{
		"BRANCH":                    true,
		"GOFLAGS":                   true,
		"GOCACHE":                   true,
		"GOMODCACHE":                true,
		"GOPATH":                    true,
		"HOME":                      true,
		"INSTALL_DIR":               true,
		"SCIP_SEARCH_INSTALL_ARCH":  true,
		"SCIP_SEARCH_INSTALL_OS":    true,
		"SCIP_SEARCH_RELEASES_FILE": true,
		"SCIP_SEARCH_RELEASES_URL":  true,
		"SCIP_SEARCH_SOURCE_REPO":   true,
		"SCIP_SEARCH_SOURCE_TMPDIR": true,
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

func (h *sourceBranchE2EHarness) installedPath() string {
	return filepath.Join(h.installDir, "scip-search")
}

func (h *sourceBranchE2EHarness) requireExecutableSourceBranch(t *testing.T, branch, revision string) {
	t.Helper()

	info, err := os.Stat(h.installedPath())
	if err != nil {
		t.Fatalf("expected installed scip-search at %s: %v", h.installedPath(), err)
	}
	if info.Mode()&0o111 == 0 {
		t.Fatalf("installed scip-search is not executable: mode %s", info.Mode())
	}

	cmd := exec.Command(h.installedPath(), "--version")
	cmd.Dir = filepath.Join(h.root, "verify-outside-source")
	if err := os.MkdirAll(cmd.Dir, 0o755); err != nil {
		t.Fatalf("create verification directory: %v", err)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("%s --version failed outside source checkout: %v\nstdout:\n%s\nstderr:\n%s", h.installedPath(), err, stdout.String(), stderr.String())
	}
	got := stdout.String()
	for _, want := range []string{"scip-search", "source", "branch:" + branch, revision} {
		if !strings.Contains(got, want) {
			t.Fatalf("--version output = %q, want substring %q", got, want)
		}
	}
	if strings.Contains(got, "release") {
		t.Fatalf("--version output = %q, source branch install must not claim release provenance", got)
	}
	if stderr.Len() > 0 {
		t.Fatalf("--version wrote stderr outside source checkout:\n%s", stderr.String())
	}
}

func (h *sourceBranchE2EHarness) requireNoInstalledBinary(t *testing.T) {
	t.Helper()

	if _, err := os.Stat(h.installedPath()); err == nil {
		t.Fatalf("expected no installed scip-search at %s", h.installedPath())
	} else if !os.IsNotExist(err) && !errors.Is(err, syscall.ENOTDIR) {
		t.Fatalf("stat installed scip-search: %v", err)
	}
}

func (h *sourceBranchE2EHarness) requireNoReleaseFallback(t *testing.T) {
	t.Helper()

	if data, err := os.ReadFile(h.releaseLog); err == nil {
		t.Fatalf("branch source install used release artifact or forbidden tool:\n%s", data)
	} else if !os.IsNotExist(err) {
		t.Fatalf("read release fallback log: %v", err)
	}
}

func (r installerResult) requireSourceBranchSuccessOutput(t *testing.T, installedPath, branch, revision string) {
	t.Helper()

	for _, want := range []string{installedPath, "source", branch, revision} {
		if !strings.Contains(r.stdout, want) {
			t.Fatalf("success output missing %q\nstdout:\n%s", want, r.stdout)
		}
	}
	if strings.Contains(r.stdout, "release") {
		t.Fatalf("success output claimed release provenance:\n%s", r.stdout)
	}
}

func (r installerResult) requireNoProvisioningOrOutOfScopeDiagnostics(t *testing.T) {
	t.Helper()

	combined := r.stdout + r.stderr
	for _, forbidden := range []string{
		"apt",
		"apk",
		"brew",
		"dnf",
		"yum",
		"go install",
		"make install",
		"scip-go",
		"scip-typescript",
		"scip-python",
		"rust-analyzer",
		"--index",
		"symbols",
		"references",
		"implementations",
		"packages",
	} {
		if strings.Contains(combined, forbidden) {
			t.Fatalf("failure diagnostic referenced provisioning or out-of-scope workflow %q\nstdout:\n%s\nstderr:\n%s", forbidden, r.stdout, r.stderr)
		}
	}
}

func runE2ECommand(t *testing.T, name string, args ...string) {
	t.Helper()

	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("%s %s failed: %v\nstdout:\n%s\nstderr:\n%s", name, strings.Join(args, " "), err, stdout.String(), stderr.String())
	}
}

func commandOutput(t *testing.T, name string, args ...string) string {
	t.Helper()

	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("%s %s failed: %v\nstdout:\n%s\nstderr:\n%s", name, strings.Join(args, " "), err, stdout.String(), stderr.String())
	}
	return strings.TrimSpace(stdout.String())
}
