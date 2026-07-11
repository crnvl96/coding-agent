# `read_file`

Reads the contents of a file at a given path.

## Input

| Field  | Type   | Required | Description                       |
| ------ | ------ | -------- | --------------------------------- |
| `path` | string | yes      | Relative path of the file to read |

## Behavior

- Returns the full file contents as a string.
- Returns an error if the file doesn't exist, is a directory, or can't be read.
- No size limit — the entire file is returned. Large files may consume significant context window space.

## Example

```
tool: read_file({"path":"internal/agent/tools.go"})
```

```
result: package agent

import (
    "encoding/json"
    ...
)
```

## Implementation

`internal/agent/tools_read_file.go`
