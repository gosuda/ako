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

const fxFileTemplate = `package %s

import (
	"context"

	"go.uber.org/fx"
)

// %sRegister is the fx.Provide function for the %s client.
// It registers the client as a dependency in the fx application.
// You can append interfaces into the fx.As() function to register multiple interfaces.
var %sRegister = fx.Provide(New%s, fx.As())

type %s struct {
}

func New%s(ctx context.Context, lc fx.Lifecycle) *%s {
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
	return &%s{}
}`

func createFxFile(path string, name string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	name = strings.ToUpper(name[:1]) + name[1:]

	packageName := filepath.Base(path)
	file, err := os.Create(filepath.Join(path, fmt.Sprintf(fxFileName, name)))
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := fmt.Fprintf(file, fxFileTemplate, packageName, name, name, name, name, name, name, name, name); err != nil {
		return err
	}

	return nil
}
