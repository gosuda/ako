package main

import (
	"context"

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
	},
}
