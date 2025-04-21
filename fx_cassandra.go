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
	cassandraTemplate           = `package cassandra

import (
	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v2"
)

type %s struct {
	clusterConfig gocql.ClusterConfig
}

func New(host ...string) *%s {
	cfg := gocql.NewCluster(host...)

	return &%s{
		clusterConfig: *cfg,
	}
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
)

func createFxCassandraFile(path string, name string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	name = strings.ToUpper(name[:1]) + name[1:]

	fileName := filepath.Join(path, fmt.Sprintf(fxFileName, name))
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	content := fmt.Sprintf(cassandraTemplate, name, name, name, name, name, name, name)
	if _, err := file.WriteString(content); err != nil {
		return err
	}

	if err := file.Sync(); err != nil {
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
