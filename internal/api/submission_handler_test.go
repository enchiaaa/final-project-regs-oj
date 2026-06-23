package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"online-judge/internal/rbac"
	"online-judge/internal/test_util"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func setupRerunTestRouter(db *gorm.DB, userID uint) (*gin.Engine, chan string) {
	router := gin.New()
	jobQueue := make(chan string, 100)
	router.POST(
		"/api/submissions/:operatorId/rerun",
		testutil.MockAuthMiddleware(userID, rbac.RoleUser),
		RerunSubmissionHandler(db, jobQueue),
	)

	return router, jobQueue
}

func TestRerunSubmissionHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Rerun 成功", func(t *testing.T) {
		// 1. 建立 Test DB
		testDB := testutil.SetupTestDB(t)

		// 2. 建立測試資料
		user := testutil.CreateTestUser(t, testDB, "rerun-user")
		problem := testutil.CreateTestProblem(t, testDB, "rerun-problem")

		// 3. 設定測試 router
		router, testJobQueue := setupRerunTestRouter(testDB, user.ID)

		// 4. 測試每一種 status 的情況
		statuses := []string{
			"CE", "SE", "RE", "TLE", "AC", "WA",
		}
		for _, status := range statuses {
			newSubmission := testutil.CreateTestSubmission(t, testDB, user.ID, problem.ID, status)

			// 5. Call API
			w := httptest.NewRecorder()
			url := fmt.Sprintf("/api/submissions/%s/rerun", newSubmission.OperatorID)
			req := httptest.NewRequest(http.MethodPost, url, nil)
			router.ServeHTTP(w, req)

			// 6. 確認結果
			if w.Code != http.StatusAccepted {
				t.Fatalf("status %s: expected HTTP 202, got %d\nbody: %s", status, w.Code, w.Body.String())
			}
			if err := testDB.First(&newSubmission, newSubmission.ID).Error; err != nil {
				t.Fatalf("failed to reload submission: %v", err)
			}
			if newSubmission.Status != "Pending" {
				t.Fatalf("expected submission status Pending after rerun, got %s", newSubmission.Status)
			}

			// Queue 收到正確 operatorId
			queuedID := <-testJobQueue
			if queuedID != newSubmission.OperatorID {
				t.Fatalf("expected queued operatorId %s, got %s", newSubmission.OperatorID, queuedID)
			}
		}
	})

	t.Run("Rerun Conflict", func(t *testing.T) {
		// 1. 建立 Test DB
		testDB := testutil.SetupTestDB(t)

		// 2. 建立測試資料
		user := testutil.CreateTestUser(t, testDB, "rerun-user")
		problem := testutil.CreateTestProblem(t, testDB, "rerun-problem")

		// 3. 設定測試 router
		router, _ := setupRerunTestRouter(testDB, user.ID)

		// 4. 測試每一種 status 的情況
		statuses := []string{
			"Compiling", "Configuring", "Pending", "Judging",
		}
		for _, status := range statuses {
			newSubmission := testutil.CreateTestSubmission(t, testDB, user.ID, problem.ID, status)

			// 5. Call API
			w := httptest.NewRecorder()
			url := fmt.Sprintf("/api/submissions/%s/rerun", newSubmission.OperatorID)
			req := httptest.NewRequest(http.MethodPost, url, nil)
			router.ServeHTTP(w, req)

			// 6. 確認結果
			if w.Code != http.StatusConflict {
				t.Fatalf("status %s: expected HTTP 409, got %d\nbody: %s", status, w.Code, w.Body.String())
			}
		}
	})

	t.Run("Rerun Forbidden", func(t *testing.T) {
		// 1. 建立 Test DB
		testDB := testutil.SetupTestDB(t)

		// 2. 建立測試資料
		requestingUser := testutil.CreateTestUser(t, testDB, "requesting-user")
		ownerUser := testutil.CreateTestUser(t, testDB, "submission-owner")
		problem := testutil.CreateTestProblem(t, testDB, "rerun-problem")
		newSubmission := testutil.CreateTestSubmission(t, testDB, ownerUser.ID, problem.ID, "AC")

		// 3. 設定測試 router
		router, _ := setupRerunTestRouter(testDB, requestingUser.ID)

		// 4. Call API
		w := httptest.NewRecorder()
		url := fmt.Sprintf("/api/submissions/%s/rerun", newSubmission.OperatorID)
		req := httptest.NewRequest(http.MethodPost, url, nil)
		router.ServeHTTP(w, req)

		// 5. 確認結果
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected HTTP 403, got %d\nbody: %s", w.Code, w.Body.String())
		}
	})

	t.Run("Rerun Submission Not Found", func(t *testing.T) {
		// 1. 建立 Test DB
		testDB := testutil.SetupTestDB(t)

		// 2. 設定測試 router
		router, _ := setupRerunTestRouter(testDB, 0)

		// 3. Call API
		w := httptest.NewRecorder()
		url := "/api/submissions/1/rerun"
		req := httptest.NewRequest(http.MethodPost, url, nil)
		router.ServeHTTP(w, req)

		// 4. 確認結果
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected HTTP 404, got %d\nbody: %s", w.Code, w.Body.String())
		}
	})
}
