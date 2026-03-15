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
		name          string
		lastReturn    int
		lastErr       error
		currentReturn int
		currentErr    error
		expectedMatch bool
		expectErr     bool
	}{
		{
			name:          "matched",
			lastReturn:    10,
			currentReturn: 10,
			expectedMatch: true,
		},
		{
			name:          "error get last obj doc",
			lastErr:       mntSvc.NewError(mntSvc.ErrCodeNotFound, "last object ID not found"),
			currentReturn: 10,
			expectErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockProvider := new(MockObjectIDProvider)

			mockProvider.
				On("GetLastObjId").
				Return(tt.lastReturn, tt.lastErr)

			mockProvider.
				On("GetCurrentId").
				Return(tt.currentReturn, tt.currentErr).
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
