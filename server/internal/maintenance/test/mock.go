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

// GetTotalDocumentRecords returns the max number of object file doc or an error.
func (m *MockObjectIDProvider) GetTotalRecordsWithFilePath() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

// Get total file in file storage
func (m *MockObjectIDProvider) GetTotalFileInStorage(path string) (int64, error) {
	args := m.Called(path)
	return args.Get(0).(int64), args.Error(1)
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

// extractBoolAndErr safely extracts a bool and error from mock arguments.
// func extractBoolAndErr(args mock.Arguments) (bool, error) {
// 	var result bool
// 	if args.Get(0) != nil {
// 		result = args.Get(0).(bool)
// 	}
// 	return result, args.Error(1)
// }
