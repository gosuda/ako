package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	pkgTemplateList["[Logger/Slog] slog"] = createFxSlogFile
}

const (
	slogWriterTemplate = `package {{.package_name}}

import (
	"context"
	"io"
	"log/slog"
	"os"

	"go.uber.org/fx"
)

// Register is the fx.Provide function for the client.
// It registers the client as a dependency in the fx application.
// You can append interfaces into the fx.As() function to register multiple interfaces.
var Register = fx.Provide(fx.Annotate(New, fx.As()))

type Param struct {
	fx.In
}

type {{.client_name}} struct {
	writer io.Writer
	logger *slog.Logger
}

func New(ctx context.Context, lc fx.Lifecycle, param Param) *{{.client_name}} {
	cli := &{{.client_name}}{}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

			cli.writer = os.Stdout
			cli.logger = logger

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})

	return cli
}
`
)

func createFxSlogFile(path string, name string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	name = strings.ToUpper(name[:1]) + name[1:]

	packageName := filepath.Base(path)
	if err := writeTemplate2File(filepath.Join(path, fmt.Sprintf(fxFileName, name)), slogWriterTemplate, map[string]any{
		"package_name": packageName,
		"client_name":  name,
	}); err != nil {
		return err
	}

	return nil
}
