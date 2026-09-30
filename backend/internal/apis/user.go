package apis

import (
	a "github.com/Ankumeah/JSBEE/backend/internal/app"

	"github.com/gin-gonic/gin"

	"net/http"
	"uuid"
)

func user(r *gin.RouterGroup, app *a.App) {
	group := r.Group("/user")

	// This route returns the details of a user
	group.GET("/", func(c *gin.Context) {
		ctx := c.Request.Context()
		unparsedUserUUID := c.Query("uuid")
		userEmail := c.Query("email")

		if unparsedUserUUID != "" {
			userUUID, err := uuid.Parse(unparsedUserUUID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID"})
				return
			}

			user, err := app.DBController.GetUser(ctx, userUUID)
			if !handleError(c, "GetUser", err) {
				return
			}

			c.JSON(http.StatusOK, gin.H{"user": user})
			return
		} else if userEmail != "" {
			user, err := app.DBController.GetUserByEmail(ctx, userEmail)
			if !handleError(c, "GetUserByEmail", err) {
				return
			}

			c.JSON(http.StatusOK, gin.H{"user": user})
			return
		} else {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": "Must provide either user uuid or email"},
			)
			return
		}
	})

	// This route returns all papers of a user
	group.GET("/papers", func(c *gin.Context) {
		ctx := c.Request.Context()

		unparsedUserUUID := c.Query("uuid")
		userEmail := c.Query("email")

		var userUUID uuid.UUID

		if unparsedUserUUID != "" {
			var err error
			userUUID, err = uuid.Parse(unparsedUserUUID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID"})
				return
			}
		} else if userEmail != "" {
			user, err := app.DBController.GetUserByEmail(ctx, userEmail)
			if !handleError(c, "GetUserByEmail", err) {
				return
			}

			userUUID = user.UUID
		} else {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": "Must provide either user uuid or email"},
			)
			return
		}

		papers, err := app.DBController.GetUserPapers(ctx, userUUID)
		if !handleError(c, "GetUserPapers", err) {
			return
		}

		c.JSON(http.StatusOK, gin.H{"papers": papers})
	})
}
