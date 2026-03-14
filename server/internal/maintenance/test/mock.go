package test

import (
	"github.com/stretchr/testify/mock"
)

type MockObjectIDProvider struct {
	mock.Mock
}

func (m *MockObjectIDProvider) GetLastObjId() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}

func (m *MockObjectIDProvider) GetCurrentObjId() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}

func extractIntAndErr(args mock.Arguments) (int, error) {
	return args.Int(0), args.Error(1)
}
