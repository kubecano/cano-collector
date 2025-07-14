package interfaces

import (
	"net/http"
)

//go:generate mockgen -source=health.go -destination=../../../mocks/health_mock.go -package=mocks
type HealthInterface interface {
	RegisterHealthChecks() error
	Handler() http.Handler
}
