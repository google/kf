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
// Source: github.com/GoogleCloudPlatform/kf/pkg/kf/buildpacks/fake (interfaces: Client)

// Package fake is a generated GoMock package.
package fake

import (
	buildpacks "github.com/GoogleCloudPlatform/kf/pkg/kf/buildpacks"
	doctor "github.com/GoogleCloudPlatform/kf/pkg/kf/doctor"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
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

// List mocks base method
func (m *FakeClient) List() ([]buildpacks.Buildpack, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List")
	ret0, _ := ret[0].([]buildpacks.Buildpack)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List
func (mr *FakeClientMockRecorder) List() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*FakeClient)(nil).List))
}
