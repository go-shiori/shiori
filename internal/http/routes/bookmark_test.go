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

func TestBookmarkContentHandler(t *testing.T) {
	logger := logrus.New()

	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.Background(), logger)

	bookmark := testutil.GetValidBookmark()
	bookmark.HTML = "<html><body><h1>Bookmark HTML content</h1></body></html>"
	boomkarks, err := deps.Database.SaveBookmarks(context.TODO(), true, *bookmark)
	require.NoError(t, err)

	bookmark = &boomkarks[0]

	t.Run("not logged in", func(t *testing.T) {
		g := gin.Default()
		router := NewBookmarkRoutes(logger, deps)
		router.Setup(g.Group("/"))
		w := httptest.NewRecorder()
		path := "/" + strconv.Itoa(bookmark.ID) + "/content"
		req, _ := http.NewRequest("GET", path, nil)
		g.ServeHTTP(w, req)
		require.Equal(t, http.StatusFound, w.Code)
		require.Equal(t, "/login?dst="+path, w.Header().Get("Location"))
	})

	t.Run("get existing bookmark content", func(t *testing.T) {
		g := gin.Default()
		templates.SetupTemplates(g)
		g.Use(func(c *gin.Context) {
			c.Set(model.ContextAccountKey, "test")
		})
		router := NewBookmarkRoutes(logger, deps)
		router.Setup(g.Group("/"))
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/"+strconv.Itoa(bookmark.ID)+"/content", nil)
		g.ServeHTTP(w, req)
		t.Log(w.Header().Get("Location"))
		require.Equal(t, 200, w.Code)
		require.Contains(t, w.Body.String(), bookmark.HTML)
	})
}

func TestBookmarkFileHandlers(t *testing.T) {
	logger := logrus.New()

	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.Background(), logger)

	bookmark := testutil.GetValidBookmark()
	bookmark.HTML = "<html><body><h1>Bookmark HTML content</h1></body></html>"
	bookmark.HasArchive = true
	bookmark.CreateArchive = true
	bookmark.CreateEbook = true
	bookmarks, err := deps.Database.SaveBookmarks(context.TODO(), true, *bookmark)
	require.NoError(t, err)

	bookmark, err = deps.Domains.Archiver.DownloadBookmarkArchive(bookmarks[0])
	require.NoError(t, err)

	bookmarks, err = deps.Database.SaveBookmarks(context.TODO(), false, *bookmark)
	require.NoError(t, err)
	bookmark = &bookmarks[0]

	g := gin.Default()
	templates.SetupTemplates(g)
	g.Use(func(c *gin.Context) {
		c.Set(model.ContextAccountKey, "test")
	})
	router := NewBookmarkRoutes(logger, deps)
	router.Setup(g.Group("/"))

	t.Run("get existing bookmark archive", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/"+strconv.Itoa(bookmark.ID)+"/archive", nil)
		g.ServeHTTP(w, req)
		require.Contains(t, w.Body.String(), "iframe")
		require.Equal(t, 200, w.Code)
	})

	t.Run("get existing bookmark thumbnail", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/"+strconv.Itoa(bookmark.ID)+"/thumb", nil)
		g.ServeHTTP(w, req)
		require.Equal(t, 200, w.Code)
	})

	t.Run("bookmark without archive", func(t *testing.T) {
		bookmark := testutil.GetValidBookmark()
		bookmarks, err := deps.Database.SaveBookmarks(context.TODO(), true, *bookmark)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/"+strconv.Itoa(bookmarks[0].ID)+"/archive", nil)
		g.ServeHTTP(w, req)
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("get existing bookmark archive file", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/"+strconv.Itoa(bookmark.ID)+"/archive/file/", nil)
		g.ServeHTTP(w, req)
		require.Equal(t, 200, w.Code)
	})

	t.Run("bookmark with ebook", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/"+strconv.Itoa(bookmarks[0].ID)+"/ebook", nil)
		g.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("bookmark without ebook", func(t *testing.T) {
		bookmark := testutil.GetValidBookmark()
		bookmarks, err := deps.Database.SaveBookmarks(context.TODO(), true, *bookmark)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/"+strconv.Itoa(bookmarks[0].ID)+"/ebook", nil)
		g.ServeHTTP(w, req)
		require.Equal(t, http.StatusNotFound, w.Code)
	})
}
