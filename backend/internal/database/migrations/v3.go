package migrations

import (
	"github.com/jmoiron/sqlx"

	"context"
	"time"
)

// This creates the `state` table
type v3 struct{}

func init() {
	migrations = append(migrations, v3{})
}

func (v3) Version() uint64 {
	return 3
}

func (v3) Apply(
	ctx context.Context,
	tx *sqlx.Tx,
) error {
	if _, err := tx.ExecContext(
		ctx, tx.Rebind(`
      CREATE TABLE state (
        id INTEGER PRIMARY KEY,
        volume INTEGER NOT NULL,
        issue INTEGER NOT NULL,

        CHECK (id = 1)
      );
    `),
	); err != nil {
		return err
	}

	if _, err := tx.ExecContext(
		ctx, tx.Rebind(`
    INSERT INTO state (id, volume, issue)
    VALUES (1, 1, 1)
    ON CONFLICT DO NOTHING;
  `),
	); err != nil {
		return err
	}

	if _, err := tx.ExecContext(
		ctx, tx.Rebind(`
    INSERT INTO migrations (version, applied_at)
    VALUES (3, ?);
  `), time.Now().Unix(),
	); err != nil {
		return err
	}

	return tx.Commit()
}
