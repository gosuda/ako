package main

import "os"

func init() {
	ciTemplates["gitlab actions"] = createGitlabCICDConfig
}

const (
	githubActionFileName = ".github/workflows/%s.yml"
	githubActionTemplate = `name: "Code Analysis"

on:
  pull_request:
  push:
    paths:
      - '**.go'
      - 'go.mod'
      - 'go.sum'

permissions:
  contents: read

jobs:
  basic:
    name: "Run Basic Code Analysis"
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: 'stable'
      - run: "go test ./..."
  golangci-lint:
    name: "Run GolangCI-Lint Code Analysis"
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: 'stable'
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v7
        with:
          version: latest
`
)

func createGithubActionConfig(name string) error {
	f, err := os.Create(githubActionFileName)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(githubActionTemplate); err != nil {
		return err
	}

	if err := f.Sync(); err != nil {
		return err
	}

	return nil
}
