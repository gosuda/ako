package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	pkgTemplateList["[MessageQueue/NATS] nats"] = createFxNatsFile
}

const (
	natsDependency     = `github.com/nats-io/nats.go`
	natsClientTemplate = `package {{.package_name}}

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"

	"github.com/nats-io/nats.go"
	"go.uber.org/fx"
)

// Register is the fx.Provide function for the client.
// It registers the client as a dependency in the fx application.
// You can append interfaces into the fx.As() function to register multiple interfaces.
var Register = fx.Provide(fx.Annotate(New, fx.As()), ConfigRegister())

// ConfigRegister is the fx.Provide function for the config.
// Modify the config according to your needs.
func ConfigRegister() func() *Config {
	return func() *Config {
		u := os.Getenv("NATS_{{.client_name}}_URL")
		if len(u) == 0 {
			u = nats.DefaultURL
		}

		return &Config{
			Url:      u,
			Username: os.Getenv("NATS_{{.client_name}}_USERNAME"),
			Password: os.Getenv("NATS_{{.client_name}}_PASSWORD"),
			Token:    os.Getenv("NATS_{{.client_name}}_TOKEN"),
		}
	}
}

type Param struct {
	fx.In
	Cfg *Config
}

type Config struct {
	Url       string // ex) 10.0.4.12:4222,10.0.4.13:4222,10.0.4.14:4222
	Username  string
	Password  string
	Token     string
	TlsConfig *tls.Config
}

type {{.client_name}} struct {
	conn *nats.Conn
}

func New(ctx context.Context, lc fx.Lifecycle, param Param) *{{.client_name}} {
	cli := &{{.client_name}}{}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			conn, err := nats.Connect(param.Cfg.Url, func(options *nats.Options) error {
				options.User = param.Cfg.Username
				options.Password = param.Cfg.Password
				options.Token = param.Cfg.Token
				options.TLSConfig = param.Cfg.TlsConfig

				return nil
			})
			if err != nil {
				return fmt.Errorf("nats.Connect: %w", err)
			}

			cli.conn = conn

			return nil
		},
		OnStop: func(ctx context.Context) error {
			cli.conn.Close()
			return nil
		},
	})

	return cli
}`
)

func createFxNatsFile(path string, name string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	name = strings.ToUpper(name[:1]) + name[1:]

	packageName := filepath.Base(path)
	if err := writeTemplate2File(filepath.Join(path, fmt.Sprintf(fxFileName, name)), natsClientTemplate, map[string]any{
		"package_name": packageName,
		"client_name":  name,
	}); err != nil {
		return err
	}

	if err := getGoModule(natsDependency); err != nil {
		return fmt.Errorf("getGoModule: %w", err)
	}

	return nil
}
