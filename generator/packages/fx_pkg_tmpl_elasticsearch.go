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
	pkgTemplateList["[SearchEngine/Elasticsearch] Elasticsearch"] = createFxElasticsearchClientFile
}

const (
	elasticsearchDependency     = `github.com/elastic/go-elasticsearch/v9@latest`
	elasticsearchClientTemplate = `package {{.package_name}}

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/go-elasticsearch/v9/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"go.uber.org/fx"
)

var Module = fx.Module("{{.package_name}}",
	fx.Provide(ConfigRegister()),
	fx.Provide(fx.Annotate(New, fx.As(/* implemented interfaces */))),
)

func ConfigRegister() func() *Config {
	return func() *Config {
		return &Config{
			Address:     strings.Split(os.Getenv("ELASTICSEARCH_{{.client_name}}_ADDRESS"), ","),
			Username:    os.Getenv("ELASTICSEARCH_{{.client_name}}_USERNAME"),
			Password:    os.Getenv("ELASTICSEARCH_{{.client_name}}_PASSWORD"),
		}
	}
}

type Param struct {
	fx.In
	Cfg *Config
}

type Config struct {
	Address     []string
	Username 	string
	Password 	string
}

type {{.client_name}} struct {
	client *elasticsearch.TypedClient
}

func New(ctx context.Context, lc fx.Lifecycle, param Param) *{{.client_name}} {
	cli := &{{.client_name}}{}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			tc, err := elasticsearch.NewTypedClient(elasticsearch.Config{
				Addresses: param.Cfg.Address,
				Username: param.Cfg.Username,
				Password: param.Cfg.Password,
			})
			if err != nil {
				return fmt.Errorf("elasticsearch.NewTypedClient: %w", err)
			}

			cli.client = tc

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})

	return cli
}

func (e *{{.client_name}}) createIndex(ctx context.Context, index string) error {
	_, err := e.client.Indices.Create(index).Do(ctx)
	if err != nil {
		return fmt.Errorf("create index %s: %w", index, err)
	}

	return nil
}

func (e *{{.client_name}}) deleteIndex(ctx context.Context, index string) error {
	_, err := e.client.Indices.Delete(index).Do(ctx)
	if err != nil {
		return fmt.Errorf("delete index %s: %w", index, err)
	}

	return nil
}

func (e *{{.client_name}}) indexExists(ctx context.Context, index string) (bool, error) {
	res, err := e.client.Indices.Exists(index).Do(ctx)
	if err != nil {
		return false, fmt.Errorf("index exists %s: %w", index, err)
	}

	return res, nil
}

func (e *{{.client_name}}) searchDocuments(ctx context.Context, index string, query string, from, to time.Time, count int) ([]types.Hit, error) {
	res, err := e.client.Search().Request(&search.Request{
		Query: &types.Query{
			QueryString: &types.QueryStringQuery{
				Query: query,
			},
			Range: map[string]types.RangeQuery{
				"timestamp": map[string]any{
					"gte": from.Format(time.RFC3339),
					"lt":  to.Format(time.RFC3339),
				},
			},
		},
	}).Size(count).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("search index %s: %w", index, err)
	}

	return res.Hits.Hits, nil
}

func (e *{{.client_name}}) deleteDocumentsByQuery(ctx context.Context, index string, query string) error {
	_, err := e.client.DeleteByQuery(index).Q(query).Do(ctx)
	if err != nil {
		return fmt.Errorf("delete documents by query %s: %w", index, err)
	}

	return nil
}

func (e *{{.client_name}}) deleteDocument(ctx context.Context, index string, id string) error {
	_, err := e.client.Delete(index, id).Do(ctx)
	if err != nil {
		return fmt.Errorf("delete document %s: %w", id, err)
	}

	return nil
}

func (e *{{.client_name}}) createDocument(ctx context.Context, index string, id string, document any) error {
	_, err := e.client.Index(index).Id(id).Document(document).Do(ctx)
	if err != nil {
		return fmt.Errorf("create document %s: %w", id, err)
	}

	return nil
}

func (e *{{.client_name}}) updateDocument(ctx context.Context, index string, id string, document any) error {
	_, err := e.client.Update(index, id).Doc(document).Do(ctx)
	if err != nil {
		return fmt.Errorf("update document %s: %w", id, err)
	}

	return nil
}
`
)

func createFxElasticsearchClientFile(path string, name string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	name = strings.ToUpper(name[:1]) + name[1:]

	packageName := filepath.Base(path)
	if err := template.WriteTemplate2File(filepath.Join(path, fmt.Sprintf(fxFileName, name)), elasticsearchClientTemplate, map[string]any{
		"package_name": packageName,
		"client_name":  name,
	}); err != nil {
		return err
	}

	if err := module.GetGoModule(elasticsearchDependency); err != nil {
		return err
	}

	return nil
}
