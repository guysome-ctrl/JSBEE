package apis

import (
	a "github.com/Ankumeah/JSBEE/backend/internal/app"
	"github.com/Ankumeah/JSBEE/backend/internal/middlewares"
	"github.com/Ankumeah/JSBEE/backend/internal/roles"

	"github.com/gin-gonic/gin"

	"errors"
	"fmt"
	"net/http"
	"uuid"
)

// This route handles all review oprations
func review(r *gin.RouterGroup, app *a.App) {
	group := r.Group("/review",
		middlewares.FireBaseAuthMiddleware(app),
		middlewares.GetUserMiddleware(app),
		func(c *gin.Context) { // Checks if user can review
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

			if !role.CanReview {
				c.AbortWithStatusJSON(
					http.StatusForbidden,
					gin.H{"error": "You are not allowed to review"},
				)
				return
			}

			c.Next()
		},
	)

	// This routes gets all unapproved papers
	group.GET("/papers", func(c *gin.Context) {
		ctx := c.Request.Context()

		papers, err := app.DBController.GetUnapprovedPapers(ctx)
		if !handleError(c, "GetUnapprovedPapers", err) {
			return
		}

		c.JSON(http.StatusOK, gin.H{"papers": papers})
	})

	// This route approves a paper
	group.PATCH("/approve/:paperUUID", func(c *gin.Context) {
		ctx := c.Request.Context()

		paperUUID, err := uuid.Parse(c.Param("paperUUID"))
		if err != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": "Invalid uuid"},
			)
			return
		}

		if !handleError(
			c, "ApprovePaper",
			app.DBController.ApprovePaper(ctx, paperUUID),
		) {
			return
		}

		// TODO: Email author

		c.Status(http.StatusNoContent)
	})

	group.DELETE("/reject/:paperUUID", func(c *gin.Context) {
		ctx := c.Request.Context()

		paperUUID, err := uuid.Parse(c.Param("paperUUID"))
		if err != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": "Invalid uuid"},
			)
			return
		}

		if !handleError(
			c, "RejectPaper",
			app.DBController.RejectPaper(ctx, paperUUID),
		) {
			return
		}

		// TODO: Email author

		c.Status(http.StatusNoContent)
	})
}
