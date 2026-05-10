package database

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListMigrationsPreservesFullFilenameVersion(t *testing.T) {
	dir := t.TempDir()
	files := []string{
		"000002_add.sql",
		"000002_fix.sql",
		"000001_init.sql",
	}

	for _, file := range files {
		if err := os.WriteFile(filepath.Join(dir, file), []byte("-- migration\n"), 0o644); err != nil {
			t.Fatalf("write migration %s: %v", file, err)
		}
	}

	migrations, err := listMigrations(os.DirFS(dir), ".")
	if err != nil {
		t.Fatalf("list migrations: %v", err)
	}

	got := make([]string, 0, len(migrations))
	for _, migration := range migrations {
		got = append(got, migration.Version)
	}

	want := []string{
		"000001_init.sql",
		"000002_add.sql",
		"000002_fix.sql",
	}

	if len(got) != len(want) {
		t.Fatalf("got %d migrations, want %d: %v", len(got), len(want), got)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("version[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
