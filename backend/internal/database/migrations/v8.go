package migrations

import (
	"github.com/jmoiron/sqlx"

	"context"
	"time"
)

// This creates the `core_members` table. Member text (name, role,
// description) lives here; member photos live in the public bucket
// of the object store and are referenced by `image_filename`.
type v8 struct{}

func init() {
	migrations = append(migrations, v8{})
}

func (v8) Version() uint64 {
	return 8
}

func (v8) Apply(
	ctx context.Context,
	tx *sqlx.Tx,
) error {
	if _, err := tx.ExecContext(
		ctx, tx.Rebind(`
      CREATE TABLE core_members (
        uuid TEXT PRIMARY KEY CHECK (uuid != ''),
        name TEXT NOT NULL CHECK (length(name) > 0 AND length(name) <= 120),
        role TEXT NOT NULL UNIQUE
        CHECK (length(role) > 0 AND length(role) <= 120),
        description TEXT NOT NULL CHECK (length(description) > 0),
        image_filename TEXT NOT NULL UNIQUE CHECK (image_filename != ''),
        sort_order INTEGER NOT NULL
      );
    `),
	); err != nil {
		return err
	}

	if _, err := tx.ExecContext(
		ctx, tx.Rebind(`
      CREATE INDEX idx_core_members_order
      ON core_members(sort_order);
    `),
	); err != nil {
		return err
	}

	if _, err := tx.ExecContext(
		ctx, tx.Rebind(`
      INSERT INTO migrations (version, applied_at)
      VALUES (8, ?);
    `), time.Now().Unix(),
	); err != nil {
		return err
	}

	return tx.Commit()
}
