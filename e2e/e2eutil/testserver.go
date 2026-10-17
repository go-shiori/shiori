package e2eutil

import (
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"runtime"
	"testing"
)

// StartTestServer starts an HTTP file server serving the e2e/testserver/ directory.
// It returns the base URL reachable from Docker containers (using host.docker.internal).
// The server is automatically stopped when the test completes.
func StartTestServer(t *testing.T) string {
	t.Helper()

	// Resolve the testserver content directory relative to this source file
	_, thisFile, _, _ := runtime.Caller(0)
	contentDir := filepath.Join(filepath.Dir(thisFile), "..", "testserver")

	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		t.Fatalf("failed to start test server listener: %v", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	server := &http.Server{
		Handler: http.FileServer(http.Dir(contentDir)),
	}

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			t.Logf("test server error: %v", err)
		}
	}()

	t.Cleanup(func() {
		server.Close()
	})

	return fmt.Sprintf("http://host.docker.internal:%d", port)
}
