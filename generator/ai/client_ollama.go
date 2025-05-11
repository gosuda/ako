package ai

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/ollama/ollama/api"
)

type OllamaClient struct {
	client *api.Client
	model  string
}

func NewOllamaClient(host string, model string) (*OllamaClient, error) {
	urlValue, err := url.Parse(host)
	if err != nil {
		return nil, err
	}

	client := api.NewClient(urlValue, &http.Client{
		Timeout: 30 * time.Second,
	})

	return &OllamaClient{
		client: client,
		model:  model,
	}, nil
}

func (c *OllamaClient) chat(ctx context.Context, system []string, user []string) (<-chan string, error) {
	messages := make([]api.Message, 0, len(system)+len(user))
	for _, msg := range system {
		messages = append(messages, api.Message{
			Role:    "system",
			Content: msg,
		})
	}
	for _, msg := range user {
		messages = append(messages, api.Message{
			Role:    "user",
			Content: msg,
		})
	}

	ch := make(chan string, 1024)

	if err := c.client.Chat(ctx, &api.ChatRequest{
		Model:    c.model,
		Stream:   Wrap(true),
		Messages: messages,
	}, func(response api.ChatResponse) error {
		if response.Done {
			close(ch)
			return nil
		}

		ch <- response.Message.Content
		return nil
	}); err != nil {
		close(ch)
		return nil, err
	}

	return ch, nil
}

func (c *OllamaClient) GenerateCommitMessage(ctx context.Context, gitDiff string) (<-chan string, error) {
	ch, err := c.chat(ctx, []string{CommitMessageGenerationPrompt}, []string{gitDiff})
	if err != nil {
		return nil, err
	}

	return ch, nil
}
