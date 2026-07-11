# Roadmap

Building a coding agent from scratch, following the [how-to-build-a-coding-agent](../how-to-build-a-coding-agent/) workshop's progressive accretion strategy.

## Done

| #   | Stage        | What it adds                                                                    | Source                                              |
| --- | ------------ | ------------------------------------------------------------------------------- | --------------------------------------------------- |
| 1   | Basic chat   | Event loop, API integration, conversation state                                 | `internal/agent/agent.go`                           |
| 2   | `read_file`  | Tool system (`ToolDefinition`, `GenerateSchema`, inner loop), file reading      | `internal/agent/tools.go`, `tools_read_file.go`     |
| 3   | `list_files` | Directory exploration via `filepath.Walk`, JSON output to model                 | `internal/agent/tools_list_files.go`                |
| 4   | `bash`       | Shell execution via `os/exec`, stdout+stderr capture, inline error reporting    | `internal/agent/tools_bash.go`                      |

## Upcoming

| #   | Tool          | What it adds                                                                              | Dependencies        |
| --- | ------------- | ----------------------------------------------------------------------------------------- | ------------------- |
| 5   | `edit_file`   | File mutation — `old_str`→`new_str` exact-match replacement, file creation                | stdlib only         |
| 6   | `code_search` | Wraps ripgrep for pattern search, CLI arg building, result truncation, exit code handling | `rg` binary in PATH |

## Architecture

```
cmd/agent/main.go          → wiring: flags, client (DeepSeek), agent
internal/agent/
  agent.go                 → Agent struct, event loop, inference, tool loop
  agent_test.go            → behavioural tests for the event loop and tool orchestration
  tools.go                 → ToolDefinition, GenerateSchema
  tools_read_file.go       → ReadFile tool (definition + handler)
  tools_list_files.go      → ListFiles tool (definition + handler)
  tools_list_files_test.go → behavioural tests for ListFiles
  tools_bash.go            → Bash tool (definition + handler)
  tools_bash_test.go       → behavioural tests for Bash
```

Each tool lives in its own file: a `ToolDefinition` (name, description, JSON Schema, handler function) plus any helper types. Adding a tool means adding one implementation file.
