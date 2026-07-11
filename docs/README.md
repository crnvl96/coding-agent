# Docs

| Document                          | Description                                                                          |
| --------------------------------- | ------------------------------------------------------------------------------------ |
| [Usage](usage.md)                 | How to interact with the agent — tools, terminal output, gotchas, example sessions   |
| [Architecture](architecture.md)   | High-level design — package layout, event loop, tool system, testing strategy        |
| [Data flow](data-flow.md)         | Lifecycle of a user request — from stdin through tool execution and back             |
| [Configuration](configuration.md) | Authentication — `auth.json`, environment variables, fallback chain, model selection |

## Tools

| Tool                                  | Description                                                                         |
| ------------------------------------- | ----------------------------------------------------------------------------------- |
| [`read_file`](tools/read_file.md)     | Read file contents at a given path                                                  |
| [`list_files`](tools/list_files.md)   | Recursively list files and directories as a JSON array                              |
| [`bash`](tools/bash.md)               | Execute shell commands via `bash -c`                                                |
| [`edit_file`](tools/edit_file.md)     | Exact-match text replacement, append, and file creation                             |
| [`code_search`](tools/code_search.md) | Search code with ripgrep — pattern matching, file type filtering, result truncation |
