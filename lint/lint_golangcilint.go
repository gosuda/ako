package lint

import (
	"github.com/gosuda/ako/module"
	"github.com/gosuda/ako/template"
)

const (
	golangcilintToolDependency = `github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2`
	golangcilintToolName       = `golangci-lint`
	golangcilintFileName       = ".golangci.yaml"
)

func InstallGolangcilint() error {
	if err := module.GetGoModuleAsTool(golangcilintToolDependency); err != nil {
		return err
	}

	return nil
}

func RunGolangcilint() error {
	if err := module.RunGoModuleTool(golangcilintToolName, "run"); err != nil {
		return err
	}

	return nil
}

func CreateGolangcilintConfig() error {
	if err := template.WriteTemplate2File(golangcilintFileName, golangcilintConfig, map[string]any{}); err != nil {
		return err
	}

	return nil
}
