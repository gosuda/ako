package main

const (
	golangcilintToolDependency = `github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2`
	golangcilintToolName       = `golangci-lint`
	golangcilintFileName       = ".golangci.yaml"
)

func installGolangcilint() error {
	if err := getGoModuleAsTool(golangcilintToolDependency); err != nil {
		return err
	}

	return nil
}

func runGolangcilint() error {
	if err := runGoModuleTool(golangcilintToolName, "run"); err != nil {
		return err
	}

	return nil
}

func createGolangcilintConfig() error {
	if err := writeTemplate2File(golangcilintFileName, golangcilintConfig, map[string]any{}); err != nil {
		return err
	}

	return nil
}
