//go:build tools
// +build tools

package tools

import (
	_ "github.com/golang/mock/mockgen"
)

//go:generate mockgen -destination=../mocks/tracerprovider_mock.go -package=mocks go.opentelemetry.io/otel/trace TracerProvider
