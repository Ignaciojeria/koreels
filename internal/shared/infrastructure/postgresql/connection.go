package postgresql

import (
	"embed"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"koreels/internal/shared/configuration"

	"github.com/Ignaciojeria/ioc"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

var _ = ioc.Register(NewConnection)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func NewConnection(env configuration.Conf) (*sqlx.DB, error) {
	dsn := env.DATABASE_URL
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}
	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}
	u, err := url.Parse(dsn)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("invalid DATABASE_URL format: %w", err)
	}
	dbName := strings.TrimPrefix(u.Path, "/")
	if err := internalRunMigrations(db, dbName, migrationsFS); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}
	slog.Info("Database migrations validated/applied successfully")
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
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{DatabaseName: dbName})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithInstance("iofs", d, dbName, driver)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
