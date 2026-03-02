# AGENTS.md - Code Guidelines for This Repository

## Project Overview

This is a Go (Golang) project for analysis. The module is `github.com/adeelkhan/analytics_service`.

## Build, Lint, and Test Commands

### Running the Application

```bash
# Build the application
go build -o bin/app main.go

# Run the application
go run main.go
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run a single test (specify package and test name)
go test -run TestFunctionName ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...
```

### Linting and Formatting

```bash
# Format code (gofmt)
go fmt ./...

# Run go vet (static analysis)
go vet ./...

# Run all checks
go fmt ./... && go vet ./...
```

### Dependency Management

```bash
go mod download
go mod tidy
```

## Code Style Guidelines

### Naming Conventions

- **Variables/Functions**: `camelCase` (e.g., `calculateDiff`, `processFile`)
- **Constants**: `PascalCase` or `ALL_CAPS` (e.g., `MaxRetries`, `DEFAULT_TIMEOUT`)
- **Types/Interfaces**: `PascalCase` (e.g., `DiffResult`, `Processor`)
- **Packages**: short, lowercase (e.g., `diff`, `parser`)
- **Files**: lowercase with underscores (e.g., `diff_parser.go`)
- **Tests**: `*_test.go` suffix

### Import Organization

Organize imports in three groups with blank lines between:

1. Standard library imports
2. Third-party imports
3. Local packages

```go
import (
    "fmt"
    "os"

    "github.com/pkg/errors"

    "github.com/adeelkhan/analytics_service/internal"
)
```

### Formatting

- Use `gofmt` or `goimports` automatically
- Use tabs for indentation
- No trailing whitespace

### Error Handling

- Always handle errors explicitly; never ignore with `_`
- Return errors early (fail fast)
- Wrap errors with context: `fmt.Errorf("failed to %s: %w", action, err)`
- Use sentinel errors for known conditions

```go
if err != nil {
    return fmt.Errorf("to process diff: %w", err)
}

var (
    ErrInvalidInput = errors.New("invalid input")
    ErrNotFound     = errors.New("not found")
)
```

### Concurrency

- Use goroutines for concurrent operations
- Use channels for communication
- Use `sync.WaitGroup` for coordination
- Use `context.Context` for cancellation and timeouts

### Testing

- Write tests for all exported functions
- Use table-driven tests for multiple cases
- Use descriptive names: `TestFunctionName_Condition_Expected`
- Use subtests for related cases

```go
func TestParseDiff_ValidInput(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    []Change
        wantErr bool
    }{
        {"single line add", "+\nfoo", []Change{{Line: "foo", Type: Add}}, false},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseDiff(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseDiff() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("ParseDiff() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Documentation

- Document all exported functions, types, and constants
- Use doc comments starting with the name being documented

```go
// DiffResult represents the result of a diff operation.
type DiffResult struct {
    Changes []Change
    Stats   Stats
}
```

## Project Structure

```
.
├── cmd/app/main.go
├── internal/diff/
├── internal/parser/
├── pkg/utils/
├── go.mod
└── go.sum
```

## Pre-commit Checklist

1. Run `go fmt ./...`
2. Run `go vet ./...`
3. Run tests: `go test ./...`
4. Check for hardcoded secrets
