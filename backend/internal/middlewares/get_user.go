package middlewares

import (
	a "github.com/Ankumeah/JSBEE/backend/internal/app"

	"github.com/gin-gonic/gin"

	"net/http"
	"uuid"
)

func GetUserMiddleware(app *a.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		userUUID, err := uuid.Parse(c.GetString(UUIDField))
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				gin.H{"error": "Invalid user uuid in auth token"},
			)
			return
		}

		user, err := app.DBController.GetUser(ctx, userUUID)
		if !handleError(c, "GetUser", err) {
			return
		}

		c.Set(RoleField, user.Role)
		c.Set(EmailField, user.Email)
		c.Set(NameField, user.Name)
		c.Set(UUIDField, user.UUID.String())
		c.Set(SubscribedField, user.Subscribed)
		c.Next()
	}
}
