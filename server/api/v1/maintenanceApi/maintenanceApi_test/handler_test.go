package maintenanceApi_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"2_Go/api/v1/maintenanceApi"
	mnt "2_Go/internal/maintenance"
	mntTest "2_Go/internal/maintenance/test"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// newRouter registers the maintenance routes; userID 0 means not authenticated.
func newRouter(repo *mntTest.MockObjectIDProvider, userID int) *gin.Engine {
	gin.SetMode(gin.TestMode)
	logger, _ := test.NewNullLogger()
	svc := mnt.NewMaintenanceService(repo, "", "")

	r := gin.New()
	g := r.Group("/v1/auth", func(c *gin.Context) {
		if userID != 0 {
			c.Set("userId", userID)
		}
	})
	maintenanceApi.Register(g, maintenanceApi.NewHandler(svc, logger))
	return r
}

func do(r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func TestHandleVerify(t *testing.T) {
	t.Run("returns empty issues list", func(t *testing.T) {
		repo := new(mntTest.MockObjectIDProvider)
		repo.On("GetDocFileMeta", mock.Anything, 7).Return([]mnt.DocFileMeta{}, nil)

		w := do(newRouter(repo, 7), "GET", "/v1/auth/maintenance/verify", "")

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"issues":[]}`, w.Body.String())
		repo.AssertExpectations(t)
	})

	t.Run("not authenticated", func(t *testing.T) {
		w := do(newRouter(new(mntTest.MockObjectIDProvider), 0), "GET", "/v1/auth/maintenance/verify", "")
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		repo := new(mntTest.MockObjectIDProvider)
		repo.On("GetDocFileMeta", mock.Anything, 7).Return(nil, errors.New("db down"))

		w := do(newRouter(repo, 7), "GET", "/v1/auth/maintenance/verify", "")
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestHandleRepair(t *testing.T) {
	t.Run("invalid bodies", func(t *testing.T) {
		for name, body := range map[string]string{
			"empty body":   "",
			"no issues":    `{"issues":[]}`,
			"missing code": `{"issues":[{"objId":8}]}`,
			"not json":     `issues`,
		} {
			t.Run(name, func(t *testing.T) {
				repo := new(mntTest.MockObjectIDProvider)
				w := do(newRouter(repo, 7), "POST", "/v1/auth/maintenance/repair", body)
				assert.Equal(t, http.StatusBadRequest, w.Code)
				repo.AssertNotCalled(t, "GetDocFileMeta", mock.Anything, mock.Anything)
			})
		}
	})

	t.Run("unknown issue is skipped", func(t *testing.T) {
		repo := new(mntTest.MockObjectIDProvider)
		repo.On("GetDocFileMeta", mock.Anything, 7).Return([]mnt.DocFileMeta{}, nil)

		w := do(newRouter(repo, 7), "POST", "/v1/auth/maintenance/repair", `{"issues":[{"objId":8,"code":1}]}`)

		assert.Equal(t, http.StatusOK, w.Code)
		var result mnt.RepairResult
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
		assert.Empty(t, result.Repaired)
		assert.Len(t, result.Skipped, 1)
		assert.Equal(t, int64(8), result.Skipped[0].ObjId)
	})

	t.Run("not authenticated", func(t *testing.T) {
		w := do(newRouter(new(mntTest.MockObjectIDProvider), 0), "POST", "/v1/auth/maintenance/repair", `{"issues":[{"objId":8,"code":1}]}`)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
