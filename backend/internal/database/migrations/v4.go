package migrations

import (
	"github.com/jmoiron/sqlx"

	"context"
	"time"
)

// This adds the `reviewed` column to the `papers` table
type v4 struct{}

func init() {
	migrations = append(migrations, v4{})
}

func (v4) Version() uint64 {
	return 4
}

func (v4) Apply(
	ctx context.Context,
	tx *sqlx.Tx,
) error {
	if _, err := tx.ExecContext(
		ctx, tx.Rebind(`
      ALTER TABLE papers
      ADD COLUMN reviewed INTEGER NOT NULL DEFAULT 0
      CHECK (reviewed IN (0, 1));
    `),
	); err != nil {
		return err
	}

	if _, err := tx.ExecContext(
		ctx, tx.Rebind(`
      INSERT INTO migrations (version, applied_at)
      VALUES (4, ?);
    `), time.Now().Unix(),
	); err != nil {
		return err
	}

	return tx.Commit()
}
