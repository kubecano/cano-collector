package sender

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubecano/cano-collector/mocks"
	issuepkg "github.com/kubecano/cano-collector/pkg/core/issue"
	logger_interfaces "github.com/kubecano/cano-collector/pkg/logger/interfaces"
	"github.com/kubecano/cano-collector/pkg/util"
)

// Mock implementations for testing
type mockDestinationSender struct {
	client util.HTTPClient
	logger logger_interfaces.LoggerInterface
}

func (m *mockDestinationSender) Send(ctx context.Context, issue *issuepkg.Issue) error {
	return nil
}

func (m *mockDestinationSender) SetClient(client util.HTTPClient) {
	m.client = client
}

func (m *mockDestinationSender) SetLogger(logger logger_interfaces.LoggerInterface) {
	m.logger = logger
}

func TestWithHTTPClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockHTTPClient(ctrl)
	sender := &mockDestinationSender{}

	// Test that WithHTTPClient option works
	option := WithHTTPClient(mockClient)
	require.NotNil(t, option)

	// Apply the option
	option(sender)

	// Verify that the client was set
	assert.Equal(t, mockClient, sender.client)
}

func TestWithHTTPClient_NoSetClientMethod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Test with sender that doesn't implement SetClient
	mockSender := mocks.NewMockDestinationSenderInterface(ctrl)

	mockClient := mocks.NewMockHTTPClient(ctrl)
	option := WithHTTPClient(mockClient)

	// Should not panic when sender doesn't have SetClient method
	assert.NotPanics(t, func() {
		option(mockSender)
	})
}

func TestWithLogger(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLog := mocks.NewMockLoggerInterface(ctrl)
	sender := &mockDestinationSender{}

	// Test that WithLogger option works
	option := WithLogger(mockLog)
	require.NotNil(t, option)

	// Apply the option
	option(sender)

	// Verify that the logger was set
	assert.Equal(t, mockLog, sender.logger)
}

func TestWithLogger_NoSetLoggerMethod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Test with sender that doesn't implement SetLogger
	mockSender := mocks.NewMockDestinationSenderInterface(ctrl)

	mockLog := mocks.NewMockLoggerInterface(ctrl)
	option := WithLogger(mockLog)

	// Should not panic when sender doesn't have SetLogger method
	assert.NotPanics(t, func() {
		option(mockSender)
	})
}

func TestApplyOptions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sender := &mockDestinationSender{}
	mockClient := mocks.NewMockHTTPClient(ctrl)
	mockLog := mocks.NewMockLoggerInterface(ctrl)

	// Test applying multiple options
	result := ApplyOptions(sender, WithHTTPClient(mockClient), WithLogger(mockLog))

	// Should return the same sender instance
	assert.Same(t, sender, result)

	// Both options should have been applied
	assert.Equal(t, mockClient, sender.client)
	assert.Equal(t, mockLog, sender.logger)
}

func TestApplyOptions_NoOptions(t *testing.T) {
	sender := &mockDestinationSender{}

	// Test with no options
	result := ApplyOptions(sender)

	// Should return the same sender instance
	assert.Same(t, sender, result)

	// Fields should remain nil
	assert.Nil(t, sender.client)
	assert.Nil(t, sender.logger)
}

func TestApplyOptions_NilSender(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockHTTPClient(ctrl)

	// Test with nil sender - should not panic
	assert.NotPanics(t, func() {
		result := ApplyOptions(nil, WithHTTPClient(mockClient))
		assert.Nil(t, result)
	})
}
