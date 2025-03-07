package health

import (
	"github.com/hellofresh/health-go/v5"

	"github.com/kubecano/cano-collector/config"
	"github.com/kubecano/cano-collector/pkg/logger"
)

func RegisterHealthChecks() *health.Health {
	h, err := health.New(health.WithComponent(
		health.Component{
			Name:    config.GlobalConfig.AppName,
			Version: config.GlobalConfig.AppVersion,
		}),
	)
	if err != nil {
		logger.Errorf("Failed to create health check: %v", err)
	}

	return h
}
