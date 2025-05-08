package packages

import (
	"fmt"
	"slices"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

var pkgTemplateList = map[string]func(string, string) error{
	"[Plain] empty structure": createFxStructFile,
}

func getPkgTemplateKeyList() []string {
	keys := make([]string, 0, len(pkgTemplateList))
	for k := range pkgTemplateList {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

func GetPkgTemplateWriter(key string) (func(string, string) error, error) {
	writer, ok := pkgTemplateList[key]
	if !ok {
		return nil, fmt.Errorf("invalid fx package template key: %s", key)
	}
	return writer, nil
}

func SelectFxPkgTemplateKey() (string, error) {
	keys := getPkgTemplateKeyList()
	var key string
	if err := survey.AskOne(&survey.Select{
		Message: "Select the fx package template key:",
		Options: keys,
	}, &key, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}
	key = strings.TrimSpace(key)
	return key, nil
}
