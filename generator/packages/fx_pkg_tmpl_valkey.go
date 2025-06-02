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
	pkgTemplateList["[Cache/Valkey] valkey"] = createFxValkeyFile
}

const (
	valkeyDependency     = "github.com/valkey-io/valkey-go@v1"
	valkeyClientTemplate = `package {{.package_name}}

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/valkey-io/valkey-go"
	"go.uber.org/fx"
)

const Name = "{{.client_name}}"

var Module = fx.Module("{{.package_name}}",
	fx.Provide(ConfigRegister()),
	fx.Provide(fx.Annotate(New, fx.As(/* implemented interfaces */))),
)

// ConfigRegister is the fx.Provide function for the config.
// Modify the config according to your needs.
func ConfigRegister() func() *Config {
	return func() *Config {
		addr := os.Getenv("VALKEY_{{.client_name}}_ADDR")
		username := os.Getenv("VALKEY_{{.client_name}}_USERNAME")
		password := os.Getenv("VALKEY_{{.client_name}}_PASSWORD")

		return &Config{
			Addr:     strings.Split(addr, ","),
			Username: username,
			Password: password,
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
}

type {{.client_name}} struct {
	conn valkey.Client
}

func New(ctx context.Context, lc fx.Lifecycle, param Param) *{{.client_name}} {
	cli := &{{.client_name}}{}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			conn, err := valkey.NewClient(valkey.ClientOption{
				InitAddress: param.Cfg.Addr,
				Username:    param.Cfg.Username,
				Password:    param.Cfg.Password,
			})
			if err != nil {
				return fmt.Errorf("valkey.New: %w", err)
			}

			cli.conn = conn

			return nil
		},
		OnStop: func(ctx context.Context) error {
			// Clean up resources if needed
			cli.conn.Close()
			return nil
		},
	})

	return cli
}`
)

func createFxValkeyFile(path string, name string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	name = strings.ToUpper(name[:1]) + name[1:]

	packageName := filepath.Base(path)
	if err := template.WriteTemplate2File(filepath.Join(path, fmt.Sprintf(fxFileName, name)), valkeyClientTemplate, map[string]any{
		"package_name": packageName,
		"client_name":  name,
	}); err != nil {
		return err
	}

	if err := module.GetGoModule(valkeyDependency); err != nil {
		return fmt.Errorf("getGoModule: %w", err)
	}

	return nil
}
