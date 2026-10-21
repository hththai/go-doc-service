package test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	mnt "2_Go/internal/maintenance"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const testUserID = 1

// storage is a temporary document store with a quarantine folder outside it.
type storage struct {
	base       string
	quarantine string
}

func newStorage(t *testing.T) storage {
	t.Helper()
	root := t.TempDir()
	s := storage{base: filepath.Join(root, "0"), quarantine: filepath.Join(root, "quarantine")}
	require.NoError(t, os.MkdirAll(filepath.Join(s.base, "0"), 0755))
	return s
}

// write creates base/0/<name> with content and returns its path.
func (s storage) write(t *testing.T, name, content string) string {
	t.Helper()
	p := filepath.Join(s.base, "0", name)
	require.NoError(t, os.WriteFile(p, []byte(content), 0644))
	return p
}

func (s storage) quarantined(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir(s.quarantine)
	if os.IsNotExist(err) {
		return nil
	}
	require.NoError(t, err)
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

func newService(repo *MockObjectIDProvider, s storage) *mnt.MaintenanceService {
	return mnt.NewMaintenanceService(repo, s.base, s.quarantine)
}

func codes(issues []mnt.DocIssue) []mnt.IssueCode {
	var c []mnt.IssueCode
	for _, i := range issues {
		c = append(c, i.Code)
	}
	return c
}

func TestVerifyUserDocuments(t *testing.T) {
	t.Run("healthy document has no issues", func(t *testing.T) {
		s := newStorage(t)
		p := s.write(t, "8.png", "1234")
		s.write(t, "80.pdf", "other doc") // different stem, not a duplicate
		s.write(t, ".DS_Store", "hidden") // hidden, ignored

		repo := new(MockObjectIDProvider)
		repo.On("GetDocFileMeta", mock.Anything, testUserID).Return([]mnt.DocFileMeta{
			{ObjId: 8, Title: "Receipt", FilePath: p, Extension: ".png", FileSize: 4},
		}, nil)

		issues, err := newService(repo, s).VerifyUserDocuments(context.Background(), testUserID)
		assert.NoError(t, err)
		assert.Empty(t, issues)
		assert.NotNil(t, issues, "must be an empty list, not nil, for JSON")
	})

	t.Run("duplicate and stale metadata", func(t *testing.T) {
		s := newStorage(t)
		p := s.write(t, "8.png", "1234")
		s.write(t, "8.pdf", "old")

		repo := new(MockObjectIDProvider)
		repo.On("GetDocFileMeta", mock.Anything, testUserID).Return([]mnt.DocFileMeta{
			{ObjId: 8, Title: "Receipt", FilePath: p, Extension: ".pdf", FileSize: 3},
		}, nil)

		issues, err := newService(repo, s).VerifyUserDocuments(context.Background(), testUserID)
		assert.NoError(t, err)
		assert.Equal(t, []mnt.IssueCode{mnt.IssueDuplicateFile, mnt.IssueMetadataMismatch}, codes(issues))
		assert.Equal(t, []string{"8.pdf"}, issues[0].ExtraFiles)
		assert.Equal(t, "8.png", issues[0].FileName)
		assert.True(t, issues[0].Repairable)
		assert.True(t, issues[1].Repairable)
		assert.NotContains(t, issues[0].Message, s.base, "server paths must not leak")
	})

	t.Run("extension compare ignores dot and case", func(t *testing.T) {
		s := newStorage(t)
		p := s.write(t, "8.PNG", "1234")

		repo := new(MockObjectIDProvider)
		repo.On("GetDocFileMeta", mock.Anything, testUserID).Return([]mnt.DocFileMeta{
			{ObjId: 8, FilePath: p, Extension: "png", FileSize: 4},
		}, nil)

		issues, err := newService(repo, s).VerifyUserDocuments(context.Background(), testUserID)
		assert.NoError(t, err)
		assert.Empty(t, issues)
	})

	t.Run("missing file is not repairable", func(t *testing.T) {
		s := newStorage(t)
		repo := new(MockObjectIDProvider)
		repo.On("GetDocFileMeta", mock.Anything, testUserID).Return([]mnt.DocFileMeta{
			{ObjId: 9, FilePath: filepath.Join(s.base, "0", "9.pdf"), Extension: ".pdf", FileSize: 10},
		}, nil)

		issues, err := newService(repo, s).VerifyUserDocuments(context.Background(), testUserID)
		assert.NoError(t, err)
		assert.Equal(t, []mnt.IssueCode{mnt.IssueMissingFile}, codes(issues))
		assert.False(t, issues[0].Repairable)
	})

	t.Run("multiple path rows is reported only", func(t *testing.T) {
		s := newStorage(t)
		p := s.write(t, "8.png", "1234")
		repo := new(MockObjectIDProvider)
		repo.On("GetDocFileMeta", mock.Anything, testUserID).Return([]mnt.DocFileMeta{
			{ObjId: 8, FilePath: p, Extension: ".png", FileSize: 4},
			{ObjId: 8, FilePath: filepath.Join(s.base, "0", "8.pdf"), Extension: ".png", FileSize: 4},
		}, nil)

		issues, err := newService(repo, s).VerifyUserDocuments(context.Background(), testUserID)
		assert.NoError(t, err)
		assert.Equal(t, []mnt.IssueCode{mnt.IssueMultiplePaths}, codes(issues))
		assert.False(t, issues[0].Repairable)
	})

	t.Run("database error", func(t *testing.T) {
		s := newStorage(t)
		repo := new(MockObjectIDProvider)
		repo.On("GetDocFileMeta", mock.Anything, testUserID).Return(nil, errors.New("db down"))

		_, err := newService(repo, s).VerifyUserDocuments(context.Background(), testUserID)
		customErr, ok := err.(*mnt.Error)
		require.True(t, ok)
		assert.Equal(t, mnt.ErrCodeDatabaseError, customErr.Code)
	})
}

func TestRepairUserDocuments(t *testing.T) {
	ctx := context.Background()

	t.Run("moves duplicate to quarantine and keeps stored file", func(t *testing.T) {
		s := newStorage(t)
		stored := s.write(t, "8.png", "1234")
		extra := s.write(t, "8.pdf", "old")

		repo := new(MockObjectIDProvider)
		repo.On("GetDocFileMeta", mock.Anything, testUserID).Return([]mnt.DocFileMeta{
			{ObjId: 8, FilePath: stored, Extension: ".png", FileSize: 4},
		}, nil)
		repo.On("InsertIssueRecord", mock.Anything, mock.MatchedBy(func(r mnt.ObjVerifyRecord) bool {
			return r.ObjId == 8 && !r.HasIssue && r.IssueCode == int(mnt.IssueDuplicateFile)
		})).Return(int64(1), nil)

		result, err := newService(repo, s).RepairUserDocuments(ctx, testUserID,
			[]mnt.IssueRef{{ObjId: 8, Code: mnt.IssueDuplicateFile}})

		assert.NoError(t, err)
		assert.Len(t, result.Repaired, 1)
		assert.Empty(t, result.Skipped)
		assert.FileExists(t, stored)
		assert.NoFileExists(t, extra)
		q := s.quarantined(t)
		require.Len(t, q, 1)
		assert.Regexp(t, `^8_\d+_8\.pdf$`, q[0])
		repo.AssertExpectations(t)
	})

	t.Run("updates metadata from the file on disk", func(t *testing.T) {
		s := newStorage(t)
		stored := s.write(t, "8.png", "1234")

		repo := new(MockObjectIDProvider)
		repo.On("GetDocFileMeta", mock.Anything, testUserID).Return([]mnt.DocFileMeta{
			{ObjId: 8, FilePath: stored, Extension: ".pdf", FileSize: 99},
		}, nil)
		repo.On("UpdateFileMetadata", mock.Anything, int64(8), testUserID, ".png", float64(4)).Return(nil)
		repo.On("InsertIssueRecord", mock.Anything, mock.Anything).Return(int64(1), nil)

		result, err := newService(repo, s).RepairUserDocuments(ctx, testUserID,
			[]mnt.IssueRef{{ObjId: 8, Code: mnt.IssueMetadataMismatch}})

		assert.NoError(t, err)
		assert.Len(t, result.Repaired, 1)
		repo.AssertExpectations(t)
	})

	t.Run("repairs only the selected issues", func(t *testing.T) {
		s := newStorage(t)
		stored := s.write(t, "8.png", "1234")
		extra := s.write(t, "8.pdf", "old")

		repo := new(MockObjectIDProvider)
		repo.On("GetDocFileMeta", mock.Anything, testUserID).Return([]mnt.DocFileMeta{
			{ObjId: 8, FilePath: stored, Extension: ".pdf", FileSize: 99}, // duplicate + mismatch
		}, nil)
		repo.On("UpdateFileMetadata", mock.Anything, int64(8), testUserID, ".png", float64(4)).Return(nil)
		repo.On("InsertIssueRecord", mock.Anything, mock.Anything).Return(int64(1), nil)

		result, err := newService(repo, s).RepairUserDocuments(ctx, testUserID,
			[]mnt.IssueRef{{ObjId: 8, Code: mnt.IssueMetadataMismatch}})

		assert.NoError(t, err)
		assert.Len(t, result.Repaired, 1)
		assert.FileExists(t, extra, "unselected duplicate must stay")
		assert.Empty(t, s.quarantined(t))
	})

	t.Run("skips missing file, unknown and other user's documents", func(t *testing.T) {
		s := newStorage(t)

		repo := new(MockObjectIDProvider)
		// Only the user's documents are returned; obj 50 belongs to someone else.
		repo.On("GetDocFileMeta", mock.Anything, testUserID).Return([]mnt.DocFileMeta{
			{ObjId: 9, FilePath: filepath.Join(s.base, "0", "9.pdf"), Extension: ".pdf", FileSize: 10},
		}, nil)

		result, err := newService(repo, s).RepairUserDocuments(ctx, testUserID, []mnt.IssueRef{
			{ObjId: 9, Code: mnt.IssueMissingFile},
			{ObjId: 50, Code: mnt.IssueDuplicateFile},
			{ObjId: 9, Code: mnt.IssueDuplicateFile}, // issue that does not exist
		})

		assert.NoError(t, err)
		assert.Empty(t, result.Repaired)
		assert.Len(t, result.Skipped, 3)
		repo.AssertNotCalled(t, "InsertIssueRecord", mock.Anything, mock.Anything)
		repo.AssertNotCalled(t, "UpdateFileMetadata", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("refuses to move files outside storage", func(t *testing.T) {
		s := newStorage(t)
		// Stored path outside the base path, e.g. a misconfigured FILE_DATA_BASE_PATH.
		outside := t.TempDir()
		stored := filepath.Join(outside, "8.png")
		require.NoError(t, os.WriteFile(stored, []byte("1234"), 0644))
		extra := filepath.Join(outside, "8.pdf")
		require.NoError(t, os.WriteFile(extra, []byte("old"), 0644))

		repo := new(MockObjectIDProvider)
		repo.On("GetDocFileMeta", mock.Anything, testUserID).Return([]mnt.DocFileMeta{
			{ObjId: 8, FilePath: stored, Extension: ".png", FileSize: 4},
		}, nil)

		result, err := newService(repo, s).RepairUserDocuments(ctx, testUserID,
			[]mnt.IssueRef{{ObjId: 8, Code: mnt.IssueDuplicateFile}})

		assert.NoError(t, err)
		assert.Empty(t, result.Repaired)
		require.Len(t, result.Skipped, 1)
		assert.Error(t, result.Skipped[0].Err)
		assert.FileExists(t, extra)
	})

	t.Run("duplicate selections are repaired once", func(t *testing.T) {
		s := newStorage(t)
		stored := s.write(t, "8.png", "1234")
		s.write(t, "8.pdf", "old")

		repo := new(MockObjectIDProvider)
		repo.On("GetDocFileMeta", mock.Anything, testUserID).Return([]mnt.DocFileMeta{
			{ObjId: 8, FilePath: stored, Extension: ".png", FileSize: 4},
		}, nil)
		repo.On("InsertIssueRecord", mock.Anything, mock.Anything).Return(int64(1), nil).Once()

		ref := mnt.IssueRef{ObjId: 8, Code: mnt.IssueDuplicateFile}
		result, err := newService(repo, s).RepairUserDocuments(ctx, testUserID, []mnt.IssueRef{ref, ref})

		assert.NoError(t, err)
		assert.Len(t, result.Repaired, 1)
		assert.Empty(t, result.Skipped)
		repo.AssertExpectations(t)
	})
}
