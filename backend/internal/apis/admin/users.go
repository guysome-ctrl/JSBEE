package admin

import (
	"github.com/Ankumeah/JSBEE/backend/internal/apis/shared"
	a "github.com/Ankumeah/JSBEE/backend/internal/app"
	"github.com/Ankumeah/JSBEE/backend/internal/roles"

	"github.com/gin-gonic/gin"

	"net/http"
	"uuid"
)

func users(group *gin.RouterGroup, app *a.App) {
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
		if !shared.HandleError(c, "GetUser", err) {
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
		if !shared.HandleError(c, "ChangeRole", err) {
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
		); !shared.HandleError(c, "EditCityLead", err) {
			return
		}

		leaders, err := app.DBController.GetCityLeaders(ctx)
		if !shared.HandleError(c, "GetCityLeaders", err) {
			return
		}
		app.Cache.Store(a.CacheKeyLeaders, leaders)

		user, err := app.DBController.GetUser(ctx, userUUID)
		if !shared.HandleError(c, "GetUser", err) {
			return
		}

		c.JSON(http.StatusOK, gin.H{"user": user})
	})
}
