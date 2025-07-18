// Code generated by MockGen. DO NOT EDIT.
// Source: projects.go
//
// Generated by this command:
//
//	mockgen -package=mocks -destination=../../../mocks/handler/projects.go -source=projects.go ProjectsService
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

// MockProjectsService is a mock of ProjectsService interface.
type MockProjectsService struct {
	ctrl     *gomock.Controller
	recorder *MockProjectsServiceMockRecorder
	isgomock struct{}
}

// MockProjectsServiceMockRecorder is the mock recorder for MockProjectsService.
type MockProjectsServiceMockRecorder struct {
	mock *MockProjectsService
}

// NewMockProjectsService creates a new mock instance.
func NewMockProjectsService(ctrl *gomock.Controller) *MockProjectsService {
	mock := &MockProjectsService{ctrl: ctrl}
	mock.recorder = &MockProjectsServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProjectsService) EXPECT() *MockProjectsServiceMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockProjectsService) Create(ctx context.Context, input *model.CreateProjectInput) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, input)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create.
func (mr *MockProjectsServiceMockRecorder) Create(ctx, input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockProjectsService)(nil).Create), ctx, input)
}

// DeleteByID mocks base method.
func (m *MockProjectsService) DeleteByID(ctx context.Context, input *model.DeleteProjectInput) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteByID", ctx, input)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteByID indicates an expected call of DeleteByID.
func (mr *MockProjectsServiceMockRecorder) DeleteByID(ctx, input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteByID", reflect.TypeOf((*MockProjectsService)(nil).DeleteByID), ctx, input)
}

// GetByID mocks base method.
func (m *MockProjectsService) GetByID(ctx context.Context, id, userID uuid.UUID) (*model.Project, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByID", ctx, id, userID)
	ret0, _ := ret[0].(*model.Project)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByID indicates an expected call of GetByID.
func (mr *MockProjectsServiceMockRecorder) GetByID(ctx, id, userID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockProjectsService)(nil).GetByID), ctx, id, userID)
}

// List mocks base method.
func (m *MockProjectsService) List(ctx context.Context, input *model.ListProjectsInput) (*model.ListProjectsOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", ctx, input)
	ret0, _ := ret[0].(*model.ListProjectsOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockProjectsServiceMockRecorder) List(ctx, input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockProjectsService)(nil).List), ctx, input)
}

// UpdateByID mocks base method.
func (m *MockProjectsService) UpdateByID(ctx context.Context, input *model.UpdateProjectInput) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateByID", ctx, input)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateByID indicates an expected call of UpdateByID.
func (mr *MockProjectsServiceMockRecorder) UpdateByID(ctx, input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateByID", reflect.TypeOf((*MockProjectsService)(nil).UpdateByID), ctx, input)
}
