package ai

import (
	"context"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type OpenAIClient struct {
	client *openai.Client
	model  string
}

func NewOpenAIClient(apiKey string, model string) (*OpenAIClient, error) {
	client := openai.NewClient(option.WithAPIKey(apiKey))

	return &OpenAIClient{
		client: &client,
		model:  model,
	}, nil
}

func (c *OpenAIClient) GenerateCommitMessage(ctx context.Context, gitDiff string) (<-chan string, error) {
	ch := make(chan string, 1)
	chat, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(CommitMessageGenerationPrompt),
			openai.UserMessage(gitDiff),
		},
		Model: c.model,
	})
	if err != nil {
		close(ch)
		return nil, err
	}

	ch <- chat.Choices[0].Message.Content
	close(ch)

	return ch, nil
}
