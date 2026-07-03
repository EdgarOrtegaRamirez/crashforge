// Package reporter generates reports from analysis results.
package reporter

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/EdgarOrtegaRamirez/crashforge/analyzer"
	"github.com/EdgarOrtegaRamirez/crashforge/parser"
)

// Format represents the output format.
type Format string

const (
	FormatText     Format = "text"
	FormatJSON     Format = "json"
	FormatMarkdown Format = "markdown"
)

// Reporter generates reports from analysis results.
type Reporter struct {
	format    Format
	verbose   bool
	analyzer  *analyzer.Analyzer
}

// New creates a new Reporter.
func New(format Format, verbose bool) *Reporter {
	return &Reporter{
		format:   format,
		verbose:  verbose,
		analyzer: analyzer.New(),
	}
}

// Report generates a report from an analysis result.
func (r *Reporter) Report(result *analyzer.AnalysisResult) string {
	switch r.format {
	case FormatJSON:
		return r.reportJSON(result)
	case FormatMarkdown:
		return r.reportMarkdown(result)
	default:
		return r.reportText(result)
	}
}

func (r *Reporter) reportJSON(result *analyzer.AnalysisResult) string {
	data, _ := json.MarshalIndent(result, "", "  ")
	return string(data)
}

func (r *Reporter) reportText(result *analyzer.AnalysisResult) string {
	var b strings.Builder

	b.WriteString("╔══════════════════════════════════════════════════════════╗\n")
	b.WriteString("║              CrashForge Analysis Report                  ║\n")
	b.WriteString("╚══════════════════════════════════════════════════════════╝\n\n")

	b.WriteString(fmt.Sprintf("Total Errors: %d\n", result.TotalErrors))
	b.WriteString(fmt.Sprintf("Unique Errors: %d\n\n", result.UniqueCount))

	// Error groups
	if len(result.Groups) > 0 {
		b.WriteString("━━━ Error Groups (by frequency) ━━━\n\n")
		for i, group := range result.Groups {
			if i >= 10 && !r.verbose {
				b.WriteString(fmt.Sprintf("  ... and %d more groups (use --verbose to see all)\n", len(result.Groups)-10))
				break
			}
			b.WriteString(fmt.Sprintf("  [%d] %s (×%d)\n", i+1, group.ErrorType, group.Count))
			b.WriteString(fmt.Sprintf("      %s\n", truncate(group.Message, 80)))

			if len(group.TopFrames) > 0 {
				top := group.TopFrames[0]
				b.WriteString(fmt.Sprintf("      → %s:%d in %s\n", top.File, top.Line, top.Function))
			}

			// Root cause suggestions
			suggestions := r.analyzer.SuggestRootCause(group)
			if len(suggestions) > 0 {
				b.WriteString("      💡 Suggestion: " + suggestions[0] + "\n")
			}
			b.WriteString("\n")
		}
	}

	// Top files
	if len(result.TopFiles) > 0 {
		b.WriteString("━━━ Top Files with Errors ━━━\n\n")
		for i, fs := range result.TopFiles {
			if i >= 5 && !r.verbose {
				break
			}
			bar := strings.Repeat("█", min(fs.ErrorCount, 30))
			b.WriteString(fmt.Sprintf("  %s %d\n", fs.File, fs.ErrorCount))
			b.WriteString(fmt.Sprintf("    %s\n\n", bar))
		}
	}

	// Error type distribution
	if len(result.TopErrors) > 0 {
		b.WriteString("━━━ Error Type Distribution ━━━\n\n")
		for _, es := range result.TopErrors {
			pct := float64(es.Count) / float64(result.TotalErrors) * 100
			bar := strings.Repeat("█", int(pct/3))
			b.WriteString(fmt.Sprintf("  %-20s %d (%.1f%%)\n", es.ErrorType, es.Count, pct))
			b.WriteString(fmt.Sprintf("    %s\n", bar))
		}
	}

	return b.String()
}

func (r *Reporter) reportMarkdown(result *analyzer.AnalysisResult) string {
	var b strings.Builder

	b.WriteString("# CrashForge Analysis Report\n\n")
	b.WriteString(fmt.Sprintf("- **Total Errors:** %d\n", result.TotalErrors))
	b.WriteString(fmt.Sprintf("- **Unique Errors:** %d\n\n", result.UniqueCount))

	if len(result.Groups) > 0 {
		b.WriteString("## Error Groups\n\n")
		b.WriteString("| # | Type | Count | Message | Location |\n")
		b.WriteString("|---|------|-------|---------|----------|\n")
		for i, group := range result.Groups {
			if i >= 10 && !r.verbose {
				break
			}
			loc := ""
			if len(group.TopFrames) > 0 {
				top := group.TopFrames[0]
				loc = fmt.Sprintf("%s:%d", top.File, top.Line)
			}
			b.WriteString(fmt.Sprintf("| %d | %s | %d | %s | %s |\n",
				i+1, group.ErrorType, group.Count, truncate(group.Message, 50), loc))
		}
		b.WriteString("\n")
	}

	if len(result.TopFiles) > 0 {
		b.WriteString("## Top Files\n\n")
		for i, fs := range result.TopFiles {
			if i >= 5 && !r.verbose {
				break
			}
			b.WriteString(fmt.Sprintf("- `%s` — %d errors\n", fs.File, fs.ErrorCount))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// ReportSingle generates a report for a single parsed error.
func (r *Reporter) ReportSingle(info *parser.ErrorInfo) string {
	switch r.format {
	case FormatJSON:
		data, _ := json.MarshalIndent(info, "", "  ")
		return string(data)
	case FormatMarkdown:
		return r.singleMarkdown(info)
	default:
		return r.singleText(info)
	}
}

func (r *Reporter) singleText(info *parser.ErrorInfo) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Error Type: %s\n", info.ErrorType))
	b.WriteString(fmt.Sprintf("Language:   %s\n", info.Language))
	b.WriteString(fmt.Sprintf("Message:    %s\n", info.Message))
	b.WriteString(fmt.Sprintf("Hash:       %s\n\n", info.Hash))

	if len(info.Frames) > 0 {
		b.WriteString("Stack Trace:\n")
		for i, frame := range info.Frames {
			b.WriteString(fmt.Sprintf("  #%d %s()\n", i, frame.Function))
			b.WriteString(fmt.Sprintf("      %s:%d\n", frame.File, frame.Line))
			if frame.Package != "" {
				b.WriteString(fmt.Sprintf("      package: %s\n", frame.Package))
			}
		}
	}

	return b.String()
}

func (r *Reporter) singleMarkdown(info *parser.ErrorInfo) string {
	var b strings.Builder
	b.WriteString("# Error Details\n\n")
	b.WriteString(fmt.Sprintf("- **Type:** %s\n", info.ErrorType))
	b.WriteString(fmt.Sprintf("- **Language:** %s\n", info.Language))
	b.WriteString(fmt.Sprintf("- **Message:** %s\n", info.Message))
	b.WriteString(fmt.Sprintf("- **Hash:** `%s`\n\n", info.Hash))

	if len(info.Frames) > 0 {
		b.WriteString("## Stack Trace\n\n")
		b.WriteString("```")
		b.WriteString(string(info.Language))
		b.WriteString("\n")
		for i, frame := range info.Frames {
			b.WriteString(fmt.Sprintf("#%d %s()\n", i, frame.Function))
			b.WriteString(fmt.Sprintf("    %s:%d\n", frame.File, frame.Line))
		}
		b.WriteString("```\n")
	}

	return b.String()
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
