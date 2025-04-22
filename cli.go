package main

import (
	"context"
	"path/filepath"

	"github.com/urfave/cli/v3"
)

var rootCmd = &cli.Command{
	Name:  "ako",
	Usage: "Manage your Go project with ako",
	Commands: []*cli.Command{
		{
			Name:    "init",
			Aliases: []string{"i"},
			Arguments: []cli.Argument{
				&cli.StringArg{
					Name:      "module-name",
					Value:     "github.com/ako/ako",
					UsageText: "The name of the module to initialize",
					Config:    cli.StringConfig{TrimSpace: true},
				},
			},
			Usage: "Initialize a new Go module",
			Action: func(ctx context.Context, command *cli.Command) error {
				if len(command.Arguments) < 1 {
					return cli.Exit("Module name is required", 1)
				}

				moduleName, ok := command.Arguments[0].Get().(string)
				if !ok {
					return cli.Exit("Invalid module name", 1)
				}

				if err := initGoModule(moduleName); err != nil {
					return cli.Exit(err.Error(), 1)
				}

				if err := createPackageTemplate(); err != nil {
					return cli.Exit(err.Error(), 1)
				}

				if err := createBufTemplate(); err != nil {
					return cli.Exit(err.Error(), 1)
				}

				if err := createProtobufExample(); err != nil {
					return cli.Exit(err.Error(), 1)
				}

				if err := getFxDependency(); err != nil {
					return cli.Exit(err.Error(), 1)
				}

				return nil
			},
		},
		{
			Name:    "buf",
			Aliases: []string{"f"},
			Usage:   "Generate protobuf files using buf",
			Action: func(ctx context.Context, command *cli.Command) error {
				if err := runGoModuleTool("buf", "generate"); err != nil {
					return cli.Exit(err.Error(), 1)
				}

				return nil
			},
		},
		{
			Name:        "lib",
			Usage:       "Generate core abstraction layer (in lib/)",
			Description: "Scaffolds the core abstraction layer (lib/) of your Go project.\n   This layer contains interface definitions, shared data structures (DTOs, VOs, Entities),\n   and domain models, free of concrete implementations. It establishes the contracts\n   and core concepts for other layers (internal, pkg) to depend on.",
			Aliases:     []string{"l"},
			Action: func(ctx context.Context, command *cli.Command) error {
				base, err := selectLibraryBase()
				if err != nil {
					return cli.Exit(err.Error(), 1)
				}

				packageName, err := inputLibraryPackage()
				if err != nil {
					return cli.Exit(err.Error(), 1)
				}

				name, err := inputLibraryName()
				if err != nil {
					return cli.Exit(err.Error(), 1)
				}

				path := makeLibraryPath(base, packageName)
				if err := createLibraryFile(path, name); err != nil {
					return cli.Exit(err.Error(), 1)
				}

				return nil
			},
		},
		{
			Name:        "pkg",
			Aliases:     []string{"p"},
			Usage:       "Generate new package implementation (in pkg/)",
			Description: "Generates a new package within the pkg/ directory. This layer contains\n   the concrete implementations of interfaces defined in the lib/ layer. Packages\n   within pkg/ are typically organized based on the specific technology or external\n   dependency they integrate with (e.g., postgres, redis, zerolog, stripe).\n   This command helps scaffold the necessary directory structure and boilerplate\n   files for the implementation.",
			Action: func(ctx context.Context, command *cli.Command) error {
				base, err := inputPackageBase()
				if err != nil {
					return cli.Exit(err.Error(), 1)
				}

				category, err := inputPackageCategory()
				if err != nil {
					return cli.Exit(err.Error(), 1)
				}

				packageName, err := inputPackageName()
				if err != nil {
					return cli.Exit(err.Error(), 1)
				}

				templateKey, err := selectFxPkgTemplateKey()
				if err != nil {
					return cli.Exit(err.Error(), 1)
				}

				templateWriter, err := getPkgTemplateWriter(templateKey)
				if err != nil {
					return cli.Exit(err.Error(), 1)
				}

				path := makePackagePath(base, category, packageName)

				if err := templateWriter(path, packageName); err != nil {
					return cli.Exit(err.Error(), 1)
				}

				return nil
			},
		},
	},
}

var pkgGenerateArguments = append([]cli.Argument{
	&cli.StringArg{
		Name:      "path",
		Value:     "client/http",
		UsageText: "The path to the package to create [relative to the pkg folder, e.g. client/http]",
		Config:    cli.StringConfig{TrimSpace: true},
	},
}, pkgGenerateSpecificArguments...)

var pkgGenerateSpecificArguments = []cli.Argument{
	&cli.StringArg{
		Name:      "name",
		Value:     "client",
		UsageText: "The name of the struct to create [e.g. client]",
		Config:    cli.StringConfig{TrimSpace: true},
	},
}

func getPkgGenerateArguments(ctx context.Context, command *cli.Command) (string, string, error) {
	if len(command.Arguments) < 2 {
		return "", "", cli.Exit("Path and name are required", 1)
	}

	path, ok := command.Arguments[0].Get().(string)
	if !ok {
		return "", "", cli.Exit("Invalid path", 1)
	}
	path = filepath.Join("pkg", path)

	name, ok := command.Arguments[1].Get().(string)
	if !ok {
		return "", "", cli.Exit("Invalid name", 1)
	}

	return path, name, nil
}
