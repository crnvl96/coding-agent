package agent

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const (
	defaultModel              = "deepseek-v4-pro"
	defaultInferenceMaxTokens = 4096

	// ANSI escape sequences for terminal colours in the chat prompt.
	// ansiReset reverts to the default terminal colour after a coloured label.
	ansiBlue   = "\u001b[94m"
	ansiYellow = "\u001b[93m"
	ansiReset  = "\u001b[0m"
)

type chatCreator interface {
	New(ctx context.Context, params anthropic.MessageNewParams,
		opts ...option.RequestOption) (*anthropic.Message, error)
}

type agent struct {
	chat           chatCreator
	getUserMessage func() (string, bool)
}

// NewAgent creates an Agent with the given message creator and user message input function.
func NewAgent(
	chat chatCreator,
	getUserMessage func() (string, bool),
) *agent {
	return &agent{
		chat:           chat,
		getUserMessage: getUserMessage,
	}
}

// Run starts the interactive chat loop, reading user messages and printing AI responses.
// It returns when the user signals EOF (ctrl-d) or when an inference error occurs.
func (a *agent) Run(ctx context.Context) error {
	conversation := []anthropic.MessageParam{}

	fmt.Println("Chat with AI (use 'ctrl-c' to quit)")

	for {
		fmt.Print(ansiBlue + "You" + ansiReset + ": ")
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
				fmt.Printf(ansiYellow+"AI"+ansiReset+": %s\n", content.Text)
			}
		}
	}

	return nil
}

func (a *agent) runInference(ctx context.Context, conversation []anthropic.MessageParam) (*anthropic.Message, error) {
	message, err := a.chat.New(ctx, anthropic.MessageNewParams{
		Model:     defaultModel,
		MaxTokens: int64(defaultInferenceMaxTokens),
		Messages:  conversation,
	})

	return message, err
}
