# architecture-guidelines

> Hexagonal Architecture with ports: in (what the app exposes) and out (what the app needs)

## Ports structure

Ports live in `application/ports/`, split by direction:

```
app/application/ports/
├── in/              <- Input ports: use case executors (what the app exposes to controllers)
│   └── get_template.go
└── out/             <- Output ports: repositories, publishers (what the app needs from adapters)
    ├── template_repository.go
    └── event_publisher.go
```

**Input ports (in):** Interfaces that controllers call. One executor per use case.

**Output ports (out):** Interfaces that the use case needs. Implementations live in `adapter/out` (Postgres, NATS, etc.).

**Domain:** `domain/entity/` only. No interfaces in the domain—they live in ports.

Why:
- **DIP:** Usecase imports `ports/out` only. It never knows about Postgres or eventbus.
- **Testability:** Mock `ports/out` in usecase tests. No DB drivers.
- **Enterprise-ready:** Recognizable hexagonal layout. Scalable and standard.

## Example flow

```go
// application/ports/out/template_repository.go
package out

import ("context"; "archetype/app/domain/entity")

type TemplateRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Template, error)
}
```

```go
// application/ports/in/get_template.go
package in

import "context"

type GetTemplateExecutor interface {
	Execute(ctx context.Context, id string) (GetTemplateOutput, error)
}

type GetTemplateOutput struct { ID string; Name string }
```

```go
// application/usecase/get_template.go - imports ports/out, NOT postgres
package usecase

import ("context"; "archetype/app/application/ports/in"; "archetype/app/application/ports/out"; "github.com/Ignaciojeria/ioc")

func NewGetTemplateUseCase(repo out.TemplateRepository) (in.GetTemplateExecutor, error) {
	return &getTemplateUseCase{repo: repo}, nil
}
```

```go
// adapter/out/postgres/template_repository.go - implements ports/out
package postgres

import ("archetype/app/application/ports/out"; "archetype/app/domain/entity"; ...)

func NewTemplateRepository(db *sqlx.DB) (out.TemplateRepository, error) { ... }
```

```go
// adapter/in/fuegoapi/get_template.go - injects ports/in
func NewGetTemplate(s *httpserver.Server, uc in.GetTemplateExecutor) { ... }
```

Flow: `ports/out` (interface) → usecase implements logic → `ports/in` (executor) → controller calls it. Adapters implement `ports/out`.

## Data mapping: the firewall rule

To keep the Domain pure, we strictly separate **Infrastructure Models** (how data is stored/received) from **Domain Entities** (how the business thinks).

**The rule:**
- Domain entities must have **zero tags** (`db`, `json`, etc.) and no external dependencies.
- Adapters define their own **local structs** (models).
- **Mappers** live inside the adapter and translate between adapter model and domain entity.

### Output adapter (Repository)

```go
// adapter/out/postgres/template_repository.go

// 1. Local struct with infrastructure tags
type templateDB struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
}

// 2. Mapper (private function inside the adapter)
func toDomain(m templateDB) *entity.Template {
	return &entity.Template{
		ID:   m.ID,
		Name: m.Name,
	}
}

// 3. Reverse mapper for Save (entity → DB model)
func fromDomain(e *entity.Template) templateDB {
	return templateDB{ID: e.ID, Name: e.Name, CreatedAt: time.Now()}
}

func (r *templateRepository) FindByID(ctx context.Context, id string) (*entity.Template, error) {
	var m templateDB
	err := r.db.GetContext(ctx, &m, "SELECT * FROM templates WHERE id=$1", id)
	if err != nil {
		return nil, err
	}
	return toDomain(m), nil
}
```

The port always receives and returns domain entities. The adapter hides DB details.

### Input adapter (Controller)

The controller receives HTTP/event data and calls the use case with a DTO (`XxxInput`). The port returns a DTO (`XxxOutput`). The controller passes that DTO to the client; no extra mapping needed because the port already exposes DTOs, not entities.

### Why this matters (for AI and humans)

| Aspect | Impact |
|--------|--------|
| **Encapsulation** | No `sql.NullString` or DB types leak into the domain. |
| **Predictability** | Every adapter: Fetch → Map → Return (or Input → Map → Insert). |
| **Refactoring** | DB column rename? Change `templateDB` and `toDomain`/`fromDomain` only. Rest of app unchanged. |

### Mapper vibe check

- **Where do mappers live?** Always inside the adapter file. No global `mappers/` folder—mapping is a private adapter concern.
- **Domain → Infra?** Yes. On `Save`, the adapter receives `entity.Template` and maps to `templateDB` before the INSERT.
- **Infra → Domain?** Yes. On `FindByID`, map `templateDB` to `entity.Template` before returning.

## Always inject interfaces (one implementation per interface)

**Always inject and return interfaces.** Each interface must have **exactly one implementation** in the IoC. If an interface has N implementations, the container cannot resolve unambiguously.

## Use case output: DTOs over domain entities

**Prefer use case-specific DTOs** (`XxxInput`, `XxxOutput`) in the port. Protects the API; if the domain changes, external JSON stays stable.

## Summary

| Avoid | Use |
|-------|-----|
| Interfaces scattered (domain/repository, usecase, etc.) | Input ports in `ports/in/`; output ports in `ports/out/` |
| Usecase importing postgres/eventbus (breaks DIP) | Usecase imports `ports/out`; adapter implements and imports ports |
| Domain entities with `db:"id"` or infra types | Adapter defines local model + mapper; port sees only domain entities |
| Injecting concrete types | Always inject and return interfaces |
| One interface with N implementations | One implementation per interface; use factory if needed |
| Returning domain entities from use cases | DTOs in the port (`XxxOutput`) |

---

## Vibe check

| Rule | Why it matters |
|------|----------------|
| Output ports in `ports/out/` | Several use cases share the same contract. Adapters (Postgres, NATS) implement them. |
| Input ports in `ports/in/` | Controllers import one file to know what to call. Clear entry points. |
| DTOs in ports | Protects the API. Domain changes don't break external JSON. |
| One implementation per interface | Keeps einar-ioc predictable and startup deterministic. |
| Domain = entities only | Domain stays pure. No infrastructure knowledge. |
| Mappers inside adapter | Local model + `toDomain`/`fromDomain`; no global mappers folder. |
