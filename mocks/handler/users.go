// Code generated by MockGen. DO NOT EDIT.
// Source: users.go
//
// Generated by this command:
//
//	mockgen -package=mocks -destination=../../../mocks/handler/users.go -source=users.go UserService
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

// MockUserService is a mock of UserService interface.
type MockUserService struct {
	ctrl     *gomock.Controller
	recorder *MockUserServiceMockRecorder
	isgomock struct{}
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

// Create mocks base method.
func (m *MockUserService) Create(ctx context.Context, input *service.CreateUserInput) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, input)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create.
func (mr *MockUserServiceMockRecorder) Create(ctx, input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockUserService)(nil).Create), ctx, input)
}

// Delete mocks base method.
func (m *MockUserService) Delete(ctx context.Context, input *service.DeleteUserInput) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, input)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockUserServiceMockRecorder) Delete(ctx, input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockUserService)(nil).Delete), ctx, input)
}

// GetByEmail mocks base method.
func (m *MockUserService) GetByEmail(ctx context.Context, email string) (*service.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByEmail", ctx, email)
	ret0, _ := ret[0].(*service.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByEmail indicates an expected call of GetByEmail.
func (mr *MockUserServiceMockRecorder) GetByEmail(ctx, email any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByEmail", reflect.TypeOf((*MockUserService)(nil).GetByEmail), ctx, email)
}

// GetByID mocks base method.
func (m *MockUserService) GetByID(ctx context.Context, id uuid.UUID) (*service.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByID", ctx, id)
	ret0, _ := ret[0].(*service.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByID indicates an expected call of GetByID.
func (mr *MockUserServiceMockRecorder) GetByID(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockUserService)(nil).GetByID), ctx, id)
}

// HealthCheck mocks base method.
func (m *MockUserService) HealthCheck(ctx context.Context) (service.Health, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HealthCheck", ctx)
	ret0, _ := ret[0].(service.Health)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// HealthCheck indicates an expected call of HealthCheck.
func (mr *MockUserServiceMockRecorder) HealthCheck(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HealthCheck", reflect.TypeOf((*MockUserService)(nil).HealthCheck), ctx)
}

// List mocks base method.
func (m *MockUserService) List(ctx context.Context, input *service.ListUserInput) (*service.ListUsersOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", ctx, input)
	ret0, _ := ret[0].(*service.ListUsersOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockUserServiceMockRecorder) List(ctx, input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockUserService)(nil).List), ctx, input)
}

// Update mocks base method.
func (m *MockUserService) Update(ctx context.Context, input *service.UpdateUserInput) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", ctx, input)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *MockUserServiceMockRecorder) Update(ctx, input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockUserService)(nil).Update), ctx, input)
}
