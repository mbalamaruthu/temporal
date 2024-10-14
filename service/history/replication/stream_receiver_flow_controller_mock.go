// The MIT License
//
// Copyright (c) 2020 Temporal Technologies Inc.  All rights reserved.
//
// Copyright (c) 2020 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// Code generated by MockGen. DO NOT EDIT.
// Source: stream_receiver_flow_controller.go
//
// Generated by this command:
//
//	mockgen -copyright_file ../../../LICENSE -package replication -source stream_receiver_flow_controller.go -destination stream_receiver_flow_controller_mock.go
//

// Package replication is a generated GoMock package.
package replication

import (
	reflect "reflect"

	enums "go.temporal.io/server/api/enums/v1"
	gomock "go.uber.org/mock/gomock"
)

// MockReceiverFlowController is a mock of ReceiverFlowController interface.
type MockReceiverFlowController struct {
	ctrl     *gomock.Controller
	recorder *MockReceiverFlowControllerMockRecorder
}

// MockReceiverFlowControllerMockRecorder is the mock recorder for MockReceiverFlowController.
type MockReceiverFlowControllerMockRecorder struct {
	mock *MockReceiverFlowController
}

// NewMockReceiverFlowController creates a new mock instance.
func NewMockReceiverFlowController(ctrl *gomock.Controller) *MockReceiverFlowController {
	mock := &MockReceiverFlowController{ctrl: ctrl}
	mock.recorder = &MockReceiverFlowControllerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockReceiverFlowController) EXPECT() *MockReceiverFlowControllerMockRecorder {
	return m.recorder
}

// GetFlowControlInfo mocks base method.
func (m *MockReceiverFlowController) GetFlowControlInfo(priority enums.TaskPriority) enums.ReplicationFlowControlCommand {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFlowControlInfo", priority)
	ret0, _ := ret[0].(enums.ReplicationFlowControlCommand)
	return ret0
}

// GetFlowControlInfo indicates an expected call of GetFlowControlInfo.
func (mr *MockReceiverFlowControllerMockRecorder) GetFlowControlInfo(priority any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFlowControlInfo", reflect.TypeOf((*MockReceiverFlowController)(nil).GetFlowControlInfo), priority)
}
