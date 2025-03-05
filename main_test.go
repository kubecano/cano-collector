package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"syscall"
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

	ts := httptest.NewServer(router)
	defer ts.Close()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		StartServer(router)
	}()
	// Wait for server to start
	time.Sleep(1 * time.Second)

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Fatalf("Failed to close response body: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Unexpected response code: got %v want %v", resp.StatusCode, http.StatusOK)
	}

	quit <- syscall.SIGTERM

	time.Sleep(2 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
		t.Error("Server did not shut down in time")
	case <-quit:
		t.Log("Server shut down successfully")
	}
}
