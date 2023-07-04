package testutil

import (
	"net/http"
	"net/http/httptest"
)

type Header struct {
	Name  string
	Value string
}

func PerformRequest(r http.Handler, method, path string, headers ...Header) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	return PerformRequestWithRecorder(w, r, method, path, headers...)
}

func PerformRequestWithRecorder(rec *httptest.ResponseRecorder, r http.Handler, method, path string, headers ...Header) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	for _, h := range headers {
		req.Header.Add(h.Name, h.Value)
	}
	r.ServeHTTP(rec, req)
	return rec
}
