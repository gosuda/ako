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
	pkgTemplateList["[Storage] vault"] = createFxVaultFile
}

const (
	vaultDependency     = `github.com/hashicorp/vault-client-go`
	vaultClientTemplate = `package {{.package_name}}

import (
	"context"
	"os"
	"time"

	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
	"go.uber.org/fx"
)

// Register is the fx.Provide function for the client.
// It registers the client as a dependency in the fx application.
// You can append interfaces into the fx.As() function to register multiple interfaces.
var Register = fx.Provide(fx.Annotate(New, fx.As()))

// ConfigRegister is the fx.Provide function for the config.
// Modify the config according to your needs.
func ConfigRegister() func() *Config {
	return func() *Config {
		timeoutValue := os.Getenv("VAULT_{{.client_name}}_TIMEOUT")
		if timeoutValue == "" {
			timeoutValue = "10s"
		}
		timeout, _ := time.ParseDuration(timeoutValue)

		return &Config{
			Address: os.Getenv("VAULT_{{.client_name}}_ADDRESS"),
			Token: os.Getenv("VAULT_{{.client_name}}_TOKEN"),
			RoleId: os.Getenv("VAULT_{{.client_name}}_ROLE_ID"),
			SecretId: os.Getenv("VAULT_{{.client_name}}_SECRET_ID"),
			Timeout: timeout,
		}
	}
}

type Param struct {
	fx.In
	Cfg *Config
}

type Config struct {
	Address  string
	Token    string
	RoleId   string
	SecretId string
	Timeout  time.Duration
}

type {{.client_name}} struct {
	client *vault.Client
}

func New(ctx context.Context, lc fx.Lifecycle, param Param) *{{.client_name}} {
	cli := &{{.client_name}}{}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			client, err := vault.New(vault.WithAddress(param.Cfg.Address), vault.WithRequestTimeout(param.Cfg.Timeout))
			if err != nil {
				return err
			}

			switch len(param.Cfg.Token) {
			case 0:
				resp, err := client.Auth.AppRoleLogin(ctx, schema.AppRoleLoginRequest{
					RoleId:   param.Cfg.RoleId,
					SecretId: param.Cfg.SecretId,
				})
				if err != nil {
					return err
				}
		
				if err := client.SetToken(resp.Auth.ClientToken); err != nil {
					return err
				}
			default:
				if err := client.SetToken(param.Cfg.Token); err != nil {
					return err
				}
			}

			cli.client = client

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})

	return cli
}

type serializable[T any] interface {
	Serialize() (map[string]any, error)
	Deserialize(data map[string]any) error
	Init() T
}

type mountOption struct {
	mountPath            string
	requestSpecificToken string
}

func newMountOption() *mountOption {
	return &mountOption{}
}

func (o *mountOption) withMountPath(mountPath string) *mountOption {
	o.mountPath = mountPath
	return o
}

func (o *mountOption) withRequestSpecificToken(token string) *mountOption {
	o.requestSpecificToken = token
	return o
}

type mountedSecret[T serializable[T]] struct {
	client *vault.Client
	option *mountOption
}

func mount[T serializable[T]](client *{{.client_name}}, option *mountOption) (*mountedSecret[T], error) {
	return &mountedSecret[T]{
		client: client.client,
		option: option,
	}, nil
}

func (m *mountedSecret[T]) write(ctx context.Context, path string, data T) error {
	d, err := data.Serialize()
	if err != nil {
		return err
	}

	opt := []vault.RequestOption{
		vault.WithMountPath(m.option.mountPath),
	}
	if m.option.requestSpecificToken != "" {
		opt = append(opt, vault.WithToken(m.option.requestSpecificToken))
	}

	if _, err := m.client.Secrets.KvV2Write(ctx, path, schema.KvV2WriteRequest{
		Data: d,
	}, opt...); err != nil {
		return err
	}

	return nil
}

func (m *mountedSecret[T]) read(ctx context.Context, path string) (T, error) {
	result := *new(T)

	opt := []vault.RequestOption{
		vault.WithMountPath(m.option.mountPath),
	}
	if m.option.requestSpecificToken != "" {
		opt = append(opt, vault.WithToken(m.option.requestSpecificToken))
	}

	resp, err := m.client.Secrets.KvV2Read(ctx, path, opt...)
	if err != nil {
		return result, err
	}

	result = result.Init()
	if err := result.Deserialize(resp.Data.Data); err != nil {
		return result, err
	}

	return result, nil
}

func (m *mountedSecret[T]) delete(ctx context.Context, path string) error {
	opt := []vault.RequestOption{
		vault.WithMountPath(m.option.mountPath),
	}
	if m.option.requestSpecificToken != "" {
		opt = append(opt, vault.WithToken(m.option.requestSpecificToken))
	}

	if _, err := m.client.Secrets.KvV2Delete(ctx, path, opt...); err != nil {
		return err
	}

	return nil
}
`
	vaultObjectTemplate = `package {{.package_name}}

import (
	"encoding/json"
	"fmt"
)

// Write wrapper for the vault client

// SimplePassword is a struct that holds a username and password
type SimplePassword struct {
	Username string
	Password string
	Number   int64
}

func (s *SimplePassword) Serialize() (map[string]any, error) {
	return map[string]any{
		"username": s.Username,
		"password": s.Password,
		"aged":     s.Number,
	}, nil
}

func (s *SimplePassword) Deserialize(data map[string]any) error {
	username, ok := data["username"].(string)
	if !ok {
		return fmt.Errorf("username is not a string")
	}

	password, ok := data["password"].(string)
	if !ok {
		return fmt.Errorf("password is not a string")
	}

	aged, ok := data["aged"].(json.Number)
	if !ok {
		return fmt.Errorf("aged is not a int64")
	}

	s.Username = username
	s.Password = password
	s.Number, _ = aged.Int64()

	return nil
}

func (s *SimplePassword) Init() *SimplePassword {
	if s == nil {
		return &SimplePassword{}
	}

	return s
}`
)

func createFxVaultFile(path string, name string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	name = strings.ToUpper(name[:1]) + name[1:]

	packageName := filepath.Base(path)
	if err := template.WriteTemplate2File(filepath.Join(path, fmt.Sprintf(fxFileName, name)), vaultClientTemplate, map[string]any{
		"package_name": packageName,
		"client_name":  name,
	}); err != nil {
		return err
	}

	if err := template.WriteTemplate2File(filepath.Join(path, "vault_objects.go"), vaultObjectTemplate, map[string]any{
		"package_name": packageName,
		"client_name":  name,
	}); err != nil {
		return err
	}

	if err := module.GetGoModule(vaultDependency); err != nil {
		return fmt.Errorf("getGoModule: %w", err)
	}

	return nil
}
