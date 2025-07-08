package slack

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/slack-go/slack"
	"go.uber.org/zap"

	logger_interfaces "github.com/kubecano/cano-collector/pkg/logger/interfaces"
	sender_interfaces "github.com/kubecano/cano-collector/pkg/sender/interfaces"
)

type ThreadManager struct {
	client       sender_interfaces.SlackClientInterface
	channel      string
	logger       logger_interfaces.LoggerInterface
	cache        map[string]*threadCacheEntry
	cacheMutex   sync.RWMutex
	cacheTTL     time.Duration
	searchLimit  int
	searchWindow time.Duration
}

type threadCacheEntry struct {
	threadTS  string
	timestamp time.Time
}

func NewThreadManager(client sender_interfaces.SlackClientInterface, channel string, logger logger_interfaces.LoggerInterface, cacheTTL time.Duration, searchLimit int, searchWindow time.Duration) *ThreadManager {
	return &ThreadManager{
		client:       client,
		channel:      channel,
		logger:       logger,
		cache:        make(map[string]*threadCacheEntry),
		cacheTTL:     cacheTTL,
		searchLimit:  searchLimit,
		searchWindow: searchWindow,
	}
}

func (tm *ThreadManager) GetThreadTS(ctx context.Context, fingerprint string) (string, error) {
	// First check cache
	tm.cacheMutex.RLock()
	entry, exists := tm.cache[fingerprint]
	tm.cacheMutex.RUnlock()

	if exists && time.Since(entry.timestamp) < tm.cacheTTL {
		tm.logger.Debug("Thread found in cache", zap.String("fingerprint", fingerprint), zap.String("threadTS", entry.threadTS))
		return entry.threadTS, nil
	}

	// Search Slack for existing message with this fingerprint
	threadTS, err := tm.searchSlackForThread(fingerprint)
	if err != nil {
		tm.logger.Error("Failed to search Slack for thread", zap.String("fingerprint", fingerprint), zap.Error(err))
		return "", err
	}

	// Cache the result if found
	if threadTS != "" {
		tm.SetThreadTS(fingerprint, threadTS)
		tm.logger.Debug("Thread found via Slack search", zap.String("fingerprint", fingerprint), zap.String("threadTS", threadTS))
	}

	return threadTS, nil
}

func (tm *ThreadManager) SetThreadTS(fingerprint, threadTS string) {
	tm.cacheMutex.Lock()
	tm.cache[fingerprint] = &threadCacheEntry{
		threadTS:  threadTS,
		timestamp: time.Now(),
	}
	tm.cacheMutex.Unlock()
	tm.logger.Debug("Thread cached", zap.String("fingerprint", fingerprint), zap.String("threadTS", threadTS))
}

func (tm *ThreadManager) InvalidateThread(fingerprint string) {
	tm.cacheMutex.Lock()
	delete(tm.cache, fingerprint)
	tm.cacheMutex.Unlock()
	tm.logger.Debug("Thread invalidated", zap.String("fingerprint", fingerprint))
}

func (tm *ThreadManager) Cleanup() {
	tm.cacheMutex.Lock()
	defer tm.cacheMutex.Unlock()

	now := time.Now()
	for fingerprint, entry := range tm.cache {
		if now.Sub(entry.timestamp) > tm.cacheTTL {
			delete(tm.cache, fingerprint)
			tm.logger.Debug("Thread cache entry expired", zap.String("fingerprint", fingerprint))
		}
	}
}

// searchSlackForThread searches Slack for a message containing the fingerprint
func (tm *ThreadManager) searchSlackForThread(fingerprint string) (string, error) {
	// Calculate search timeframe
	earliest := time.Now().Add(-tm.searchWindow)

	// Use conversations.history to get recent messages
	params := &slack.GetConversationHistoryParameters{
		ChannelID: tm.channel,
		Oldest:    fmt.Sprintf("%.6f", float64(earliest.Unix())),
		Limit:     tm.searchLimit,
		Inclusive: true,
	}

	historyResponse, err := tm.client.GetConversationHistory(params)
	if err != nil {
		return "", fmt.Errorf("failed to get conversation history: %w", err)
	}

	// Look through messages for one containing the fingerprint
	for _, message := range historyResponse.Messages {
		// Check if message text contains fingerprint
		if message.Text != "" && tm.messageContainsFingerprint(message.Text, fingerprint) {
			tm.logger.Debug("Found matching message in conversation history",
				zap.String("fingerprint", fingerprint),
				zap.String("timestamp", message.Timestamp),
				zap.String("channel", tm.channel))

			return message.Timestamp, nil
		}

		// Also check message attachments for fingerprint
		for _, attachment := range message.Attachments {
			if attachment.Text != "" && tm.messageContainsFingerprint(attachment.Text, fingerprint) {
				tm.logger.Debug("Found matching message in attachment",
					zap.String("fingerprint", fingerprint),
					zap.String("timestamp", message.Timestamp),
					zap.String("channel", tm.channel))

				return message.Timestamp, nil
			}
		}
	}

	return "", nil
}

// messageContainsFingerprint checks if a message contains the alert fingerprint
func (tm *ThreadManager) messageContainsFingerprint(text, fingerprint string) bool {
	// Simple string contains check - could be more sophisticated
	return len(fingerprint) > 0 && len(text) > 0 &&
		(text == fingerprint ||
			fmt.Sprintf("```%s```", fingerprint) == text ||
			fmt.Sprintf("`%s`", fingerprint) == text)
}
