package interfaces

import (
	"github.com/gin-gonic/gin"
)

//go:generate mockgen -source=handler.go -destination=../../../mocks/alert_handler_mock.go -package=mocks
type AlertHandlerInterface interface {
	HandleAlert(c *gin.Context)
}
