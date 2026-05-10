package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Migration struct {
	Version string
	Path    string
}

func Migrate(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	if dir == "" {
		dir = "migrations"
	}

	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version text PRIMARY KEY,
			applied_at timestamptz NOT NULL DEFAULT now()
		)
	`); err != nil {
		return fmt.Errorf("ensure schema_migrations: %w", err)
	}

	migrations, err := listMigrations(dir)
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		applied, err := isApplied(ctx, pool, migration.Version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		if err := applyMigration(ctx, pool, migration); err != nil {
			return err
		}
	}

	return nil
}

func listMigrations(dir string) ([]Migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}

	migrations := make([]Migration, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		version := strings.TrimSuffix(entry.Name(), ".sql")
		if cut := strings.Index(version, "_"); cut > 0 {
			version = version[:cut]
		}

		migrations = append(migrations, Migration{
			Version: version,
			Path:    filepath.Join(dir, entry.Name()),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Path < migrations[j].Path
	})

	return migrations, nil
}

func isApplied(ctx context.Context, pool *pgxpool.Pool, version string) (bool, error) {
	var exists bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`, version).Scan(&exists); err != nil {
		return false, fmt.Errorf("check migration %s: %w", version, err)
	}
	return exists, nil
}

func applyMigration(ctx context.Context, pool *pgxpool.Pool, migration Migration) error {
	sql, err := os.ReadFile(migration.Path)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", migration.Path, err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", migration.Version, err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, string(sql)); err != nil {
		return fmt.Errorf("apply migration %s: %w", migration.Version, err)
	}

	if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, migration.Version); err != nil {
		return fmt.Errorf("record migration %s: %w", migration.Version, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit migration %s: %w", migration.Version, err)
	}

	return nil
}
