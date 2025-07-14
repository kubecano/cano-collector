package interfaces

import (
	"github.com/kubecano/cano-collector/pkg/core/event"
	issuepkg "github.com/kubecano/cano-collector/pkg/core/issue"
)

//go:generate mockgen -source=converter.go -destination=../../../mocks/alert_converter_mock.go -package=mocks
type ConverterInterface interface {
	ConvertAlertManagerEventToIssues(event *event.AlertManagerEvent) ([]*issuepkg.Issue, error)
}
