package apis

import (
	a "github.com/Ankumeah/JSBEE/backend/internal/app"
	"github.com/Ankumeah/JSBEE/backend/internal/database"

	"github.com/gin-gonic/gin"

	"io"
	"net/http"
	"uuid"
)

// This route deals with reading blog entries,
// writing is done by admins in `admin.go`
func blog(r *gin.RouterGroup, app *a.App) {
	group := r.Group("/blogs")

	// This route returns all blog entries, newest first
	group.GET("", func(c *gin.Context) {
		ctx := c.Request.Context()

		blogs, err := app.DBController.GetBlogs(ctx)
		if !handleError(c, "GetBlogs", err) {
			return
		}

		if blogs == nil {
			blogs = []database.Blog{}
		}

		c.JSON(http.StatusOK, gin.H{"blogs": blogs})
	})

	// This route returns the details of a blog entry
	group.GET("/:blogUUID", func(c *gin.Context) {
		ctx := c.Request.Context()

		blogUUID, err := uuid.Parse(c.Param("blogUUID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid uuid"})
			return
		}

		blog, err := app.DBController.GetBlog(ctx, blogUUID)
		if !handleError(c, "GetBlog", err) {
			return
		}

		c.JSON(http.StatusOK, gin.H{"blog": blog})
	})

	// This route returns the raw markdown content of a blog entry
	group.GET("/:blogUUID/content", func(c *gin.Context) {
		ctx := c.Request.Context()

		blogUUID, err := uuid.Parse(c.Param("blogUUID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid uuid"})
			return
		}

		blog, err := app.DBController.GetBlog(ctx, blogUUID)
		if !handleError(c, "GetBlog", err) {
			return
		}

		content, err := app.ObjectStore.GetFile(ctx, blog.Filename)
		if err != nil {
			c.JSON(
				http.StatusNotFound,
				gin.H{"error": "Blog content not found"},
			)
			return
		}
		defer content.Close()

		markdown, err := io.ReadAll(content)
		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{"error": "Internal server error"},
			)
			return
		}

		c.Data(http.StatusOK, "text/markdown; charset=utf-8", markdown)
	})
}
