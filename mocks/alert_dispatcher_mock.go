// Code generated by MockGen. DO NOT EDIT.
// Source: dispatcher.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	config_team "github.com/kubecano/cano-collector/config/team"
	issue "github.com/kubecano/cano-collector/pkg/core/issue"
)

// MockAlertDispatcherInterface is a mock of AlertDispatcherInterface interface.
type MockAlertDispatcherInterface struct {
	ctrl     *gomock.Controller
	recorder *MockAlertDispatcherInterfaceMockRecorder
}

// MockAlertDispatcherInterfaceMockRecorder is the mock recorder for MockAlertDispatcherInterface.
type MockAlertDispatcherInterfaceMockRecorder struct {
	mock *MockAlertDispatcherInterface
}

// NewMockAlertDispatcherInterface creates a new mock instance.
func NewMockAlertDispatcherInterface(ctrl *gomock.Controller) *MockAlertDispatcherInterface {
	mock := &MockAlertDispatcherInterface{ctrl: ctrl}
	mock.recorder = &MockAlertDispatcherInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAlertDispatcherInterface) EXPECT() *MockAlertDispatcherInterfaceMockRecorder {
	return m.recorder
}

// DispatchIssues mocks base method.
func (m *MockAlertDispatcherInterface) DispatchIssues(ctx context.Context, issues []*issue.Issue, team *config_team.Team) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DispatchIssues", ctx, issues, team)
	ret0, _ := ret[0].(error)
	return ret0
}

// DispatchIssues indicates an expected call of DispatchIssues.
func (mr *MockAlertDispatcherInterfaceMockRecorder) DispatchIssues(ctx, issues, team interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DispatchIssues", reflect.TypeOf((*MockAlertDispatcherInterface)(nil).DispatchIssues), ctx, issues, team)
}
