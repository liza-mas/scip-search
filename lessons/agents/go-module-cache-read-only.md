---
title: "Go module cache read-only fallback"
trigger: "When Go commands fail because /home/tangi/go/pkg/mod is read-only"
keywords: [go, GOMODCACHE, GOCACHE, "read-only file system", "go.mod cache"]
date: 2026-05-17
---

## Context

Temporary Liza worktrees may inherit `GOPATH=/home/tangi/go` while the module cache under `/home/tangi/go/pkg/mod` is not writable from the agent sandbox.

## Failure Mode

Go commands that need module metadata or downloads fail before reaching the project code, for example:

```text
go: writing go.mod cache: mkdir /home/tangi/go/pkg/mod/cache/download/...: read-only file system
```

## Solution

Run Go commands with per-command writable caches under `/home/tangi/.cache` instead of changing global Go configuration:

```bash
GOPATH=/home/tangi/.cache/go GOMODCACHE=/home/tangi/.cache/go/pkg/mod GOCACHE=/home/tangi/.cache/go/build rtk go test ./...
```

For pre-commit hooks, also set `GOPATH=/home/tangi/.cache/go` so hook-managed Go tools use the writable cache tree.

## References

- `.pre-commit-config.yaml`
