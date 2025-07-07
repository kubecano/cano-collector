package health

import (
	"net/http"

	"github.com/hellofresh/health-go/v5"

	"github.com/kubecano/cano-collector/config"
	logger_interfaces "github.com/kubecano/cano-collector/pkg/logger/interfaces"
)

type HealthChecker struct {
	cfg    config.Config
	logger logger_interfaces.LoggerInterface
	health *health.Health
}

func NewHealthChecker(cfg config.Config, logger logger_interfaces.LoggerInterface) *HealthChecker {
	return &HealthChecker{cfg: cfg, logger: logger}
}

func (hc *HealthChecker) RegisterHealthChecks() error {
	hc.logger.Debug("Starting health check registration")

	var err error
	hc.health, err = health.New(health.WithComponent(
		health.Component{
			Name:    hc.cfg.AppName,
			Version: hc.cfg.AppVersion,
		}),
	)
	if err != nil {
		hc.logger.Errorf("Failed to create health check: %v", err)
		return err
	}

	hc.logger.Info("Health check registration completed successfully")
	return nil
}

func (hc *HealthChecker) Handler() http.Handler {
	return hc.health.Handler()
}
