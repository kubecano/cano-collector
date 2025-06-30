package util

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// HTTPClient defines the methods needed for sending HTTP requests
//
//go:generate mockgen -destination=../../mocks/http_client_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/util HTTPClient
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// DefaultHTTPClient returns a standard HTTP client with sane defaults and connection pooling
func DefaultHTTPClient() HTTPClient {
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 10 * time.Second,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			DisableCompression:  false,
		},
	}
}

// SharedHTTPClient is a singleton HTTP client for better connection pooling
var (
	sharedHTTPClient HTTPClient
	once             sync.Once
)

// GetSharedHTTPClient returns a shared HTTP client instance for better connection pooling
func GetSharedHTTPClient() HTTPClient {
	once.Do(func() {
		sharedHTTPClient = DefaultHTTPClient()
	})
	return sharedHTTPClient
}
