package admin

import (
	"github.com/Ankumeah/JSBEE/backend/internal/apis/shared"
	a "github.com/Ankumeah/JSBEE/backend/internal/app"
	"github.com/Ankumeah/JSBEE/backend/internal/frontend"
	"github.com/Ankumeah/JSBEE/backend/internal/objectstore"

	"github.com/gin-gonic/gin"

	"errors"
	"net/http"
)

func about(group *gin.RouterGroup, app *a.App) {
	// This route issues a presigned upload URL to update the about
	group.POST("/about/upload-url", func(c *gin.Context) {
		ctx := c.Request.Context()

		uploadURL, err := app.ObjectStore.PresignedUploadURL(
			ctx, frontend.AboutFilename, shared.UploadURLTTL,
		)
		if !shared.HandleError(c, "PresignedUploadURL", err) {
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
		if !shared.HandleError(c, "PrivateFileSize", err) {
			return
		}
		if size <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Empty content"})
			app.ObjectStore.DeleteFile(ctx, frontend.AboutFilename)
			return
		}

		if err := app.ObjectStore.PublicFile(
			ctx, frontend.AboutFilename,
		); !shared.HandleError(c, "PublicFile", err) {
			return
		}

		c.Status(http.StatusOK)
	})
}
