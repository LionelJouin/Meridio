// Code generated by MockGen. DO NOT EDIT.
// Source: ./pkg/ambassador/tap/types/registry.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	ambassador "github.com/nordix/meridio/api/ambassador/v1"
	v1 "github.com/nordix/meridio/api/nsp/v1"
	types "github.com/nordix/meridio/pkg/ambassador/tap/types"
)

// MockRegistry is a mock of Registry interface.
type MockRegistry struct {
	ctrl     *gomock.Controller
	recorder *MockRegistryMockRecorder
}

// MockRegistryMockRecorder is the mock recorder for MockRegistry.
type MockRegistryMockRecorder struct {
	mock *MockRegistry
}

// NewMockRegistry creates a new mock instance.
func NewMockRegistry(ctrl *gomock.Controller) *MockRegistry {
	mock := &MockRegistry{ctrl: ctrl}
	mock.recorder = &MockRegistryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRegistry) EXPECT() *MockRegistryMockRecorder {
	return m.recorder
}

// Add mocks base method.
func (m *MockRegistry) Add(arg0 context.Context, arg1 *v1.Stream) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Add", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Add indicates an expected call of Add.
func (mr *MockRegistryMockRecorder) Add(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockRegistry)(nil).Add), arg0, arg1)
}

// Remove mocks base method.
func (m *MockRegistry) Remove(arg0 context.Context, arg1 *v1.Stream) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Remove", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Remove indicates an expected call of Remove.
func (mr *MockRegistryMockRecorder) Remove(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Remove", reflect.TypeOf((*MockRegistry)(nil).Remove), arg0, arg1)
}

// SetStatus mocks base method.
func (m *MockRegistry) SetStatus(arg0 *v1.Stream, arg1 ambassador.StreamStatus_Status) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetStatus", arg0, arg1)
}

// SetStatus indicates an expected call of SetStatus.
func (mr *MockRegistryMockRecorder) SetStatus(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetStatus", reflect.TypeOf((*MockRegistry)(nil).SetStatus), arg0, arg1)
}

// Watch mocks base method.
func (m *MockRegistry) Watch(arg0 context.Context, arg1 *v1.Stream) (types.Watcher, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Watch", arg0, arg1)
	ret0, _ := ret[0].(types.Watcher)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Watch indicates an expected call of Watch.
func (mr *MockRegistryMockRecorder) Watch(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Watch", reflect.TypeOf((*MockRegistry)(nil).Watch), arg0, arg1)
}

// MockWatcher is a mock of Watcher interface.
type MockWatcher struct {
	ctrl     *gomock.Controller
	recorder *MockWatcherMockRecorder
}

// MockWatcherMockRecorder is the mock recorder for MockWatcher.
type MockWatcherMockRecorder struct {
	mock *MockWatcher
}

// NewMockWatcher creates a new mock instance.
func NewMockWatcher(ctrl *gomock.Controller) *MockWatcher {
	mock := &MockWatcher{ctrl: ctrl}
	mock.recorder = &MockWatcherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWatcher) EXPECT() *MockWatcherMockRecorder {
	return m.recorder
}

// ResultChan mocks base method.
func (m *MockWatcher) ResultChan() <-chan []*ambassador.StreamStatus {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ResultChan")
	ret0, _ := ret[0].(<-chan []*ambassador.StreamStatus)
	return ret0
}

// ResultChan indicates an expected call of ResultChan.
func (mr *MockWatcherMockRecorder) ResultChan() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ResultChan", reflect.TypeOf((*MockWatcher)(nil).ResultChan))
}

// Stop mocks base method.
func (m *MockWatcher) Stop() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Stop")
}

// Stop indicates an expected call of Stop.
func (mr *MockWatcherMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockWatcher)(nil).Stop))
}
