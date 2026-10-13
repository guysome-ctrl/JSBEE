package migrations

import (
	"github.com/jmoiron/sqlx"

	"context"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

// Applies every migration at once on a fresh database and checks the
// final schema. There are no per-migration tests; at this scale one
// end-to-end test is enough.
func TestApplyAllMigrations(t *testing.T) {
	ctx := context.Background()

	db := sqlx.MustConnect("sqlite", "file::memory:")
	defer db.Close()

	if err := ApplyMigrations(ctx, db); err != nil {
		t.Fatalf("ApplyMigrations failed: %v", err)
	}

	var version uint64
	if err := db.GetContext(ctx, &version,
		`SELECT MAX(version) FROM migrations;`,
	); err != nil {
		t.Fatalf("Could not read migration version: %v", err)
	}
	if version != 8 {
		t.Fatalf("Expected version 8 after all migrations, got %v", version)
	}

	for _, table := range []string{
		"migrations", "users", "papers", "state", "blogs", "core_members",
	} {
		var name string
		if err := db.GetContext(ctx, &name,
			`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?;`,
			table,
		); err != nil || name != table {
			t.Fatalf("Expected table %q to exist, got %q, %v", table, name, err)
		}
	}

	// Core member photos are referenced by image_filename.
	var cols []struct {
		Name string `db:"name"`
	}
	if err := db.SelectContext(ctx, &cols,
		`SELECT name FROM pragma_table_info('core_members');`,
	); err != nil {
		t.Fatalf("Could not inspect core_members: %v", err)
	}
	names := make(map[string]bool, len(cols))
	for _, c := range cols {
		names[c.Name] = true
	}
	for _, col := range []string{
		"uuid", "name", "role", "description", "image_filename", "sort_order",
	} {
		if !names[col] {
			t.Fatalf("Expected core_members column %q, have %v", col, names)
		}
	}

	// Bios have no upper length cap: a very long description must insert fine.
	long := strings.Repeat("a", 20000)
	firstUUID := "11111111-1111-1111-1111-111111111111"
	firstImage := "core-11111111-1111-1111-1111-111111111111.jpg"
	secondUUID := "22222222-2222-2222-2222-222222222222"
	secondImage := "core-22222222-2222-2222-2222-222222222222.jpg"
	db.MustExecContext(ctx, `
    INSERT INTO core_members
    (uuid, name, role, description, image_filename, sort_order)
    VALUES (?, 'A', 'Role A', ?, ?, 0);
  `, firstUUID, long, firstImage)

	var gotDesc, gotImage string
	if err := db.QueryRowxContext(ctx,
		`SELECT description, image_filename
		FROM core_members WHERE uuid = ?;`,
		firstUUID,
	).Scan(&gotDesc, &gotImage); err != nil {
		t.Fatalf("Could not read core member back: %v", err)
	}
	if gotDesc != long || gotImage != firstImage {
		t.Fatal("Core member did not round-trip")
	}

	// Empty bios are still rejected, as are overlong names and duplicate roles.
	if _, err := db.ExecContext(ctx, `
    INSERT INTO core_members
    (uuid, name, role, description, image_filename, sort_order)
    VALUES (?, 'B', 'Role B', '', ?, 1);
  `, secondUUID, secondImage); err == nil {
		t.Fatal("Expected rejection for empty description")
	}
	if _, err := db.ExecContext(ctx, `
    INSERT INTO core_members
    (uuid, name, role, description, image_filename, sort_order)
    VALUES (?, ?, 'Role B', 'Desc', ?, 1);
  `, secondUUID, strings.Repeat("a", 121), secondImage); err == nil {
		t.Fatal("Expected rejection for 121-character name")
	}
	if _, err := db.ExecContext(ctx, `
    INSERT INTO core_members
    (uuid, name, role, description, image_filename, sort_order)
    VALUES (?, 'B', 'Role A', 'Desc', ?, 1);
  `, secondUUID, secondImage); err == nil {
		t.Fatal("Expected rejection for duplicate role")
	}
}
