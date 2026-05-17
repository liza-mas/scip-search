---
title: "Use explicit test binary for RTK directory checks"
trigger: "When an RTK-prefixed `test -d` worktree check fails with `Illegal option -d`"
keywords: [rtk, test, /usr/bin/test, Illegal option -d, worktree]
date: 2026-05-17
---

## Context

Liza bootstrap prompts may require a literal worktree existence check such as `test -d <worktree>`, while Codex sessions are expected to prefix shell commands with `rtk`.

## Failure Mode

Running `rtk test -d <worktree>` can fail with:

```text
sh: 0: Illegal option -d
```

RTK can route the bare `test` command through shell handling where `-d` is interpreted as a shell option instead of the `test` predicate.

## Solution

Use the explicit binary under RTK:

```bash
rtk /usr/bin/test -d /home/tangi/Workspace/scip-search/.worktrees/<task-id>
```

This preserves the required directory-existence check while avoiding shell-option parsing.

## References

- `~/.liza/AGENT_TOOLS.md`
