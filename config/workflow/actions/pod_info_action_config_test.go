package actions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	actions_interfaces "github.com/kubecano/cano-collector/pkg/workflow/actions/interfaces"
)

func TestNewPodInfoActionConfigWithDefaults(t *testing.T) {
	baseConfig := actions_interfaces.ActionConfig{
		Name: "test-action",
		Type: "pod_info",
	}

	config := NewPodInfoActionConfigWithDefaults(baseConfig)

	assert.Equal(t, "test-action", config.Name)
	assert.Equal(t, "pod_info", config.Type)
	assert.True(t, config.IncludePreviousState)
	assert.False(t, config.IncludeInitContainers)
	assert.Equal(t, int32(0), config.MinRestartCount)
}

func TestPodInfoActionConfig_GetActionType(t *testing.T) {
	config := PodInfoActionConfig{}

	assert.Equal(t, "pod_info", config.GetActionType())
}

func TestPodInfoActionConfig_Validate_Success(t *testing.T) {
	config := PodInfoActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name: "test-action",
			Type: "pod_info",
		},
		MinRestartCount: 3,
	}

	err := config.Validate()
	require.NoError(t, err)
}

func TestPodInfoActionConfig_Validate_EmptyName(t *testing.T) {
	config := PodInfoActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Type: "pod_info",
		},
	}

	err := config.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "action name cannot be empty")
}

func TestPodInfoActionConfig_Validate_NegativeMinRestartCount(t *testing.T) {
	config := PodInfoActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name: "test-action",
			Type: "pod_info",
		},
		MinRestartCount: -1,
	}

	err := config.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "min_restart_count cannot be negative")
}

func TestPodInfoActionConfig_UpdateFromParameters_AllFields(t *testing.T) {
	config := PodInfoActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name: "test-action",
			Type: "pod_info",
		},
	}

	params := map[string]interface{}{
		"include_previous_state":  true,
		"include_init_containers": true,
		"min_restart_count":       5,
		"pod_name":                "my-pod",
		"container":               "my-container",
	}

	err := config.UpdateFromParameters(params)
	require.NoError(t, err)

	assert.True(t, config.IncludePreviousState)
	assert.True(t, config.IncludeInitContainers)
	assert.Equal(t, int32(5), config.MinRestartCount)
	assert.Equal(t, "my-pod", config.PodName)
	assert.Equal(t, "my-container", config.Container)
}

func TestPodInfoActionConfig_UpdateFromParameters_Float64MinRestartCount(t *testing.T) {
	config := PodInfoActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name: "test-action",
			Type: "pod_info",
		},
	}

	params := map[string]interface{}{
		"min_restart_count": float64(7),
	}

	err := config.UpdateFromParameters(params)
	require.NoError(t, err)
	assert.Equal(t, int32(7), config.MinRestartCount)
}

func TestPodInfoActionConfig_UpdateFromParameters_NilParams(t *testing.T) {
	config := PodInfoActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name: "test-action",
			Type: "pod_info",
		},
		IncludePreviousState: true,
	}

	err := config.UpdateFromParameters(nil)
	require.NoError(t, err)

	// Should preserve existing values
	assert.True(t, config.IncludePreviousState)
}

func TestPodInfoActionConfig_UpdateFromParameters_EmptyParams(t *testing.T) {
	config := PodInfoActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name: "test-action",
			Type: "pod_info",
		},
		IncludePreviousState: true,
	}

	params := map[string]interface{}{}

	err := config.UpdateFromParameters(params)
	require.NoError(t, err)

	// Should preserve existing values
	assert.True(t, config.IncludePreviousState)
}

func TestPodInfoActionConfig_UpdateFromParameters_PartialParams(t *testing.T) {
	config := PodInfoActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name: "test-action",
			Type: "pod_info",
		},
		IncludePreviousState:  true,
		IncludeInitContainers: false,
		MinRestartCount:       0,
	}

	params := map[string]interface{}{
		"include_init_containers": true,
		"min_restart_count":       10,
	}

	err := config.UpdateFromParameters(params)
	require.NoError(t, err)

	// Updated fields
	assert.True(t, config.IncludeInitContainers)
	assert.Equal(t, int32(10), config.MinRestartCount)

	// Preserved fields
	assert.True(t, config.IncludePreviousState)
}

func TestPodInfoActionConfig_UpdateFromParameters_WrongTypes(t *testing.T) {
	config := PodInfoActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name: "test-action",
			Type: "pod_info",
		},
	}

	// Parameters with wrong types should be ignored
	params := map[string]interface{}{
		"include_previous_state":  "invalid_bool",
		"include_init_containers": 123,
		"min_restart_count":       "invalid_int",
		"pod_name":                123,
		"container":               true,
	}

	err := config.UpdateFromParameters(params)
	require.NoError(t, err)

	// All fields should remain at their zero values
	assert.False(t, config.IncludePreviousState)
	assert.False(t, config.IncludeInitContainers)
	assert.Equal(t, int32(0), config.MinRestartCount)
	assert.Empty(t, config.PodName)
	assert.Empty(t, config.Container)
}
