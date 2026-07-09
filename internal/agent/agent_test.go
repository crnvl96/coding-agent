package agent

import (
	"bytes"
	"context"
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

	a := NewAgent(client, getUserMessage)

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

	a := NewAgent(client, getUserMessage)
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

	a := NewAgent(client, getUserMessage)
	err := a.Run(context.Background())
	if err == nil {
		t.Fatal("Run() expected an error from inference, got nil")
	}
}
