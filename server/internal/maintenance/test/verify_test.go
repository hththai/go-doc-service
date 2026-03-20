package test

import (
	"testing"

	mntSvc "2_Go/internal/maintenance"

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
func TestGetTotalDocument(t *testing.T) {
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
			input:   4,
			want:    3,
			funcErr: mntSvc.NewError(mntSvc.ErrCodeNotFound, "error of getting total doc path"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := new(MockObjectIDProvider)

			mockProvider.
				On("GetTotalDocument",
					int64(tt.input)).Return(tt.want, tt.funcErr)

			ms := &mntSvc.MaintenanceService{
				Repo: nil,
			}

			got, gotErr := ms.GetTotalDocument(tt.input)
			t.Logf("Result is::: %v", got)
			if tt.wantErr {
				assert.Error(t, gotErr)

				customErr, ok := gotErr.(*mntSvc.Error)
				assert.True(t, ok)
				t.Logf("\x1b[33mError code: %d, message: %s\x1b[0m", customErr.Code, customErr.Message)
			} else {
				assert.NoError(t, gotErr)
			}

			assert.Equal(t, tt.want, got)
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
