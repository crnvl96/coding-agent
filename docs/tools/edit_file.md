# `edit_file`

Replaces text in a file using exact-match `old_str` → `new_str` substitution. Can also append to existing files or create new ones.

## Input

| Field     | Type   | Required | Description                |
| --------- | ------ | -------- | -------------------------- |
| `path`    | string | yes      | Path to the file to modify |
| `old_str` | string | yes      | Text to find and replace   |
| `new_str` | string | yes      | Replacement text           |

## Behavior

The tool operates in three modes depending on `old_str` and whether the file exists:

### Replacement mode (`old_str` non-empty, file exists)

`old_str` must appear **exactly once** in the file. Zero matches returns an error (`"old_str not found in file"`). Multiple matches returns an error with the count (`"old_str found 3 times in file, must be unique"`). The single occurrence is replaced with `new_str`.

### Append mode (`old_str` empty, file exists)

`new_str` is appended to the end of the file. No newline is automatically added — include `\n` in `new_str` if you want one.

### Creation mode (`old_str` empty, file doesn't exist)

The file is created with `new_str` as its contents. Parent directories are created automatically if needed.

## Constraints

- `old_str` must differ from `new_str`. Passing the same value for both returns an error.
- `path` must not be empty.
- When `old_str` is non-empty and the file doesn't exist, an error is returned (you can't replace text in a nonexistent file — use creation mode instead).

## Example

```
tool: edit_file({"path":"tools.go","old_str":"func ReadFile(","new_str":"func Read("})
```

```
result: OK
```

## Implementation

`internal/agent/tools_edit_file.go`
