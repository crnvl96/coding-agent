## 📐 Architecture Analysis: `how-to-build-a-coding-agent`

### 🎯 What It Teaches

This is a **progressive workshop** that teaches how to build an AI coding agent from scratch. The core idea: start with a dumb chatbot, then bolt on tools one at a time until it becomes a capable coding assistant. The pedagogical strategy is **learning by accretion** — each step adds one concept, and every file is a self-contained, runnable milestone.

---

### 🧬 The Architecture

Every file is a standalone Go program with the **same skeleton**:

```
main()                    → wiring: CLI flags, client, tools, agent
  ├── Agent.Run()         → event loop (the "heartbeat")
  │     ├── read user input
  │     ├── runInference() → call LLM API with conversation + tools
  │     └── tool loop      → if LLM returns tool_use blocks:
  │           ├── match tool by name
  │           ├── execute tool function
  │           └── send results back → repeat inference
  └── runInference()      → converts ToolDefinition[] to Anthropic tool params
```

**Three layers, same in every file:**

| Layer                                                     | What it does                                                                     | Where it lives        |
| --------------------------------------------------------- | -------------------------------------------------------------------------------- | --------------------- |
| **Agent** (`Agent` struct)                                | Event loop: input → Claude → tools → repeat                                      | Copied into each file |
| **Tool registry** (`ToolDefinition` + `[]ToolDefinition`) | Declares tools: name, description, JSON schema, handler function                 | Copied into each file |
| **Tool implementations**                                  | Actual work: `ReadFile()`, `ListFiles()`, `Bash()`, `EditFile()`, `CodeSearch()` | Copied into each file |

This duplication is **intentional** — each file is a self-contained snapshot of the agent at that stage of development. You can `go run read.go` without needing any shared library.

---

### 📈 The 6-Stage Progression

| #   | File                  | Tools added     | What it teaches                                                                               |
| --- | --------------------- | --------------- | --------------------------------------------------------------------------------------------- |
| 1   | `chat.go`             | (none)          | Basic Anthropic API integration, conversation loop                                            |
| 2   | `read.go`             | `read_file`     | **Tool system**: defining tools, JSON schema from Go structs, `runInference` with tools param |
| 3   | `list_files.go`       | + `list_files`  | Multiple tools coexisting, `filepath.Walk`, JSON output to Claude                             |
| 4   | `bash_tool.go`        | + `bash`        | Shell execution, `os/exec`, stdout+stderr capture, error handling                             |
| 5   | `edit_tool.go`        | + `edit_file`   | File mutation, `old_str`→`new_str` replacement, exact-match constraint, file creation         |
| 6   | `code_search_tool.go` | + `code_search` | External tool (ripgrep), building CLI args, result truncation, exit code handling             |

The workshop uses **prompt files** (`prompts/`) as guided exercises for each stage.

---

### 🧩 Key Design Decisions Worth Understanding

**1. The Event Loop (tool orchestration)**

```go
// Pseudocode of the core loop:
for {
    message = runInference(conversation)
    conversation += message
    for {
        for each content block in message:
            if text → print to user
            if tool_use → execute tool, collect results
        if no tool_uses → break (back to user input)
        send tool results back → runInference again
        conversation += new message
    }
}
```

Claude can request **multiple tools in one response**, and the agent executes them all before calling back. This inner loop continues until Claude produces a text-only response.

**2. Tool Definition Pattern**

```go
type ToolDefinition struct {
    Name        string
    Description string
    InputSchema anthropic.ToolInputSchemaParam  // auto-generated from Go struct
    Function    func(json.RawMessage) (string, error)
}
```

Tools are **generic** — they receive raw JSON and return strings. The `GenerateSchema[T]()` helper uses `jsonschema.Reflector` to auto-derive the JSON Schema from Go struct tags, eliminating manual schema writing.

**3. Single-file, no shared code**
Deliberately no `package agent` or shared library. Each `.go` file is a complete program. This forces you to internalize every part of the architecture by reading and modifying it directly. For your own agent, you'd refactor the shared Agent/Run/runInference code into a reusable package.

**4. Anthropic SDK usage (LLM-agnostic via Anthropic-compatible API)**

- Model: `claude-opus-4-6` (hardcoded) — auto-mapped by compatible providers
- Primary backend: DeepSeek via `https://api.deepseek.com/anthropic` (Anthropic-compatible endpoint)
- `anthropic.NewClient()` reads `ANTHROPIC_API_KEY` from env (set to your DeepSeek key)
- Custom base URL via `option.WithBaseURL()` or `ANTHROPIC_BASE_URL` env var
- Conversation is a `[]anthropic.MessageParam` slice that grows unbounded (no context window management)
- Max tokens: 1024 (quite small for a coding agent)
- See [`docs/deepseek.md`](./deepseek.md) for full integration guide

---

### 🔍 What It Leaves for You to Build

The workshop deliberately stops at the "toy agent" level. Here's what a **real** coding agent would need beyond this:

| Missing piece                 | Why it matters                                                                                                                                                     |
| ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Shared library**            | Agent loop and tool system duplicated 6 times — needs extraction                                                                                                   |
| **System prompt**             | No system message controlling agent behavior, personality, or safety                                                                                               |
| **Context window management** | Conversation grows forever, no truncation or summarization                                                                                                         |
| **Tool output size limits**   | Claude gets raw unfiltered tool output; real agents truncate aggressively                                                                                          |
| **Streaming**                 | No SSE/streaming — user waits for full response                                                                                                                    |
| **Error recovery**            | `panic(err)` on JSON unmarshal failures; no retry logic                                                                                                            |
| **Safety/sandboxing**         | `bash` tool runs arbitrary commands with no restrictions                                                                                                           |
| **Multi-model support**       | Hardcoded to Claude Opus 4.6 — but Anthropic-compatible API means any provider with that endpoint works (DeepSeek, etc.). See [`docs/deepseek.md`](./deepseek.md). |
| **File watching / LSP**       | No language server integration for semantic understanding                                                                                                          |
| **Persistent memory**         | No cross-session memory or project indexing                                                                                                                        |
| **Git awareness**             | No git diff, branch awareness, or safe git operations                                                                                                              |

---

### 🔄 LLM Backend: DeepSeek via Anthropic-Compatible API

Instead of Anthropic Claude, this project uses **DeepSeek** as the LLM backend. DeepSeek provides an Anthropic-compatible API at `https://api.deepseek.com/anthropic`, meaning the existing `anthropic-sdk-go` code works **unchanged** — only the base URL and API key change.

Key advantages:

- **No SDK changes**: Same `anthropic.Client`, same tool format, same message protocol
- **Auto model mapping**: `claude-opus-*` → `deepseek-v4-pro`, `claude-haiku/sonnet-*` → `deepseek-v4-flash`
- **1M token context**: vs Claude's 200K — more room for the workshop's unbounded conversation pattern
- **~96% cheaper**: ~$0.88/M output vs Claude Opus ~$75/M

For setup instructions and compatibility details, see [`docs/deepseek.md`](./deepseek.md).

---

### 🗺️ Summary

The workshop's **goal** is to teach the core agent architecture pattern: an event loop that mediates between an LLM and a toolbox. The **method** is progressive disclosure — each file adds one tool to the same skeleton, so you see exactly what changes. The **intended takeaway** is that once you understand this loop + tool pattern, you can build any agent by adding more sophisticated tools and improving the orchestration layer in `Agent.Run()`.
