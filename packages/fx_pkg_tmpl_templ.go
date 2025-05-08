package packages

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"

	"github.com/gosuda/ako/module"
	"github.com/gosuda/ako/template"
)

func init() {
	pkgTemplateList["[Template/Templ] templ"] = createFxTemplFile
}

const (
	templToolDependency = `github.com/a-h/templ/cmd/templ`
	templDependency     = `github.com/a-h/templ`
	templObjectTemplate = `package {{.package_name}}

import (
	"context"

	"go.uber.org/fx"
)

//go:generate go tool templ generate

// Register is the fx.Provide function for the client.
// It registers the client as a dependency in the fx application.
// You can append interfaces into the fx.As() function to register multiple interfaces.
var Register = fx.Provide(fx.Annotate(New, fx.As()))

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
}
`
	templGeneratorFilename = "%s.templ"
	templGeneratorTemplate = `package {{.package_name}}

import "fmt"

templ (c *{{.client_name}}) {{.method_name}}(name string) {
	<div>{fmt.Sprintf("Hello, %s!", name)}</div>
}
`
)

func inputTemplateName() (string, error) {
	var name string
	if err := survey.AskOne(&survey.Input{
		Message: "Enter the name of the template",
	}, &name, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}

	name = strings.TrimSpace(name)
	if len(name) == 0 {
		return "", fmt.Errorf("name cannot be empty")
	}

	name = strings.ToUpper(name[:1]) + name[1:]

	return name, nil
}

func createFxTemplFile(path string, name string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	name = strings.ToUpper(name[:1]) + name[1:]

	packageName := filepath.Base(path)
	objectFilePath := filepath.Join(path, fmt.Sprintf(fxFileName, name))
	if _, err := os.Stat(objectFilePath); os.IsNotExist(err) {
		if err := template.WriteTemplate2File(objectFilePath, templObjectTemplate, map[string]any{
			"package_name": packageName,
			"client_name":  name,
		}); err != nil {
			return err
		}
	}

	templateName, err := inputTemplateName()
	if err != nil {
		return err
	}

	generatorFilePath := filepath.Join(path, fmt.Sprintf(templGeneratorFilename, templateName))
	if _, err := os.Stat(generatorFilePath); os.IsNotExist(err) {
		if err := template.WriteTemplate2File(generatorFilePath, templGeneratorTemplate, map[string]any{
			"package_name": packageName,
			"client_name":  name,
			"method_name":  templateName,
		}); err != nil {
			return err
		}
	}

	if err := module.GetGoModule(templDependency); err != nil {
		return fmt.Errorf("getGoModule: %w", err)
	}

	if err := module.GetGoModule(templToolDependency); err != nil {
		return fmt.Errorf("getGoModule: %w", err)
	}

	return nil
}
