// Code generated by MockGen. DO NOT EDIT.
// Source: users.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	uuid "github.com/google/uuid"
	service "github.com/p2p-b2b/go-rest-api-service-template/internal/service"
)

// MockUserService is a mock of UserService interface.
type MockUserService struct {
	ctrl     *gomock.Controller
	recorder *MockUserServiceMockRecorder
}

// MockUserServiceMockRecorder is the mock recorder for MockUserService.
type MockUserServiceMockRecorder struct {
	mock *MockUserService
}

// NewMockUserService creates a new mock instance.
func NewMockUserService(ctrl *gomock.Controller) *MockUserService {
	mock := &MockUserService{ctrl: ctrl}
	mock.recorder = &MockUserServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserService) EXPECT() *MockUserServiceMockRecorder {
	return m.recorder
}

// CreateUser mocks base method.
func (m *MockUserService) CreateUser(ctx context.Context, user *service.CreateUserInput) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateUser", ctx, user)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateUser indicates an expected call of CreateUser.
func (mr *MockUserServiceMockRecorder) CreateUser(ctx, user interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUser", reflect.TypeOf((*MockUserService)(nil).CreateUser), ctx, user)
}

// DeleteUser mocks base method.
func (m *MockUserService) DeleteUser(ctx context.Context, user *service.DeleteUserInput) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteUser", ctx, user)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteUser indicates an expected call of DeleteUser.
func (mr *MockUserServiceMockRecorder) DeleteUser(ctx, user interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteUser", reflect.TypeOf((*MockUserService)(nil).DeleteUser), ctx, user)
}

// GetUserByID mocks base method.
func (m *MockUserService) GetUserByID(ctx context.Context, id uuid.UUID) (*service.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserByID", ctx, id)
	ret0, _ := ret[0].(*service.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserByID indicates an expected call of GetUserByID.
func (mr *MockUserServiceMockRecorder) GetUserByID(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserByID", reflect.TypeOf((*MockUserService)(nil).GetUserByID), ctx, id)
}

// ListUsers mocks base method.
func (m *MockUserService) ListUsers(ctx context.Context, params *service.ListUserInput) (*service.ListUserOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListUsers", ctx, params)
	ret0, _ := ret[0].(*service.ListUserOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListUsers indicates an expected call of ListUsers.
func (mr *MockUserServiceMockRecorder) ListUsers(ctx, params interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListUsers", reflect.TypeOf((*MockUserService)(nil).ListUsers), ctx, params)
}

// UpdateUser mocks base method.
func (m *MockUserService) UpdateUser(ctx context.Context, user *service.UpdateUserInput) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateUser", ctx, user)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateUser indicates an expected call of UpdateUser.
func (mr *MockUserServiceMockRecorder) UpdateUser(ctx, user interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateUser", reflect.TypeOf((*MockUserService)(nil).UpdateUser), ctx, user)
}

// UserHealthCheck mocks base method.
func (m *MockUserService) UserHealthCheck(ctx context.Context) (service.Health, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UserHealthCheck", ctx)
	ret0, _ := ret[0].(service.Health)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UserHealthCheck indicates an expected call of UserHealthCheck.
func (mr *MockUserServiceMockRecorder) UserHealthCheck(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UserHealthCheck", reflect.TypeOf((*MockUserService)(nil).UserHealthCheck), ctx)
}
