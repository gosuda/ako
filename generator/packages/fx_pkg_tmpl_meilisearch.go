package packages

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gosuda/ako/util/module"
	"github.com/gosuda/ako/util/template"
)

func init() {
	pkgTemplateList["[SearchEngine/Meilisearch] Meilisearch"] = createMeiliSearchClientFile
}

const (
	meilisearchDependency     = `github.com/meilisearch/meilisearch-go@v0.31.0`
	meilisearchClientTemplate = `package {{.package_name}}

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/meilisearch/meilisearch-go"
	"go.uber.org/fx"
)

var Module = fx.Module("{{.package_name}}",
	fx.Provide(ConfigRegister()),
	fx.Provide(fx.Annotate(New, fx.As(/* implemented interfaces */))),
)

func ConfigRegister() func() *Config {
	return func() *Config {
		return &Config{
			Host:   os.Getenv("MEILISEARCH_{{.client_name}}_HOST"),
			APIKey: os.Getenv("MEILISEARCH_{{.client_name}}_API_KEY"),
		}
	}
}

type Param struct {
	fx.In
	Cfg *Config
}

type Config struct {
	Host string
	APIKey string
}

type {{.client_name}} struct {
	client meilisearch.ServiceManager
}

func New(ctx context.Context, lc fx.Lifecycle, param Param) *{{.client_name}} {
	cli := &{{.client_name}}{}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			cli.client = meilisearch.New(param.Cfg.Host, meilisearch.WithAPIKey(param.Cfg.APIKey))
			return nil
		},
		OnStop: func(ctx context.Context) error {
			cli.client.Close()
			return nil
		},
	})

	return cli
}

type SearchResult[T any] struct {
	Hits               []*T   ` + "`" + `json:"hits"` + "`" + `
	Offset             int64  ` + "`" + `json:"offset"` + "`" + `
	Limit              int64  ` + "`" + `json:"limit"` + "`" + `
	EstimatedTotalHits int64  ` + "`" + `json:"estimatedTotalHits"` + "`" + `
	ProcessingTimeMs   int64  ` + "`" + `json:"processingTimeMs"` + "`" + `
	Query              string ` + "`" + `json:"query"` + "`" + `
}

type SearchOption struct {
	Limit  int64
	Offset int64
}

func search[T any](c *{{.client_name}}, index string, query string, option SearchOption) (*SearchResult[T], error) {
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

func insert[T any](c *{{.client_name}}, index string, documents ...T) error {
	_, err := c.client.Index(index).AddDocuments(documents)
	if err != nil {
		return fmt.Errorf("meilisearch: addDocuments: %w", err)
	}
	return nil
}

func remove(c *{{.client_name}}, index string, documents ...string) error {
	_, err := c.client.Index(index).DeleteDocuments(documents)
	if err != nil {
		return fmt.Errorf("meilisearch: deleteDocuments: %w", err)
	}
	return nil
}

func updateSynonyms(c *{{.client_name}}, index string, synonyms map[string][]string) error {
	_, err := c.client.Index(index).UpdateSynonyms(&synonyms)
	if err != nil {
		return fmt.Errorf("meilisearch: updateSynonyms: %w", err)
	}
	return nil
}
`
)

func createMeiliSearchClientFile(path string, name string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	name = strings.ToUpper(name[:1]) + name[1:]

	packageName := filepath.Base(path)
	if err := template.WriteTemplate2File(filepath.Join(path, fmt.Sprintf(fxFileName, name)), meilisearchClientTemplate, map[string]any{
		"package_name": packageName,
		"client_name":  name,
	}); err != nil {
		return err
	}

	if err := module.GetGoModule(meilisearchDependency); err != nil {
		return err
	}

	return nil
}
