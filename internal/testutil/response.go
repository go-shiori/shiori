package testutil

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/stretchr/testify/require"
)

type testResponse struct {
	Response response.Response
}

func (r *testResponse) AssertMessageIsEmptyList(t *testing.T) {
	var jsonData []any
	err := json.Unmarshal(r.Response.GetData().([]byte), &jsonData)
	require.NoError(t, err)
	require.Equal(t, []any{}, jsonData)
}

func (r *testResponse) AssertMessageIsNotEmptyList(t *testing.T) {
	var jsonData []any
	err := json.Unmarshal(r.Response.GetData().([]byte), &jsonData)
	require.NoError(t, err)
	require.Greater(t, len(jsonData), 0)
}

func (r *testResponse) AssertMessageIsListLength(t *testing.T, length int) {
	var jsonData []any
	err := json.Unmarshal(r.Response.GetData().([]byte), &jsonData)
	require.NoError(t, err)
	require.Len(t, jsonData, length)
}

// ForEach iterates over the items in the response and calls the provided function
// with each item.
func (r *testResponse) ForEach(t *testing.T, fn func(item map[string]any)) {
	var jsonData []any
	err := json.Unmarshal(r.Response.GetData().([]byte), &jsonData)
	require.NoError(t, err)
	for _, item := range jsonData {
		fn(item.(map[string]any))
	}
}

func (r *testResponse) AssertNilMessage(t *testing.T) {
	require.Equal(t, nil, r.Response.GetData())
}

func (r testResponse) AssertMessageEquals(t *testing.T, expected any) {
	require.Equal(t, expected, r.Response.GetData())
}

func (r testResponse) AssertMessageJSONContains(t *testing.T, expected string) {
	require.JSONEq(t, expected, string(r.Response.GetData().([]byte)))
}

// AssertMessageJSONContainsKey asserts that the response message contains a key
// and returns the value of the key to be used in other comparisons depending on the
// value type.
func (r testResponse) AssertMessageJSONContainsKey(t *testing.T, key string) any {
	var jsonData map[string]any
	err := json.Unmarshal(r.Response.GetData().([]byte), &jsonData)
	require.NoError(t, err)
	require.Contains(t, jsonData, key)
	return jsonData[key]
}

// AssertMessageJSONKeyValue asserts that the response message contains a key
// and calls the provided function with the value of the key to be used in other
// comparisons depending on the value type.
func (r *testResponse) AssertMessageJSONKeyValue(t *testing.T, key string, valueAssertFunc func(t *testing.T, value any)) {
	value := r.AssertMessageJSONContainsKey(t, key)
	valueAssertFunc(t, value)
}

func (r *testResponse) AssertMessageContains(t *testing.T, expected string) {
	require.Contains(t, r.Response.GetData(), expected)
}

func (r *testResponse) AssertMessageIsBytes(t *testing.T, expected []byte) {
	require.Equal(t, expected, r.Response.GetData().([]byte))
}

func (r *testResponse) AssertOk(t *testing.T) {
	require.False(t, r.Response.IsError())
}

func (r *testResponse) AssertNotOk(t *testing.T) {
	require.True(t, r.Response.IsError())
}

func NewTestResponseFromRecorder(w *httptest.ResponseRecorder) *testResponse {
	return &testResponse{Response: *response.NewResponse(w.Body.Bytes(), w.Code)}
}
