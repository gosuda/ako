package main

import (
	"fmt"
	"os"
	"path/filepath"
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
      value: "{{.module_name}}/lib/gen"
  disable:
    - module: buf.build/googleapis/googleapis
      file_option: go_package_prefix
plugins:
  - remote: buf.build/grpc/go:v1.4.0
    out: lib/adapter/gen
    opt:
      - paths=source_relative
  - remote: buf.build/protocolbuffers/go
    out: lib/adapter/gen
    opt:
      - paths=source_relative`

	bufCmdPackage = "github.com/bufbuild/buf/cmd/buf"

	protobufDependencyPackage = "google.golang.org/protobuf"
	grpcDependencyPackage     = "google.golang.org/grpc"
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

	if err := writeTemplate2File(bufGenFileName, bufGenFileTemplate, map[string]any{
		"module_name": moduleName,
	}); err != nil {
		return err
	}

	if err := getGoModuleAsTool(bufCmdPackage); err != nil {
		return err
	}

	if err := getGoModule(protobufDependencyPackage); err != nil {
		return err
	}

	if err := getGoModule(grpcDependencyPackage); err != nil {
		return err
	}

	return nil
}

const protobufExample = `syntax = "proto3";

option go_package = "./person";

message Person {
  string name = 1;
  int32 age = 2;
  string email = 3;
}

message Request {
  string name = 1;
}

message Response {
  string message = 1;
}

service Greeter {
  rpc SayHello (Request) returns (Response) {}
}`

// createProtobufExample creates a protobuf example file in the proto directory.
func createProtobufExample() error {
	protoDir := filepath.Join("proto", "person")
	if err := os.MkdirAll(protoDir, 0755); err != nil {
		return err
	}

	file, err := os.Create(fmt.Sprintf("%s/person.proto", protoDir))
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.WriteString(protobufExample); err != nil {
		return err
	}

	if err := file.Sync(); err != nil {
		return err
	}

	return nil
}
