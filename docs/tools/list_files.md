# `list_files`

Recursively lists files and directories at a given path, returning a JSON array of relative paths.

## Input

| Field  | Type   | Required | Description                                                   |
| ------ | ------ | -------- | ------------------------------------------------------------- |
| `path` | string | no       | Directory to list. Defaults to the current working directory. |

## Behavior

- Walks the directory tree recursively.
- Directories are suffixed with `/` to distinguish them from files.
- The root directory itself (`.`) is excluded from the listing.
- Returns a JSON array of relative paths, e.g. `["file.go","subdir/","subdir/file.go"]`.
- An empty directory returns `[]`.
- Returns an error if the path doesn't exist or can't be read.

## Example

```
tool: list_files({"path":"internal/agent"})
```

```
result: ["agent.go","agent_test.go","tools.go","tools_bash.go","tools_bash_test.go",...]
```

## Implementation

`internal/agent/tools_list_files.go`
