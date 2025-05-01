package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	pkgTemplateList["cassandra"] = createFxCassandraFile
}

const (
	cassandraDependencyGocql  = "github.com/gocql/gocql"
	cassandraDependencyGocqlx = "github.com/scylladb/gocqlx/v2"
	cassandraClientTemplate   = `package {{.package_name}}

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v2"
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
		host := os.Getenv("CASSANDRA_HOST")
		if host == "" {
			host = "127.0.0.1:9042"
		}

		opt := []func(*gocql.ClusterConfig){
			func(c *gocql.ClusterConfig) {
				c.ProtoVersion = 4
			},
		}

		return &Config{
			Host:   strings.Split(host, ","),
			Option: opt,
		}
	}
}

type Param struct {
	fx.In
}

type Config struct {
	Host     []string
	Option   []func(*gocql.ClusterConfig)
}

type {{.client_name}} struct {
	clusterConfig gocql.ClusterConfig
	session *gocql.Session
	sessionX gocqlx.Session
}

func New(ctx context.Context, lc fx.Lifecycle, param Param) *{{.client_name}} {
	cli := &{{.client_name}}{
		clusterConfig: gocql.NewCluster(param.Cfg.Host...),
	}

	for _, opt := range param.Cfg.Option {
		opt(&cli.clusterConfig)
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			session, err := cli.clusterConfig.CreateSession()
			if err != nil {
				return fmt.Errorf("gocql.NewSession: %w", err)
			}
			cli.session = session

			sessionX := gocqlx.NewSession(session)
			if err != nil {
				return fmt.Errorf("gocqlx.NewSession: %w", err)
			}
			cli.sessionX = sessionX

			return nil
		},
		OnStop: func(ctx context.Context) error {
			if cli.session != nil {
				cli.session.Close()
			}

			return nil
		},
	})

	return &{{.client_name}}{}
}`
)

func createFxCassandraFile(path string, name string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	name = strings.ToUpper(name[:1]) + name[1:]

	packageName := filepath.Base(path)
	if err := writeTemplate2File(filepath.Join(path, fmt.Sprintf(fxFileName, name)), cassandraClientTemplate, map[string]any{
		"package_name": packageName,
		"client_name":  name,
	}); err != nil {
		return err
	}

	if err := getGoModule(cassandraDependencyGocql); err != nil {
		return fmt.Errorf("getGoModule: %w", err)
	}

	if err := getGoModule(cassandraDependencyGocqlx); err != nil {
		return fmt.Errorf("getGoModule: %w", err)
	}

	return nil
}
