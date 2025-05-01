package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	pkgTemplateList["valkey"] = createFxValkeyFile
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

// Register is the fx.Provide function for the client.
// It registers the client as a dependency in the fx application.
// You can append interfaces into the fx.As() function to register multiple interfaces.
var Register = fx.Provide(fx.Annotate(New, fx.As()), ConfigRegister())

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

func New(ctx context.Context, lc fx.Lifecycle, param Param) (*{{.client_name}}, error) {
	conn, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: param.Cfg.Addr,
		Username:    param.Cfg.Username,
		Password:    param.Cfg.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("valkey.New: %w", err)
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Initialize the client here if needed
			return nil
		},
		OnStop: func(ctx context.Context) error {
			// Clean up resources if needed
			conn.Close()
			return nil
		},
	})

	return &{{.client_name}}{
		conn: conn,
	}, nil
}`
)

func createFxValkeyFile(path string, name string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	name = strings.ToUpper(name[:1]) + name[1:]

	packageName := filepath.Base(path)
	if err := writeTemplate2File(filepath.Join(path, fmt.Sprintf(fxFileName, name)), valkeyClientTemplate, map[string]any{
		"package_name": packageName,
		"client_name":  name,
	}); err != nil {
		return err
	}

	if err := getGoModule(valkeyDependency); err != nil {
		return fmt.Errorf("getGoModule: %w", err)
	}

	return nil
}
