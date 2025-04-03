package utils

import (
	"net"
	"net/http"
	"time"
)

// HTTPClient defines the methods needed for sending HTTP requests
//
//go:generate mockgen -destination=../../mocks/http_client_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/utils HTTPClient
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// DefaultHTTPClient returns a standard HTTP client with sane defaults
func DefaultHTTPClient() HTTPClient {
	return &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 5 * time.Second,
			MaxIdleConns:        100,
			IdleConnTimeout:     90 * time.Second,
		},
	}
}
