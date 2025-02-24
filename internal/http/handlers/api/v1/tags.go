package api_v1

import (
	"net/http"

	"github.com/go-shiori/shiori/internal/http/middleware"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
)

// @Summary					List tags
// @Description				List all tags
// @Tags						Tags
// @securityDefinitions.apikey	ApiKeyAuth
// @Produce					json
// @Success					200	{array}		model.Tag
// @Failure					403	{object}	nil	"Authentication required"
// @Failure					500	{object}	nil	"Internal server error"
// @Router						/api/v1/tags [get]
func HandleListTags(deps model.Dependencies, c model.WebContext) {
	if err := middleware.RequireLoggedInUser(deps, c); err != nil {
		return
	}

	tags, err := deps.Domains().Tags().ListTags(c.Request().Context())
	if err != nil {
		deps.Logger().WithError(err).Error("failed to get tags")
		response.SendInternalServerError(c)
		return
	}

	response.Send(c, http.StatusOK, tags)
}
