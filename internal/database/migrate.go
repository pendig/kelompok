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

const migrationAdvisoryLockKey int64 = 4621001

type Migration struct {
	Version string
	Path    string
}

func Migrate(ctx context.Context, pool *pgxpool.Pool, dir string) (err error) {
	if dir == "" {
		dir = "migrations"
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire migration connection: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, `SELECT pg_advisory_lock($1)`, migrationAdvisoryLockKey); err != nil {
		return fmt.Errorf("acquire migration lock: %w", err)
	}
	defer func() {
		if _, unlockErr := conn.Exec(context.Background(), `SELECT pg_advisory_unlock($1)`, migrationAdvisoryLockKey); err == nil && unlockErr != nil {
			err = fmt.Errorf("release migration lock: %w", unlockErr)
		}
	}()

	if _, err := conn.Exec(ctx, `
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
		applied, err := isApplied(ctx, conn, migration.Version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		if err := applyMigration(ctx, conn, migration); err != nil {
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

		migrations = append(migrations, Migration{
			Version: entry.Name(),
			Path:    filepath.Join(dir, entry.Name()),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Path < migrations[j].Path
	})

	return migrations, nil
}

func isApplied(ctx context.Context, conn *pgxpool.Conn, version string) (bool, error) {
	var exists bool
	if err := conn.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`, version).Scan(&exists); err != nil {
		return false, fmt.Errorf("check migration %s: %w", version, err)
	}
	return exists, nil
}

func applyMigration(ctx context.Context, conn *pgxpool.Conn, migration Migration) error {
	sql, err := os.ReadFile(migration.Path)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", migration.Path, err)
	}

	tx, err := conn.Begin(ctx)
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
