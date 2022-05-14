package api

import "github.com/go-shiori/shiori/internal/database"

// ShioriServer holds configuration and connection pools
type ShioriServer struct {
	DB database.DB
}

// NewShioriServer creates new API server
func NewShioriServer(db database.DB) *ShioriServer {
	return &ShioriServer{
		DB: db,
	}
}
