# postgresql-connection

> PostgreSQL connection with embedded migrations via sqlx and golang-migrate

## app/shared/infrastructure/postgresql/connection.go

```go
package postgresql

import (
	"embed"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"archetype/app/shared/configuration"

	"github.com/Ignaciojeria/ioc"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib" // register pgx driver
	"github.com/jmoiron/sqlx"
)

var _ = ioc.Register(NewConnection)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// NewConnection creates a new PostgreSQL sqlx connection using the provided configuration.
// It automatically executes any pending migrations encoded in the migrationsFS embedded folder.
func NewConnection(env configuration.Conf) (*sqlx.DB, error) {

	dsn := env.DATABASE_URL
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}

	// 1️⃣ Conectar con el driver nativo puro pgx
	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	// 2️⃣ Extraer nombre de la base de datos para las migraciones
	u, err := url.Parse(dsn)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("invalid DATABASE_URL format: %w", err)
	}
	dbName := strings.TrimPrefix(u.Path, "/")

	// 3️⃣ Correr migraciones automáticamente
	if err := internalRunMigrations(db, dbName, migrationsFS); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

func internalRunMigrations(db *sqlx.DB, dbName string, fsys embed.FS) error {
	if db == nil {
		return fmt.Errorf("db connection is nil")
	}
	d, err := iofs.New(fsys, "migrations")
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(db.DB, &postgres.Config{
		DatabaseName: dbName,
	})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance(
		"iofs",
		d,
		dbName,
		driver,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	slog.Info("Database migrations validated/applied successfully")
	return nil
}

// Deprecated: use NewConnection which handles migrations internally.
// Function signature kept for backward compatibility if needed by generated code.
func runMigrations(db *sqlx.DB, dbName string) error {
	return internalRunMigrations(db, dbName, migrationsFS)
}
```

---

## Unit tests

When creating a new component, generate tests following this pattern:

### app/shared/infrastructure/postgresql/connection_test.go

```go
package postgresql

import (
	"context"
	"embed"
	"strings"
	"testing"
	"time"

	"archetype/app/shared/configuration"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestNewConnection_Success(t *testing.T) {
	ctx := context.Background()

	postgresContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %s", err)
	}

	defer func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}()

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %s", err)
	}

	conf := configuration.Conf{
		DATABASE_URL: connStr,
	}

	db, err := NewConnection(conf)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	err = db.Ping()
	assert.NoError(t, err)
	db.Close()
}

func TestNewConnection_InvalidURL(t *testing.T) {
	conf := configuration.Conf{
		DATABASE_URL: "invalid_url",
	}

	db, err := NewConnection(conf)
	if err == nil {
		t.Fatal("expected error connecting with invalid URL, got nil")
	}
	if db != nil {
		t.Errorf("expected nil db on error, got %v", db)
	}

	// Validate generic DSN parsing failure
	if !strings.Contains(err.Error(), "failed to connect") {
		t.Errorf("expected connection formatting error, got %v", err)
	}
}

func TestNewConnection_MalformedURL(t *testing.T) {
	// A URL that fails url.Parse
	conf := configuration.Conf{
		// Starting with a colon but no scheme often confuses parser
		DATABASE_URL: "postgres://user:pass@host:port/%-invalid",
	}

	db, err := NewConnection(conf)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if db != nil {
		t.Errorf("expected nil db, got %v", db)
	}
}

func TestNewConnection_EmptyURL(t *testing.T) {
	conf := configuration.Conf{
		DATABASE_URL: "",
	}

	db, err := NewConnection(conf)
	if err == nil {
		t.Fatal("expected error connecting with empty URL, got nil")
	}
	if db != nil {
		t.Errorf("expected nil db on error, got %v", db)
	}

	if !strings.Contains(err.Error(), "DATABASE_URL is not set") {
		t.Errorf("expected connection error, got %v", err)
	}
}

func TestInternalRunMigrations_Error(t *testing.T) {
	// Passing an empty embed.FS should trigger iofs.New error because "migrations" folder won't exist.
	// We need a valid DB connection (internalRunMigrations returns early if db is nil).
	ctx := context.Background()
	postgresContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %s", err)
	}
	defer func() {
		_ = postgresContainer.Terminate(ctx)
	}()

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %s", err)
	}

	db, err := sqlx.Connect("pgx", connStr)
	if err != nil {
		t.Fatalf("failed to connect: %s", err)
	}
	defer db.Close()

	err = internalRunMigrations(db, "testdb", embed.FS{})
	if err == nil {
		t.Fatal("expected error with empty embed.FS (iofs.New fails when migrations folder does not exist), got nil")
	}
}

func TestInternalRunMigrations_NilDB(t *testing.T) {
	// postgres.WithInstance(nil, ...) should fail
	err := internalRunMigrations(nil, "test", migrationsFS)
	if err == nil {
		t.Fatal("expected error with nil db, got nil")
	}
}

func TestRunMigrationsWrapper(t *testing.T) {
	// Test the public wrapper
	err := runMigrations(nil, "test")
	if err == nil {
		t.Fatal("expected error with nil db via wrapper, got nil")
	}
}
```
