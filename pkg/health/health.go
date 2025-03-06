package health

import (
	"github.com/hellofresh/health-go/v5"

	"github.com/kubecano/cano-collector/config"
)

func RegisterHealthChecks() *health.Health {
	h, _ := health.New(health.WithComponent(
		health.Component{
			Name:    config.GlobalConfig.AppName,
			Version: config.GlobalConfig.AppVersion,
		}),
	)

	return h
}
