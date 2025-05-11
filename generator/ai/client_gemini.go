package ai

import (
	"context"
	"log"

	"google.golang.org/genai"
)

type GeminiClient struct {
	client *genai.Client
	model  string
}

func NewGeminiClient(backend genai.Backend, apiKey string, model string, location string, project string) (*GeminiClient, error) {
	client, err := genai.NewClient(context.TODO(), &genai.ClientConfig{
		APIKey:   apiKey,
		Backend:  backend,
		Location: location,
		Project:  project,
	})
	if err != nil {
		return nil, err
	}

	return &GeminiClient{
		client: client,
		model:  model,
	}, nil
}

func (c *GeminiClient) GenerateCommitMessage(ctx context.Context, gitDiff string) (<-chan string, error) {
	ch := make(chan string, 1024)
	chat, err := c.client.Chats.Create(ctx, c.model, &genai.GenerateContentConfig{
		Temperature: Wrap(float32(0.75)),
	}, []*genai.Content{
		{
			Role: genai.RoleUser,
			Parts: []*genai.Part{
				genai.NewPartFromText(CommitMessageGenerationPrompt),
			},
		},
	})
	if err != nil {
		close(ch)
		return nil, err
	}

	go func() {
		defer close(ch)

		for part, err := range chat.SendMessageStream(ctx, *genai.NewPartFromText(gitDiff)) {
			if err != nil {
				log.Printf("Error occurred while receiving message: %v", err)
				return
			}

			if part == nil {
				continue
			}

			ch <- part.Text()
		}
	}()

	return ch, nil
}
