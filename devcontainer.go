package main

import (
	"path/filepath"
)

func generateGoImageFile(appName string) error {
	const template = `FROM golang:latest

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY ../../ .
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
	}); err != nil {
		return err
	}

	return nil
}
