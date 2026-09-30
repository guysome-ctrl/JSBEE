//go:build !turso

package main

import (
	a "github.com/Ankumeah/JSBEE/backend/internal/app"

	"context"
	"os"
	"time"
)

const dbBackupInterval = 24 * time.Hour
const dbBackupTimeout = 10 * time.Minute

// Starts the daily DB backup loop in the background
// Backup failures are logged but never panic
func startDailyDBBackup(ctx context.Context, app *a.App) {
	go func() {
		backup := func() {
			backupCtx, cancel := context.WithTimeout(ctx, dbBackupTimeout)
			defer cancel()
			if err := storeBackup(backupCtx, app); err != nil {
				app.Logger.ErrorContext(ctx,
					"Error while storing DB backup: "+err.Error(),
				)
			} else {
				app.Logger.InfoContext(ctx, "Stored DB backup")
			}
		}

		// Backup once on startup so restarts don't leave gaps,
		// then every `dbBackupInterval` after that.
		backup()

		ticker := time.NewTicker(dbBackupInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				backup()
			}
		}
	}()
}

// Snapshots the DB via the controller and stores the
// snapshot through the object store, then removes it.
func storeBackup(ctx context.Context, app *a.App) error {
	backupPath, err := app.DBController.BackupDB(ctx)
	if err != nil {
		return err
	}
	defer os.Remove(backupPath)

	return app.ObjectStore.StoreDBBackup(ctx, backupPath)
}
