package ai

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type AnthropicClient struct {
	client *anthropic.Client
	model  string
}

func NewAnthropicClient(apiKey string, model string) (*AnthropicClient, error) {
	client := anthropic.NewClient(option.WithAPIKey(apiKey))

	return &AnthropicClient{
		client: &client,
		model:  model,
	}, nil
}

func (c *AnthropicClient) GenerateCommitMessage(ctx context.Context, gitDiff string) (<-chan string, error) {
	ch := make(chan string, 1)
	chat, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			{
				Role: anthropic.MessageParamRoleAssistant,
				Content: []anthropic.ContentBlockParamUnion{{
					OfRequestTextBlock: &anthropic.TextBlockParam{Text: CommitMessageGenerationPrompt},
				}},
			},
			{
				Role: anthropic.MessageParamRoleUser,
				Content: []anthropic.ContentBlockParamUnion{{
					OfRequestTextBlock: &anthropic.TextBlockParam{Text: gitDiff},
				}},
			},
		},
		Model: c.model,
	})
	if err != nil {
		close(ch)
		return nil, err
	}

	go func() {
		for _, content := range chat.Content {
			ch <- content.Text
		}
	}()

	return ch, nil
}
