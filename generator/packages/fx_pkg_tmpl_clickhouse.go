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
	pkgTemplateList["[NoSQL/Columnar] clickhouse"] = createClickhouseStructFile
}

const (
	clickhouseDependency     = `github.com/ClickHouse/clickhouse-go/v2`
	clickhouseClientTemplate = `package {{.package_name}}

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"go.uber.org/fx"
)

var Module = fx.Module("{{.package_name}}",
	fx.Provide(ConfigRegister()),
	fx.Provide(fx.Annotate(New, fx.As(/* implemented interfaces */))),
)

// ConfigRegister is the fx.Provide function for the config.
// Modify the config according to your needs.
func ConfigRegister() func() *Config {
	return func() *Config {
		addr := os.Getenv("CLICKHOUSE_{{.client_name}}_ADDR")
		username := os.Getenv("CLICKHOUSE_{{.client_name}}_USERNAME")
		password := os.Getenv("CLICKHOUSE_{{.client_name}}_PASSWORD")
		database := os.Getenv("CLICKHOUSE_{{.client_name}}_DATABASE")
		insecureSkipVerify := os.Getenv("CLICKHOUSE_{{.client_name}}_INSECURE_SKIP_VERIFY")
		isk, _ := strconv.ParseBool(insecureSkipVerify)

		return &Config{
			Addr:     strings.Split(addr, ","),
			Username: username,
			Password: password,
			Database: database,
			InsecureSkipVerify: isk,
		}
	}
}

type Param struct {
	fx.In
	Cfg *Config
}

type Config struct {
	Addr     []string
	Username string
	Password string
	Database string
	InsecureSkipVerify bool

	// Add any other configuration options you need
	ClientInfo clickhouse.ClientInfo
}

type {{.client_name}} struct {
	conn driver.Conn
}

func New(ctx context.Context, lc fx.Lifecycle, param Param) *{{.client_name}} {
	cli := &{{.client_name}}{}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			opt := clickhouse.Options{}
			if param.Cfg.Addr != nil {
				opt.Addr = param.Cfg.Addr
			}
			if len(param.Cfg.Database) > 0 {
				opt.Auth.Database = param.Cfg.Database
			}
			if len(param.Cfg.Username) > 0 {
				opt.Auth.Username = param.Cfg.Username
			}
			if len(param.Cfg.Password) > 0 {
				opt.Auth.Password = param.Cfg.Password
			}
			if param.Cfg.InsecureSkipVerify {
				opt.TLS.InsecureSkipVerify = param.Cfg.InsecureSkipVerify
			}
			if len(param.Cfg.ClientInfo.Products) > 0 {
				opt.ClientInfo = param.Cfg.ClientInfo
			}
		
			conn, err := clickhouse.Open(&opt)
			if err != nil {
				return fmt.Errorf("failed to open connection: %w", err)
			}

			cli.conn = conn
			if err := cli.conn.Ping(ctx); err != nil {
				return fmt.Errorf("failed to ping connection: %w", err)
			}

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return cli.conn.Close()
		},
	})

	return cli
}`
)

func createClickhouseStructFile(path string, name string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	name = strings.ToUpper(name[:1]) + name[1:]

	packageName := filepath.Base(path)
	if err := template.WriteTemplate2File(filepath.Join(path, fmt.Sprintf(fxFileName, name)), clickhouseClientTemplate, map[string]any{
		"package_name": packageName,
		"client_name":  name,
	}); err != nil {
		return err
	}

	if err := module.GetGoModule(clickhouseDependency); err != nil {
		return fmt.Errorf("getGoModule: %w", err)
	}

	return nil
}
