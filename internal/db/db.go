package db

import (
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func InitDB(dataSourceName string) (*sqlx.DB, error) {
	var db *sqlx.DB
	var err error

	retryIntervals := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

	for i := 0; i < len(retryIntervals)+1; i++ {
		db, err = sqlx.Connect("postgres", dataSourceName)
		if err == nil {
			break
		}

		if i < len(retryIntervals) {
			if isRetriableError(err) {
				time.Sleep(retryIntervals[i])
				continue
			}
			return nil, fmt.Errorf("connect to postgres: %w", err)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("connect to postgres after retries: %w", err)
	}

	err = runMigrations(db)

	if err != nil {
		return nil, err
	}

	return db, nil
}

//go:embed migrations/*.sql
var migrationsDir embed.FS

func runMigrations(db *sqlx.DB) error {
	d, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		return fmt.Errorf("failed to return an iofs driver: %w", err)
	}

	instance, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to get instance: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs:migrations", d, "postgres", instance)
	if err != nil {
		return fmt.Errorf("failed to get a new migrate instance: %w", err)
	}

	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations to the DB: %w", err)
		}
	}

	return nil
}

func isRetriableError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.SerializationFailure,
			pgerrcode.DeadlockDetected,
			pgerrcode.LockNotAvailable,
			pgerrcode.ConnectionException,
			pgerrcode.ConnectionDoesNotExist,
			pgerrcode.ConnectionFailure:
			return true
		}
	}
	return false
}
