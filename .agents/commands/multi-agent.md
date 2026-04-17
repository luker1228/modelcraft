---
name: Multi-Agent
description: "Analyze task dependencies and dispatch work to multiple agents in parallel"
argument-hint: "<task description or list> [--agent <agent-type>]"
---

Analyze a complex task, break it into subtasks with dependency relationships, and dispatch them to multiple agents working in parallel.

**Input**: A task description or list of tasks. Optionally specify agent type(s) with `--agent <type>`.

- If `--agent` is specified, use that agent type for all spawned agents
- If no agent is specified, use `general-purpose` as the default subagent type
- Multiple agent types can be assigned per-task if the user describes them (e.g., "use Explore for research, Bash for tests")

---

## Steps

### 0. Workflow-aware preset (backend)

If the input clearly indicates `workflow backend`, apply this preset before normal decomposition:

- Create two parallel tracks immediately:
  - `backend-worker` track: implement `<backend-plan-file>`
  - `bdd-test` track: develop/adjust related BDD tests
- These two tracks run in **Wave 1** concurrently.
- After both are done, continue with:
  - `backend-reviewer` (Wave 2)
  - final `bdd-test` execution/acceptance report (Wave 3, or merged into reviewer handoff if explicitly requested)

Preset intent:
- enforce "backend implementation and BDD test development can run in parallel"
- keep reviewer and final BDD acceptance execution after parallel work converges

### 1. Analyze the task

Parse the user's input to identify:
- **Individual subtasks** — break the work into discrete, independently completable units
- **Dependencies** — which tasks must finish before others can start
- **Agent type per task** — use the user-specified agent type, or `general-purpose` if not specified

If the input is vague, use the **AskUserQuestion tool** to clarify scope before proceeding.

### 2. Present the execution plan

Display the dependency graph and get user approval before dispatching:

```
## Multi-Agent Execution Plan

### Tasks & Dependencies

1. [agent-type] Task A — description
2. [agent-type] Task B — description (blocked by: #1)
3. [agent-type] Task C — description
4. [agent-type] Task D — description (blocked by: #2, #3)

### Execution Waves

Wave 1 (parallel): #1, #3
Wave 2 (parallel): #2 (after #1)
Wave 3: #4 (after #2, #3)

### Agent Assignment

- general-purpose: #1, #2, #4
- Explore: #3

Proceed?
```

Use the **AskUserQuestion tool** to confirm. If the user wants changes, adjust and re-present.

### 3. Create the team and task list

```
TeamCreate → team_name: "multi-agent-<timestamp>"
```

Then create all tasks via **TaskCreate**, using `addBlockedBy` to express dependencies:

- Set `subject` and `description` with enough context for the agent to work autonomously
- Set `activeForm` for progress display
- Wire `addBlockedBy` / `addBlocks` to reflect the dependency graph from step 2

### 4. Dispatch agents

Spawn agents for all unblocked tasks (Wave 1) in parallel using the **Agent tool**:

- Set `subagent_type` to the determined agent type for each task
- Set `run_in_background: true` so agents run concurrently
- Set `team_name` to the created team
- Set `name` to a descriptive agent name (e.g., `task-1-research`, `task-2-implement`)
- Include the full task description and any relevant context in the `prompt`
- Tell each agent: "When done, mark your task as completed via TaskUpdate and check TaskList for newly unblocked tasks to claim."

**CRITICAL**: Launch all Wave 1 agents in a **single message** with multiple Agent tool calls to maximize parallelism.

### 5. Monitor and coordinate

As agents complete tasks and send back results:

- Acknowledge completion
- Check if new tasks are unblocked (via **TaskList**)
- Spawn agents for newly unblocked tasks
- If an agent reports an error or blocker, assess and either:
  - Provide guidance via **SendMessage**
  - Reassign the task
  - Ask the user for input

Show progress updates at natural milestones:

```
## Progress Update

✓ #1 Task A — complete
✓ #3 Task C — complete
⟳ #2 Task B — in progress
○ #4 Task D — blocked by #2

Progress: 2/4 tasks complete
```

### 6. Wrap up

When all tasks are complete:

- Display a final summary of what was accomplished
- Clean up: **TeamDelete** to remove the team
- Report any issues or follow-up items

```
## Multi-Agent Complete

All 4 tasks finished.

### Results
- #1 Task A — ✓ done
- #2 Task B — ✓ done
- #3 Task C — ✓ done
- #4 Task D — ✓ done

Team cleaned up.
```

---

## Agent Type Reference

| Type | Best For | Tools |
|------|----------|-------|
| `general-purpose` | Implementation, multi-step tasks | All tools |
| `Explore` | Research, codebase search, read-only investigation | Read, Grep, Glob, Bash (read-only) |
| `Bash` | Running commands, tests, builds | Bash only |
| `Plan` | Architecture, design, planning | All tools |

---

## Guardrails

- **Always get user approval** on the execution plan before dispatching agents
- **Never guess dependencies** — if unclear, ask the user
- **Each agent must be self-contained** — include all necessary context in the prompt, agents don't share conversation history
- **Prefer parallel over sequential** — maximize concurrency by identifying independent tasks
- **Respect the dependency graph** — never dispatch a task before its blockers are resolved
- **Handle failures gracefully** — if an agent fails, report to the user rather than silently retrying
- **Clean up** — always delete the team when done, even if some tasks failed
- **Default to general-purpose** — when no agent type is specified, use `general-purpose` which has access to all tools
- **Read-only agents can't write** — never assign implementation tasks to Explore or Plan agents
