# Go Module Generation Guide (with Fx)

This document provides a comprehensive guide to creating standardized, modular components using the `uber-go/fx` framework, following the architectural principles outlined in `project_structure.md`. A consistent module structure is essential for building a scalable and maintainable application where dependencies are managed explicitly and lifecycles are handled gracefully.

### Core Principles of a Module

1.  **Self-Contained**: Each module should be a self-contained unit of functionality. It defines its own configuration, dependencies, and constructor.
2.  **Explicit Dependencies**: All dependencies required by a module **must** be declared explicitly using the `fx.In` struct tag within a `Param` struct. This makes the dependency graph clear and statically analyzable.
3.  **Interface-Driven**: Modules that provide implementations (typically in `pkg`) **must** provide them as `lib` interfaces using `fx.As`. This decouples the business logic from concrete implementations.
4.  **Lifecycle Aware**: Modules that manage resources (like database connections or message queue consumers) **must** use `fx.Lifecycle` to register `OnStart` and `OnStop` hooks for proper initialization and graceful shutdown.

### Standard `fx.go` Template

Every component (`pkg`, `internal/service`, `internal/controller`) **must** have a corresponding `fx.go` file that defines its Fx module. This file acts as the single entry point for the component into the application's dependency injection graph.

Below is the standard, heavily commented template that **must** be followed.

```go
// The package name must match the directory name.
package postgres

import (
	"context"

	"go.uber.org/fx"
	// Import the LIB interface that this module implements.
	"your/project/lib/repository/user"
)

// Module exports the component's functionality to the Fx application.
// The module name (e.g., "postgres") should be descriptive and unique within the application.
var Module = fx.Module("postgres",
	// fx.Provide lists all the constructors this module offers to the DI container.
	fx.Provide(
		// The main constructor for the component.
		// fx.Annotate is used to explicitly associate the concrete implementation
		// with the interface it implements. This is a mandatory pattern.
		fx.Annotate(
			NewUserRepository, // The constructor function.
			// fx.As casts the concrete return type (*UserRepository) to one or more
			// interfaces (e.g., new(user.Repository)).
			fx.As(new(user.Repository)),
		),
		// The constructor for the module's configuration.
		ConfigRegister,
	),
)

// Config holds the configuration specific to this module.
// These values would typically be populated from a config file or environment variables.
type Config struct {
	DSN      string `yaml:"dsn"`
	PoolSize int    `yaml:"poolSize"`
}

// ConfigRegister is the standard constructor for the Config struct.
// In a real application, this function would contain logic to load and validate
// the configuration for this module.
func ConfigRegister() *Config {
	// For example: load from a config file.
	return &Config{}
}

// Param is a struct that groups all dependencies for the main constructor.
// The `fx.In` tag tells Fx to populate the fields of this struct.
// This keeps the constructor signature clean and manageable.
type Param struct {
	fx.In

	// Lifecycle is a mandatory dependency for any module that manages a resource.
	Lifecycle fx.Lifecycle
	// The module's specific configuration.
	Config *Config
	// Other dependencies, such as a logger, can be added here.
	// Logger client.Logger
}

// UserRepository is the concrete struct that implements the user.Repository interface.
type UserRepository struct {
	// e.g., db *sql.DB
}

// NewUserRepository is the constructor for the UserRepository.
// It receives all its dependencies via the Param struct, provided by Fx.
func NewUserRepository(p Param) (*UserRepository, error) {
	// The compile-time interface check is mandatory.
	var _ user.Repository = (*UserRepository)(nil)

	repo := &UserRepository{}

	// The Fx lifecycle is used to register startup and shutdown hooks.
	// This is mandatory for any resource that needs to be initialized or cleaned up.
	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Logic to run on application start.
			// e.g., connect to the database using p.Config.DSN
			// and assign the connection to repo.db.
			return nil
		},
		OnStop: func(ctx context.Context) error {
			// Logic to run on application stop.
			// e.g., close the database connection.
			return nil
		},
	})

	return repo, nil
}
```

### Module Composition in `main.go`

As described in `entry_point.md`, the `cmd/.../main.go` file is responsible for assembling all the application's modules. Fx will build the dependency graph, execute the constructors in the correct order, and run the application.

```go
package main

import (
	"go.uber.org/fx"

	// Import all necessary modules
	"your/project/pkg/client/postgres"
	"your/project/internal/service/user"
	"your/project/internal/controller/userapi"
	// ... and so on
)

func main() {
	fx.New(
		// List all modules here. Fx resolves the dependency order.
		postgres.Module,
		user.Module,
		userapi.Module,
	).Run()
}
```
