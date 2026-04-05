package test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	mntSvc "2_Go/internal/maintenance"
	"2_Go/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
				On("GetLastObjId", mock.Anything).
				Return(tt.lastObjId, tt.lastObjIdErr)
			mockProvider.
				On("GetCurrentId", mock.Anything).
				Return(tt.currentObjId, tt.currentObjIdErr).
				Maybe()
			ms := &mntSvc.MaintenanceService{
				Repo: mockProvider,
			}
			match, err := ms.IsMatchedObjectDocAndCounter(context.Background())

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
				On("GetTotalRecordsWithFilePath", mock.Anything).
				Return(tt.input, tt.funcErr)

			ms := &mntSvc.MaintenanceService{
				Repo: mockProvider,
			}

			got, gotErr := ms.GetTotalFileRecord(context.Background())
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

	// test case
	tests := []struct {
		name        string
		path        string
		expected    int64
		expectError bool
	}{
		{
			name:        "existing directory with files",
			path:        "/Users/huythai/Documents/0_Projects/2_Go/server/filedata/0", // Assuming this directory exists and has files in it
			expected:    11,
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

			// ms := &mntSvc.MaintenanceService{}

			ms := mntSvc.NewMaintenanceRepository(nil)

			count, err := ms.GetTotalFileInStorage(tt.path)
			if tt.expectError {
				assert.Error(t, err, utils.Colorize("red", "An error is expected but got nil"))
				assert.Equalf(t, tt.expected, count, utils.Colorize("red", "fail negative by returning wrong number of files"))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, count, utils.Colorize("orange", "not match expect number of file"))
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
		fileProcess struct {
			totalFile      int64
			fileProcessErr error
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
			fileProcess: struct {
				totalFile      int64
				fileProcessErr error
			}{
				totalFile:      10,
				fileProcessErr: nil,
			},
			expectFuncErr: nil,
			isVerify:      true,
			expectErr:     false,
		},
		{
			name: "unmatched file path with valid path input and record",
			path: defaultPath,
			recordProcess: struct {
				totalRecord    int64
				countRecordErr error
			}{
				totalRecord:    8,
				countRecordErr: nil,
			},
			fileProcess: struct {
				totalFile      int64
				fileProcessErr error
			}{
				totalFile:      10,
				fileProcessErr: nil,
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
			fileProcess: struct {
				totalFile      int64
				fileProcessErr error
			}{
				totalFile:      0,
				fileProcessErr: mntSvc.NewError(mntSvc.ErrCodeInvalidInput, ""),
			},
			expectFuncErr: mntSvc.NewError(mntSvc.ErrCodeInvalidInput, ""),
			isVerify:      false,
			expectErr:     true,
		},
		{
			name: "cannot retrieve total record with valid input",
			path: defaultPath,
			recordProcess: struct {
				totalRecord    int64
				countRecordErr error
			}{
				totalRecord:    -1,
				countRecordErr: mntSvc.NewError(mntSvc.ErrCodeDatabaseError, "error of getting total doc path records"),
			}, fileProcess: struct {
				totalFile      int64
				fileProcessErr error
			}{
				totalFile:      10,
				fileProcessErr: nil,
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
			fileProcess: struct {
				totalFile      int64
				fileProcessErr error
			}{
				totalFile:      10,
				fileProcessErr: nil,
			},
			expectFuncErr: mntSvc.NewError(mntSvc.ErrCodeInvalidInput, ""),
			isVerify:      false,
			expectErr:     true,
		},
		{
			name: "records higher than files in storage",
			path: defaultPath,
			recordProcess: struct {
				totalRecord    int64
				countRecordErr error
			}{
				totalRecord:    11,
				countRecordErr: nil,
			},
			fileProcess: struct {
				totalFile      int64
				fileProcessErr error
			}{
				totalFile:      10,
				fileProcessErr: nil,
			},
			expectFuncErr: mntSvc.NewError(mntSvc.ErrCodeVerificationFailed, ""),
			isVerify:      false,
			expectErr:     false,
		},
		{
			name: "records less than files in storage",
			path: defaultPath,
			recordProcess: struct {
				totalRecord    int64
				countRecordErr error
			}{
				totalRecord:    8,
				countRecordErr: nil,
			},
			fileProcess: struct {
				totalFile      int64
				fileProcessErr error
			}{
				totalFile:      10,
				fileProcessErr: nil,
			},
			expectFuncErr: mntSvc.NewError(mntSvc.ErrCodeVerificationFailed, ""),
			isVerify:      false,
			expectErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := new(MockObjectIDProvider)

			mockProvider.
				On("GetTotalRecordsWithFilePath", mock.Anything).
				Return(tt.recordProcess.totalRecord, tt.recordProcess.countRecordErr).
				Maybe()

			mockProvider.
				On("GetTotalFileInStorage", tt.path).
				Return(tt.fileProcess.totalFile, tt.fileProcess.fileProcessErr).
				Maybe()

			ms := &mntSvc.MaintenanceService{
				Repo: mockProvider,
			}

			isFileVerify, err := ms.IsFilePathEqualToCountFile(context.Background(), tt.path)

			// t.Logf("isFileVerify = %v", isFileVerify)

			if tt.expectErr {
				assert.Error(t, err)

				// Compare error type.
				expectedRecordErr := tt.expectFuncErr

				customErr, ok := err.(*mntSvc.Error)
				assert.True(t, ok, utils.Colorize("red", "expected *mntSvc.Error type"))

				// Compare error code + message
				expCustomErr := expectedRecordErr.(*mntSvc.Error)
				assert.Equal(t, expCustomErr.Code, customErr.Code, utils.Colorize("red", "unmatched error code expected"))

				return
			}

			assert.Equalf(t, tt.isVerify, isFileVerify,
				utils.Colorize("red", "false result of file verification"))
			mockProvider.AssertExpectations(t)
		})
	}
}

// test for scanfile get duplication
func TestScanFileFolder(t *testing.T) {
	// Create a temporary directory for testing
	testDir := "./filedata/0/"
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	defer func() {
		removeDir := "./filedata/"
		fmt.Println("Removing:", removeDir)
		err := os.RemoveAll(removeDir)
		fmt.Println("RemoveAll err:", err)

		// Check if it exists immediately after deletion
		_, statErr := os.Stat(removeDir)
		fmt.Println("Exists right after RemoveAll:", !os.IsNotExist(statErr))
	}()

	// Set up the test environment
	tempDir := filepath.Join(testDir, "temp")
	os.MkdirAll(tempDir, 0755)

	// Create some test files with duplicate names
	files := []string{
		"8.pdf",
		"8.png",
		"9.pdf",
		"9.png",
		"1.png",
	}

	for _, fileName := range files {
		filePath := filepath.Join(testDir, fileName)
		if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		fmt.Printf("Created file: %s\n", filePath) // Print the file path
	}

	// Set the FILE_BASE_PATH environment variable
	os.Setenv("FILE_BASE_PATH", testDir)

	// read the FILE_BASE_PATH env
	fileBasePath := os.Getenv("FILE_BASE_PATH")

	// Combine the base path with a relative path
	fullbasePath, err := filepath.Abs(fileBasePath)
	if err != nil {
		t.Fatalf("Failed to get full path: %v", err)
	}

	fmt.Println(">>>Full base path: ", fullbasePath)

	ms := mntSvc.NewMaintenanceRepository(nil)

	// Call the function under test
	err = ms.ScanFileFolder()
	if err != nil {
		t.Errorf("ScanFileFolder failed: %v", err)
	}

	// Check if the duplicate files were renamed and moved to the temp directory
	expectedFiles := []string{
		"8_1.pdf",
		"8_2.png",
		"9_1.pdf",
		"9_2.png",
	}

	for _, expectedFileName := range expectedFiles {
		expectedFilePath := filepath.Join(tempDir, expectedFileName)
		if _, err := os.Stat(expectedFilePath); os.IsNotExist(err) {
			t.Errorf("Expected file %s does not exist", expectedFilePath)
		}
	}

}
