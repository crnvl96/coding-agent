# Configuration

The agent authenticates with the LLM backend and configures its endpoint through a fallback chain: `auth.json` first, then environment variables.

## `auth.json`

Place a JSON file named `auth.json` in the working directory (alongside the agent binary or wherever you run it from):

```json
{
  "api_key": "sk-your-deepseek-api-key",
  "base_url": "https://api.deepseek.com/anthropic"
}
```

| Field      | Required | Description                                                                                            |
| ---------- | -------- | ------------------------------------------------------------------------------------------------------ |
| `api_key`  | yes      | Your DeepSeek API key. If empty or missing, the file is ignored and the agent falls back to env vars.  |
| `base_url` | no       | Custom Anthropic-compatible API endpoint. Defaults to `https://api.deepseek.com/anthropic` if omitted. |

If the file is missing, malformed, or has an empty `api_key`, the agent silently falls back to environment variables — it does not print a warning.

## Environment variables

Set these when not using `auth.json`, or when `auth.json` has an empty `api_key`:

| Variable             | Description                                                                                    |
| -------------------- | ---------------------------------------------------------------------------------------------- |
| `ANTHROPIC_API_KEY`  | Your DeepSeek API key                                                                          |
| `ANTHROPIC_BASE_URL` | Anthropic-compatible API endpoint (optional, defaults to `https://api.deepseek.com/anthropic`) |

The `ANTHROPIC_BASE_URL` default is always applied regardless of auth method. If `auth.json` provides a `base_url`, it overrides the environment variable and the default.

## Fallback chain

```
1. auth.json with non-empty api_key → use api_key (and base_url if present)
2. ANTHROPIC_API_KEY env var        → use env var key
3. Neither                          → anthropic-sdk-go looks for ANTHROPIC_API_KEY internally;
                                       fails if not set
```

## Model

The model is hardcoded to `deepseek-v4-pro` in `internal/agent/agent.go`. This is the best-quality DeepSeek model for tool-calling. To switch models, change the `defaultModel` constant.

The original Anthropic SDK model names (e.g. `claude-opus-4-6`) are auto-mapped by DeepSeek's Anthropic-compatible endpoint — you can use either naming convention.
