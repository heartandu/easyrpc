# AGENTS.md

## Build Commands
- `just build` - Build the binary to `bin/easyrpc`
- `just test` - Run all tests with race detection and coverage
- `go test -v -race -count=1 test ./pkg/...` - Run tests for specific package
- `just lint` - Run golangci-lint with comprehensive rules

## Code Style
- **Go version**: 1.24.6
- **Imports**: Use gci formatting with standard/default/local sections, local prefix `github.com/heartandu/easyrpc`
- **Formatting**: Use gofumpt (stricter than gofmt) and goimports
- **Line length**: Max 120 characters
- **Constants**: Use const blocks with descriptive names, prefer lowercase with camelCase
- **Error handling**: Wrap errors with context using `fmt.Errorf("msg: %w", err)`, avoid bare returns
- **Naming**: Use camelCase, avoid abbreviations, descriptive variable names
- **Comments**: Package comments disabled, avoid excessive inline comments
- **Types**: Use explicit types for struct fields, avoid naked returns
- **Testing**: Use testify for assertions, place test files in separate test/ directory for integration tests

## Linting Rules
- Comprehensive linting with most rules enabled except: depguard, exhaustruct, funcorder, ireturn, mnd, noinlineerr, varnamelen, wsl
- Cognitive complexity limit: 20
- Test files have relaxed rules for dupl, err113, errcheck, forcetypeassert, funlen, gocyclo, gosec, lll, revive, varnamelen, wrapcheck
