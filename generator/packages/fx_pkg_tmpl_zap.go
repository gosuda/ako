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

// Register is the fx.Provide function for the client.
// It registers the client as a dependency in the fx application.
// You can append interfaces into the fx.As() function to register multiple interfaces.
var Register = fx.Provide(fx.Annotate(New, fx.As()))

type Param struct {
	fx.In
}

type {{.client_name}} struct {
	writer io.Writer
	logger *zap.Logger
}

func New(ctx context.Context, lc fx.Lifecycle, param Param) *{{.client_name}} {
	cli := &{{.client_name}}{}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger := zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), os.Stdout, zapcore.DebugLevel))

			cli.writer = os.Stdout
			cli.logger = logger

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
