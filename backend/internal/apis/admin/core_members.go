package admin

import (
	"github.com/Ankumeah/JSBEE/backend/internal/apis/shared"
	a "github.com/Ankumeah/JSBEE/backend/internal/app"
	"github.com/Ankumeah/JSBEE/backend/internal/database"
	"github.com/Ankumeah/JSBEE/backend/internal/objectstore"

	"github.com/gin-gonic/gin"

	"errors"
	"net/http"
	"strings"
	"uuid"
)

func coreMembers(group *gin.RouterGroup, app *a.App) {
	// This route issues a presigned upload URL for a new core member photo
	group.POST("/core-members/upload-url", func(c *gin.Context) {
		ctx := c.Request.Context()

		var body struct {
			Extension string `json:"extension"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": "Must provide an image extension"},
			)
			return
		}
		switch body.Extension {
		case ".jpg", ".jpeg", ".png", ".webp":
		default:
			c.JSON(http.StatusBadRequest,
				gin.H{"error": "Photo must be a .jpg, .jpeg, .png or .webp image"},
			)
			return
		}

		memberUUID := uuid.New()
		filename := "core-" + memberUUID.String() + body.Extension

		uploadURL, err := app.ObjectStore.PresignedUploadURL(
			ctx, filename, shared.UploadURLTTL,
		)
		if !shared.HandleError(c, "PresignedUploadURL", err) {
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"uuid":           memberUUID,
			"image_filename": filename,
			"upload_url":     uploadURL,
		})
	})

	// This confirms a core member photo upload
	group.POST("/core-members/confirm", func(c *gin.Context) {
		ctx := c.Request.Context()

		var body struct {
			UUID          uuid.UUID `json:"uuid"`
			Name          string    `json:"name"`
			Role          string    `json:"role"`
			Description   string    `json:"description"`
			ImageFilename string    `json:"image_filename"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": "Must provide all member fields"},
			)
			return
		}
		name, role, description, err := shared.ValidateCoreMemberFields(
			body.Name, body.Role, body.Description,
		)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if !strings.HasPrefix(body.ImageFilename, "core-"+body.UUID.String()) {
			c.JSON(http.StatusBadRequest,
				gin.H{"error": "Image filename does not match the issued upload"},
			)
			return
		}

		size, err := app.ObjectStore.PrivateFileSize(ctx, body.ImageFilename)
		if errors.Is(err, objectstore.ErrNoSuchUpload) {
			c.JSON(http.StatusBadRequest,
				gin.H{"error": "Upload not found, request a new upload URL"},
			)
			return
		}
		if !shared.HandleError(c, "PrivateFileSize", err) {
			return
		}
		if size <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Empty photo"})
			app.ObjectStore.DeleteFile(ctx, body.ImageFilename)
			return
		}
		if size > shared.MaxCoreMemberPhotoBytes {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Photo must be at most 20MB"})
			app.ObjectStore.DeleteFile(ctx, body.ImageFilename)
			return
		}

		if err := app.ObjectStore.PublicFile(
			ctx, body.ImageFilename,
		); !shared.HandleError(c, "PublicFile", err) {
			app.ObjectStore.DeleteFile(ctx, body.ImageFilename)
			return
		}

		existing, err := app.DBController.GetCoreMembers(ctx)
		if !shared.HandleError(c, "GetCoreMembers", err) {
			app.ObjectStore.DeleteFile(ctx, body.ImageFilename)
			return
		}
		var nextOrder int64
		for _, m := range existing {
			if m.SortOrder >= nextOrder {
				nextOrder = m.SortOrder + 1
			}
		}

		member := database.CoreMember{
			UUID:          body.UUID,
			Name:          name,
			Role:          role,
			Description:   description,
			ImageFilename: body.ImageFilename,
			SortOrder:     nextOrder,
		}
		if !shared.HandleError(
			c, "AddCoreMember",
			app.DBController.AddCoreMember(ctx, member),
		) {
			app.ObjectStore.DeleteFile(ctx, body.ImageFilename)
			return
		}

		member.ImageURL = strings.TrimSuffix(
			app.ObjectStore.PublicBaseURL(), "/",
		) + "/" + member.ImageFilename
		if cached, ok := app.Cache.Load(a.CacheKeyCoreMembers); ok {
			if members, ok := cached.([]database.CoreMember); ok {
				app.Cache.Store(a.CacheKeyCoreMembers, append(members, member))
			}
		}

		c.JSON(http.StatusCreated, gin.H{"member": member})
	})

	// This updates a core member.
	group.PATCH("/core-members/:memberUUID", func(c *gin.Context) {
		ctx := c.Request.Context()

		memberUUID, err := uuid.Parse(c.Param("memberUUID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid uuid"})
			return
		}

		var body struct {
			Name        *string `json:"name"`
			Role        *string `json:"role"`
			Description *string `json:"description"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": "Must provide a name, role or description"},
			)
			return
		}
		if body.Name == nil && body.Role == nil && body.Description == nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": "Nothing to update"},
			)
			return
		}

		member, err := app.DBController.GetCoreMember(ctx, memberUUID)
		if !shared.HandleError(c, "GetCoreMember", err) {
			return
		}

		name, role, description := member.Name, member.Role, member.Description
		if body.Name != nil {
			name = *body.Name
		}
		if body.Role != nil {
			role = *body.Role
		}
		if body.Description != nil {
			description = *body.Description
		}
		if name, role, description, err = shared.ValidateCoreMemberFields(
			name, role, description,
		); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		member.Name, member.Role, member.Description = name, role, description

		if !shared.HandleError(
			c, "UpdateCoreMember",
			app.DBController.UpdateCoreMember(ctx, member),
		) {
			return
		}

		member.ImageURL = strings.TrimSuffix(
			app.ObjectStore.PublicBaseURL(), "/",
		) + "/" + member.ImageFilename
		if cached, ok := app.Cache.Load(a.CacheKeyCoreMembers); ok {
			if members, ok := cached.([]database.CoreMember); ok {
				for i := range members {
					if members[i].UUID == member.UUID {
						members[i] = member
						break
					}
				}
				app.Cache.Store(a.CacheKeyCoreMembers, members)
			}
		}

		c.JSON(http.StatusOK, gin.H{"member": member})
	})

	// This route issues a presigned upload URL to replace
	// the photo of an existing core member.
	group.POST("/core-members/:memberUUID/photo/upload-url", func(c *gin.Context) {
		ctx := c.Request.Context()

		memberUUID, err := uuid.Parse(c.Param("memberUUID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid uuid"})
			return
		}
		_, getErr := app.DBController.GetCoreMember(
			ctx, memberUUID,
		)
		if !shared.HandleError(c, "GetCoreMember", getErr) {
			return
		}

		var body struct {
			Extension string `json:"extension"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": "Must provide an image extension"},
			)
			return
		}
		switch body.Extension {
		case ".jpg", ".jpeg", ".png", ".webp":
		default:
			c.JSON(http.StatusBadRequest,
				gin.H{"error": "Photo must be a .jpg, .jpeg, .png or .webp image"},
			)
			return
		}

		filename := "core-" + uuid.New().String() + body.Extension

		uploadURL, err := app.ObjectStore.PresignedUploadURL(
			ctx, filename, shared.UploadURLTTL,
		)
		if !shared.HandleError(c, "PresignedUploadURL", err) {
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"image_filename": filename,
			"upload_url":     uploadURL,
		})
	})

	// This confirms a replacement photo for an existing core member.
	group.POST("/core-members/:memberUUID/photo/confirm", func(c *gin.Context) {
		ctx := c.Request.Context()

		memberUUID, err := uuid.Parse(c.Param("memberUUID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid uuid"})
			return
		}

		var body struct {
			ImageFilename string `json:"image_filename"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": "Must provide an image filename"},
			)
			return
		}

		member, err := app.DBController.GetCoreMember(ctx, memberUUID)
		if !shared.HandleError(c, "GetCoreMember", err) {
			return
		}
		if body.ImageFilename == member.ImageFilename {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Photo is already set"})
			return
		}

		size, err := app.ObjectStore.PrivateFileSize(ctx, body.ImageFilename)
		if errors.Is(err, objectstore.ErrNoSuchUpload) {
			c.JSON(http.StatusBadRequest,
				gin.H{"error": "Upload not found, request a new upload URL"},
			)
			return
		}
		if !shared.HandleError(c, "PrivateFileSize", err) {
			return
		}
		if size <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Empty photo"})
			app.ObjectStore.DeleteFile(ctx, body.ImageFilename)
			return
		}
		if size > shared.MaxCoreMemberPhotoBytes {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Photo must be at most 20MB"})
			app.ObjectStore.DeleteFile(ctx, body.ImageFilename)
			return
		}

		if err := app.ObjectStore.PublicFile(
			ctx, body.ImageFilename,
		); !shared.HandleError(c, "PublicFile", err) {
			app.ObjectStore.DeleteFile(ctx, body.ImageFilename)
			return
		}

		oldImageFilename := member.ImageFilename
		if err := app.DBController.UpdateCoreMemberPhoto(
			ctx, memberUUID, body.ImageFilename,
		); !shared.HandleError(c, "UpdateCoreMemberPhoto", err) {
			app.ObjectStore.DeleteFile(ctx, body.ImageFilename)
			return
		}
		if oldImageFilename != "" && oldImageFilename != body.ImageFilename {
			_ = app.ObjectStore.DeleteFile(ctx, oldImageFilename)
		}

		member.ImageFilename = body.ImageFilename
		member.ImageURL = strings.TrimSuffix(
			app.ObjectStore.PublicBaseURL(), "/",
		) + "/" + member.ImageFilename
		if cached, ok := app.Cache.Load(a.CacheKeyCoreMembers); ok {
			if members, ok := cached.([]database.CoreMember); ok {
				for i := range members {
					if members[i].UUID == member.UUID {
						members[i] = member
						break
					}
				}
				app.Cache.Store(a.CacheKeyCoreMembers, members)
			}
		}

		c.JSON(http.StatusOK, gin.H{"member": member})
	})

	// This removes a core member and their photo.
	group.DELETE("/core-members/:memberUUID", func(c *gin.Context) {
		ctx := c.Request.Context()

		memberUUID, err := uuid.Parse(c.Param("memberUUID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid uuid"})
			return
		}

		member, err := app.DBController.DeleteCoreMember(ctx, memberUUID)
		if !shared.HandleError(c, "DeleteCoreMember", err) {
			return
		}
		if member.ImageFilename != "" {
			shared.HandleError(
				c, "DeleteFile", app.ObjectStore.DeleteFile(
					ctx, member.ImageFilename,
				),
			)
		}

		if cached, ok := app.Cache.Load(a.CacheKeyCoreMembers); ok {
			if members, ok := cached.([]database.CoreMember); ok {
				kept := make([]database.CoreMember, 0, len(members))
				for _, m := range members {
					if m.UUID != memberUUID {
						kept = append(kept, m)
					}
				}
				app.Cache.Store(a.CacheKeyCoreMembers, kept)
			}
		}

		c.Status(http.StatusNoContent)
	})
}
