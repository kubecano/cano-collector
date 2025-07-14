package interfaces

import (
	"context"

	"github.com/gin-gonic/gin"
)

//go:generate mockgen -source=tracer.go -destination=../../../mocks/tracer_mock.go -package=mocks
type TracerInterface interface {
	InitTracer(ctx context.Context) error
	TraceLoggerMiddleware() gin.HandlerFunc
	ShutdownTracer(ctx context.Context) error
}
