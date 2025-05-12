package packages

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"

	"github.com/gosuda/ako/util/module"
	"github.com/gosuda/ako/util/template"
)

func init() {
	pkgTemplateList["[SQL/Generation] sqlc (mysql, postgresql, postgis, pgvector)"] = createFxSqlcFile
}

const (
	sqlcPostgresDependencyPgxV5  = `github.com/jackc/pgx/v5`
	sqlcPostgresDependencyGoGeom = `github.com/twpayne/go-geom`
	sqlcPostgresConfigTemplate   = `version: "2"
sql:
  - engine: "postgresql"
    queries:
      - "./queries/queries.sql"
    schema:
      - "./queries/schema.sql"
    gen:
      go:
        package: "queries"
        sql_package: "pgx/v5"
        out: "queries"
        overrides:
          - "db_type": "geometry"
            "go_type":
              "import": "github.com/twpayne/go-geom"
              "package": "geom"
              "type": "T"
plugins: []
rules: []
options: {}
`
	sqlcMysqlDependency     = `github.com/go-sql-driver/mysql`
	sqlcMysqlConfigTemplate = `version: "2"
sql:
  - engine: "mysql"
    queries: "queries/queries.sql"
    schema: "queries/schema.sql"
    gen:
      go:
        package: "queries"
        out: "queries"
`

	sqlcToolDependency               = `github.com/sqlc-dev/sqlc/cmd/sqlc`
	sqlcPostgresGenerateFileTemplate = `package {{.package_name}}

//go:generate go tool sqlc generate

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
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
			ConnectionString: os.Getenv("SQLC_{{.client_name}}_CONNECTION_STRING"),
			DriverName:       os.Getenv("SQLC_{{.client_name}}_DRIVER_NAME"),
		}
	}
}

type Param struct {
	fx.In
	Cfg *Config
}

type Config struct {
	ConnectionString string
	DriverName       string
}

type {{.client_name}} struct {
	conn    *pgx.Conn
	queries *queries.Queries
}

func New(ctx context.Context, lc fx.Lifecycle, param Param) *{{.client_name}} {
	cli := &{{.client_name}}{}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			conn, err := pgx.Connect(ctx, param.Cfg.ConnectionString)
			if err != nil {
				return fmt.Errorf("pgx.Connect: %w", err)
			}

			cli.conn = conn
			cli.queries = queries.New(conn)

			return nil
		},
		OnStop: func(ctx context.Context) error {
			if err := cli.conn.Close(ctx); err != nil {
				return fmt.Errorf("conn.Close: %w", err)
			}

			return nil
		},
	})

	return cli
}
`
	sqlcMySqlGenerateFileTemplate = `package {{.package_name}}

//go:generate go tool sqlc generate

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/fx"
)

// Register is the fx.Provide function for the client.
// It registers the client as a dependency in the fx application.
// You can append interfaces into the fx.As() function to register multiple interfaces.
var Register = fx.Provide(fx.Annotate(New, fx.As()))

// ConfigRegister is the fx.Provide function for the config.
// Modify the config according to your needs.
func ConfigRegister() func() *Config {
	return func() *Config {
		return &Config{
			ConnectionString: os.Getenv("SQLC_{{.client_name}}_CONNECTION_STRING"),
			DriverName:       os.Getenv("SQLC_{{.client_name}}_DRIVER_NAME"),
		}
	}
}

type Param struct {
	fx.In
	Cfg *Config
}

type Config struct {
	ConnectionString string
	DriverName       string
}

type {{.client_name}} struct {
	db      *sql.DB
	queries *queries.Queries
}

func New(ctx context.Context, lc fx.Lifecycle, param Param) *{{.client_name}} {
	cli := &{{.client_name}}{}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			db, err := sql.Open(param.Cfg.DriverName, param.Cfg.ConnectionString)
			if err != nil {
				return fmt.Errorf("sql.Open: %w", err)
			}

			cli.db = db
			cli.queries = queries.New(db)

			return nil
		},
		OnStop: func(ctx context.Context) error {
			if err := cli.db.Close(ctx); err != nil {
				return fmt.Errorf("db.Close: %w", err)
			}

			return nil
		},
	})

	return cli
}
`
	sqlcQueriesFolder          = `queries`
	sqlcQueriesQueriesTemplate = `-- name: GetPerson :one
SELECT * FROM person WHERE id = $1;

-- name: AllPersons :many
SELECT * FROM person;

-- name: AddPerson :exec
INSERT INTO person (name, age) VALUES ($1, $2);

-- name: UpdatePerson :exec
UPDATE person SET name = $1, age = $2 WHERE id = $3;

-- name: UpsertPerson :exec
INSERT INTO person (name, age) VALUES ($1, $2);

-- name: DeletePerson :exec
DELETE FROM person WHERE id = $1;`
	sqlcQueriesSchemaTemplate = `CREATE TABLE IF NOT EXISTS person
(
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    age INT NOT NULL
);`
)

const (
	SqlcDatabaseEnginePostgres = "postgres"
	SqlcDatabaseEngineMysql    = "mysql"
)

func selectSqlcDatabaseEngine() (string, error) {
	var engine string
	if err := survey.AskOne(&survey.Select{
		Message: "Select the database engine for sqlc:",
		Options: []string{SqlcDatabaseEnginePostgres, SqlcDatabaseEngineMysql},
	}, &engine); err != nil {
		return "", err
	}

	engine = strings.TrimSpace(engine)

	return engine, nil
}

func createFxSqlcFile(path string, name string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	engine, err := selectSqlcDatabaseEngine()
	if err != nil {
		return fmt.Errorf("selectSqlcDatabaseEngine: %w", err)
	}

	name = strings.ToUpper(name[:1]) + name[1:]
	packageName := filepath.Base(path)

	configTemplate := ""
	clientTemplate := ""
	switch engine {
	case SqlcDatabaseEnginePostgres:
		configTemplate = sqlcPostgresConfigTemplate
		clientTemplate = sqlcPostgresGenerateFileTemplate
	case SqlcDatabaseEngineMysql:
		configTemplate = sqlcMysqlConfigTemplate
		clientTemplate = sqlcMySqlGenerateFileTemplate
	}

	if err := template.WriteTemplate2File(filepath.Join(path, "sqlc.yaml"), configTemplate, map[string]any{
		"package_name": packageName,
	}); err != nil {
		return err
	}

	if err := template.WriteTemplate2File(filepath.Join(path, fmt.Sprintf(fxFileName, name)), clientTemplate, map[string]any{
		"package_name": packageName,
		"client_name":  name,
	}); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(path, sqlcQueriesFolder), os.ModePerm); err != nil {
		return err
	}

	if err := template.WriteTemplate2File(filepath.Join(path, sqlcQueriesFolder, "queries.sql"), sqlcQueriesQueriesTemplate, nil); err != nil {
		return err
	}

	if err := template.WriteTemplate2File(filepath.Join(path, sqlcQueriesFolder, "schema.sql"), sqlcQueriesSchemaTemplate, nil); err != nil {
		return err
	}

	switch engine {
	case SqlcDatabaseEnginePostgres:
		if err := module.GetGoModule(sqlcPostgresDependencyPgxV5); err != nil {
			return fmt.Errorf("getGoModule: %w", err)
		}

		if err := module.GetGoModule(sqlcPostgresDependencyGoGeom); err != nil {
			return fmt.Errorf("getGoModule: %w", err)
		}
	case SqlcDatabaseEngineMysql:
		if err := module.GetGoModule(sqlcMysqlDependency); err != nil {
			return fmt.Errorf("getGoModule: %w", err)
		}
	}

	if err := module.GetGoModuleAsTool(sqlcToolDependency); err != nil {
		return fmt.Errorf("getGoModule: %w", err)
	}

	cmd := exec.Command("go", "tool", "sqlc", "generate")
	cmd.Dir = path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go tool sqlc generate: %w", err)
	}

	return nil
}
