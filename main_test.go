package main

import (
	"github.com/getsentry/sentry-go"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitSentry(t *testing.T) {
	err := initSentry()
	assert.NoError(t, err)
}

func TestInitSentry_Fail(t *testing.T) {
	sentryDsn := sentry.ClientOptions{Dsn: ""}
	err := sentry.Init(sentryDsn)

	assert.Error(t, err, "Expected an error when DSN is empty")
}

func TestHelloWorld(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Hello world!", w.Body.String())
}

func TestPingRoute(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "pong", w.Body.String())
}
