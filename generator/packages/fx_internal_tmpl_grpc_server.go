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
	internalControllerTemplateList["[Grpc/Server] grpc server"] = createFxGrpcFile
}

const (
	grpcServerDependency             = `google.golang.org/grpc`
	grpcServerDependencyProtobuf     = `google.golang.org/protobuf`
	grpcServerDependencyProxyProtoV2 = `github.com/pires/go-proxyproto`
	grpcServerTemplate               = `package {{.package_name}}

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/pires/go-proxyproto"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const Name = "{{.server_name}}"

var Module = fx.Module("{{.package_name}}",
	fx.Provide(New, ConfigRegister()),
	fx.Invoke(func(svr *{{.server_name}}) {}),
)

func ConfigRegister() func() *Config {
	opts := []grpc.ServerOption{
		grpc.Creds(insecure.NewCredentials()),
	}

	return func() *Config {
		return &Config{
			Addr: os.Getenv("GRPC_{{.server_name}}_ADDR"),
			Opts: opts,
		}
	}
}

type Param struct {
	fx.In
	Cfg *Config
}

type {{.server_name}} struct {
}

type Config struct {
	Addr string
	Opts []grpc.ServerOption
}

func New(ctx context.Context, lc fx.Lifecycle, param Param) *{{.server_name}} {
	grpcServer := grpc.NewServer(param.Cfg.Opts...)
	svr := &{{.server_name}}{}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Register your gRPC services here

			lis, err := net.Listen("tcp", param.Cfg.Addr)
			if err != nil {
				return fmt.Errorf("failed to listen on %s: %w", param.Cfg.Addr, err)
			}

			go func() {
				proxyListener := &proxyproto.Listener{Listener: lis}
				
				if err := grpcServer.Serve(proxyListener); err != nil {
					log.Fatalf("failed to start server: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			grpcServer.GracefulStop()
			return nil
		},
	})

	return svr
}`
)

func createFxGrpcFile(path string, name string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	name = strings.ToUpper(name[:1]) + name[1:]

	packageName := filepath.Base(path)
	if err := template.WriteTemplate2File(filepath.Join(path, fmt.Sprintf(fxFileName, name)), grpcServerTemplate, map[string]any{
		"package_name": packageName,
		"server_name":  name,
	}); err != nil {
		return err
	}

	if err := module.GetGoModule(grpcServerDependency); err != nil {
		return fmt.Errorf("getGoModule: %w", err)
	}

	if err := module.GetGoModule(grpcServerDependencyProtobuf); err != nil {
		return fmt.Errorf("getGoModule: %w", err)
	}

	if err := module.GetGoModule(grpcServerDependencyProxyProtoV2); err != nil {
		return fmt.Errorf("getGoModule: %w", err)
	}

	return nil
}
