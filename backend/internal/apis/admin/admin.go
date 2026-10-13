package admin

import (
	a "github.com/Ankumeah/JSBEE/backend/internal/app"
	"github.com/Ankumeah/JSBEE/backend/internal/middlewares"
	"github.com/Ankumeah/JSBEE/backend/internal/roles"

	"github.com/gin-gonic/gin"

	"errors"
	"fmt"
	"net/http"
)

// This route handles all admin oprations
func Admin(r *gin.RouterGroup, app *a.App) {
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

	papers(group, app)
	users(group, app)
	blog(group, app)
	about(group, app)
	coreMembers(group, app)
}
