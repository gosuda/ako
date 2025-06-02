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
	pkgTemplateList["[Logger/Zap] zap"] = createFxZapFile
}

const (
	zapDependency     = "go.uber.org/zap"
	zapWriterTemplate = `package {{.package_name}}

import (
	"context"
	"fmt"
	"io"
	"os"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const Name = "{{.client_name}}"

var Module = fx.Module("{{.package_name}}",x
	fx.Provide(fx.Annotate(New, fx.As(/* implemented interfaces */))),
)

type Param struct {
	fx.In
}

type {{.client_name}} struct {
	writer io.Writer
	logger *zap.Logger
}

func New(ctx context.Context, lc fx.Lifecycle, param Param) *{{.client_name}} {
	cli := &{{.client_name}}{}

	logger := zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), os.Stdout, zapcore.DebugLevel))

	cli.writer = os.Stdout
	cli.logger = logger

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if err := cli.logger.Sync(); err != nil {
				return fmt.Errorf("zap.Sync: %w", err)
			}
			return nil
		},
	})

	return cli
}
`
)

func createFxZapFile(path string, name string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	name = strings.ToUpper(name[:1]) + name[1:]

	packageName := filepath.Base(path)
	if err := template.WriteTemplate2File(filepath.Join(path, fmt.Sprintf(fxFileName, name)), zapWriterTemplate, map[string]any{
		"package_name": packageName,
		"client_name":  name,
	}); err != nil {
		return err
	}

	if err := module.GetGoModule(zapDependency); err != nil {
		return fmt.Errorf("getGoModule: %w", err)
	}

	return nil
}
