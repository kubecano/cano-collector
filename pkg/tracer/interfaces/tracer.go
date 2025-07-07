package interfaces

import (
	"context"

	"github.com/gin-gonic/gin"
)

//go:generate mockgen -destination=../../../mocks/tracer_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/tracer/interfaces TracerInterface
type TracerInterface interface {
	InitTracer(ctx context.Context) error
	TraceLoggerMiddleware() gin.HandlerFunc
	ShutdownTracer(ctx context.Context) error
}
