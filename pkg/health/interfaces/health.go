package interfaces

import (
	"net/http"
)

//go:generate mockgen -destination=../../../mocks/health_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/health/interfaces HealthInterface
type HealthInterface interface {
	RegisterHealthChecks() error
	Handler() http.Handler
}
