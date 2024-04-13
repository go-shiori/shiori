package testutil

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

type testResponse struct {
	Response response.Response
}

func (r *testResponse) AssertMessageIsEmptyList(t *testing.T) {
	require.Equal(t, []interface{}{}, r.Response.GetMessage())
}

func (r *testResponse) AssertNilMessage(t *testing.T) {
	require.Equal(t, nil, r.Response.GetMessage())
}

func (r testResponse) AssertMessageEquals(t *testing.T, expected interface{}) {
	require.Equal(t, expected, r.Response.GetMessage())
}

func (r *testResponse) AssertOk(t *testing.T) {
	require.True(t, !r.Response.IsError())
}

func (r *testResponse) AssertNotOk(t *testing.T) {
	require.False(t, r.Response.IsError())
}

func NewTestResponseFromBytes(b []byte) (*testResponse, error) {
	r := testResponse{}
	if err := json.Unmarshal(b, &r.Response); err != nil {
		return nil, errors.Wrap(err, "error parsing response")
	}
	return &r, nil
}

func NewTestResponseFromReader(r io.Reader) (*testResponse, error) {
	response := testResponse{}
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&response.Response); err != nil {
		return nil, errors.Wrap(err, "error parsing response")
	}
	return &response, nil
}
