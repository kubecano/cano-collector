package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/getsentry/sentry-go"

	"github.com/stretchr/testify/assert"
)

func TestInitSentry(t *testing.T) {
	err := initSentry("https://xxx@yyy.example.com/111")
	assert.NoError(t, err)
}

func TestInitSentry_Fail(t *testing.T) {
	sentryDsn := sentry.ClientOptions{Dsn: "foo"}
	err := sentry.Init(sentryDsn)

	assert.Error(t, err, "Expected an error when DSN is invalid")
}

func TestHelloWorld(t *testing.T) {
	router := setupRouter()

	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Hello world!", w.Body.String())
}

func TestStartServer(t *testing.T) {
	router := setupRouter()

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
		if err == nil && resp.StatusCode == http.StatusOK {
			if err := resp.Body.Close(); err != nil {
				t.Logf("Failed to close response body: %v", err)
			}
			serverReady = true
			break
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
