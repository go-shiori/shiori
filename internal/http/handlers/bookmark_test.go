package handlers

import (
	"context"
	"net/http"
	"strconv"
	"testing"

	"github.com/go-shiori/shiori/internal/http/templates"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestGetBookmark(t *testing.T) {
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.TODO(), logger)

	err := templates.SetupTemplates()
	require.NoError(t, err)

	// Create a private and a public bookmark to use in tests
	publicBookmark := testutil.GetValidBookmark()
	publicBookmark.Public = 1
	bookmarks, err := deps.Database().SaveBookmarks(context.TODO(), true, []model.BookmarkDTO{
		*testutil.GetValidBookmark(),
		*publicBookmark,
	}...)
	require.NoError(t, err)

	t.Run("bookmark ID is not parsable number", func(t *testing.T) {
		c, w := testutil.NewTestWebContextWithMethod("GET", "/bookmark/notanumber")
		testutil.SetRequestPathValue(c, "id", "notanumber")
		bookmark, _ := getBookmark(deps, c)
		require.Nil(t, bookmark)
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("bookmark ID does not exist", func(t *testing.T) {
		c, w := testutil.NewTestWebContextWithMethod("GET", "/bookmark/99999")
		testutil.SetRequestPathValue(c, "id", "99999")
		bookmark, _ := getBookmark(deps, c)
		require.Nil(t, bookmark)
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("bookmark ID exists but user is not logged in", func(t *testing.T) {
		c, _ := testutil.NewTestWebContextWithMethod("GET", "/bookmark/"+strconv.Itoa(bookmarks[0].ID))
		testutil.SetRequestPathValue(c, "id", strconv.Itoa(bookmarks[0].ID))
		bookmark, _ := getBookmark(deps, c)
		require.NoError(t, err) // No error because it redirects
		require.Nil(t, bookmark)
	})

	t.Run("bookmark ID exists and its public and user is not logged in", func(t *testing.T) {
		c, _ := testutil.NewTestWebContextWithMethod("GET", "/bookmark/"+strconv.Itoa(bookmarks[1].ID))
		testutil.SetRequestPathValue(c, "id", strconv.Itoa(bookmarks[1].ID))
		bookmark, _ := getBookmark(deps, c)
		require.NoError(t, err)
		require.NotNil(t, bookmark)
	})

	t.Run("bookmark ID exists and user is logged in", func(t *testing.T) {
		c, _ := testutil.NewTestWebContextWithMethod("GET", "/bookmark/"+strconv.Itoa(bookmarks[0].ID)+"/content")
		testutil.SetFakeUser(c)
		testutil.SetRequestPathValue(c, "id", strconv.Itoa(bookmarks[0].ID))
		bookmark, _ := getBookmark(deps, c)
		require.NoError(t, err)
		require.NotNil(t, bookmark)
	})
}

func TestBookmarkContentHandler(t *testing.T) {
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.Background(), logger)

	err := templates.SetupTemplates()
	require.NoError(t, err)

	bookmark := testutil.GetValidBookmark()
	bookmark.HTML = "<html><body><h1>Bookmark HTML content</h1></body></html>"
	bookmarks, err := deps.Database().SaveBookmarks(context.TODO(), true, *bookmark)
	require.NoError(t, err)
	bookmark = &bookmarks[0]

	t.Run("not logged in", func(t *testing.T) {
		c, w := testutil.NewTestWebContextWithMethod("GET", "/bookmark/"+strconv.Itoa(bookmark.ID)+"/content")
		testutil.SetRequestPathValue(c, "id", strconv.Itoa(bookmark.ID))
		HandleBookmarkContent(deps, c)
		require.Equal(t, http.StatusFound, w.Code) // Redirects to login
	})

	t.Run("get existing bookmark content", func(t *testing.T) {
		c, w := testutil.NewTestWebContextWithMethod("GET", "/bookmark/"+strconv.Itoa(bookmark.ID)+"/content")
		testutil.SetFakeUser(c)
		testutil.SetRequestPathValue(c, "id", strconv.Itoa(bookmark.ID))
		HandleBookmarkContent(deps, c)
		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), bookmark.HTML)
	})
}

func TestBookmarkFileHandlers(t *testing.T) {
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.Background(), logger)

	err := templates.SetupTemplates()
	require.NoError(t, err)

	bookmark := testutil.GetValidBookmark()
	bookmark.HTML = "<html><body><h1>Bookmark HTML content</h1></body></html>"
	bookmark.HasArchive = true
	bookmark.CreateArchive = true
	bookmark.CreateEbook = true
	bookmarks, err := deps.Database().SaveBookmarks(context.TODO(), true, *bookmark)
	require.NoError(t, err)

	bookmark, err = deps.Domains().Archiver().DownloadBookmarkArchive(bookmarks[0])
	require.NoError(t, err)

	bookmarks, err = deps.Database().SaveBookmarks(context.TODO(), false, *bookmark)
	require.NoError(t, err)
	bookmark = &bookmarks[0]

	t.Run("get existing bookmark archive", func(t *testing.T) {
		c, w := testutil.NewTestWebContextWithMethod("GET", "/bookmark/"+strconv.Itoa(bookmark.ID)+"/archive")
		testutil.SetFakeUser(c)
		testutil.SetRequestPathValue(c, "id", strconv.Itoa(bookmark.ID))
		HandleBookmarkArchive(deps, c)
		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), "iframe")
	})

	t.Run("get existing bookmark thumbnail", func(t *testing.T) {
		c, w := testutil.NewTestWebContextWithMethod("GET", "/bookmark/"+strconv.Itoa(bookmark.ID)+"/thumb")
		testutil.SetFakeUser(c)
		testutil.SetRequestPathValue(c, "id", strconv.Itoa(bookmark.ID))
		HandleBookmarkThumbnail(deps, c)
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("bookmark without archive", func(t *testing.T) {
		bookmark := testutil.GetValidBookmark()
		bookmarks, err := deps.Database().SaveBookmarks(context.TODO(), true, *bookmark)
		require.NoError(t, err)

		c, w := testutil.NewTestWebContextWithMethod("GET", "/bookmark/"+strconv.Itoa(bookmarks[0].ID)+"/archive")
		testutil.SetFakeUser(c)
		testutil.SetRequestPathValue(c, "id", strconv.Itoa(bookmarks[0].ID))
		HandleBookmarkArchive(deps, c)
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("get existing bookmark archive file", func(t *testing.T) {
		c, w := testutil.NewTestWebContextWithMethod("GET", "/bookmark/"+strconv.Itoa(bookmark.ID)+"/archive/file/")
		testutil.SetFakeUser(c)
		testutil.SetRequestPathValue(c, "id", strconv.Itoa(bookmark.ID))
		HandleBookmarkArchiveFile(deps, c)
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("bookmark with ebook", func(t *testing.T) {
		c, w := testutil.NewTestWebContextWithMethod("GET", "/bookmark/"+strconv.Itoa(bookmark.ID)+"/ebook")
		testutil.SetFakeUser(c)
		testutil.SetRequestPathValue(c, "id", strconv.Itoa(bookmark.ID))
		HandleBookmarkEbook(deps, c)
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("bookmark without ebook", func(t *testing.T) {
		bookmark := testutil.GetValidBookmark()
		bookmarks, err := deps.Database().SaveBookmarks(context.TODO(), true, *bookmark)
		require.NoError(t, err)

		c, w := testutil.NewTestWebContextWithMethod("GET", "/bookmark/"+strconv.Itoa(bookmarks[0].ID)+"/ebook")
		testutil.SetFakeUser(c)
		testutil.SetRequestPathValue(c, "id", strconv.Itoa(bookmarks[0].ID))
		HandleBookmarkEbook(deps, c)
		require.Equal(t, http.StatusNotFound, w.Code)
	})
}
