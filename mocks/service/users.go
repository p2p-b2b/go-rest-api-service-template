// Code generated by MockGen. DO NOT EDIT.
// Source: users.go
//
// Generated by this command:
//
//	mockgen -package=mocks -destination=../../mocks/service/users.go -source=users.go UsersRepository
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

// MockUsersRepository is a mock of UsersRepository interface.
type MockUsersRepository struct {
	ctrl     *gomock.Controller
	recorder *MockUsersRepositoryMockRecorder
	isgomock struct{}
}

// MockUsersRepositoryMockRecorder is the mock recorder for MockUsersRepository.
type MockUsersRepositoryMockRecorder struct {
	mock *MockUsersRepository
}

// NewMockUsersRepository creates a new mock instance.
func NewMockUsersRepository(ctrl *gomock.Controller) *MockUsersRepository {
	mock := &MockUsersRepository{ctrl: ctrl}
	mock.recorder = &MockUsersRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUsersRepository) EXPECT() *MockUsersRepositoryMockRecorder {
	return m.recorder
}

// Delete mocks base method.
func (m *MockUsersRepository) Delete(ctx context.Context, input *model.DeleteUserInput) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, input)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockUsersRepositoryMockRecorder) Delete(ctx, input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockUsersRepository)(nil).Delete), ctx, input)
}

// Insert mocks base method.
func (m *MockUsersRepository) Insert(ctx context.Context, input *model.InsertUserInput) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Insert", ctx, input)
	ret0, _ := ret[0].(error)
	return ret0
}

// Insert indicates an expected call of Insert.
func (mr *MockUsersRepositoryMockRecorder) Insert(ctx, input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Insert", reflect.TypeOf((*MockUsersRepository)(nil).Insert), ctx, input)
}

// Select mocks base method.
func (m *MockUsersRepository) Select(ctx context.Context, input *model.SelectUsersInput) (*model.SelectUsersOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Select", ctx, input)
	ret0, _ := ret[0].(*model.SelectUsersOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Select indicates an expected call of Select.
func (mr *MockUsersRepositoryMockRecorder) Select(ctx, input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Select", reflect.TypeOf((*MockUsersRepository)(nil).Select), ctx, input)
}

// SelectByEmail mocks base method.
func (m *MockUsersRepository) SelectByEmail(ctx context.Context, email string) (*model.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelectByEmail", ctx, email)
	ret0, _ := ret[0].(*model.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SelectByEmail indicates an expected call of SelectByEmail.
func (mr *MockUsersRepositoryMockRecorder) SelectByEmail(ctx, email any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectByEmail", reflect.TypeOf((*MockUsersRepository)(nil).SelectByEmail), ctx, email)
}

// SelectByID mocks base method.
func (m *MockUsersRepository) SelectByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelectByID", ctx, id)
	ret0, _ := ret[0].(*model.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SelectByID indicates an expected call of SelectByID.
func (mr *MockUsersRepositoryMockRecorder) SelectByID(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectByID", reflect.TypeOf((*MockUsersRepository)(nil).SelectByID), ctx, id)
}

// Update mocks base method.
func (m *MockUsersRepository) Update(ctx context.Context, input *model.UpdateUserInput) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", ctx, input)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *MockUsersRepositoryMockRecorder) Update(ctx, input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockUsersRepository)(nil).Update), ctx, input)
}
