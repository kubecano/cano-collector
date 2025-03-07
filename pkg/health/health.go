package health

import (
	"github.com/hellofresh/health-go/v5"

	"github.com/kubecano/cano-collector/config"
	"github.com/kubecano/cano-collector/pkg/logger"
)

func RegisterHealthChecks() (*health.Health, error) {
	logger.Debug("Starting health check registration")

	h, err := health.New(health.WithComponent(
		health.Component{
			Name:    config.GlobalConfig.AppName,
			Version: config.GlobalConfig.AppVersion,
		}),
	)
	if err != nil {
		logger.Errorf("Failed to create health check: %v", err)
		return nil, err
	}

	logger.Info("Health check registration completed successfully")
	return h, nil
}
