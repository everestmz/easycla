// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT
//

// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/communitybridge/easycla/cla-backend-go/github_organizations (interfaces: Repository)

// Package github_organizations is a generated GoMock package.
package github_organizations

import (
	context "context"
	reflect "reflect"

	models "github.com/communitybridge/easycla/cla-backend-go/gen/v1/models"
	gomock "github.com/golang/mock/gomock"
)

// MockRepository is a mock of Repository interface
type MockRepository struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryMockRecorder
}

// MockRepositoryMockRecorder is the mock recorder for MockRepository
type MockRepositoryMockRecorder struct {
	mock *MockRepository
}

// NewMockRepository creates a new mock instance
func NewMockRepository(ctrl *gomock.Controller) *MockRepository {
	mock := &MockRepository{ctrl: ctrl}
	mock.recorder = &MockRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockRepository) EXPECT() *MockRepositoryMockRecorder {
	return m.recorder
}

// AddGithubOrganization mocks base method
func (m *MockRepository) AddGithubOrganization(arg0 context.Context, arg1, arg2 string, arg3 *models.CreateGithubOrganization) (*models.GithubOrganization, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddGithubOrganization", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(*models.GithubOrganization)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddGithubOrganization indicates an expected call of AddGithubOrganization
func (mr *MockRepositoryMockRecorder) AddGithubOrganization(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddGithubOrganization", reflect.TypeOf((*MockRepository)(nil).AddGithubOrganization), arg0, arg1, arg2, arg3)
}

// DeleteGithubOrganization mocks base method
func (m *MockRepository) DeleteGithubOrganization(arg0 context.Context, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteGithubOrganization", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteGithubOrganization indicates an expected call of DeleteGithubOrganization
func (mr *MockRepositoryMockRecorder) DeleteGithubOrganization(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteGithubOrganization", reflect.TypeOf((*MockRepository)(nil).DeleteGithubOrganization), arg0, arg1, arg2)
}

// DeleteGithubOrganizationByParent mocks base method
func (m *MockRepository) DeleteGithubOrganizationByParent(arg0 context.Context, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteGithubOrganizationByParent", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteGithubOrganizationByParent indicates an expected call of DeleteGithubOrganizationByParent
func (mr *MockRepositoryMockRecorder) DeleteGithubOrganizationByParent(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteGithubOrganizationByParent", reflect.TypeOf((*MockRepository)(nil).DeleteGithubOrganizationByParent), arg0, arg1, arg2)
}

// GetGithubOrganization mocks base method
func (m *MockRepository) GetGithubOrganization(arg0 context.Context, arg1 string) (*models.GithubOrganization, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGithubOrganization", arg0, arg1)
	ret0, _ := ret[0].(*models.GithubOrganization)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetGithubOrganization indicates an expected call of GetGithubOrganization
func (mr *MockRepositoryMockRecorder) GetGithubOrganization(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGithubOrganization", reflect.TypeOf((*MockRepository)(nil).GetGithubOrganization), arg0, arg1)
}

// GetGithubOrganizationByName mocks base method
func (m *MockRepository) GetGithubOrganizationByName(arg0 context.Context, arg1 string) (*models.GithubOrganizations, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGithubOrganizationByName", arg0, arg1)
	ret0, _ := ret[0].(*models.GithubOrganizations)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetGithubOrganizationByName indicates an expected call of GetGithubOrganizationByName
func (mr *MockRepositoryMockRecorder) GetGithubOrganizationByName(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGithubOrganizationByName", reflect.TypeOf((*MockRepository)(nil).GetGithubOrganizationByName), arg0, arg1)
}

// GetGithubOrganizations mocks base method
func (m *MockRepository) GetGithubOrganizations(arg0 context.Context, arg1 string) (*models.GithubOrganizations, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGithubOrganizations", arg0, arg1)
	ret0, _ := ret[0].(*models.GithubOrganizations)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetGithubOrganizations indicates an expected call of GetGithubOrganizations
func (mr *MockRepositoryMockRecorder) GetGithubOrganizations(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGithubOrganizations", reflect.TypeOf((*MockRepository)(nil).GetGithubOrganizations), arg0, arg1)
}

// GetGithubOrganizationsByParent mocks base method
func (m *MockRepository) GetGithubOrganizationsByParent(arg0 context.Context, arg1 string) (*models.GithubOrganizations, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGithubOrganizationsByParent", arg0, arg1)
	ret0, _ := ret[0].(*models.GithubOrganizations)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetGithubOrganizationsByParent indicates an expected call of GetGithubOrganizationsByParent
func (mr *MockRepositoryMockRecorder) GetGithubOrganizationsByParent(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGithubOrganizationsByParent", reflect.TypeOf((*MockRepository)(nil).GetGithubOrganizationsByParent), arg0, arg1)
}

// UpdateGithubOrganization mocks base method
func (m *MockRepository) UpdateGithubOrganization(arg0 context.Context, arg1, arg2 string, arg3 bool, arg4 string, arg5 bool, arg6 *bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateGithubOrganization", arg0, arg1, arg2, arg3, arg4, arg5, arg6)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateGithubOrganization indicates an expected call of UpdateGithubOrganization
func (mr *MockRepositoryMockRecorder) UpdateGithubOrganization(arg0, arg1, arg2, arg3, arg4, arg5, arg6 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateGithubOrganization", reflect.TypeOf((*MockRepository)(nil).UpdateGithubOrganization), arg0, arg1, arg2, arg3, arg4, arg5, arg6)
}
