package agent

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
)

const (
	// defaultModel is the LLM model used when none is explicitly configured.
	defaultModel = "deepseek-v4-pro"
)

type Agent struct {
	client         *anthropic.Client
	getUserMessage func() (string, bool)
}

func NewAgent(
	client *anthropic.Client,
	getUserMessage func() (string, bool),
) *Agent {
	return &Agent{
		client:         client,
		getUserMessage: getUserMessage,
	}
}

func (a *Agent) Run(ctx context.Context) error {
	conversation := []anthropic.MessageParam{}

	fmt.Println("Chat with AI (use 'ctrl-c' to quit)")

	for {
		fmt.Print("\u001b[94mYou\u001b[0m: ")
		userInput, ok := a.getUserMessage()
		if !ok {
			break
		}

		if userInput == "" {
			continue
		}

		userMessage := anthropic.NewUserMessage(anthropic.NewTextBlock(userInput))
		conversation = append(conversation, userMessage)

		message, err := a.runInference(ctx, conversation)
		if err != nil {
			return err
		}
		conversation = append(conversation, message.ToParam())

		for _, content := range message.Content {
			switch content.Type {
			case "text":
				fmt.Printf("\u001b[93mAI\u001b[0m: %s\n", content.Text)
			}
		}
	}

	return nil
}

func (a *Agent) runInference(ctx context.Context, conversation []anthropic.MessageParam) (*anthropic.Message, error) {
	message, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     defaultModel,
		MaxTokens: int64(4096),
		Messages:  conversation,
	})

	return message, err
}
