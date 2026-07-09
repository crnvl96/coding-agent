# Agent Instructions

## Go tooling

- **Type checking**: run `go vet ./...` — catches shadowed variables, bad printf verbs, unreachable code, and more. Also run `go build ./...` to verify everything compiles.
- **Formatting**: run `gofumpt -w .` to format all Go source files. Never run plain `go fmt` — always use `gofumpt`.

## Testing guidelines

- Mock with stdlib only (interfaces, function fields, `httptest`, `io.Reader`). No third-party mocking libraries.
- Test behavior, not implementation. Prefer exercising the public API over unit-testing unexported helpers in isolation.
