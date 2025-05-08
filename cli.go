package main

import (
	"context"
	"log"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/urfave/cli/v3"

	"github.com/gosuda/ako/generator/ai"
	"github.com/gosuda/ako/generator/ci"
	docker2 "github.com/gosuda/ako/generator/docker"
	k8s2 "github.com/gosuda/ako/generator/k8s"
	"github.com/gosuda/ako/generator/lint"
	packages2 "github.com/gosuda/ako/generator/packages"
	"github.com/gosuda/ako/generator/protocol"
	git2 "github.com/gosuda/ako/util/git"
	"github.com/gosuda/ako/util/module"
	"github.com/gosuda/ako/util/table"
)

var rootCmd = &cli.Command{
	Name:  "ako",
	Usage: "Manage your Go project with ako",
	Commands: []*cli.Command{
		{
			Name:    "init",
			Aliases: []string{"i"},
			Usage:   "Initialize a new Go module and Git repository",
			Action: func(ctx context.Context, command *cli.Command) error {
				moduleName, err := module.InputGoModuleName()
				if err != nil {
					return cli.Exit(err.Error(), 1)
				}

				ciTemplate, err := ci.SelectCITemplate()
				if err != nil {
					return cli.Exit(err.Error(), 1)
				}

				loggerLibrary, err := packages2.SelectLoggerLibrary()
				if err != nil {
					return cli.Exit(err.Error(), 1)
				}

				if err := module.InitGoModule(moduleName); err != nil {
					return cli.Exit(err.Error(), 1)
				}

				if err := packages2.CreatePackageTemplate(); err != nil {
					return cli.Exit(err.Error(), 1)
				}

				if err := protocol.CreateBufTemplate(); err != nil {
					return cli.Exit(err.Error(), 1)
				}

				if err := protocol.CreateProtobufExample(); err != nil {
					return cli.Exit(err.Error(), 1)
				}

				if err := packages2.GetFxDependency(); err != nil {
					return cli.Exit(err.Error(), 1)
				}

				if err := docker2.GenerateDevContainerFile(filepath.Base(moduleName)); err != nil {
					return cli.Exit(err.Error(), 1)
				}

				if err := packages2.CreateLoggerWriterFile(loggerLibrary); err != nil {
					return cli.Exit(err.Error(), 1)
				}

				if err := lint.CreateGolangcilintConfig(); err != nil {
					return cli.Exit(err.Error(), 1)
				}

				if err := lint.InstallGolangcilint(); err != nil {
					return cli.Exit(err.Error(), 1)
				}

				if err := ci.CreateCITemplate(ciTemplate); err != nil {
					return cli.Exit(err.Error(), 1)
				}

				if err := git2.CreateGitIgnoreFile(); err != nil {
					return cli.Exit(err.Error(), 1)
				}

				if err := ai.CreateVsCodeCopilotSettings(); err != nil {
					return cli.Exit(err.Error(), 1)
				}

				if err := git2.GenerateCommitMessageRule(); err != nil {
					return cli.Exit(err.Error(), 1)
				}

				if err := git2.InitGit(git2.GitBranchPrefixRelease); err != nil {
					return cli.Exit(err.Error(), 1)
				}

				if err := git2.AddGitFiles("."); err != nil {
					return cli.Exit(err.Error(), 1)
				}

				if err := git2.CommitGitFiles("feat(all): initialized project"); err != nil {
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
						if err := module.RunGoModuleTool("buf", "generate"); err != nil {
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
						base, err := packages2.SelectLibraryBase()
						if err != nil {
							return cli.Exit(err.Error(), 1)
						}

						packageName, err := packages2.InputLibraryPackage()
						if err != nil {
							return cli.Exit(err.Error(), 1)
						}

						name, err := packages2.InputLibraryName()
						if err != nil {
							return cli.Exit(err.Error(), 1)
						}

						path := packages2.MakeLibraryPath(base, packageName)
						if err := packages2.CreateLibraryFile(path, name); err != nil {
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
						base, err := packages2.InputPackageBase()
						if err != nil {
							return cli.Exit(err.Error(), 1)
						}

						packageName, err := packages2.InputPackageName()
						if err != nil {
							return cli.Exit(err.Error(), 1)
						}

						templateKey, err := packages2.SelectFxPkgTemplateKey()
						if err != nil {
							return cli.Exit(err.Error(), 1)
						}

						templateWriter, err := packages2.GetPkgTemplateWriter(templateKey)
						if err != nil {
							return cli.Exit(err.Error(), 1)
						}

						path := packages2.MakePackagePath(base, packageName)

						if err := templateWriter(path, filepath.Base(path)); err != nil {
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
						base, err := packages2.SelectInternalPackageBase()
						if err != nil {
							return cli.Exit(err.Error(), 1)
						}

						packageName, err := packages2.InputInternalPackageName()
						if err != nil {
							return cli.Exit(err.Error(), 1)
						}

						path := filepath.Join(base, packageName)

						if err := packages2.CreateInternalPackage(filepath.Dir(path), filepath.Base(path)); err != nil {
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
						name, err := packages2.InputCmdName()
						if err != nil {
							return cli.Exit(err.Error(), 1)
						}

						dir := filepath.Join("cmd", name)

						if err := packages2.CreateFxExecutableFile(dir); err != nil {
							return cli.Exit(err.Error(), 1)
						}

						if err := docker2.GenerateGoImageFile(name); err != nil {
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
						branch, err := git2.GetGitBranchName()
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
						message, err := git2.BuildGitCommitMessage()
						if err != nil {
							return cli.Exit(err.Error(), 1)
						}

						log.Printf("Git commit message: %s", message)

						if err := git2.CommitGitFiles(message); err != nil {
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
						currentBranch, err := git2.GetGitBranchName()
						if err != nil {
							return cli.Exit(err.Error(), 1)
						}

						created, err := git2.MakeGitSubBranchName(currentBranch)
						if err != nil {
							return cli.Exit(err.Error(), 1)
						}

						if err := git2.SwitchOrCreateGitBranchTo(created); err != nil {
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
						branches, err := git2.GetParentBranchName()
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

						if err := git2.SwitchGitBranchTo(selectedBranch); err != nil {
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
						branches, err := git2.GetChildrenBranchName()
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

						if err := git2.SwitchGitBranchTo(selectedBranch); err != nil {
							return cli.Exit(err.Error(), 1)
						}

						return nil
					},
				},
			},
		},
		{
			Name:    "linter",
			Aliases: []string{"l"},
			Usage:   "Run linter",
			Action: func(ctx context.Context, command *cli.Command) error {
				if err := lint.RunGolangcilint(); err != nil {
					return cli.Exit(err.Error(), 1)
				}

				return nil
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
								name, err := k8s2.InputK3dRegistryName()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								if err := k8s2.CreateK3dRegistry(name); err != nil {
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
								selected, err := k8s2.SelectK3dRegistryNames()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								for _, name := range selected {
									if err := k8s2.DeleteK3dRegistry(name); err != nil {
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
								registries, err := k8s2.GetK3dRegistries()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								if len(registries) == 0 {
									log.Println("No registries found")
									return nil
								}

								tbl := table.NewTableBuilder("NAME", "IMAGE BUILD TAG", "MANIFEST TAG", "STATUS")

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
								name, err := k8s2.InputK3dClusterName()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								agents, err := k8s2.InputK3dClusterAgents()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								portMap, err := k8s2.InputK3dClusterLoadBalancerPortMap()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								registryData, err := k8s2.SelectK3dRegistryName()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								if err := k8s2.CreateK3dCluster(name, agents, registryData, portMap); err != nil {
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
								selected, err := k8s2.SelectK3dClusterNames()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								for _, name := range selected {
									if err := k8s2.DeleteK3dCluster(name); err != nil {
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
								clusters, err := k8s2.GetK3dClusters()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								if len(clusters) == 0 {
									log.Println("No clusters found")
									return nil
								}

								tbl := table.NewTableBuilder("NAME", "SERVERS", "AGENTS", "RUNNING", "OUTBOUND PORT (HOST -> CONTAINER)")

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
								selected, err := k8s2.SelectK3dClusterNames()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								portMap, err := k8s2.InputK3dClusterLoadBalancerPortMap()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								for _, name := range selected {
									for h, c := range portMap {
										if err := k8s2.AddK3dClusterPort(name, h, c); err != nil {
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
								selectedCluster, err := k8s2.SelectK3dClusterName()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								selectedLocalRegistry, err := k8s2.SelectK3dRegistryForCluster()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								namespace, err := k8s2.InputK8sNamespace()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								remoteRegistry, err := k8s2.InputK8sRemoteRegistry()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								k8s2.GlobalConfig.Cluster = selectedCluster
								k8s2.GlobalConfig.Namespace = namespace
								k8s2.GlobalConfig.LocalRegistry = selectedLocalRegistry
								k8s2.GlobalConfig.RemoteRegistry = remoteRegistry
								if err := k8s2.SaveK3dConfig(); err != nil {
									return cli.Exit(err.Error(), 1)
								}

								if err := k8s2.GenerateK8sNamespaceFile(namespace); err != nil {
									return cli.Exit(err.Error(), 1)
								}

								if err := k8s2.GenerateK8sIngressFile(namespace, "public"); err != nil {
									return cli.Exit(err.Error(), 1)
								}

								if err := k8s2.GenerateK8sIngressFile(namespace, "private"); err != nil {
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
								selectedCmd, err := packages2.SelectCmdName()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								selectedKind, err := k8s2.SelectK8sManifestKind()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								cmds := strings.Split(selectedCmd, "/")

								switch selectedKind {
								case k8s2.K8sManifestKindDeployment:
									tier, err := k8s2.SelectK8sDeploymentTier()
									if err != nil {
										return cli.Exit(err.Error(), 1)
									}

									if err := k8s2.GenerateK8sDeploymentFile(tier, k8s2.GlobalConfig.Namespace, cmds...); err != nil {
										return cli.Exit(err.Error(), 1)
									}

									if err := k8s2.GenerateK8sServiceFile(k8s2.GlobalConfig.Namespace, cmds...); err != nil {
										return cli.Exit(err.Error(), 1)
									}
								case k8s2.K8sManifestKindCronJob:
									if err := k8s2.GenerateK8sCronJobFile(k8s2.GlobalConfig.Namespace, cmds...); err != nil {
										return cli.Exit(err.Error(), 1)
									}
								default:
									log.Printf("Unknown K3D manifest kind: %s", selectedKind)
									return cli.Exit("Unknown K3D manifest kind", 1)
								}

								if err := k8s2.GenerateK8sConfigMap(k8s2.GlobalConfig.Namespace, cmds...); err != nil {
									return cli.Exit(err.Error(), 1)
								}

								if err := k8s2.GenerateK8sPvcFile(k8s2.GlobalConfig.Namespace, cmds...); err != nil {
									return cli.Exit(err.Error(), 1)
								}

								log.Printf("Created K3D manifest for command: %s", selectedCmd)

								return nil
							},
						},
						{
							Name:    "build",
							Aliases: []string{"b", "d", "deploy"},
							Usage:   "Build cmd and push to local registry",
							Action: func(ctx context.Context, command *cli.Command) error {
								selectedCmd, err := packages2.SelectCmdName()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								cmds := strings.Split(selectedCmd, "/")

								if err := docker2.BuildDockerImage(cmds...); err != nil {
									return cli.Exit(err.Error(), 1)
								}

								log.Printf("Built K3D manifest for command: %s", selectedCmd)

								return nil
							},
						},
						{
							Name:    "apply",
							Aliases: []string{"a"},
							Usage:   "Apply K3D manifest",
							Action: func(ctx context.Context, command *cli.Command) error {
								selectedManifests, err := k8s2.SelectK8sManifest()
								if err != nil {
									return cli.Exit(err.Error(), 1)
								}

								for _, manifest := range selectedManifests {
									log.Printf("Applied K3D manifest: %s", manifest)
									if err := k8s2.ApplyK8sManifest(manifest); err != nil {
										return cli.Exit(err.Error(), 1)
									}

									log.Printf("Applied K3D manifest successfully")
								}

								return nil
							},
						},
						{
							Name:    "get",
							Aliases: []string{"g"},
							Usage:   "Get K3D resources",
							Commands: []*cli.Command{
								{
									Name:    "pods",
									Aliases: []string{"p", "po"},
									Usage:   "Get K3D pods",
									Action: func(ctx context.Context, command *cli.Command) error {
										if err := k8s2.RunK8sGetPods(); err != nil {
											return cli.Exit(err.Error(), 1)
										}

										return nil
									},
								},
								{
									Name:    "services",
									Aliases: []string{"s", "svc"},
									Usage:   "Get K3D services",
									Action: func(ctx context.Context, command *cli.Command) error {
										if err := k8s2.RunK8sGetServices(); err != nil {
											return cli.Exit(err.Error(), 1)
										}

										return nil
									},
								},
								{
									Name:    "deployments",
									Aliases: []string{"d", "deploy"},
									Usage:   "Get K3D deployments",
									Action: func(ctx context.Context, command *cli.Command) error {
										if err := k8s2.RunK8sGetDeployments(); err != nil {
											return cli.Exit(err.Error(), 1)
										}

										return nil
									},
								},
								{
									Name:    "ingress",
									Aliases: []string{"i"},
									Usage:   "Get K3D ingress",
									Action: func(ctx context.Context, command *cli.Command) error {
										if err := k8s2.RunK8sGetIngress(); err != nil {
											return cli.Exit(err.Error(), 1)
										}

										return nil
									},
								},
							},
						},
					},
				},
			},
		},
	},
}
