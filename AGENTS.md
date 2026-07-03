# AGENTS.md

## Project Overview

CrashForge is a crash and error log analyzer CLI that parses stack traces from multiple programming languages, groups similar errors, and provides actionable root cause suggestions.

## Building

```bash
go build -o crashforge ./cmd/crashforge/
```

## Testing

```bash
go test ./...
```

Run tests with verbose output:

```bash
go test ./... -v
```

Run a specific package's tests:

```bash
go test ./parser/...
go test ./analyzer/...
go test ./reporter/...
```

## Architecture

- `parser/`: Stack trace parsing for Go, Python, JavaScript, Java, Rust, C
- `analyzer/`: Error grouping, frequency analysis, root cause suggestions
- `reporter/`: Report generation in text, JSON, and Markdown formats
- `cmd/crashforge/`: CLI entry point using Cobra

## Adding a New Language

1. Add a new `Language` constant in `parser/parser.go`
2. Create a `register*` method in `parser/languages.go`
3. Implement the language-specific frame parsing
4. Add tests in `parser/parser_test.go`

## Key Data Structures

- `parser.ErrorInfo`: Parsed error with message, type, language, frames
- `parser.StackFrame`: Single frame with function, file, line
- `analyzer.ErrorGroup`: Grouped similar errors with statistics
- `analyzer.AnalysisResult`: Full analysis with groups, top files, top errors

## Dependencies

- `github.com/spf13/cobra`: CLI framework
- Standard library only for core logic (no external dependencies for parsing/analysis)

## Code Style

- Follow Go conventions
- Use meaningful variable names
- Add comments for exported functions
- Handle errors explicitly (no bare `if err != nil { return err }` without context)
