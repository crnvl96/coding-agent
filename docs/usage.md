# Usage

This document assumes the agent is running in your terminal (you see `Chat with AI (use 'ctrl-c' to quit)`). It covers what you can ask the agent to do, how it responds, and behavioural details worth knowing.

## How it works

You type a message, press Enter. The agent sends your message to the LLM along with the full conversation history and a catalog of available tools. The model decides whether to reply with text or to invoke tools. If it calls tools, the agent executes them locally, feeds the results back, and asks the model again — repeating until the model produces a final text answer. Then it waits for your next message.

## Available tools

The agent has five tools the model can use. You don't call them directly — you ask in natural language, and the model decides which to use.

| Tool          | What it does                                 | Example prompt                                    |
| ------------- | -------------------------------------------- | ------------------------------------------------- |
| `read_file`   | Reads the contents of a file                 | "Show me the contents of main.go"                 |
| `list_files`  | Lists files and directories                  | "What files are in the internal/agent directory?" |
| `bash`        | Executes a shell command                     | "Run go test ./..."                               |
| `edit_file`   | Replaces text in a file or creates a new one | "Rename the variable `x` to `count` in tools.go"  |
| `code_search` | Searches code with ripgrep                   | "Find all places where ToolDefinition is used"    |

The model can combine multiple tools in one turn — e.g., search for a function definition, then read the file containing it.

## What you'll see in the terminal

```
You: find where error handling is done
```

```
AI: Let me search for that.

tool: code_search({"pattern":"if err != nil","file_type":"go"})
result: agent.go:77:    if err != nil {
agent.go:102:       if err != nil {
tools_edit_file.go:50: if err != nil {
...

AI: The main error handling is in agent.go (the event loop) and in each tool file.
```

- **`You:`** — your input prompt.
- **`AI:`** — text responses from the model.
- **`tool:`** — the model requested a tool; shows which tool and with what input.
- **`result:`** — the tool's output, fed back to the model.
- **`error:`** — something went wrong (tool not found, execution failed). This is shown to you and also sent back to the model so it can recover.

## Session behaviour

**Conversation persists.** The model remembers everything said within a session. If you ask "what was that file again?", it can refer to earlier results.

**No context window management.** The conversation grows without bound. On very long sessions the prompt will eventually exceed the model's context limit (1M tokens with DeepSeek), at which point the API will return an error and the session ends.

**No memory across sessions.** Quitting and restarting starts fresh. The agent has no database, no project index — it re-discovers your codebase on each session.

## Gotchas

### The model drives tool selection

You can't force a specific tool. "Read main.go with the read_file tool" works, but so does "show me main.go" — the model infers the right tool. If it insists on using `bash` with `cat` instead of `read_file`, rephrase your request.

### `edit_file` requires exact matches

When the model uses `edit_file`, the `old_str` must match the file content exactly (same whitespace, same indentation). If the model guesses wrong, the edit fails with "old_str not found". The model usually reads the file first to get the exact text, but this isn't guaranteed — if you see repeated edit failures, ask the model to read the file first, then edit.

### `edit_file` requires unique matches

If `old_str` appears multiple times in the file, `edit_file` rejects the edit. The model needs to include enough surrounding context to make the match unique.

### `bash` runs without restrictions

Shell commands execute with your user's permissions in the agent's working directory. Commands like `rm`, `mv`, `curl`, or `git push` will run if the model requests them. The agent does not sandbox, confirm, or restrict commands. Treat it like a terminal session — the model acts on your behalf.

### `code_search` depends on `rg`

The `code_search` tool wraps ripgrep. If `rg` is not in your PATH, searches will fail. Install it (`apt install ripgrep`, `brew install ripgrep`, etc.) or use `bash` with `grep` as a fallback.

### Empty input is ignored

Pressing Enter on an empty line does nothing — the agent skips it and re-prompts.

### Exiting

Press **Ctrl+C** to quit. Pressing **Ctrl+D** (EOF) also exits cleanly. The agent does not save session state on exit.

## Example sessions

**Exploring a codebase:**

```
You: what packages does this project have?

AI: Let me check the directory structure.
tool: list_files({})
result: ["cmd/","cmd/agent/","internal/","internal/agent/",...]

AI: The project has two packages: cmd/agent (the entry point) and internal/agent (the core library).
```

**Editing with read-then-edit:**

```
You: rename the function ReadFile to Read in tools_read_file.go

AI: I'll read the file first to find the exact text.
tool: read_file({"path":"internal/agent/tools_read_file.go"})
result: package agent
...

AI: Now I'll replace the function name.
tool: edit_file({"path":"internal/agent/tools_read_file.go","old_str":"func ReadFile(","new_str":"func Read("})
result: OK

AI: I've renamed ReadFile to Read in the function definition.
```

**Running tests through bash:**

```
You: do all the tests pass?

AI: Let me run the tests.
tool: bash({"command":"go test ./..."})
result: ok  github.com/adran/coding-agent/internal/agent  0.023s

AI: Yes, all tests pass.
```
