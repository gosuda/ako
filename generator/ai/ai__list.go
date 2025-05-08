package ai

import (
	"fmt"
	"slices"

	"github.com/AlecAivazis/survey/v2"
)

var (
	aiTemplateList = map[string]func() error{
		"github copilot": CreateVsCodeCopilotSettings,
		"continue":       CreateVsCodeContinueSettings,
	}
)

func SelectAiTemplate() (string, error) {
	candidates := make([]string, 0, len(aiTemplateList))
	for k := range aiTemplateList {
		candidates = append(candidates, k)
	}
	slices.Sort(candidates)

	var selected string
	if err := survey.AskOne(&survey.Select{
		Message: "Select the prefer AI solution to use:",
		Options: candidates,
	}, &selected, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}

	return selected, nil
}

func CreateAiTemplate(name string) error {
	if fn, ok := aiTemplateList[name]; ok {
		return fn()
	}

	return fmt.Errorf("template %s not found", name)
}
