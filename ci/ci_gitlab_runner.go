package ci

import "os"

func init() {
	ciTemplates["gitlab ci/cd"] = createGitlabCICDConfig
}

const (
	gitlabCICDFileName = ".gitlab-ci.yml"
	gitlabCICDTemplate = `image: golang:1

stages:
  - lint
  - test

lint_job:
  stage: lint
  script:
    - echo "Stsarting linting..."
    - go tool golangci-lint run
    - echo "Linting completed."

test_job:
  stage: test
  script:
    - echo "Starting tests..."
    - go test ./...
    - echo "Tests completed."
`
)

func createGitlabCICDConfig(name string) error {
	f, err := os.Create(gitlabCICDFileName)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(gitlabCICDTemplate); err != nil {
		return err
	}

	if err := f.Sync(); err != nil {
		return err
	}

	return nil
}
