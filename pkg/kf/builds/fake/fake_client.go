// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/GoogleCloudPlatform/kf/pkg/kf/builds/fake (interfaces: Client)

// Package fake is a generated GoMock package.
package fake

import (
	reflect "reflect"

	builds "github.com/GoogleCloudPlatform/kf/pkg/kf/builds"
	doctor "github.com/GoogleCloudPlatform/kf/pkg/kf/doctor"
	gomock "github.com/golang/mock/gomock"
	v1alpha1 "github.com/knative/build/pkg/apis/build/v1alpha1"
)

// FakeClient is a mock of Client interface
type FakeClient struct {
	ctrl     *gomock.Controller
	recorder *FakeClientMockRecorder
}

// FakeClientMockRecorder is the mock recorder for FakeClient
type FakeClientMockRecorder struct {
	mock *FakeClient
}

// NewFakeClient creates a new mock instance
func NewFakeClient(ctrl *gomock.Controller) *FakeClient {
	mock := &FakeClient{ctrl: ctrl}
	mock.recorder = &FakeClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *FakeClient) EXPECT() *FakeClientMockRecorder {
	return m.recorder
}

// Create mocks base method
func (m *FakeClient) Create(arg0 string, arg1 v1alpha1.TemplateInstantiationSpec, arg2 ...builds.CreateOption) (*v1alpha1.Build, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Create", varargs...)
	ret0, _ := ret[0].(*v1alpha1.Build)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create
func (mr *FakeClientMockRecorder) Create(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*FakeClient)(nil).Create), varargs...)
}

// Delete mocks base method
func (m *FakeClient) Delete(arg0 string, arg1 ...builds.DeleteOption) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Delete", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete
func (mr *FakeClientMockRecorder) Delete(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*FakeClient)(nil).Delete), varargs...)
}

// Diagnose mocks base method
func (m *FakeClient) Diagnose(arg0 *doctor.Diagnostic) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Diagnose", arg0)
}

// Diagnose indicates an expected call of Diagnose
func (mr *FakeClientMockRecorder) Diagnose(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Diagnose", reflect.TypeOf((*FakeClient)(nil).Diagnose), arg0)
}

// Status mocks base method
func (m *FakeClient) Status(arg0 string, arg1 ...builds.StatusOption) (bool, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Status", varargs...)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Status indicates an expected call of Status
func (mr *FakeClientMockRecorder) Status(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Status", reflect.TypeOf((*FakeClient)(nil).Status), varargs...)
}
