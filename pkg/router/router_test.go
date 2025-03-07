package router

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/kubecano/cano-collector/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestHelloWorld(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l, _ := zap.NewDevelopment()
	logger.SetLogger(l)
	router := SetupRouter(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Hello world!", w.Body.String())
}

func TestStartServer(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	l, _ := zap.NewDevelopment()
	logger.SetLogger(l)
	router := SetupRouter(nil)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	var wg sync.WaitGroup
	wg.Add(1)

	serverErrChan := make(chan error, 1)

	go func() {
		defer wg.Done()
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrChan <- err
		}
	}()

	// Wait for server to start using a simple health check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var serverReady bool
	for {
		resp, err := http.Get("http://127.0.0.1:8080/")
		if err == nil {
			func() {
				if err := resp.Body.Close(); err != nil {
					t.Logf("Failed to close response body: %v", err)
				}
			}()
			if resp.StatusCode == http.StatusOK {
				serverReady = true
				break
			}
		}
		if ctx.Err() != nil {
			t.Fatal("Server did not start in time")
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !serverReady {
		t.Fatal("Server did not start correctly")
	}

	select {
	case err := <-serverErrChan:
		if err != nil {
			t.Fatalf("Server encountered an error: %v", err)
		}
	default:
	}

	// Gracefully shut down the server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	shutdownErrChan := make(chan error, 1)
	go func() {
		shutdownErrChan <- srv.Shutdown(shutdownCtx)
	}()

	// Wait for server shutdown
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-shutdownCtx.Done():
		t.Fatal("Server did not shut down in time")
	case <-done:
		t.Log("Server shut down successfully")
	case err := <-shutdownErrChan:
		if err != nil {
			t.Fatalf("Server shutdown failed: %v", err)
		}
	}
}

func TestMetricsEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	l, _ := zap.NewDevelopment()
	logger.SetLogger(l)
	router := SetupRouter(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/metrics", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "http_requests_total")
}
