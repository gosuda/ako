package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	clickhouseDependencyPackage = "github.com/ClickHouse/clickhouse-go/v2"
	clickhouseTemplate1         = `package {{.package_name}}

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"go.uber.org/fx"
)

// ClientRegister is the fx.Provide function for the client.
// It registers the client as a dependency in the fx application.
// You can append interfaces into the fx.As() function to register multiple interfaces.
var ClientRegister = fx.Provide(NewClient, fx.As())

type Config struct {
	addr               []string
	database           *string
	username           *string
	password           *string
	insecureSkipVerify *bool

	clientInfo clickhouse.ClientInfo
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) WithAddr(addr ...string) *Config {
	c.addr = addr
	return c
}

func (c *Config) WithDatabase(database string) *Config {
	c.database = &database
	return c
}

func (c *Config) WithUsername(username string) *Config {
	c.username = &username
	return c
}

func (c *Config) WithPassword(password string) *Config {
	c.password = &password
	return c
}

func (c *Config) WithInsecureSkipVerify(insecureSkipVerify bool) *Config {
	c.insecureSkipVerify = &insecureSkipVerify
	return c
}

func (c *Config) AddClientInfo(name string, version string) *Config {
	c.clientInfo.Products = append(c.clientInfo.Products, struct {
		Name    string
		Version string
	}{Name: name, Version: version})
	return c
}

type Client struct {
	conn driver.Conn
}

func NewClient(lc fx.Lifecycle, cfg Config) (*Client, error) {
	opt := clickhouse.Options{}
	if cfg.addr != nil {
		opt.Addr = cfg.addr
	}
	if cfg.database != nil {
		opt.Auth.Database = *cfg.database
	}
	if cfg.username != nil {
		opt.Auth.Username = *cfg.username
	}
	if cfg.password != nil {
		opt.Auth.Password = *cfg.password
	}
	if cfg.insecureSkipVerify != nil {
		opt.TLS.InsecureSkipVerify = *cfg.insecureSkipVerify
	}
	if len(cfg.clientInfo.Products) > 0 {
		opt.ClientInfo = cfg.clientInfo
	}

	conn, err := clickhouse.Open(&opt)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Initialize the client here if needed
			return nil
		},
		OnStop: func(ctx context.Context) error {
			// Clean up resources if needed
			if err := conn.Close(); err != nil {
				return fmt.Errorf("failed to close connection: %w", err)
			}
			return nil
		},
	})

	return &Client{
		conn: conn,
	}, nil
}`
	clickhouseTemplate2 = `package {{.package_name}}

import (
	"context"
	"fmt"
	"time"
)

func (c *Client) FindExample(ctx context.Context, name string) ([]int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := c.conn.Query(ctx, "SELECT age FROM person WHERE name = ?", name)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	ages := make([]int64, 0, 4)
	for rows.Next() {
		var age int64
		if err := rows.Scan(&age); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		ages = append(ages, age)
	}

	return ages, nil
}`
)

func createClickhouseFile(path string, name string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	name = strings.ToUpper(name[:1]) + name[1:]

	packageName := filepath.Base(path)

	if err := writeTemplate2File(filepath.Join(path, fmt.Sprintf(fxFileName, name)), clickhouseTemplate1, map[string]any{
		"package_name": packageName,
	}); err != nil {
		return err
	}

	if err := writeTemplate2File(filepath.Join(path, "query.go"), clickhouseTemplate2, map[string]any{
		"package_name": packageName,
	}); err != nil {
		return err
	}

	if err := getGoModule(clickhouseDependencyPackage); err != nil {
		return err
	}

	return nil
}
