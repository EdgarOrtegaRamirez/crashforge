# CrashForge

A crash and error log analyzer CLI that parses stack traces from multiple programming languages, groups similar errors, and provides actionable root cause suggestions.

## Features

- **Multi-language support**: Parses stack traces from Go, Python, JavaScript, Java, Rust, and C
- **Error grouping**: Groups similar errors by type, message, and stack trace
- **Root cause analysis**: Suggests potential root causes based on error patterns
- **Frequency analysis**: Identifies the most common errors and affected files
- **Multiple output formats**: Text, JSON, and Markdown reports
- **Batch analysis**: Analyze multiple stack traces from a single log file

## Installation

```bash
go install github.com/EdgarOrtegaRamirez/crashforge/cmd/crashforge@latest
```

Or build from source:

```bash
git clone https://github.com/EdgarOrtegaRamirez/crashforge
cd crashforge
go build -o crashforge ./cmd/crashforge/
```

## Quick Start

### Parse a single stack trace

```bash
echo 'panic: runtime error: nil pointer dereference

goroutine 1 [running]:
main.processA()
	/tmp/app.go:42 +0x1a
main.main()
	/tmp/app.go:10 +0x1a' | crashforge parse
```

Output:
```
Error Type: panic
Language:   go
Message:    runtime error: nil pointer dereference
Hash:       4bdfb8a60690

Stack Trace:
  #0 main.processA()
      /tmp/app.go:42
      package: main
  #1 main.main()
      /tmp/app.go:10
      package: main
```

### Analyze a log file

```bash
crashforge analyze /var/log/app.log
```

Output:
```
╔══════════════════════════════════════════════════════════╗
║              CrashForge Analysis Report                  ║
╚══════════════════════════════════════════════════════════╝

Total Errors: 150
Unique Errors: 12

━━━ Error Groups (by frequency) ━━━

  [1] panic (×45)
      runtime error: nil pointer dereference
      → /app/server.go:142 in handleRequest
      💡 Suggestion: Nil/null dereference: Check for uninitialized pointers

  [2] TypeError (×32)
      Cannot read property 'map' of undefined
      → /app/client.js:85 in processItems
      💡 Suggestion: Type error: Verify input types match expected types
```

### JSON output

```bash
crashforge analyze --format json /var/log/app.log
```

### Markdown report

```bash
crashforge analyze --format markdown /var/log/app.log > report.md
```

## Supported Languages

| Language | Error Format | Example |
|----------|--------------|---------|
| Go | `panic: message` | `panic: runtime error: nil pointer dereference` |
| Python | `Traceback (most recent call last):` | `KeyError: 'key'` |
| JavaScript | `TypeError: message` | `TypeError: Cannot read property 'x' of undefined` |
| Java | `java.lang.Exception: message` | `java.lang.NullPointerException: null` |
| Rust | `thread 'main' panicked at 'message'` | `thread 'main' panicked at 'assertion failed'` |
| C | `Segmentation fault` | `Segmentation fault (core dumped)` |

## Commands

### `crashforge parse [file]`

Parse a single stack trace from a file or stdin.

**Options:**
- `-f, --format`: Output format (text, json, markdown)
- `-v, --verbose`: Verbose output

### `crashforge analyze [file]`

Analyze multiple stack traces from a log file.

**Options:**
- `-f, --format`: Output format (text, json, markdown)
- `-v, --verbose`: Show all error groups (default: top 10)

### `crashforge watch [file]`

Watch a log file for new errors (coming soon).

### `crashforge version`

Print version information.

## Architecture

```
crashforge/
├── parser/          # Stack trace parsing for multiple languages
│   ├── parser.go    # Core parser and data structures
│   └── languages.go # Language-specific parsers
├── analyzer/        # Error grouping and analysis
│   └── analyzer.go  # Grouping, frequency analysis, root cause suggestions
├── reporter/        # Report generation
│   └── reporter.go  # Text, JSON, Markdown formatters
├── cmd/
│   └── crashforge/  # CLI entry point
│       └── main.go
└── tests/           # Integration tests
```

## Root Cause Suggestions

CrashForge analyzes error patterns and suggests potential root causes:

| Pattern | Suggestion |
|---------|------------|
| Nil/null dereference | Check for uninitialized pointers or missing null checks |
| Index out of bounds | Add bounds checking before array/slice access |
| Division by zero | Add zero-check before division operations |
| File not found | Verify file paths and ensure files exist |
| Permission denied | Check file permissions and user access rights |
| Connection refused | Verify the target service is running |
| Timeout | Check network connectivity and increase timeout values |
| Type mismatch | Verify input types match expected types |
| Stack overflow | Check for missing base cases in recursive functions |
| Memory exhaustion | Check for memory leaks or excessive allocation |

## Testing

```bash
go test ./...
```

## License

MIT License - see [LICENSE](LICENSE) for details.
