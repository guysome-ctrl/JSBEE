package main

import (
	a "github.com/Ankumeah/JSBEE/backend/internal/app"
	"github.com/Ankumeah/JSBEE/backend/internal/database"
	"github.com/Ankumeah/JSBEE/backend/internal/database/migrations"
	"github.com/Ankumeah/JSBEE/backend/internal/frontend"
	"github.com/Ankumeah/JSBEE/backend/internal/frontend/assets"
	"github.com/Ankumeah/JSBEE/backend/internal/objectstore"
	"github.com/Ankumeah/JSBEE/backend/internal/provider"

	"bytes"
	"context"
	"log"
	"log/slog"
	"os"
	"path"
	"strings"
	"uuid"
)

// Initalises the logger
// Should be initalised before anything
// else as almost everything can have errors
func initLogger(app *a.App) {
	app.Logger = slog.New(
		slog.NewTextHandler(os.Stdout, nil),
	)
}

// Runs DB migrations
// Exits program on any errors
func runDBMigrations(ctx context.Context, app *a.App) {
	log.Println("Running DB migrations")

	if err := migrations.ApplyMigrations(
		ctx, app.DBController.DB(),
	); err != nil {
		log.Fatalf("Error while running DB migrations: %v\n", err)
	}

	log.Println("DB migrations completed")
}

// Creates the ComponentUpdater struct
// Creates the save dir if it dosent exist
// Exits program on any errors with mkdir
func getComponentUpdater(app *a.App) {
	log.Println("Getting component updater")
	var err error
	app.ComponentUpdater, err = frontend.GetComponentUpdater(
		app.Config.FrontendSaveDir,
		app.Config.FireBaseClientConfig,
	)
	if err != nil {
		log.Fatalf("Error while getting component updater: %v\n", err.Error())
	}
	log.Println("Got component updater")
}

// Asset files may exist at launch, this
// writes the bundled asset files
// Exits on any error
func saveAssets(ctx context.Context, app *a.App) {
	log.Println("Saveing assets")

	if err := assets.SaveAssets(
		app.Config.FrontendSaveDir,
	); err != nil {
		log.Fatalf("Error while saveing assets: %v\n", err.Error())
	}

	log.Println("Assets saved")
}

// Component files may not exist at launch, this generates
// them and updates them if they alreday exist.
// Volumes and team pages are static shells, their data
// is fetched at runtime from the public API.
// Exits on any error
func generateInitalComponents(ctx context.Context, app *a.App) {
	log.Println("Generating inital components")

	if err := app.ComponentUpdater.UpdateAll(
		ctx, app.ObjectStore.PublicBaseURL(),
	); err != nil {
		log.Fatalf("Error while generating inital components: %v\n", err.Error())
	}

	log.Println("Generated inital components")
}

func connectObjectStore(ctx context.Context, app *a.App) {
	log.Println("Connecting to object store")

	config, err := objectstore.NewS3StaticConfigFromJSON(
		app.Config.ObjectStoreConfig,
	)
	if err != nil {
		log.Fatalf("Error while getting s3 config: %v\n", err.Error())
	}

	app.ObjectStore, err = objectstore.GetStaticS3Client(
		ctx, config,
	)
	if err != nil {
		log.Fatalf("Error while getting s3 client: %v\n", err.Error())
	}

	log.Println("Connected to object store")
}

func initObjectStore(ctx context.Context, app *a.App) {
	log.Println("Initalizing object store")

	if err := app.ObjectStore.Init(ctx); err != nil {
		log.Fatalf("Error while initalizing object store: %v\n", err.Error())
	}

	log.Println("Object store initalized")
}

// Creates the about us page in the object store
// if it dosent exist yet, exits on any error
func seedAboutPage(ctx context.Context, app *a.App) {
	log.Println("Seeding about page")

	content, err := app.ObjectStore.GetFile(ctx, frontend.AboutFilename)
	if err == nil {
		content.Close()
		log.Println("About page already exists")
		return
	}

	buf := bytes.NewBufferString(provider.AboutSeedContent)
	if err := app.ObjectStore.AddFile(
		ctx, frontend.AboutFilename, buf, int64(buf.Len()),
	); err != nil {
		log.Fatalf("Error while seeding about page: %v\n", err.Error())
	}
	if err := app.ObjectStore.PublicFile(ctx, frontend.AboutFilename); err != nil {
		log.Fatalf("Error while publishing about page: %v\n", err.Error())
	}

	log.Println("About page seeded")
}

// Seeds core members
func seedCoreMembers(ctx context.Context, app *a.App) {
	log.Println("Seeding core members")

	existing, err := app.DBController.GetCoreMembers(ctx)
	if err != nil {
		log.Fatalf("Error while checking core members: %v\n", err.Error())
	}
	present := make(map[string]bool, len(existing))
	for _, m := range existing {
		present[m.Role] = true
	}

	seeded := 0
	for i, role := range provider.CoreMemberOrder {
		if present[role] {
			continue
		}
		m := provider.CoreMembers[role]

		assetKey := strings.TrimPrefix(m.ImagePath, "/")
		data, ok := assets.GetAsset(assetKey)
		if !ok || len(data) == 0 {
			log.Fatalf("Missing photo for core member %q\n", role)
		}

		ext := strings.ToLower(path.Ext(assetKey))
		if ext == "" {
			ext = ".jpg"
		}

		memberUUID := uuid.New()
		filename := "core-" + memberUUID.String() + ext

		buf := bytes.NewReader(data)
		if err := app.ObjectStore.AddFile(
			ctx, filename, buf, int64(len(data)),
		); err != nil {
			log.Fatalf("Error seeding core member photo %q: %v\n", role, err.Error())
		}
		if err := app.ObjectStore.PublicFile(ctx, filename); err != nil {
			log.Fatalf("Error publishing core member photo %q: %v\n", role, err.Error())
		}

		if err := app.DBController.AddCoreMember(ctx, database.CoreMember{
			UUID:          memberUUID,
			Name:          m.Name,
			Role:          role,
			Description:   m.Description,
			ImageFilename: filename,
			SortOrder:     int64(i),
		}); err != nil {
			_ = app.ObjectStore.DeleteFile(ctx, filename)
			log.Fatalf("Error while seeding core member %q: %v\n", role, err.Error())
		}
		seeded++
	}

	log.Printf("Core members seeded: %d new\n", seeded)
}
