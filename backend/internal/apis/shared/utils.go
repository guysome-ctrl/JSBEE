package shared

import (
	"github.com/Ankumeah/JSBEE/backend/internal/database"

	"github.com/gin-gonic/gin"

	"errors"
	"fmt"
	"net/http"
	"strings"
)

// Any errors obtained can be passed to this func to
// check for nil, handle and properly respond to. If you
// add any new defined errors which require special
// treatment then consider adding it here
//
// If this func returns false that means it received a
// non nil error in which case you are to return immediately
// else you can continue with the route
func HandleError(c *gin.Context, scope string, err error) bool {
	if errors.Is(err, database.ErrInvalid) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return false
	} else if errors.Is(err, database.ErrExists) {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return false
	} else if err != nil { // Fallback, any new cases should be added above this
		c.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "Internal server error"},
		)
		c.Error(errors.New(fmt.Sprintf("%v - %v: %v", c.FullPath(), scope, err)))
		return false
	}

	return true
}

// trims and checks the location and category
func ValidatePaperMetadata(
	location string, category string,
) (string, string, error) {
	location = strings.TrimSpace(location)
	if location == "" {
		return "", "", errors.New("Location is required")
	}
	if len([]rune(location)) > maxPaperFieldLength {
		return "", "", errors.New("Location must be at most 120 characters")
	}

	category = strings.TrimSpace(category)
	if category == "" {
		return "", "", errors.New("Category is required")
	}
	if len([]rune(category)) > maxPaperFieldLength {
		return "", "", errors.New("Category must be at most 120 characters")
	}

	return location, category, nil
}

// trims and checks the name, role and description of a core member
func ValidateCoreMemberFields(
	name, role, description string,
) (string, string, string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", "", "", errors.New("Name is required")
	}
	if len([]rune(name)) > maxCoreMemberNameLength {
		return "", "", "", errors.New("Name must be at most 120 characters")
	}

	role = strings.TrimSpace(role)
	if role == "" {
		return "", "", "", errors.New("Role is required")
	}
	if len([]rune(role)) > maxCoreMemberRoleLength {
		return "", "", "", errors.New("Role must be at most 120 characters")
	}

	description = strings.TrimSpace(description)
	if description == "" {
		return "", "", "", errors.New("Description is required")
	}

	return name, role, description, nil
}

// fills ImageURL for each member from the public base URL
func WithPhotoURLs(members []database.CoreMember, baseURL string) {
	base := strings.TrimSuffix(baseURL, "/")
	for i := range members {
		if members[i].ImageFilename != "" {
			members[i].ImageURL = base + "/" + members[i].ImageFilename
		}
	}
}
