package main

import (
	"github.com/Ankumeah/JSBEE/backend/internal/apis"
	a "github.com/Ankumeah/JSBEE/backend/internal/app"
	"github.com/Ankumeah/JSBEE/backend/internal/middlewares"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"context"
	"sync"
	"time"
)

var Ctx = context.Background()
var app = &a.App{Config: &a.Config{}}
var ginLambda *ginadapter.GinLambda

func Handler(
	ctx context.Context,
	req events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	initLogger(app)
	loadEnv(Ctx, app)

	app.Cache = map[string]any{}

	var wg sync.WaitGroup
	wg.Go(func() { initFirebase(Ctx, app) })
	wg.Go(func() { getDBConnection(Ctx, app) })
	wg.Go(func() { connectObjectStore(Ctx, app) })
	wg.Wait()

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

	ginLambda = ginadapter.New(r)

	app.Logger.InfoContext(Ctx, "Starting lambda handler")
	app.Logger.InfoContext(Ctx, "API_VERSION: "+app.Config.APIVersion)
	lambda.Start(Handler)
}
