package database

import (
	"database/sql"
	"embed"
	"fmt"
	"net/url"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

type Config struct {
	DatabaseSourceName string
	MaxIdleConns       int
	ApplyDownMigration bool // Field to indicate whether to apply down migration
}

//go:embed migrations/*.sql
var Migrations embed.FS

type DB struct {
	*sql.DB
}

func New(cfg *Config) (*DB, error) {
	db, err := sql.Open("postgres", cfg.DatabaseSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxIdleConns(cfg.MaxIdleConns)

	return &DB{db}, nil
}

func NewWithMigrate(cfg *Config) (*DB, error) {
	db, err := New(cfg)
	if err != nil {
		return nil, err
	}

	if err := applyMigrations(cfg, db.DB); err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	return db, nil
}

func applyMigrations(cfg *Config, database *sql.DB) error {
	driver, err := postgres.WithInstance(database, &postgres.Config{})
	if err != nil {
		return err
	}

	mgtDriver, err := iofs.New(Migrations, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create iofs driver: %w", err)
	}

	// Extract database name from DSN
	u, err := url.Parse(cfg.DatabaseSourceName)
	if err != nil {
		return err
	}
	dbName := strings.TrimPrefix(u.Path, "/")

	m, err := migrate.NewWithInstance("iofs", mgtDriver, dbName, driver)
	if err != nil {
		return err
	}

	if cfg.ApplyDownMigration {
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			return err
		}
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
