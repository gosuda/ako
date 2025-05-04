package main

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
    - echo "린팅을 시작합니다..."
    - go tool golangci-lint run
    - echo "린팅 완료."

test_job:
  stage: test
  script:
    - echo "테스트를 시작합니다..."
    - go test ./...
    - echo "테스트 완료."
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
