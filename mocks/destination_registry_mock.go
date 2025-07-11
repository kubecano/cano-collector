// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/kubecano/cano-collector/pkg/destination/interfaces (interfaces: DestinationRegistryInterface)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	config_destination "github.com/kubecano/cano-collector/config/destination"
	interfaces "github.com/kubecano/cano-collector/pkg/destination/interfaces"
)

// MockDestinationRegistryInterface is a mock of DestinationRegistryInterface interface.
type MockDestinationRegistryInterface struct {
	ctrl     *gomock.Controller
	recorder *MockDestinationRegistryInterfaceMockRecorder
}

// MockDestinationRegistryInterfaceMockRecorder is the mock recorder for MockDestinationRegistryInterface.
type MockDestinationRegistryInterfaceMockRecorder struct {
	mock *MockDestinationRegistryInterface
}

// NewMockDestinationRegistryInterface creates a new mock instance.
func NewMockDestinationRegistryInterface(ctrl *gomock.Controller) *MockDestinationRegistryInterface {
	mock := &MockDestinationRegistryInterface{ctrl: ctrl}
	mock.recorder = &MockDestinationRegistryInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDestinationRegistryInterface) EXPECT() *MockDestinationRegistryInterfaceMockRecorder {
	return m.recorder
}

// GetDestination mocks base method.
func (m *MockDestinationRegistryInterface) GetDestination(arg0 string) (interfaces.DestinationInterface, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDestination", arg0)
	ret0, _ := ret[0].(interfaces.DestinationInterface)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDestination indicates an expected call of GetDestination.
func (mr *MockDestinationRegistryInterfaceMockRecorder) GetDestination(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDestination", reflect.TypeOf((*MockDestinationRegistryInterface)(nil).GetDestination), arg0)
}

// GetDestinations mocks base method.
func (m *MockDestinationRegistryInterface) GetDestinations(arg0 []string) ([]interfaces.DestinationInterface, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDestinations", arg0)
	ret0, _ := ret[0].([]interfaces.DestinationInterface)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDestinations indicates an expected call of GetDestinations.
func (mr *MockDestinationRegistryInterfaceMockRecorder) GetDestinations(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDestinations", reflect.TypeOf((*MockDestinationRegistryInterface)(nil).GetDestinations), arg0)
}

// LoadFromConfig mocks base method.
func (m *MockDestinationRegistryInterface) LoadFromConfig(arg0 config_destination.DestinationsConfig) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LoadFromConfig", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// LoadFromConfig indicates an expected call of LoadFromConfig.
func (mr *MockDestinationRegistryInterfaceMockRecorder) LoadFromConfig(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LoadFromConfig", reflect.TypeOf((*MockDestinationRegistryInterface)(nil).LoadFromConfig), arg0)
}

// RegisterDestination mocks base method.
func (m *MockDestinationRegistryInterface) RegisterDestination(arg0 string, arg1 interfaces.DestinationInterface) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RegisterDestination", arg0, arg1)
}

// RegisterDestination indicates an expected call of RegisterDestination.
func (mr *MockDestinationRegistryInterfaceMockRecorder) RegisterDestination(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterDestination", reflect.TypeOf((*MockDestinationRegistryInterface)(nil).RegisterDestination), arg0, arg1)
}
