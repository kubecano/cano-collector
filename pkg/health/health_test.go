package health

import (
	"context"
	"testing"

	"github.com/hellofresh/health-go/v5"

	"github.com/kubecano/cano-collector/config"

	"github.com/stretchr/testify/assert"
)

func TestRegisterHealthChecks(t *testing.T) {
	config.GlobalConfig = config.Config{
		AppName:    "cano-collector",
		AppVersion: "1.0.0",
	}

	h := RegisterHealthChecks()

	assert.NotNil(t, h, "Healthcheck instance should not be nil")

	ctx := context.Background()

	healthStatus := h.Measure(ctx)

	assert.Equal(t, "cano-collector", healthStatus.Component.Name, "AppName should be set correctly")
	assert.Equal(t, "1.0.0", healthStatus.Component.Version, "AppVersion should be set correctly")

	assert.Equal(t, health.StatusOK, healthStatus.Status, "Healthcheck should initially return StatusOK")
}
