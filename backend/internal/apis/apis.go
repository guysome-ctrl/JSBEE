package apis

import (
	"github.com/Ankumeah/JSBEE/backend/internal/apis/admin"
	a "github.com/Ankumeah/JSBEE/backend/internal/app"
	"github.com/gin-gonic/gin"
)

// This func registers all routes within the passed router
func Apis(r *gin.RouterGroup, app *a.App) {
	ping(r)
	account(r, app)
	user(r, app)
	paper(r, app)
	review(r, app)
	admin.Admin(r, app)
	blog(r, app)
	public(r, app)
}
