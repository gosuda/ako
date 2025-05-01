package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

func makePackagePath(base string, name string) string {
	return filepath.Join("pkg", base, name)
}

func inputPackageBase() (string, error) {
	suggestions := []string{
		"persistence: Contains implementations related to storing and retrieving data from persistent storage (e.g., databases).",
		"cache: Contains implementations for interacting with caching systems (e.g., Redis, Memcached).",
		"client: Contains implementations for clients that interact with external services or APIs.",
		"transport: Contains implementations related to network transport layers (e.g., HTTP server setup, gRPC specifics).",
		"notification: Contains implementations for sending notifications via different channels (e.g., SMTP, SMS).",
		"payment: Contains implementations for interacting with payment gateways (e.g., Stripe, PayPal).",
		"config: Contains implementations for loading and managing application configuration (e.g., using Viper).",
		"encryption: Contains implementations for cryptographic operations like hashing or encryption (e.g., using bcrypt).",
		"logger: Contains implementations for logging interfaces (e.g., using Zerolog, Zap).",
		"template: Contains implementations for template rendering engines (e.g., HTML templates).",
		"validation: Contains implementations for data validation logic or integrating with validation libraries.",
		"analytics: Contains implementations for analytics tracking (e.g., Google Analytics, Mixpanel).",
		"monitoring: Contains implementations for monitoring and observability (e.g., Prometheus, Grafana).",
		"transformation: Contains implementations for data transformation or ETL processes. (e.g., converting data formats, duckdb).",
	}
	var base string
	if err := survey.AskOne(&survey.Input{
		Message: "Enter the package base [pkg/<base>/<package>]:",
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

func inputPackageName() (string, error) {
	var packageName string
	if err := survey.AskOne(&survey.Input{
		Message: "Enter the package name:",
	}, &packageName, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}

	packageName = strings.TrimSpace(packageName)
	if packageName == "" {
		return "", fmt.Errorf("invalid packageName: %s", packageName)
	}

	return packageName, nil
}
