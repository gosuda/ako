package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	pkgTemplateList["[SQL/ORM] ent (sqlite, mysql, postgresql)"] = createFxEntgoFile
}

const (
	entgoGenerateFileName = "ent.generate.go"
	entgoToolDependency   = "entgo.io/ent/cmd/ent"
	entgoDependency       = "entgo.io/ent"
	entgoClientTemplate   = `package {{.package_name}}

import (
	"context"
	"fmt"
	"os"

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
		return &Config{
			ConnectionString: os.Getenv("ENTGO_{{.client_name}}_CONNECTION_STRING"),
			Driver:           os.Getenv("ENTGO_{{.client_name}}_DRIVER"),
		}
	}
}

type Param struct {
	fx.In
	Cfg *Config
}

type Config struct {
	ConnectionString string
	Driver string
}

type {{.client_name}} struct {
	client *ent.Client
}

func New(ctx context.Context, lc fx.Lifecycle, param Param) *{{.client_name}} {
	cli := &{{.client_name}}{}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			client, err := ent.Open(param.Cfg.Driver, param.Cfg.ConnectionString)
			if err != nil {
				return fmt.Errorf("ent.Open: %w", err)
			}

			cli.client = client

			if err := cli.client.Schema.Create(ctx); err != nil {
				return fmt.Errorf("ent.Schema.Create: %w", err)
			}

			return nil
		},
		OnStop: func(ctx context.Context) error {
			if err := cli.client.Close(); err != nil {
				return fmt.Errorf("ent.Close: %w", err)
			}

			return nil
		},
	})

	return &{{.client_name}}{}
}
`
	entgoGenerateTemplate = `package {{.package_name}}

//go:generate go tool ent new <entity_name>
`
)

func createFxEntgoFile(path string, name string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	name = strings.ToUpper(name[:1]) + name[1:]

	packageName := filepath.Base(path)
	if err := writeTemplate2File(filepath.Join(path, fmt.Sprintf(fxFileName, name)), entgoClientTemplate, map[string]any{
		"package_name": packageName,
		"client_name":  name,
	}); err != nil {
		return err
	}

	if err := writeTemplate2File(filepath.Join(path, entgoGenerateFileName), entgoGenerateTemplate, map[string]any{
		"package_name": packageName,
	}); err != nil {
		return err
	}

	if err := getGoModuleAsTool(entgoToolDependency); err != nil {
		return fmt.Errorf("getGoModule: %w", err)
	}

	if err := getGoModule(entgoDependency); err != nil {
		return fmt.Errorf("getGoModule: %w", err)
	}

	return nil
}
