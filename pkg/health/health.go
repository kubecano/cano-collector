package health

import (
	"github.com/hellofresh/health-go/v5"

	"github.com/kubecano/cano-collector/config"
	"github.com/kubecano/cano-collector/pkg/logger"
)

type HealthInterface interface {
	RegisterHealthChecks() (*health.Health, error)
}

type HealthChecker struct {
	cfg    config.Config
	logger logger.LoggerInterface
}

func NewHealthChecker(cfg config.Config, logger logger.LoggerInterface) *HealthChecker {
	return &HealthChecker{cfg: cfg, logger: logger}
}

func (hc *HealthChecker) RegisterHealthChecks() (*health.Health, error) {
	hc.logger.Debug("Starting health check registration")

	h, err := health.New(health.WithComponent(
		health.Component{
			Name:    hc.cfg.AppName,
			Version: hc.cfg.AppVersion,
		}),
	)
	if err != nil {
		hc.logger.Errorf("Failed to create health check: %v", err)
		return nil, err
	}

	hc.logger.Info("Health check registration completed successfully")
	return h, nil
}
