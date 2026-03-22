package test

import (
	"fmt"
	"testing"

	mntSvc "2_Go/internal/maintenance"
	"2_Go/utils"

	"github.com/stretchr/testify/assert"
)

// type mockProvider struct {
// 	lastObjFn    func() (int, error)
// 	currentObjFn func() (int, error)
// }

// func (m *mockProvider) GetLast() (int, error)    { return m.lastObjFn() }
// func (m *mockProvider) GetCurrent() (int, error) { return m.currentObjFn() }

func TestIsMatchedObjectDocAndCounter(t *testing.T) {
	tests := []struct {
		name            string
		lastObjId       int
		lastObjIdErr    error
		currentObjId    int
		currentObjIdErr error
		expectedMatch   bool
		expectErr       bool
	}{
		{
			name:          "matched",
			lastObjId:     10,
			currentObjId:  10,
			expectedMatch: true,
		},
		{
			name:         "error get last obj doc",
			lastObjIdErr: mntSvc.NewError(mntSvc.ErrCodeNotFound, "last object ID not found"),
			currentObjId: 10,
			expectErr:    true,
		},
		{
			name:            "error get current obj doc",
			lastObjId:       10,
			currentObjIdErr: mntSvc.NewError(mntSvc.ErrCodeNotFound, "current object ID not found"),
			expectErr:       true,
		},
		{
			name:          "not matched last objId and current objId",
			lastObjId:     10,
			currentObjId:  11,
			expectedMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := new(MockObjectIDProvider)
			mockProvider.
				On("GetLastObjId").
				Return(tt.lastObjId, tt.lastObjIdErr)
			mockProvider.
				On("GetCurrentId").
				Return(tt.currentObjId, tt.currentObjIdErr).
				Maybe()
			ms := &mntSvc.MaintenanceService{
				Repo: mockProvider,
			}
			match, err := ms.IsMatchedObjectDocAndCounter()

			if tt.expectErr {
				assert.Error(t, err)

				customErr, ok := err.(*mntSvc.Error)
				assert.True(t, ok)
				t.Logf("\x1b[33mError code: %d, message: %s\x1b[0m", customErr.Code, customErr.Message)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedMatch, match)
			mockProvider.AssertExpectations(t)
		})
	}
}

// Test Count total documents
func TestGetTotalFilePathRecord(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		input   int64
		want    int64
		funcErr error
		wantErr bool
	}{
		{
			name:    "return valid document count",
			input:   3,
			want:    3,
			funcErr: nil,
			wantErr: false,
		},
		{
			name:    "return invalid document count",
			input:   7,
			want:    -1,
			funcErr: mntSvc.NewError(mntSvc.ErrCodeDatabaseError, "error of getting total doc path"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := new(MockObjectIDProvider)

			mockProvider.
				On("CountRecordsInFilePath").
				Return(tt.input, tt.funcErr)

			ms := &mntSvc.MaintenanceService{
				Repo: mockProvider,
			}

			got, gotErr := ms.GetTotalFileRecord()
			t.Logf("Result is::: %v", got)

			if tt.wantErr {
				assert.Error(t, gotErr)
				customErr, ok := gotErr.(*mntSvc.Error)
				assert.True(t, ok, "Error is not a custom error")
				// server/internal/maintenance/test/verify_test.go (130-130)
				t.Log(utils.Colorize("orange",
					fmt.Sprintf("Error code: %d, message: %s", customErr.Code, customErr.Message)))
				// Compare the error code
				expectedCode := tt.funcErr.(*mntSvc.Error).Code
				assert.Equal(t, expectedCode, customErr.Code,
					utils.Colorize("red", "Error codes do not match"))

				return
			}

			assert.Equal(t, tt.want, got, utils.Colorize("red", "Error GetTotalDocument() = %v, want %"))
			mockProvider.AssertExpectations(t)
		})
	}
}

// Test for GetTotalFilesInPath function
func TestGetTotalFilesInPath(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expected    int64
		expectError bool
	}{
		{
			name:        "existing directory with files",
			path:        "/Users/huythai/Documents/0_Projects/2_Go/server/filedata/0", // Assuming this directory exists and has files in it
			expected:    10,
			expectError: false,
		},
		{
			name:        "non-existent directory",
			path:        "./non_existent_directory",
			expected:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ms := &mntSvc.MaintenanceService{}

			count, err := ms.GetTotalFilesInPath(tt.path)
			if tt.expectError {
				assert.Error(t, err)
				assert.Equalf(t, tt.expected, count, "\x1b[31mfail negative by returning wrong number of files\x1b[0m")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, count)
			}
		})
	}
}

// Test if verification of number path file and path record.
func TestIsFilePathEqualToCountFile(t *testing.T) {
	defaultPath := "/Users/huythai/Documents/0_Projects/2_Go/server/filedata/"
	tests := []struct {
		name          string
		path          string
		recordProcess struct {
			totalRecord    int64
			countRecordErr error
		}
		expectFuncErr error
		isVerify      bool
		expectErr     bool
	}{
		// Add test case
		{
			name: "matched file path and record",
			path: defaultPath,
			recordProcess: struct {
				totalRecord    int64
				countRecordErr error
			}{
				totalRecord:    10,
				countRecordErr: nil,
			},
			expectFuncErr: nil,
			isVerify:      true,
			expectErr:     false,
		},
		{
			name: "unmatched file path and record",
			path: defaultPath,
			recordProcess: struct {
				totalRecord    int64
				countRecordErr error
			}{
				totalRecord:    8,
				countRecordErr: nil,
			},
			expectFuncErr: nil,
			isVerify:      false,
			expectErr:     false,
		},
		{
			name: "invalid path input",
			path: "",
			recordProcess: struct {
				totalRecord    int64
				countRecordErr error
			}{
				totalRecord:    8,
				countRecordErr: nil,
			},
			expectFuncErr: mntSvc.NewError(mntSvc.ErrCodeInvalidInput, ""),
			isVerify:      false,
			expectErr:     true,
		},
		{
			name: "cannot retrieve total record with valid input",
			path: "/Users/huythai/Documents/0_Projects/2_Go/server/filedata/",
			recordProcess: struct {
				totalRecord    int64
				countRecordErr error
			}{
				totalRecord:    -1,
				countRecordErr: mntSvc.NewError(mntSvc.ErrCodeDatabaseError, "error of getting total doc path records"),
			},
			expectFuncErr: mntSvc.NewError(mntSvc.ErrCodeDatabaseError, ""),
			isVerify:      false,
			expectErr:     true,
		},
		{
			name: "cannot retrieve total record with invalid input",
			recordProcess: struct {
				totalRecord    int64
				countRecordErr error
			}{
				totalRecord:    -1,
				countRecordErr: mntSvc.NewError(mntSvc.ErrCodeDatabaseError, "error of getting total doc path records"),
			},
			expectFuncErr: mntSvc.NewError(mntSvc.ErrCodeInvalidInput, ""),
			isVerify:      false,
			expectErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := new(MockObjectIDProvider)

			mockProvider.
				On("CountRecordsInFilePath").
				Return(tt.recordProcess.totalRecord, tt.recordProcess.countRecordErr).
				Maybe()
			// mockProvider.
			// On("IsFilePathEqualToCountFile",path).
			// Return(true,false)

			ms := &mntSvc.MaintenanceService{
				Repo: mockProvider,
			}

			isFileVerify, err := ms.IsFilePathEqualToCountFile(tt.path)

			// t.Logf("isFileVerify = %v", isFileVerify)

			if tt.expectErr {
				assert.Error(t, err)

				// Compare error type.
				expectedRecordErr := tt.expectFuncErr

				// TODO: expected Path error
				customErr, ok := err.(*mntSvc.Error)
				assert.True(t, ok, utils.Colorize("red", "expected *mntSvc.Error type"))

				// Compare error code + message
				expCustomErr := expectedRecordErr.(*mntSvc.Error)
				assert.Equal(t, expCustomErr.Code, customErr.Code, utils.Colorize("red", "unmatched error code"))

				// If want to check error message.
				// assert.Equal(t, expCustomErr.Message, customErr.Message, utils.Colorize("red", "unmatched error message"))
				// if fail check record

				// if fail to check file doc
				return
			}

			assert.Equalf(t, tt.isVerify, isFileVerify,
				utils.Colorize("red", "false result of file verification"))
			mockProvider.AssertExpectations(t)
		})
	}
}
