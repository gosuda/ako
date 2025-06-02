package packages

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/AlecAivazis/survey/v2"

	"github.com/gosuda/ako/util/template"
)

var internalControllerTemplateList = map[string]func(string, string) error{}

func getInternalControllerTemplateKeyList() []string {
	keys := make([]string, 0, len(internalControllerTemplateList))
	for k := range internalControllerTemplateList {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

const (
	internalPackageController = "controller"
	internalPackageService    = "service"
)

func SelectInternalPackageBase() (string, error) {
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

func InputInternalPackageName() (string, error) {
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

const Name = "{{.client_name}}"

var Module = fx.Module("{{.package_name}}",
	fx.Provide(New),
)

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

func CreateInternalPackage(path, packageName string) error {
	dir := filepath.Join("internal", path, packageName)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	switch {
	case strings.HasPrefix(path, internalPackageController):
		creator, err := selectInternalControllerPackage()
		if err != nil {
			return err
		}

		if err := creator(dir, packageName); err != nil {
			return err
		}
	default:
		structName := strings.ToUpper(packageName[:1]) + packageName[1:]
		fileName := filepath.Join(dir, fmt.Sprintf(fxFileName, structName))
		if err := template.WriteTemplate2File(fileName, internalTemplate, map[string]any{
			"package_name": packageName,
			"client_name":  structName,
		}); err != nil {
			return err
		}
	}

	return nil
}

func selectInternalControllerPackage() (func(string, string) error, error) {
	var packageName string
	if err := survey.AskOne(&survey.Select{
		Message: "Select the internal package type [internal/<base>/<package>]:",
		Options: getInternalControllerTemplateKeyList(),
	}, &packageName, survey.WithValidator(survey.Required)); err != nil {
		return nil, err
	}

	if fn, ok := internalControllerTemplateList[packageName]; ok {
		return fn, nil
	}

	return nil, fmt.Errorf("invalid internal package name: %s", packageName)
}
