package main

import (
	"context"
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
	// Utwórz router Gin
	router := setupRouter()

	// Utwórz testowy serwer HTTP
	ts := httptest.NewServer(router)
	defer ts.Close()

	// Kanał na sygnały do zamknięcia serwera
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Uruchomienie serwera w gorutynie
	go func() {
		StartServer(router)
	}()

	// Poczekaj na uruchomienie serwera
	time.Sleep(1 * time.Second)

	// Sprawdź, czy serwer odpowiada na żądania
	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Unexpected response code: got %v want %v", resp.StatusCode, http.StatusOK)
	}

	// Symulacja zamknięcia serwera
	quit <- syscall.SIGTERM

	// Poczekaj na zamknięcie
	time.Sleep(2 * time.Second)

	// Sprawdzenie czy serwer został zamknięty poprawnie
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	select {
	case <-ctx.Done():
		t.Error("Server did not shut down in time")
	default:
	}
}
