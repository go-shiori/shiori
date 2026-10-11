package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/go-shiori/shiori/e2e/e2eutil"
	"github.com/stretchr/testify/require"
)

// apiRequest is a helper to make authenticated API requests and return the response.
func apiRequest(t *testing.T, method, url, token string, body []byte) *http.Response {
	t.Helper()
	var req *http.Request
	var err error
	if body != nil {
		req, err = http.NewRequest(method, url, bytes.NewReader(body))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

// bookmarkResponse represents the bookmark fields we care about in tests.
type bookmarkResponse struct {
	ID      int    `json:"ID"`
	URL     string `json:"URL"`
	Title   string `json:"Title"`
	Excerpt string `json:"Excerpt"`
	Public  int    `json:"Public"`
}

func TestBookmarkCreation(t *testing.T) {
	testServerURL := e2eutil.StartTestServer(t)

	container := e2eutil.NewShioriContainer(t, "")
	baseURL := "http://localhost:" + container.GetPort()
	token := apiLogin(t, baseURL, "shiori", "gopher")

	t.Run("create bookmark with URL only", func(t *testing.T) {
		body := fmt.Sprintf(`{"URL":"%s/index.html"}`, testServerURL)
		resp := apiRequest(t, "POST", baseURL+"/api/v1/bookmarks", token, []byte(body))
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var bm bookmarkResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&bm))
		require.NotZero(t, bm.ID)
		require.Equal(t, testServerURL+"/index.html", bm.URL)
	})

	t.Run("create bookmark with custom title and excerpt", func(t *testing.T) {
		body := fmt.Sprintf(`{"URL":"%s/custom.html","Title":"My Title","Excerpt":"My Excerpt"}`, testServerURL)
		resp := apiRequest(t, "POST", baseURL+"/api/v1/bookmarks", token, []byte(body))
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var bm bookmarkResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&bm))
		require.Equal(t, "My Title", bm.Title)
		require.Equal(t, "My Excerpt", bm.Excerpt)
	})

	t.Run("create bookmark with archive option", func(t *testing.T) {
		body := fmt.Sprintf(`{"URL":"%s/archive.html","Options":{"CreateArchive":true}}`, testServerURL)
		resp := apiRequest(t, "POST", baseURL+"/api/v1/bookmarks", token, []byte(body))
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var bm bookmarkResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&bm))
		require.NotZero(t, bm.ID)
	})

	t.Run("create bookmark with ebook option", func(t *testing.T) {
		body := fmt.Sprintf(`{"URL":"%s/ebook.html","Options":{"CreateEbook":true}}`, testServerURL)
		resp := apiRequest(t, "POST", baseURL+"/api/v1/bookmarks", token, []byte(body))
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var bm bookmarkResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&bm))
		require.NotZero(t, bm.ID)
	})

	t.Run("create bookmark and assign tags", func(t *testing.T) {
		// Create bookmark
		body := fmt.Sprintf(`{"URL":"%s/tagged.html"}`, testServerURL)
		resp := apiRequest(t, "POST", baseURL+"/api/v1/bookmarks", token, []byte(body))
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var bm bookmarkResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&bm))

		// Create a tag
		tagBody := `{"Name":"test-tag"}`
		tagResp := apiRequest(t, "POST", baseURL+"/api/v1/tags", token, []byte(tagBody))
		defer tagResp.Body.Close()
		require.Equal(t, http.StatusCreated, tagResp.StatusCode)

		var tag struct {
			ID int `json:"ID"`
		}
		require.NoError(t, json.NewDecoder(tagResp.Body).Decode(&tag))
		require.NotZero(t, tag.ID)

		// Add tag to bookmark
		addTagBody := fmt.Sprintf(`{"TagID":%d}`, tag.ID)
		addTagResp := apiRequest(t, "POST", fmt.Sprintf("%s/api/v1/bookmarks/%d/tags", baseURL, bm.ID), token, []byte(addTagBody))
		defer addTagResp.Body.Close()
		require.Equal(t, http.StatusCreated, addTagResp.StatusCode)

		// Verify tag is on the bookmark
		getTagsResp := apiRequest(t, "GET", fmt.Sprintf("%s/api/v1/bookmarks/%d/tags", baseURL, bm.ID), token, nil)
		defer getTagsResp.Body.Close()
		require.Equal(t, http.StatusOK, getTagsResp.StatusCode)

		var tags []struct {
			ID   int    `json:"ID"`
			Name string `json:"Name"`
		}
		require.NoError(t, json.NewDecoder(getTagsResp.Body).Decode(&tags))
		require.Len(t, tags, 1)
		require.Equal(t, "test-tag", tags[0].Name)
	})

	t.Run("create bookmark marked as public", func(t *testing.T) {
		body := fmt.Sprintf(`{"URL":"%s/public.html","Public":1}`, testServerURL)
		resp := apiRequest(t, "POST", baseURL+"/api/v1/bookmarks", token, []byte(body))
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var bm bookmarkResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&bm))
		require.Equal(t, 1, bm.Public, "Bookmark should be public")
	})

	t.Run("created bookmark appears in library", func(t *testing.T) {
		// Create a bookmark with a unique URL
		body := fmt.Sprintf(`{"URL":"%s/index.html?lib=1","Title":"Library Test"}`, testServerURL)
		createResp := apiRequest(t, "POST", baseURL+"/api/v1/bookmarks", token, []byte(body))
		defer createResp.Body.Close()
		require.Equal(t, http.StatusCreated, createResp.StatusCode)

		var created bookmarkResponse
		require.NoError(t, json.NewDecoder(createResp.Body).Decode(&created))

		// List bookmarks and verify it appears
		listResp := apiRequest(t, "GET", baseURL+"/api/v1/bookmarks", token, nil)
		defer listResp.Body.Close()
		require.Equal(t, http.StatusOK, listResp.StatusCode)

		var listResult struct {
			Items []bookmarkResponse `json:"Items"`
			Total int                `json:"Total"`
		}
		require.NoError(t, json.NewDecoder(listResp.Body).Decode(&listResult))

		found := false
		for _, bm := range listResult.Items {
			if bm.ID == created.ID {
				found = true
				break
			}
		}
		require.True(t, found, "Created bookmark should appear in the library listing")
	})

	t.Run("empty URL returns error", func(t *testing.T) {
		body := `{"URL":""}`
		resp := apiRequest(t, "POST", baseURL+"/api/v1/bookmarks", token, []byte(body))
		defer resp.Body.Close()
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("missing URL returns error", func(t *testing.T) {
		body := `{}`
		resp := apiRequest(t, "POST", baseURL+"/api/v1/bookmarks", token, []byte(body))
		defer resp.Body.Close()
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestBookmarkBulkTags(t *testing.T) {
	testServerURL := e2eutil.StartTestServer(t)

	container := e2eutil.NewShioriContainer(t, "")
	baseURL := "http://localhost:" + container.GetPort()
	token := apiLogin(t, baseURL, "shiori", "gopher")

	createBookmark := func(path string) int {
		body := fmt.Sprintf(`{"URL":"%s%s"}`, testServerURL, path)
		resp := apiRequest(t, "POST", baseURL+"/api/v1/bookmarks", token, []byte(body))
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var bm bookmarkResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&bm))
		return bm.ID
	}

	createTag := func(name string) int {
		resp := apiRequest(t, "POST", baseURL+"/api/v1/tags", token, []byte(fmt.Sprintf(`{"Name":%q}`, name)))
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var tag struct {
			ID int `json:"ID"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&tag))
		return tag.ID
	}

	bulkUpdate := func(bookmarkIDs, tagIDs []int) {
		body, err := json.Marshal(map[string][]int{"BookmarkIDs": bookmarkIDs, "TagIDs": tagIDs})
		require.NoError(t, err)
		resp := apiRequest(t, "PUT", baseURL+"/api/v1/bookmarks/bulk/tags", token, body)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	}

	tagNames := func(bookmarkID int) []string {
		resp := apiRequest(t, "GET", fmt.Sprintf("%s/api/v1/bookmarks/%d/tags", baseURL, bookmarkID), token, nil)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var tags []struct {
			Name string `json:"Name"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&tags))
		names := make([]string, len(tags))
		for i, tag := range tags {
			names[i] = tag.Name
		}
		return names
	}

	first := createBookmark("/index.html?bulk=1")
	second := createBookmark("/index.html?bulk=2")
	alpha := createTag("bulk-alpha")
	beta := createTag("bulk-beta")

	t.Run("sets tags on multiple bookmarks", func(t *testing.T) {
		bulkUpdate([]int{first, second}, []int{alpha})
		require.ElementsMatch(t, []string{"bulk-alpha"}, tagNames(first))
		require.ElementsMatch(t, []string{"bulk-alpha"}, tagNames(second))
	})

	t.Run("replaces existing tags", func(t *testing.T) {
		bulkUpdate([]int{first, second}, []int{alpha})
		bulkUpdate([]int{first}, []int{beta})
		require.ElementsMatch(t, []string{"bulk-beta"}, tagNames(first))
		require.ElementsMatch(t, []string{"bulk-alpha"}, tagNames(second), "Other bookmarks must not change")
	})
}
