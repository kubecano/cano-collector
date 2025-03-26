package health

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/kubecano/cano-collector/mocks"

	"github.com/hellofresh/health-go/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubecano/cano-collector/config"
)

func TestRegisterHealthChecks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()

	cfg := config.Config{
		AppName:    "cano-collector",
		AppVersion: "1.0.0",
	}

	healthChecker := NewHealthChecker(cfg, mockLogger)
	err := healthChecker.RegisterHealthChecks()

	require.NoError(t, err, "RegisterHealthChecks should not return an error")
	assert.NotNil(t, healthChecker.health, "Healthcheck instance should not be nil")

	ctx := context.Background()
	healthStatus := healthChecker.health.Measure(ctx)

	assert.Equal(t, "cano-collector", healthStatus.Component.Name, "AppName should be set correctly")
	assert.Equal(t, "1.0.0", healthStatus.Component.Version, "AppVersion should be set correctly")

	assert.Equal(t, health.StatusOK, healthStatus.Status, "Healthcheck should initially return StatusOK")
}
