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
	model "github.com/p2p-b2b/go-rest-api-service-template/internal/model"
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
func (m *MockUsersService) Create(ctx context.Context, input *model.CreateUserInput) error {
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

// DeleteByID mocks base method.
func (m *MockUsersService) DeleteByID(ctx context.Context, input *model.DeleteUserInput) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteByID", ctx, input)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteByID indicates an expected call of DeleteByID.
func (mr *MockUsersServiceMockRecorder) DeleteByID(ctx, input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteByID", reflect.TypeOf((*MockUsersService)(nil).DeleteByID), ctx, input)
}

// GetByID mocks base method.
func (m *MockUsersService) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByID", ctx, id)
	ret0, _ := ret[0].(*model.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByID indicates an expected call of GetByID.
func (mr *MockUsersServiceMockRecorder) GetByID(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockUsersService)(nil).GetByID), ctx, id)
}

// LinkRoles mocks base method.
func (m *MockUsersService) LinkRoles(ctx context.Context, input *model.LinkRolesToUserInput) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LinkRoles", ctx, input)
	ret0, _ := ret[0].(error)
	return ret0
}

// LinkRoles indicates an expected call of LinkRoles.
func (mr *MockUsersServiceMockRecorder) LinkRoles(ctx, input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LinkRoles", reflect.TypeOf((*MockUsersService)(nil).LinkRoles), ctx, input)
}

// List mocks base method.
func (m *MockUsersService) List(ctx context.Context, input *model.ListUsersInput) (*model.ListUsersOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", ctx, input)
	ret0, _ := ret[0].(*model.ListUsersOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockUsersServiceMockRecorder) List(ctx, input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockUsersService)(nil).List), ctx, input)
}

// ListByRoleID mocks base method.
func (m *MockUsersService) ListByRoleID(ctx context.Context, roleID uuid.UUID, input *model.ListUsersInput) (*model.ListUsersOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListByRoleID", ctx, roleID, input)
	ret0, _ := ret[0].(*model.ListUsersOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListByRoleID indicates an expected call of ListByRoleID.
func (mr *MockUsersServiceMockRecorder) ListByRoleID(ctx, roleID, input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListByRoleID", reflect.TypeOf((*MockUsersService)(nil).ListByRoleID), ctx, roleID, input)
}

// SelectAuthz mocks base method.
func (m *MockUsersService) SelectAuthz(ctx context.Context, userID uuid.UUID) (map[string]any, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelectAuthz", ctx, userID)
	ret0, _ := ret[0].(map[string]any)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SelectAuthz indicates an expected call of SelectAuthz.
func (mr *MockUsersServiceMockRecorder) SelectAuthz(ctx, userID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectAuthz", reflect.TypeOf((*MockUsersService)(nil).SelectAuthz), ctx, userID)
}

// UnLinkRoles mocks base method.
func (m *MockUsersService) UnLinkRoles(ctx context.Context, input *model.UnLinkRolesFromUsersInput) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UnLinkRoles", ctx, input)
	ret0, _ := ret[0].(error)
	return ret0
}

// UnLinkRoles indicates an expected call of UnLinkRoles.
func (mr *MockUsersServiceMockRecorder) UnLinkRoles(ctx, input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UnLinkRoles", reflect.TypeOf((*MockUsersService)(nil).UnLinkRoles), ctx, input)
}

// UpdateByID mocks base method.
func (m *MockUsersService) UpdateByID(ctx context.Context, input *model.UpdateUserInput) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateByID", ctx, input)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateByID indicates an expected call of UpdateByID.
func (mr *MockUsersServiceMockRecorder) UpdateByID(ctx, input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateByID", reflect.TypeOf((*MockUsersService)(nil).UpdateByID), ctx, input)
}
