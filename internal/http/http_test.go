package http

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

type testMiddleware struct {
	onRequestCalled  bool
	onResponseCalled bool
	returnError      bool
}

func (m *testMiddleware) OnRequest(deps model.Dependencies, c model.WebContext) error {
	m.onRequestCalled = true
	if m.returnError {
		return errors.New("test error")
	}
	return nil
}

func (m *testMiddleware) OnResponse(deps model.Dependencies, c model.WebContext) error {
	m.onResponseCalled = true
	if m.returnError {
		return errors.New("test error")
	}
	return nil
}

func TestToHTTPHandler(t *testing.T) {
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.TODO(), logger)

	t.Run("executes handler without middleware", func(t *testing.T) {
		handlerCalled := false
		handler := func(deps model.Dependencies, c model.WebContext) {
			handlerCalled = true
			c.ResponseWriter().WriteHeader(http.StatusOK)
		}

		c, w := testutil.NewTestWebContext()
		httpHandler := ToHTTPHandler(deps, handler)
		httpHandler.ServeHTTP(w, c.Request())

		require.True(t, handlerCalled)
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("executes middleware chain", func(t *testing.T) {
		middleware1 := &testMiddleware{}
		middleware2 := &testMiddleware{}

		handlerCalled := false
		handler := func(deps model.Dependencies, c model.WebContext) {
			handlerCalled = true
			c.ResponseWriter().WriteHeader(http.StatusOK)
		}

		c, w := testutil.NewTestWebContext()
		httpHandler := ToHTTPHandler(deps, handler, middleware1, middleware2)
		httpHandler.ServeHTTP(w, c.Request())

		require.True(t, handlerCalled)
		require.True(t, middleware1.onRequestCalled)
		require.True(t, middleware1.onResponseCalled)
		require.True(t, middleware2.onRequestCalled)
		require.True(t, middleware2.onResponseCalled)
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("stops on middleware request error", func(t *testing.T) {
		middleware1 := &testMiddleware{returnError: true}
		middleware2 := &testMiddleware{}

		handlerCalled := false
		handler := func(deps model.Dependencies, c model.WebContext) {
			handlerCalled = true
		}

		c, w := testutil.NewTestWebContext()
		httpHandler := ToHTTPHandler(deps, handler, middleware1, middleware2)
		httpHandler.ServeHTTP(w, c.Request())

		require.False(t, handlerCalled)
		require.True(t, middleware1.onRequestCalled)
		require.False(t, middleware1.onResponseCalled)
		require.False(t, middleware2.onRequestCalled)
		require.False(t, middleware2.onResponseCalled)
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
