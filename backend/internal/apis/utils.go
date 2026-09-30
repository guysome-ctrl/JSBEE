package apis

import (
	"github.com/Ankumeah/JSBEE/backend/internal/database"

	"github.com/gin-gonic/gin"

	"errors"
	"fmt"
	"net/http"
)

// Any errors obtained can be passed to this func to
// check for nil, handle and properly respond to. If you
// add any new defined errors which require special
// treatment then consider adding it here
//
// If this func returns false that means it received a
// non nil error in which case you are to return immediately
// else you can continue with the route
func handleError(c *gin.Context, scope string, err error) bool {
	if errors.Is(err, database.ErrInvalid) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return false
	} else if errors.Is(err, database.ErrExists) {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return false
	} else if err != nil { // Fallback, any new cases should be added above this
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		c.Error(errors.New(fmt.Sprintf("%v - %v: %v", c.FullPath(), scope, err)))
		return false
	}

	return true
}
