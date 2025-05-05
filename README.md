# ako: Opinionated Go Project Manager

[![Go Report Card](https://goreportcard.com/badge/github.com/gosuda/ako)](https://goreportcard.com/report/github.com/gosuda/ako)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

`ako` is a CLI tool for enhancing the productivity and standardization of Go projects. It automates repetitive setup, code structuring, Git management, and local K3d environment configuration, helping developers focus on core logic.

## Problem to Solve

Starting and continuously managing Go projects often involves the following inefficiencies and difficulties:

1.  Complex and time-consuming initial setup:
  * Every new project requires basic tasks like creating a Go module, initializing a Git repository, and setting up a basic `.gitignore` file.
  * Configuring CI/CD pipelines (e.g., GitHub Actions workflows), setting up linters (`golangci-lint`) for code quality management, and configuring `buf` for Protobuf usage require specialized knowledge, are cumbersome, and prone to errors.
  * Considering Dev Container setup for development environment consistency adds significant time and effort to the initial setup.

2.  Inconsistent project structure:
  * Without clear guidelines, teams or individuals interpret and use directories like `pkg`, `internal`, `cmd`, and `lib` differently.
  * This lowers code cohesion, complicates dependency management, and causes new team members to spend unnecessary time understanding the project structure, ultimately increasing maintenance costs.

3.  Repetitive boilerplate code writing:
  * Basic code structures, Fx module setups, and interface definitions required for specific layers (e.g., HTTP handlers in `internal`, database clients in `pkg`) need to be written similarly each time.
  * This repetitive task slows down development speed and distracts from focusing on core feature development.

4.  Inefficient Git workflow management:
  * Without rules for creating and managing branches for features, bug fixes, etc., branch names become haphazard, making history tracking difficult.
  * Manually managing hierarchical branch structures, especially for complex features involving multiple sub-tasks, is very cumbersome.
  * Not enforcing a commit message format (e.g., Conventional Commits) makes it hard to understand the intent of changes and hinders automated changelog generation and version management.

5.  Difficult local cloud-native environment setup:
  * Configuring a local Kubernetes environment for container-based application development requires significant effort.
  * Setting up a local image registry, creating and configuring K3d clusters and networks, and writing and managing manifests (Deployment, Service, Ingress, etc.) for application deployment are complex and time-consuming.

## `ako`'s Solution & Goals

`ako` aims to solve the problems mentioned above by providing the following standardized and automated features:

1.  One-click project initialization (`ako init` / `ako i`):
  * Automatically configures almost everything needed to start a project with a single command execution: Go module, Git repository (including `.gitignore`), selectable CI/CD templates, `buf` setup and examples, Uber Fx dependency, Dev Container setup, `golangci-lint` setup and binary installation, Conventional Commits rules setup, and default `release` branch creation.
  * This frees developers from the complex initial setup process, allowing them to focus on core code development immediately after project creation.

2.  Enforced standardized layer architecture and code generation (`ako go` / `ako g`):
  * Clearly presents a layer structure (`lib`, `pkg`, `internal`, `cmd`) and guides code generation suitable for each layer's role.
    * `lib/`: Defines core application abstractions like domain models, interfaces, and DTOs to manage dependency direction.
    * `pkg/`: Handles integration with external infrastructure and specific technology implementations like databases and external API clients. Implements interfaces from `lib/`.
    * `internal/`: Implements application use cases and business logic, such as HTTP handlers and gRPC services. Mainly depends on `lib/`.
    * `cmd/`: Serves as the application entry point, responsible for loading configuration, dependency injection using Uber Fx and component assembly, and running the application.
  * The `ako go internal` and `ako go pkg` commands allow selecting predefined templates (e.g., Chi handler, Redis client) to automatically generate Fx module-based boilerplate code, thus speeding up development and ensuring code consistency.
  * Simplifies running `buf generate` with the `ako go buf` (`ako g f`) command, facilitating Protobuf-based development.

3.  Systematic Git workflow support (`ako branch` / `ako b`):
  * `ako branch commit` (`ako b m`) supports writing commit messages following the Conventional Commits convention through an interactive prompt, helping to easily create consistent and meaningful Git history.
  * `ako branch create` (`ako b c`) automates the creation of hierarchical branches in the `type/scope/description` format, and the `up` (`b u`), `down` (`b d`) commands simplify navigation between parent/child branches, enabling systematic management of complex feature development flows.

4.  Easy code quality checks (`ako linter` / `ako l`):
  * A single `ako linter` (`ako l`) command runs `golangci-lint` across the entire project, helping to detect code style issues early and maintain consistent code quality.

5.  Simplified local K3d environment management (`ako k3d` / `ako k`):
  * Automates K3d registry and cluster creation/deletion/listing (`ako k3d registry`, `ako k3d cluster`) with simple commands, removing the complexity of setting up local Kubernetes infrastructure.
  * `ako k3d manifest init` (`ako k m i`) initializes basic K8s manifest settings like namespaces and ingresses suitable for the project.
  * `ako k3d manifest create` (`ako k m c`) automatically generates necessary manifest files (Deployment, Service, ConfigMap, etc.) for `cmd` applications.
  * `ako k3d manifest build` (`ako k m b`) builds the application's Docker image and pushes it to the local K3d registry, while `ako k3d manifest apply` (`ako k m a`) allows easy deployment to the cluster.
  * `ako k3d manifest get` (`ako k m g`) makes it easy to check `kubectl get` command results for monitoring deployment status.

## `ako`'s Philosophy

`ako` is designed based on the following core philosophies:

* Standardization: Aims to increase code consistency, facilitate collaboration within and between teams, and reduce long-term maintenance costs by presenting a well-defined project structure and development workflow. This reduces the cognitive load on developers.
* Developer Experience: Helps developers focus more time and energy on core logic development that creates actual business value by automating cumbersome and error-prone tasks like project initialization, repetitive code writing, and complex Git and K3d management.
* Opinionated: Instead of "How should I do this?", it proposes clear and specific development methods based on industry-proven best practices like layered structure, Conventional Commits, and K3d-based local environments. This reduces the fatigue of technical decision-making and supports rapid development speed.
* Cloud-Native Ready: Provides default support for API design using Protobuf, containerization via Docker, and local Kubernetes environments using K3d, enabling the configuration of projects optimized for development and deployment in modern cloud environments from the start.
* Efficiency: Seeks to maximize overall development productivity by eliminating inefficiencies and accelerating tasks at each stage of the development lifecycle (initial setup, coding, branch management, testing, deployment).

## Getting Started

### Prerequisites

* [Go 1.18+](https://go.dev/dl/)
* [Git](https://git-scm.com/downloads)
* [Docker](https://docs.docker.com/get-docker/)
* [K3d](https://k3d.io/#installation) (Optional)
* [Buf](https://buf.build/docs/installation) (Optional)

### Installation

'''bash
go install github.com/gosuda/ako@latest
'''

### Basic Usage (Examples)

1.  Create a new project:
    '''bash
    mkdir my-project && cd my-project
    ako init # or ako i
    '''
2.  Generate code:
    '''bash
    ako go lib # or ako g l (lib layer)
    ako go pkg # or ako g p (pkg layer, select template)
    ako go internal # or ako g n (internal layer, select template)
    ako go cmd # or ako g c (cmd layer)
    ako go buf # or ako g f (Generate Protobuf)
    '''
3.  Manage Git:
    '''bash
    ako branch create # or ako b c (Create branch)
    git add .
    ako branch commit # or ako b m (Commit)
    ako branch up # or ako b u (Move to parent branch)
    '''
4.  Run linter:
    '''bash
    ako linter # or ako l
    '''
5.  Manage K3d:
    '''bash
    ako k3d registry create my-reg # or ako k r c my-reg
    ako k3d cluster create my-clu --registry my-reg # or ako k c c my-clu --registry my-reg
    ako k3d manifest init # or ako k m i
    ako k3d manifest create # or ako k m c (Create manifest)
    ako k3d manifest build api-server # or ako k m b api-server (Build image)
    ako k3d manifest apply ./deployments/manifests/api-server/*.yaml # or ako k m a ...
    ako k3d manifest get pods # or ako k m g p
    '''

## Command Aliases

Here is a list of frequently used commands and their shortest aliases.

* `ako init` -> `ako i`
* `ako go lib` -> `ako g l`
* `ako go pkg` -> `ako g p`
* `ako go internal` -> `ako g n`
* `ako go cmd` -> `ako g c`
* `ako go buf` -> `ako g f`
* `ako branch current` -> `ako b n`
* `ako branch commit` -> `ako b m`
* `ako branch create` -> `ako b c`
* `ako branch up` -> `ako b u`
* `ako branch down` -> `ako b d`
* `ako linter` -> `ako l`
* `ako k3d registry list` -> `ako k r l` / `ls`
* `ako k3d registry create` -> `ako k r c`
* `ako k3d registry delete` -> `ako k r d` / `rm`
* `ako k3d cluster list` -> `ako k c l` / `ls`
* `ako k3d cluster create` -> `ako k c c`
* `ako k3d cluster delete` -> `ako k c d` / `rm`
* `ako k3d cluster append-port` -> `ako k c a` / `ap`
* `ako k3d manifest init` -> `ako k m i` / `f i`
* `ako k3d manifest create` -> `ako k m c` / `f c`
* `ako k3d manifest build` -> `ako k m b` / `f b` / `f d`
* `ako k3d manifest apply` -> `ako k m a` / `f a`
* `ako k3d manifest get pods` -> `ako k m g p` / `f g p` / `f g po`
* `ako k3d manifest get services` -> `ako k m g s` / `f g s` / `f g svc`
* `ako k3d manifest get deployments` -> `ako k m g d` / `f g d` / `f g deploy`
* `ako k3d manifest get ingress` -> `ako k m g i` / `f g i`

## Core Technologies

* Go: Core development language.
* urfave/cli/v3 & AlecAivazis/survey/v2: Building the CLI interface.
* Git: Version control and workflow automation based on Conventional Commits.
* Buf: Protobuf schema management and code generation.
* golangci-lint: Code static analysis.
* Docker: Containerization and image build automation.
* K3d: Local Kubernetes environment configuration automation.
* Uber Fx: Dependency injection framework (template-based).

## Supported Code Templates

Fx-based templates selectable when running `ako go internal` and `ako go pkg`.

### Internal Layer Templates (`ako g n`)

* `chi`: Handler based on go-chi/chi router.
* `fiber`: Handler based on gofiber/fiber framework.
* `grpc_server`: gRPC server implementation.
* `empty`: Includes only the basic Fx module structure.

### Pkg Layer Templates (`ako g p`)

* `cassandra`: Client using gocql/gocql.
* `clickhouse`: Client using ClickHouse/clickhouse-go.
* `duckdb`: Client using marcboeker/go-duckdb.
* `entgo`: Setup for ent/ent ORM.
* `grpc_client`: gRPC client implementation.
* `http_client`: Client based on standard `net/http` package.
* `kafka`: Producer/Consumer using IBM/sarama.
* `minio`: Client using minio/minio-go.
* `mssql`: Client using microsoft/go-mssqldb.
* `nats`: Client using nats-io/nats.go.
* `qdrant`: Client using qdrant/go-client.
* `redis`: Client using redis/rueidis.
* `sqlc`: Setup for sqlc-dev/sqlc (mainly for SQL DBs like PostgreSQL, MySQL).
* `templ`: Server-side rendering (SSR) component using a-h/templ.
* `valkey`: Client using valkey-io/valkey-go.
* `vault`: Client using hashicorp/vault/api.
* `empty`: Includes only the basic Fx module structure.

## License

MIT License. See the `LICENSE` file for details.
