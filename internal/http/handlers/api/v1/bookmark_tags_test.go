package api_v1_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	api_v1 "github.com/go-shiori/shiori/internal/http/handlers/api/v1"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Define the BookmarkTagPayload struct to match the one in the API
type bookmarkTagPayload struct {
	TagID int `json:"tag_id"`
}

func TestBookmarkTagsAPI(t *testing.T) {
	ctx := context.Background()
	logger := logrus.New()

	// Setup using the test configuration and dependencies
	_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
	db := deps.Database()

	// Create a test bookmark
	bookmark := model.BookmarkDTO{
		URL:   "https://example.com/api-tags-test",
		Title: "API Tags Test",
	}
	savedBookmarks, err := db.SaveBookmarks(ctx, true, bookmark)
	require.NoError(t, err)
	require.Len(t, savedBookmarks, 1)
	bookmarkID := savedBookmarks[0].ID

	// Create a test tag
	tag := model.Tag{
		Name: "api-test-tag",
	}
	createdTags, err := db.CreateTags(ctx, tag)
	require.NoError(t, err)
	require.Len(t, createdTags, 1)
	tagID := createdTags[0].ID

	// Test authentication requirements
	t.Run("AuthenticationRequirements", func(t *testing.T) {
		// Test unauthenticated user for GetBookmarkTags
		t.Run("UnauthenticatedUserGetTags", func(t *testing.T) {
			rec := testutil.PerformRequest(
				deps,
				api_v1.HandleGetBookmarkTags,
				http.MethodGet,
				"/api/v1/bookmarks/"+strconv.Itoa(bookmarkID)+"/tags",
				testutil.WithRequestPathValue("id", strconv.Itoa(bookmarkID)),
			)

			require.Equal(t, http.StatusUnauthorized, rec.Code)
		})

		// Test unauthenticated user for AddTagToBookmark
		t.Run("UnauthenticatedUserAddTag", func(t *testing.T) {
			payload := bookmarkTagPayload{
				TagID: tagID,
			}
			payloadBytes, err := json.Marshal(payload)
			require.NoError(t, err)

			rec := testutil.PerformRequest(
				deps,
				api_v1.HandleAddTagToBookmark,
				http.MethodPost,
				"/api/v1/bookmarks/"+strconv.Itoa(bookmarkID)+"/tags",
				testutil.WithRequestPathValue("id", strconv.Itoa(bookmarkID)),
				testutil.WithBody(string(payloadBytes)),
			)

			require.Equal(t, http.StatusUnauthorized, rec.Code)
		})

		// Test non-admin user for AddTagToBookmark (which requires admin)
		t.Run("NonAdminUserAddTag", func(t *testing.T) {
			payload := bookmarkTagPayload{
				TagID: tagID,
			}
			payloadBytes, err := json.Marshal(payload)
			require.NoError(t, err)

			rec := testutil.PerformRequest(
				deps,
				api_v1.HandleAddTagToBookmark,
				http.MethodPost,
				"/api/v1/bookmarks/"+strconv.Itoa(bookmarkID)+"/tags",
				testutil.WithFakeUser(), // Regular user, not admin
				testutil.WithRequestPathValue("id", strconv.Itoa(bookmarkID)),
				testutil.WithBody(string(payloadBytes)),
			)

			// Just check the status code since the response might vary
			require.Equal(t, http.StatusForbidden, rec.Code)
		})

		// Test unauthenticated user for RemoveTagFromBookmark
		t.Run("UnauthenticatedUserRemoveTag", func(t *testing.T) {
			payload := bookmarkTagPayload{
				TagID: tagID,
			}
			payloadBytes, err := json.Marshal(payload)
			require.NoError(t, err)

			rec := testutil.PerformRequest(
				deps,
				api_v1.HandleRemoveTagFromBookmark,
				http.MethodDelete,
				"/api/v1/bookmarks/"+strconv.Itoa(bookmarkID)+"/tags",
				testutil.WithRequestPathValue("id", strconv.Itoa(bookmarkID)),
				testutil.WithBody(string(payloadBytes)),
			)

			require.Equal(t, http.StatusUnauthorized, rec.Code)
		})
	})

	// Test BulkUpdateBookmarkTags
	t.Run("BulkUpdateBookmarkTags", func(t *testing.T) {
		// Define the payload struct
		type bulkUpdatePayload struct {
			BookmarkIDs []int `json:"bookmark_ids"`
			TagIDs      []int `json:"tag_ids"`
		}

		// Test successful bulk update
		t.Run("SuccessfulBulkUpdate", func(t *testing.T) {
			payload := bulkUpdatePayload{
				BookmarkIDs: []int{bookmarkID},
				TagIDs:      []int{tagID},
			}
			payloadBytes, err := json.Marshal(payload)
			require.NoError(t, err)

			rec := testutil.PerformRequest(
				deps,
				api_v1.HandleBulkUpdateBookmarkTags,
				http.MethodPut,
				"/api/v1/bookmarks/bulk/tags",
				testutil.WithFakeAdmin(),
				testutil.WithBody(string(payloadBytes)),
			)

			require.Equal(t, http.StatusOK, rec.Code)

			testResp := testutil.NewTestResponseFromRecorder(rec)
			testResp.AssertOk(t)
		})

		// Test unauthenticated user
		t.Run("UnauthenticatedUser", func(t *testing.T) {
			payload := bulkUpdatePayload{
				BookmarkIDs: []int{bookmarkID},
				TagIDs:      []int{tagID},
			}
			payloadBytes, err := json.Marshal(payload)
			require.NoError(t, err)

			rec := testutil.PerformRequest(
				deps,
				api_v1.HandleBulkUpdateBookmarkTags,
				http.MethodPut,
				"/api/v1/bookmarks/bulk/tags",
				testutil.WithBody(string(payloadBytes)),
			)

			require.Equal(t, http.StatusUnauthorized, rec.Code)
		})

		// Test invalid request payload
		t.Run("InvalidRequestPayload", func(t *testing.T) {
			invalidPayload := []byte(`{"bookmark_ids": "invalid", "tag_ids": [1]}`)

			rec := testutil.PerformRequest(
				deps,
				api_v1.HandleBulkUpdateBookmarkTags,
				http.MethodPut,
				"/api/v1/bookmarks/bulk/tags",
				testutil.WithFakeAdmin(),
				testutil.WithBody(string(invalidPayload)),
			)

			require.Equal(t, http.StatusBadRequest, rec.Code)

			testResp := testutil.NewTestResponseFromRecorder(rec)
			testResp.AssertNotOk(t)
			testResp.AssertMessageJSONKeyValue(t, "error", func(t *testing.T, value any) {
				require.Equal(t, "Invalid request payload", value)
			})
		})

		// Test empty bookmark IDs
		t.Run("EmptyBookmarkIDs", func(t *testing.T) {
			payload := bulkUpdatePayload{
				BookmarkIDs: []int{},
				TagIDs:      []int{tagID},
			}
			payloadBytes, err := json.Marshal(payload)
			require.NoError(t, err)

			rec := testutil.PerformRequest(
				deps,
				api_v1.HandleBulkUpdateBookmarkTags,
				http.MethodPut,
				"/api/v1/bookmarks/bulk/tags",
				testutil.WithFakeAdmin(),
				testutil.WithBody(string(payloadBytes)),
			)

			require.Equal(t, http.StatusBadRequest, rec.Code)

			testResp := testutil.NewTestResponseFromRecorder(rec)
			testResp.AssertNotOk(t)
			testResp.AssertMessageJSONKeyValue(t, "error", func(t *testing.T, value any) {
				require.Equal(t, "bookmark_ids should not be empty", value)
			})
		})

		// Test empty tag IDs
		t.Run("EmptyTagIDs", func(t *testing.T) {
			payload := bulkUpdatePayload{
				BookmarkIDs: []int{bookmarkID},
				TagIDs:      []int{},
			}
			payloadBytes, err := json.Marshal(payload)
			require.NoError(t, err)

			rec := testutil.PerformRequest(
				deps,
				api_v1.HandleBulkUpdateBookmarkTags,
				http.MethodPut,
				"/api/v1/bookmarks/bulk/tags",
				testutil.WithFakeAdmin(),
				testutil.WithBody(string(payloadBytes)),
			)

			require.Equal(t, http.StatusBadRequest, rec.Code)

			testResp := testutil.NewTestResponseFromRecorder(rec)
			testResp.AssertNotOk(t)
			testResp.AssertMessageJSONKeyValue(t, "error", func(t *testing.T, value any) {
				require.Equal(t, "tag_ids should not be empty", value)
			})
		})

		// Test bookmark not found
		t.Run("BookmarkNotFound", func(t *testing.T) {
			payload := bulkUpdatePayload{
				BookmarkIDs: []int{9999}, // Non-existent bookmark ID
				TagIDs:      []int{tagID},
			}
			payloadBytes, err := json.Marshal(payload)
			require.NoError(t, err)

			rec := testutil.PerformRequest(
				deps,
				api_v1.HandleBulkUpdateBookmarkTags,
				http.MethodPut,
				"/api/v1/bookmarks/bulk/tags",
				testutil.WithFakeAdmin(),
				testutil.WithBody(string(payloadBytes)),
			)

			require.Equal(t, http.StatusInternalServerError, rec.Code)

			testResp := testutil.NewTestResponseFromRecorder(rec)
			testResp.AssertNotOk(t)
			testResp.AssertMessageJSONKeyValue(t, "error", func(t *testing.T, value any) {
				require.Equal(t, "Failed to update bookmarks", value)
			})
		})
	})

	// Test GetBookmarkTags
	t.Run("GetBookmarkTags", func(t *testing.T) {
		// Add a tag to the bookmark first
		err := db.AddTagToBookmark(ctx, bookmarkID, tagID)
		require.NoError(t, err)

		// Create a request to get the tags
		rec := testutil.PerformRequest(
			deps,
			api_v1.HandleGetBookmarkTags,
			http.MethodGet,
			"/api/v1/bookmarks/"+strconv.Itoa(bookmarkID)+"/tags",
			testutil.WithFakeAdmin(),
			testutil.WithRequestPathValue("id", strconv.Itoa(bookmarkID)),
		)

		// Check the response
		require.Equal(t, http.StatusOK, rec.Code)

		// Parse the response
		testResp := testutil.NewTestResponseFromRecorder(rec)
		testResp.AssertOk(t)

		testResp.AssertMessageIsNotEmptyList(t)

		testResp.ForEach(t, func(item map[string]any) {
			require.NotZero(t, item["id"])
			require.NotEmpty(t, item["name"])
		})
	})

	// Test AddTagToBookmark
	t.Run("AddTagToBookmark", func(t *testing.T) {
		// Remove the tag first to ensure a clean state
		err := db.RemoveTagFromBookmark(ctx, bookmarkID, tagID)
		require.NoError(t, err)

		// Create a request to add the tag
		payload := bookmarkTagPayload{
			TagID: tagID,
		}
		payloadBytes, err := json.Marshal(payload)
		require.NoError(t, err)

		rec := testutil.PerformRequest(
			deps,
			api_v1.HandleAddTagToBookmark,
			http.MethodPost,
			"/api/v1/bookmarks/"+strconv.Itoa(bookmarkID)+"/tags",
			testutil.WithFakeAdmin(),
			testutil.WithRequestPathValue("id", strconv.Itoa(bookmarkID)),
			testutil.WithBody(string(payloadBytes)),
		)

		// Check the response
		require.Equal(t, http.StatusCreated, rec.Code)

		// Verify the tag was added
		tags, err := deps.Domains().Tags().ListTags(ctx, model.ListTagsOptions{
			BookmarkID: bookmarkID,
		})
		require.NoError(t, err)
		require.Len(t, tags, 1)
		assert.Equal(t, tagID, tags[0].ID)
	})

	// Test RemoveTagFromBookmark
	t.Run("RemoveTagFromBookmark", func(t *testing.T) {
		// Add the tag first to ensure it exists
		err := db.AddTagToBookmark(ctx, bookmarkID, tagID)
		require.NoError(t, err)

		// Create a request to remove the tag
		payload := bookmarkTagPayload{
			TagID: tagID,
		}
		payloadBytes, err := json.Marshal(payload)
		require.NoError(t, err)

		rec := testutil.PerformRequest(
			deps,
			api_v1.HandleRemoveTagFromBookmark,
			http.MethodDelete,
			"/api/v1/bookmarks/"+strconv.Itoa(bookmarkID)+"/tags",
			testutil.WithFakeAdmin(),
			testutil.WithRequestPathValue("id", strconv.Itoa(bookmarkID)),
			testutil.WithBody(string(payloadBytes)),
		)

		// Check the response
		require.Equal(t, http.StatusOK, rec.Code)

		// Verify the tag was removed
		tags, err := deps.Domains().Tags().ListTags(ctx, model.ListTagsOptions{
			BookmarkID: bookmarkID,
		})
		require.NoError(t, err)
		require.Len(t, tags, 0)
	})

	// Test error cases
	t.Run("ErrorCases", func(t *testing.T) {
		// Test non-existent bookmark
		t.Run("NonExistentBookmark", func(t *testing.T) {
			// Create a request to get tags for a non-existent bookmark
			rec := testutil.PerformRequest(
				deps,
				api_v1.HandleGetBookmarkTags,
				http.MethodGet,
				"/api/v1/bookmarks/9999/tags",
				testutil.WithFakeAdmin(),
				testutil.WithRequestPathValue("id", "9999"),
			)

			// Check the response
			require.Equal(t, http.StatusNotFound, rec.Code)

			// Parse the response
			testResp := testutil.NewTestResponseFromRecorder(rec)
			testResp.AssertNotOk(t)
			testResp.AssertMessageJSONKeyValue(t, "error", func(t *testing.T, value any) {
				require.Equal(t, "Bookmark not found", value)
			})
		})

		// Test non-existent tag
		t.Run("NonExistentTag", func(t *testing.T) {
			// Create a request to add a non-existent tag
			payload := bookmarkTagPayload{
				TagID: 9999,
			}
			payloadBytes, err := json.Marshal(payload)
			require.NoError(t, err)

			rec := testutil.PerformRequest(
				deps,
				api_v1.HandleAddTagToBookmark,
				http.MethodPost,
				"/api/v1/bookmarks/"+strconv.Itoa(bookmarkID)+"/tags",
				testutil.WithFakeAdmin(),
				testutil.WithRequestPathValue("id", strconv.Itoa(bookmarkID)),
				testutil.WithBody(string(payloadBytes)),
			)

			// Check the response
			require.Equal(t, http.StatusNotFound, rec.Code)

			// Parse the response
			testResp := testutil.NewTestResponseFromRecorder(rec)
			testResp.AssertNotOk(t)
			testResp.AssertMessageJSONKeyValue(t, "error", func(t *testing.T, value any) {
				require.Equal(t, "Tag not found", value)
			})
		})

		// Test non-existent bookmark for AddTagToBookmark
		t.Run("NonExistentBookmarkForAddTag", func(t *testing.T) {
			// Create a request to add a tag to a non-existent bookmark
			payload := bookmarkTagPayload{
				TagID: tagID,
			}
			payloadBytes, err := json.Marshal(payload)
			require.NoError(t, err)

			rec := testutil.PerformRequest(
				deps,
				api_v1.HandleAddTagToBookmark,
				http.MethodPost,
				"/api/v1/bookmarks/9999/tags",
				testutil.WithFakeAdmin(),
				testutil.WithRequestPathValue("id", "9999"),
				testutil.WithBody(string(payloadBytes)),
			)

			// Check the response
			require.Equal(t, http.StatusNotFound, rec.Code)

			// Parse the response
			testResp := testutil.NewTestResponseFromRecorder(rec)
			testResp.AssertNotOk(t)
			testResp.AssertMessageJSONKeyValue(t, "error", func(t *testing.T, value any) {
				require.Equal(t, "Bookmark not found", value)
			})
		})

		// Test non-existent bookmark for RemoveTagFromBookmark
		t.Run("NonExistentBookmarkForRemoveTag", func(t *testing.T) {
			// Create a request to remove a tag from a non-existent bookmark
			payload := bookmarkTagPayload{
				TagID: tagID,
			}
			payloadBytes, err := json.Marshal(payload)
			require.NoError(t, err)

			rec := testutil.PerformRequest(
				deps,
				api_v1.HandleRemoveTagFromBookmark,
				http.MethodDelete,
				"/api/v1/bookmarks/9999/tags",
				testutil.WithFakeAdmin(),
				testutil.WithRequestPathValue("id", "9999"),
				testutil.WithBody(string(payloadBytes)),
			)

			// Check the response
			require.Equal(t, http.StatusNotFound, rec.Code)

			// Parse the response
			testResp := testutil.NewTestResponseFromRecorder(rec)
			testResp.AssertNotOk(t)
			testResp.AssertMessageJSONKeyValue(t, "error", func(t *testing.T, value any) {
				require.Equal(t, "Bookmark not found", value)
			})
		})

		// Test non-existent tag for RemoveTagFromBookmark
		t.Run("NonExistentTagForRemoveTag", func(t *testing.T) {
			// Create a request to remove a non-existent tag
			payload := bookmarkTagPayload{
				TagID: 9999,
			}
			payloadBytes, err := json.Marshal(payload)
			require.NoError(t, err)

			rec := testutil.PerformRequest(
				deps,
				api_v1.HandleRemoveTagFromBookmark,
				http.MethodDelete,
				"/api/v1/bookmarks/"+strconv.Itoa(bookmarkID)+"/tags",
				testutil.WithFakeAdmin(),
				testutil.WithRequestPathValue("id", strconv.Itoa(bookmarkID)),
				testutil.WithBody(string(payloadBytes)),
			)

			// Check the response
			require.Equal(t, http.StatusNotFound, rec.Code)

			// Parse the response
			testResp := testutil.NewTestResponseFromRecorder(rec)
			testResp.AssertNotOk(t)
			testResp.AssertMessageJSONKeyValue(t, "error", func(t *testing.T, value any) {
				require.Equal(t, "Tag not found", value)
			})
		})

		// Test invalid bookmark ID
		t.Run("InvalidBookmarkID", func(t *testing.T) {
			// Create a request with an invalid bookmark ID
			rec := testutil.PerformRequest(
				deps,
				api_v1.HandleGetBookmarkTags,
				http.MethodGet,
				"/api/v1/bookmarks/invalid/tags",
				testutil.WithFakeAdmin(),
				testutil.WithRequestPathValue("id", "invalid"),
			)

			// Check the response
			require.Equal(t, http.StatusBadRequest, rec.Code)

			// Parse the response
			testResp := testutil.NewTestResponseFromRecorder(rec)
			testResp.AssertNotOk(t)
			testResp.AssertMessageJSONKeyValue(t, "error", func(t *testing.T, value any) {
				require.Equal(t, "Invalid bookmark ID", value)
			})
		})

		// Test invalid bookmark ID for AddTagToBookmark
		t.Run("InvalidBookmarkIDForAddTag", func(t *testing.T) {
			// Create a request with an invalid bookmark ID
			payload := bookmarkTagPayload{
				TagID: tagID,
			}
			payloadBytes, err := json.Marshal(payload)
			require.NoError(t, err)

			rec := testutil.PerformRequest(
				deps,
				api_v1.HandleAddTagToBookmark,
				http.MethodPost,
				"/api/v1/bookmarks/invalid/tags",
				testutil.WithFakeAdmin(),
				testutil.WithRequestPathValue("id", "invalid"),
				testutil.WithBody(string(payloadBytes)),
			)

			// Check the response
			require.Equal(t, http.StatusBadRequest, rec.Code)

			// Parse the response
			testResp := testutil.NewTestResponseFromRecorder(rec)
			testResp.AssertNotOk(t)
			testResp.AssertMessageJSONKeyValue(t, "error", func(t *testing.T, value any) {
				require.Equal(t, "Invalid bookmark ID", value)
			})
		})

		// Test invalid bookmark ID for RemoveTagFromBookmark
		t.Run("InvalidBookmarkIDForRemoveTag", func(t *testing.T) {
			// Create a request with an invalid bookmark ID
			payload := bookmarkTagPayload{
				TagID: tagID,
			}
			payloadBytes, err := json.Marshal(payload)
			require.NoError(t, err)

			rec := testutil.PerformRequest(
				deps,
				api_v1.HandleRemoveTagFromBookmark,
				http.MethodDelete,
				"/api/v1/bookmarks/invalid/tags",
				testutil.WithFakeAdmin(),
				testutil.WithRequestPathValue("id", "invalid"),
				testutil.WithBody(string(payloadBytes)),
			)

			// Check the response
			require.Equal(t, http.StatusBadRequest, rec.Code)

			// Parse the response
			testResp := testutil.NewTestResponseFromRecorder(rec)
			testResp.AssertNotOk(t)
			testResp.AssertMessageJSONKeyValue(t, "error", func(t *testing.T, value any) {
				require.Equal(t, "Invalid bookmark ID", value)
			})
		})

		// Test invalid payload
		t.Run("InvalidPayload", func(t *testing.T) {
			// Create a request with an invalid payload
			invalidPayload := []byte(`{"tag_id": "invalid"}`)

			rec := testutil.PerformRequest(
				deps,
				api_v1.HandleAddTagToBookmark,
				http.MethodPost,
				"/api/v1/bookmarks/"+strconv.Itoa(bookmarkID)+"/tags",
				testutil.WithFakeAdmin(),
				testutil.WithRequestPathValue("id", strconv.Itoa(bookmarkID)),
				testutil.WithBody(string(invalidPayload)),
			)

			// Check the response
			require.Equal(t, http.StatusBadRequest, rec.Code)

			// Parse the response
			testResp := testutil.NewTestResponseFromRecorder(rec)
			testResp.AssertNotOk(t)
			testResp.AssertMessageJSONKeyValue(t, "error", func(t *testing.T, value any) {
				require.Equal(t, "Invalid request payload", value)
			})
		})

		// Test zero tag ID
		t.Run("ZeroTagID", func(t *testing.T) {
			// Create a request with a zero tag ID
			payload := bookmarkTagPayload{
				TagID: 0,
			}
			payloadBytes, err := json.Marshal(payload)
			require.NoError(t, err)

			rec := testutil.PerformRequest(
				deps,
				api_v1.HandleAddTagToBookmark,
				http.MethodPost,
				"/api/v1/bookmarks/"+strconv.Itoa(bookmarkID)+"/tags",
				testutil.WithFakeAdmin(),
				testutil.WithRequestPathValue("id", strconv.Itoa(bookmarkID)),
				testutil.WithBody(string(payloadBytes)),
			)

			// Check the response
			require.Equal(t, http.StatusBadRequest, rec.Code)

			// Parse the response
			testResp := testutil.NewTestResponseFromRecorder(rec)
			testResp.AssertNotOk(t)
			testResp.AssertMessageJSONKeyValue(t, "error", func(t *testing.T, value any) {
				require.Equal(t, "tag_id should be a positive integer", value)
			})
		})

		// Test negative tag ID
		t.Run("NegativeTagID", func(t *testing.T) {
			// Create a request with a negative tag ID
			payload := bookmarkTagPayload{
				TagID: -1,
			}
			payloadBytes, err := json.Marshal(payload)
			require.NoError(t, err)

			rec := testutil.PerformRequest(
				deps,
				api_v1.HandleAddTagToBookmark,
				http.MethodPost,
				"/api/v1/bookmarks/"+strconv.Itoa(bookmarkID)+"/tags",
				testutil.WithFakeAdmin(),
				testutil.WithRequestPathValue("id", strconv.Itoa(bookmarkID)),
				testutil.WithBody(string(payloadBytes)),
			)

			// Check the response
			require.Equal(t, http.StatusBadRequest, rec.Code)

			// Parse the response
			testResp := testutil.NewTestResponseFromRecorder(rec)
			testResp.AssertNotOk(t)
			testResp.AssertMessageJSONKeyValue(t, "error", func(t *testing.T, value any) {
				require.Equal(t, "tag_id should be a positive integer", value)
			})
		})

		// Test validation for RemoveTagFromBookmark
		t.Run("RemoveTagValidation", func(t *testing.T) {
			// Create a request with a zero tag ID
			payload := bookmarkTagPayload{
				TagID: 0,
			}
			payloadBytes, err := json.Marshal(payload)
			require.NoError(t, err)

			rec := testutil.PerformRequest(
				deps,
				api_v1.HandleRemoveTagFromBookmark,
				http.MethodDelete,
				"/api/v1/bookmarks/"+strconv.Itoa(bookmarkID)+"/tags",
				testutil.WithFakeAdmin(),
				testutil.WithRequestPathValue("id", strconv.Itoa(bookmarkID)),
				testutil.WithBody(string(payloadBytes)),
			)

			// Check the response
			require.Equal(t, http.StatusBadRequest, rec.Code)

			// Parse the response
			testResp := testutil.NewTestResponseFromRecorder(rec)
			testResp.AssertNotOk(t)
			testResp.AssertMessageJSONKeyValue(t, "error", func(t *testing.T, value any) {
				require.Equal(t, "tag_id should be a positive integer", value)
			})
		})
	})
}
