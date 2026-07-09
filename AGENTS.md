# Agent Instructions

## Go tooling

- **Compilation**: run `go build ./...` to verify everything compiles.
- **Static analysis**: run `go vet ./...` — catches shadowed variables, bad printf verbs, unreachable code, and more.
- **Formatting**: run `gofumpt -w .` to format all Go source files. Never run plain `go fmt` — always use `gofumpt`.

## Documentation guidelines

- **Document all exported symbols** with a `// SymbolName` doc comment describing their purpose and behaviour. Unexported symbols should not have doc comments — a brief inline comment is acceptable only when the intent is not obvious from the code.

## Testing guidelines

- Mock with stdlib only (interfaces, function fields, `httptest`, `io.Reader`). No third-party mocking libraries.
- Test behavior, not implementation. Prefer exercising the public API over unit-testing unexported helpers in isolation.
