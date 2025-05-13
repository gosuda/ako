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
	internalControllerTemplateList["[Empty] Empty"] = createFxInternalEmptyFile
}

const (
	fxFileName          = "init_%s.go"
	fxDependencyPackage = "go.uber.org/fx"
)

func GetFxDependency() error {
	if err := module.GetGoModuleAsTool(fxDependencyPackage); err != nil {
		return err
	}

	if err := module.TidyGoModule(); err != nil {
		return err
	}

	return nil
}

const fxStructFileTemplate = `package {{.package_name}}

import (
	"context"

	"go.uber.org/fx"
)

var Module = fx.Module("{{.package_name}}",
	fx.Provide(New, ConfigRegister()),
)

func ConfigRegister() func() *Config {
	return func() *Config {
		return &Config{}
	}
}

type Param struct {
	fx.In
}

type Config struct {
}

type {{.client_name}} struct {
}

func New(ctx context.Context, lc fx.Lifecycle, param Param) *{{.client_name}} {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Initialize the client here if needed
			return nil
		},
		OnStop: func(ctx context.Context) error {
			// Clean up resources if needed
			return nil
		},
	})

	return &{{.client_name}}{}
}`

const fxInternalEmptyFileTemplate = `package {{.package_name}}

import (
	"context"

	"go.uber.org/fx"
)

var Module = fx.Module(
	fx.Provide(New, ConfigRegister()),
	fx.Invoke(func(svr *{{.client_name}}) {}),
)

type Param struct {
	fx.In
}

type {{.client_name}} struct {
}

func New(ctx context.Context, lc fx.Lifecycle, param Param) *{{.client_name}} {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Initialize the client here if needed
			return nil
		},
		OnStop: func(ctx context.Context) error {
			// Clean up resources if needed
			return nil
		},
	})

	return &{{.client_name}}{}
}`

const fxInterfaceFileTemplate = `package {{.package_name}}

type {{.client_name}} interface {
}`

func createFxStructFile(path string, name string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	name = strings.ToUpper(name[:1]) + name[1:]

	packageName := filepath.Base(path)
	if err := template.WriteTemplate2File(filepath.Join(path, fmt.Sprintf(fxFileName, name)), fxStructFileTemplate, map[string]any{
		"package_name": packageName,
		"client_name":  name,
	}); err != nil {
		return err
	}

	return nil
}

func createFxInternalEmptyFile(path string, name string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	name = strings.ToUpper(name[:1]) + name[1:]

	packageName := filepath.Base(path)
	if err := template.WriteTemplate2File(filepath.Join(path, fmt.Sprintf(fxFileName, name)), fxInternalEmptyFileTemplate, map[string]any{
		"package_name": packageName,
		"client_name":  name,
	}); err != nil {
		return err
	}

	return nil
}

func createFxInterfaceFile(path string, name string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	name = strings.ToUpper(name[:1]) + name[1:]

	packageName := filepath.Base(path)
	if err := template.WriteTemplate2File(filepath.Join(path, fmt.Sprintf(fxFileName, name)), fxInterfaceFileTemplate, map[string]any{
		"package_name": packageName,
		"client_name":  name,
	}); err != nil {
		return err
	}

	return nil
}

const (
	fxExecutableFileName     = "main.go"
	fxExecutableFileTemplate = `package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.uber.org/fx"
)

func newStartupContext() context.Context {
	return context.Background()
}

func main() {
	app := fx.New(fx.Provide(newStartupContext),
		fx.StartTimeout(15*time.Second),
		fx.StopTimeout(15*time.Second),
		fx.Invoke(func() {
			fmt.Println("Hello, world!")
		}),
	)
	app.Run()

	if err := app.Err(); err != nil {
		log.Printf("app is exiting with error: %v", err)
	}
}`
)

func CreateFxExecutableFile(path string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(path, fxExecutableFileName))
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.WriteString(fxExecutableFileTemplate); err != nil {
		return err
	}

	if err := file.Sync(); err != nil {
		return err
	}

	return nil
}
