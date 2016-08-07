// Copyright (c) 2016 Uber Technologies, Inc.
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

// Automatically generated by MockGen. DO NOT EDIT!
// Source: github.com/m3db/m3db/interfaces/m3db (interfaces: FileSetWriter,FileSetReader)

package mocks

import (
	gomock "github.com/golang/mock/gomock"
	time0 "github.com/m3db/m3x/time"
	time "time"
)

// Mock of FileSetWriter interface
type MockFileSetWriter struct {
	ctrl     *gomock.Controller
	recorder *_MockFileSetWriterRecorder
}

// Recorder for MockFileSetWriter (not exported)
type _MockFileSetWriterRecorder struct {
	mock *MockFileSetWriter
}

func NewMockFileSetWriter(ctrl *gomock.Controller) *MockFileSetWriter {
	mock := &MockFileSetWriter{ctrl: ctrl}
	mock.recorder = &_MockFileSetWriterRecorder{mock}
	return mock
}

func (_m *MockFileSetWriter) EXPECT() *_MockFileSetWriterRecorder {
	return _m.recorder
}

func (_m *MockFileSetWriter) Close() error {
	ret := _m.ctrl.Call(_m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockFileSetWriterRecorder) Close() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Close")
}

func (_m *MockFileSetWriter) Open(_param0 uint32, _param1 time.Time) error {
	ret := _m.ctrl.Call(_m, "Open", _param0, _param1)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockFileSetWriterRecorder) Open(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Open", arg0, arg1)
}

func (_m *MockFileSetWriter) Write(_param0 string, _param1 []byte) error {
	ret := _m.ctrl.Call(_m, "Write", _param0, _param1)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockFileSetWriterRecorder) Write(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Write", arg0, arg1)
}

func (_m *MockFileSetWriter) WriteAll(_param0 string, _param1 [][]byte) error {
	ret := _m.ctrl.Call(_m, "WriteAll", _param0, _param1)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockFileSetWriterRecorder) WriteAll(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "WriteAll", arg0, arg1)
}

// Mock of FileSetReader interface
type MockFileSetReader struct {
	ctrl     *gomock.Controller
	recorder *_MockFileSetReaderRecorder
}

// Recorder for MockFileSetReader (not exported)
type _MockFileSetReaderRecorder struct {
	mock *MockFileSetReader
}

func NewMockFileSetReader(ctrl *gomock.Controller) *MockFileSetReader {
	mock := &MockFileSetReader{ctrl: ctrl}
	mock.recorder = &_MockFileSetReaderRecorder{mock}
	return mock
}

func (_m *MockFileSetReader) EXPECT() *_MockFileSetReaderRecorder {
	return _m.recorder
}

func (_m *MockFileSetReader) Close() error {
	ret := _m.ctrl.Call(_m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockFileSetReaderRecorder) Close() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Close")
}

func (_m *MockFileSetReader) Entries() int {
	ret := _m.ctrl.Call(_m, "Entries")
	ret0, _ := ret[0].(int)
	return ret0
}

func (_mr *_MockFileSetReaderRecorder) Entries() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Entries")
}

func (_m *MockFileSetReader) EntriesRead() int {
	ret := _m.ctrl.Call(_m, "EntriesRead")
	ret0, _ := ret[0].(int)
	return ret0
}

func (_mr *_MockFileSetReaderRecorder) EntriesRead() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "EntriesRead")
}

func (_m *MockFileSetReader) Open(_param0 uint32, _param1 time.Time) error {
	ret := _m.ctrl.Call(_m, "Open", _param0, _param1)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockFileSetReaderRecorder) Open(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Open", arg0, arg1)
}

func (_m *MockFileSetReader) Range() time0.Range {
	ret := _m.ctrl.Call(_m, "Range")
	ret0, _ := ret[0].(time0.Range)
	return ret0
}

func (_mr *_MockFileSetReaderRecorder) Range() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Range")
}

func (_m *MockFileSetReader) Read() (string, []byte, error) {
	ret := _m.ctrl.Call(_m, "Read")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].([]byte)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (_mr *_MockFileSetReaderRecorder) Read() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Read")
}

func (_m *MockFileSetReader) Validate() error {
	ret := _m.ctrl.Call(_m, "Validate")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockFileSetReaderRecorder) Validate() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Validate")
}
