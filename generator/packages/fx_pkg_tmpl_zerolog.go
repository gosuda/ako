package packages

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gosuda/ako/util/module"
	"github.com/gosuda/ako/util/template"
)

func init() {
	pkgTemplateList["[Logger/Zerolog] zerolog"] = createFxZerologFile
}

const (
	zerologDependency     = "github.com/rs/zerolog"
	zerologWriterTemplate = `package {{.package_name}}

import (
	"context"
	"io"
	"os"

	"github.com/rs/zerolog"
	"go.uber.org/fx"
)

const Name = "{{.client_name}}"

var Module = fx.Module("{{.package_name}}",
	fx.Provide(fx.Annotate(New, fx.As(/* implemented interfaces */))),
)

type Param struct {
	fx.In
}

type {{.client_name}} struct {
	writer io.Writer
	logger zerolog.Logger
}

func New(ctx context.Context, lc fx.Lifecycle, param Param) *{{.client_name}} {
	cli := &{{.client_name}}{}

	logger := zerolog.New(os.Stdout).
		Level(zerolog.DebugLevel).
		With().Caller().Timestamp().Logger()

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

func createFxZerologFile(path string, name string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	name = strings.ToUpper(name[:1]) + name[1:]

	packageName := filepath.Base(path)
	if err := template.WriteTemplate2File(filepath.Join(path, fmt.Sprintf(fxFileName, name)), zerologWriterTemplate, map[string]any{
		"package_name": packageName,
		"client_name":  name,
	}); err != nil {
		return err
	}

	if err := module.GetGoModule(zerologDependency); err != nil {
		return fmt.Errorf("getGoModule: %w", err)
	}

	return nil
}
