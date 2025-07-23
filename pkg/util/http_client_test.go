package util

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultHTTPClient(t *testing.T) {
	client := DefaultHTTPClient()

	// Test that we get a valid HTTP client
	require.NotNil(t, client)

	// Type assert to access underlying http.Client
	httpClient, ok := client.(*http.Client)
	require.True(t, ok, "Expected client to be *http.Client")

	// Test that timeout is configured
	assert.Equal(t, 30*time.Second, httpClient.Timeout)

	// Test that transport is configured
	require.NotNil(t, httpClient.Transport)

	transport, ok := httpClient.Transport.(*http.Transport)
	require.True(t, ok, "Expected transport to be *http.Transport")

	// Test connection pool settings
	assert.Equal(t, 100, transport.MaxIdleConns)
	assert.Equal(t, 10, transport.MaxIdleConnsPerHost)         // Fixed: should be 10, not 30
	assert.Equal(t, 90*time.Second, transport.IdleConnTimeout) // Fixed: should be 90s, not 30s
}

func TestGetSharedHTTPClient(t *testing.T) {
	// Test that we always get the same instance (singleton pattern)
	client1 := GetSharedHTTPClient()
	client2 := GetSharedHTTPClient()

	require.NotNil(t, client1)
	require.NotNil(t, client2)

	// Both calls should return the same instance
	assert.Same(t, client1, client2, "GetSharedHTTPClient should return the same instance")

	// Test that the shared client has the same configuration as DefaultHTTPClient
	defaultClient := DefaultHTTPClient()

	// Type assert both clients
	sharedHTTPClient, ok1 := client1.(*http.Client)
	defaultHTTPClient, ok2 := defaultClient.(*http.Client)

	require.True(t, ok1, "Expected shared client to be *http.Client")
	require.True(t, ok2, "Expected default client to be *http.Client")

	assert.Equal(t, defaultHTTPClient.Timeout, sharedHTTPClient.Timeout)

	// Transport types should be the same
	assert.IsType(t, defaultHTTPClient.Transport, sharedHTTPClient.Transport)
}
