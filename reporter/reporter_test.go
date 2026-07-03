package reporter

import (
	"strings"
	"testing"

	"github.com/EdgarOrtegaRamirez/crashforge/analyzer"
	"github.com/EdgarOrtegaRamirez/crashforge/parser"
)

func TestReportText(t *testing.T) {
	result := &analyzer.AnalysisResult{
		TotalErrors: 5,
		UniqueCount: 2,
		Groups: []*analyzer.ErrorGroup{
			{
				ErrorType: "panic",
				Message:   "nil pointer dereference",
				Count:     3,
				TopFrames: []analyzer.FrameStat{
					{Function: "main", File: "app.go", Line: 42},
				},
			},
		},
		TopFiles: []analyzer.FileStat{
			{File: "app.go", ErrorCount: 3},
		},
	}

	r := New(FormatText, false)
	output := r.Report(result)

	if !strings.Contains(output, "Total Errors: 5") {
		t.Error("expected total errors in output")
	}
	if !strings.Contains(output, "nil pointer dereference") {
		t.Error("expected error message in output")
	}
}

func TestReportJSON(t *testing.T) {
	result := &analyzer.AnalysisResult{
		TotalErrors: 1,
		UniqueCount: 1,
	}

	r := New(FormatJSON, false)
	output := r.Report(result)

	if !strings.Contains(output, "total_errors") {
		t.Error("expected JSON format")
	}
}

func TestReportMarkdown(t *testing.T) {
	result := &analyzer.AnalysisResult{
		TotalErrors: 1,
		UniqueCount: 1,
	}

	r := New(FormatMarkdown, false)
	output := r.Report(result)

	if !strings.Contains(output, "# CrashForge Analysis Report") {
		t.Error("expected markdown header")
	}
}

func TestReportSingle(t *testing.T) {
	info := &parser.ErrorInfo{
		ErrorType: "panic",
		Language:  parser.LanguageGo,
		Message:   "nil pointer dereference",
		Hash:      "abc123",
		Frames: []parser.StackFrame{
			{Function: "main", File: "app.go", Line: 42},
		},
	}

	r := New(FormatText, false)
	output := r.ReportSingle(info)

	if !strings.Contains(output, "Error Type: panic") {
		t.Error("expected error type in output")
	}
	if !strings.Contains(output, "Language:") {
		t.Error("expected language in output")
	}
	if !strings.Contains(output, "#0 main()") {
		t.Error("expected stack frame in output")
	}
}
