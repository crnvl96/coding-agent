package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type mocketChat struct {
	responseText string
	err          error
}

func (f *mocketChat) New(
	_ context.Context,
	_ anthropic.MessageNewParams,
	_ ...option.RequestOption,
) (*anthropic.Message, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &anthropic.Message{
		Content: []anthropic.ContentBlockUnion{
			{Type: "text", Text: f.responseText},
		},
	}, nil
}

func TestAgent_Run_respondsToUserMessage(t *testing.T) {
	inputs := []string{"hello"}
	getUserMessage := func() (string, bool) {
		if len(inputs) == 0 {
			return "", false
		}
		msg := inputs[0]
		inputs = inputs[1:]
		return msg, true
	}

	wantText := "Hi there!"
	client := &mocketChat{responseText: wantText}

	a := NewAgent(client, getUserMessage, nil)

	var stdout bytes.Buffer
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	errCh := make(chan error, 1)
	go func() {
		errCh <- a.Run(context.Background())
		w.Close()
	}()

	io.Copy(&stdout, r)
	os.Stdout = rescueStdout

	err := <-errCh
	if err != nil {
		t.Fatalf("Run() returned unexpected error: %v", err)
	}

	if !strings.Contains(stdout.String(), wantText) {
		t.Errorf("Run() output should contain AI response %q; got:\n%s", wantText, stdout.String())
	}
}

func TestAgent_Run_handlesEOF(t *testing.T) {
	getUserMessage := func() (string, bool) {
		return "", false
	}

	client := &mocketChat{}

	a := NewAgent(client, getUserMessage, nil)
	err := a.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() with immediate EOF returned error: %v", err)
	}
}

func TestAgent_Run_returnsInferenceError(t *testing.T) {
	inputs := []string{"hello"}
	getUserMessage := func() (string, bool) {
		if len(inputs) == 0 {
			return "", false
		}
		msg := inputs[0]
		inputs = inputs[1:]
		return msg, true
	}

	client := &mocketChat{
		err: errors.New("inference error"),
	}

	a := NewAgent(client, getUserMessage, nil)
	err := a.Run(context.Background())
	if err == nil {
		t.Fatal("Run() expected an error from inference, got nil")
	}
}

// --- Tool tests ---

// mockChat is a chatCreator double that returns pre-configured responses
// and records all inference requests for later inspection.
type mockChat struct {
	responses []*anthropic.Message
	params    []anthropic.MessageNewParams
	err       error
}

func (m *mockChat) New(
	_ context.Context,
	params anthropic.MessageNewParams,
	_ ...option.RequestOption,
) (*anthropic.Message, error) {
	m.params = append(m.params, params)
	if m.err != nil {
		return nil, m.err
	}
	if len(m.responses) == 0 {
		return &anthropic.Message{}, nil
	}
	resp := m.responses[0]
	m.responses = m.responses[1:]
	return resp, nil
}

// assistantMessage constructs a Message with the given JSON array of content blocks.
func assistantMessage(contentBlocksJSON string) *anthropic.Message {
	raw := `{"id":"msg_1","type":"message","role":"assistant","model":"test","content":` + contentBlocksJSON + `}`
	var msg anthropic.Message
	json.Unmarshal([]byte(raw), &msg)
	return &msg
}

// suppressStdout redirects os.Stdout to a pipe that is drained in the
// background. Returns a function that restores stdout.
func suppressStdout(t *testing.T) func() {
	t.Helper()
	rescue := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	done := make(chan struct{})
	go func() {
		io.Copy(io.Discard, r)
		close(done)
	}()
	return func() {
		w.Close()
		<-done
		os.Stdout = rescue
	}
}

func TestAgent_Run_executesToolAndSendsResult(t *testing.T) {
	defer suppressStdout(t)()

	var capturedInput json.RawMessage
	tool := ToolDefinition{
		Name:        "greet",
		Description: "Greets someone",
		InputSchema: GenerateSchema[struct {
			Name string `json:"name"`
		}](),
		Function: func(input json.RawMessage) (string, error) {
			capturedInput = input
			return "Hello, Alice!", nil
		},
	}

	// First response: tool_use, second: text (ends the inner loop).
	mock := &mockChat{
		responses: []*anthropic.Message{
			assistantMessage(`[{"type":"tool_use","id":"toolu_001","name":"greet","input":{"name":"Alice"}}]`),
			assistantMessage(`[{"type":"text","text":"Done!"}]`),
		},
	}

	inputs := []string{"greet Alice"}
	getUserMessage := func() (string, bool) {
		if len(inputs) == 0 {
			return "", false
		}
		msg := inputs[0]
		inputs = inputs[1:]
		return msg, true
	}

	agt := NewAgent(mock, getUserMessage, []ToolDefinition{tool})
	err := agt.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}

	// Tool should have been called with the correct input.
	if capturedInput == nil {
		t.Fatal("tool was never called")
	}
	if string(capturedInput) != `{"name":"Alice"}` {
		t.Errorf("tool input = %s, want {\"name\":\"Alice\"}", capturedInput)
	}

	// Two inference calls: one with tool_use response, one after tool results.
	if len(mock.params) != 2 {
		t.Fatalf("expected 2 inference calls, got %d", len(mock.params))
	}

	// The second call should include the tool result message.
	secondCall := mock.params[1]
	var foundToolResult bool
	for _, msg := range secondCall.Messages {
		for _, block := range msg.Content {
			if block.OfToolResult != nil {
				foundToolResult = true
				if block.OfToolResult.ToolUseID != "toolu_001" {
					t.Errorf("tool result ID = %q, want toolu_001", block.OfToolResult.ToolUseID)
				}
				if block.OfToolResult.IsError.Value {
					t.Error("tool result should not be an error")
				}
			}
		}
	}
	if !foundToolResult {
		t.Error("second inference call missing tool result message")
	}
}

func TestAgent_Run_reportsUnknownToolAsError(t *testing.T) {
	defer suppressStdout(t)()

	tool := ToolDefinition{
		Name:        "known",
		Description: "A known tool",
		InputSchema: GenerateSchema[struct{}](),
		Function:    func(json.RawMessage) (string, error) { return "ok", nil },
	}

	// Model requests a tool that isn't registered.
	mock := &mockChat{
		responses: []*anthropic.Message{
			assistantMessage(`[{"type":"tool_use","id":"t1","name":"nonexistent","input":{}}]`),
			assistantMessage(`[{"type":"text","text":"ok"}]`),
		},
	}

	inputs := []string{"go"}
	getUserMessage := func() (string, bool) {
		if len(inputs) == 0 {
			return "", false
		}
		msg := inputs[0]
		inputs = inputs[1:]
		return msg, true
	}

	agt := NewAgent(mock, getUserMessage, []ToolDefinition{tool})
	err := agt.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}

	// Second inference call should contain an error tool result.
	if len(mock.params) < 2 {
		t.Fatal("expected at least 2 inference calls")
	}
	secondCall := mock.params[1]
	var foundErrorResult bool
	for _, msg := range secondCall.Messages {
		for _, block := range msg.Content {
			if block.OfToolResult != nil && block.OfToolResult.IsError.Value {
				foundErrorResult = true
			}
		}
	}
	if !foundErrorResult {
		t.Error("expected an error tool result for unknown tool")
	}
}

func TestAgent_Run_reportsToolFunctionError(t *testing.T) {
	defer suppressStdout(t)()

	tool := ToolDefinition{
		Name:        "failing",
		Description: "Always fails",
		InputSchema: GenerateSchema[struct{}](),
		Function:    func(json.RawMessage) (string, error) { return "", errors.New("boom") },
	}

	mock := &mockChat{
		responses: []*anthropic.Message{
			assistantMessage(`[{"type":"tool_use","id":"t1","name":"failing","input":{}}]`),
			assistantMessage(`[{"type":"text","text":"handled"}]`),
		},
	}

	inputs := []string{"go"}
	getUserMessage := func() (string, bool) {
		if len(inputs) == 0 {
			return "", false
		}
		msg := inputs[0]
		inputs = inputs[1:]
		return msg, true
	}

	agt := NewAgent(mock, getUserMessage, []ToolDefinition{tool})
	err := agt.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}

	// The error from the tool function should be sent back as an error result.
	secondCall := mock.params[1]
	var foundErrorResult bool
	for _, msg := range secondCall.Messages {
		for _, block := range msg.Content {
			if block.OfToolResult != nil && block.OfToolResult.IsError.Value {
				foundErrorResult = true
			}
		}
	}
	if !foundErrorResult {
		t.Error("expected an error tool result from failing tool function")
	}
}

func TestAgent_Run_executesMultipleToolUses(t *testing.T) {
	defer suppressStdout(t)()

	var calls []string
	toolA := ToolDefinition{
		Name:        "a",
		Description: "Tool A",
		InputSchema: GenerateSchema[struct{}](),
		Function:    func(json.RawMessage) (string, error) { calls = append(calls, "a"); return "A", nil },
	}
	toolB := ToolDefinition{
		Name:        "b",
		Description: "Tool B",
		InputSchema: GenerateSchema[struct{}](),
		Function:    func(json.RawMessage) (string, error) { calls = append(calls, "b"); return "B", nil },
	}

	// One response with two tool_use blocks, then a text response.
	mock := &mockChat{
		responses: []*anthropic.Message{
			assistantMessage(`[
				{"type":"tool_use","id":"t1","name":"a","input":{}},
				{"type":"tool_use","id":"t2","name":"b","input":{}}
			]`),
			assistantMessage(`[{"type":"text","text":"done"}]`),
		},
	}

	inputs := []string{"go"}
	getUserMessage := func() (string, bool) {
		if len(inputs) == 0 {
			return "", false
		}
		msg := inputs[0]
		inputs = inputs[1:]
		return msg, true
	}

	agt := NewAgent(mock, getUserMessage, []ToolDefinition{toolA, toolB})
	err := agt.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}

	// Both tools should have been called.
	if len(calls) != 2 || calls[0] != "a" || calls[1] != "b" {
		t.Errorf("expected calls [a b], got %v", calls)
	}

	// Second inference should contain both tool results.
	secondCall := mock.params[1]
	var resultCount int
	for _, msg := range secondCall.Messages {
		for _, block := range msg.Content {
			if block.OfToolResult != nil {
				resultCount++
			}
		}
	}
	if resultCount != 2 {
		t.Errorf("expected 2 tool results in second call, got %d", resultCount)
	}
}

func TestAgent_Run_innerLoopContinuesUntilText(t *testing.T) {
	defer suppressStdout(t)()

	var calls int
	tool := ToolDefinition{
		Name:        "t",
		Description: "A tool",
		InputSchema: GenerateSchema[struct{}](),
		Function:    func(json.RawMessage) (string, error) { calls++; return "r", nil },
	}

	// Two tool_use rounds before a final text response.
	mock := &mockChat{
		responses: []*anthropic.Message{
			assistantMessage(`[{"type":"tool_use","id":"t1","name":"t","input":{}}]`),
			assistantMessage(`[{"type":"tool_use","id":"t2","name":"t","input":{}}]`),
			assistantMessage(`[{"type":"text","text":"finally"}]`),
		},
	}

	inputs := []string{"go"}
	getUserMessage := func() (string, bool) {
		if len(inputs) == 0 {
			return "", false
		}
		msg := inputs[0]
		inputs = inputs[1:]
		return msg, true
	}

	agt := NewAgent(mock, getUserMessage, []ToolDefinition{tool})
	err := agt.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}

	// Tool should have been called twice (once per tool_use round).
	if calls != 2 {
		t.Errorf("tool called %d times, want 2", calls)
	}

	// Three inference calls total: user msg → t1, t1 result → t2, t2 result → text.
	if len(mock.params) != 3 {
		t.Errorf("expected 3 inference calls, got %d", len(mock.params))
	}
}
