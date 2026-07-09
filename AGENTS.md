# Agent Instructions

## Go tooling

- **Type checking**: run `gopls check .` (or per-package: `gopls check ./internal/agent`) before committing.
- **Formatting**: run `gofumpt -w .` to format all Go source files. Never run plain `go fmt` — always use `gofumpt`.

## Testing guidelines

- Mock with stdlib only (interfaces, function fields, `httptest`, `io.Reader`). No third-party mocking libraries.
- Test behavior, not implementation. Prefer exercising the public API over unit-testing unexported helpers in isolation.
