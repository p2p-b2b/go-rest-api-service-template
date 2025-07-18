// Code generated by MockGen. DO NOT EDIT.
// Source: authn.go
//
// Generated by this command:
//
//	mockgen -package=mocks -destination=../../../mocks/handler/authn.go -source=authn.go AuthnService
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	model "github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	gomock "go.uber.org/mock/gomock"
)

// MockAuthnService is a mock of AuthnService interface.
type MockAuthnService struct {
	ctrl     *gomock.Controller
	recorder *MockAuthnServiceMockRecorder
	isgomock struct{}
}

// MockAuthnServiceMockRecorder is the mock recorder for MockAuthnService.
type MockAuthnServiceMockRecorder struct {
	mock *MockAuthnService
}

// NewMockAuthnService creates a new mock instance.
func NewMockAuthnService(ctrl *gomock.Controller) *MockAuthnService {
	mock := &MockAuthnService{ctrl: ctrl}
	mock.recorder = &MockAuthnServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAuthnService) EXPECT() *MockAuthnServiceMockRecorder {
	return m.recorder
}

// LoggingOutUser mocks base method.
func (m *MockAuthnService) LoggingOutUser(ctx context.Context, userID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LoggingOutUser", ctx, userID)
	ret0, _ := ret[0].(error)
	return ret0
}

// LoggingOutUser indicates an expected call of LoggingOutUser.
func (mr *MockAuthnServiceMockRecorder) LoggingOutUser(ctx, userID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LoggingOutUser", reflect.TypeOf((*MockAuthnService)(nil).LoggingOutUser), ctx, userID)
}

// LoginUser mocks base method.
func (m *MockAuthnService) LoginUser(ctx context.Context, input *model.LoginUserInput) (*model.LoginUserOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LoginUser", ctx, input)
	ret0, _ := ret[0].(*model.LoginUserOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LoginUser indicates an expected call of LoginUser.
func (mr *MockAuthnServiceMockRecorder) LoginUser(ctx, input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LoginUser", reflect.TypeOf((*MockAuthnService)(nil).LoginUser), ctx, input)
}

// ReVerifyUser mocks base method.
func (m *MockAuthnService) ReVerifyUser(ctx context.Context, email string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReVerifyUser", ctx, email)
	ret0, _ := ret[0].(error)
	return ret0
}

// ReVerifyUser indicates an expected call of ReVerifyUser.
func (mr *MockAuthnServiceMockRecorder) ReVerifyUser(ctx, email any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReVerifyUser", reflect.TypeOf((*MockAuthnService)(nil).ReVerifyUser), ctx, email)
}

// RefreshAccessToken mocks base method.
func (m *MockAuthnService) RefreshAccessToken(ctx context.Context, input *model.RefreshAccessTokenInput) (*model.RefreshAccessTokenOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RefreshAccessToken", ctx, input)
	ret0, _ := ret[0].(*model.RefreshAccessTokenOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RefreshAccessToken indicates an expected call of RefreshAccessToken.
func (mr *MockAuthnServiceMockRecorder) RefreshAccessToken(ctx, input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RefreshAccessToken", reflect.TypeOf((*MockAuthnService)(nil).RefreshAccessToken), ctx, input)
}

// RegisterUser mocks base method.
func (m *MockAuthnService) RegisterUser(ctx context.Context, input *model.RegisterUserInput) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RegisterUser", ctx, input)
	ret0, _ := ret[0].(error)
	return ret0
}

// RegisterUser indicates an expected call of RegisterUser.
func (mr *MockAuthnServiceMockRecorder) RegisterUser(ctx, input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterUser", reflect.TypeOf((*MockAuthnService)(nil).RegisterUser), ctx, input)
}

// VerifyUser mocks base method.
func (m *MockAuthnService) VerifyUser(ctx context.Context, jwtToken string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "VerifyUser", ctx, jwtToken)
	ret0, _ := ret[0].(error)
	return ret0
}

// VerifyUser indicates an expected call of VerifyUser.
func (mr *MockAuthnServiceMockRecorder) VerifyUser(ctx, jwtToken any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "VerifyUser", reflect.TypeOf((*MockAuthnService)(nil).VerifyUser), ctx, jwtToken)
}
