package packages

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gosuda/ako/util/template"
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

var Module = fx.Module("{{.package_name}}",
	fx.Provide(fx.Annotate(New, fx.As(/* implemented interfaces */))),
)

type Param struct {
	fx.In
}

type {{.client_name}} struct {
	writer io.Writer
	logger *slog.Logger
}

func New(ctx context.Context, lc fx.Lifecycle, param Param) *{{.client_name}} {
	cli := &{{.client_name}}{}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cli.writer = os.Stdout
	cli.logger = logger

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
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
	if err := template.WriteTemplate2File(filepath.Join(path, fmt.Sprintf(fxFileName, name)), slogWriterTemplate, map[string]any{
		"package_name": packageName,
		"client_name":  name,
	}); err != nil {
		return err
	}

	return nil
}
