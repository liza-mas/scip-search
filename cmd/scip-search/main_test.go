package main

import (
	"bytes"
	"testing"

	runtimecontract "scip-search/internal/runtime"
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
