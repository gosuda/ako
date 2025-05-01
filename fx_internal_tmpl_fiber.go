package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	internalControllerTemplateList["fiber (http, https)"] = createFxFiberFile
}

const (
	fiberDependency     = `github.com/gofiber/fiber/v2`
	fiberServerTemplate = `package {{.package_name}}

import (
	"context"
	"os"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
)

// Register is the fx.Provide function for the client.
// It registers the client as a dependency in the fx application.
// You can append interfaces into the fx.As() function to register multiple interfaces.
var Register = fx.Provide(New, ConfigRegister())

func ConfigRegister() func() *Config {
	return func() *Config {
		return &Config{
			Addr: os.Getenv("FIBER_{{.server_name}}_ADDR"),
		}
	}
}

type Param struct {
	fx.In
}

type {{.server_name}} struct {
	app *fiber.App
}

type Config struct {
	Addr string
}

func New(ctx context.Context, lc fx.Lifecycle, param Param) *{{.server_name}} {
	svr := &{{.server_name}}{}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			svr.app = fiber.New()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return svr.app.Shutdown()
		},
	})

	return svr
}`
)

func createFxFiberFile(path string, name string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	name = strings.ToUpper(name[:1]) + name[1:]

	packageName := filepath.Base(path)
	if err := writeTemplate2File(filepath.Join(path, fmt.Sprintf(fxFileName, name)), fiberServerTemplate, map[string]any{
		"package_name": packageName,
		"server_name":  name,
	}); err != nil {
		return err
	}

	if err := getGoModule(fiberDependency); err != nil {
		return fmt.Errorf("getGoModule: %w", err)
	}

	return nil
}
