package actions

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubecano/cano-collector/mocks"
	actions_interfaces "github.com/kubecano/cano-collector/pkg/workflow/actions/interfaces"
)

func TestNewBaseAction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockMetrics := mocks.NewMockMetricsInterface(ctrl)

	config := actions_interfaces.ActionConfig{
		Name:    "test-action",
		Type:    "test_type",
		Enabled: true,
		Timeout: 60,
		Parameters: map[string]interface{}{
			"param1": "value1",
			"param2": 42,
		},
	}

	baseAction := NewBaseAction(config, mockLogger, mockMetrics)

	require.NotNil(t, baseAction)
	assert.Equal(t, config, baseAction.config)
	assert.Equal(t, mockLogger, baseAction.logger)
	assert.Equal(t, mockMetrics, baseAction.metrics)
}

func TestBaseAction_GetName(t *testing.T) {
	config := actions_interfaces.ActionConfig{
		Name: "test-action-name",
	}

	baseAction := NewBaseAction(config, nil, nil)
	assert.Equal(t, "test-action-name", baseAction.GetName())
}

func TestBaseAction_GetType(t *testing.T) {
	config := actions_interfaces.ActionConfig{
		Type: "test-action-type",
	}

	baseAction := NewBaseAction(config, nil, nil)
	assert.Equal(t, "test-action-type", baseAction.GetType())
}

func TestBaseAction_GetTimeout(t *testing.T) {
	tests := []struct {
		name           string
		timeoutConfig  int
		expectedResult time.Duration
	}{
		{
			name:           "custom timeout",
			timeoutConfig:  60,
			expectedResult: 60 * time.Second,
		},
		{
			name:           "zero timeout uses default",
			timeoutConfig:  0,
			expectedResult: 30 * time.Second,
		},
		{
			name:           "negative timeout uses default",
			timeoutConfig:  -10,
			expectedResult: 30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := actions_interfaces.ActionConfig{
				Timeout: tt.timeoutConfig,
			}

			baseAction := NewBaseAction(config, nil, nil)
			assert.Equal(t, tt.expectedResult, baseAction.GetTimeout())
		})
	}
}

func TestBaseAction_IsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		enabled  bool
		expected bool
	}{
		{
			name:     "enabled action",
			enabled:  true,
			expected: true,
		},
		{
			name:     "disabled action",
			enabled:  false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := actions_interfaces.ActionConfig{
				Enabled: tt.enabled,
			}

			baseAction := NewBaseAction(config, nil, nil)
			assert.Equal(t, tt.expected, baseAction.IsEnabled())
		})
	}
}

func TestBaseAction_GetParameter(t *testing.T) {
	config := actions_interfaces.ActionConfig{
		Parameters: map[string]interface{}{
			"existing_key": "existing_value",
			"number_key":   42,
		},
	}

	baseAction := NewBaseAction(config, nil, nil)

	// Test existing parameter
	value, exists := baseAction.GetParameter("existing_key")
	assert.True(t, exists)
	assert.Equal(t, "existing_value", value)

	// Test non-existing parameter
	value, exists = baseAction.GetParameter("non_existing_key")
	assert.False(t, exists)
	assert.Nil(t, value)
}

func TestBaseAction_GetStringParameter(t *testing.T) {
	config := actions_interfaces.ActionConfig{
		Parameters: map[string]interface{}{
			"string_key": "string_value",
			"number_key": 42,
		},
	}

	baseAction := NewBaseAction(config, nil, nil)

	// Test existing string parameter
	result := baseAction.GetStringParameter("string_key", "default")
	assert.Equal(t, "string_value", result)

	// Test non-existing parameter
	result = baseAction.GetStringParameter("non_existing_key", "default")
	assert.Equal(t, "default", result)

	// Test parameter with wrong type
	result = baseAction.GetStringParameter("number_key", "default")
	assert.Equal(t, "default", result)
}

func TestBaseAction_GetIntParameter(t *testing.T) {
	config := actions_interfaces.ActionConfig{
		Parameters: map[string]interface{}{
			"int_key":    42,
			"float_key":  3.14,
			"string_key": "not_a_number",
		},
	}

	baseAction := NewBaseAction(config, nil, nil)

	// Test existing int parameter
	result := baseAction.GetIntParameter("int_key", 999)
	assert.Equal(t, 42, result)

	// Test float parameter (should be converted)
	result = baseAction.GetIntParameter("float_key", 999)
	assert.Equal(t, 3, result)

	// Test non-existing parameter
	result = baseAction.GetIntParameter("non_existing_key", 999)
	assert.Equal(t, 999, result)

	// Test parameter with wrong type
	result = baseAction.GetIntParameter("string_key", 999)
	assert.Equal(t, 999, result)
}

func TestBaseAction_GetBoolParameter(t *testing.T) {
	config := actions_interfaces.ActionConfig{
		Parameters: map[string]interface{}{
			"bool_true":  true,
			"bool_false": false,
			"string_key": "not_a_bool",
		},
	}

	baseAction := NewBaseAction(config, nil, nil)

	// Test existing bool parameter (true)
	result := baseAction.GetBoolParameter("bool_true", false)
	assert.True(t, result)

	// Test existing bool parameter (false)
	result = baseAction.GetBoolParameter("bool_false", true)
	assert.False(t, result)

	// Test non-existing parameter
	result = baseAction.GetBoolParameter("non_existing_key", true)
	assert.True(t, result)

	// Test parameter with wrong type
	result = baseAction.GetBoolParameter("string_key", true)
	assert.True(t, result)
}

func TestBaseAction_ValidateBasicConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  actions_interfaces.ActionConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: actions_interfaces.ActionConfig{
				Name:    "test-action",
				Type:    "test_type",
				Timeout: 30,
			},
			wantErr: false,
		},
		{
			name: "empty name",
			config: actions_interfaces.ActionConfig{
				Name: "",
				Type: "test_type",
			},
			wantErr: true,
			errMsg:  "action name cannot be empty",
		},
		{
			name: "empty type",
			config: actions_interfaces.ActionConfig{
				Name: "test-action",
				Type: "",
			},
			wantErr: true,
			errMsg:  "action type cannot be empty",
		},
		{
			name: "negative timeout",
			config: actions_interfaces.ActionConfig{
				Name:    "test-action",
				Type:    "test_type",
				Timeout: -1,
			},
			wantErr: true,
			errMsg:  "action timeout cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseAction := NewBaseAction(tt.config, nil, nil)
			err := baseAction.ValidateBasicConfig()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestBaseAction_CreateSuccessResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	// Mock the debug call that might be made
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

	config := actions_interfaces.ActionConfig{Name: "test"}
	baseAction := NewBaseAction(config, mockLogger, nil)

	testData := map[string]string{"key": "value"}
	result := baseAction.CreateSuccessResult(testData)

	require.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, testData, result.Data)
	require.NoError(t, result.Error)
	assert.NotNil(t, result.Metadata)
}

func TestBaseAction_CreateErrorResult(t *testing.T) {
	config := actions_interfaces.ActionConfig{Name: "test"}
	baseAction := NewBaseAction(config, nil, nil)

	testData := map[string]string{"key": "value"}
	testError := assert.AnError
	result := baseAction.CreateErrorResult(testError, testData)

	require.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Equal(t, testData, result.Data)
	assert.Equal(t, testError, result.Error)
	assert.NotNil(t, result.Metadata)
}
