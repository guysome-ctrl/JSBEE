package sqlite_test

import (
	"github.com/Ankumeah/JSBEE/backend/internal/database"
	"github.com/Ankumeah/JSBEE/backend/internal/database/migrations"
	"github.com/Ankumeah/JSBEE/backend/internal/database/sqlite"

	"errors"
	"testing"
	"uuid"
)

func TestCoreMembers(t *testing.T) {
	db, err := database.GetDBConnection(Ctx, "file::memory:", sqlite.DriverName)
	if err != nil {
		t.Fatalf("Error while getting DB connection: %v\n", err.Error())
	}
	defer db.Close()

	if err := migrations.ApplyMigrations(Ctx, db); err != nil {
		t.Fatalf("Error while applying migrations: %v\n", err.Error())
	}

	sqlxDB := sqlite.GetSqlxDBController(db)

	first := database.CoreMember{
		UUID:          uuid.New(),
		Name:          "A",
		Role:          "Role A",
		Description:   "Desc A",
		ImageFilename: "core-" + uuid.New().String() + ".jpg",
		SortOrder:     1,
	}
	second := database.CoreMember{
		UUID:          uuid.New(),
		Name:          "B",
		Role:          "Role B",
		Description:   "Desc B",
		ImageFilename: "core-" + uuid.New().String() + ".png",
		SortOrder:     0,
	}

	if err := sqlxDB.AddCoreMember(Ctx, first); err != nil {
		t.Fatalf("Error while adding core member: %v\n", err.Error())
	}
	if err := sqlxDB.AddCoreMember(Ctx, second); err != nil {
		t.Fatalf("Error while adding core member: %v\n", err.Error())
	}

	// Duplicate role must map to ErrExistCoreMember.
	dup := second
	dup.UUID = uuid.New()
	dup.ImageFilename = "core-" + uuid.New().String() + ".jpg"
	err = sqlxDB.AddCoreMember(Ctx, dup)
	if !errors.Is(err, database.ErrExistCoreMember) {
		t.Fatalf("Expected ErrExistCoreMember on duplicate role, got %v\n", err)
	}

	members, err := sqlxDB.GetCoreMembers(Ctx)
	if err != nil {
		t.Fatalf("Error while getting core members: %v\n", err.Error())
	}
	if len(members) != 2 {
		t.Fatalf("Expected 2 core members, got %v\n", len(members))
	}
	// Ordered by sort_order: second (0) before first (1).
	if members[0].UUID != second.UUID || members[1].UUID != first.UUID {
		t.Fatal("Core members are not ordered by sort_order")
	}

	got, err := sqlxDB.GetCoreMember(Ctx, first.UUID)
	if err != nil {
		t.Fatalf("Error while getting core member: %v\n", err.Error())
	}
	if got != first {
		t.Fatalf("Expected member %v, got %v\n", first, got)
	}

	_, err = sqlxDB.GetCoreMember(Ctx, uuid.New())
	if !errors.Is(err, database.ErrInvalidCoreMember) {
		t.Fatalf("Expected ErrInvalidCoreMember for unknown uuid, got %v\n", err)
	}

	first.Name = "A2"
	first.Role = "Role A2"
	first.Description = "Desc A2"
	if err := sqlxDB.UpdateCoreMember(Ctx, first); err != nil {
		t.Fatalf("Error while updating core member: %v\n", err.Error())
	}
	got, err = sqlxDB.GetCoreMember(Ctx, first.UUID)
	if err != nil {
		t.Fatalf("Error while getting updated member: %v\n", err.Error())
	}
	if got.Name != "A2" || got.Role != "Role A2" || got.Description != "Desc A2" {
		t.Fatalf("Update did not persist: %v\n", got)
	}

	newImageFilename := "core-" + uuid.New().String() + ".webp"
	if err := sqlxDB.UpdateCoreMemberPhoto(
		Ctx, first.UUID, newImageFilename,
	); err != nil {
		t.Fatalf("Error while updating photo: %v\n", err.Error())
	}
	got, err = sqlxDB.GetCoreMember(Ctx, first.UUID)
	if err != nil {
		t.Fatalf("Error while getting member after photo update: %v\n", err.Error())
	}
	if got.ImageFilename != newImageFilename {
		t.Fatalf(
			"Expected image filename %q, got %q\n",
			newImageFilename, got.ImageFilename,
		)
	}

	deleted, err := sqlxDB.DeleteCoreMember(Ctx, first.UUID)
	if err != nil {
		t.Fatalf("Error while deleting core member: %v\n", err.Error())
	}
	if deleted.ImageFilename != newImageFilename {
		t.Fatalf(
			"Delete should return the old row for file cleanup, got %v\n",
			deleted,
		)
	}
	_, err = sqlxDB.GetCoreMember(Ctx, first.UUID)
	if !errors.Is(err, database.ErrInvalidCoreMember) {
		t.Fatalf("Expected ErrInvalidCoreMember after delete, got %v\n", err)
	}
}
