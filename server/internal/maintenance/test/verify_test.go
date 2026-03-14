package maintenance

import (
	"testing"

	mntSvc "2_Go/internal/maintenance"

	"github.com/stretchr/testify/assert"
)

type mockProvider struct {
	lastObjFn    func() (int, error)
	currentObjFn func() (int, error)
}

func (m *mockProvider) GetLast() (int, error)    { return m.lastObjFn() }
func (m *mockProvider) GetCurrent() (int, error) { return m.currentObjFn() }

func TestIsMatchedObjectDocAndCounter(t *testing.T) {
	tests := []struct {
		name          string
		lastObjFn     func() (int, error)
		currentObjFn  func() (int, error)
		expectedMatch bool
		expectErr     bool
	}{
		{
			name:          "matched",
			lastObjFn:     func() (int, error) { return 10, nil },
			currentObjFn:  func() (int, error) { return 10, nil },
			expectedMatch: true,
		},
		{
			name:         "error get last obj doc",
			lastObjFn:    func() (int, error) { return 0, mntSvc.NewError(mntSvc.ErrCodeNotFound, "last object ID not found") },
			currentObjFn: func() (int, error) { return 10, nil },
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &mntSvc.MaintenanceService{
				Provider: &mockProvider{
					lastObjFn:    tt.lastObjFn,
					currentObjFn: tt.currentObjFn,
				},
			}

			match, err := ms.IsMatchedObjectDocAndCounter()

			if tt.expectErr {
				assert.Error(t, err)

				customErr, ok := err.(*mntSvc.Error)
				assert.True(t, ok, "error should be *mntSvc.Error")

				t.Logf("\x1b[31mError code: %d, message: %s\x1b[0m", customErr.Code, customErr.Message)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedMatch, match)
		})
	}
}
