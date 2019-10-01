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
// Source: github.com/google/kf/pkg/kf/services/fake (interfaces: Client)

// Package fake is a generated GoMock package.
package fake

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	services "github.com/google/kf/pkg/kf/services"
	v1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	reflect "reflect"
	time "time"
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
func (m *FakeClient) Create(arg0 string, arg1 *v1beta1.ServiceInstance, arg2 ...services.CreateOption) (*v1beta1.ServiceInstance, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Create", varargs...)
	ret0, _ := ret[0].(*v1beta1.ServiceInstance)
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
func (m *FakeClient) Delete(arg0, arg1 string, arg2 ...services.DeleteOption) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Delete", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete
func (mr *FakeClientMockRecorder) Delete(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*FakeClient)(nil).Delete), varargs...)
}

// Get mocks base method
func (m *FakeClient) Get(arg0, arg1 string, arg2 ...services.GetOption) (*v1beta1.ServiceInstance, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Get", varargs...)
	ret0, _ := ret[0].(*v1beta1.ServiceInstance)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get
func (mr *FakeClientMockRecorder) Get(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*FakeClient)(nil).Get), varargs...)
}

// List mocks base method
func (m *FakeClient) List(arg0 string, arg1 ...services.ListOption) ([]v1beta1.ServiceInstance, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "List", varargs...)
	ret0, _ := ret[0].([]v1beta1.ServiceInstance)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List
func (mr *FakeClientMockRecorder) List(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*FakeClient)(nil).List), varargs...)
}

// Transform mocks base method
func (m *FakeClient) Transform(arg0, arg1 string, arg2 services.Mutator) (*v1beta1.ServiceInstance, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Transform", arg0, arg1, arg2)
	ret0, _ := ret[0].(*v1beta1.ServiceInstance)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Transform indicates an expected call of Transform
func (mr *FakeClientMockRecorder) Transform(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Transform", reflect.TypeOf((*FakeClient)(nil).Transform), arg0, arg1, arg2)
}

// Update mocks base method
func (m *FakeClient) Update(arg0 string, arg1 *v1beta1.ServiceInstance, arg2 ...services.UpdateOption) (*v1beta1.ServiceInstance, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Update", varargs...)
	ret0, _ := ret[0].(*v1beta1.ServiceInstance)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Update indicates an expected call of Update
func (mr *FakeClientMockRecorder) Update(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*FakeClient)(nil).Update), varargs...)
}

// Upsert mocks base method
func (m *FakeClient) Upsert(arg0 string, arg1 *v1beta1.ServiceInstance, arg2 services.Merger) (*v1beta1.ServiceInstance, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Upsert", arg0, arg1, arg2)
	ret0, _ := ret[0].(*v1beta1.ServiceInstance)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Upsert indicates an expected call of Upsert
func (mr *FakeClientMockRecorder) Upsert(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Upsert", reflect.TypeOf((*FakeClient)(nil).Upsert), arg0, arg1, arg2)
}

// WaitFor mocks base method
func (m *FakeClient) WaitFor(arg0 context.Context, arg1, arg2 string, arg3 time.Duration, arg4 services.Predicate) (*v1beta1.ServiceInstance, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WaitFor", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(*v1beta1.ServiceInstance)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WaitFor indicates an expected call of WaitFor
func (mr *FakeClientMockRecorder) WaitFor(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WaitFor", reflect.TypeOf((*FakeClient)(nil).WaitFor), arg0, arg1, arg2, arg3, arg4)
}

// WaitForDeletion mocks base method
func (m *FakeClient) WaitForDeletion(arg0 context.Context, arg1, arg2 string, arg3 time.Duration) (*v1beta1.ServiceInstance, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WaitForDeletion", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(*v1beta1.ServiceInstance)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WaitForDeletion indicates an expected call of WaitForDeletion
func (mr *FakeClientMockRecorder) WaitForDeletion(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WaitForDeletion", reflect.TypeOf((*FakeClient)(nil).WaitForDeletion), arg0, arg1, arg2, arg3)
}

// WaitForE mocks base method
func (m *FakeClient) WaitForE(arg0 context.Context, arg1, arg2 string, arg3 time.Duration, arg4 services.ConditionFuncE) (*v1beta1.ServiceInstance, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WaitForE", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(*v1beta1.ServiceInstance)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WaitForE indicates an expected call of WaitForE
func (mr *FakeClientMockRecorder) WaitForE(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WaitForE", reflect.TypeOf((*FakeClient)(nil).WaitForE), arg0, arg1, arg2, arg3, arg4)
}

// WaitForProvisionSuccess mocks base method
func (m *FakeClient) WaitForProvisionSuccess(arg0 context.Context, arg1, arg2 string, arg3 time.Duration) (*v1beta1.ServiceInstance, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WaitForProvisionSuccess", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(*v1beta1.ServiceInstance)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WaitForProvisionSuccess indicates an expected call of WaitForProvisionSuccess
func (mr *FakeClientMockRecorder) WaitForProvisionSuccess(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WaitForProvisionSuccess", reflect.TypeOf((*FakeClient)(nil).WaitForProvisionSuccess), arg0, arg1, arg2, arg3)
}
