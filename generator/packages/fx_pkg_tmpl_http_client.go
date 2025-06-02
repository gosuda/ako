package packages

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gosuda/ako/util/template"
)

func init() {
	pkgTemplateList["[Http] http client"] = createFxHttpClientFile
}

const (
	httpClientTemplate = `package {{.package_name}}

import (
	"context"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"go.uber.org/fx"
)

const Name = "{{.client_name}}"

var Module = fx.Module("{{.package_name}}",
	fx.Provide(ConfigRegister()),
	fx.Provide(fx.Annotate(New, fx.As(/* implemented interfaces */))),
)

// ConfigRegister is the fx.Provide function for the config.
// Modify the config according to your needs.
func ConfigRegister() func() *Config {
	return func() *Config {
		return &Config{}
	}
}

type Param struct {
	fx.In
}

type Config struct {}

type {{.client_name}} struct {
	client *http.Client
}

func New(ctx context.Context, lc fx.Lifecycle, param Param) *{{.client_name}} {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Initialize the client here if needed
			return nil
		},
		OnStop: func(ctx context.Context) error {
			// Clean up resources if needed
			return nil
		},
	})

	return &{{.client_name}}{}
}

func (c *{{.client_name}}) withCookie(url *url.URL, cookie ...*http.Cookie) *{{.client_name}} {
	c.client.Jar.SetCookies(url, cookie)
	return c
}

func (c *{{.client_name}}) resetCookie() *{{.client_name}} {
	c.client.Jar = new(cookiejar.Jar)
	return c
}

func (c *{{.client_name}}) get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	context.AfterFunc(ctx, func() {
		resp.Body.Close()
	})

	return resp, nil
}

func (c *{{.client_name}}) post(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	context.AfterFunc(ctx, func() {
		resp.Body.Close()
	})

	return resp, nil
}

func (c *{{.client_name}}) put(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	context.AfterFunc(ctx, func() {
		resp.Body.Close()
	})

	return resp, nil
}

func (c *{{.client_name}}) patch(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	context.AfterFunc(ctx, func() {
		resp.Body.Close()
	})

	return resp, nil
}

func (c *{{.client_name}}) delete(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	context.AfterFunc(ctx, func() {
		resp.Body.Close()
	})

	return resp, nil
}`
)

func createFxHttpClientFile(path string, name string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	name = strings.ToUpper(name[:1]) + name[1:]

	packageName := filepath.Base(path)
	if err := template.WriteTemplate2File(filepath.Join(path, fmt.Sprintf(fxFileName, name)), httpClientTemplate, map[string]any{
		"package_name": packageName,
		"client_name":  name,
	}); err != nil {
		return err
	}

	return nil
}
