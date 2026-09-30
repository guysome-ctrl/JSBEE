//go:build turso

// Some databases do not support VACCUM. Use
// this db backup stub for those unless they
// impliment some driver specific backups

package main

import (
	a "github.com/Ankumeah/JSBEE/backend/internal/app"

	"context"
)

// Starts the daily DB backup loop in the background
// Backup failures are logged but never panic
func startDailyDBBackup(ctx context.Context, app *a.App) {}
