package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

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

func TestPingRoute(t *testing.T) {
	router := setupRouter()

	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "pong", w.Body.String())
}

func TestStartServer(t *testing.T) {
	router := setupRouter()

	gin.SetMode(gin.TestMode)

	err := StartServer(router)
	assert.NoError(t, err)
}
