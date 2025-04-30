package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

const (
	libraryAdapter = "adapter"
	libRepository  = "repository"
	libDomain      = "domain"
)

func makeLibraryPath(base string, name string) string {
	return filepath.Join("lib", base, name)
}

func selectLibraryBase() (string, error) {
	suggestions := []string{
		libraryAdapter + ": Defines interfaces abstracting communication with external systems (APIs, message queues, etc.).",
		libRepository + ": Defines interfaces for accessing the data persistence layer (DB, cache, etc.).",
		libDomain + ": Defines core business logic, rules, domain models, and domain service interfaces.",
	}
	var base string
	if err := survey.AskOne(&survey.Input{
		Message: "Select the library category:",
		Suggest: func(toComplete string) []string {
			result := make([]string, 0, len(suggestions))
			for _, candidate := range suggestions {
				if strings.Contains(strings.ToLower(candidate), strings.ToLower(toComplete)) {
					result = append(result, candidate)
				}
			}

			return result
		},
	}, &base, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}

	sp := strings.Split(base, ":")
	base = strings.TrimSpace(sp[0])

	return base, nil
}

func inputLibraryPackage() (string, error) {
	var packageName string
	if err := survey.AskOne(&survey.Input{
		Message: "Enter the library package path:",
	}, &packageName, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}

	packageName = strings.TrimSpace(packageName)
	if packageName == "" {
		return "", fmt.Errorf("invalid library packageName: %s", packageName)
	}

	return packageName, nil
}

func inputLibraryName() (string, error) {
	var name string
	if err := survey.AskOne(&survey.Input{
		Message: "Enter the library interface name:",
	}, &name, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("invalid library name: %s", name)
	}

	return name, nil
}

func createLibraryFile(path string, name string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	packageName := filepath.Base(path)

	fileName := filepath.Join(path, fmt.Sprintf(fxFileName, name))
	if err := writeTemplate2File(fileName, fxInterfaceFileTemplate, map[string]any{
		"package_name": packageName,
		"client_name":  name,
	}); err != nil {
		return err
	}

	return nil
}
