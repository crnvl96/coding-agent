# `code_search`

Searches code for patterns using ripgrep (`rg`).

## Input

| Field            | Type   | Required | Description                                                                                   |
| ---------------- | ------ | -------- | --------------------------------------------------------------------------------------------- |
| `pattern`        | string | yes      | Search pattern or regex                                                                       |
| `path`           | string | no       | File or directory to search. Defaults to the current directory.                               |
| `file_type`      | string | no       | File extension to filter by (e.g. `"go"`, `"js"`, `"py"`). Passed to ripgrep's `--type` flag. |
| `case_sensitive` | bool   | no       | Whether the search is case-sensitive. Defaults to `false`.                                    |

## Behavior

- Runs `rg --line-number --with-filename --color=never` with the provided options.
- Adds `--ignore-case` unless `case_sensitive` is `true`.
- Adds `--type <file_type>` if `file_type` is set.
- If ripgrep finds no matches (exit code 1), returns `"No matches found"` — this is not treated as an error.
- If ripgrep fails for another reason (exit code > 1, or `rg` not found), returns an error.
- Output is **truncated to 50 lines** if there are more matches. A summary line like `... (showing first 50 of 120 matches)` is appended.

## Example

```
tool: code_search({"pattern":"func NewAgent","file_type":"go","path":"internal/"})
```

```
result: internal/agent/agent.go:42:func NewAgent(
internal/agent/agent_test.go:55:   agt := NewAgent(mock, getUserMessage, []ToolDefinition{tool})
```

## Dependency

Requires `rg` (ripgrep) in PATH. If `rg` is not installed, searches will fail with an error.

## Implementation

`internal/agent/tools_code_search.go`
