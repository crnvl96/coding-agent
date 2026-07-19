# coding-agent

> **⚠️ Beta — not ready for production use.** APIs, tool behaviours, and configuration may change without notice.

A terminal-based coding agent written in Go. Describe what you want in natural language and the agent reads files, runs shell commands, edits code, and searches your project — all through a chat interface in your terminal.

It uses **DeepSeek** (via an Anthropic-compatible API) with tool-calling to operate on your codebase.

## Features

- **Natural language interface** — ask the agent to explore, refactor, or test your code
- **Five built-in tools** — read files, list directories, run bash, edit files, search code
- **Multi-turn tool use** — the agent chains multiple tools together within a single request (e.g., search → read → edit)
- **Conversational memory** — context persists across turns within a session
- **Zero setup beyond auth** — no databases, indexes, or daemons

## Prerequisites

- **Go 1.26+** (toolchain specified in `go.mod`)
- **A DeepSeek API key** — [platform.deepseek.com](https://platform.deepseek.com/)
- **ripgrep (`rg`)** — required for the `code_search` tool. Install with your package manager (`apt install ripgrep`, `brew install ripgrep`, etc.)

## Installation

```bash
git clone git@github.com:crnvl96/coding-agent.git
cd coding-agent
go build ./cmd/agent/
```

This produces a binary named `agent` in the current directory.

## Configuration

The agent reads its API key and endpoint from `auth.json` in the working directory, falling back to environment variables.

### Using `auth.json` (recommended)

```json
{
  "api_key": "sk-your-deepseek-api-key",
  "base_url": "https://api.deepseek.com/anthropic"
}
```

| Field      | Required | Description                                       |
| ---------- | -------- | ------------------------------------------------- |
| `api_key`  | Yes      | Your DeepSeek API key                             |
| `base_url` | No       | Custom endpoint; defaults to DeepSeek's Anthropic-compatible URL |

### Using environment variables

```bash
export ANTHROPIC_API_KEY="sk-your-deepseek-api-key"
```

The base URL can only be customised through `auth.json` (the `base_url` field). There is no environment variable for it.

### Fallback chain

1. `auth.json` with a non-empty `api_key`
2. Otherwise — the SDK reads `ANTHROPIC_API_KEY` from the environment; the request fails if it's not set

See [docs/configuration.md](docs/configuration.md) for full details.

## Usage

```bash
./agent
```

You'll see:

```
Chat with AI (use 'ctrl-c' to quit)
You:
```

Type a message and press Enter. The agent sends your message to the LLM, executes any tools the model requests, and prints the result. Press **Ctrl+C** or **Ctrl+D** to exit.

### Example session

```
You: what files are in the internal/agent directory?

AI: Let me check.
tool: list_files({"path":"internal/agent"})
result: ["agent.go","agent_test.go","tools.go","tools_bash.go","..."]
AI: The directory contains 12 files: the main agent logic, tool implementations, and tests.

You: run the tests for that package

AI: Running the tests now.
tool: bash({"command":"go test ./internal/agent/"})
result: ok  github.com/adran/coding-agent/internal/agent 0.023s
AI: All tests pass.
```

See [docs/usage.md](docs/usage.md) for a detailed walkthrough with more examples.

## Available tools

| Tool          | Description                                          |
| ------------- | ---------------------------------------------------- |
| `read_file`   | Read file contents at a given path                   |
| `list_files`  | Recursively list files and directories as JSON       |
| `bash`        | Execute shell commands via `bash -c`                 |
| `edit_file`   | Exact-match text replacement, append, and file creation |
| `code_search` | Search code with ripgrep — patterns, file types, truncation |

The model decides which tools to use based on your prompts. You don't call them directly — just describe what you want.

## Project structure

```
cmd/agent/main.go           Entry point — auth, client setup, tool registration
internal/agent/
  agent.go                  Agent struct, chat loop, inference
  agent_test.go             Behavioural tests for the event loop
  tools.go                  ToolDefinition type and schema generation
  tools_bash.go             Bash tool
  tools_code_search.go      Code search tool (ripgrep wrapper)
  tools_edit_file.go        Edit file tool
  tools_list_files.go       List files tool
  tools_read_file.go        Read file tool
  tools_*_test.go           Tests for each tool
docs/                       User-facing documentation
```

## Development

```bash
# Build
go build ./...

# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with the race detector
go test -race ./...

# Format code (requires gofumpt)
gofumpt -w .

# Static analysis
go vet ./...
```

The project follows these conventions:

- **No third-party mocking libraries** — mocks use stdlib interfaces and function fields
- **Behavioural tests** — tests exercise the public API, not internal helpers
- **`gofumpt` for formatting** — a stricter `gofmt`; run `gofumpt -w .` before committing

## Architecture

The agent runs a two-level event loop:

1. **Outer loop** — reads user messages from stdin, appends to conversation history
2. **Inner loop** — processes tool-use responses from the LLM, executes tools locally, feeds results back, repeats until the model produces a text-only response

```
User → LLM → tool_use? → execute tools → send results → LLM → text → User
              ↑___________________________________________|
```

The LLM receives the full conversation history on every inference call, giving it context across turns. Tools are defined as `ToolDefinition` structs with auto-derived JSON Schema (via `GenerateSchema[T]`) — adding a new tool means writing one file, registering it in `main.go`, and adding tests.

See [docs/architecture.md](docs/architecture.md) and [docs/data-flow.md](docs/data-flow.md) for a deeper dive.

## Limitations

- **No context window management** — very long sessions will eventually exceed the model's context limit
- **No memory across sessions** — restarting starts fresh; no project index or database
- **No command sandboxing** — shell commands run with your user's permissions without confirmation
- **Single model** — hardcoded to `deepseek-v4-pro`; change the `defaultModel` constant to switch


