package migrations

import (
	"github.com/jmoiron/sqlx"

	"context"
	"time"
)

// This creates the `papers` table
type v2 struct{}

func init() {
	migrations = append(migrations, v2{})
}

func (v2) Version() uint64 {
	return 2
}

func (v2) Apply(
	ctx context.Context,
	tx *sqlx.Tx,
) error {
	if _, err := tx.ExecContext(
		ctx, tx.Rebind(`
      CREATE TABLE papers (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        uuid TEXT NOT NULL UNIQUE CHECK (uuid != ''),
        title TEXT NOT NULL CHECK (title != ''),
        number INTEGER,
        filename TEXT NOT NULL UNIQUE CHECK (filename != ''),
        volume INTEGER CHECK (volume > 0),
        issue INTEGER CHECK (issue > 0),

        owner_uuid TEXT REFERENCES users(uuid) ON DELETE SET NULL,

        UNIQUE(title, owner_uuid),
        UNIQUE(volume, number, issue)
      );
    `),
	); err != nil {
		return err
	}

	if _, err := tx.ExecContext(
		ctx, tx.Rebind(`
      INSERT INTO migrations (version, applied_at)
      VALUES (2, ?);
    `), time.Now().Unix(),
	); err != nil {
		return err
	}

	return tx.Commit()
}
