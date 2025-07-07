package interfaces

import (
	"github.com/gin-gonic/gin"
)

//go:generate mockgen -destination=../../../mocks/alert_handler_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/alert/interfaces AlertHandlerInterface
type AlertHandlerInterface interface {
	HandleAlert(c *gin.Context)
}
