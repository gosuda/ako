package packages

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gosuda/ako/module"
	"github.com/gosuda/ako/template"
)

func init() {
	pkgTemplateList["[VectorDB] qdrant"] = createFxQdrantFile
}

const (
	qdrantDependency     = `github.com/qdrant/go-client`
	qdrantClientTemplate = `package {{.package_name}}

import (
	"context"
	"crypto/tls"
	"os"
	"strconv"

	"github.com/qdrant/go-client/qdrant"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

// Register is the fx.Provide function for the client.
// It registers the client as a dependency in the fx application.
// You can append interfaces into the fx.As() function to register multiple interfaces.
var Register = fx.Provide(fx.Annotate(New, fx.As()), ConfigRegister())

// ConfigRegister is the fx.Provide function for the config.
// Modify the config according to your needs.
func ConfigRegister() func() *Config {
	return func() *Config {
		grpcOpts := ([]grpc.DialOption)(nil)
		tlsConfig := (*tls.Config)(nil)
		port, _ := strconv.Atoi(os.Getenv("QDRANT_{{.client_name}}_PORT"))

		return &Config{
			Host: os.Getenv("QDRANT_{{.client_name}}_HOST"),
			Port: port,
			ApiKey: os.Getenv("QDRANT_{{.client_name}}_API_KEY"),
			TlsConfig: tlsConfig,
			GrpcOptions: grpcOpts,
		}
	}
}

type Param struct {
	fx.In
	Cfg *Config
}

type Config struct {
	Host        string
	Port        int
	ApiKey      string
	TlsConfig   *tls.Config
	GrpcOptions []grpc.DialOption
}

type {{.client_name}} struct {
	client *qdrant.Client
}

func New(ctx context.Context, lc fx.Lifecycle, param Param) *{{.client_name}} {
	cli := &{{.client_name}}{}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			cfg := &qdrant.Config{
				Host: param.Cfg.Host,
				Port: param.Cfg.Port,
				APIKey: param.Cfg.ApiKey,
			}
			if param.Cfg.TlsConfig != nil {
				cfg.UseTLS = true
				cfg.TLSConfig = param.Cfg.TlsConfig
			}
			if param.Cfg.GrpcOptions != nil {
				cfg.GrpcOptions = param.Cfg.GrpcOptions
			}

			client, err := qdrant.NewClient(cfg)
			if err != nil {
				return err
			}

			cli.client = client
			
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if err := cli.client.Close(); err != nil {
				return err
			}			

			return nil
		},
	})

	return cli
}`
)

func createFxQdrantFile(path string, name string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	name = strings.ToUpper(name[:1]) + name[1:]

	packageName := filepath.Base(path)
	if err := template.WriteTemplate2File(filepath.Join(path, fmt.Sprintf(fxFileName, name)), qdrantClientTemplate, map[string]any{
		"package_name": packageName,
		"client_name":  name,
	}); err != nil {
		return err
	}

	if err := module.GetGoModule(qdrantDependency); err != nil {
		return fmt.Errorf("getGoModule: %w", err)
	}

	return nil
}
