package distribution_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestValidateDistributionTargetRunsDistributionScopedChecks(t *testing.T) {
	t.Parallel()

	harness := newValidationTargetHarness(t)

	result := harness.runMake(t)

	result.requireSuccess(t)
	harness.requireGoTestCommands(t, []string{
		"go test ./tests/docs",
		"go test ./tests/e2e -run ^TestVersionExecutableSmokeReleaseAndSourceBuildsWithoutIndex$",
		"go test ./tests/e2e -run ^TestReleaseInstallerEndToEnd",
		"go test ./tests/e2e -run ^TestSourceBranchInstallEndToEnd",
		"go test ./tests/install -run ^TestMakeInstall",
	})
	result.requireOutputContains(t, []string{
		"distribution validation: docs drift",
		"distribution validation: version smoke",
		"distribution validation: release install packaging",
		"distribution validation: source install packaging",
		"distribution validation: local clone source install",
	})
	harness.requireNoOutOfScopeValidation(t)
}

func TestValidateDistributionTargetLabelsFailingCategory(t *testing.T) {
	t.Parallel()

	harness := newValidationTargetHarness(t)
	harness.writeFailingGoShim(t, "./tests/docs")

	result := harness.runScript(t)

	result.requireFailure(t)
	result.requireOutputContains(t, []string{"distribution validation: docs drift"})
	result.requireErrorContains(t, []string{
		"controlled failure for ./tests/docs",
		"distribution validation failed: docs drift",
	})
	result.requireErrorExcludes(t, []string{
		"query runtime",
		"symbols",
		"references",
		"implementations",
		"packages",
	})
}

type validationTargetHarness struct {
	repoRoot   string
	logPath    string
	toolDir    string
	failReason string
}

type makeResult struct {
	err    error
	stdout string
	stderr string
}

func newValidationTargetHarness(t *testing.T) *validationTargetHarness {
	t.Helper()

	root := t.TempDir()
	harness := &validationTargetHarness{
		repoRoot: repoRoot(t),
		logPath:  filepath.Join(root, "go-test-commands.log"),
		toolDir:  filepath.Join(root, "tools"),
	}
	if err := os.MkdirAll(harness.toolDir, 0o755); err != nil {
		t.Fatalf("create tool dir: %v", err)
	}
	harness.writeGoShim(t)

	return harness
}

func (h *validationTargetHarness) writeGoShim(t *testing.T) {
	t.Helper()

	script := `#!/bin/sh
printf 'go %s\n' "$*" >> "$SCIP_SEARCH_VALIDATE_DISTRIBUTION_LOG"
exit 0
`
	if err := os.WriteFile(filepath.Join(h.toolDir, "go"), []byte(script), 0o755); err != nil {
		t.Fatalf("write go shim: %v", err)
	}
}

func (h *validationTargetHarness) writeFailingGoShim(t *testing.T, reason string) {
	t.Helper()

	if reason == "" {
		t.Fatal("failing go shim reason must not be empty")
	}
	h.failReason = reason

	script := `#!/bin/sh
printf 'go %s\n' "$*" >> "$SCIP_SEARCH_VALIDATE_DISTRIBUTION_LOG"
printf 'controlled failure for %s\n' "$SCIP_SEARCH_VALIDATE_DISTRIBUTION_FAIL_REASON" >&2
exit 73
`
	if err := os.WriteFile(filepath.Join(h.toolDir, "go"), []byte(script), 0o755); err != nil {
		t.Fatalf("write failing go shim: %v", err)
	}
}

func (h *validationTargetHarness) runMake(t *testing.T) makeResult {
	t.Helper()

	cmd := exec.Command("make", "validate-distribution", "GO="+filepath.Join(h.toolDir, "go"))
	cmd.Dir = h.repoRoot
	return h.runCommand(t, cmd)
}

func (h *validationTargetHarness) runScript(t *testing.T) makeResult {
	t.Helper()

	cmd := exec.Command("./scripts/validate-distribution.sh")
	cmd.Dir = h.repoRoot
	cmd.Env = append(cleanDistributionValidationEnv(),
		"GO="+filepath.Join(h.toolDir, "go"),
		"PATH="+h.toolDir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"SCIP_SEARCH_VALIDATE_DISTRIBUTION_LOG="+h.logPath,
		"SCIP_SEARCH_VALIDATE_DISTRIBUTION_FAIL_REASON="+h.failReason,
	)
	return h.runCommand(t, cmd)
}

func (h *validationTargetHarness) runCommand(t *testing.T, cmd *exec.Cmd) makeResult {
	t.Helper()

	env := cmd.Env
	if env == nil {
		env = cleanDistributionValidationEnv()
	}
	cmd.Env = append(env,
		"PATH="+h.toolDir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"SCIP_SEARCH_VALIDATE_DISTRIBUTION_LOG="+h.logPath,
		"SCIP_SEARCH_VALIDATE_DISTRIBUTION_FAIL_REASON="+h.failReason,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	return makeResult{
		err:    err,
		stdout: stdout.String(),
		stderr: stderr.String(),
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current test filename")
	}
	if root, ok := findRepoRoot(filepath.Dir(filename)); ok {
		return root
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("resolve working directory: %v", err)
	}
	if root, ok := findRepoRoot(wd); ok {
		return root
	}
	t.Fatalf("resolve repository root from %s or %s", filename, wd)
	return ""
}

func findRepoRoot(start string) (string, bool) {
	dir := filepath.Clean(start)
	for {
		if _, err := os.Stat(filepath.Join(dir, "Makefile")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "scripts", "validate-distribution.sh")); err == nil {
				return dir, true
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

func cleanDistributionValidationEnv() []string {
	blocked := map[string]bool{
		"GO": true,
		"SCIP_SEARCH_VALIDATE_DISTRIBUTION_FAIL_REASON": true,
		"SCIP_SEARCH_VALIDATE_DISTRIBUTION_LOG":         true,
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

func (h *validationTargetHarness) requireGoTestCommands(t *testing.T, want []string) {
	t.Helper()

	data, err := os.ReadFile(h.logPath)
	if err != nil {
		t.Fatalf("read validation command log: %v", err)
	}
	got := string(data)
	for _, command := range want {
		if !strings.Contains(got, command) {
			t.Fatalf("validation command log missing %q:\n%s", command, got)
		}
	}
}

func (h *validationTargetHarness) requireNoOutOfScopeValidation(t *testing.T) {
	t.Helper()

	data, err := os.ReadFile(h.logPath)
	if err != nil {
		t.Fatalf("read validation command log: %v", err)
	}
	got := string(data)
	for _, forbidden := range []string{
		"./internal/query",
		"./internal/traversal",
		"--index",
		"testdata/golden",
		"scip-go",
		"scip-python",
		"scip-typescript",
	} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("distribution validation invoked out-of-scope surface %q:\n%s", forbidden, got)
		}
	}
}

func (r makeResult) requireSuccess(t *testing.T) {
	t.Helper()

	if r.err != nil {
		t.Fatalf("distribution validation command failed: %v\nstdout:\n%s\nstderr:\n%s", r.err, r.stdout, r.stderr)
	}
}

func (r makeResult) requireFailure(t *testing.T) {
	t.Helper()

	if r.err == nil {
		t.Fatalf("distribution validation command unexpectedly succeeded\nstdout:\n%s\nstderr:\n%s", r.stdout, r.stderr)
	}
}

func (r makeResult) requireOutputContains(t *testing.T, want []string) {
	t.Helper()

	for _, substring := range want {
		if !strings.Contains(r.stdout, substring) {
			t.Fatalf("stdout missing %q:\nstdout:\n%s\nstderr:\n%s", substring, r.stdout, r.stderr)
		}
	}
}

func (r makeResult) requireErrorContains(t *testing.T, want []string) {
	t.Helper()

	for _, substring := range want {
		if !strings.Contains(r.stderr, substring) {
			t.Fatalf("stderr missing %q:\nstdout:\n%s\nstderr:\n%s", substring, r.stdout, r.stderr)
		}
	}
}

func (r makeResult) requireErrorExcludes(t *testing.T, forbidden []string) {
	t.Helper()

	for _, substring := range forbidden {
		if strings.Contains(r.stderr, substring) {
			t.Fatalf("stderr contains out-of-scope diagnostic %q:\nstdout:\n%s\nstderr:\n%s", substring, r.stdout, r.stderr)
		}
	}
}
