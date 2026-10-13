package apis

import (
	"github.com/Ankumeah/JSBEE/backend/internal/apis/shared"
	a "github.com/Ankumeah/JSBEE/backend/internal/app"
	"github.com/Ankumeah/JSBEE/backend/internal/database"

	"github.com/gin-gonic/gin"

	"net/http"
)

// This app handles public info that dont fit into
// other categories
func public(r *gin.RouterGroup, app *a.App) {
	// Returns all city leaders
	r.GET("/leaders", func(c *gin.Context) {
		ctx := c.Request.Context()

		if cached, ok := app.Cache.Load(a.CacheKeyLeaders); ok {
			if leaders, ok := cached.([]database.User); ok {
				c.JSON(http.StatusOK, gin.H{"leaders": leaders})
				return
			}
		}

		leaders, err := app.DBController.GetCityLeaders(ctx)
		if !shared.HandleError(c, "GetCityLeaders", err) {
			return
		}

		app.Cache.Store(a.CacheKeyLeaders, leaders)

		if leaders == nil {
			leaders = []database.User{}
		}

		c.JSON(http.StatusOK, gin.H{"leaders": leaders})
	})

	// Returns all core members with their role, text and
	// public photo link. Photos live in the public bucket.
	r.GET("/core-members", func(c *gin.Context) {
		ctx := c.Request.Context()

		if cached, ok := app.Cache.Load(a.CacheKeyCoreMembers); ok {
			if members, ok := cached.([]database.CoreMember); ok {
				c.JSON(http.StatusOK, gin.H{"core_members": members})
				return
			}
		}

		members, err := app.DBController.GetCoreMembers(ctx)
		if !shared.HandleError(c, "GetCoreMembers", err) {
			return
		}

		if members == nil {
			members = []database.CoreMember{}
		}
		shared.WithPhotoURLs(members, app.ObjectStore.PublicBaseURL())
		app.Cache.Store(a.CacheKeyCoreMembers, members)

		c.JSON(http.StatusOK, gin.H{"core_members": members})
	})
}
