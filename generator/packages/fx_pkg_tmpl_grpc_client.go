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
	pkgTemplateList["[Grpc] grpc client"] = createFxGrpcClientFile
}

const (
	grpcClientDependency         = "google.golang.org/grpc"
	grpcClientDependencyProtobuf = "google.golang.org/protobuf"
	grpcClientTemplate           = `package {{.package_name}}

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
		return &Config{
			Addr: os.Getenv("GRPC_{{.client_name}}_ADDR"),
		}
	}
}

type Param struct {
	fx.In
	Cfg *Config
}

type Config struct {
	Addr string
}

type {{.client_name}} struct {
	client *grpc.ClientConn
}

func New(ctx context.Context, lc fx.Lifecycle, param Param) *{{.client_name}} {
	cli := &{{.client_name}}{}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			conn, err := grpc.NewClient(param.Cfg.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				return fmt.Errorf("grpc.Dial: %w", err)
			}

			cli.client = conn

			return nil
		},
		OnStop: func(ctx context.Context) error {
			if err := cli.client.Close(); err != nil {
				return fmt.Errorf("grpc.ClientConn.Close: %w", err)
			}

			return nil
		},
	})

	return cli
}`
)

func createFxGrpcClientFile(path string, name string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	name = strings.ToUpper(name[:1]) + name[1:]

	packageName := filepath.Base(path)
	if err := template.WriteTemplate2File(filepath.Join(path, fmt.Sprintf(fxFileName, name)), grpcClientTemplate, map[string]any{
		"package_name": packageName,
		"client_name":  name,
	}); err != nil {
		return err
	}

	if err := module.GetGoModule(grpcClientDependency); err != nil {
		return fmt.Errorf("getGoModule: %w", err)
	}

	if err := module.GetGoModule(grpcClientDependencyProtobuf); err != nil {
		return fmt.Errorf("getGoModule: %w", err)
	}

	return nil
}
