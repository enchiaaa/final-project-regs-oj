package middleware

import (
	"net/http"
	"net/http/httptest"
	"online-judge/internal/rbac"
	"online-judge/internal/test_util"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequirePermission(t *testing.T){
	// 建立 Test DB
	testDB := testutil.SetupTestDB(t)

	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(AuthMiddleware(), RequirePermission(testDB, rbac.PermissionProblemUpsert))

	router.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	t.Setenv("JWT_PRIVATE_KEY_PATH", "../../keys/private.pem")
	t.Setenv("JWT_PUBLIC_KEY_PATH", "../../keys/public.pem")
	userToken, err := GenerateJWT(1, "username", "User")
	if err != nil {
		t.Fatalf("got error in GenerateJWT: %s", err)
	}

	adminToken, err := GenerateJWT(1, "username", "Admin")
	if err != nil {
		t.Fatalf("got error in GenerateJWT: %s", err)
	}

	// user: StatusForbidden
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer " + userToken)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected StatusForbidden, got %d", w.Code)
	}

	// valid token -> expect 200
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+ adminToken)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}