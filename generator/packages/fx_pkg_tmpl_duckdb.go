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
	pkgTemplateList["[Analyze/Transform] duckdb"] = createFxDuckDBFile
}

const (
	duckdbDependency     = `github.com/marcboeker/go-duckdb/v2`
	duckdbClientTemplate = `package {{.package_name}}
import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/marcboeker/go-duckdb/v2"
	"go.uber.org/fx"
)

const driverName = "duckdb"

// Register is the fx.Provide function for the client.
// It registers the client as a dependency in the fx application.
// You can append interfaces into the fx.As() function to register multiple interfaces.
var Register = fx.Provide(fx.Annotate(New, fx.As()), ConfigRegister())

// ConfigRegister is the fx.Provide function for the config.
// Modify the config according to your needs.
func ConfigRegister() func() *Config {
	return func() *Config {
		return &Config{
			Datasource: os.Getenv("DUCKDB_DATASOURCE"),
		}
	}
}

type Param struct {
	fx.In
	Cfg *Config
}

type Config struct {
	Datasource string
}

type {{.client_name}} struct {
	db *sql.DB
}

func New(ctx context.Context, lc fx.Lifecycle, param Param) *{{.client_name}} {
	cli := &{{.client_name}}{}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			db, err := sql.Open(driverName, param.Cfg.Datasource)
			if err != nil {
				return fmt.Errorf("sql.Open: %w", err)
			}

			cli.db = db

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return cli.db.Close()
		},
	})

	return cli
}`
)

func createFxDuckDBFile(path string, name string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	name = strings.ToUpper(name[:1]) + name[1:]

	packageName := filepath.Base(path)
	if err := template.WriteTemplate2File(filepath.Join(path, fmt.Sprintf(fxFileName, name)), duckdbClientTemplate, map[string]any{
		"package_name": packageName,
		"client_name":  name,
	}); err != nil {
		return err
	}

	if err := module.GetGoModule(duckdbDependency); err != nil {
		return fmt.Errorf("getGoModule: %w", err)
	}

	return nil
}
