package model

import "github.com/gin-gonic/gin"

const ContextAccountKey = "account"

type Routes interface {
	Setup(group *gin.RouterGroup) Routes
}
