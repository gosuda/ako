package packages

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gosuda/ako/module"
	"github.com/gosuda/ako/template"
)

func init() {
	pkgTemplateList["[Storage] minio (s3 compatible)"] = createFxMinioFile
}

const (
	minioDependency     = `github.com/minio/minio-go/v7`
	minioClientTemplate = `package {{.package_name}}

import (
	"context"
	"io"
	"net/url"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/fx"
)

// Register is the fx.Provide function for the client.
// It registers the client as a dependency in the fx application.
// You can append interfaces into the fx.As() function to register multiple interfaces.
var Register = fx.Provide(fx.Annotate(New, fx.As()), ConfigRegister)

func ConfigRegister() func() *Config {
	return func() *Config {
		return &Config{
			AccessKey: os.Getenv("MINIO_{{.client_name}}_ACCESS_KEY"),
			SecretKey: os.Getenv("MINIO_{{.client_name}}_SECRET_KEY"),
			Endpoint:  os.Getenv("MINIO_{{.client_name}}_ENDPOINT"),
			Region:    os.Getenv("MINIO_{{.client_name}}_REGION"),
			Bucket:    os.Getenv("MINIO_{{.client_name}}_BUCKET"),
			UseSSL:    os.Getenv("MINIO_{{.client_name}}_USE_SSL") == "true",
		}
	}
}

type Param struct {
	fx.In
	Cfg *Config
}

type Config struct {
	AccessKey string
	SecretKey string
	Endpoint  string
	Region    string
	Bucket    string
	UseSSL    bool
}

type {{.client_name}} struct {
	config *Config
	conn   *minio.Client
}

func New(ctx context.Context, lc fx.Lifecycle, param Param) *{{.client_name}} {
	cli := {{.client_name}}{}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			client, err := minio.New(param.Cfg.Endpoint, &minio.Options{
				Creds:  credentials.NewStaticV4(param.Cfg.AccessKey, param.Cfg.SecretKey, ""),
				Secure: param.Cfg.UseSSL,
			})
			if err != nil {
				return err
			}

			cli.config = param.Cfg
			cli.conn = client

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})

	return &{{.client_name}}{}
}

type PutOptions minio.PutObjectOptions

func (c *{{.client_name}}) putObject(ctx context.Context, objectName string, reader io.Reader, option PutOptions) error {
	if _, err := c.conn.PutObject(ctx, c.config.Bucket, objectName, reader, -1, minio.PutObjectOptions(option)); err != nil {
		return err
	}

	return nil
}

func (c *{{.client_name}}) presignedPutObject(ctx context.Context, objectName string, expires time.Duration) (string, error) {
	u, err := c.conn.PresignedPutObject(ctx, c.config.Bucket, objectName, expires)
	if err != nil {
		return "", err
	}

	return u.String(), nil
}

type GetOptions minio.GetObjectOptions

func (c *{{.client_name}}) getObject(ctx context.Context, objectName string, option GetOptions) (io.ReadCloser, error) {
	reader, err := c.conn.GetObject(ctx, c.config.Bucket, objectName, minio.GetObjectOptions(option))
	if err != nil {
		return nil, err
	}

	context.AfterFunc(ctx, func() {
		reader.Close()
	})

	return reader, nil
}

func (c *{{.client_name}}) presignedGetObject(ctx context.Context, objectName string, expires time.Duration, reqParams url.Values) (string, error) {
	u, err := c.conn.PresignedGetObject(ctx, c.config.Bucket, objectName, expires, reqParams)
	if err != nil {
		return "", err
	}

	return u.String(), nil
}

type RemoveOption minio.RemoveObjectOptions

func (c *{{.client_name}}) removeObject(ctx context.Context, objectName string, option RemoveOption) error {
	if err := c.conn.RemoveObject(ctx, c.config.Bucket, objectName, minio.RemoveObjectOptions(option)); err != nil {
		return err
	}

	return nil
}

type ListOption minio.ListObjectsOptions

func (c *{{.client_name}}) listObjects(ctx context.Context, option ListOption) <-chan minio.ObjectInfo {
	return c.conn.ListObjects(ctx, c.config.Bucket, minio.ListObjectsOptions(option))
}
`
)

func createFxMinioFile(path string, name string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	name = strings.ToUpper(name[:1]) + name[1:]

	packageName := filepath.Base(path)
	if err := template.WriteTemplate2File(filepath.Join(path, fmt.Sprintf(fxFileName, name)), minioClientTemplate, map[string]any{
		"package_name": packageName,
		"client_name":  name,
	}); err != nil {
		return err
	}

	if err := module.GetGoModule(minioDependency); err != nil {
		return fmt.Errorf("getGoModule: %w", err)
	}

	return nil
}
