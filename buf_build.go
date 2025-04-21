package main

import (
	"fmt"
	"os"
)

const (
	bufYamlFileName = "buf.yaml"
	bufYamlTemplate = `version: v2
modules:
  - path: ./proto
lint:
  use:
    - DEFAULT
breaking:
  use:
    - FILE`

	bufGenFileName     = "buf.gen.yaml"
	bufGenFileTemplate = `version: v2
managed:
  enabled: true
  override:
    - file_option: go_package_prefix
      # <module_name>   : name in go.mod
      # <relative_path> : where generated code should be output
      value: "%s/lib/gen"
  disable:
    - module: buf.build/googleapis/googleapis
      file_option: go_package_prefix
plugins:
  - remote: buf.build/grpc/go:v1.4.0
    out: gen
    opt:
      - paths=source_relative
  - remote: buf.build/protocolbuffers/go
    out: gen
    opt:
      - paths=source_relative`

	bufCmdPackage = "github.com/bufbuild/buf/cmd/buf"
)

// createBufTemplate creates buf.yaml and buf.gen.yaml files in the current directory.
func createBufTemplate() error {
	moduleName, err := getGoModuleName()
	if err != nil {
		return err
	}

	if err := func() error {
		file, err := os.Create(bufYamlFileName)
		if err != nil {
			return err
		}
		defer file.Close()

		if _, err := file.WriteString(bufYamlTemplate); err != nil {
			return err
		}

		if err := file.Sync(); err != nil {
			return err
		}

		return nil
	}(); err != nil {
		return err
	}

	if err := func() error {
		file, err := os.Create(bufGenFileName)
		if err != nil {
			return err
		}
		defer file.Close()

		if _, err := fmt.Fprintf(file, bufGenFileTemplate, moduleName); err != nil {
			return err
		}

		if err := file.Sync(); err != nil {
			return err
		}

		return nil
	}(); err != nil {
		return err
	}

	if err := getGoModuleAsTool(bufCmdPackage); err != nil {
		return err
	}

	return nil
}
