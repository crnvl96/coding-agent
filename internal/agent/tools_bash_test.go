package agent

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBash_returnsStdoutForSuccessfulCommand(t *testing.T) {
	input, _ := json.Marshal(BashInput{Command: "echo hello"})
	result, err := Bash(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

func TestBash_returnsErrorDetailsForFailedCommand(t *testing.T) {
	input, _ := json.Marshal(BashInput{Command: "false"})
	result, err := Bash(input)
	// Failed commands should not return Go errors — the failure is in the output.
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "Command failed with error") {
		t.Errorf("expected failure message in output, got %q", result)
	}
}

func TestBash_includesStderrInOutput(t *testing.T) {
	input, _ := json.Marshal(BashInput{Command: "echo stdout; echo stderr >&2"})
	result, err := Bash(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "stdout") {
		t.Errorf("expected stdout in output, got %q", result)
	}
	if !strings.Contains(result, "stderr") {
		t.Errorf("expected stderr in output, got %q", result)
	}
}

func TestBash_emptyCommand(t *testing.T) {
	input, _ := json.Marshal(BashInput{Command: ""})
	result, err := Bash(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Bash -c "" returns empty output with no error.
	if result != "" {
		t.Errorf("expected empty output, got %q", result)
	}
}

func TestBash_invalidJSONReturnsError(t *testing.T) {
	_, err := Bash(json.RawMessage(`not json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON input")
	}
}

func TestBash_trimsTrailingWhitespace(t *testing.T) {
	// printf writes exactly what we ask without appending or modifying.
	input, _ := json.Marshal(BashInput{Command: "printf 'hello  '"})
	result, err := Bash(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello" {
		t.Errorf("expected trailing spaces trimmed; got %q", result)
	}
}
