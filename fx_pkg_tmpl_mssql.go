package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	pkgTemplateList["[SQL] mssql"] = createFxMSSQLFile
}

const (
	mssqlDependency     = "github.com/microsoft/go-mssqldb"
	mssqlClientTemplate = `package {{.package_name}}

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/microsoft/go-mssqldb/azuread"
	"go.uber.org/fx"
)

// Register is the fx.Provide function for the client.
// It registers the client as a dependency in the fx application.
// You can append interfaces into the fx.As() function to register multiple interfaces.
var Register = fx.Provide(fx.Annotate(New, fx.As()), ConfigRegister())

func ConfigRegister() func() *Config {
	return func() *Config {
		return &Config{
			DriverName:       DriverNameMSSQL,
			ConnectionString: os.Getenv("MSSQL_{{.client_name}}_CONNECTION_STRING"),
			ConnectionNumber: 10,
		}
	}
}

type Param struct {
	fx.In
	Cfg *Config
}

type Config struct {
	DriverName       DriverName
	ConnectionString string
	ConnectionNumber int
}

type DriverName string

const (
	DriverNameMSSQL   DriverName = "mssql"
	DriverNameAzureAD DriverName = azuread.DriverName
)

type {{.client_name}} struct {
	conn *sql.DB
}

func New(ctx context.Context, lc fx.Lifecycle, param Param) *{{.client_name}} {
	cli := &{{.client_name}}{}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			db, err := sql.Open(string(param.Cfg.DriverName), param.Cfg.ConnectionString)
			if err != nil {
				return fmt.Errorf("sql.Open: %w", err)
			}

			cli.conn = db

			return nil
		},
		OnStop: func(ctx context.Context) error {
			if err := cli.conn.Close(); err != nil {
				return fmt.Errorf("sql.Close: %w", err)
			}

			return nil
		},
	})

	return cli
}`
	mssqlConnectionStringTemplate = `package {{.package_name}}

import (
	"github.com/microsoft/go-mssqldb/msdsn"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	scheme      = "sqlserver://"
	defaultPort = 1433
)

type ConnectionStringBuilder struct {
	host                   string
	port                   int
	username               *string
	password               *string
	instance               *string
	database               *string
	ConnectionTimeout      *time.Duration
	TrustServerCertificate *bool
	encrypt                *string
}

func NewConnectionStringBuilder(host string, port int) *ConnectionStringBuilder {
	if port == 0 {
		port = defaultPort
	}

	return &ConnectionStringBuilder{
		host: host,
		port: port,
	}
}

func (b *ConnectionStringBuilder) WithUserAuth(username string, password string) *ConnectionStringBuilder {
	b.username = &username
	b.password = &password
	return b
}

func (b *ConnectionStringBuilder) WithInstance(instance string) *ConnectionStringBuilder {
	b.instance = &instance
	return b
}

func (b *ConnectionStringBuilder) WithDatabase(database string) *ConnectionStringBuilder {
	b.database = &database
	return b
}

func (b *ConnectionStringBuilder) WithConnectionTimeout(connectionTimeout time.Duration) *ConnectionStringBuilder {
	b.ConnectionTimeout = &connectionTimeout
	return b
}

func (b *ConnectionStringBuilder) WithTrustServerCertificate(trustServerCertificate bool) *ConnectionStringBuilder {
	b.TrustServerCertificate = &trustServerCertificate
	return b
}

func (b *ConnectionStringBuilder) WithEncrypt(encrypt string) *ConnectionStringBuilder {
	b.encrypt = &encrypt
	return b
}

func (b *ConnectionStringBuilder) BuildToURL() string {
	q := url.Values{}
	if b.database != nil {
		q.Add(msdsn.Database, *b.database)
	}
	if b.ConnectionTimeout != nil {
		q.Add(msdsn.ConnectionTimeout, strconv.FormatInt(int64(*b.ConnectionTimeout/time.Second), 10))
	}
	if b.TrustServerCertificate != nil {
		q.Add(msdsn.TrustServerCertificate, strconv.FormatBool(*b.TrustServerCertificate))
	}
	if b.encrypt != nil {
		q.Add(msdsn.Encrypt, *b.encrypt)
	}

	u := url.URL{
		Scheme: scheme,
		Host:   b.host + ":" + strconv.FormatInt(int64(b.port), 10),
	}

	if b.username != nil && b.password != nil {
		u.User = url.UserPassword(*b.username, *b.password)
	}

	if b.instance != nil {
		u.Path = *b.instance
	}

	u.RawQuery = q.Encode()

	return u.String()
}

func (b *ConnectionStringBuilder) BuildToString() string {
	builder := strings.Builder{}
	builder.WriteString("Server=")
	builder.WriteString(b.host)
	builder.WriteString(",")
	builder.WriteString(strconv.FormatInt(int64(b.port), 10))
	if b.instance != nil {
		builder.WriteString("\\")
		builder.WriteString(*b.instance)
	}
	if b.username != nil && b.password != nil {
		builder.WriteString(";User Id=")
		builder.WriteString(*b.username)
		builder.WriteString(";Password=")
		builder.WriteString(*b.password)
	}
	if b.database != nil {
		builder.WriteString(";Database=")
		builder.WriteString(*b.database)
	}
	if b.ConnectionTimeout != nil {
		builder.WriteString(";Connect Timeout=")
		builder.WriteString(strconv.FormatInt(int64(*b.ConnectionTimeout/time.Second), 10))
	}
	if b.TrustServerCertificate != nil {
		builder.WriteString(";Trust Server Certificate=")
		switch *b.TrustServerCertificate {
		case true:
			builder.WriteString("True")
		case false:
			builder.WriteString("False")
		}
	}
	if b.encrypt != nil {
		builder.WriteString(";Encrypt=")
		builder.WriteString(*b.encrypt)
	}

	return builder.String()
}`
)

func createFxMSSQLFile(path string, name string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	name = strings.ToUpper(name[:1]) + name[1:]

	packageName := filepath.Base(path)
	if err := writeTemplate2File(filepath.Join(path, fmt.Sprintf(fxFileName, name)), mssqlClientTemplate, map[string]any{
		"package_name": packageName,
		"client_name":  name,
	}); err != nil {
		return err
	}

	if err := writeTemplate2File(filepath.Join(path, "mssql_connection_string.go"), mssqlConnectionStringTemplate, map[string]any{
		"package_name": packageName,
	}); err != nil {
		return err
	}

	if err := getGoModule(mssqlDependency); err != nil {
		return fmt.Errorf("getGoModule: %w", err)
	}

	return nil
}
