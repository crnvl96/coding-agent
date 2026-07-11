# Data flow

This document walks through the lifecycle of a single user request — from stdin to LLM, through tool execution, and back to the terminal. It traces the data at each step, showing what structures are built, transformed, and passed between components.

## 1. Startup

`cmd/agent/main.go` wires three things before the loop begins:

1. **An LLM client** — `anthropic.Client` pointed at DeepSeek's Anthropic-compatible endpoint, authenticated via `auth.json` or the `ANTHROPIC_API_KEY` environment variable.
2. **A tool registry** — a `[]ToolDefinition` slice containing all five tools (`read_file`, `list_files`, `bash`, `edit_file`, `code_search`).
3. **An input function** — a closure `func() (string, bool)` that reads lines from stdin via `bufio.Scanner`.

These are passed to `NewAgent`, which stores them on the `agent` struct. The agent holds no other state — all conversational context lives in the `conversation` slice built during the loop.

## 2. The outer loop: user turn

```
┌─ Agent.Run()
│
│  conversation = []
│
│  ┌─ loop (per user message) ────────────────────┐
│  │                                                │
│  │  1. Read user input from stdin                 │
│  │  2. Wrap into a MessageParam (role: "user")    │
│  │  3. Append to conversation                     │
│  │  4. Call runInference(conversation)             │
│  │  5. Enter inner loop (tool handling)           │
│  │  6. Display final text to user                 │
│  │                                                │
│  └────────────────────────────────────────────────┘
```

**Step 2 — wrapping user input:**

```go
// Raw input: "read the file tools.go"
userMessage := anthropic.NewUserMessage(
    anthropic.NewTextBlock("read the file tools.go"),
)
// Produces a MessageParam{Role: "user", Content: [{Type: "text", Text: "read the file tools.go"}]}

conversation = append(conversation, userMessage)
// conversation = [MessageParam{role: user, ...}]
```

## 3. runInference: calling the LLM

`runInference` takes the current conversation and sends it to the LLM along with the tool definitions.

**Tool conversion:**

Each `ToolDefinition` is converted to the Anthropic API format:

```go
// ToolDefinition{Name: "read_file", Description: "...", InputSchema: {...}, Function: ReadFile}
//      ↓
// ToolUnionParam{
//     OfTool: &ToolParam{
//         Name:        "read_file",
//         Description: "Read the contents...",
//         InputSchema: {Properties: {path: {type: "string"}}},
//     },
// }
```

The full `MessageNewParams` sent to the API:

```go
anthropic.MessageNewParams{
    Model:     "deepseek-v4-pro",
    MaxTokens: 4096,
    Messages:  conversation,  // accumulated history
    Tools:     anthropicTools, // all registered tools
}
```

The LLM receives the full conversation history plus the tool catalog and returns a `Message` containing one or more content blocks.

## 4. Processing the LLM response

The response's `Content` is a `[]ContentBlockUnion` — each block has a `Type` field. The agent iterates blocks and branches:

```
┌─ for each content block in message.Content
│
│  type == "text"
│  ├── Print to terminal ("AI: ...")
│  └── (no state change)
│
│  type == "tool_use"
│  ├── Print to terminal ("tool: read_file({...})")
│  ├── Match tool by name against agent.tools
│  ├── Call tool.Function(input) → (result string, err error)
│  ├── Print result or error to terminal
│  └── Collect into toolResults slice
│
└─ end for
```

**Tool matching:**

```go
for _, tool := range a.tools {
    if tool.Name == "read_file" {
        toolResult, toolError = tool.Function(json.RawMessage(`{"path":"tools.go"}`))
        break
    }
}
```

The tool function receives raw JSON and returns `(string, error)`. The agent doesn't know or care what the tool does — it just passes input and collects output.

## 5. Tool result handling

For each tool execution, the agent builds a result block:

```go
// Success:
anthropic.NewToolResultBlock("toolu_001", "package agent\n...", false)
// → ContentBlockParamUnion with OfToolResult set, IsError=false

// Failure (tool not found, function error, etc.):
anthropic.NewToolResultBlock("toolu_001", `tool "x" not found`, true)
// → ContentBlockParamUnion with OfToolResult set, IsError=true
```

All results are collected and wrapped into a single user message:

```go
toolResultMessage := anthropic.NewUserMessage(
    toolResultForReadFile,
    toolResultForBash,
)
// → MessageParam{Role: "user", Content: [toolResultBlock, toolResultBlock]}
```

This message is appended to the conversation, giving the LLM access to every tool's output in the next turn.

## 6. The inner loop

The inner loop is what makes the agent multi-turn within a single user request:

```
┌─ inner loop
│
│  conversation = [...user, ...assistant-with-tool_use]
│
│  hasToolUse = true ──────────────┐
│  │                               │
│  │  Execute all tools            │
│  │  Collect results              │
│  │  Append results to            │
│  │  conversation as user msg     │
│  │                               │
│  │  runInference(conversation)   │
│  │       │                       │
│  │       ├── hasToolUse == true ─┘ (loop)
│  │       │
│  │       └── hasToolUse == false → break
│  │
│  └── Print text response to user
│
└─ end inner loop
```

**Example — two rounds of tool use:**

```
User:   "find and show the main function"
→ [user: "find and show the main function"]

LLM:    tool_use: code_search(pattern="func main")
→ [user: "...", assistant: tool_use(code_search)]

Agent:  executes code_search → result: "main.go:5:func main() {"
→ [user: "...", assistant: tool_use, user: tool_result("main.go:5:...")]

LLM:    tool_use: read_file(path="main.go")
→ [..., assistant: tool_use(read_file)]

Agent:  executes read_file → result: "package main\nfunc main() {...}"
→ [..., assistant: tool_use, user: tool_result("package main...")]

LLM:    text: "The main function is in main.go at line 5..."
→ [..., assistant: text("The main function...")]
→ hasToolUse == false → break

AI: The main function is in main.go at line 5...
```

Each tool round expands the conversation. The LLM sees the full history and decides whether to call more tools or respond.

## 7. End of turn

When the LLM produces a text-only response (no `tool_use` blocks), the inner loop breaks. The text is printed to the terminal, control returns to the outer loop, and the agent waits for the next user message. The conversation accumulates across turns — the LLM retains context from all previous interactions within a session.

## Key data structures

| Structure                                          | Role                                                       | Grows over time?                                                  |
| -------------------------------------------------- | ---------------------------------------------------------- | ----------------------------------------------------------------- |
| `[]anthropic.MessageParam` (conversation)          | Full message history sent to the LLM                       | Yes — each turn appends user, assistant, and tool-result messages |
| `[]ToolDefinition` (agent.tools)                   | Catalog of available tools with their schemas and handlers | No — fixed at startup                                             |
| `[]anthropic.ContentBlockUnion` (message.Content)  | One LLM response: text blocks and/or tool_use blocks       | No — per-response                                                 |
| `[]anthropic.ContentBlockParamUnion` (toolResults) | Tool results collected within one inner-loop iteration     | No — rebuilt each iteration                                       |

## Error paths

| Error                      | Where            | Effect                                              |
| -------------------------- | ---------------- | --------------------------------------------------- |
| LLM API error              | `runInference`   | Propagated to `Run()`, terminates the session       |
| Tool not found             | Inner loop       | Wrapped as error tool result, sent back to LLM      |
| Tool function error        | Inner loop       | Wrapped as error tool result, sent back to LLM      |
| Invalid JSON input to tool | Tool handler     | Returned as error, treated like tool function error |
| EOF (ctrl-d)               | `getUserMessage` | Outer loop breaks, `Run()` returns nil              |
