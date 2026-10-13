package database

import (
	"github.com/Ankumeah/JSBEE/backend/internal/roles"

	"github.com/jmoiron/sqlx"

	"context"
	"uuid"
)

// This interface is reposonsible for executing all
// DB queries and returning type safe results
type DBController interface {
	// This exposes the raw underlying DB object.
	// This is only to be used to execute migrations
	DB() *sqlx.DB

	// Add a new user
	//
	// May return the following errors:
	//   - `ErrExistUser`
	//   - Errors by the underlying DB
	AddUser(ctx context.Context, user User) error

	// Delete a user
	//
	// May return the following errors:
	//   - `ErrInvalidUser`
	//   - Errors by the underlying DB
	DeleteUser(ctx context.Context, uuid uuid.UUID) error

	// Sets a user's subscription
	//
	// May return the following errors:
	//   - `ErrInvalidUser`
	//   - Errors by the underlying DB
	SetSubscription(ctx context.Context, subscribed bool, uuid uuid.UUID) error

	// Get the details of a user
	//
	// May return the following errors:
	//   - `ErrInvalidUser`
	//   - Errors by the underlying DB
	GetUser(ctx context.Context, uuid uuid.UUID) (User, error)

	// Get the details of a user by email
	//
	// May return the following errors:
	//   - `ErrInvalidUser`
	//   - Errors by the underlying DB
	GetUserByEmail(ctx context.Context, email string) (User, error)

	// Adds a new unapproved paper
	//
	// May return the following errors:
	//   - `ErrInvalidUser`
	//   - `ErrExistPaper`
	//   - Errors by the underlying DB
	AddPaper(ctx context.Context, paper Paper) error

	// Gets the details of a paper
	//
	// If the owner of a paper has been deleted
	// `paper.OwnerUUID` will be `nil`
	//
	// May return the following errors:
	//   - `ErrInvalidPaper`
	//   - Errors by the underlying DB
	GetPaper(ctx context.Context, uuid uuid.UUID) (Paper, error)

	// Get all papers by a user
	//
	// May return the following errors:
	//   - `ErrInvalidUser`
	//   - Errors by the underlying DB
	GetUserPapers(ctx context.Context, userUUID uuid.UUID) ([]Paper, error)

	// Marks an unpublished paper as reviewed, moving it to the
	// waiting area for publishing. Safe to run on an already
	// reviewed (but unpublished) paper. Published papers are
	// rejected with `ErrInvalidPaper`
	//
	// May return the following errors:
	//   - `ErrInvalidPaper`
	//   - Errors by the underlying DB
	ApprovePaper(ctx context.Context, uuid uuid.UUID) error

	// Deletes a paper if it is unapproved.
	// In case the paper trying to be deleted is approved,
	// `ErrInvalidPaper` is returned
	//
	// May return the following errors:
	//   - `ErrInvalidPaper`
	//   - Errors by the underlying DB
	RejectPaper(ctx context.Context, paperUUID uuid.UUID) error

	// Get all volumes
	//
	// In case any paper's author has been deleted the
	// `Paper.OwnerUUID` feild will be `nil`
	//
	// May return the following errors:
	//   - Errors by the underlying DB
	GetVolumes(ctx context.Context) ([]Volume, error)

	// Get all unapproved papers (not yet reviewed)
	//
	// May return the following errors:
	//   - Errors by the underlying DB
	GetUnapprovedPapers(ctx context.Context) ([]Paper, error)

	// Get all reviewed papers waiting to be published
	//
	// May return the following errors:
	//   - Errors by the underlying DB
	GetReviewedPapers(ctx context.Context) ([]Paper, error)

	// Publish all reviewed papers by assigning numbers and
	// current volume/issue from state. Returns the count of published papers.
	//
	// May return the following errors:
	//   - Errors by the underlying DB
	PublishPapers(ctx context.Context) (int64, error)

	// Increment state.volume by 1
	//
	// May return the following errors:
	//   - Errors by the underlying DB
	IncrementVolume(ctx context.Context) error

	// Increment state.issue by 1
	//
	// May return the following errors:
	//   - Errors by the underlying DB
	IncrementIssue(ctx context.Context) error

	// Executes the backup procedure
	//
	// May return the following errors:
	//   - Errors from the underlying DB
	BackupDB(ctx context.Context) (string, error)

	// Changes the role of the given user
	//
	// May return the following errors:
	//   - `ErrInvalidUser`
	//   - Errors from the underlying DB
	ChangeRole(ctx context.Context, userUUID uuid.UUID, newRole roles.Role) error

	// Sets a user's city lead to the given value.
	// A nil newCityLead clears it back to NULL.
	//
	// May return the following errors:
	//   - `ErrInvalidUser`
	//   - Errors by the underlying DB
	EditCityLead(ctx context.Context, userUUID uuid.UUID, newCityLead *string) error

	// Get all users whose city lead is not NULL
	//
	// May return the following errors:
	//   - Errors by the underlying DB
	GetCityLeaders(ctx context.Context) ([]User, error)

	// Adds a new blog entry
	//
	// May return the following errors:
	//   - `ErrExistBlog`
	//   - Errors by the underlying DB
	AddBlog(ctx context.Context, blog Blog) error

	// Gets the details of a blog entry
	//
	// May return the following errors:
	//   - `ErrInvalidBlog`
	//   - Errors by the underlying DB
	GetBlog(ctx context.Context, uuid uuid.UUID) (Blog, error)

	// Get all blog entries, newest first
	//
	// May return the following errors:
	//   - Errors by the underlying DB
	GetBlogs(ctx context.Context) ([]Blog, error)

	// Updates the title and/or file of a blog entry
	//
	// May return the following errors:
	//   - `ErrInvalidBlog`
	//   - Errors by the underlying DB
	UpdateBlog(ctx context.Context, blog Blog) error

	// Deletes a blog entry
	//
	// May return the following errors:
	//   - `ErrInvalidBlog`
	//   - Errors by the underlying DB
	DeleteBlog(ctx context.Context, uuid uuid.UUID) error

	// Adds a new core member
	//
	// May return the following errors:
	//   - `ErrExistCoreMember`
	//   - Errors by the underlying DB
	AddCoreMember(ctx context.Context, member CoreMember) error

	// Gets the details of a core member
	//
	// May return the following errors:
	//   - `ErrInvalidCoreMember`
	//   - Errors by the underlying DB
	GetCoreMember(ctx context.Context, uuid uuid.UUID) (CoreMember, error)

	// Get all core members
	//
	// May return the following errors:
	//   - Errors by the underlying DB
	GetCoreMembers(ctx context.Context) ([]CoreMember, error)

	// Updates a core member. Ignores photo
	//
	// May return the following errors:
	//   - `ErrInvalidCoreMember`
	//   - `ErrExistCoreMember` (role conflict)
	//   - Errors by the underlying DB
	UpdateCoreMember(ctx context.Context, member CoreMember) error

	// Updates photo of a core member
	//
	// May return the following errors:
	//   - `ErrInvalidCoreMember`
	//   - `ErrExistCoreMember` (image filename conflict)
	//   - Errors by the underlying DB
	UpdateCoreMemberPhoto(
		ctx context.Context, uuid uuid.UUID, imageFilename string,
	) error

	// Deletes a core member
	//
	// May return the following errors:
	//   - `ErrInvalidCoreMember`
	//   - Errors by the underlying DB
	DeleteCoreMember(ctx context.Context, uuid uuid.UUID) (CoreMember, error)
}

type User struct {
	UUID       uuid.UUID  `json:"uuid" binding:"required" db:"uuid"`
	Name       string     `json:"name" binding:"required" db:"name"`
	Email      string     `json:"email" binding:"required" db:"email"`
	Role       roles.Role `json:"role" binding:"required" db:"role"`
	Subscribed bool       `json:"subscribed" binding:"required" db:"subscribed"`
	CityLead   *string    `json:"city_lead" db:"city_lead"`
}

type Paper struct {
	UUID      uuid.UUID  `json:"uuid" binding:"required" db:"uuid"`
	Title     string     `json:"title" binding:"required" db:"title"`
	Number    *uint64    `json:"number" binding:"required" db:"number"`
	Filename  string     `json:"filename" binding:"required" db:"filename"`
	OwnerUUID *uuid.UUID `json:"owner_uuid" binding:"required" db:"owner_uuid"`
	Reviewed  bool       `json:"reviewed" binding:"required" db:"reviewed"`
	Location  string     `json:"location" binding:"required" db:"location"`
	Category  string     `json:"category" binding:"required" db:"category"`
}

type Issue struct {
	Number uint64  `json:"number" binding:"required"`
	Papers []Paper `json:"papers" binding:"required"`
}

type Volume struct {
	Number uint64  `json:"number" binding:"required"`
	Issues []Issue `json:"issues" binding:"required"`
}

type Blog struct {
	UUID      uuid.UUID `json:"uuid" binding:"required" db:"uuid"`
	Title     string    `json:"title" binding:"required" db:"title"`
	Filename  string    `json:"filename" binding:"required" db:"filename"`
	CreatedAt int64     `json:"created_at" binding:"required" db:"created_at"`
	UpdatedAt int64     `json:"updated_at" binding:"required" db:"updated_at"`
}

type CoreMember struct {
	UUID          uuid.UUID `json:"uuid" binding:"required" db:"uuid"`
	Name          string    `json:"name" binding:"required" db:"name"`
	Role          string    `json:"role" binding:"required" db:"role"`
	Description   string    `json:"description" binding:"required" db:"description"`
	ImageFilename string    `json:"image_filename" binding:"required" db:"image_filename"`
	SortOrder     int64     `json:"sort_order" binding:"required" db:"sort_order"`
	ImageURL      string    `json:"image_url" db:"-"`
}
