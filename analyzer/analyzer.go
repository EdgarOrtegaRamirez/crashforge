// Package analyzer provides error grouping, frequency analysis, and root cause detection.
package analyzer

import (
	"sort"
	"strings"

	"github.com/EdgarOrtegaRamirez/crashforge/parser"
)

// ErrorGroup represents a group of similar errors.
type ErrorGroup struct {
	ID         string              `json:"id"`
	ErrorType  string              `json:"error_type"`
	Message    string              `json:"message"`
	Count      int                 `json:"count"`
	FirstSeen  string              `json:"first_seen,omitempty"`
	LastSeen   string              `json:"last_seen,omitempty"`
	Errors     []*parser.ErrorInfo `json:"errors"`
	TopFrames  []FrameStat         `json:"top_frames"`
	Sample     *parser.ErrorInfo   `json:"sample"`
}

// FrameStat represents statistics for a stack frame.
type FrameStat struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Count    int    `json:"count"`
}

// AnalysisResult contains the full analysis of error logs.
type AnalysisResult struct {
	TotalErrors int          `json:"total_errors"`
	UniqueCount int          `json:"unique_count"`
	Groups      []*ErrorGroup `json:"groups"`
	TopFiles    []FileStat    `json:"top_files"`
	TopErrors   []ErrorStat   `json:"top_errors"`
}

// FileStat represents file error frequency.
type FileStat struct {
	File      string `json:"file"`
	ErrorCount int   `json:"error_count"`
}

// ErrorStat represents error type frequency.
type ErrorStat struct {
	ErrorType string `json:"error_type"`
	Count     int    `json:"count"`
}

// Analyzer analyzes parsed error information.
type Analyzer struct{}

// New creates a new Analyzer.
func New() *Analyzer {
	return &Analyzer{}
}

// Analyze performs a full analysis on a set of parsed errors.
func (a *Analyzer) Analyze(errors []*parser.ErrorInfo) *AnalysisResult {
	if len(errors) == 0 {
		return &AnalysisResult{}
	}

	result := &AnalysisResult{
		TotalErrors: len(errors),
		Groups:      a.groupErrors(errors),
	}
	result.UniqueCount = len(result.Groups)
	result.TopFiles = a.findTopFiles(errors)
	result.TopErrors = a.findTopErrors(errors)

	return result
}

// groupErrors groups similar errors together.
func (a *Analyzer) groupErrors(errors []*parser.ErrorInfo) []*ErrorGroup {
	groupMap := make(map[string]*ErrorGroup)
	var groupOrder []string

	for _, err := range errors {
		key := a.errorKey(err)
		if _, exists := groupMap[key]; !exists {
			groupMap[key] = &ErrorGroup{
				ID:        err.Hash,
				ErrorType: err.ErrorType,
				Message:   err.Message,
				Errors:    []*parser.ErrorInfo{},
			}
			groupOrder = append(groupOrder, key)
		}
		group := groupMap[key]
		group.Count++
		group.Errors = append(group.Errors, err)
	}

	// Build groups and compute statistics
	var groups []*ErrorGroup
	for _, key := range groupOrder {
		group := groupMap[key]
		group.Sample = group.Errors[0]
		group.TopFrames = a.computeFrameStats(group.Errors)
		groups = append(groups, group)
	}

	// Sort by count (most frequent first)
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Count > groups[j].Count
	})

	return groups
}

// errorKey generates a grouping key for an error.
func (a *Analyzer) errorKey(err *parser.ErrorInfo) string {
	parts := []string{err.ErrorType, err.Message}
	if len(err.Frames) > 0 {
		parts = append(parts, parser.FrameKey(err.Frames[0]))
	}
	return strings.Join(parts, "||")
}

// computeFrameStats computes statistics for stack frames.
func (a *Analyzer) computeFrameStats(errors []*parser.ErrorInfo) []FrameStat {
	frameCounts := make(map[string]*FrameStat)

	for _, err := range errors {
		seen := make(map[string]bool)
		for _, frame := range err.Frames {
			key := parser.FrameKey(frame)
			if seen[key] {
				continue
			}
			seen[key] = true

			if _, exists := frameCounts[key]; !exists {
				frameCounts[key] = &FrameStat{
					Function: frame.Function,
					File:     frame.File,
					Line:     frame.Line,
					Count:    0,
				}
			}
			frameCounts[key].Count++
		}
	}

	var stats []FrameStat
	for _, stat := range frameCounts {
		stats = append(stats, *stat)
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Count > stats[j].Count
	})

	if len(stats) > 5 {
		stats = stats[:5]
	}

	return stats
}

// findTopFiles finds files with the most errors.
func (a *Analyzer) findTopFiles(errors []*parser.ErrorInfo) []FileStat {
	fileCounts := make(map[string]int)
	for _, err := range errors {
		for _, frame := range err.Frames {
			if frame.File != "" {
				fileCounts[frame.File]++
			}
		}
	}

	var stats []FileStat
	for file, count := range fileCounts {
		stats = append(stats, FileStat{File: file, ErrorCount: count})
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].ErrorCount > stats[j].ErrorCount
	})

	if len(stats) > 10 {
		stats = stats[:10]
	}

	return stats
}

// findTopErrors finds the most common error types.
func (a *Analyzer) findTopErrors(errors []*parser.ErrorInfo) []ErrorStat {
	typeCounts := make(map[string]int)
	for _, err := range errors {
		t := err.ErrorType
		if t == "" {
			t = "unknown"
		}
		typeCounts[t]++
	}

	var stats []ErrorStat
	for t, count := range typeCounts {
		stats = append(stats, ErrorStat{ErrorType: t, Count: count})
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Count > stats[j].Count
	})

	return stats
}

// SuggestRootCause analyzes an error group and suggests potential root causes.
func (a *Analyzer) SuggestRootCause(group *ErrorGroup) []string {
	var suggestions []string

	msg := strings.ToLower(group.Message)
	errType := strings.ToLower(group.ErrorType)

	// Memory-related errors
	if strings.Contains(msg, "nil pointer") || strings.Contains(msg, "null") || strings.Contains(msg, "nil") {
		suggestions = append(suggestions, "Nil/null dereference: Check for uninitialized pointers or missing null checks")
	}
	if strings.Contains(msg, "out of memory") || strings.Contains(msg, "oom") || strings.Contains(msg, "cannot allocate") {
		suggestions = append(suggestions, "Memory exhaustion: Check for memory leaks or excessive allocation")
	}
	if strings.Contains(msg, "stack overflow") || strings.Contains(msg, "recursion") {
		suggestions = append(suggestions, "Infinite recursion: Check for missing base cases in recursive functions")
	}

	// File/resource errors
	if strings.Contains(msg, "no such file") || strings.Contains(msg, "file not found") || strings.Contains(msg, "enoent") {
		suggestions = append(suggestions, "Missing file: Verify file paths and ensure files exist before access")
	}
	if strings.Contains(msg, "permission denied") || strings.Contains(msg, "eacces") {
		suggestions = append(suggestions, "Permission error: Check file permissions and user access rights")
	}
	if strings.Contains(msg, "too many open") || strings.Contains(msg, "emfile") {
		suggestions = append(suggestions, "File descriptor leak: Check for unclosed file handles or connections")
	}

	// Network errors
	if strings.Contains(msg, "connection refused") || strings.Contains(msg, "econnrefused") {
		suggestions = append(suggestions, "Connection refused: Verify the target service is running and accessible")
	}
	if strings.Contains(msg, "timeout") || strings.Contains(msg, "deadline exceeded") {
		suggestions = append(suggestions, "Timeout: Check network connectivity and increase timeout values")
	}
	if strings.Contains(msg, "broken pipe") || strings.Contains(msg, "epipe") {
		suggestions = append(suggestions, "Broken pipe: The remote end closed the connection unexpectedly")
	}

	// Type/conversion errors
	if strings.Contains(msg, "type mismatch") || strings.Contains(msg, "cannot convert") || strings.Contains(errType, "type") {
		suggestions = append(suggestions, "Type error: Verify input types match expected types")
	}
	if strings.Contains(msg, "index out of range") || strings.Contains(msg, "range") {
		suggestions = append(suggestions, "Index out of bounds: Add bounds checking before array/slice access")
	}
	if strings.Contains(msg, "division by zero") {
		suggestions = append(suggestions, "Division by zero: Add zero-check before division operations")
	}

	// Concurrency errors
	if strings.Contains(msg, "deadlock") || strings.Contains(msg, "channel") {
		suggestions = append(suggestions, "Deadlock: Check for circular wait conditions in goroutines/channels")
	}
	if strings.Contains(msg, "race condition") || strings.Contains(msg, "data race") {
		suggestions = append(suggestions, "Data race: Add proper synchronization (mutexes, channels)")
	}

	// Generic suggestions based on frame analysis
	if len(group.TopFrames) > 0 {
		topFrame := group.TopFrames[0]
		suggestions = append(suggestions, "Most frequent crash location: "+topFrame.File+":"+itoa(topFrame.Line)+" in "+topFrame.Function)
	}

	if len(suggestions) == 0 {
		suggestions = append(suggestions, "No specific root cause identified. Review the stack trace and error message for clues.")
	}

	return suggestions
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}
