package context

import (
	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/model"
)

type Context struct {
	*gin.Context

	Account *model.Account
}

func NewContextFromGin(c *gin.Context) *Context {
	return &Context{
		Context: c,
	}
}
