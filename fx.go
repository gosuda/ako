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

const fxFileTemplate = `package {{.package_name}}

import (
	"context"

	"go.uber.org/fx"
)

// %sRegister is the fx.Provide function for the %s client.
// It registers the client as a dependency in the fx application.
// You can append interfaces into the fx.As() function to register multiple interfaces.
var {{.client_name}}Register = fx.Provide(New%s, fx.As())

type {{.client_name}}Param struct {
	fx.In
}

type {{.client_name}} struct {
}

func New{{.client_name}}(ctx context.Context, lc fx.Lifecycle, param {{.client_name}}Param) *%s {
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

func createFxFile(path string, name string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	name = strings.ToUpper(name[:1]) + name[1:]

	packageName := filepath.Base(path)
	if err := writeTemplate2File(filepath.Join(path, fmt.Sprintf(fxFileName, name)), fxFileTemplate, map[string]any{
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
