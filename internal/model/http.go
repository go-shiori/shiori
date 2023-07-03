package model

import "github.com/gin-gonic/gin"

type Routes interface {
	Setup(group *gin.RouterGroup) Routes
}
