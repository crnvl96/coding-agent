# 🔄 DeepSeek API Integration

## 🎯 Goal

Use DeepSeek API instead of Anthropic Claude as the LLM backend for the `how-to-build-a-coding-agent` workshop — with **zero SDK changes** and minimal configuration changes.

---

## ✅ Viability: Fully Compatible

DeepSeek provides an **Anthropic-compatible API** at `https://api.deepseek.com/anthropic`. This means the existing `github.com/anthropics/anthropic-sdk-go` Go SDK works **unchanged** when pointed at DeepSeek's endpoint.

### Compatibility Matrix

| Feature                                                  | Used in workshop?            | DeepSeek support?                         |
| -------------------------------------------------------- | ---------------------------- | ----------------------------------------- |
| Text messages (`type: "text"`)                           | ✅ Every stage               | ✅ Fully supported                        |
| Tool definitions (`name`, `input_schema`, `description`) | ✅ Stages 2–6                | ✅ Fully supported                        |
| Tool use (`type: "tool_use"`)                            | ✅ Stages 2–6                | ✅ Fully supported                        |
| Tool results (`tool_use_id`, `content`)                  | ✅ Stages 2–6                | ✅ Fully supported                        |
| `max_tokens`                                             | ✅ All stages                | ✅ Fully supported                        |
| `stop_sequences`                                         | ❌ Not used                  | ✅ Fully supported                        |
| `system` prompt                                          | ❌ Not used                  | ✅ Fully supported                        |
| `temperature`                                            | ❌ Not used                  | ✅ Range [0.0, 2.0]                       |
| `stream` (SSE)                                           | ❌ Not used                  | ✅ Fully supported                        |
| Image / document blocks                                  | ❌ Not used                  | ❌ Not supported                          |
| `is_error` in tool results                               | ✅ Used (for error handling) | ⚠️ Ignored (but error string still works) |

**Only caveat**: The `is_error` flag on tool results is **ignored** by DeepSeek. However, since the workshop sends errors as human-readable strings (e.g., `"tool 'x' not found"`), the model still understands the failure context. This has no practical impact.

---

## 🔧 Integration: Two Approaches

### Approach A: Environment Variables Only (Zero Code Changes)

Set two environment variables and run the existing code as-is — DeepSeek auto-maps Claude model names:

```bash
export ANTHROPIC_BASE_URL="https://api.deepseek.com/anthropic"
export ANTHROPIC_API_KEY="sk-your-deepseek-api-key"

# Run any stage unchanged:
go run read.go
go run edit_tool.go
```

**Why this works:**

1. The `anthropic-sdk-go` reads `ANTHROPIC_BASE_URL` from the environment
2. DeepSeek auto-maps model names: `claude-opus-*` → `deepseek-v4-pro`, `claude-haiku/sonnet-*` → `deepseek-v4-flash`
3. Since the workshop hardcodes `claude-opus-4-6`, it automatically resolves to `deepseek-v4-pro`

### Approach B: Explicit Code Changes (Recommended for Clarity)

Change the model string and explicitly configure the base URL in code. This is clearer for future readers and doesn't rely on auto-mapping.

#### 1. Add import

```go
import (
    // ... existing imports ...
    "github.com/anthropics/anthropic-sdk-go/option"
)
```

#### 2. Change client initialization

```go
// Before:
client := anthropic.NewClient()

// After:
client := anthropic.NewClient(
    option.WithBaseURL("https://api.deepseek.com/anthropic"),
)
```

> **Note**: The API key is still read from `ANTHROPIC_API_KEY` env var. DeepSeek uses the same header name (`x-api-key`), so you just set it to your DeepSeek key.

#### 3. Change the model constant (optional — auto-mapping works too)

```go
// Before:
Model: anthropic.ModelClaudeOpus4_6,

// After:
Model: "deepseek-v4-pro",
```

The `anthropic.ModelClaudeOpus4_6` is `"claude-opus-4-6"` — since DeepSeek auto-maps models starting with `claude-opus` to `deepseek-v4-pro`, the old code works without this change. Change it for documentation clarity.

---

## 📊 Model Selection

| Model               | Context   | Cost (input / output per 1M tokens) | Best for                                         |
| ------------------- | --------- | ----------------------------------- | ------------------------------------------------ |
| `deepseek-v4-pro`   | 1M tokens | $0.435 / $0.87                      | Best quality — use for complex coding tasks      |
| `deepseek-v4-flash` | 1M tokens | $0.14 / $0.28                       | Budget — faster, cheaper, good for simpler tasks |

**Recommendation for the workshop**: Use `deepseek-v4-pro` — it maps naturally from `claude-opus-*` and gives the best tool-calling behavior.

---

## 📦 Step-by-Step Setup

### 1. Get a DeepSeek API Key

1. Go to [platform.deepseek.com](https://platform.deepseek.com/)
2. Sign up / log in
3. Navigate to **API Keys** → Create a new key
4. Top up your balance (minimum ~$2)

### 2. Set Environment Variables

```bash
export ANTHROPIC_API_KEY="sk-your-deepseek-api-key"  # DeepSeek key, same env var name
export ANTHROPIC_BASE_URL="https://api.deepseek.com/anthropic"
```

### 3. Run the Workshop

```bash
cd how-to-build-a-coding-agent

# Stage 1: Basic chat
go run chat.go

# Stage 2: With file reading
go run read.go

# Stage 5: Full toolset (read, list, bash, edit)
go run edit_tool.go

# Stage 6: All tools
go run code_search_tool.go
```

---

## ⚡ Performance Comparison

| Metric                  | Claude Opus 4.6 | DeepSeek V4 Pro | DeepSeek V4 Flash |
| ----------------------- | --------------- | --------------- | ----------------- |
| Input cost / 1M tokens  | $15.00          | $0.435          | $0.14             |
| Output cost / 1M tokens | $75.00          | $0.87           | $0.28             |
| Context window          | 200K            | 1M              | 1M                |
| Max output tokens       | 32K             | 384K            | 384K              |

For the workshop's unbounded conversation pattern (no context window management), the 1M token context is a significant advantage.

---

## 🔮 Architecture Implications

Switching from Anthropic to DeepSeek does **not** change the agent architecture at all:

```
Agent event loop     → unchanged (same anthropic-sdk-go)
Tool registry        → unchanged (same JSON Schema definition)
Tool implementations → unchanged (Go functions operate on filesystem)
API client           → unchanged (same anthropic.Client, different base URL)
```

The only change is what model processes the prompts. Every architectural concept taught by the workshop — the event loop, tool orchestration, JSON Schema generation from Go structs — remains identical.

---

## 📚 References

- [DeepSeek API Docs](https://api-docs.deepseek.com/)
- [DeepSeek Anthropic API Guide](https://api-docs.deepseek.com/guides/anthropic_api)
- [DeepSeek Tool Calls Guide](https://api-docs.deepseek.com/guides/tool_calls)
- [DeepSeek Models & Pricing](https://api-docs.deepseek.com/quick_start/pricing)
- [Anthropic Go SDK](https://github.com/anthropics/anthropic-sdk-go)
