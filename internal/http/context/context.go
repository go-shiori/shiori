package context

import (
	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/model"
)

// Context is a wrapper of gin.Context that contains authentication information
type Context struct {
	*gin.Context

	account *model.Account
}

// NewContextFromGin returns a new Context instance from gin.Context
func NewContextFromGin(c *gin.Context) *Context {
	return &Context{
		Context: c,
	}
}

// New returns a new Context instance
func New() *Context {
	return NewContextFromGin(&gin.Context{})
}
