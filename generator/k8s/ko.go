package k8s

import "github.com/gosuda/ako/util/module"

const (
	koToolDependency = `github.com/google/ko`
	koToolName       = "ko"
)

func InstallKo() error {
	if err := module.GetGoModuleAsTool(koToolDependency); err != nil {
		return err
	}

	return nil
}

type KoBuildOption struct {
	Path string
}

func BuildKo(option KoBuildOption) error {
	if err := module.RunGoModuleToolWithEnv(koToolName, map[string]string{}, "build", option.Path, "--platform=all"); err != nil {
		return err
	}

	return nil
}
