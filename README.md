# ako: Opinionated Go Project Manager

[![Go Report Card](https://goreportcard.com/badge/github.com/gosuda/ako)](https://goreportcard.com/report/github.com/gosuda/ako)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

`ako` is a CLI tool for enhancing the productivity and standardization of Go projects. It aims to efficiently manage a monorepo environment where multiple related services are managed in a single repository, and these services within a single namespace on Kubernetes. It automates repetitive setup, code structuring, Git management, and local K3d environment configuration, helping developers focus on core logic.

## Problem to Solve

Starting and continuously managing Go projects often involves the following inefficiencies and difficulties:

1.  Complex and time-consuming initial setup:
    * Every new project requires basic tasks like creating a Go module, initializing a Git repository, and setting up a basic `.gitignore` file.
    * Configuring CI/CD pipelines (e.g., GitHub Actions workflows), setting up linters (`golangci-lint`) for code quality management, and configuring `buf` for Protobuf usage require knowledge, are cumbersome, and prone to errors.
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
    * Automatically configures many elements needed to start a project with a single command execution: Go module, Git repository (including `.gitignore`), selectable CI/CD templates, `buf` setup and examples, Uber Fx dependency, Dev Container setup, `golangci-lint` setup and binary installation, Conventional Commits rules setup, and default `release` branch creation.
    * This frees developers from the complex initial setup process, allowing them to focus on core code development immediately after project creation.

2.  Standardized layer structure proposal and code generation (`ako go` / `ako g`):
    * `ako` proposes a layer structure considering separation of concerns and unidirectional dependency principles, and automates code generation for each layer. This can help improve project maintainability, scalability, and testability.
    * `lib/` (Core Abstraction Layer):
        * Role: Aims to handle the project's fundamental abstractions. Can be used to define technology-agnostic interfaces (e.g., Repository, Domain Validator, External Adapter) and supporting core data structures (Value Object, Entity, etc.). Data Transfer Objects (DTOs) used for inter-layer communication can also be considered for definition here, along with related interfaces in sub-packages (e.g., `lib/repository/user`).
        * Characteristics: Generally contains pure definitions without concrete implementation code. It's recommended to minimize dependencies on other internal packages (`internal`, `pkg`, `cmd`) and external libraries.
        * `ako go lib` (`ako g l`): Helps generate interface or basic data structure files needed for this layer.
    * `pkg/` (Implementation Layer):
        * Role: Can contain concrete implementations of interfaces defined in `lib/`. May include code related to infrastructure and external libraries, such as actual database interaction logic, external API call logic, or specific algorithm implementations.
        * Characteristics: Implements interfaces from `lib/` and can depend on `lib/` to use other `lib/` definitions if necessary. However, it's advisable not to depend on `internal` or `cmd`. Instead of directly depending on other implementation packages within `pkg`, the recommended approach is to receive necessary dependencies via injection in `cmd`. It's common to organize subdirectories based on the implementation method (e.g., `postgres`, `redis`, `kafka`).
        * `ako go pkg` (`ako g p`): Supports generating Fx module-based implementation templates for specific tech stacks (e.g., `redis`, `sqlc`), reducing repetitive setup and coding.
    * `internal/` (Business Logic Composition Layer):
        * Role: The area for composing the application's core business logic. Handles tasks like external request processing, business rule application, and data processing flow control. It's suggested to manage this by dividing into `controller` and `service` subdirectories.
        * Characteristics: Uses interfaces defined in `lib/` to compose the logic flow. It's advisable not to depend directly on `pkg/` or `cmd/`. Due to Go's `internal` directory nature, it cannot be directly imported by external projects.
        * `internal/controller/`: Acts as the entry point for handling external requests (HTTP, gRPC, etc.), calling appropriate `service`s, and transforming results into a format understandable by external systems. Can depend on `internal/service` and `lib`.
        * `internal/service/`: Performs core business logic and orchestrates the flow of use cases by combining multiple `lib/` interfaces (mainly repository, domain). Recommended to depend only on the `lib/` package.
        * `ako go internal` (`ako g n`): Supports generating Fx module-based templates (e.g., `chi`, `fiber`, `grpc_server`) for `controller` or `service` roles, helping focus on implementing business logic.
    * `cmd/` (Execution and Assembly Layer):
        * Role: The application's execution entry point (`main` package). Responsible for assembling (wiring) implementations from each layer and running the application.
        * Characteristics: Can perform configuration loading, flag parsing, initialization of `pkg/` implementations, initialization of `internal/service` and `internal/controller` components, and dependency injection (DI) using Uber Fx. It's best to focus solely on configuration, assembly, and execution, without including business logic. Can depend on all internal layers like `internal`, `pkg`, and `lib`.
        * `ako go cmd` (`ako g c`): Automates the generation of the basic structure and Dockerfile for new executables (e.g., API server, batch worker).
    * Protobuf Management:
        * `proto/`: Manages IDL source files like Protocol Buffers.
        * `lib/gen/`: Locates Go code automatically generated from `proto/` files.
        * `ako go buf` (`ako g f`): Simplifies running the `buf generate` command for easy IDL-based code generation.
    * This structure and automation tools allow developers to efficiently build robust and flexible applications based on clear separation of concerns and dependency management.

3.  Systematic Git workflow and branch strategy support (`ako branch` / `ako b`):
    * `ako` proposes a hierarchical branch strategy for Git management and automates related tasks.
    * Proposed Main Branch Hierarchy:
        * `release`: Manages deployable production code. (Created by `ako init`)
        * `staging`: For testing release candidates. (Can branch from `release`)
        * `develop`: Integrates the latest development code for the next release. (Can branch from `staging`)
        * `epic/{epic-name}`: Manages large feature units. (Can branch from `develop`)
        * `feature/{epic-name}/{feature-name}`: For developing new features while maintaining backward compatibility. (Can branch from `epic`)
        * `patch/{epic-name}/{patch-name}`: For fixing bugs while maintaining backward compatibility. (Can branch from `epic`)
        * `break/{epic-name}/{break-name}`: For developing changes that might break backward compatibility. (Can branch from `epic`)
        * `proposal/{feature|patch|break-name}/{proposal-name}`: Temporary branch for experimental ideas or tasks requiring discussion. (Can branch from `feature`, `patch`, `break`)
        * `hotfix/*`: For urgent bug fixes in the production environment. (Can branch from `release`, not directly created by `ako branch create`)
    * Branch Creation Automation:
        * The `ako branch create` (`ako b c`) command interactively creates branches of allowed subtypes based on the current branch. For example, if the current branch is `epic/user-auth`, you can create `feature/user-auth/login`, `patch/user-auth/validation-fix`, etc.
        * Branch names are structured as `type/parent-scope/task-name` or `type/task-name`.
    * Hierarchical Navigation:
        * `ako branch up` (`ako b u`) and `ako branch down` (`ako b d`) commands allow easy navigation between defined parent or child branches of the current branch, facilitating exploration even in complex branch structures.
    * Conventional Commits Support:
        * `ako branch commit` (`ako b m`) supports writing commit messages following the Conventional Commits convention through an interactive prompt.
        * This can improve commit history readability, clarify the intent of changes, and serve as a basis for automated version management and changelog generation.
    * This strategy and automation tools can help teams manage branches consistently, track code change history effectively, and build stable development and release pipelines.

4.  Easy code quality checks (`ako linter` / `ako l`):
    * A single `ako linter` (`ako l`) command runs `golangci-lint` across the entire project, helping to detect code style issues early and maintain consistent code quality.
    * You can customize linting rules by modifying the `.golangcilint.yaml` file created by `ako init` to enable/disable rules or change settings according to project needs. Refer to the [official golangci-lint documentation](https://golangci-lint.run/usage/linters/) for available linters and configuration options.

5.  Simplified local K3d environment management (`ako k3d` / `ako k`):
    * `ako` provides a workflow for setting up and managing local Kubernetes development environments, useful for deploying and managing multiple services within a single namespace in a monorepo environment.
    * K3d Cluster and Registry Management:
        * `ako k3d cluster create/delete/list` (`ako k c c/d/l`): Easily create, delete, and list K3d clusters. You can specify a local registry to use during cluster creation.
        * `ako k3d registry create/delete/list` (`ako k r c/d/l`): Create, delete, and list local Docker registries within the K3d environment for development purposes.
        * `ako k3d cluster append-port` (`ako k c a`): Add port forwarding rules to the load balancer of a created cluster.
    * Kubernetes Manifest Management Workflow:
        * Initialization (`ako k3d manifest init` / `ako k m i`):
            * Select the target K3d cluster and the local registry to use.
            * Input the Kubernetes namespace where applications will be deployed. (Services within the monorepo will share this single namespace).
            * (Optional) Input the address of a remote registry for environments like production.
            * Save the entered information (cluster, namespace, local/remote registry) to the project configuration file (`manifests/k3d_config.yaml`).
            * Generate default manifest files for the specified namespace (`namespace.yaml`) and basic ingress manifests for public/private access (`ingress-public.yaml`, `ingress-private.yaml`).
        * Creation (`ako k3d manifest create` / `ako k m c`):
            * Select one of the applications (executables) defined under the `cmd/` directory.
            * Choose the type of manifest to create (Deployment or CronJob).
            * Based on the selected application path (e.g., `cmd/api/auth`), create a corresponding directory structure under `deployments/manifests/` (e.g., `deployments/manifests/api/auth`) and automatically generate Kubernetes manifest files (Deployment/CronJob, Service, ConfigMap, etc.) for that application. The namespace uses the value set during the `init` step, ensuring all services are deployed to the same namespace.
        * Build (`ako k3d manifest build` / `ako k m b`):
            * Select an application from under `cmd/`.
            * Build a Docker image using the application's Dockerfile. (Builds utilizing common `lib/`, `pkg/` code within the monorepo).
            * Push the built image to the local K3d registry configured during the `init` step. The image tag is generated including the local registry address (e.g., `k3d-my-registry.localhost:5000/api-server:latest`).
            * This image can be referenced by manifests within the local K3d cluster.
        * Apply (`ako k3d manifest apply` / `ako k m a`):
            * Display a list of manifest files generated under `deployments/manifests/` and allow the user to select one or more files for deployment. (Manifests for multiple services can be selected together for deployment into the single namespace).
            * Sequentially execute the `kubectl apply -f <filepath>` command for each selected manifest file to create or update resources in the K3d cluster.
        * Get (`ako k3d manifest get` / `ako k m g`):
            * Select frequently checked Kubernetes resource types like `pods`, `services`, `deployments`, `ingress` to easily run the `kubectl get <resource>` command and view the results. (Checks all resources within the single namespace).
    * This workflow allows developers, even without deep knowledge of complex `kubectl` commands or manifest file structures, to easily build, deploy, and test containerized applications within a single namespace in a local K3d environment while managing multiple services in a monorepo.

## `ako`'s Philosophy

`ako` is designed based on the following core philosophies:

* Standardization: Aims to increase code consistency, facilitate collaboration within and between teams, and contribute to reducing long-term maintenance costs by presenting a well-defined project structure and development workflow. This can reduce the cognitive load on developers.
* Developer Experience: Helps developers focus more time and energy on core logic development that creates actual business value by automating cumbersome and error-prone tasks like project initialization, repetitive code writing, and complex Git and K3d management.
* Opinionated: Instead of "How should I do this?", following the development methods proposed by `ako`, such as layered structure, Conventional Commits, and K3d-based local environments, can help reduce the burden of technical decision-making and increase development speed.
* Cloud-Native Ready: Provides default support for API design using Protobuf, containerization via Docker, and local Kubernetes environments using K3d, supporting the configuration of projects suitable for development and deployment in modern cloud environments from the start.
* Efficiency: Seeks to maximize overall development productivity by reducing potential inefficiencies and accelerating tasks at each stage of the development lifecycle (initial setup, coding, branch management, testing, deployment).

## Getting Started

### Prerequisites

* [Go 1.24+](https://go.dev/dl/)
* [Git](https://git-scm.com/downloads)
* [Docker](https://docs.docker.com/get-docker/)
* [K3d](https://k3d.io/#installation) (Optional)
* [Buf](https://buf.build/docs/installation) (Embedded)

### Installation

```bash
go install [github.com/gosuda/ako@latest](https://github.com/gosuda/ako@latest)
```

### Basic Usage (Examples)

1.  Create a new project:
    ```bash
    mkdir my-project && cd my-project
    ako init # or ako i
    ```
2.  Generate code:
    ```bash
    ako go lib # or ako g l (lib layer)
    ako go pkg # or ako g p (pkg layer, select template)
    ako go internal # or ako g n (internal layer, select template)
    ako go cmd # or ako g c (cmd layer)
    ako go buf # or ako g f (Generate Protobuf)
    ```
3.  Manage Git:
    ```bash
    ako branch create # or ako b c (Create branch)
    git add .
    ako branch commit # or ako b m (Commit)
    ako branch up # or ako b u (Move to parent branch)
    ```
4.  Run linter:
    ```bash
    ako linter # or ako l
    ```
5.  Manage K3d:
    ```bash
    ako k3d registry create my-reg # or ako k r c my-reg
    ako k3d cluster create my-clu --registry my-reg # or ako k c c my-clu --registry my-reg
    ako k3d manifest init # or ako k m i
    ako k3d manifest create # or ako k m c (Create manifest)
    ako k3d manifest build api-server # or ako k m b api-server (Build image)
    ako k3d manifest apply ./deployments/manifests/api-server/*.yaml # or ako k m a ...
    ako k3d manifest get pods # or ako k m g p
    ```

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
* `zap`: Logging using uber-go/zap.
* `zerolog`: Logging using github.com/rs/zerolog.
* `slog`: Logging using log/slog.
* `meilisearch`: Client using meilisearch/meilisearch-go.
* `opensearch`: Client using opensearch-project/opensearch-go.
* `elasticsearch`: Client using elastic/go-elasticsearch/v9.
* `empty`: Includes only the basic Fx module structure.

## License

MIT License. See the `LICENSE` file for details.
