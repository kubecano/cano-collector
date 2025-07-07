package interfaces

import (
	"github.com/gin-gonic/gin"
)

//go:generate mockgen -destination=../../../mocks/router_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/router/interfaces RouterInterface
type RouterInterface interface {
	SetupRouter() *gin.Engine
	StartServer(router *gin.Engine)
}
