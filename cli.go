package main

import (
	"context"
	"log"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
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
			Usage: "Initialize a new Go module and Git repository",
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

				if err := generateDevContainerFile(filepath.Base(moduleName)); err != nil {
					return cli.Exit(err.Error(), 1)
				}

				if err := generateCommitMessageRule(); err != nil {
					return cli.Exit(err.Error(), 1)
				}

				if err := initGit(gitBranchPrefixRelease); err != nil {
					return cli.Exit(err.Error(), 1)
				}

				if err := addGitFiles("."); err != nil {
					return cli.Exit(err.Error(), 1)
				}

				if err := commitGitFiles("feat(all): initialized project"); err != nil {
					return cli.Exit(err.Error(), 1)
				}

				return nil
			},
		},
		{
			Name:    "go",
			Aliases: []string{"g"},
			Usage:   "Organize Go project",
			Commands: []*cli.Command{
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

						category, err := inputLibraryCategory()
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

						path := makeLibraryPath(base, category, packageName)
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
				{
					Name:        "internal",
					Aliases:     []string{"n"},
					Usage:       "Generate new internal implementation (in internal/)",
					Description: "Scaffolds the business logic layer within the internal/ directory. This layer\n   typically contains 'controller' packages for handling requests/responses and 'service'\n   packages for orchestrating core business logic and use cases. It primarily depends\n   on the abstractions defined in lib/. Go's 'internal' visibility rules apply.\n   This command helps set up the structure for controllers and services for a given domain.",
					Action: func(ctx context.Context, command *cli.Command) error {
						base, err := selectInternalPackageBase()
						if err != nil {
							return cli.Exit(err.Error(), 1)
						}

						category, err := inputInternalPackageCategory()
						if err != nil {
							return cli.Exit(err.Error(), 1)
						}

						packageName, err := inputInternalPackageName()
						if err != nil {
							return cli.Exit(err.Error(), 1)
						}

						if err := createInternalPackage(base, category, packageName); err != nil {
							return cli.Exit(err.Error(), 1)
						}

						return nil
					},
				},
				{
					Name:        "cmd",
					Aliases:     []string{"c"},
					Usage:       "Generate new command implementation (in cmd/)",
					Description: "Creates and manages the application's execution entry point (main package).\n   Its main role is to load configuration, assemble (wire) components\n   from other layers (pkg, internal) via dependency injection, and finally\n   run the application (e.g., HTTP server, worker).\n   Does not contain business logic.",
					Action: func(ctx context.Context, command *cli.Command) error {
						name, err := inputCmdName()
						if err != nil {
							return cli.Exit(err.Error(), 1)
						}

						dir := filepath.Join("cmd", name)

						if err := createFxExecutableFile(dir); err != nil {
							return cli.Exit(err.Error(), 1)
						}

						if err := generateGoImageFile(name); err != nil {
							return cli.Exit(err.Error(), 1)
						}

						return nil
					},
				},
			},
		},
		{
			Name:    "branch",
			Aliases: []string{"b"},
			Usage:   "Organize Git branches and commits",
			Commands: []*cli.Command{
				{
					Name:    "current",
					Aliases: []string{"n"},
					Usage:   "Get the current branch name",
					Action: func(ctx context.Context, command *cli.Command) error {
						branch, err := getGitBranchName()
						if err != nil {
							return cli.Exit(err.Error(), 1)
						}

						log.Printf("Current branch name: %s", branch)
						return nil
					},
				},
				{
					Name:    "commit",
					Aliases: []string{"m"},
					Usage:   "Create a new message and commit",
					Action: func(ctx context.Context, command *cli.Command) error {
						message, err := buildGitCommitMessage()
						if err != nil {
							return cli.Exit(err.Error(), 1)
						}

						log.Printf("Git commit message: %s", message)

						if err := commitGitFiles(message); err != nil {
							return cli.Exit(err.Error(), 1)
						}

						log.Printf("Git Committed files successfully")

						return nil
					},
				},
				{
					Name:    "create",
					Aliases: []string{"c"},
					Usage:   "Create a new branch",
					Action: func(ctx context.Context, command *cli.Command) error {
						currentBranch, err := getGitBranchName()
						if err != nil {
							return cli.Exit(err.Error(), 1)
						}

						created, err := makeGitSubBranchName(currentBranch)
						if err != nil {
							return cli.Exit(err.Error(), 1)
						}

						if err := switchOrCreateGitBranchTo(created); err != nil {
							return cli.Exit(err.Error(), 1)
						}

						log.Printf("Switched to branch: %s", created)

						return nil
					},
				},
				{
					Name:    "up",
					Aliases: []string{"u"},
					Usage:   "Up to parent branch",
					Action: func(ctx context.Context, command *cli.Command) error {
						branches, err := getParentBranchName()
						if err != nil {
							return cli.Exit(err.Error(), 1)
						}

						if len(branches) == 0 {
							return cli.Exit("No parent branch found", 1)
						}

						selectedBranch := ""
						if err := survey.AskOne(&survey.Select{
							Message: "Choose branch",
							Options: branches,
						}, &selectedBranch, survey.WithValidator(survey.Required)); err != nil {
							return cli.Exit(err.Error(), 1)
						}

						if err := switchGitBranchTo(selectedBranch); err != nil {
							return cli.Exit(err.Error(), 1)
						}

						return nil
					},
				},
				{
					Name:    "down",
					Aliases: []string{"d"},
					Usage:   "Down to child branch",
					Action: func(ctx context.Context, command *cli.Command) error {
						branches, err := getChildrenBranchName()
						if err != nil {
							return cli.Exit(err.Error(), 1)
						}

						if len(branches) == 0 {
							return cli.Exit("No child branch found", 1)
						}

						selectedBranch := ""
						if err := survey.AskOne(&survey.Select{
							Message: "Choose branch",
							Options: branches,
						}, &selectedBranch, survey.WithValidator(survey.Required)); err != nil {
							return cli.Exit(err.Error(), 1)
						}

						if err := switchGitBranchTo(selectedBranch); err != nil {
							return cli.Exit(err.Error(), 1)
						}

						return nil
					},
				},
			},
		},
		{
			Name:    "k3d",
			Aliases: []string{"k"},
			Usage:   "Manage K3S manifests and clusters",
			Commands: []*cli.Command{
				{
					Name:    "registry",
					Aliases: []string{"r"},
					Usage:   "Manage K3D registries",
					Commands: []*cli.Command{
						{
							Name:    "create",
							Aliases: []string{"c"},
							Usage:   "Create a new K3D registry",
							Action: func(ctx context.Context, command *cli.Command) error {
								name, err := inputK3dRegistryName()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								if err := createK3dRegistry(name); err != nil {
									return cli.Exit(err.Error(), 1)
								}

								log.Printf("Created K3D registry: %s", name)
								return nil
							},
						},
						{
							Name:    "delete",
							Aliases: []string{"d", "rm"},
							Usage:   "Delete a K3D registry",
							Action: func(ctx context.Context, command *cli.Command) error {
								selected, err := selectK3dRegistryName()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								for _, name := range selected {
									if err := deleteK3dRegistry(name); err != nil {
										return cli.Exit(err.Error(), 1)
									}

									log.Printf("Deleted K3D registry: %s", name)
								}

								return nil
							},
						},
						{
							Name:    "list",
							Aliases: []string{"ls", "l"},
							Usage:   "List K3D registries",
							Action: func(ctx context.Context, command *cli.Command) error {
								registries, err := getK3dRegistries()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								if len(registries) == 0 {
									log.Println("No registries found")
									return nil
								}

								tbl := NewTableBuilder("NAME", "IMAGE BUILD TAG", "MANIFEST TAG", "STATUS")

								for _, registry := range registries {
									addr := "localhost:" + registry.PortMappings.Five000TCP[0].HostPort
									tbl.AppendRow(registry.Name, addr, registry.Name+"."+addr, registry.State.Status)
								}

								tbl.Print()

								return nil
							},
						},
					},
				},
				{
					Name:    "cluster",
					Aliases: []string{"c"},
					Usage:   "Manage K3D clusters",
					Commands: []*cli.Command{
						{
							Name:    "create",
							Aliases: []string{"c"},
							Usage:   "Create a new K3D cluster",
							Action: func(ctx context.Context, command *cli.Command) error {
								name, err := inputK3dClusterName()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								agents, err := inputK3dClusterAgents()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								portMap, err := inputK3dClusterLoadBalancerPortMap()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								registryData, err := selectK3dRegistryForCluster()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								if err := createK3dCluster(name, agents, registryData, portMap); err != nil {
									return cli.Exit(err.Error(), 1)
								}

								log.Printf("Created K3D cluster: %s", name)
								return nil
							},
						},
						{
							Name:    "delete",
							Aliases: []string{"d", "rm"},
							Usage:   "Delete a K3D cluster",
							Action: func(ctx context.Context, command *cli.Command) error {
								selected, err := selectK3dClusterNames()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								for _, name := range selected {
									if err := deleteK3dCluster(name); err != nil {
										return cli.Exit(err.Error(), 1)
									}

									log.Printf("Deleted K3D cluster: %s", name)
								}

								return nil
							},
						},
						{
							Name:    "list",
							Aliases: []string{"ls", "l"},
							Usage:   "List K3D clusters",
							Action: func(ctx context.Context, command *cli.Command) error {
								clusters, err := getK3dClusters()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								if len(clusters) == 0 {
									log.Println("No clusters found")
									return nil
								}

								tbl := NewTableBuilder("NAME", "SERVERS", "AGENTS", "RUNNING", "OUTBOUND PORT (HOST -> CONTAINER)")

								for _, cluster := range clusters {
									builder := strings.Builder{}
									for _, node := range cluster.Nodes {
										const loadBalancerRole = "loadbalancer"
										if node.Role == loadBalancerRole {
											for k, v := range node.PortMappings {
												for _, v := range v {
													builder.WriteString(v.HostPort)
													builder.WriteString(" -> ")
													builder.WriteString(k)
													builder.WriteString("\n")
												}
											}
										}
									}
									tbl.AppendRow(cluster.Name, cluster.ServersCount, cluster.AgentsCount, (cluster.AgentsRunning > 0) && (cluster.ServersRunning > 0), builder.String())
								}

								tbl.Print()

								return nil
							},
						},
						{
							Name:    "append-port",
							Aliases: []string{"ap", "a"},
							Usage:   "Append port to K3D cluster",
							Action: func(ctx context.Context, command *cli.Command) error {
								selected, err := selectK3dClusterNames()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								portMap, err := inputK3dClusterLoadBalancerPortMap()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								for _, name := range selected {
									for h, c := range portMap {
										if err := addK3dClusterPort(name, h, c); err != nil {
											return cli.Exit(err.Error(), 1)
										}
									}

									log.Printf("Appended port to K3D cluster: %s", name)
								}

								return nil
							},
						},
					},
				},
				{
					Name:    "manifest",
					Aliases: []string{"m", "f"},
					Usage:   "Manage K3D manifests",
					Commands: []*cli.Command{
						{
							Name:    "init",
							Aliases: []string{"i"},
							Usage:   "Initialize a new K3D manifest",
							Action: func(ctx context.Context, command *cli.Command) error {
								selectedCluster, err := SelectK3dClusterName()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								selectedLocalRegistry, err := selectK3dRegistryForCluster()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								namespace, err := inputK8sNamespace()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								remoteRegistry, err := inputK8sRemoteRegistry()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								globalConfig.Cluster = selectedCluster
								globalConfig.Namespace = namespace
								globalConfig.LocalRegistry = selectedLocalRegistry
								globalConfig.RemoteRegistry = remoteRegistry
								if err := saveK3dConfig(); err != nil {
									return cli.Exit(err.Error(), 1)
								}

								log.Printf("Initialized K3D manifest for cluster: %s", selectedCluster)
								log.Printf("Local registry: %s", selectedLocalRegistry)
								log.Printf("Remote registry: %s", remoteRegistry)
								log.Printf("Namespace: %s", namespace)

								log.Printf("K3D manifest initialized successfully")

								return nil
							},
						},
						{
							Name:    "create",
							Aliases: []string{"c"},
							Usage:   "Create a new K3D manifest",
							Action: func(ctx context.Context, command *cli.Command) error {
								selectedCmd, err := selectCmdName()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								selectedKind, err := selectK8sManifestKind()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								log.Printf("Created K3D manifest: %s for %s", selectedKind, selectedCmd)

								return nil
							},
						},
					},
				},
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
