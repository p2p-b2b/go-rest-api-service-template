// Code generated by MockGen. DO NOT EDIT.
// Source: users.go
//
// Generated by this command:
//
//	mockgen -package=mocks -destination=../../mocks/service/users.go -source=users.go UserRepository
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	sql "database/sql"
	reflect "reflect"

	uuid "github.com/google/uuid"
	repository "github.com/p2p-b2b/go-rest-api-service-template/internal/repository"
	gomock "go.uber.org/mock/gomock"
)

// MockUserRepository is a mock of UserRepository interface.
type MockUserRepository struct {
	ctrl     *gomock.Controller
	recorder *MockUserRepositoryMockRecorder
	isgomock struct{}
}

// MockUserRepositoryMockRecorder is the mock recorder for MockUserRepository.
type MockUserRepositoryMockRecorder struct {
	mock *MockUserRepository
}

// NewMockUserRepository creates a new mock instance.
func NewMockUserRepository(ctrl *gomock.Controller) *MockUserRepository {
	mock := &MockUserRepository{ctrl: ctrl}
	mock.recorder = &MockUserRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserRepository) EXPECT() *MockUserRepositoryMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockUserRepository) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockUserRepositoryMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockUserRepository)(nil).Close))
}

// Conn mocks base method.
func (m *MockUserRepository) Conn(ctx context.Context) (*sql.Conn, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Conn", ctx)
	ret0, _ := ret[0].(*sql.Conn)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Conn indicates an expected call of Conn.
func (mr *MockUserRepositoryMockRecorder) Conn(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Conn", reflect.TypeOf((*MockUserRepository)(nil).Conn), ctx)
}

// Delete mocks base method.
func (m *MockUserRepository) Delete(ctx context.Context, input *repository.DeleteUserInput) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, input)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockUserRepositoryMockRecorder) Delete(ctx, input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockUserRepository)(nil).Delete), ctx, input)
}

// DriverName mocks base method.
func (m *MockUserRepository) DriverName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DriverName")
	ret0, _ := ret[0].(string)
	return ret0
}

// DriverName indicates an expected call of DriverName.
func (mr *MockUserRepositoryMockRecorder) DriverName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DriverName", reflect.TypeOf((*MockUserRepository)(nil).DriverName))
}

// Insert mocks base method.
func (m *MockUserRepository) Insert(ctx context.Context, input *repository.InsertUserInput) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Insert", ctx, input)
	ret0, _ := ret[0].(error)
	return ret0
}

// Insert indicates an expected call of Insert.
func (mr *MockUserRepositoryMockRecorder) Insert(ctx, input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Insert", reflect.TypeOf((*MockUserRepository)(nil).Insert), ctx, input)
}

// PingContext mocks base method.
func (m *MockUserRepository) PingContext(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PingContext", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// PingContext indicates an expected call of PingContext.
func (mr *MockUserRepositoryMockRecorder) PingContext(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PingContext", reflect.TypeOf((*MockUserRepository)(nil).PingContext), ctx)
}

// Select mocks base method.
func (m *MockUserRepository) Select(ctx context.Context, input *repository.SelectUsersInput) (*repository.SelectUsersOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Select", ctx, input)
	ret0, _ := ret[0].(*repository.SelectUsersOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Select indicates an expected call of Select.
func (mr *MockUserRepositoryMockRecorder) Select(ctx, input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Select", reflect.TypeOf((*MockUserRepository)(nil).Select), ctx, input)
}

// SelectUserByEmail mocks base method.
func (m *MockUserRepository) SelectUserByEmail(ctx context.Context, email string) (*repository.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelectUserByEmail", ctx, email)
	ret0, _ := ret[0].(*repository.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SelectUserByEmail indicates an expected call of SelectUserByEmail.
func (mr *MockUserRepositoryMockRecorder) SelectUserByEmail(ctx, email any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectUserByEmail", reflect.TypeOf((*MockUserRepository)(nil).SelectUserByEmail), ctx, email)
}

// SelectUserByID mocks base method.
func (m *MockUserRepository) SelectUserByID(ctx context.Context, id uuid.UUID) (*repository.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelectUserByID", ctx, id)
	ret0, _ := ret[0].(*repository.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SelectUserByID indicates an expected call of SelectUserByID.
func (mr *MockUserRepositoryMockRecorder) SelectUserByID(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectUserByID", reflect.TypeOf((*MockUserRepository)(nil).SelectUserByID), ctx, id)
}

// Update mocks base method.
func (m *MockUserRepository) Update(ctx context.Context, input *repository.UpdateUserInput) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", ctx, input)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *MockUserRepositoryMockRecorder) Update(ctx, input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockUserRepository)(nil).Update), ctx, input)
}
