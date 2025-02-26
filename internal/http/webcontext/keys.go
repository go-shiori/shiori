package webcontext

type contextKey string

const (
	accountKey   contextKey = "account"
	requestIDKey contextKey = "requestID"
)
