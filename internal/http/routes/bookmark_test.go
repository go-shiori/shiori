package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	sctx "github.com/go-shiori/shiori/internal/http/context"
	"github.com/go-shiori/shiori/internal/http/templates"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestBookmarkRoutesGetBookmark(t *testing.T) {
	logger := logrus.New()

	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.TODO(), logger)

	g := gin.Default()
	templates.SetupTemplates(g)
	w := httptest.NewRecorder()

	// Create a private and a public bookmark to use in tests
	publicBookmark := testutil.GetValidBookmark()
	publicBookmark.Public = 1
	bookmarks, err := deps.Database.SaveBookmarks(context.TODO(), true, []model.BookmarkDTO{
		*testutil.GetValidBookmark(),
		*publicBookmark,
	}...)

	require.NoError(t, err)

	router := NewBookmarkRoutes(logger, deps)

	t.Run("bookmark ID is not present", func(t *testing.T) {
		gctx := gin.CreateTestContextOnly(w, g)
		c := sctx.NewContextFromGin(gctx)
		_, err := router.getBookmark(c)
		require.Error(t, err)
		require.Equal(t, http.StatusBadRequest, c.Writer.Status())
	})

	t.Run("bookmark ID is not parsable number", func(t *testing.T) {
		gctx := gin.CreateTestContextOnly(w, g)
		c := sctx.NewContextFromGin(gctx)
		c.Params = append(c.Params, gin.Param{Key: "id", Value: "not a number"})
		_, err := router.getBookmark(c)
		require.Error(t, err)
		require.Equal(t, http.StatusInternalServerError, c.Writer.Status())
	})

	t.Run("bookmark ID does not exist", func(t *testing.T) {
		gctx := gin.CreateTestContextOnly(w, g)
		c := sctx.NewContextFromGin(gctx)
		c.Params = append(c.Params, gin.Param{Key: "id", Value: "99"})
		bookmark, err := router.getBookmark(c)
		require.Equal(t, http.StatusNotFound, c.Writer.Status())
		require.Nil(t, bookmark)
		require.Error(t, err)
	})

	t.Run("bookmark ID exists but user is not logged in", func(t *testing.T) {
		gctx := gin.CreateTestContextOnly(w, g)
		c := sctx.NewContextFromGin(gctx)
		c.Request = httptest.NewRequest(http.MethodGet, "/bookmark/1", nil)
		c.Params = append(c.Params, gin.Param{Key: "id", Value: "1"})
		bookmark, err := router.getBookmark(c)
		require.Equal(t, http.StatusFound, c.Writer.Status())
		require.Nil(t, bookmark)
		require.Error(t, err)
	})

	t.Run("bookmark ID exists and its public and user is not logged in", func(t *testing.T) {
		gctx := gin.CreateTestContextOnly(w, g)
		c := sctx.NewContextFromGin(gctx)
		c.Request = httptest.NewRequest(http.MethodGet, "/bookmark/"+strconv.Itoa(bookmarks[0].ID), nil)
		c.Params = append(c.Params, gin.Param{Key: "id", Value: strconv.Itoa(bookmarks[1].ID)})
		bookmark, err := router.getBookmark(c)
		require.Equal(t, http.StatusOK, c.Writer.Status())
		require.NotNil(t, bookmark)
		require.NoError(t, err)
	})

	t.Run("bookmark ID exists and user is logged in", func(t *testing.T) {
		g := gin.Default()
		templates.SetupTemplates(g)
		gctx := gin.CreateTestContextOnly(w, g)
		c := sctx.NewContextFromGin(gctx)
		c.Set(model.ContextAccountKey, &model.Account{})
		c.Request = httptest.NewRequest(http.MethodGet, "/bookmark/"+strconv.Itoa(bookmarks[0].ID), nil)
		c.Params = append(c.Params, gin.Param{Key: "id", Value: strconv.Itoa(bookmarks[0].ID)})
		bookmark, err := router.getBookmark(c)
		require.Equal(t, http.StatusOK, c.Writer.Status())
		require.NotNil(t, bookmark)
		require.NoError(t, err)
	})
}
