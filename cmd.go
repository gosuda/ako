package main

import (
	"os"
	"path/filepath"
	"slices"
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

func getCmdList(prefix string) ([]string, error) {
	if prefix == "" {
		prefix = RootPackageCmd
	}

	dir, err := os.ReadDir(prefix)
	if err != nil {
		return nil, err
	}

	cmdList := map[string]struct{}{}

	for _, entry := range dir {
		if entry.IsDir() {
			entries, err := getCmdList(filepath.Join(prefix, entry.Name()))
			if err != nil {
				return nil, err
			}

			for _, entry := range entries {
				cmdList[filepath.ToSlash(entry)] = struct{}{}
			}
			continue
		}

		if strings.HasSuffix(entry.Name(), ".go") {
			cmdList[filepath.ToSlash(prefix)] = struct{}{}
		}
	}

	delete(cmdList, RootPackageCmd)

	var cmdListSlice []string
	for cmd := range cmdList {
		cmdListSlice = append(cmdListSlice, cmd)
	}

	slices.Sort(cmdListSlice)

	return cmdListSlice, nil
}

func selectCmdName() (string, error) {
	cmdList, err := getCmdList(RootPackageCmd)
	if err != nil {
		return "", err
	}

	for i := range cmdList {
		cmdList[i] = strings.TrimPrefix(cmdList[i], RootPackageCmd+"/")
	}

	var cmdName string
	if err := survey.AskOne(&survey.Select{
		Message: "Select the command name:",
		Options: cmdList,
	}, &cmdName, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}

	return cmdName, nil
}
