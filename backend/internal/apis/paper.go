package apis

import (
	"github.com/Ankumeah/JSBEE/backend/internal/apis/shared"
	a "github.com/Ankumeah/JSBEE/backend/internal/app"
	"github.com/Ankumeah/JSBEE/backend/internal/database"
	"github.com/Ankumeah/JSBEE/backend/internal/middlewares"
	"github.com/Ankumeah/JSBEE/backend/internal/objectstore"

	"github.com/gin-gonic/gin"

	"errors"
	"net/http"
	"strings"
	"uuid"
)

// This route deals with all paper related apis
func paper(r *gin.RouterGroup, app *a.App) {
	group := r.Group("/paper")

	// This route issues a presigned upload URL for a new paper
	group.POST("/upload-url", middlewares.FireBaseAuthMiddleware(app),
		func(c *gin.Context) {
			ctx := c.Request.Context()

			paperUUID := uuid.New()
			filename := paperUUID.String() + ".pdf"

			uploadURL, err := app.ObjectStore.PresignedUploadURL(
				ctx, filename, shared.UploadURLTTL,
			)
			if !shared.HandleError(c, "PresignedUploadURL", err) {
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"uuid":       paperUUID,
				"filename":   filename,
				"upload_url": uploadURL,
			})
		},
	)

	// This route confirms a paper upload
	group.POST("/confirm", middlewares.FireBaseAuthMiddleware(app),
		func(c *gin.Context) {
			ctx := c.Request.Context()

			userUUID, err := uuid.Parse(c.GetString(middlewares.UUIDField))
			if !shared.HandleError(c, "uuid.Parse", err) {
				return
			}

			var body struct {
				UUID     uuid.UUID `json:"uuid"`
				Title    string    `json:"title"`
				Location string    `json:"location"`
				Category string    `json:"category"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest,
					gin.H{"error": "Must provide uuid, title, location and category"},
				)
				return
			}
			if strings.TrimSpace(body.Title) == "" {
				c.JSON(http.StatusBadRequest,
					gin.H{"error": "Title is required"},
				)
				return
			}
			location, category, err := shared.ValidatePaperMetadata(
				body.Location, body.Category,
			)
			if err != nil {
				c.JSON(http.StatusBadRequest,
					gin.H{"error": err.Error()},
				)
				return
			}

			filename := body.UUID.String() + ".pdf"

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
				c.JSON(http.StatusBadRequest, gin.H{"error": "Empty file"})
				app.ObjectStore.DeleteFile(ctx, filename)
				return
			}

			if err := app.ObjectStore.PublicFile(
				ctx, filename,
			); !shared.HandleError(c, "PublicFile", err) {
				app.ObjectStore.DeleteFile(ctx, filename)
				return
			}

			if err := app.DBController.AddPaper(ctx, database.Paper{
				UUID:      body.UUID,
				Title:     strings.TrimSpace(body.Title),
				Filename:  filename,
				OwnerUUID: &userUUID,
				Location:  location,
				Category:  category,
			}); !shared.HandleError(c, "AddPaper", err) {
				app.ObjectStore.DeleteFile(ctx, filename)
				return
			}

			c.JSON(http.StatusCreated, gin.H{"uuid": body.UUID})
		},
	)

	// This route returns the details of a paper
	group.GET("/:paperUUID", func(c *gin.Context) {
		ctx := c.Request.Context()

		paperUUID, err := uuid.Parse(c.Param("paperUUID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid uuid"})
			return
		}

		paper, err := app.DBController.GetPaper(ctx, paperUUID)
		if !shared.HandleError(c, "GetPaper", err) {
			return
		}

		fileURL := app.ObjectStore.PublicBaseURL() + "/" + paper.Filename

		c.JSON(http.StatusOK, gin.H{"paper": paper, "file_url": fileURL})
	})

	// Returns all published volumes
	group.GET("/volumes", func(c *gin.Context) {
		ctx := c.Request.Context()

		if cached, ok := app.Cache.Load(a.CacheKeyVolumes); ok {
			if volumes, ok := cached.([]database.Volume); ok {
				c.JSON(http.StatusOK, gin.H{"volumes": volumes})
				return
			}
		}

		volumes, err := app.DBController.GetVolumes(ctx)
		if !shared.HandleError(c, "GetVolumes", err) {
			return
		}

		if volumes == nil {
			volumes = []database.Volume{}
		}
		app.Cache.Store(a.CacheKeyVolumes, volumes)

		c.JSON(http.StatusOK, gin.H{"volumes": volumes})
	})
}
