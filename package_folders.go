package main

import "path/filepath"

const (
	packagePersistence = "persistence"
	packageCache       = "cache"
	packageClient      = "client"
	packageTransport   = "transport"
)

func makePackagePath(base string, category string, name string) string {
	return filepath.Join("pkg", base, category, name)
}
