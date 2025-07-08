package slack

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubecano/cano-collector/mocks"
)

func TestNewThreadManager(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockSlackClientInterface(ctrl)
	mockLogger := mocks.NewMockLoggerInterface(ctrl)

	cacheTTL := 10 * time.Minute
	searchLimit := 50
	searchWindow := 24 * time.Hour
	channel := "#test-channel"

	tm := NewThreadManager(mockClient, channel, mockLogger, cacheTTL, searchLimit, searchWindow)

	assert.NotNil(t, tm)
	assert.Equal(t, mockClient, tm.client)
	assert.Equal(t, channel, tm.channel)
	assert.Equal(t, mockLogger, tm.logger)
	assert.Equal(t, cacheTTL, tm.cacheTTL)
	assert.Equal(t, searchLimit, tm.searchLimit)
	assert.Equal(t, searchWindow, tm.searchWindow)
	assert.NotNil(t, tm.cache)
	assert.Empty(t, tm.cache)
}

func TestThreadManager_SetThreadTS(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockSlackClientInterface(ctrl)
	mockLogger := mocks.NewMockLoggerInterface(ctrl)

	// Expect debug log call
	mockLogger.EXPECT().Debug("Thread cached", gomock.Any()).Times(1)

	tm := NewThreadManager(mockClient, "#test", mockLogger, 10*time.Minute, 50, 24*time.Hour)

	fingerprint := "test-fingerprint"
	threadTS := "1234567890.123456"

	tm.SetThreadTS(fingerprint, threadTS)

	// Verify cache entry was created
	tm.cacheMutex.RLock()
	entry, exists := tm.cache[fingerprint]
	tm.cacheMutex.RUnlock()

	assert.True(t, exists)
	assert.Equal(t, threadTS, entry.threadTS)
	assert.WithinDuration(t, time.Now(), entry.timestamp, 1*time.Second)
}

func TestThreadManager_InvalidateThread(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockSlackClientInterface(ctrl)
	mockLogger := mocks.NewMockLoggerInterface(ctrl)

	// Expect debug log calls
	mockLogger.EXPECT().Debug("Thread cached", gomock.Any()).Times(1)
	mockLogger.EXPECT().Debug("Thread invalidated", gomock.Any()).Times(1)

	tm := NewThreadManager(mockClient, "#test", mockLogger, 10*time.Minute, 50, 24*time.Hour)

	fingerprint := "test-fingerprint"
	threadTS := "1234567890.123456"

	// First set a thread
	tm.SetThreadTS(fingerprint, threadTS)

	// Verify it exists
	tm.cacheMutex.RLock()
	_, exists := tm.cache[fingerprint]
	tm.cacheMutex.RUnlock()
	assert.True(t, exists)

	// Invalidate it
	tm.InvalidateThread(fingerprint)

	// Verify it's gone
	tm.cacheMutex.RLock()
	_, exists = tm.cache[fingerprint]
	tm.cacheMutex.RUnlock()
	assert.False(t, exists)
}

func TestThreadManager_GetThreadTS_FromCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockSlackClientInterface(ctrl)
	mockLogger := mocks.NewMockLoggerInterface(ctrl)

	// Expect debug log calls
	mockLogger.EXPECT().Debug("Thread cached", gomock.Any()).Times(1)
	mockLogger.EXPECT().Debug("Thread found in cache", gomock.Any()).Times(1)

	tm := NewThreadManager(mockClient, "#test", mockLogger, 10*time.Minute, 50, 24*time.Hour)

	fingerprint := "test-fingerprint"
	threadTS := "1234567890.123456"

	// Set a thread in cache
	tm.SetThreadTS(fingerprint, threadTS)

	// Get it from cache
	ctx := context.Background()
	result, err := tm.GetThreadTS(ctx, fingerprint)

	require.NoError(t, err)
	assert.Equal(t, threadTS, result)
}

func TestThreadManager_GetThreadTS_CacheExpired(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockSlackClientInterface(ctrl)
	mockLogger := mocks.NewMockLoggerInterface(ctrl)

	// Expect debug log calls
	mockLogger.EXPECT().Debug("Thread cached", gomock.Any()).Times(1)

	// Very short TTL for testing
	tm := NewThreadManager(mockClient, "#test", mockLogger, 1*time.Millisecond, 50, 24*time.Hour)

	fingerprint := "test-fingerprint"
	threadTS := "1234567890.123456"

	// Set a thread in cache
	tm.SetThreadTS(fingerprint, threadTS)

	// Wait for cache to expire
	time.Sleep(2 * time.Millisecond)

	// Mock Slack search returning no results
	mockClient.EXPECT().GetConversationHistory(gomock.Any()).Return(&slack.GetConversationHistoryResponse{
		Messages: []slack.Message{},
	}, nil).Times(1)

	ctx := context.Background()
	result, err := tm.GetThreadTS(ctx, fingerprint)

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestThreadManager_GetThreadTS_SlackSearchSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockSlackClientInterface(ctrl)
	mockLogger := mocks.NewMockLoggerInterface(ctrl)

	// Expect debug log calls
	mockLogger.EXPECT().Debug("Found matching message in conversation history", gomock.Any()).Times(1)
	mockLogger.EXPECT().Debug("Thread found via Slack search", gomock.Any()).Times(1)
	mockLogger.EXPECT().Debug("Thread cached", gomock.Any()).Times(1)

	tm := NewThreadManager(mockClient, "#test", mockLogger, 10*time.Minute, 50, 24*time.Hour)

	fingerprint := "test-fingerprint"
	threadTS := "1234567890.123456"

	// Mock Slack search returning a matching message
	mockClient.EXPECT().GetConversationHistory(gomock.Any()).Return(&slack.GetConversationHistoryResponse{
		Messages: []slack.Message{
			{
				Msg: slack.Msg{
					Text:      fmt.Sprintf("```%s```", fingerprint),
					Timestamp: threadTS,
				},
			},
		},
	}, nil).Times(1)

	ctx := context.Background()
	result, err := tm.GetThreadTS(ctx, fingerprint)

	require.NoError(t, err)
	assert.Equal(t, threadTS, result)
}

func TestThreadManager_GetThreadTS_SlackSearchError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockSlackClientInterface(ctrl)
	mockLogger := mocks.NewMockLoggerInterface(ctrl)

	// Expect error log call
	mockLogger.EXPECT().Error("Failed to search Slack for thread", gomock.Any()).Times(1)

	tm := NewThreadManager(mockClient, "#test", mockLogger, 10*time.Minute, 50, 24*time.Hour)

	fingerprint := "test-fingerprint"

	// Mock Slack search returning error
	expectedError := fmt.Errorf("slack api error")
	mockClient.EXPECT().GetConversationHistory(gomock.Any()).Return(nil, expectedError).Times(1)

	ctx := context.Background()
	result, err := tm.GetThreadTS(ctx, fingerprint)

	require.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "slack api error")
}

func TestThreadManager_Cleanup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockSlackClientInterface(ctrl)
	mockLogger := mocks.NewMockLoggerInterface(ctrl)

	// Expect debug log calls - use AnyTimes() to avoid strict ordering issues
	mockLogger.EXPECT().Debug("Thread cached", gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug("Thread cache entry expired", gomock.Any()).AnyTimes()

	// Very short TTL for testing
	tm := NewThreadManager(mockClient, "#test", mockLogger, 1*time.Millisecond, 50, 24*time.Hour)

	fingerprint1 := "test-fingerprint-1"
	fingerprint2 := "test-fingerprint-2"
	threadTS1 := "1234567890.123456"
	threadTS2 := "1234567890.123457"

	// Set threads in cache
	tm.SetThreadTS(fingerprint1, threadTS1)
	tm.SetThreadTS(fingerprint2, threadTS2)

	// Verify both exist
	tm.cacheMutex.RLock()
	assert.Len(t, tm.cache, 2)
	tm.cacheMutex.RUnlock()

	// Wait for entries to expire
	time.Sleep(5 * time.Millisecond)

	// Run cleanup
	tm.Cleanup()

	// Verify expired entries were removed
	tm.cacheMutex.RLock()
	assert.Empty(t, tm.cache)
	tm.cacheMutex.RUnlock()
}

func TestThreadManager_searchSlackForThread_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockSlackClientInterface(ctrl)
	mockLogger := mocks.NewMockLoggerInterface(ctrl)

	// Expect debug log call
	mockLogger.EXPECT().Debug("Found matching message in conversation history", gomock.Any()).Times(1)

	tm := NewThreadManager(mockClient, "#test", mockLogger, 10*time.Minute, 50, 24*time.Hour)

	fingerprint := "test-fingerprint"
	threadTS := "1234567890.123456"

	// Mock successful conversation history response
	mockClient.EXPECT().GetConversationHistory(gomock.Any()).Return(&slack.GetConversationHistoryResponse{
		Messages: []slack.Message{
			{
				Msg: slack.Msg{
					Text:      fmt.Sprintf("```%s```", fingerprint),
					Timestamp: threadTS,
				},
			},
		},
	}, nil).Times(1)

	result, err := tm.searchSlackForThread(fingerprint)

	require.NoError(t, err)
	assert.Equal(t, threadTS, result)
}

func TestThreadManager_searchSlackForThread_AttachmentMatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockSlackClientInterface(ctrl)
	mockLogger := mocks.NewMockLoggerInterface(ctrl)

	// Expect debug log call
	mockLogger.EXPECT().Debug("Found matching message in attachment", gomock.Any()).Times(1)

	tm := NewThreadManager(mockClient, "#test", mockLogger, 10*time.Minute, 50, 24*time.Hour)

	fingerprint := "test-fingerprint"
	threadTS := "1234567890.123456"

	// Mock conversation history response with attachment match
	mockClient.EXPECT().GetConversationHistory(gomock.Any()).Return(&slack.GetConversationHistoryResponse{
		Messages: []slack.Message{
			{
				Msg: slack.Msg{
					Text:      "Some other text",
					Timestamp: threadTS,
					Attachments: []slack.Attachment{
						{
							Text: fmt.Sprintf("```%s```", fingerprint),
						},
					},
				},
			},
		},
	}, nil).Times(1)

	result, err := tm.searchSlackForThread(fingerprint)

	require.NoError(t, err)
	assert.Equal(t, threadTS, result)
}

func TestThreadManager_searchSlackForThread_NoMatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockSlackClientInterface(ctrl)
	mockLogger := mocks.NewMockLoggerInterface(ctrl)

	tm := NewThreadManager(mockClient, "#test", mockLogger, 10*time.Minute, 50, 24*time.Hour)

	fingerprint := "test-fingerprint"

	// Mock conversation history response with no matches
	mockClient.EXPECT().GetConversationHistory(gomock.Any()).Return(&slack.GetConversationHistoryResponse{
		Messages: []slack.Message{
			{
				Msg: slack.Msg{
					Text:      "Some other text",
					Timestamp: "1234567890.123456",
				},
			},
		},
	}, nil).Times(1)

	result, err := tm.searchSlackForThread(fingerprint)

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestThreadManager_searchSlackForThread_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockSlackClientInterface(ctrl)
	mockLogger := mocks.NewMockLoggerInterface(ctrl)

	tm := NewThreadManager(mockClient, "#test", mockLogger, 10*time.Minute, 50, 24*time.Hour)

	fingerprint := "test-fingerprint"

	// Mock conversation history error
	expectedError := fmt.Errorf("slack api error")
	mockClient.EXPECT().GetConversationHistory(gomock.Any()).Return(nil, expectedError).Times(1)

	result, err := tm.searchSlackForThread(fingerprint)

	require.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "failed to get conversation history")
}

func TestThreadManager_messageContainsFingerprint(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockSlackClientInterface(ctrl)
	mockLogger := mocks.NewMockLoggerInterface(ctrl)

	tm := NewThreadManager(mockClient, "#test", mockLogger, 10*time.Minute, 50, 24*time.Hour)

	tests := []struct {
		name        string
		text        string
		fingerprint string
		expected    bool
	}{
		{
			name:        "exact match",
			text:        "test-fingerprint",
			fingerprint: "test-fingerprint",
			expected:    true,
		},
		{
			name:        "code block match",
			text:        "```test-fingerprint```",
			fingerprint: "test-fingerprint",
			expected:    true,
		},
		{
			name:        "inline code match",
			text:        "`test-fingerprint`",
			fingerprint: "test-fingerprint",
			expected:    true,
		},
		{
			name:        "no match",
			text:        "some other text",
			fingerprint: "test-fingerprint",
			expected:    false,
		},
		{
			name:        "empty text",
			text:        "",
			fingerprint: "test-fingerprint",
			expected:    false,
		},
		{
			name:        "empty fingerprint",
			text:        "some text",
			fingerprint: "",
			expected:    false,
		},
		{
			name:        "both empty",
			text:        "",
			fingerprint: "",
			expected:    false,
		},
		{
			name:        "partial match",
			text:        "test-fingerprint-other",
			fingerprint: "test-fingerprint",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tm.messageContainsFingerprint(tt.text, tt.fingerprint)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestThreadManager_GetThreadTS_ConcurrentAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockSlackClientInterface(ctrl)
	mockLogger := mocks.NewMockLoggerInterface(ctrl)

	// Expect debug log calls (may be called multiple times due to concurrency)
	mockLogger.EXPECT().Debug("Thread cached", gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug("Thread found in cache", gomock.Any()).AnyTimes()

	tm := NewThreadManager(mockClient, "#test", mockLogger, 10*time.Minute, 50, 24*time.Hour)

	fingerprint := "test-fingerprint"
	threadTS := "1234567890.123456"

	// Set a thread in cache
	tm.SetThreadTS(fingerprint, threadTS)

	// Test concurrent access
	done := make(chan bool)
	errors := make(chan error, 10)

	for range 10 {
		go func() {
			defer func() { done <- true }()
			ctx := context.Background()
			result, err := tm.GetThreadTS(ctx, fingerprint)
			if err != nil {
				errors <- err
				return
			}
			if result != threadTS {
				errors <- fmt.Errorf("expected %s, got %s", threadTS, result)
				return
			}
		}()
	}

	// Wait for all goroutines to complete
	for range 10 {
		<-done
	}

	// Check for errors
	select {
	case err := <-errors:
		t.Fatalf("Concurrent access error: %v", err)
	default:
		// No errors, test passed
	}
}

func TestThreadManager_SetThreadTS_ConcurrentAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockSlackClientInterface(ctrl)
	mockLogger := mocks.NewMockLoggerInterface(ctrl)

	// Expect debug log calls (may be called multiple times due to concurrency)
	mockLogger.EXPECT().Debug("Thread cached", gomock.Any()).AnyTimes()

	tm := NewThreadManager(mockClient, "#test", mockLogger, 10*time.Minute, 50, 24*time.Hour)

	done := make(chan bool)

	// Test concurrent writes
	for i := range 10 {
		go func(index int) {
			defer func() { done <- true }()
			fingerprint := fmt.Sprintf("test-fingerprint-%d", index)
			threadTS := fmt.Sprintf("1234567890.12345%d", index)
			tm.SetThreadTS(fingerprint, threadTS)
		}(i)
	}

	// Wait for all goroutines to complete
	for range 10 {
		<-done
	}

	// Verify all entries were set
	tm.cacheMutex.RLock()
	assert.Len(t, tm.cache, 10)
	tm.cacheMutex.RUnlock()
}
