package test

import "github.com/stretchr/testify/mock"

// MockObjectIDProvider is a test double for ObjectIDProvider.
type MockObjectIDProvider struct {
	mock.Mock
}

// GetLastObjId returns the last object ID or an error.
func (m *MockObjectIDProvider) GetLastObjId() (int, error) {
	return extractIntAndErr(m.Called())
}

// GetCurrentObjId returns the current object ID or an error.
func (m *MockObjectIDProvider) GetCurrentObjId() (int, error) {
	return extractIntAndErr(m.Called())
}

// extractIntAndErr safely extracts an int and error from mock arguments.
func extractIntAndErr(args mock.Arguments) (int, error) {
	// Defensive: avoid panic if Int(0) is missing
	val := 0
	if args.Get(0) != nil {
		val = args.Int(0)
	}
	return val, args.Error(1)
}
