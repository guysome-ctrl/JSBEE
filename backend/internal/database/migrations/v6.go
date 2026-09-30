package migrations

import (
	"github.com/jmoiron/sqlx"

	"context"
	"time"
)

// This adds the `city_lead` column to the `users` table
type v6 struct{}

func init() {
	migrations = append(migrations, v6{})
}

func (v6) Version() uint64 {
	return 6
}

func (v6) Apply(
	ctx context.Context,
	tx *sqlx.Tx,
) error {
	if _, err := tx.ExecContext(
		ctx, tx.Rebind(`
      ALTER TABLE users
      ADD COLUMN city_lead TEXT DEFAULT NULL;
    `),
	); err != nil {
		return err
	}

	if _, err := tx.ExecContext(
		ctx, tx.Rebind(`
      INSERT INTO migrations (version, applied_at)
      VALUES (6, ?);
    `), time.Now().Unix(),
	); err != nil {
		return err
	}

	return tx.Commit()
}
