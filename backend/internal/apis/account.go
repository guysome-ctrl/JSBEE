package apis

import (
	a "github.com/Ankumeah/JSBEE/backend/internal/app"
	"github.com/Ankumeah/JSBEE/backend/internal/database"
	"github.com/Ankumeah/JSBEE/backend/internal/middlewares"
	"github.com/Ankumeah/JSBEE/backend/internal/provider"

	"github.com/gin-gonic/gin"

	"errors"
	"net/http"
	"strconv"
	"strings"
	"uuid"
)

// This route deals with account logic
func account(r *gin.RouterGroup, app *a.App) {
	// This route creates a user
	r.POST("/account", func(c *gin.Context) {
		ctx := c.Request.Context()
		authHeader := c.GetHeader("Authorization")

		subscribedString := c.Query("sub")
		subscribed, err := strconv.ParseBool(subscribedString)
		if subscribedString == "" {
			subscribed = false
		} else if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		// Reject on unexpected prefix
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(
				http.StatusUnauthorized,
				gin.H{"error": "No auth token provided"},
			)
			return
		}

		// Verify the JWT
		idToken := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := app.FireBaseClient.VerifyIDToken(ctx, idToken)
		if err != nil {
			c.JSON(
				http.StatusUnauthorized,
				gin.H{"error": "Invalid token"},
			)
			return
		}

		// Get needed feilds
		email, ok := token.Claims["email"].(string)
		if !ok {
			c.AbortWithStatusJSON(
				http.StatusUnprocessableEntity,
				gin.H{"error": "No email claim in token"},
			)
			return
		}
		verified, ok := token.Claims["email_verified"].(bool)
		if !ok || !verified {
			c.AbortWithStatusJSON(
				http.StatusUnprocessableEntity,
				gin.H{"error": "Email not verified"},
			)
			return
		}
		name, ok := token.Claims["name"].(string)
		if !ok {
			c.AbortWithStatusJSON(
				http.StatusUnprocessableEntity,
				gin.H{"error": "No name claim in token"},
			)
			return
		}

		// Get UUID is user exists or else create it
		var userUUID uuid.UUID
		isNewUser := true
		uuidInToken, ok := token.Claims[provider.SiteName+"-uuid"]
		if !ok {
			userUUID = uuid.New()
			if err := app.DBController.AddUser(ctx, database.User{
				UUID:       userUUID,
				Name:       name,
				Email:      email,
				Subscribed: subscribed,
			}); errors.Is(err, database.ErrExistUser) {
				existing, lookupErr := app.DBController.GetUserByEmail(ctx, email)
				if handleError(c, "GetUserByEmail", lookupErr) {
					return
				}
				userUUID = existing.UUID
				isNewUser = false
			} else if !handleError(c, "AddUser", err) {
				return
			}

			// Both new and returning users need the uuid claim in their
			// token, otherwise authed routes keep rejecting them
			if err := app.FireBaseClient.SetCustomUserClaims(
				ctx,
				token.UID,
				map[string]any{
					provider.SiteName + "-uuid": userUUID.String(),
				},
			); !handleError(c, "SetCustomUserClaims", err) {
				if isNewUser {
					app.DBController.DeleteUser(ctx, userUUID)
				}
				return
			}
		} else {
			uuidInTokenString, ok := uuidInToken.(string)
			if !ok {
				c.AbortWithStatusJSON(
					http.StatusBadRequest,
					gin.H{"error": "Invalid user uuid in auth token"},
				)
				return
			}

			userUUID, err = uuid.Parse(uuidInTokenString)
			if err != nil {
				c.AbortWithStatusJSON(
					http.StatusBadRequest,
					gin.H{"error": "Invalid uuid in auth token"},
				)
				return
			}
			isNewUser = false
		}

		if isNewUser {
			c.JSON(http.StatusCreated, gin.H{"uuid": userUUID})
		} else {
			c.JSON(http.StatusOK, gin.H{"uuid": userUUID})
		}
	})

	group := r.Group("/account", middlewares.FireBaseAuthMiddleware(app))

	// This route handles deletion of a user
	group.DELETE("", func(c *gin.Context) {
		ctx := c.Request.Context()

		userUUID, err := uuid.Parse(c.GetString(middlewares.UUIDField))
		if err != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": "Invalid user uuid in auth token"},
			)
			return
		}

		if !handleError(c, "DeleteUser",
			app.DBController.DeleteUser(ctx, userUUID),
		) {
			return
		}

		c.Status(http.StatusNoContent)
	})

	// TODO: Add email change
	group.PATCH("", func(c *gin.Context) {
		ctx := c.Request.Context()

		userUUID, err := uuid.Parse(
			c.GetString(middlewares.UUIDField),
		)
		if err != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": "Invalid user uuid in auth token"},
			)
			return
		}

		subscribedString := c.Query("sub")
		if subscribedString != "" {
			subscribed, err := strconv.ParseBool(c.Query("sub"))
			if err != nil {
				c.JSON(
					http.StatusBadRequest,
					gin.H{"error": err.Error()},
				)
				return
			}

			if !handleError(c, "SetSubscription",
				app.DBController.SetSubscription(ctx, subscribed, userUUID),
			) {
				return
			}
		}

		c.Status(http.StatusOK)
	})
}
