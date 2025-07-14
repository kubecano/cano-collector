package interfaces

import (
	"github.com/gin-gonic/gin"
)

//go:generate mockgen -source=router.go -destination=../../../mocks/router_mock.go -package=mocks
type RouterInterface interface {
	SetupRouter() *gin.Engine
	StartServer(router *gin.Engine)
}
