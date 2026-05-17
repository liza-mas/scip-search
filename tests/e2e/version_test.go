package e2e_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersionExecutableSmokeReleaseAndSourceBuildsWithoutIndex(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		ldflags    []string
		want       []string
		forbid     []string
		forbidKind string
	}{
		{
			name: "controlled release identity",
			ldflags: []string{
				"-X", "scip-search/internal/version.ReleaseIdentity=v7.8.9",
				"-X", "scip-search/internal/version.SourceRef=ignored-source-ref",
				"-X", "scip-search/internal/version.SourceRevision=ignored-source-revision",
			},
			want:       []string{"scip-search", "release", "v7.8.9"},
			forbid:     []string{"ignored-source-ref", "ignored-source-revision"},
			forbidKind: "release output exposed source provenance",
		},
		{
			name: "controlled source provenance",
			ldflags: []string{
				"-X", "scip-search/internal/version.SourceRef=feature/version-smoke",
				"-X", "scip-search/internal/version.SourceRevision=abc123def456",
			},
			want:       []string{"scip-search", "source", "feature/version-smoke", "abc123def456"},
			forbid:     []string{"release", "v7.8.9"},
			forbidKind: "source output masqueraded as a release",
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			harness := newVersionSmokeHarness(t)
			executable := harness.buildBinary(t, tc.ldflags)

			result := harness.runVersion(t, executable)

			if result.err != nil {
				t.Fatalf("%s --version failed: %v\nstdout:\n%s\nstderr:\n%s", executable, result.err, result.stdout, result.stderr)
			}
			if result.stdout == "" {
				t.Fatal("stdout is empty, want version identity")
			}
			for _, want := range tc.want {
				if !strings.Contains(result.stdout, want) {
					t.Fatalf("stdout = %q, want substring %q", result.stdout, want)
				}
			}
			for _, forbidden := range tc.forbid {
				if strings.Contains(result.stdout, forbidden) {
					t.Fatalf("%s: stdout = %q contains forbidden substring %q", tc.forbidKind, result.stdout, forbidden)
				}
			}
			if result.stderr != "" {
				t.Fatalf("stderr = %q, want empty", result.stderr)
			}
			harness.requireNoExternalWorkflow(t)
		})
	}
}

type versionSmokeHarness struct {
	root        string
	runDir      string
	pathDir     string
	workflowLog string
}

type versionSmokeResult struct {
	err    error
	stdout string
	stderr string
}

func newVersionSmokeHarness(t *testing.T) *versionSmokeHarness {
	t.Helper()

	root := t.TempDir()
	harness := &versionSmokeHarness{
		root:        root,
		runDir:      filepath.Join(root, "empty-run-dir"),
		pathDir:     filepath.Join(root, "path"),
		workflowLog: filepath.Join(root, "external-workflow.log"),
	}

	for _, dir := range []string{harness.runDir, harness.pathDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("create %s: %v", dir, err)
		}
	}

	for _, name := range []string{
		"bash",
		"curl",
		"git",
		"go",
		"make",
		"scip-go",
		"scip-python",
		"scip-search",
		"scip-typescript",
		"sh",
	} {
		path := filepath.Join(harness.pathDir, name)
		script := fmt.Sprintf("#!/bin/sh\nprintf '%%s\\n' %q >> %q\nexit 97\n", name, harness.workflowLog)
		if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
			t.Fatalf("write forbidden workflow shim %s: %v", path, err)
		}
	}

	return harness
}

func (h *versionSmokeHarness) buildBinary(t *testing.T, ldflags []string) string {
	t.Helper()

	executable := filepath.Join(h.root, "bin", "scip-search")
	if err := os.MkdirAll(filepath.Dir(executable), 0o755); err != nil {
		t.Fatalf("create binary directory: %v", err)
	}

	args := []string{"build", "-o", executable}
	if len(ldflags) > 0 {
		args = append(args, "-ldflags", strings.Join(ldflags, " "))
	}
	args = append(args, "./cmd/scip-search")

	cmd := exec.Command("go", args...)
	cmd.Dir = filepath.Join("..", "..")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("go build failed: %v\nstdout:\n%s\nstderr:\n%s", err, stdout.String(), stderr.String())
	}

	return executable
}

func (h *versionSmokeHarness) runVersion(t *testing.T, executable string) versionSmokeResult {
	t.Helper()

	cmd := exec.Command(executable, "--version")
	cmd.Dir = h.runDir
	cmd.Env = append(cleanVersionSmokeEnv(),
		"PATH="+h.pathDir,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	return versionSmokeResult{
		err:    err,
		stdout: stdout.String(),
		stderr: stderr.String(),
	}
}

func cleanVersionSmokeEnv() []string {
	blocked := map[string]bool{
		"BRANCH":                    true,
		"INSTALL_DIR":               true,
		"SCIP_SEARCH_INDEX":         true,
		"SCIP_SEARCH_RELEASES_FILE": true,
		"SCIP_SEARCH_RELEASES_URL":  true,
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

func (h *versionSmokeHarness) requireNoExternalWorkflow(t *testing.T) {
	t.Helper()

	if data, err := os.ReadFile(h.workflowLog); err == nil {
		t.Fatalf("--version invoked external workflow tools:\n%s", data)
	} else if !os.IsNotExist(err) {
		t.Fatalf("read external workflow log: %v", err)
	}
}
