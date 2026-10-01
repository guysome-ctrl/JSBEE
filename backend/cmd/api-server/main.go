package main

import (
	"github.com/Ankumeah/JSBEE/backend/internal/apis"
	a "github.com/Ankumeah/JSBEE/backend/internal/app"
	"github.com/Ankumeah/JSBEE/backend/internal/middlewares"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"context"
	"os"
	"sync"
	"time"
)

var Ctx = context.Background()
var app = &a.App{Config: &a.Config{}}

func main() {
	initLogger(app)
	loadEnv(Ctx, app)

	var wg sync.WaitGroup
	wg.Go(func() { initFirebase(Ctx, app) })
	wg.Go(func() { getComponentUpdater(Ctx, app) })
	wg.Go(func() { getDBConnection(Ctx, app) })
	wg.Go(func() { connectObjectStore(Ctx, app) })
	wg.Wait()

	app.Logger.InfoContext(Ctx, "Starting daily DB backups")
	startDailyDBBackup(Ctx, app)

	app.Logger.InfoContext(Ctx, "Starting http server")
	r := gin.Default()
	r.Use(cors.New(cors.Config{ // TODO: Acctly fill these
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	apiGroup := r.Group(
		"/api/"+app.Config.APIVersion+"/",
		middlewares.LogMiddleware(app, Ctx),
	)
	apis.Apis(apiGroup, app)

	app.Logger.InfoContext(Ctx, "Running backend on port: "+app.Config.Port)
	app.Logger.InfoContext(Ctx, "API_VERSION: "+app.Config.APIVersion)

	err := r.Run(":" + app.Config.Port)
	if err != nil {
		app.Logger.ErrorContext(Ctx, "Error while running server: "+err.Error())
		os.Exit(1)
	}
}
