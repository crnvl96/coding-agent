package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEditFile_replacesUniqueOldStr(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	os.WriteFile(path, []byte("hello world"), 0o644)

	input, _ := json.Marshal(EditFileInput{Path: path, OldStr: "hello", NewStr: "hi"})
	result, err := EditFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "OK" {
		t.Errorf("expected OK, got %q", result)
	}

	content, _ := os.ReadFile(path)
	if string(content) != "hi world" {
		t.Errorf("expected 'hi world', got %q", string(content))
	}
}

func TestEditFile_oldStrNotFoundReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	os.WriteFile(path, []byte("hello world"), 0o644)

	input, _ := json.Marshal(EditFileInput{Path: path, OldStr: "xyz", NewStr: "abc"})
	_, err := EditFile(input)
	if err == nil {
		t.Fatal("expected error for old_str not found")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got %q", err.Error())
	}
}

func TestEditFile_multipleMatchesReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	os.WriteFile(path, []byte("ab ab ab"), 0o644)

	input, _ := json.Marshal(EditFileInput{Path: path, OldStr: "ab", NewStr: "cd"})
	_, err := EditFile(input)
	if err == nil {
		t.Fatal("expected error for multiple matches")
	}
	if !strings.Contains(err.Error(), "3 times") {
		t.Errorf("expected count in error, got %q", err.Error())
	}
}

func TestEditFile_appendsWhenOldStrEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	os.WriteFile(path, []byte("hello"), 0o644)

	input, _ := json.Marshal(EditFileInput{Path: path, OldStr: "", NewStr: " world"})
	result, err := EditFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "OK" {
		t.Errorf("expected OK, got %q", result)
	}

	content, _ := os.ReadFile(path)
	if string(content) != "hello world" {
		t.Errorf("expected 'hello world', got %q", string(content))
	}
}

func TestEditFile_createsNewFileWithEmptyOldStr(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "new.txt")

	input, _ := json.Marshal(EditFileInput{Path: path, OldStr: "", NewStr: "fresh content"})
	result, err := EditFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "Successfully created") {
		t.Errorf("expected creation message, got %q", result)
	}

	content, _ := os.ReadFile(path)
	if string(content) != "fresh content" {
		t.Errorf("expected 'fresh content', got %q", string(content))
	}
}

func TestEditFile_createsParentDirectories(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a", "b", "c.txt")

	input, _ := json.Marshal(EditFileInput{Path: path, OldStr: "", NewStr: "deep"})
	result, err := EditFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "Successfully created") {
		t.Errorf("expected creation message, got %q", result)
	}

	content, _ := os.ReadFile(path)
	if string(content) != "deep" {
		t.Errorf("expected 'deep', got %q", string(content))
	}
}

func TestEditFile_nonexistentFileWithNonEmptyOldStrReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nope.txt")

	input, _ := json.Marshal(EditFileInput{Path: path, OldStr: "x", NewStr: "y"})
	_, err := EditFile(input)
	if err == nil {
		t.Fatal("expected error for nonexistent file with non-empty old_str")
	}
}

func TestEditFile_invalidInputReturnsError(t *testing.T) {
	tests := []struct {
		name  string
		input EditFileInput
	}{
		{"empty path", EditFileInput{Path: "", OldStr: "a", NewStr: "b"}},
		{"old equals new", EditFileInput{Path: "f.txt", OldStr: "a", NewStr: "a"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			input, _ := json.Marshal(tc.input)
			_, err := EditFile(input)
			if err == nil {
				t.Fatal("expected error for invalid input")
			}
		})
	}
}

func TestEditFile_invalidJSONReturnsError(t *testing.T) {
	_, err := EditFile(json.RawMessage(`not json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON input")
	}
}

func TestEditFile_reportsFileReadErrorOnNonexistentPath(t *testing.T) {
	input, _ := json.Marshal(EditFileInput{Path: "/nonexistent/xyz", OldStr: "a", NewStr: "b"})
	_, err := EditFile(input)
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
}
