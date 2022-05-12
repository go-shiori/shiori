package api

// ShioriServer holds configuration and connection pools
type ShioriServer struct{}

// NewShioriServer creates new API server
func NewShioriServer() *ShioriServer {
	return &ShioriServer{}
}
