# `bash`

Executes a shell command via `bash -c` and returns its combined output (stdout + stderr).

## Input

| Field     | Type   | Required | Description                 |
| --------- | ------ | -------- | --------------------------- |
| `command` | string | yes      | The bash command to execute |

## Behavior

- Runs the command through `bash -c`.
- Captures combined stdout and stderr.
- Command failures (non-zero exit code) are reported **inline in the output string**, not as tool errors. This lets the model read and react to failure details.
- Successful output has trailing whitespace trimmed.
- Runs with the agent's user permissions in the current working directory — **no sandboxing, no confirmation prompts, no restrictions**.

## Example (success)

```
tool: bash({"command":"go test ./internal/agent/"})
```

```
result: ok  github.com/adran/coding-agent/internal/agent  0.023s
```

## Example (failure)

```
tool: bash({"command":"cat /nonexistent"})
```

```
result: Command failed with error: exit status 1
Output: cat: /nonexistent: No such file or directory
```

## Implementation

`internal/agent/tools_bash.go`
