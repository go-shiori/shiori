package api_v1

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-shiori/shiori/internal/http/middleware"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
)

// @Summary					List tags
// @Description				List all tags
// @Tags						Tags
// @securityDefinitions.apikey	ApiKeyAuth
// @Produce					json
// @Param						with_bookmark_count	query		boolean	false	"Include bookmark count for each tag"
// @Param						bookmark_id			query		integer	false	"Filter tags by bookmark ID"
// @Param						search				query		string	false	"Search tags by name"
// @Success					200					{array}		model.TagDTO
// @Failure					403					{object}	nil	"Authentication required"
// @Failure					500					{object}	nil	"Internal server error"
// @Router						/api/v1/tags [get]
func HandleListTags(deps model.Dependencies, c model.WebContext) {
	if err := middleware.RequireLoggedInUser(deps, c); err != nil {
		return
	}

	// Parse query parameters
	withBookmarkCount := c.Request().URL.Query().Get("with_bookmark_count") == "true"
	search := c.Request().URL.Query().Get("search")

	var bookmarkID int
	if bookmarkIDStr := c.Request().URL.Query().Get("bookmark_id"); bookmarkIDStr != "" {
		var err error
		bookmarkID, err = strconv.Atoi(bookmarkIDStr)
		if err != nil {
			response.SendError(c, http.StatusBadRequest, "Invalid bookmark ID")
			return
		}
	}

	// Create options and validate
	opts := model.ListTagsOptions{
		WithBookmarkCount: withBookmarkCount,
		BookmarkID:        bookmarkID,
		OrderBy:           model.DBTagOrderByTagName,
		Search:            search,
	}

	if err := opts.IsValid(); err != nil {
		response.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	tags, err := deps.Domains().Tags().ListTags(c.Request().Context(), opts)
	if err != nil {
		deps.Logger().WithError(err).Error("failed to get tags")
		response.SendInternalServerError(c)
		return
	}

	response.SendJSON(c, http.StatusOK, tags)
}

// @Summary					Get tag
// @Description				Get a tag by ID
// @Tags						Tags
// @securityDefinitions.apikey	ApiKeyAuth
// @Produce					json
// @Param						id	path		int	true	"Tag ID"
// @Success					200	{object}	model.TagDTO
// @Failure					403	{object}	nil	"Authentication required"
// @Failure					404	{object}	nil	"Tag not found"
// @Failure					500	{object}	nil	"Internal server error"
// @Router						/api/v1/tags/{id} [get]
func HandleGetTag(deps model.Dependencies, c model.WebContext) {
	if err := middleware.RequireLoggedInUser(deps, c); err != nil {
		return
	}

	idParam := c.Request().PathValue("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		response.SendError(c, http.StatusBadRequest, "Invalid tag ID")
		return
	}

	tag, err := deps.Domains().Tags().GetTag(c.Request().Context(), id)
	if err != nil {
		if err == model.ErrNotFound {
			response.NotFound(c)
			return
		}
		deps.Logger().WithError(err).Error("failed to get tag")
		response.SendInternalServerError(c)
		return
	}

	response.SendJSON(c, http.StatusOK, tag)
}

// @Summary					Create tag
// @Description				Create a new tag
// @Tags						Tags
// @securityDefinitions.apikey	ApiKeyAuth
// @Accept						json
// @Produce					json
// @Param						tag	body		model.TagDTO	true	"Tag data"
// @Success					201	{object}	model.TagDTO
// @Failure					400	{object}	nil	"Invalid request"
// @Failure					403	{object}	nil	"Authentication required"
// @Failure					500	{object}	nil	"Internal server error"
// @Router						/api/v1/tags [post]
func HandleCreateTag(deps model.Dependencies, c model.WebContext) {
	if err := middleware.RequireLoggedInUser(deps, c); err != nil {
		return
	}

	var tag model.TagDTO
	err := json.NewDecoder(c.Request().Body).Decode(&tag)
	if err != nil {
		response.SendError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	if tag.Name == "" {
		response.SendError(c, http.StatusBadRequest, "Tag name is required")
		return
	}

	createdTag, err := deps.Domains().Tags().CreateTag(c.Request().Context(), tag)
	if err != nil {
		deps.Logger().WithError(err).Error("failed to create tag")
		response.SendInternalServerError(c)
		return
	}

	response.SendJSON(c, http.StatusCreated, createdTag)
}

// @Summary					Update tag
// @Description				Update an existing tag
// @Tags						Tags
// @securityDefinitions.apikey	ApiKeyAuth
// @Accept						json
// @Produce					json
// @Param						id	path		int				true	"Tag ID"
// @Param						tag	body		model.TagDTO	true	"Tag data"
// @Success					200	{object}	model.TagDTO
// @Failure					400	{object}	nil	"Invalid request"
// @Failure					403	{object}	nil	"Authentication required"
// @Failure					404	{object}	nil	"Tag not found"
// @Failure					500	{object}	nil	"Internal server error"
// @Router						/api/v1/tags/{id} [put]
func HandleUpdateTag(deps model.Dependencies, c model.WebContext) {
	if err := middleware.RequireLoggedInUser(deps, c); err != nil {
		return
	}

	idParam := c.Request().PathValue("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		response.SendError(c, http.StatusBadRequest, "Invalid tag ID")
		return
	}

	var tag model.TagDTO
	err = json.NewDecoder(c.Request().Body).Decode(&tag)
	if err != nil {
		response.SendError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	if tag.Name == "" {
		response.SendError(c, http.StatusBadRequest, "Tag name is required")
		return
	}

	// Ensure the ID in the URL matches the ID in the body
	tag.ID = id

	updatedTag, err := deps.Domains().Tags().UpdateTag(c.Request().Context(), tag)
	if err != nil {
		if err == model.ErrNotFound {
			response.NotFound(c)
			return
		}
		deps.Logger().WithError(err).Error("failed to update tag")
		response.SendInternalServerError(c)
		return
	}

	response.SendJSON(c, http.StatusOK, updatedTag)
}

// @Summary					Delete tag
// @Description				Delete a tag
// @Tags						Tags
// @securityDefinitions.apikey	ApiKeyAuth
// @Param						id	path		int	true	"Tag ID"
// @Success					204	{object}	nil
// @Failure					403	{object}	nil	"Authentication required"
// @Failure					404	{object}	nil	"Tag not found"
// @Failure					500	{object}	nil	"Internal server error"
// @Router						/api/v1/tags/{id} [delete]
func HandleDeleteTag(deps model.Dependencies, c model.WebContext) {
	if err := middleware.RequireLoggedInAdmin(deps, c); err != nil {
		return
	}

	idParam := c.Request().PathValue("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		response.SendError(c, http.StatusBadRequest, "Invalid tag ID")
		return
	}

	err = deps.Domains().Tags().DeleteTag(c.Request().Context(), id)
	if err != nil {
		if err == model.ErrNotFound {
			response.NotFound(c)
			return
		}
		deps.Logger().WithError(err).Error("failed to delete tag")
		response.SendInternalServerError(c)
		return
	}

	response.SendJSON(c, http.StatusNoContent, nil)
}
