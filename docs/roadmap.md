# Roadmap

Building a coding agent from scratch, following the [how-to-build-a-coding-agent](../how-to-build-a-coding-agent/) workshop's progressive accretion strategy.

## Done

| #   | Stage        | What it adds                                                                    | Source                               |
| --- | ------------ | ------------------------------------------------------------------------------- | ------------------------------------ |
| 1   | Basic chat   | Event loop, API integration, conversation state                                 | `internal/agent/agent.go`            |
| 2   | `read_file`  | Tool system (`ToolDefinition`, `GenerateSchema`, inner loop), file reading      | `internal/agent/tools.go`, `agent.go` |

## Upcoming

| #   | Tool          | What it adds                                                                              | Dependencies        |
| --- | ------------- | ----------------------------------------------------------------------------------------- | ------------------- |
| 3   | `list_files`  | Directory exploration via `filepath.Walk`, JSON output to model                           | stdlib only         |
| 4   | `bash`        | Shell execution via `os/exec`, stdout+stderr capture, error handling                      | stdlib only         |
| 5   | `edit_file`   | File mutation — `old_str`→`new_str` exact-match replacement, file creation                | stdlib only         |
| 6   | `code_search` | Wraps ripgrep for pattern search, CLI arg building, result truncation, exit code handling | `rg` binary in PATH |

## Architecture

```
cmd/agent/main.go          → wiring: flags, client (DeepSeek), agent
internal/agent/
  agent.go                 → Agent struct, event loop, inference, tool loop
  tools.go                 → ToolDefinition, GenerateSchema, tool implementations
```

Each tool is a `ToolDefinition` (name, description, JSON Schema, handler function) — adding a tool is a single-file change.
