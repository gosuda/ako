package meilisearch

import (
	"encoding/json"
	"fmt"

	"github.com/meilisearch/meilisearch-go"
)

type Config struct {
	Host   *string
	ApiKey *string
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) WithHost(host string) *Config {
	c.Host = &host
	return c
}

func (c *Config) WithApiKey(apiKey string) *Config {
	c.ApiKey = &apiKey
	return c
}

type Client struct {
	client meilisearch.ServiceManager
}

func New(config *Config) *Client {
	host := "http://localhost:7700"
	if config.Host != nil {
		host = *config.Host
	}
	apiKey := ""
	if config.ApiKey != nil {
		apiKey = *config.ApiKey
	}

	client := meilisearch.New(host, meilisearch.WithAPIKey(apiKey))

	return &Client{client: client}
}

type SearchResult[T any] struct {
	Hits               []*T   `json:"hits"`
	Offset             int64  `json:"offset"`
	Limit              int64  `json:"limit"`
	EstimatedTotalHits int64  `json:"estimatedTotalHits"`
	ProcessingTimeMs   int64  `json:"processingTimeMs"`
	Query              string `json:"query"`
}

type SearchOption struct {
	Limit  int64
	Offset int64
}

func Search[T any](c *Client, index string, query string, option SearchOption) (*SearchResult[T], error) {
	res, err := c.client.Index(index).SearchRaw(query, &meilisearch.SearchRequest{
		Limit:            option.Limit,
		Offset:           option.Offset,
		ShowRankingScore: true,
	})
	if err != nil {
		return nil, fmt.Errorf("meilisearch: searchRaw: %w", err)
	}

	sr := new(SearchResult[T])
	if err := json.Unmarshal(*res, sr); err != nil {
		return nil, fmt.Errorf("meilisearch: unmarshal: %w", err)
	}

	return sr, nil
}

func Insert[T any](c *Client, index string, documents ...T) error {
	_, err := c.client.Index(index).AddDocuments(documents)
	if err != nil {
		return fmt.Errorf("meilisearch: addDocuments: %w", err)
	}
	return nil
}

func UpdateSynonyms(c *Client, index string, synonyms map[string][]string) error {
	_, err := c.client.Index(index).UpdateSynonyms(&synonyms)
	if err != nil {
		return fmt.Errorf("meilisearch: updateSynonyms: %w", err)
	}
	return nil
}
