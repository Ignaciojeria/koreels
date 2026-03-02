# ports

> Hexagonal ports - in (executors) and out (repositories, publishers)

## app/application/ports/in/get_template.go

```go
package in

import "context"

// GetTemplateExecutor is the input port for the GetTemplate use case.
// Controllers (Fuego, gRPC) call this; implementations live in usecase/.
type GetTemplateExecutor interface {
	Execute(ctx context.Context, id string) (GetTemplateOutput, error)
}

// GetTemplateOutput is the DTO returned by the use case.
type GetTemplateOutput struct {
	ID   string
	Name string
}
```

---

## app/application/ports/out/template_repository.go

```go
package out

import (
	"context"

	"archetype/app/domain/entity"
)

// TemplateRepository defines the persistence contract for Template.
// Implementations (Postgres, MongoDB, etc.) live in adapter/out.
type TemplateRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Template, error)
}
```

---

## app/application/ports/out/event_publisher.go

```go
package out

import "context"

// Event is a domain event to be published.
// Implementations convert it to the broker format (e.g. CloudEvents) in the adapter.
type Event interface {
	EventType() string
}

// DomainEventPublisher publishes domain events to a broker.
// Implementations (NATS, GCP Pub/Sub, Kafka) live in adapter/out.
type DomainEventPublisher interface {
	Publish(ctx context.Context, e Event) error
}
```
