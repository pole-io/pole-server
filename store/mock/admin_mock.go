// Code generated by MockGen. DO NOT EDIT.
// Source: mysql/admin.go

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	admin "github.com/GovernSea/sergo-server/common/model/admin"
)

// MockLeaderElectionStore is a mock of LeaderElectionStore interface.
type MockLeaderElectionStore struct {
	ctrl     *gomock.Controller
	recorder *MockLeaderElectionStoreMockRecorder
}

// MockLeaderElectionStoreMockRecorder is the mock recorder for MockLeaderElectionStore.
type MockLeaderElectionStoreMockRecorder struct {
	mock *MockLeaderElectionStore
}

// NewMockLeaderElectionStore creates a new mock instance.
func NewMockLeaderElectionStore(ctrl *gomock.Controller) *MockLeaderElectionStore {
	mock := &MockLeaderElectionStore{ctrl: ctrl}
	mock.recorder = &MockLeaderElectionStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLeaderElectionStore) EXPECT() *MockLeaderElectionStoreMockRecorder {
	return m.recorder
}

// CheckMtimeExpired mocks base method.
func (m *MockLeaderElectionStore) CheckMtimeExpired(key string, leaseTime int32) (string, bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckMtimeExpired", key, leaseTime)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(bool)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// CheckMtimeExpired indicates an expected call of CheckMtimeExpired.
func (mr *MockLeaderElectionStoreMockRecorder) CheckMtimeExpired(key, leaseTime interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckMtimeExpired", reflect.TypeOf((*MockLeaderElectionStore)(nil).CheckMtimeExpired), key, leaseTime)
}

// CompareAndSwapVersion mocks base method.
func (m *MockLeaderElectionStore) CompareAndSwapVersion(key string, curVersion, newVersion int64, leader string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CompareAndSwapVersion", key, curVersion, newVersion, leader)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CompareAndSwapVersion indicates an expected call of CompareAndSwapVersion.
func (mr *MockLeaderElectionStoreMockRecorder) CompareAndSwapVersion(key, curVersion, newVersion, leader interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CompareAndSwapVersion", reflect.TypeOf((*MockLeaderElectionStore)(nil).CompareAndSwapVersion), key, curVersion, newVersion, leader)
}

// CreateLeaderElection mocks base method.
func (m *MockLeaderElectionStore) CreateLeaderElection(key string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateLeaderElection", key)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateLeaderElection indicates an expected call of CreateLeaderElection.
func (mr *MockLeaderElectionStoreMockRecorder) CreateLeaderElection(key interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateLeaderElection", reflect.TypeOf((*MockLeaderElectionStore)(nil).CreateLeaderElection), key)
}

// GetVersion mocks base method.
func (m *MockLeaderElectionStore) GetVersion(key string) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVersion", key)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetVersion indicates an expected call of GetVersion.
func (mr *MockLeaderElectionStoreMockRecorder) GetVersion(key interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVersion", reflect.TypeOf((*MockLeaderElectionStore)(nil).GetVersion), key)
}

// ListLeaderElections mocks base method.
func (m *MockLeaderElectionStore) ListLeaderElections() ([]*admin.LeaderElection, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListLeaderElections")
	ret0, _ := ret[0].([]*admin.LeaderElection)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListLeaderElections indicates an expected call of ListLeaderElections.
func (mr *MockLeaderElectionStoreMockRecorder) ListLeaderElections() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListLeaderElections", reflect.TypeOf((*MockLeaderElectionStore)(nil).ListLeaderElections))
}
