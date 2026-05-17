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

func TestReleaseInstallerEndToEndLatestExplicitVersionAndCustomInstallDir(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		env        map[string]string
		installDir string
		want       string
	}{
		{
			name: "latest release installs newest supported artifact",
			want: "v1.2.0",
		},
		{
			name: "explicit VERSION installs requested artifact",
			env: map[string]string{
				"VERSION": "v1.0.0",
			},
			want: "v1.0.0",
		},
		{
			name:       "custom INSTALL_DIR installs directly executable binary",
			installDir: "custom-bin",
			want:       "v1.2.0",
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			harness := newReleaseE2EHarness(t)
			harness.addArtifact("v1.0.0", "linux", "amd64", "2026-01-01T00:00:00Z")
			harness.addArtifact("v1.2.0", "linux", "amd64", "2026-02-01T00:00:00Z")
			harness.addArtifact("v9.9.9", "darwin", "arm64", "2026-03-01T00:00:00Z")

			installDir := ""
			if tc.installDir != "" {
				installDir = filepath.Join(harness.root, tc.installDir)
			}
			result := harness.runInstaller(t, installDir, tc.env)

			result.requireSuccess(t)
			installedPath := harness.defaultInstalledPath()
			if installDir != "" {
				installedPath = filepath.Join(installDir, "scip-search")
			}
			requireExecutableRelease(t, installedPath, tc.want)
			result.requireSuccessOutput(t, installedPath, tc.want)
			harness.requireNoForbiddenWorkflow(t)
			harness.requireNoQueryExecution(t)
		})
	}
}

func TestReleaseInstallerEndToEndFailuresAreActionable(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		setup      func(*releaseE2EHarness)
		env        map[string]string
		installDir func(*releaseE2EHarness) string
		want       []string
	}{
		{
			name: "unsupported platform",
			setup: func(h *releaseE2EHarness) {
				h.addArtifact("v1.2.0", "linux", "amd64", "2026-02-01T00:00:00Z")
			},
			env: map[string]string{
				"SCIP_SEARCH_INSTALL_OS":   "freebsd",
				"SCIP_SEARCH_INSTALL_ARCH": "amd64",
			},
			want: []string{"platform freebsd/amd64"},
		},
		{
			name: "unavailable latest release",
			setup: func(h *releaseE2EHarness) {
				h.addArtifact("v1.2.0", "darwin", "arm64", "2026-02-01T00:00:00Z")
			},
			want: []string{"latest release", "platform linux/amd64"},
		},
		{
			name: "unavailable requested VERSION",
			setup: func(h *releaseE2EHarness) {
				h.addArtifact("v1.2.0", "linux", "amd64", "2026-02-01T00:00:00Z")
			},
			env: map[string]string{
				"VERSION": "v9.9.9",
			},
			want: []string{"VERSION=v9.9.9"},
		},
		{
			name: "unusable INSTALL_DIR",
			setup: func(h *releaseE2EHarness) {
				h.addArtifact("v1.2.0", "linux", "amd64", "2026-02-01T00:00:00Z")
			},
			installDir: func(h *releaseE2EHarness) string {
				path := filepath.Join(h.root, "not-a-directory")
				if err := os.WriteFile(path, []byte("not a directory"), 0o644); err != nil {
					h.t.Fatalf("create unusable install dir sentinel: %v", err)
				}
				return path
			},
			want: []string{"INSTALL_DIR"},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			harness := newReleaseE2EHarness(t)
			tc.setup(harness)
			installDir := ""
			if tc.installDir != nil {
				installDir = tc.installDir(harness)
			}

			result := harness.runInstaller(t, installDir, tc.env)

			result.requireFailure(t)
			for _, diagnostic := range tc.want {
				result.requireDiagnostic(t, diagnostic)
			}
			harness.requireNoInstalledBinary(t)
			harness.requireNoForbiddenWorkflow(t)
			harness.requireNoQueryExecution(t)
		})
	}
}

type releaseE2EHarness struct {
	root            string
	t               *testing.T
	metadataPath    string
	artifactDir     string
	defaultInstall  string
	pathDir         string
	toolLog         string
	queryLog        string
	metadataContent bytes.Buffer
}

type installerResult struct {
	err    error
	stdout string
	stderr string
}

func newReleaseE2EHarness(t *testing.T) *releaseE2EHarness {
	t.Helper()

	root := t.TempDir()
	h := &releaseE2EHarness{
		t:              t,
		root:           root,
		metadataPath:   filepath.Join(root, "releases.tsv"),
		artifactDir:    filepath.Join(root, "artifacts"),
		defaultInstall: filepath.Join(root, "home", ".local", "bin"),
		pathDir:        filepath.Join(root, "path"),
		toolLog:        filepath.Join(root, "forbidden-tools.log"),
		queryLog:       filepath.Join(root, "query-commands.log"),
	}

	for _, dir := range []string{h.artifactDir, h.defaultInstall, h.pathDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("create %s: %v", dir, err)
		}
	}

	for _, name := range []string{"go", "make", "git", "scip-go", "scip-typescript", "scip-python", "scip-search"} {
		path := filepath.Join(h.pathDir, name)
		content := fmt.Sprintf("#!/bin/sh\nprintf '%%s\\n' %q >> %q\nexit 97\n", name, h.toolLog)
		if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
			t.Fatalf("write forbidden tool shim %s: %v", path, err)
		}
	}

	return h
}

func (h *releaseE2EHarness) addArtifact(version, goos, goarch, publishedAt string) {
	artifact := filepath.Join(h.artifactDir, fmt.Sprintf("scip-search-%s-%s-%s", version, goos, goarch))
	content := fmt.Sprintf(`#!/bin/sh
if [ "$1" = "--version" ] && [ "$#" -eq 1 ]; then
	printf 'scip-search release %s\n'
	exit 0
fi
printf 'unexpected invocation: %%s\n' "$*" >> %q
exit 64
`, version, h.queryLog)
	if err := os.WriteFile(artifact, []byte(content), 0o755); err != nil {
		h.t.Fatalf("write release artifact %s: %v", artifact, err)
	}
	fmt.Fprintf(&h.metadataContent, "%s\t%s\t%s\t%s\t%s\n", version, publishedAt, goos, goarch, artifact)
	if err := os.WriteFile(h.metadataPath, h.metadataContent.Bytes(), 0o644); err != nil {
		h.t.Fatalf("write release metadata %s: %v", h.metadataPath, err)
	}
}

func (h *releaseE2EHarness) runInstaller(t *testing.T, installDir string, env map[string]string) installerResult {
	t.Helper()

	scriptPath, err := filepath.Abs(filepath.Join("..", "..", "install.sh"))
	if err != nil {
		t.Fatalf("resolve installer path: %v", err)
	}
	cmd := exec.Command("sh", scriptPath)
	cmd.Dir = filepath.Join(h.root, "run")
	if err := os.MkdirAll(cmd.Dir, 0o755); err != nil {
		t.Fatalf("create installer working directory: %v", err)
	}
	cmd.Env = append(cleanReleaseE2EEnv(),
		"HOME="+filepath.Join(h.root, "home"),
		"PATH="+h.pathDir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"SCIP_SEARCH_INSTALL_OS=linux",
		"SCIP_SEARCH_INSTALL_ARCH=amd64",
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
	err = cmd.Run()

	return installerResult{
		err:    err,
		stdout: stdout.String(),
		stderr: stderr.String(),
	}
}

func cleanReleaseE2EEnv() []string {
	blocked := map[string]bool{
		"BRANCH":                    true,
		"HOME":                      true,
		"INSTALL_DIR":               true,
		"SCIP_SEARCH_INSTALL_ARCH":  true,
		"SCIP_SEARCH_INSTALL_OS":    true,
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

func (h *releaseE2EHarness) defaultInstalledPath() string {
	return filepath.Join(h.defaultInstall, "scip-search")
}

func (h *releaseE2EHarness) requireNoInstalledBinary(t *testing.T) {
	t.Helper()

	path := h.defaultInstalledPath()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("expected no installed binary at %s", path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat installed binary %s: %v", path, err)
	}
}

func (h *releaseE2EHarness) requireNoForbiddenWorkflow(t *testing.T) {
	t.Helper()

	if data, err := os.ReadFile(h.toolLog); err == nil {
		t.Fatalf("installer required forbidden workflow tool:\n%s", data)
	} else if !os.IsNotExist(err) {
		t.Fatalf("read forbidden tool log: %v", err)
	}
}

func (h *releaseE2EHarness) requireNoQueryExecution(t *testing.T) {
	t.Helper()

	if data, err := os.ReadFile(h.queryLog); err == nil {
		t.Fatalf("installer or verification executed a query command:\n%s", data)
	} else if !os.IsNotExist(err) {
		t.Fatalf("read query command log: %v", err)
	}
}

func (r installerResult) requireSuccess(t *testing.T) {
	t.Helper()

	if r.err != nil {
		t.Fatalf("installer failed: %v\nstdout:\n%s\nstderr:\n%s", r.err, r.stdout, r.stderr)
	}
	if !strings.Contains(r.stdout, "Installed scip-search") {
		t.Fatalf("success output did not claim completed install:\nstdout:\n%s", r.stdout)
	}
}

func (r installerResult) requireFailure(t *testing.T) {
	t.Helper()

	if r.err == nil {
		t.Fatalf("installer unexpectedly succeeded\nstdout:\n%s\nstderr:\n%s", r.stdout, r.stderr)
	}
	if strings.Contains(r.stdout, "Installed scip-search") {
		t.Fatalf("failure output claimed success:\nstdout:\n%s", r.stdout)
	}
	if strings.Contains(r.stdout, "scip-search --version") || strings.Contains(r.stderr, "scip-search --version") {
		t.Fatalf("failure output instructed verification of a missing binary\nstdout:\n%s\nstderr:\n%s", r.stdout, r.stderr)
	}
	for _, forbidden := range []string{"go install", "make install", "git clone", "scip-go", "scip-typescript", "scip-python", "--index", "symbols", "references", "implementations", "packages"} {
		if strings.Contains(r.stdout, forbidden) || strings.Contains(r.stderr, forbidden) {
			t.Fatalf("failure diagnostic referenced out-of-scope workflow %q\nstdout:\n%s\nstderr:\n%s", forbidden, r.stdout, r.stderr)
		}
	}
}

func (r installerResult) requireDiagnostic(t *testing.T, want string) {
	t.Helper()

	if !strings.Contains(r.stderr, want) {
		t.Fatalf("diagnostic %q missing\nstdout:\n%s\nstderr:\n%s", want, r.stdout, r.stderr)
	}
}

func (r installerResult) requireSuccessOutput(t *testing.T, installedPath, release string) {
	t.Helper()

	if !strings.Contains(r.stdout, installedPath) {
		t.Fatalf("success output did not identify installed path %q:\nstdout:\n%s", installedPath, r.stdout)
	}
	if !strings.Contains(r.stdout, release) {
		t.Fatalf("success output did not identify release %q:\nstdout:\n%s", release, r.stdout)
	}
}

func requireExecutableRelease(t *testing.T, path, release string) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected installed executable at %s: %v", path, err)
	}
	if info.Mode()&0o111 == 0 {
		t.Fatalf("installed file is not executable: mode %s", info.Mode())
	}

	cmd := exec.Command(path, "--version")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("%s --version failed: %v\nstdout:\n%s\nstderr:\n%s", path, err, stdout.String(), stderr.String())
	}
	if got := stdout.String(); !strings.Contains(got, release) {
		t.Fatalf("%s --version output %q does not identify release %s", path, got, release)
	}
	if stderr.Len() > 0 {
		t.Fatalf("%s --version wrote stderr:\n%s", path, stderr.String())
	}
}
