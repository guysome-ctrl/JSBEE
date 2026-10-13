package admin

import (
	"github.com/Ankumeah/JSBEE/backend/internal/apis/shared"
	a "github.com/Ankumeah/JSBEE/backend/internal/app"

	"github.com/gin-gonic/gin"

	"net/http"
)

func papers(group *gin.RouterGroup, app *a.App) {
	// Increments either volume or issue
	group.POST("/increment", func(c *gin.Context) {
		ctx := c.Request.Context()
		field := c.Query("field")

		if field == "volume" {
			if !shared.HandleError(
				c, "IncrementVolume",
				app.DBController.IncrementVolume(ctx),
			) {
				return
			}
		} else if field == "issue" {
			if !shared.HandleError(
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
		if !shared.HandleError(c, "GetReviewedPapers", err) {
			return
		}

		c.JSON(http.StatusOK, gin.H{"papers": papers})
	})

	group.POST("/publish", func(c *gin.Context) {
		ctx := c.Request.Context()

		count, err := app.DBController.PublishPapers(ctx)
		if !shared.HandleError(c, "PublishPapers", err) {
			return
		}

		if count > 0 {
			volumes, err := app.DBController.GetVolumes(ctx)
			if !shared.HandleError(c, "GetVolumes", err) {
				return
			}
			app.Cache.Store(a.CacheKeyVolumes, volumes)
		}

		c.JSON(http.StatusOK, gin.H{"count": count})
	})
}
