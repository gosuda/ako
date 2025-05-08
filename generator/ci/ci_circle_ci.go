package ci

import "os"

func init() {
	ciTemplates["circleci"] = createCircleCiConfig
}

const (
	circleCiFileName = ".circleci/config.yml"
	circleCiTemplate = `version: 2.1

jobs:
  lint:
    docker:
      - image: golang:1
    steps:
      - checkout
      - run:
          name: Run golangci-lint
          command: go tool golangci-lint run
  test:
    docker:
      - image: golang:1
    steps:
      - checkout
      - run:
          name: Run Go Unit Tests
          command: go test ./...

workflows:
  lint_and_test:
    jobs:
      - lint
      - test
`
)

func createCircleCiConfig(name string) error {
	f, err := os.Create(circleCiFileName)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(circleCiTemplate); err != nil {
		return err
	}

	if err := f.Sync(); err != nil {
		return err
	}

	return nil
}
