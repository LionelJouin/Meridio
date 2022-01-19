// Code generated by MockGen. DO NOT EDIT.
// Source: ./pkg/ambassador/tap/types/conduit.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v1 "github.com/nordix/meridio/api/nsp/v1"
	types "github.com/nordix/meridio/pkg/ambassador/tap/types"
)

// MockConduit is a mock of Conduit interface.
type MockConduit struct {
	ctrl     *gomock.Controller
	recorder *MockConduitMockRecorder
}

// MockConduitMockRecorder is the mock recorder for MockConduit.
type MockConduitMockRecorder struct {
	mock *MockConduit
}

// NewMockConduit creates a new mock instance.
func NewMockConduit(ctrl *gomock.Controller) *MockConduit {
	mock := &MockConduit{ctrl: ctrl}
	mock.recorder = &MockConduitMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockConduit) EXPECT() *MockConduitMockRecorder {
	return m.recorder
}

// AddStream mocks base method.
func (m *MockConduit) AddStream(arg0 context.Context, arg1 *v1.Stream) (types.Stream, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddStream", arg0, arg1)
	ret0, _ := ret[0].(types.Stream)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddStream indicates an expected call of AddStream.
func (mr *MockConduitMockRecorder) AddStream(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddStream", reflect.TypeOf((*MockConduit)(nil).AddStream), arg0, arg1)
}

// Connect mocks base method.
func (m *MockConduit) Connect(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Connect", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Connect indicates an expected call of Connect.
func (mr *MockConduitMockRecorder) Connect(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Connect", reflect.TypeOf((*MockConduit)(nil).Connect), ctx)
}

// Disconnect mocks base method.
func (m *MockConduit) Disconnect(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Disconnect", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Disconnect indicates an expected call of Disconnect.
func (mr *MockConduitMockRecorder) Disconnect(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Disconnect", reflect.TypeOf((*MockConduit)(nil).Disconnect), ctx)
}

// Equals mocks base method.
func (m *MockConduit) Equals(arg0 *v1.Conduit) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Equals", arg0)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Equals indicates an expected call of Equals.
func (mr *MockConduitMockRecorder) Equals(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Equals", reflect.TypeOf((*MockConduit)(nil).Equals), arg0)
}

// GetStreams mocks base method.
func (m *MockConduit) GetStreams(arg0 context.Context) []types.Stream {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStreams", arg0)
	ret0, _ := ret[0].([]types.Stream)
	return ret0
}

// GetStreams indicates an expected call of GetStreams.
func (mr *MockConduitMockRecorder) GetStreams(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStreams", reflect.TypeOf((*MockConduit)(nil).GetStreams), arg0)
}

// RemoveStream mocks base method.
func (m *MockConduit) RemoveStream(arg0 context.Context, arg1 *v1.Stream) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveStream", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveStream indicates an expected call of RemoveStream.
func (mr *MockConduitMockRecorder) RemoveStream(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveStream", reflect.TypeOf((*MockConduit)(nil).RemoveStream), arg0, arg1)
}
