# Templ Renderer Fx Module Generation Guide

This guide is for generating a standardized Fx module for a `templ` renderer. This allows `templ` components to be used for server-side rendering within an HTTP controller (like Chi or Fiber).

**File to Create**: `pkg/renderer/templ/fx.go`

**LLM Prompt**:
"Create an Fx module for a `templ` renderer. The package name is `templ`. It should implement a `lib/renderer.Renderer` interface. This module will allow `templ` components to be rendered in HTTP handlers."

---

### `lib/renderer/renderer.go` (Interface Definition)

First, define a generic interface for a renderer in the `lib` layer.

```go
package renderer

import (
	"context"
	"io"

	"github.com/a-h/templ"
)

// Renderer defines a standard interface for rendering components.
type Renderer interface {
	Render(ctx context.Context, w io.Writer, component templ.Component) error
}
```

### Generated `fx.go`

```go
package templ

import (
	"context"
	"io"

	"github.com/a-h/templ"
	"go.uber.org/fx"

	"your/project/lib/renderer"
)

// Module exports the templ renderer component to the Fx application.
var Module = fx.Module("templ-renderer",
	fx.Provide(
		fx.Annotate(
			NewRenderer,
			fx.As(new(renderer.Renderer)),
		),
	),
)

// templRenderer is the concrete implementation of the renderer.Renderer interface.
type templRenderer struct{}

// NewRenderer creates a new templ renderer.
// This component is stateless, so it has no dependencies or config.
func NewRenderer() (renderer.Renderer, error) {
	// Compile-time interface check.
	var _ renderer.Renderer = (*templRenderer)(nil)
	return &templRenderer{}, nil
}

// Render executes the given templ component and writes the output.
func (r *templRenderer) Render(ctx context.Context, w io.Writer, component templ.Component) error {
	return component.Render(ctx, w)
}
```

### Usage in an HTTP Handler (Example)

```go
// In your internal/controller/userapi/controller.go

// ...
import (
	"your/project/lib/renderer"
	// your generated templ components
	// "your/project/internal/controller/userapi/view"
)

// ...

func makeGetUserPageHandler(renderer renderer.Renderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// ... fetch user data
		// component := view.Profile(user)
		// renderer.Render(r.Context(), w, component)
	}
}
```
