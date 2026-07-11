# Architecture

## Overview

A terminal-based coding agent built in Go. It reads user messages from stdin, calls an LLM, and executes tools the model requests — reading files, listing directories, running shell commands, editing files, and searching code. The core loop runs until the model produces a text-only response, then returns control to the user.

## Package layout

```
cmd/agent/main.go          Entry point: auth, client, tool wiring
internal/agent/
  agent.go                 Agent, event loop, inference
  agent_test.go            Behavioural tests for the event loop and tool orchestration
  tools.go                 ToolDefinition, GenerateSchema (shared infrastructure)
  tools_<name>.go          One file per tool: definition + handler
  tools_<name>_test.go     Behavioural tests for each tool
```

## The event loop

```
for each user message:
    send conversation + tools to LLM
    append response to conversation
    while response contains tool_use blocks:
        execute each matching tool
        collect results (or errors)
        send results back → LLM responds
        append to conversation
    display text-only response to user
```

The inner `while` loop absorbs all tool calls before talking to the user. The model can request multiple tools in one response (executed sequentially), and the loop continues through as many inference rounds as needed until the model produces a text-only message.

## Tools

Each tool is a `ToolDefinition` — a name, a description, an auto-derived JSON Schema (from a Go struct via `GenerateSchema`), and a handler function `func(json.RawMessage) (string, error)`. Tools live in their own files:

| Tool          | File                   | What it does                                                                 |
| ------------- | ---------------------- | ---------------------------------------------------------------------------- |
| `read_file`   | `tools_read_file.go`   | Reads file contents at a given path                                          |
| `list_files`  | `tools_list_files.go`  | Recursively lists directories as JSON arrays                                 |
| `bash`        | `tools_bash.go`        | Executes shell commands via `bash -c`                                        |
| `edit_file`   | `tools_edit_file.go`   | Exact-match `old_str` → `new_str` replacement, with append and file creation |
| `code_search` | `tools_code_search.go` | Wraps ripgrep for pattern search, with result truncation                     |

Adding a tool means writing one implementation file, registering it in `main.go`, and adding behavioural tests.

## Schema generation

`GenerateSchema[T any]()` uses `jsonschema.Reflector` to derive a JSON Schema from a Go struct with `jsonschema_description` tags. This eliminates manual schema authoring — the struct is the single source of truth for both the Go code and the tool's input contract.

## LLM backend

The agent uses **DeepSeek** via an Anthropic-compatible API endpoint (`https://api.deepseek.com/anthropic`). The `anthropic-sdk-go` client works unchanged — only the base URL and API key differ. The model is `deepseek-v4-pro` (mapped from `claude-opus-*` by the provider).

Authentication is loaded from `auth.json` (a JSON file with `api_key` and optional `base_url`) or falls back to the `ANTHROPIC_API_KEY` environment variable.

## Testing strategy

Tests exercise **behaviour, not implementation details**. The event loop and tool orchestration are tested through the public `Run()` API using a mock chat creator that returns pre-configured responses. Individual tools are tested directly against their handler functions using temp directories and real system commands (ripgrep tests skip gracefully when `rg` is unavailable). No third-party mocking libraries — only stdlib interfaces and function fields.
