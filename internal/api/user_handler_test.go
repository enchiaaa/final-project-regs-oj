package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"online-judge/internal/test_util"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestUserLoginHandler(t *testing.T) {
	// 定義 request 和 response
	type LoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	type LoginResponse struct {
		Message string `json:"message"`
		User    string `json:"user"`
		Token   string `json:"token"`
	}

	t.Setenv("JWT_PRIVATE_KEY_PATH", "../../keys/private.pem")

	t.Run("登入成功", func(t *testing.T) {
		// 1. 建立 Test DB
		testDB := testutil.SetupTestDB(t)

		// 2. 建立 User
		user := testutil.CreateTestUser(t, testDB, "login-user")

		// 3. 建立 router
		router := gin.New()
		router.POST("/api/users/login", UserLoginHandler(testDB))

		// 4. 建立 JSON request
		requestBody := UserLoginRequest{
			Username: user.Username,
			Password: testutil.TestPassword,
		}
		body, err := json.Marshal(requestBody)
		if err != nil {
			t.Fatalf("failed to marshal login request: %v", err)
		}
		req := httptest.NewRequest(http.MethodPost, "/api/users/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		// 5. Call API
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 6. 確認結果
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d\nbody: %s", w.Code, w.Body.String())
		}

		// 7. Parse response JSON
		var response LoginResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatal(err)
		}

		// 8. 確認 token 不是空字串
		if response.Token == "" {
			t.Fatal("expected JWT token, got empty string")
		}
	})

	t.Run("密碼錯誤", func(t *testing.T) {
		// 1. 建立 Test DB
		testDB := testutil.SetupTestDB(t)

		// 2. 建立 User
		user := testutil.CreateTestUser(t, testDB, "login-user")

		// 3. 建立 router
		router := gin.New()
		router.POST("/api/users/login", UserLoginHandler(testDB))

		// 4. 建立 JSON request
		requestBody := UserLoginRequest{
			Username: user.Username,
			Password: "wrong password",
		}
		body, err := json.Marshal(requestBody)
		if err != nil {
			t.Fatalf("failed to marshal login request: %v", err)
		}
		req := httptest.NewRequest(http.MethodPost, "/api/users/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		// 5. Call API
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 6. 確認結果
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d\nbody: %s", w.Code, w.Body.String())
		}
	})
}
