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
