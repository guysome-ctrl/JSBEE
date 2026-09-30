package migrations

import (
	"github.com/jmoiron/sqlx"

	"context"
	"time"
)

// This migration adds the `users` table
// and creates an index on it
type v1 struct{}

func init() {
	migrations = append(migrations, v1{})
}

func (v1) Version() uint64 {
	return 1
}

func (v1) Apply(
	ctx context.Context,
	tx *sqlx.Tx,
) error {
	if _, err := tx.ExecContext(
		ctx, tx.Rebind(`
      CREATE TABLE users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        uuid TEXT NOT NULL UNIQUE CHECK (uuid != ''),
        name TEXT NOT NULL CHECK (name != ''),
        email TEXT NOT NULL UNIQUE CHECK (email LIKE '%@%'),
        subscribed INTEGER NOT NULL DEFAULT 0 CHECK (subscribed IN (0, 1)),
        role TEXT NOT NULL DEFAULT 'viewer' CHECK (role IN (
          'owner',
          'admin',
          'reviewer',
          'viewer'
        ))
      );
    `),
	); err != nil {
		return err
	}

	if _, err := tx.ExecContext(
		ctx, tx.Rebind(`
      CREATE INDEX idx_users
      ON users(name, role);
    `),
	); err != nil {
		return err
	}

	if _, err := tx.ExecContext(
		ctx, tx.Rebind(`
      INSERT INTO migrations (version, applied_at)
      VALUES (1, ?);
    `), time.Now().Unix(),
	); err != nil {
		return err
	}

	return tx.Commit()
}
