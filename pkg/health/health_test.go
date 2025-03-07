package health

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"github.com/kubecano/cano-collector/pkg/logger"

	"github.com/stretchr/testify/require"

	"github.com/hellofresh/health-go/v5"

	"github.com/kubecano/cano-collector/config"

	"github.com/stretchr/testify/assert"
)

func TestRegisterHealthChecks(t *testing.T) {
	l, _ := zap.NewDevelopment()
	logger.SetLogger(l)
	config.GlobalConfig = config.Config{
		AppName:    "cano-collector",
		AppVersion: "1.0.0",
	}

	h, err := RegisterHealthChecks()
	require.NoError(t, err, "RegisterHealthChecks should not return an error")
	assert.NotNil(t, h, "Healthcheck instance should not be nil")

	ctx := context.Background()

	healthStatus := h.Measure(ctx)

	assert.Equal(t, "cano-collector", healthStatus.Component.Name, "AppName should be set correctly")
	assert.Equal(t, "1.0.0", healthStatus.Component.Version, "AppVersion should be set correctly")

	assert.Equal(t, health.StatusOK, healthStatus.Status, "Healthcheck should initially return StatusOK")
}
