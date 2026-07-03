package analyzer

import (
	"testing"

	"github.com/EdgarOrtegaRamirez/crashforge/parser"
)

func TestAnalyzeGroupsErrors(t *testing.T) {
	errors := []*parser.ErrorInfo{
		{
			ErrorType: "panic",
			Message:   "nil pointer",
			Frames:    []parser.StackFrame{{File: "a.go", Function: "main"}},
		},
		{
			ErrorType: "panic",
			Message:   "nil pointer",
			Frames:    []parser.StackFrame{{File: "a.go", Function: "main"}},
		},
		{
			ErrorType: "panic",
			Message:   "index out of range",
			Frames:    []parser.StackFrame{{File: "b.go", Function: "other"}},
		},
	}

	a := New()
	result := a.Analyze(errors)

	if result.TotalErrors != 3 {
		t.Errorf("expected 3 total errors, got %d", result.TotalErrors)
	}
	if result.UniqueCount != 2 {
		t.Errorf("expected 2 unique errors, got %d", result.UniqueCount)
	}
	if len(result.Groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(result.Groups))
	}
	if result.Groups[0].Count != 2 {
		t.Errorf("expected first group count 2, got %d", result.Groups[0].Count)
	}
}

func TestAnalyzeEmpty(t *testing.T) {
	a := New()
	result := a.Analyze(nil)

	if result.TotalErrors != 0 {
		t.Errorf("expected 0 total errors, got %d", result.TotalErrors)
	}
	if result.UniqueCount != 0 {
		t.Errorf("expected 0 unique errors, got %d", result.UniqueCount)
	}
}

func TestSuggestRootCause(t *testing.T) {
	a := New()
	group := &ErrorGroup{
		ErrorType: "panic",
		Message:   "runtime error: nil pointer dereference",
	}

	suggestions := a.SuggestRootCause(group)
	if len(suggestions) == 0 {
		t.Error("expected at least one suggestion")
	}

	found := false
	for _, s := range suggestions {
		if contains(s, "Nil/null dereference") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected nil dereference suggestion, got: %v", suggestions)
	}
}

func TestFindTopFiles(t *testing.T) {
	a := New()
	errors := []*parser.ErrorInfo{
		{Frames: []parser.StackFrame{{File: "a.go"}}},
		{Frames: []parser.StackFrame{{File: "a.go"}}},
		{Frames: []parser.StackFrame{{File: "b.go"}}},
	}

	topFiles := a.findTopFiles(errors)
	if len(topFiles) == 0 {
		t.Fatal("expected top files")
	}
	if topFiles[0].File != "a.go" {
		t.Errorf("expected top file a.go, got %s", topFiles[0].File)
	}
	if topFiles[0].ErrorCount != 2 {
		t.Errorf("expected count 2, got %d", topFiles[0].ErrorCount)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
