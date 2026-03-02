# postgres-repository

> PostgreSQL adapter - implements ports/out.TemplateRepository

## app/adapter/out/postgres/template_repository.go

```go
package postgres

import (
	"context"

	"archetype/app/application/ports/out"
	"archetype/app/domain/entity"

	"github.com/Ignaciojeria/ioc"
	"github.com/jmoiron/sqlx"
)

var _ = ioc.Register(NewTemplateRepository)

type templateRepository struct {
	db *sqlx.DB
}

// NewTemplateRepository returns an implementation of ports/out.TemplateRepository.
func NewTemplateRepository(db *sqlx.DB) (out.TemplateRepository, error) {
	return &templateRepository{db: db}, nil
}

func (r *templateRepository) FindByID(ctx context.Context, id string) (*entity.Template, error) {
	var dest struct {
		ID string `db:"id"`
	}
	err := r.db.GetContext(ctx, &dest, "SELECT id FROM template_table WHERE id = $1 LIMIT 1", id)
	if err != nil {
		return nil, err
	}
	return &entity.Template{ID: dest.ID}, nil
}
```

---

## Unit tests

When creating a new component, generate tests following this pattern:

### app/adapter/out/postgres/template_repository_test.go

```go
package postgres

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestNewTemplateRepository(t *testing.T) {
	repo, err := NewTemplateRepository(nil)
	if err != nil {
		t.Fatalf("expected no error during repository creation, got %v", err)
	}
	if repo == nil {
		t.Fatal("expected repository instance, got nil")
	}
}

func TestTemplateRepository_FindByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open stub database connection: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo, _ := NewTemplateRepository(sqlxDB)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery("SELECT id FROM template_table WHERE id = \\$1 LIMIT 1").
			WithArgs("123").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("123"))

		result, err := repo.FindByID(ctx, "123")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result.ID != "123" {
			t.Errorf("expected ID 123, got %s", result.ID)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery("SELECT id FROM template_table WHERE id = \\$1 LIMIT 1").
			WithArgs("999").
			WillReturnError(sql.ErrNoRows)

		_, err := repo.FindByID(ctx, "999")
		if err == nil {
			t.Error("expected error sql.ErrNoRows, got nil")
		}
	})
}
```
