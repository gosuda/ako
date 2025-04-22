package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	fxFileName          = "init_%s.go"
	fxDependencyPackage = "go.uber.org/fx"
)

func getFxDependency() error {
	if err := getGoModuleAsTool(fxDependencyPackage); err != nil {
		return err
	}

	if err := tidyGoModule(); err != nil {
		return err
	}

	return nil
}

const fxStructFileTemplate = `package {{.package_name}}

import (
	"context"

	"go.uber.org/fx"
)

// {{.client_name}}Register is the fx.Provide function for the {{.client_name}} client.
// It registers the client as a dependency in the fx application.
// You can append interfaces into the fx.As() function to register multiple interfaces.
var {{.client_name}}Register = fx.Provide(New{{.client_name}}, fx.As())

type {{.client_name}}Param struct {
	fx.In
}

type {{.client_name}} struct {
}

func New{{.client_name}}(ctx context.Context, lc fx.Lifecycle, param {{.client_name}}Param) *{{.client_name}} {
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

import (
	"context"

	"go.uber.org/fx"
)

type {{.client_name}} interface {
}

type {{.client_name}}Data struct {
}`

func createFxStructFile(path string, name string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	name = strings.ToUpper(name[:1]) + name[1:]

	packageName := filepath.Base(path)
	if err := writeTemplate2File(filepath.Join(path, fmt.Sprintf(fxFileName, name)), fxStructFileTemplate, map[string]any{
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
	if err := writeTemplate2File(filepath.Join(path, fmt.Sprintf(fxFileName, name)), fxInterfaceFileTemplate, map[string]any{
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
	"time"

	"go.uber.org/fx"
)

func main() {
	fx.New(fx.StartTimeout(15*time.Second),
		fx.StopTimeout(15*time.Second)).Run()
}`
)

func createFxExecutableFile(path string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(path, fxExecutableFileName))
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := fmt.Fprintf(file, fxExecutableFileTemplate); err != nil {
		return err
	}

	if err := file.Sync(); err != nil {
		return err
	}

	return nil
}
