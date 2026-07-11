package agent

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// rgAvailable reports whether ripgrep is installed and usable.
func rgAvailable() bool {
	_, err := exec.LookPath("rg")
	return err == nil
}

func TestCodeSearch_invalidJSONReturnsError(t *testing.T) {
	_, err := CodeSearch(json.RawMessage(`not json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON input")
	}
}

func TestCodeSearch_emptyPatternReturnsError(t *testing.T) {
	input, _ := json.Marshal(CodeSearchInput{Pattern: ""})
	_, err := CodeSearch(input)
	if err == nil {
		t.Fatal("expected error for empty pattern")
	}
}

func TestCodeSearch_returnsMatches(t *testing.T) {
	if !rgAvailable() {
		t.Skip("rg not found in PATH")
	}

	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.go"), []byte("package main\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n"), 0o644)

	input, _ := json.Marshal(CodeSearchInput{Pattern: "hello", Path: dir})
	result, err := CodeSearch(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "hello") {
		t.Errorf("expected match for 'hello', got %q", result)
	}
}

func TestCodeSearch_noMatchesReturnsMessage(t *testing.T) {
	if !rgAvailable() {
		t.Skip("rg not found in PATH")
	}

	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("nothing here"), 0o644)

	input, _ := json.Marshal(CodeSearchInput{Pattern: "xyz_not_present", Path: dir})
	result, err := CodeSearch(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "No matches found" {
		t.Errorf("expected 'No matches found', got %q", result)
	}
}

func TestCodeSearch_caseInsensitiveByDefault(t *testing.T) {
	if !rgAvailable() {
		t.Skip("rg not found in PATH")
	}

	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("HELLO\nhello"), 0o644)

	input, _ := json.Marshal(CodeSearchInput{Pattern: "hello", Path: dir, CaseSensitive: false})
	result, err := CodeSearch(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Count(result, "\n") < 1 {
		t.Errorf("expected at least 2 matches (case insensitive), got %q", result)
	}
}

func TestCodeSearch_caseSensitive(t *testing.T) {
	if !rgAvailable() {
		t.Skip("rg not found in PATH")
	}

	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("HELLO\nhello"), 0o644)

	input, _ := json.Marshal(CodeSearchInput{Pattern: "HELLO", Path: dir, CaseSensitive: true})
	result, err := CodeSearch(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Count(result, "\n") > 0 {
		t.Errorf("expected exactly 1 match (case sensitive), got %q", result)
	}
}

func TestCodeSearch_fileTypeFilter(t *testing.T) {
	if !rgAvailable() {
		t.Skip("rg not found in PATH")
	}

	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.go"), []byte("hello go"), 0o644)
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello txt"), 0o644)

	input, _ := json.Marshal(CodeSearchInput{Pattern: "hello", Path: dir, FileType: "go"})
	result, err := CodeSearch(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(result, "a.txt") {
		t.Errorf("expected only .go files, got %q", result)
	}
	if !strings.Contains(result, "a.go") {
		t.Errorf("expected a.go to be in results, got %q", result)
	}
}

func TestCodeSearch_truncatesOutput(t *testing.T) {
	if !rgAvailable() {
		t.Skip("rg not found in PATH")
	}

	dir := t.TempDir()
	var b strings.Builder
	for range 60 {
		b.WriteString("needle\n")
	}
	os.WriteFile(filepath.Join(dir, "f.txt"), []byte(b.String()), 0o644)

	input, _ := json.Marshal(CodeSearchInput{Pattern: "needle", Path: dir})
	result, err := CodeSearch(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "showing first 50") {
		t.Errorf("expected truncation message, got %q", result)
	}
}

func TestCodeSearch_rgNotFoundReturnsError(t *testing.T) {
	if rgAvailable() {
		t.Skip("rg is available; can't test missing rg")
	}

	input, _ := json.Marshal(CodeSearchInput{Pattern: "x"})
	_, err := CodeSearch(input)
	if err == nil {
		t.Fatal("expected error when rg is not available")
	}
}
