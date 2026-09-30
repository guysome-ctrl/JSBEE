package migrations

import (
	"github.com/jmoiron/sqlx"

	"context"
	"time"
)

// This is the core migration which tracks
// all other migrations
type v0 struct{}

func (v0) Version() uint64 {
	return 0
}

func (v0) Apply(
	ctx context.Context,
	db *sqlx.DB,
) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(
		ctx, tx.Rebind(`
      CREATE TABLE IF NOT EXISTS migrations (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        version BIGINT UNIQUE NOT NULL,
        applied_at BIGINT NOT NULL
      );
    `),
	); err != nil {
		return err
	}

  if _, err := tx.ExecContext(
    ctx, tx.Rebind(`
      INSERT INTO migrations (version, applied_at)
      VALUES (0, ?)
      ON CONFLICT(version) DO NOTHING;
    `), time.Now().Unix(),
  ); err != nil {
    return err
  }

	return tx.Commit()
}
