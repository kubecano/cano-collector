package main

import (
	"testing"

	"github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/assert"
)

func resetSentryState() {
	sentry.Flush(0)
	_ = sentry.Init(sentry.ClientOptions{Dsn: ""})
}

func TestInitSentry_Success(t *testing.T) {
	defer resetSentryState()

	err := initSentry("https://xxx@yyy.example.com/111")
	assert.NoError(t, err, "Expected no error when DSN is valid")
}

func TestInitSentry_Fail(t *testing.T) {
	defer resetSentryState()

	err := initSentry("invalid-dsn")
	assert.Error(t, err, "Expected an error when DSN is invalid")
}

func TestInitSentry_Disabled(t *testing.T) {
	defer resetSentryState()

	err := initSentry("")
	assert.NoError(t, err, "Expected no error when Sentry DSN is empty")
}
