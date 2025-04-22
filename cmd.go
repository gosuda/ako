package main

import (
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

func inputCmdName() (string, error) {
	var cmdName string
	if err := survey.AskOne(&survey.Input{
		Message: "Enter the command name [cmd/<name>]:",
	}, &cmdName, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}

	cmdName = strings.TrimSpace(cmdName)

	return cmdName, nil
}
