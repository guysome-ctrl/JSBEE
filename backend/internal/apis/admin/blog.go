package admin

import (
	"github.com/Ankumeah/JSBEE/backend/internal/apis/shared"
	a "github.com/Ankumeah/JSBEE/backend/internal/app"
	"github.com/Ankumeah/JSBEE/backend/internal/database"
	"github.com/Ankumeah/JSBEE/backend/internal/objectstore"

	"github.com/gin-gonic/gin"

	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"uuid"
)

func blog(group *gin.RouterGroup, app *a.App) {
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
			ctx, filename, shared.UploadURLTTL,
		)
		if !shared.HandleError(c, "PresignedUploadURL", err) {
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
		if !shared.HandleError(c, "PrivateFileSize", err) {
			return
		}
		if size <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Empty content"})
			app.ObjectStore.DeleteFile(ctx, filename)
			return
		}

		if err := app.ObjectStore.PublicFile(
			ctx, filename,
		); !shared.HandleError(c, "PublicFile", err) {
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
		}); !shared.HandleError(c, "AddBlog", err) {
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
		if !shared.HandleError(c, "GetBlog", err) {
			return
		}

		uploadURL, err := app.ObjectStore.PresignedUploadURL(
			ctx, blog.Filename, shared.UploadURLTTL,
		)
		if !shared.HandleError(c, "PresignedUploadURL", err) {
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
		if !shared.HandleError(c, "GetBlog", err) {
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
			if !shared.HandleError(c, "PrivateFileSize", err) {
				return
			}
			if size <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Empty content"})
				app.ObjectStore.DeleteFile(ctx, blog.Filename)
				return
			}

			if err := app.ObjectStore.PublicFile(
				ctx, blog.Filename,
			); !shared.HandleError(c, "PublicFile", err) {
				return
			}
		}

		blog.UpdatedAt = time.Now().Unix()
		if !shared.HandleError(c, "UpdateBlog",
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
		if !shared.HandleError(c, "GetBlog", err) {
			return
		}

		if err := app.DBController.DeleteBlog(
			ctx, blogUUID,
		); !shared.HandleError(c, "DeleteBlog", err) {
			return
		}
		shared.HandleError(
			c, "DeleteFile", app.ObjectStore.DeleteFile(
				ctx, blog.Filename,
			),
		)

		c.Status(http.StatusNoContent)
	})
}
