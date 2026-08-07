package storage

import (
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func newMigrate(pool *pgxpool.Pool) (*migrate.Migrate, error) {
	db := stdlib.OpenDBFromPool(pool)

	driver, err := pgx.WithInstance(db, &pgx.Config{})
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate instance: %w", err)
	}

	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		driver.Close()
		db.Close()
		return nil, fmt.Errorf("migrate source: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "pgx", driver)
	if err != nil {
		src.Close()
		driver.Close()
		db.Close()
		return nil, fmt.Errorf("migrate init: %w", err)
	}

	return m, nil
}

// MigrateUp applies all pending migrations.
func MigrateUp(pool *pgxpool.Pool) error {
	m, err := newMigrate(pool)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}

// MigrateDown reverts all migrations. Useful for integration-test teardown.
func MigrateDown(pool *pgxpool.Pool) error {
	m, err := newMigrate(pool)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate down: %w", err)
	}
	return nil
}
