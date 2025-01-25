// Code generated by MockGen. DO NOT EDIT.
// Source: users.go
//
// Generated by this command:
//
//	mockgen -package=mocks -destination=../../../mocks/handler/users.go -source=users.go UsersService
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	uuid "github.com/google/uuid"
	service "github.com/p2p-b2b/go-rest-api-service-template/internal/service"
	gomock "go.uber.org/mock/gomock"
)

// MockUsersService is a mock of UsersService interface.
type MockUsersService struct {
	ctrl     *gomock.Controller
	recorder *MockUsersServiceMockRecorder
	isgomock struct{}
}

// MockUsersServiceMockRecorder is the mock recorder for MockUsersService.
type MockUsersServiceMockRecorder struct {
	mock *MockUsersService
}

// NewMockUsersService creates a new mock instance.
func NewMockUsersService(ctrl *gomock.Controller) *MockUsersService {
	mock := &MockUsersService{ctrl: ctrl}
	mock.recorder = &MockUsersServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUsersService) EXPECT() *MockUsersServiceMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockUsersService) Create(ctx context.Context, input *service.CreateUserInput) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, input)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create.
func (mr *MockUsersServiceMockRecorder) Create(ctx, input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockUsersService)(nil).Create), ctx, input)
}

// Delete mocks base method.
func (m *MockUsersService) Delete(ctx context.Context, input *service.DeleteUserInput) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, input)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockUsersServiceMockRecorder) Delete(ctx, input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockUsersService)(nil).Delete), ctx, input)
}

// GetByEmail mocks base method.
func (m *MockUsersService) GetByEmail(ctx context.Context, email string) (*service.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByEmail", ctx, email)
	ret0, _ := ret[0].(*service.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByEmail indicates an expected call of GetByEmail.
func (mr *MockUsersServiceMockRecorder) GetByEmail(ctx, email any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByEmail", reflect.TypeOf((*MockUsersService)(nil).GetByEmail), ctx, email)
}

// GetByID mocks base method.
func (m *MockUsersService) GetByID(ctx context.Context, id uuid.UUID) (*service.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByID", ctx, id)
	ret0, _ := ret[0].(*service.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByID indicates an expected call of GetByID.
func (mr *MockUsersServiceMockRecorder) GetByID(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockUsersService)(nil).GetByID), ctx, id)
}

// HealthCheck mocks base method.
func (m *MockUsersService) HealthCheck(ctx context.Context) (service.Health, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HealthCheck", ctx)
	ret0, _ := ret[0].(service.Health)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// HealthCheck indicates an expected call of HealthCheck.
func (mr *MockUsersServiceMockRecorder) HealthCheck(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HealthCheck", reflect.TypeOf((*MockUsersService)(nil).HealthCheck), ctx)
}

// List mocks base method.
func (m *MockUsersService) List(ctx context.Context, input *service.ListUsersInput) (*service.ListUsersOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", ctx, input)
	ret0, _ := ret[0].(*service.ListUsersOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockUsersServiceMockRecorder) List(ctx, input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockUsersService)(nil).List), ctx, input)
}

// Update mocks base method.
func (m *MockUsersService) Update(ctx context.Context, input *service.UpdateUserInput) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", ctx, input)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *MockUsersServiceMockRecorder) Update(ctx, input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockUsersService)(nil).Update), ctx, input)
}
