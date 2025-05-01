package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	internalControllerTemplateList["chi (http, https)"] = createFxChiFile
}

const (
	chiDependency             = `github.com/go-chi/chi/v5`
	chiDependencyProxyProtoV2 = `github.com/pires/go-proxyproto`
	chiServerTemplate         = `package {{.package_name}}

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pires/go-proxyproto"
	"go.uber.org/fx"
)

// Register is the fx.Provide function for the client.
// It registers the client as a dependency in the fx application.
// You can append interfaces into the fx.As() function to register multiple interfaces.
var Register = fx.Provide(New, ConfigRegister())

func ConfigRegister() func() *Config {
	return func() *Config {
		return &Config{
			Addr: os.Getenv("CHI_{{.server_name}}_ADDR"),
		}
	}
}

type Param struct {
	fx.In
	Cfg *Config
}

type {{.server_name}} struct {
	mux *chi.Mux
	svr *http.Server
}

type Config struct {
	Addr string
}

func New(ctx context.Context, lc fx.Lifecycle, param Param) *{{.server_name}} {
	mux := chi.NewRouter()
	mux.Use(middleware.Heartbeat("/ping"))
	mux.Use(middleware.Recoverer)
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Timeout(60 * time.Second))

	svr := &{{.server_name}}{
		mux: mux,
		svr: &http.Server{
			Addr:    param.Cfg.Addr,
			Handler: mux,
		},
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			lis, err := net.Listen("tcp", param.Cfg.Addr)
			if err != nil {
				return fmt.Errorf("net.Listen: %w", err)
			}

			go func() {
				proxyListener := &proxyproto.Listener{Listener: lis}

				if err := svr.svr.Serve(proxyListener); err != nil {
					log.Printf("failed to set listener: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return svr.svr.Shutdown(ctx)
		},
	})

	return svr
}`
)

func createFxChiFile(path string, name string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	name = strings.ToUpper(name[:1]) + name[1:]

	packageName := filepath.Base(path)
	if err := writeTemplate2File(filepath.Join(path, fmt.Sprintf(fxFileName, name)), chiServerTemplate, map[string]any{
		"package_name": packageName,
		"server_name":  name,
	}); err != nil {
		return err
	}

	if err := getGoModule(chiDependency); err != nil {
		return fmt.Errorf("getGoModule: %w", err)
	}

	if err := getGoModule(chiDependencyProxyProtoV2); err != nil {
		return fmt.Errorf("getGoModule: %w", err)
	}

	return nil
}
