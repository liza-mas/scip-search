---
title: "Run Liza state commands from the project root"
trigger: "When a Liza state-mutating CLI command returns an internal error from a task worktree"
keywords: [liza, write-checkpoint, mark-blocked, worktree, internal error]
date: 2026-05-17
---

## Context

Liza agents may work inside task worktrees for artifact and git operations, while runtime blackboard state lives at the project root under `.liza/`.

## Failure Mode

Running state-mutating commands such as `liza write-checkpoint` from a task worktree can return:

```json
{"ok":false,"result":null,"error":{"code":"internal","message":"internal error"}}
```

The command is resolving state relative to the wrong workspace context rather than the project blackboard.

## Solution

Run Liza state commands from the project root:

```bash
rtk liza write-checkpoint <task-id> --agent-id <agent-id> ...
```

Keep git and artifact file operations in the assigned worktree, but use `/home/tangi/Workspace/scip-search` as the working directory for `liza get`, `liza write-checkpoint`, `liza set-task-output`, `liza submit-for-review`, `liza await-verdict`, and `liza mark-blocked`. When `liza set-task-output` needs an output file that exists only in the task worktree, pass the absolute worktree path to `--output`.

## References

- `~/.liza/MULTI_AGENT_MODE.md`
- `/home/tangi/Workspace/scip-search/.liza/state.yaml`
