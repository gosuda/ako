package ai

import "github.com/AlecAivazis/survey/v2"

func Confirm(message string) (bool, error) {
	confirm := false
	if err := survey.AskOne(&survey.Confirm{
		Message: message,
	}, &confirm, survey.WithValidator(survey.Required)); err != nil {
		return false, err
	}

	return confirm, nil
}
