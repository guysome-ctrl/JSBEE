package apis

import (
	a "github.com/Ankumeah/JSBEE/backend/internal/app"
	"github.com/Ankumeah/JSBEE/backend/internal/database"
	"github.com/Ankumeah/JSBEE/backend/internal/frontend"
	"github.com/Ankumeah/JSBEE/backend/internal/middlewares"
	"github.com/Ankumeah/JSBEE/backend/internal/objectstore"
	"github.com/Ankumeah/JSBEE/backend/internal/roles"

	"github.com/gin-gonic/gin"

	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"uuid"
)

// This route handles all admin oprations
func admin(r *gin.RouterGroup, app *a.App) {
	group := r.Group("/admin",
		middlewares.FireBaseAuthMiddleware(app),
		middlewares.GetUserMiddleware(app),
		func(c *gin.Context) { // Checks if user can edit db
			value, ok := c.Get(middlewares.RoleField)
			if !ok {
				c.AbortWithStatusJSON(
					http.StatusInternalServerError,
					gin.H{"error": "Internal server error"},
				)
				c.Error(
					errors.New(c.FullPath() + ": Role field not set"),
				)
				return
			}

			role, ok := value.(roles.Role)
			if !ok {
				c.AbortWithStatusJSON(
					http.StatusInternalServerError,
					gin.H{"error": "Internal server error"},
				)
				c.Error(errors.New(fmt.Sprintf(
					"%v: Invalid role %v", c.FullPath(), value,
				)))
				return
			}

			if !role.CanEditDB {
				c.AbortWithStatusJSON(
					http.StatusForbidden,
					gin.H{"error": "You are not allowed to preform this action"},
				)
				return
			}

			c.Next()
		},
	)

	// Increments either volume or issue
	group.POST("/increment", func(c *gin.Context) {
		ctx := c.Request.Context()
		field := c.Query("field")

		if field == "volume" {
			if !handleError(
				c, "IncrementVolume",
				app.DBController.IncrementVolume(ctx),
			) {
				return
			}
		} else if field == "issue" {
			if !handleError(
				c, "IncrementIssue",
				app.DBController.IncrementIssue(ctx),
			) {
				return
			}

			// TODO: Email authors
		} else {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": "Must provide a valid field"},
			)
		}

		c.Status(http.StatusNoContent)
	})

	group.GET("/reviewed", func(c *gin.Context) {
		ctx := c.Request.Context()

		papers, err := app.DBController.GetReviewedPapers(ctx)
		if !handleError(c, "GetReviewedPapers", err) {
			return
		}

		c.JSON(http.StatusOK, gin.H{"papers": papers})
	})

	group.POST("/publish", func(c *gin.Context) {
		ctx := c.Request.Context()

		count, err := app.DBController.PublishPapers(ctx)
		if !handleError(c, "PublishPapers", err) {
			return
		}

		if count > 0 {
			volumes, err := app.DBController.GetVolumes(ctx)
			if !handleError(c, "GetVolumes", err) {
				return
			}
			if app.Cache == nil {
				app.Cache = map[string]any{}
			}
			app.Cache[a.CacheKeyVolumes] = volumes
		}

		c.JSON(http.StatusOK, gin.H{"count": count})
	})

	group.PATCH("/role/:userUUID", func(c *gin.Context) {
		ctx := c.Request.Context()

		userUUID, err := uuid.Parse(c.Param("userUUID"))
		if err != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": "Invalid user UUID"},
			)
			return
		}

		var newRole roles.Role
		if err = newRole.Scan(c.Query("role")); err != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": "Invalid role"},
			)
			return
		} else if newRole.Role == roles.Owner.Role {
			c.JSON(
				http.StatusForbidden,
				gin.H{"error": "Cannot change role to owner"},
			)
			return
		}

		user, err := app.DBController.GetUser(ctx, userUUID)
		if !handleError(c, "GetUser", err) {
			return
		}
		if user.Role.Role == roles.Owner.Role {
			c.JSON(
				http.StatusForbidden,
				gin.H{"error": "Cannot change the role of owner"},
			)
			return
		}

		err = app.DBController.ChangeRole(ctx, userUUID, newRole)
		if !handleError(c, "ChangeRole", err) {
			return
		}

		c.Status(http.StatusOK)
	})

	// This route sets a user's city lead.
	group.PATCH("/city/:userUUID", func(c *gin.Context) {
		ctx := c.Request.Context()

		userUUID, err := uuid.Parse(c.Param("userUUID"))
		if err != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": "Invalid user UUID"},
			)
			return
		}

		var cityLead *string

		city := c.Query("city")
		cityLead = &city
		if *cityLead == "" {
			cityLead = nil
		}

		if err := app.DBController.EditCityLead(
			ctx, userUUID, cityLead,
		); !handleError(c, "EditCityLead", err) {
			return
		}

		leaders, err := app.DBController.GetCityLeaders(ctx)
		if !handleError(c, "GetCityLeaders", err) {
			return
		}
		if app.Cache == nil {
			app.Cache = map[string]any{}
		}
		app.Cache[a.CacheKeyLeaders] = leaders

		user, err := app.DBController.GetUser(ctx, userUUID)
		if !handleError(c, "GetUser", err) {
			return
		}

		c.JSON(http.StatusOK, gin.H{"user": user})
	})

	// This route issues a presigned upload URL for a new blog
	group.POST("/blog/upload-url", func(c *gin.Context) {
		ctx := c.Request.Context()

		var body struct {
			Title string `json:"title"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": "Must provide a title"},
			)
			return
		}
		if strings.TrimSpace(body.Title) == "" {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": "Title is required"},
			)
			return
		}

		blogUUID := uuid.New()
		filename := fmt.Sprintf("blog-%s.md", blogUUID.String())

		uploadURL, err := app.ObjectStore.PresignedUploadURL(
			ctx, filename, uploadURLTTL,
		)
		if !handleError(c, "PresignedUploadURL", err) {
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"uuid":       blogUUID,
			"filename":   filename,
			"upload_url": uploadURL,
		})
	})

	// This confirms a blog upload
	group.POST("/blog/confirm", func(c *gin.Context) {
		ctx := c.Request.Context()

		var body struct {
			UUID  uuid.UUID `json:"uuid"`
			Title string    `json:"title"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": "Must provide uuid and title"},
			)
			return
		}
		if strings.TrimSpace(body.Title) == "" {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": "Title is required"},
			)
			return
		}

		filename := fmt.Sprintf("blog-%s.md", body.UUID.String())

		size, err := app.ObjectStore.PrivateFileSize(ctx, filename)
		if errors.Is(err, objectstore.ErrNoSuchUpload) {
			c.JSON(http.StatusBadRequest,
				gin.H{"error": "Upload not found, request a new upload URL"},
			)
			return
		}
		if !handleError(c, "PrivateFileSize", err) {
			return
		}
		if size <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Empty content"})
			app.ObjectStore.DeleteFile(ctx, filename)
			return
		}

		if err := app.ObjectStore.PublicFile(
			ctx, filename,
		); !handleError(c, "PublicFile", err) {
			app.ObjectStore.DeleteFile(ctx, filename)
			return
		}

		now := time.Now().Unix()
		if err := app.DBController.AddBlog(ctx, database.Blog{
			UUID:      body.UUID,
			Title:     strings.TrimSpace(body.Title),
			Filename:  filename,
			CreatedAt: now,
			UpdatedAt: now,
		}); !handleError(c, "AddBlog", err) {
			app.ObjectStore.DeleteFile(ctx, filename)
			return
		}

		c.JSON(http.StatusCreated, gin.H{"uuid": body.UUID})
	})

	// This route issues a presigned upload URL to update a blog
	group.POST("/blog/:blogUUID/upload-url", func(c *gin.Context) {
		ctx := c.Request.Context()

		blogUUID, err := uuid.Parse(c.Param("blogUUID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid uuid"})
			return
		}

		blog, err := app.DBController.GetBlog(ctx, blogUUID)
		if !handleError(c, "GetBlog", err) {
			return
		}

		uploadURL, err := app.ObjectStore.PresignedUploadURL(
			ctx, blog.Filename, uploadURLTTL,
		)
		if !handleError(c, "PresignedUploadURL", err) {
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"filename":   blog.Filename,
			"upload_url": uploadURL,
		})
	})

	// This confirms a blog update
	group.PATCH("/blog/:blogUUID", func(c *gin.Context) {
		ctx := c.Request.Context()

		blogUUID, err := uuid.Parse(c.Param("blogUUID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid uuid"})
			return
		}

		var body struct {
			Title           *string `json:"title"`
			ContentReplaced bool    `json:"content_replaced"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": "Must provide a title or content"},
			)
			return
		}
		if body.Title == nil && !body.ContentReplaced {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": "Nothing to update"},
			)
			return
		}

		blog, err := app.DBController.GetBlog(ctx, blogUUID)
		if !handleError(c, "GetBlog", err) {
			return
		}

		if body.Title != nil {
			if strings.TrimSpace(*body.Title) == "" {
				c.JSON(
					http.StatusBadRequest,
					gin.H{"error": "Title is required"},
				)
				return
			}
			blog.Title = strings.TrimSpace(*body.Title)
		}

		if body.ContentReplaced {
			size, err := app.ObjectStore.PrivateFileSize(ctx, blog.Filename)
			if errors.Is(err, objectstore.ErrNoSuchUpload) {
				c.JSON(http.StatusBadRequest,
					gin.H{"error": "Upload not found, request a new upload URL"},
				)
				return
			}
			if !handleError(c, "PrivateFileSize", err) {
				return
			}
			if size <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Empty content"})
				app.ObjectStore.DeleteFile(ctx, blog.Filename)
				return
			}

			if err := app.ObjectStore.PublicFile(
				ctx, blog.Filename,
			); !handleError(c, "PublicFile", err) {
				return
			}
		}

		blog.UpdatedAt = time.Now().Unix()
		if !handleError(c, "UpdateBlog",
			app.DBController.UpdateBlog(ctx, blog),
		) {
			return
		}

		c.Status(http.StatusOK)
	})

	// This deletes a blog
	group.DELETE("/blog/:blogUUID", func(c *gin.Context) {
		ctx := c.Request.Context()

		blogUUID, err := uuid.Parse(c.Param("blogUUID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid uuid"})
			return
		}

		blog, err := app.DBController.GetBlog(ctx, blogUUID)
		if !handleError(c, "GetBlog", err) {
			return
		}

		if err := app.DBController.DeleteBlog(
			ctx, blogUUID,
		); !handleError(c, "DeleteBlog", err) {
			return
		}
		handleError(
			c, "DeleteFile", app.ObjectStore.DeleteFile(
				ctx, blog.Filename,
			),
		)

		c.Status(http.StatusNoContent)
	})

	// This route issues a presigned upload URL to update the about
	group.POST("/about/upload-url", func(c *gin.Context) {
		ctx := c.Request.Context()

		uploadURL, err := app.ObjectStore.PresignedUploadURL(
			ctx, frontend.AboutFilename, uploadURLTTL,
		)
		if !handleError(c, "PresignedUploadURL", err) {
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"filename":   frontend.AboutFilename,
			"upload_url": uploadURL,
		})
	})

	// This confirms the updated about
	group.PUT("/about/confirm", func(c *gin.Context) {
		ctx := c.Request.Context()

		size, err := app.ObjectStore.PrivateFileSize(ctx, frontend.AboutFilename)
		if errors.Is(err, objectstore.ErrNoSuchUpload) {
			c.JSON(http.StatusBadRequest,
				gin.H{"error": "Upload not found, request a new upload URL"},
			)
			return
		}
		if !handleError(c, "PrivateFileSize", err) {
			return
		}
		if size <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Empty content"})
			app.ObjectStore.DeleteFile(ctx, frontend.AboutFilename)
			return
		}

		if err := app.ObjectStore.PublicFile(
			ctx, frontend.AboutFilename,
		); !handleError(c, "PublicFile", err) {
			return
		}

		c.Status(http.StatusOK)
	})
}
