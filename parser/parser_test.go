package parser

import (
	"testing"
)

func TestParseGoNilPointer(t *testing.T) {
	input := `panic: runtime error: nil pointer dereference

goroutine 1 [running]:
main.processA()
	/tmp/app.go:42 +0x1a
main.main()
	/tmp/app.go:10 +0x1a`

	p := New()
	info, err := p.Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.Language != LanguageGo {
		t.Errorf("expected language go, got %s", info.Language)
	}
	if info.ErrorType != "panic" {
		t.Errorf("expected error type panic, got %s", info.ErrorType)
	}
	if info.Message != "runtime error: nil pointer dereference" {
		t.Errorf("unexpected message: %s", info.Message)
	}
	if len(info.Frames) != 2 {
		t.Errorf("expected 2 frames, got %d", len(info.Frames))
	}
}

func TestParseGoIndexOutOfRange(t *testing.T) {
	input := `panic: runtime error: index out of range [3] with length 3

goroutine 1 [running]:
main.sliceAccess()
	/tmp/app.go:25 +0x2a`

	p := New()
	info, err := p.Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.Language != LanguageGo {
		t.Errorf("expected language go, got %s", info.Language)
	}
	if info.ErrorType != "panic" {
		t.Errorf("expected error type panic, got %s", info.ErrorType)
	}
	if len(info.Frames) != 1 {
		t.Errorf("expected 1 frame, got %d", len(info.Frames))
	}
	if info.Frames[0].Function != "main.sliceAccess" {
		t.Errorf("expected function main.sliceAccess, got %s", info.Frames[0].Function)
	}
}

func TestParseGoPackageExtraction(t *testing.T) {
	input := `panic: test error

goroutine 1 [running]:
github.com/example/pkg.MyFunc()
	/tmp/app.go:10 +0x5`

	p := New()
	info, err := p.Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.Frames[0].Package != "github.com/example/pkg" {
		t.Errorf("expected package github.com/example/pkg, got %s", info.Frames[0].Package)
	}
}

func TestParsePythonKeyError(t *testing.T) {
	input := `Traceback (most recent call last):
  File "app.py", line 15, in process_data
    result = data["key"]
KeyError: 'key'`

	p := New()
	info, err := p.Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.Language != LanguagePython {
		t.Errorf("expected language python, got %s", info.Language)
	}
	if info.ErrorType != "KeyError" {
		t.Errorf("expected error type KeyError, got %s", info.ErrorType)
	}
	if len(info.Frames) != 1 {
		t.Errorf("expected 1 frame, got %d", len(info.Frames))
	}
}

func TestParsePythonAttributeError(t *testing.T) {
	input := `Traceback (most recent call last):
  File "main.py", line 5, in <module>
    obj.nonexistent()
AttributeError: 'NoneType' object has no attribute 'nonexistent'`

	p := New()
	info, err := p.Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.ErrorType != "AttributeError" {
		t.Errorf("expected error type AttributeError, got %s", info.ErrorType)
	}
}

func TestParseJavaScriptTypeError(t *testing.T) {
	input := `TypeError: Cannot read properties of undefined (reading 'map')
    at processItems (/app/src/utils.js:15:20)
    at Object.handle (/app/src/handler.js:42:5)`

	p := New()
	info, err := p.Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.Language != LanguageJavaScript {
		t.Errorf("expected language javascript, got %s", info.Language)
	}
	if info.ErrorType != "TypeError" {
		t.Errorf("expected error type TypeError, got %s", info.ErrorType)
	}
	if len(info.Frames) != 2 {
		t.Errorf("expected 2 frames, got %d", len(info.Frames))
	}
}

func TestParseJavaScriptReferenceError(t *testing.T) {
	input := `ReferenceError: myVar is not defined
    at /app/src/index.js:10:5`

	p := New()
	info, err := p.Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.ErrorType != "ReferenceError" {
		t.Errorf("expected error type ReferenceError, got %s", info.ErrorType)
	}
}

func TestParseJavaNullPointerException(t *testing.T) {
	input := `java.lang.NullPointerException: Cannot invoke method on null
	at com.example.MyClass.process(MyClass.java:42)
	at com.example.Main.run(Main.java:15)
	at com.example.App.main(App.java:8)`

	p := New()
	info, err := p.Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.Language != LanguageJava {
		t.Errorf("expected language java, got %s", info.Language)
	}
	if info.ErrorType != "java.lang.NullPointerException" {
		t.Errorf("expected error type java.lang.NullPointerException, got %s", info.ErrorType)
	}
	if len(info.Frames) != 3 {
		t.Errorf("expected 3 frames, got %d", len(info.Frames))
	}
	if info.Frames[0].Package != "com.example.MyClass" {
		t.Errorf("expected package com.example.MyClass, got %s", info.Frames[0].Package)
	}
	if info.Frames[0].Function != "process" {
		t.Errorf("expected function process, got %s", info.Frames[0].Function)
	}
}

func TestParseJavaArrayIndexOutOfBounds(t *testing.T) {
	input := `java.lang.ArrayIndexOutOfBoundsException: Index 5 out of bounds for length 3
	at java.base/java.util.Arrays.checkIndex(Arrays.java:455)
	at com.example.App.main(App.java:20)`

	p := New()
	info, err := p.Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.ErrorType != "java.lang.ArrayIndexOutOfBoundsException" {
		t.Errorf("expected error type, got %s", info.ErrorType)
	}
}

func TestParseRustPanic(t *testing.T) {
	input := `thread 'main' panicked at 'assertion failed: left == right', src/main.rs:42:5`

	p := New()
	info, err := p.Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.Language != LanguageRust {
		t.Errorf("expected language rust, got %s", info.Language)
	}
	if info.ErrorType != "panic" {
		t.Errorf("expected error type panic, got %s", info.ErrorType)
	}
	if info.Message != "assertion failed: left == right" {
		t.Errorf("unexpected message: %s", info.Message)
	}
}

func TestParseMultipleGoPanics(t *testing.T) {
	input := `panic: first error

goroutine 1 [running]:
main.func1()
	/tmp/a.go:1 +0x0

panic: second error

goroutine 1 [running]:
main.func2()
	/tmp/b.go:2 +0x0

panic: third error

goroutine 1 [running]:
main.func3()
	/tmp/c.go:3 +0x0`

	p := New()
	errors := p.ParseMultiple(input)

	if len(errors) != 3 {
		t.Fatalf("expected 3 errors, got %d", len(errors))
	}

	if errors[0].Message != "first error" {
		t.Errorf("expected first error message, got %s", errors[0].Message)
	}
	if errors[1].Message != "second error" {
		t.Errorf("expected second error message, got %s", errors[1].Message)
	}
	if errors[2].Message != "third error" {
		t.Errorf("expected third error message, got %s", errors[2].Message)
	}
}

func TestParseEmpty(t *testing.T) {
	p := New()
	_, err := p.Parse("")
	if err == nil {
		t.Error("expected error for empty input")
	}
}

func TestParseUnknownFormat(t *testing.T) {
	p := New()
	info, err := p.Parse("some random text that doesn't match any pattern")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.Language != LanguageUnknown {
		t.Errorf("expected language unknown, got %s", info.Language)
	}
	if info.Message != "some random text that doesn't match any pattern" {
		t.Errorf("unexpected message: %s", info.Message)
	}
}

func TestFrameKey(t *testing.T) {
	frame := StackFrame{
		File:     "app.go",
		Function: "main",
	}

	key := FrameKey(frame)
	if key != "app.go:main" {
		t.Errorf("expected app.go:main, got %s", key)
	}
}

func TestSortFrames(t *testing.T) {
	frames := []StackFrame{
		{Line: 42},
		{Line: 10},
		{Line: 25},
	}

	SortFrames(frames)

	if frames[0].Line != 10 || frames[1].Line != 25 || frames[2].Line != 42 {
		t.Errorf("frames not sorted correctly: %v", frames)
	}
}
