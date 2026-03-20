package test

import "github.com/stretchr/testify/mock"

// MockObjectIDProvider is a test double for ObjectIDProvider.
type MockObjectIDProvider struct {
	mock.Mock
}

// GetLastObjId returns the last object ID or an error.
func (m *MockObjectIDProvider) GetLastObjId() (int64, error) {
	return extractIntAndErr(m.Called())
}

// GetCurrentObjId returns the current object ID or an error.
func (m *MockObjectIDProvider) GetCurrentId() (int64, error) {
	return extractIntAndErr(m.Called())
}

// GetTotalDocument returns the max number of object file doc or an error.
// func (m *MockObjectIDProvider) GetTotalDocument(input int64) (int64, error) {
// 	if input > 0 {
// 		return input, nil
// 	}

// 	return -1, NewError(ErrCodeNotFound, "error of getting total doc path")
// }

func (m *MockObjectIDProvider) GetTotalDocument(input int64) (int64, error) {
	args := m.Called(input)

	var val int64
	if v := args.Get(0); v != nil {
		switch t := v.(type) {
		case int:
			val = int64(t)
		case int64:
			val = t
		}
	}

	return val, args.Error(1)
}

// extractIntAndErr safely extracts an int and error from mock arguments.
func extractIntAndErr(args mock.Arguments) (int64, error) {
	// Defensive: avoid panic if Int(0) is missing
	var val int64
	if args.Get(0) != nil {
		val = int64(args.Get(0).(int))
	}
	return val, args.Error(1)
}
