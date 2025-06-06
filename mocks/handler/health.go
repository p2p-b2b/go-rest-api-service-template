// Code generated by MockGen. DO NOT EDIT.
// Source: health.go
//
// Generated by this command:
//
//	mockgen -package=mocks -destination=../../../mocks/handler/health.go -source=health.go HealthService
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	model "github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	gomock "go.uber.org/mock/gomock"
)

// MockHealthService is a mock of HealthService interface.
type MockHealthService struct {
	ctrl     *gomock.Controller
	recorder *MockHealthServiceMockRecorder
	isgomock struct{}
}

// MockHealthServiceMockRecorder is the mock recorder for MockHealthService.
type MockHealthServiceMockRecorder struct {
	mock *MockHealthService
}

// NewMockHealthService creates a new mock instance.
func NewMockHealthService(ctrl *gomock.Controller) *MockHealthService {
	mock := &MockHealthService{ctrl: ctrl}
	mock.recorder = &MockHealthServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockHealthService) EXPECT() *MockHealthServiceMockRecorder {
	return m.recorder
}

// HealthCheck mocks base method.
func (m *MockHealthService) HealthCheck(ctx context.Context) (model.Health, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HealthCheck", ctx)
	ret0, _ := ret[0].(model.Health)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// HealthCheck indicates an expected call of HealthCheck.
func (mr *MockHealthServiceMockRecorder) HealthCheck(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HealthCheck", reflect.TypeOf((*MockHealthService)(nil).HealthCheck), ctx)
}
