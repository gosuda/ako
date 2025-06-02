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
	pkgTemplateList["[SearchEngine/OpenSearch] Opensearch"] = createFxOpensearchClientFile
}

const (
	opensearchDependency       = `github.com/opensearch-project/opensearch-go/v4`
	opensearchDependencyAws    = `github.com/aws/aws-sdk-go-v2/aws`
	opensearchDependencyConfig = `github.com/aws/aws-sdk-go-v2/config`
	opensearchClientTemplate   = `package {{.package_name}}

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	opensearch "github.com/opensearch-project/opensearch-go/v4"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	requestsigner "github.com/opensearch-project/opensearch-go/v4/signer/awsv2"
	"go.uber.org/fx"
)

const Name = "{{.client_name}}"

var Module = fx.Module("{{.package_name}}",
	fx.Provide(ConfigRegister()),
	fx.Provide(fx.Annotate(New, fx.As(/* implemented interfaces */))),
)

func ConfigRegister() func() *Config {
	return func() *Config {
		return &Config{
			Address:  strings.Split(os.Getenv("OPENSEARCH_{{.client_name}}_ADDRESS"), ","),
			AuthPassword: &AuthPassword{
				Username: os.Getenv("OPENSEARCH_{{.client_name}}_USERNAME"),
				Password: os.Getenv("OPENSEARCH_{{.client_name}}_PASSWORD"),
			},
			AuthAWS: &AuthAWS{
				Region:          os.Getenv("OPENSEARCH_{{.client_name}}_REGION"),
				AccessKeyID:     os.Getenv("OPENSEARCH_{{.client_name}}_ACCESS_KEY_ID"),
				SecretAccessKey: os.Getenv("OPENSEARCH_{{.client_name}}_SECRET_ACCESS_KEY"),
				SessionToken:    os.Getenv("OPENSEARCH_{{.client_name}}_SESSION_TOKEN"),
			},
		}
	}
}

type Param struct {
	fx.In
	Cfg *Config
}

type AuthPassword struct {
	Username string
	Password string
}

type AuthAWS struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

type Config struct {
	Address       []string
	AuthPassword  *AuthPassword
	AuthAWS       *AuthAWS
}

type {{.client_name}} struct {
	client *opensearch.Client
}

func New(ctx context.Context, lc fx.Lifecycle, param Param) *{{.client_name}} {
	cli := &{{.client_name}}{}

	authCond := [2]bool{}
	if param.Cfg.AuthPassword != nil && param.Cfg.AuthPassword.Username != "" && param.Cfg.AuthPassword.Password != "" {
		authCond[0] = true
	}
	if param.Cfg.AuthAWS != nil && param.Cfg.AuthAWS.Region != "" && param.Cfg.AuthAWS.AccessKeyID != "" && param.Cfg.AuthAWS.SecretAccessKey != "" {
		authCond[1] = true
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			cfg := opensearch.Config{
				Addresses: param.Cfg.Address,
				Transport: &http.Transport{
            		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
        		},
				MaxRetries: 5,
        		RetryOnStatus: []int{502, 503, 504},
			}

			switch authCond {
			case [2]bool{true, false}:
				cfg.Username = param.Cfg.AuthPassword.Username
				cfg.Password = param.Cfg.AuthPassword.Password
			case [2]bool{false, true}, [2]bool{true, true}:
				awsCfg, err := config.LoadDefaultConfig(ctx,
					config.WithRegion(param.Cfg.AuthAWS.Region),
					config.WithCredentialsProvider(
						getCredentialProvider(param.Cfg.AuthAWS.AccessKeyID, param.Cfg.AuthAWS.SecretAccessKey, param.Cfg.AuthAWS.SessionToken),
					),
				)
				if err != nil {
					return fmt.Errorf("config.LoadDefaultConfig: %w", err)
				}

				signer, err := requestsigner.NewSignerWithService(awsCfg, "es")
				if err != nil {
					return fmt.Errorf("requests.NewSignerWithService: %w", err)
				}

				cfg.Signer = signer
			case [2]bool{false, false}:
				// No authentication
			}

			client, err := opensearch.NewClient(cfg)
			if err != nil {
				return fmt.Errorf("opensearch.NewClient: %w", err)
			}

			cli.client = client

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})

	return cli
}

func getCredentialProvider(accessKey, secretAccessKey, token string) aws.CredentialsProviderFunc {
	return func(ctx context.Context) (aws.Credentials, error) {
		c := &aws.Credentials{
			AccessKeyID:     accessKey,
			SecretAccessKey: secretAccessKey,
			SessionToken:    token,
		}
		return *c, nil
	}
}


func (c *{{.client_name}}) createIndex(ctx context.Context, name string, shards, replicas int) error {
	requestValue := map[string]any{
		"settings": map[string]any{
			"index": map[string]any{
				"number_of_shards":   shards,
				"number_of_replicas": replicas,
			},
		},
	}

	requestBody, err := json.Marshal(requestValue)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	req := opensearchapi.IndicesCreateReq{
		Index: name,
		Body:  bytes.NewReader(requestBody),
	}

	res, err := c.client.Do(ctx, req, nil)
	if err != nil {
		return fmt.Errorf("client.Do: %w", err)
	}

	defer res.Body.Close()

	if res.IsError() {
		var errRes map[string]any
		if err := json.NewDecoder(res.Body).Decode(&errRes); err != nil {
			return fmt.Errorf("json.NewDecoder: %w", err)
		}

		return fmt.Errorf("error: %s", errRes["error"])
	}

	return nil
}

func (c *{{.client_name}}) deleteIndex(ctx context.Context, name ...string) error {
	req := opensearchapi.IndicesDeleteReq{
		Indices: name,
	}

	res, err := c.client.Do(ctx, req, nil)
	if err != nil {
		return fmt.Errorf("client.Do: %w", err)
	}

	defer res.Body.Close()

	if res.IsError() {
		var errRes map[string]any
		if err := json.NewDecoder(res.Body).Decode(&errRes); err != nil {
			return fmt.Errorf("json.NewDecoder: %w", err)
		}

		return fmt.Errorf("error: %s", errRes["error"])
	}

	return nil
}

func (c *{{.client_name}}) createDocument(ctx context.Context, index string, id string, body any) error {
	requestBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	req := opensearchapi.IndexReq{
		Index:      index,
		DocumentID: id,
		Body:       bytes.NewReader(requestBody),
	}

	res, err := c.client.Do(ctx, req, nil)
	if err != nil {
		return fmt.Errorf("client.Do: %w", err)
	}

	defer res.Body.Close()

	if res.IsError() {
		var errRes map[string]any
		if err := json.NewDecoder(res.Body).Decode(&errRes); err != nil {
			return fmt.Errorf("json.NewDecoder: %w", err)
		}

		return fmt.Errorf("error: %s", errRes["error"])
	}

	return nil
}

func (c *{{.client_name}}) deleteDocument(ctx context.Context, index string, id string) error {
	req := opensearchapi.DocumentDeleteReq{
		Index:      index,
		DocumentID: id,
	}

	res, err := c.client.Do(ctx, req, nil)
	if err != nil {
		return fmt.Errorf("client.Do: %w", err)
	}

	defer res.Body.Close()

	if res.IsError() {
		var errRes map[string]any
		if err := json.NewDecoder(res.Body).Decode(&errRes); err != nil {
			return fmt.Errorf("json.NewDecoder: %w", err)
		}

		return fmt.Errorf("error: %s", errRes["error"])
	}

	return nil
}

func (c *{{.client_name}}) deleteDocumentsByQuery(ctx context.Context, index string, query string) error {
	requestValue := map[string]any{
		"query": query,
	}

	requestBody, err := json.Marshal(requestValue)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	req := opensearchapi.DocumentDeleteByQueryReq{
		Indices: []string{index},
		Body:    bytes.NewReader(requestBody),
	}

	res, err := c.client.Do(ctx, req, nil)
	if err != nil {
		return fmt.Errorf("client.Do: %w", err)
	}

	defer res.Body.Close()

	if res.IsError() {
		var errRes map[string]any
		if err := json.NewDecoder(res.Body).Decode(&errRes); err != nil {
			return fmt.Errorf("json.NewDecoder: %w", err)
		}

		return fmt.Errorf("error: %s", errRes["error"])
	}

	return nil
}

func (c *{{.client_name}}) searchDocuments(ctx context.Context, indices []string, query string, from, to time.Time, count int) (*opensearchapi.SearchResp, error) {
	requestValue := map[string]any{
		"query": map[string]any{
			"query_string": query,
			"range": map[string]any{
				"timestamp": map[string]any{
					"gte": from.Format(time.RFC3339),
					"lt":  to.Format(time.RFC3339),
				},
			},
		},
		"size": count,
	}

	requestBody, err := json.Marshal(requestValue)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal: %w", err)
	}

	req := opensearchapi.SearchReq{
		Indices: indices,
		Body:    bytes.NewReader(requestBody),
	}

	res, err := c.client.Do(ctx, req, nil)
	if err != nil {
		return nil, fmt.Errorf("client.Do: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var errRes map[string]any
		if err := json.NewDecoder(res.Body).Decode(&errRes); err != nil {
			return nil, fmt.Errorf("json.NewDecoder: %w", err)
		}

		return nil, fmt.Errorf("error: %s", errRes["error"])
	}

	response := opensearchapi.SearchResp{}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("json.NewDecoder: %w", err)
	}

	return &response, nil
}
`
)

func createFxOpensearchClientFile(path string, name string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	name = strings.ToUpper(name[:1]) + name[1:]

	packageName := filepath.Base(path)
	if err := template.WriteTemplate2File(filepath.Join(path, fmt.Sprintf(fxFileName, name)), opensearchClientTemplate, map[string]any{
		"package_name": packageName,
		"client_name":  name,
	}); err != nil {
		return err
	}

	if err := module.GetGoModule(opensearchDependency); err != nil {
		return err
	}

	if err := module.GetGoModule(opensearchDependencyAws); err != nil {
		return err
	}

	if err := module.GetGoModule(opensearchDependencyConfig); err != nil {
		return err
	}

	return nil
}
