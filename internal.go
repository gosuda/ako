package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

const (
	internalPackageController = "controller"
	internalPackageService    = "service"
)

func selectInternalPackageBase() (string, error) {
	candidates := []string{
		internalPackageController + ": Handles incoming requests (e.g., HTTP, gRPC), validates input, invokes appropriate services, and formats responses.",
		internalPackageService + ": Implements core business logic and use cases by orchestrating operations using interfaces defined in lib/",
	}
	var base string
	if err := survey.AskOne(&survey.Select{
		Message: "Select the internal package type [internal/<base>/<package>]:",
		Options: candidates,
	}, &base, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}

	sp := strings.Split(base, ":")
	base = strings.TrimSpace(sp[0])

	return base, nil
}

func inputInternalPackageName() (string, error) {
	var packageName string
	if err := survey.AskOne(&survey.Input{
		Message: "Enter the internal package name [internal/<base>/<package>]:",
	}, &packageName, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}

	packageName = strings.TrimSpace(packageName)
	if packageName == "" {
		return "", fmt.Errorf("invalid internal package name: %s", packageName)
	}

	return packageName, nil
}

const internalTemplate = `package {{.package_name}}

import (
	"context"

	"go.uber.org/fx"
)

// Register is the fx.Provide function for the client.
// It registers the client as a dependency in the fx application.
var Register = fx.Provide(New)

type Param struct {
	fx.In
}

type {{.client_name}} struct {
}

func New(ctx context.Context, lc fx.Lifecycle, param Param) *{{.client_name}} {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Orchestrate the initialization of the client here if needed
			return nil
		},
		OnStop: func(ctx context.Context) error {
			// Clean up resources if needed
			return nil
		},
	})
	return &{{.client_name}}{}
}`

func createInternalPackage(path, packageName string) error {
	dir := filepath.Join("internal", path, packageName)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	structName := strings.ToUpper(packageName[:1]) + packageName[1:]
	fileName := filepath.Join(dir, fmt.Sprintf(fxFileName, structName))
	if err := writeTemplate2File(fileName, internalTemplate, map[string]any{
		"package_name": packageName,
		"client_name":  structName,
	}); err != nil {
		return err
	}

	return nil
}
