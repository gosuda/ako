package ollama

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ollama/ollama/api"
)

const (
	ModelGemma3p12B = "gemma3:12b"
	ModelGemma3p27B = "gemma3:27b"
)

const (
	DefaultHost = "http://localhost"
	DefaultPort = 11434

	EmbeddingLength = 3584
)

func Box[T any](value T) *T {
	return &value
}

type Client struct {
	client *api.Client
	model  string
}

type Config struct {
	Endpoint string
	Timeout  time.Duration
	Model    string
}

func New(cfg Config) (*Client, error) {
	if cfg.Endpoint == "" {
		cfg.Endpoint = DefaultHost + ":" + strconv.FormatInt(DefaultPort, 10)
	}

	u, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return nil, err
	}

	return &Client{
		client: api.NewClient(u, &http.Client{
			Timeout: cfg.Timeout,
		}),
		model: cfg.Model,
	}, nil
}

type ChatClient struct {
	client   *Client
	model    string
	messages []api.Message
}

func (c *Client) StartChat(ctx context.Context, basePrompt []string) *ChatClient {
	messages := make([]api.Message, len(basePrompt))
	for i, prompt := range basePrompt {
		messages[i] = api.Message{
			Content: prompt,
			Role:    "system",
		}
	}

	return &ChatClient{
		client:   c,
		model:    c.model,
		messages: messages,
	}
}

func (c *ChatClient) SendMessage(ctx context.Context, message string) (string, error) {
	wait := make(chan struct{}, 1)
	responseMessage := strings.Builder{}
	c.messages = append(c.messages, api.Message{
		Content: message,
		Role:    "user",
	})
	if err := c.client.client.Chat(ctx, &api.ChatRequest{
		Model:     c.model,
		Messages:  c.messages,
		Stream:    Box(true),
		KeepAlive: Box(api.Duration{Duration: 30 * time.Second}),
		Options:   nil,
	}, func(response api.ChatResponse) error {
		responseMessage.WriteString(response.Message.Content)

		if response.Done {
			close(wait)
		}

		return nil
	}); err != nil {
		return "", err
	}

	for range wait {
	}

	return responseMessage.String(), nil
}

func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	responseMessage := strings.Builder{}
	wait := make(chan struct{}, 1)
	if err := c.client.Generate(ctx, &api.GenerateRequest{
		Model:     c.model,
		Prompt:    prompt,
		Stream:    Box(true),
		KeepAlive: Box(api.Duration{Duration: 30 * time.Second}),
		Options:   nil,
	}, func(response api.GenerateResponse) error {
		responseMessage.WriteString(response.Response)

		if response.Done {
			close(wait)
		}

		return nil
	}); err != nil {
		return "", err
	}

	return responseMessage.String(), nil
}

func (c *Client) Embedding(ctx context.Context, prompt string) ([]float64, error) {
	response, err := c.client.Embeddings(ctx, &api.EmbeddingRequest{
		Model:  c.model,
		Prompt: prompt,
	})
	if err != nil {
		return nil, err
	}

	return response.Embedding, nil
}
