// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	peer "github.com/libp2p/go-libp2p-core/peer"
	mock "github.com/stretchr/testify/mock"
)

// Db is an autogenerated mock type for the Db type
type Db struct {
	mock.Mock
}

// Nodes provides a mock function with given fields:
func (_m *Db) Nodes() ([]peer.ID, error) {
	ret := _m.Called()

	var r0 []peer.ID
	if rf, ok := ret.Get(0).(func() []peer.ID); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]peer.ID)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewDb interface {
	mock.TestingT
	Cleanup(func())
}

// NewDb creates a new instance of Db. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewDb(t mockConstructorTestingTNewDb) *Db {
	mock := &Db{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
