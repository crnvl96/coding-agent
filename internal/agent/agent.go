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
	ansiGreen  = "\u001b[92m"
	ansiCyan   = "\u001b[96m"
	ansiRed    = "\u001b[91m"
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
	tools          []ToolDefinition
}

// NewAgent creates an Agent with the given message creator, user message input function,
// and tool definitions available to the model.
func NewAgent(
	chat chatCreator,
	getUserMessage func() (string, bool),
	tools []ToolDefinition,
) *agent {
	return &agent{
		chat:           chat,
		getUserMessage: getUserMessage,
		tools:          tools,
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

		// Inner loop: process tool calls until the model responds with text only.
		for {
			var toolResults []anthropic.ContentBlockParamUnion
			var hasToolUse bool

			for _, content := range message.Content {
				switch content.Type {
				case "text":
					fmt.Printf(ansiYellow+"AI"+ansiReset+": %s\n", content.Text)
				case "tool_use":
					hasToolUse = true
					toolUse := content.AsToolUse()
					fmt.Printf(ansiCyan+"tool"+ansiReset+": %s(%s)\n", toolUse.Name, string(toolUse.Input))

					var toolResult string
					var toolError error
					var toolFound bool
					for _, tool := range a.tools {
						if tool.Name == toolUse.Name {
							toolFound = true
							toolResult, toolError = tool.Function(toolUse.Input)
							break
						}
					}
					if !toolFound {
						toolError = fmt.Errorf("tool %q not found", toolUse.Name)
					}

					if toolError != nil {
						fmt.Printf(ansiRed+"error"+ansiReset+": %s\n", toolError.Error())
						toolResults = append(toolResults,
							anthropic.NewToolResultBlock(toolUse.ID, toolError.Error(), true))
					} else {
						fmt.Printf(ansiGreen+"result"+ansiReset+": %s\n", toolResult)
						toolResults = append(toolResults,
							anthropic.NewToolResultBlock(toolUse.ID, toolResult, false))
					}
				}
			}

			if !hasToolUse {
				break
			}

			toolResultMessage := anthropic.NewUserMessage(toolResults...)
			conversation = append(conversation, toolResultMessage)

			message, err = a.runInference(ctx, conversation)
			if err != nil {
				return err
			}
			conversation = append(conversation, message.ToParam())
		}
	}

	return nil
}

func (a *agent) runInference(ctx context.Context, conversation []anthropic.MessageParam) (*anthropic.Message, error) {
	anthropicTools := make([]anthropic.ToolUnionParam, len(a.tools))
	for i, tool := range a.tools {
		anthropicTools[i] = anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        tool.Name,
				Description: anthropic.String(tool.Description),
				InputSchema: tool.InputSchema,
			},
		}
	}

	message, err := a.chat.New(ctx, anthropic.MessageNewParams{
		Model:     defaultModel,
		MaxTokens: int64(defaultInferenceMaxTokens),
		Messages:  conversation,
		Tools:     anthropicTools,
	})

	return message, err
}
