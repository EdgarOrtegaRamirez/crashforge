package parser

import (
	"regexp"
	"strconv"
	"strings"
)

func (p *Parser) registerGo() {
	// Match Go panic message on a single line
	pattern := regexp.MustCompile(`(?m)^panic:\s*(.+)$`)

	p.parsers = append(p.parsers, languageParser{
		language: LanguageGo,
		pattern:  pattern,
		parse: func(matches []string, text string) *ErrorInfo {
			info := &ErrorInfo{RawText: text}
			info.Message = strings.TrimSpace(matches[1])
			info.ErrorType = "panic"
			info.Frames = parseGoFrames(text)
			return info
		},
	})
}

func parseGoFrames(text string) []StackFrame {
	var frames []StackFrame

	// Match function() followed by file:line
	framePattern := regexp.MustCompile(`(?m)^([\S]+)\(\)\s*\n\s+(.+?):(\d+)(?:\s+\+0x[0-9a-f]+)?`)
	matches := framePattern.FindAllStringSubmatch(text, -1)

	for _, m := range matches {
		frame := StackFrame{
			Function: m[1],
			File:     m[2],
			Raw:      m[0],
		}
		frame.Line, _ = strconv.Atoi(m[3])

		if idx := strings.LastIndex(frame.Function, "."); idx > 0 {
			frame.Package = frame.Function[:idx]
		}
		frames = append(frames, frame)
	}

	return frames
}

func (p *Parser) registerPython() {
	pattern := regexp.MustCompile(`(?s)Traceback \(most recent call last\):(.+)$`)

	p.parsers = append(p.parsers, languageParser{
		language: LanguagePython,
		pattern:  pattern,
		parse: func(matches []string, text string) *ErrorInfo {
			info := &ErrorInfo{RawText: text}
			info.Frames = parsePythonFrames(text)
			info.Message = extractPythonError(text)
			info.ErrorType = extractPythonErrorType(text)
			return info
		},
	})
}

func parsePythonFrames(text string) []StackFrame {
	var frames []StackFrame
	framePattern := regexp.MustCompile(`(?m)^\s+File "([^"]+)", line (\d+)(?:, in (\w+))?`)
	matches := framePattern.FindAllStringSubmatch(text, -1)
	for _, m := range matches {
		frame := StackFrame{
			File:     m[1],
			Function: m[3],
			Raw:      m[0],
		}
		frame.Line, _ = strconv.Atoi(m[2])
		frames = append(frames, frame)
	}
	return frames
}

func extractPythonError(text string) string {
	lines := strings.Split(text, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" && !strings.HasPrefix(line, "File ") && !strings.HasPrefix(line, "Traceback") {
			if len(line) > 200 {
				return line[:200] + "..."
			}
			return line
		}
	}
	return "Unknown Python error"
}

func extractPythonErrorType(text string) string {
	lines := strings.Split(text, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.Contains(line, ":") && !strings.HasPrefix(line, "File ") && !strings.HasPrefix(line, "Traceback") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) > 0 {
				errType := strings.TrimSpace(parts[0])
				if strings.Contains(errType, "Error") || strings.Contains(errType, "Exception") {
					return errType
				}
			}
		}
	}
	return ""
}

func (p *Parser) registerJavaScript() {
	pattern := regexp.MustCompile(`(?s)(?:Uncaught |Unhandled )?(Syntax|Type|Reference|Range|URI|Eval|Internal)Error[:\s]+(.+?)$`)

	p.parsers = append(p.parsers, languageParser{
		language: LanguageJavaScript,
		pattern:  pattern,
		parse: func(matches []string, text string) *ErrorInfo {
			info := &ErrorInfo{RawText: text}
			info.ErrorType = matches[1] + "Error"
			info.Message = strings.TrimSpace(matches[2])
			if idx := strings.Index(info.Message, "\n"); idx > 0 {
				info.Message = strings.TrimSpace(info.Message[:idx])
			}
			info.Frames = parseJSFrames(text)
			return info
		},
	})
}

func parseJSFrames(text string) []StackFrame {
	var frames []StackFrame
	// Node.js: at Function (file:line:column)
	pattern := regexp.MustCompile(`(?m)^\s+at\s+(\S+)\s+\(([^:]+):(\d+):(\d+)\)`)
	matches := pattern.FindAllStringSubmatch(text, -1)
	for _, m := range matches {
		frame := StackFrame{
			Function: m[1],
			File:     m[2],
			Raw:      m[0],
		}
		frame.Line, _ = strconv.Atoi(m[3])
		frame.Column, _ = strconv.Atoi(m[4])
		frames = append(frames, frame)
	}

	// Fallback: at file:line:column
	if len(frames) == 0 {
		simplePattern := regexp.MustCompile(`(?m)^\s+at\s+([^:]+):(\d+):(\d+)`)
		simpleMatches := simplePattern.FindAllStringSubmatch(text, -1)
		for _, m := range simpleMatches {
			frame := StackFrame{
				File: m[1],
				Raw:  m[0],
			}
			frame.Line, _ = strconv.Atoi(m[2])
			frame.Column, _ = strconv.Atoi(m[3])
			frames = append(frames, frame)
		}
	}

	return frames
}

func (p *Parser) registerJava() {
	pattern := regexp.MustCompile(`(?s)(java\.\w+(?:\.\w+)*(?:Exception|Error))\s*:\s*(.+?)$`)

	p.parsers = append(p.parsers, languageParser{
		language: LanguageJava,
		pattern:  pattern,
		parse: func(matches []string, text string) *ErrorInfo {
			info := &ErrorInfo{RawText: text}
			info.ErrorType = matches[1]
			info.Message = strings.TrimSpace(matches[2])
			if idx := strings.Index(info.Message, "\n"); idx > 0 {
				info.Message = strings.TrimSpace(info.Message[:idx])
			}
			info.Frames = parseJavaFrames(text)
			return info
		},
	})
}

func parseJavaFrames(text string) []StackFrame {
	var frames []StackFrame
	framePattern := regexp.MustCompile(`(?m)^\s+at\s+([\w.$]+)\(([^)]+)\)`)
	matches := framePattern.FindAllStringSubmatch(text, -1)
	for _, m := range matches {
		frame := StackFrame{
			Function: m[1],
			Raw:      m[0],
		}
		location := m[2]
		if location == "Native Method" || location == "Unknown Source" {
			frame.File = location
		} else {
			parts := strings.Split(location, ":")
			frame.File = parts[0]
			if len(parts) > 1 {
				frame.Line, _ = strconv.Atoi(parts[1])
			}
		}
		if idx := strings.LastIndex(frame.Function, "."); idx > 0 {
			frame.Package = frame.Function[:idx]
			frame.Function = frame.Function[idx+1:]
		}
		frames = append(frames, frame)
	}
	return frames
}

func (p *Parser) registerRust() {
	pattern := regexp.MustCompile(`(?s)thread '(.+?)' panicked at '(.+?)', (.+):(\d+)`)

	p.parsers = append(p.parsers, languageParser{
		language: LanguageRust,
		pattern:  pattern,
		parse: func(matches []string, text string) *ErrorInfo {
			info := &ErrorInfo{RawText: text}
			info.ErrorType = "panic"
			info.Message = matches[2]
			info.Frames = parseRustFrames(text, matches[3], matches[4])
			return info
		},
	})
}

func parseRustFrames(text, mainFile, mainLine string) []StackFrame {
	var frames []StackFrame
	framePattern := regexp.MustCompile(`(?m)^\s+\d+:\s+(.+?)\s+-\s+(.+?)\s+\(([^)]+)\)`)
	matches := framePattern.FindAllStringSubmatch(text, -1)
	for _, m := range matches {
		frame := StackFrame{
			Function: m[2],
			File:     m[3],
			Raw:      m[0],
		}
		frame.Line, _ = strconv.Atoi(m[3])
		if idx := strings.LastIndex(frame.Function, "::"); idx > 0 {
			frame.Package = frame.Function[:idx]
		}
		frames = append(frames, frame)
	}
	// Add panic location as first frame
	if mainFile != "" {
		line, _ := strconv.Atoi(mainLine)
		frames = append([]StackFrame{{
			Function: "panic",
			File:     mainFile,
			Line:     line,
			Raw:      "panicked at",
		}}, frames...)
	}
	return frames
}

func (p *Parser) registerC() {
	pattern := regexp.MustCompile(`(?s)(?:Segmentation fault|Abort|Signal \d+|Fatal)(.+?)$`)

	p.parsers = append(p.parsers, languageParser{
		language: LanguageC,
		pattern:  pattern,
		parse: func(matches []string, text string) *ErrorInfo {
			info := &ErrorInfo{RawText: text}
			info.ErrorType = "signal"
			info.Message = strings.TrimSpace(matches[1])
			if idx := strings.Index(info.Message, "\n"); idx > 0 {
				info.Message = strings.TrimSpace(info.Message[:idx])
			}
			info.Frames = parseCFrames(text)
			return info
		},
	})
}

func parseCFrames(text string) []StackFrame {
	var frames []StackFrame
	framePattern := regexp.MustCompile(`(?m)#\d+\s+0x[0-9a-f]+\s+in\s+(\S+)\s+\(([^)]*)\)\s+(?:at\s+)?([^:]+):(\d+)`)
	matches := framePattern.FindAllStringSubmatch(text, -1)
	for _, m := range matches {
		frame := StackFrame{
			Function: m[1],
			File:     m[3],
			Raw:      m[0],
		}
		frame.Line, _ = strconv.Atoi(m[4])
		frames = append(frames, frame)
	}
	return frames
}
