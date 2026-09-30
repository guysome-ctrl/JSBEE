package apis

import (
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

		if app.Cache == nil {
			app.Cache = map[string]any{}
		}

		if cached, ok := app.Cache[a.CacheKeyLeaders]; ok {
			if leaders, ok := cached.([]database.User); ok {
				c.JSON(http.StatusOK, gin.H{"leaders": leaders})
				return
			}
		}

		leaders, err := app.DBController.GetCityLeaders(ctx)
		if !handleError(c, "GetCityLeaders", err) {
			return
		}

		app.Cache[a.CacheKeyLeaders] = leaders

		if leaders == nil {
			leaders = []database.User{}
		}

		c.JSON(http.StatusOK, gin.H{"leaders": leaders})
	})
}
