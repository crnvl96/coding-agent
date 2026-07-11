package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestListFiles_returnsJSONArrayForDirectory(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), nil, 0o644)
	os.MkdirAll(filepath.Join(dir, "sub"), 0o755)
	os.WriteFile(filepath.Join(dir, "sub", "b.txt"), nil, 0o644)

	input, _ := json.Marshal(ListFilesInput{Path: dir})
	result, err := ListFiles(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var files []string
	json.Unmarshal([]byte(result), &files)

	if len(files) != 3 {
		t.Fatalf("expected 3 entries, got %d: %v", len(files), files)
	}
	if !contains(files, "a.txt") {
		t.Errorf("missing file a.txt")
	}
	if !contains(files, "sub/") {
		t.Errorf("missing directory sub/")
	}
	if !contains(files, filepath.Join("sub", "b.txt")) {
		t.Errorf("missing file sub/b.txt")
	}
}

func TestListFiles_defaultsToCWD(t *testing.T) {
	// Empty input (no path) should list the current working directory.
	result, err := ListFiles(json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var files []string
	if err := json.Unmarshal([]byte(result), &files); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	// Should at least find our own source files.
	if !contains(files, "tools_list_files.go") {
		t.Errorf("expected to find tools_list_files.go in cwd listing")
	}
}

func TestListFiles_returnsErrorForNonexistentPath(t *testing.T) {
	input, _ := json.Marshal(ListFilesInput{Path: "/nonexistent/path/xyz"})
	_, err := ListFiles(input)
	if err == nil {
		t.Fatal("expected error for nonexistent path, got nil")
	}
}

func TestListFiles_excludesRootDotEntry(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "only.txt"), nil, 0o644)

	input, _ := json.Marshal(ListFilesInput{Path: dir})
	result, err := ListFiles(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var files []string
	json.Unmarshal([]byte(result), &files)

	if contains(files, ".") {
		t.Error("result should not contain '.' entry")
	}
}

func TestListFiles_emptyDirectoryReturnsEmptyArray(t *testing.T) {
	dir := t.TempDir()

	input, _ := json.Marshal(ListFilesInput{Path: dir})
	result, err := ListFiles(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "[]" {
		t.Errorf("expected empty JSON array, got %s", result)
	}
}

func TestListFiles_invalidJSONReturnsError(t *testing.T) {
	_, err := ListFiles(json.RawMessage(`not json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON input")
	}
}

// contains reports whether s is present in ss.
func contains(ss []string, s string) bool {
	// sort.SearchStrings needs a sorted slice.
	sort.Strings(ss)
	i := sort.SearchStrings(ss, s)
	return i < len(ss) && ss[i] == s
}
