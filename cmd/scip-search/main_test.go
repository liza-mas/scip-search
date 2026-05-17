package main

import (
	"bytes"
	"strings"
	"testing"

	runtimecontract "scip-search/internal/runtime"
	"scip-search/internal/version"
)

func TestRunSharedInvocationFailuresUseCLIStreamsAndStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "no command",
			args: nil,
		},
		{
			name: "unsupported command",
			args: []string{"search", "--index", "repo.scip", "--name", "Supervisor"},
		},
		{
			name: "symbols missing index",
			args: []string{"symbols", "--name", "Supervisor"},
		},
		{
			name: "references missing index",
			args: []string{"references", "--symbol", "scip-go gomod example . pkg/Foo#"},
		},
		{
			name: "implementations missing index",
			args: []string{"implementations", "--symbol", "scip-go gomod example . pkg/Foo#"},
		},
		{
			name: "packages missing index",
			args: []string{"packages", "--prefix", "github.com/example"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var stdout bytes.Buffer
			var stderr bytes.Buffer

			status := run(test.args, &stdout, &stderr)

			if status != runtimecontract.StatusUsage {
				t.Fatalf("run(%v) status = %d, want %d", test.args, status, runtimecontract.StatusUsage)
			}
			if stdout.String() != "" {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
			if stderr.String() == "" {
				t.Fatal("stderr is empty, want usage diagnostic")
			}
		})
	}
}

func TestRunVersionUsesBuildIdentityBeforeQueryValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		identity version.BuildIdentity
		want     []string
	}{
		{
			name: "release",
			identity: version.BuildIdentity{
				Release: "v2.0.0",
			},
			want: []string{"scip-search", "release", "v2.0.0"},
		},
		{
			name: "source",
			identity: version.BuildIdentity{
				SourceRef:      "local",
				SourceRevision: "abc999",
			},
			want: []string{"scip-search", "source", "local", "abc999"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var stdout bytes.Buffer
			var stderr bytes.Buffer

			status := runWithBuildIdentity([]string{"--version"}, &stdout, &stderr, test.identity)

			if status != runtimecontract.StatusOK {
				t.Fatalf("runWithBuildIdentity(--version) status = %d, want %d", status, runtimecontract.StatusOK)
			}
			if stdout.String() == "" {
				t.Fatal("stdout is empty, want version output")
			}
			for _, want := range test.want {
				if !strings.Contains(stdout.String(), want) {
					t.Fatalf("stdout = %q, want substring %q", stdout.String(), want)
				}
			}
			if stderr.String() != "" {
				t.Fatalf("stderr = %q, want empty", stderr.String())
			}
			if strings.HasPrefix(strings.TrimSpace(stdout.String()), "{") {
				t.Fatalf("stdout = %q, want non-JSON version output", stdout.String())
			}
		})
	}
}
