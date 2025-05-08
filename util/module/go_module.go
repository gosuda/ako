package module

import (
	"os"
	"os/exec"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

func InputGoModuleName() (string, error) {
	var moduleName string
	if err := survey.AskOne(&survey.Input{
		Message: "Enter the Go module name [github.com/username/repo]:",
	}, &moduleName, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}

	moduleName = strings.TrimSpace(moduleName)

	return moduleName, nil
}

func InitGoModule(moduleName string) error {
	cmd := exec.Command("go", "mod", "init", moduleName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func GetGoModule(item string) error {
	cmd := exec.Command("go", "get", "-u", item)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func GetGoModuleAsTool(item string) error {
	cmd := exec.Command("go", "get", "-tool", item)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func TidyGoModule() error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func GetGoModuleName() (string, error) {
	cmd := exec.Command("go", "list", "-m")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func RunGoModuleTool(item string, command ...string) error {
	args := make([]string, 0, len(command)+2)
	args = append(args, "tool")
	args = append(args, item)
	args = append(args, command...)
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
