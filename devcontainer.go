package main

import (
	"os"
	"path/filepath"
	"strings"
)

func generateGoImageFile(appName string) error {
	const template = `
# Run
# docker build -f cmd/{{.cmd_name}}/Dockerfile -t <org>/<group>/{{.app_name}} .
# in root of the project to build the image.
FROM docker.io/library/golang:latest

WORKDIR /app

COPY go.mod .
COPY go.sum .

ENV CGO_ENABLED=0 # Disable CGO, If you need CGO, set it to 1

RUN go mod download

COPY . .
RUN go build -o main ./cmd/{{.cmd_name}}/.

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=0 /app/main .

CMD ["./main"]
`

	path := filepath.Join("cmd", appName, "Dockerfile")
	if err := writeTemplate2File(path, template, map[string]any{
		"cmd_name": appName,
		"app_name": makeCmdDepthToName(strings.Split(appName, "/")...),
	}); err != nil {
		return err
	}

	return nil
}

func generateDevContainerFile(name string) error {
	if err := os.MkdirAll(".devcontainer", os.ModePerm); err != nil {
		return err
	}

	const goImageTemplate = `FROM docker.io/library/golang:alpine
RUN apk add --no-cache git
RUN go env -w GOPROXY=https://proxy.golang.org,direct
RUN go env -w GOSUMDB=sum.golang.org
RUN go env -w GOSUMDB=off
RUN go env -w GOPRIVATE=*.mycompany.com
RUN go env -w GONOSUMDB=*.mycompany.com

RUN go install github.com/gosuda/ako@latest

WORKDIR /workspace

CMD ["/bin/sh", "-c", "while true; do sleep 30; done;"]
`

	path := filepath.Join(".devcontainer", "Dockerfile")
	if err := writeTemplate2File(path, goImageTemplate, nil); err != nil {
		return err
	}

	const devcontainerJsonTemplate = `{
	"name": "{{.name}}",
	"dockerComposeFile": [
		"docker-compose.yml"
	],
	"service": "{{.name}}",
	"workspaceFolder": "/workspace",
	"shutdownAction": "stopCompose",
	"customizations": {
		"vscode": {
			"extensions": [
				"ms-azuretools.vscode-docker",
				"golang.go"
			]
		}
	}
}`

	path = filepath.Join(".devcontainer", "devcontainer.json")
	if err := writeTemplate2File(path, devcontainerJsonTemplate, map[string]any{
		"name": name,
	}); err != nil {
		return err
	}

	const dockerComposeTemplate = `version: '3'
services:
  {{.name}}:
    build:
      context: .
      dockerfile: Dockerfile
    volumes:
      - ../.:/workspace:cached
`
	path = filepath.Join(".devcontainer", "docker-compose.yml")
	if err := writeTemplate2File(path, dockerComposeTemplate, map[string]any{
		"name": name,
	}); err != nil {
		return err
	}

	return nil
}
