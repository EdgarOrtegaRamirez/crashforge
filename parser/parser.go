// Package parser provides stack trace parsing for multiple languages.
package parser

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Language represents a programming language.
type Language string

const (
	LanguageGo         Language = "go"
	LanguagePython     Language = "python"
	LanguageJavaScript Language = "javascript"
	LanguageJava       Language = "java"
	LanguageRust       Language = "rust"
	LanguageC          Language = "c"
	LanguageUnknown    Language = "unknown"
)

// StackFrame represents a single frame in a stack trace.
type StackFrame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Column   int    `json:"column,omitempty"`
	Package  string `json:"package,omitempty"`
	Raw      string `json:"raw"`
}

// ErrorInfo represents parsed error information.
type ErrorInfo struct {
	Message   string       `json:"message"`
	ErrorType string       `json:"error_type,omitempty"`
	Language  Language     `json:"language"`
	Frames    []StackFrame `json:"frames"`
	RawText   string       `json:"raw_text"`
	Hash      string       `json:"hash,omitempty"`
}

// Parser parses stack traces from various languages.
type Parser struct {
	parsers []languageParser
}

type languageParser struct {
	language Language
	pattern  *regexp.Regexp
	parse    func(matches []string, text string) *ErrorInfo
}

// New creates a new Parser with all supported language parsers.
func New() *Parser {
	p := &Parser{}
	p.registerGo()
	p.registerPython()
	p.registerJavaScript()
	p.registerJava()
	p.registerRust()
	p.registerC()
	return p
}

// Parse attempts to parse a stack trace from any supported language.
func (p *Parser) Parse(text string) (*ErrorInfo, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("empty input")
	}

	for _, lp := range p.parsers {
		if matches := lp.pattern.FindStringSubmatch(text); matches != nil {
			info := lp.parse(matches, text)
			if info != nil && len(info.Frames) > 0 {
				info.Language = lp.language
				info.Hash = computeHash(info)
				return info, nil
			}
		}
	}

	return &ErrorInfo{
		Message:  extractErrorMessage(text),
		Language: LanguageUnknown,
		Frames:   []StackFrame{},
		RawText:  text,
		Hash:     simpleHash(text),
	}, nil
}

// ParseMultiple parses multiple stack traces from a log file.
func (p *Parser) ParseMultiple(text string) []*ErrorInfo {
	var results []*ErrorInfo

	// Try to split by language-specific patterns
	blocks := p.splitByLanguage(text)

	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}
		info, err := p.Parse(block)
		if err == nil && info != nil {
			results = append(results, info)
		}
	}

	return results
}

// splitByLanguage splits text by language-specific error markers.
func (p *Parser) splitByLanguage(text string) []string {
	// Define splitting patterns for each language
	splitPatterns := []*regexp.Regexp{
		// Go: panic at the start of a line
		regexp.MustCompile(`(?m)^panic:\s`),
		// Python: Traceback at the start
		regexp.MustCompile(`(?m)^Traceback \(most recent call last\):`),
		// Java: java.lang.Exception at the start
		regexp.MustCompile(`(?m)^java\.\w+(?:\.\w+)*(?:Exception|Error):`),
		// JavaScript: Error at the start (after optional "Uncaught ")
		regexp.MustCompile(`(?m)^(?:Uncaught )?(?:Syntax|Type|Reference|Range|URI|Eval|Internal)Error`),
		// Rust: thread panic
		regexp.MustCompile(`(?m)^thread '.+?' panicked at`),
	}

	// Find all split points
	type splitPoint struct {
		index int
		line  int
	}
	var points []splitPoint

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		for _, pat := range splitPatterns {
			if pat.MatchString(line) {
				// Calculate character index
				idx := 0
				for j := 0; j < i; j++ {
					idx += len(lines[j]) + 1 // +1 for \n
				}
				points = append(points, splitPoint{index: idx, line: i})
				break
			}
		}
	}

	if len(points) == 0 {
		return []string{text}
	}

	// Split at the points
	var blocks []string
	for i, pt := range points {
		end := len(text)
		if i+1 < len(points) {
			end = points[i+1].index
		}
		blocks = append(blocks, text[pt.index:end])
	}

	return blocks
}

func extractErrorMessage(text string) string {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			if len(line) > 200 {
				return line[:200] + "..."
			}
			return line
		}
	}
	return "Unknown error"
}

// SortFrames sorts stack frames by line number.
func SortFrames(frames []StackFrame) {
	sort.Slice(frames, func(i, j int) bool {
		return frames[i].Line < frames[j].Line
	})
}

// FrameKey returns a unique key for a frame.
func FrameKey(f StackFrame) string {
	return f.File + ":" + f.Function
}

func computeHash(info *ErrorInfo) string {
	h := sha256.New()
	h.Write([]byte(info.ErrorType))
	h.Write([]byte(info.Message))
	for _, f := range info.Frames {
		h.Write([]byte(f.File))
		h.Write([]byte(f.Function))
	}
	return fmt.Sprintf("%x", h.Sum(nil))[:12]
}

func simpleHash(text string) string {
	h := sha256.New()
	h.Write([]byte(text))
	return fmt.Sprintf("%x", h.Sum(nil))[:12]
}
