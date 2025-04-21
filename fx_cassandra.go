package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	cassandraDependencyPackage1 = "github.com/gocql/gocql"
	cassandraDependencyPackage2 = "github.com/scylladb/gocqlx/v2"
	cassandraTemplate1          = `package %s

import (
	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v2"

	"go.uber.org/fx"
)

// %sRegister is the fx.Provide function for the %s client.
// It registers the client as a dependency in the fx application.
// You can append interfaces into the fx.As() function to register multiple interfaces.
var %sRegister = fx.Provide(New%s, fx.As())

type Config gocql.ClusterConfig

type %s struct {
	clusterConfig gocql.ClusterConfig
}

func New%s(lc fx.Lifecycle, cfg Config) *%s {
	client := &%s{
		clusterConfig: gocql.ClusterConfig(cfg),
	}
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

	return client
}

func (c *%s) SetKeyspace(keyspace string) {
	c.clusterConfig.Keyspace = keyspace
}

func (c *%s) SetConfig(f func(*gocql.ClusterConfig)) {
	f(&c.clusterConfig)
}

func (c *%s) Connect() (*gocql.Session, error) {
	return c.clusterConfig.CreateSession()
}

func (c *%s) ConnectX() (gocqlx.Session, error) {
	session, err := gocqlx.WrapSession(c.clusterConfig.CreateSession())
	if err != nil {
		return session, err
	}
	return session, nil
}`

	cassandraTemplate2 = `package %s

import "github.com/scylladb/gocqlx/v2/table"

var sampleTableMetadata = table.Metadata{
	Name: "sample_table",
	Columns: []string{
		"column1",
		"column2",
		"column3",
	},
	PartKey: []string{"column1"},
	SortKey: []string{"column2"},
}

var sampleTable = table.New(sampleTableMetadata)

var sampleTableDdl = ` + "`" + `CREATE TABLE IF NOT EXISTS sample_table (
column1 text,
column2 text,
column3 text,
PRIMARY KEY (column1, column2)
) WITH CLUSTERING ORDER BY (column2 ASC);
` + "`" + `

type SampleTable struct {
	Column1 string
	Column2 string
	Column3 string
}

var sampleTableSelectStmt, sampleTableSelectNames = sampleTable.Select()

func (c *%s) Select(column1, column2 string) (*SampleTable, error) {
	sess, err := c.ConnectX()
	if err != nil {
		return nil, err
	}

	var result SampleTable
	if err := sess.Query(sampleTableSelectStmt, sampleTableSelectNames).BindMap(map[string]any{
		"column1": column1,
		"column2": column2,
	}).SelectRelease(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

var sampleTableInsertStmt, sampleTableInsertNames = sampleTable.Insert()

func (c *%s) Insert(data *SampleTable) error {
	sess, err := c.ConnectX()
	if err != nil {
		return err
	}

	if err := sess.Query(sampleTableInsertStmt, sampleTableInsertNames).BindStruct(data).ExecRelease(); err != nil {
		return err
	}

	return nil
}

var sampleTableUpdateStmt, sampleTableUpdateNames = sampleTable.Update()

func (c *%s) Update(data *SampleTable) error {
	sess, err := c.ConnectX()
	if err != nil {
		return err
	}

	if err := sess.Query(sampleTableUpdateStmt, sampleTableUpdateNames).BindStruct(data).ExecRelease(); err != nil {
		return err
	}

	return nil
}

var sampleTableDeleteStmt, sampleTableDeleteNames = sampleTable.Delete()

func (c *%s) Delete(column1, column2 string) error {
	sess, err := c.ConnectX()
	if err != nil {
		return err
	}

	if err := sess.Query(sampleTableDeleteStmt, sampleTableDeleteNames).BindMap(map[string]any{
		"column1": column1,
		"column2": column2,
	}).ExecRelease(); err != nil {
		return err
	}

	return nil
}`
)

func createFxCassandraFile(path string, name string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	name = strings.ToUpper(name[:1]) + name[1:]
	packageName := filepath.Base(path)

	fileName := filepath.Join(path, fmt.Sprintf(fxFileName, name))
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := func() error {
		content := fmt.Sprintf(cassandraTemplate1, packageName, name, name, name, name, name, name, name, name, name, name, name, name)
		if _, err := file.WriteString(content); err != nil {
			return err
		}

		if err := file.Sync(); err != nil {
			return err
		}

		return nil
	}(); err != nil {
		return err
	}

	if err := func() error {
		fileName = filepath.Join(path, "table.go")
		file, err = os.Create(fileName)
		if err != nil {
			return err
		}
		defer file.Close()

		content := fmt.Sprintf(cassandraTemplate2, packageName, name, name, name, name)
		if _, err := file.WriteString(content); err != nil {
			return err
		}

		if err := file.Sync(); err != nil {
			return err
		}

		return nil
	}(); err != nil {
		return err
	}

	if err := getGoModule(cassandraDependencyPackage1); err != nil {
		return err
	}

	if err := getGoModule(cassandraDependencyPackage2); err != nil {
		return err
	}

	return nil
}
