package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAuthMiddlewareMissingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Setenv("JWT_PRIVATE_KEY_PATH", "../../keys/private.pem")
	t.Setenv("JWT_PUBLIC_KEY_PATH", "../../keys/public.pem")

	router := gin.New()
	router.Use(AuthMiddleware())

	router.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	// missing header -> StatusUnauthorized
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected StatusUnauthorized, got %d", w.Code)
	}

	// invalid token -> StatusUnauthorized
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid token")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected StatusUnauthorized, got %d", w.Code)
	}

	// valid token -> StatusOK
	validToken, err := GenerateJWT(1, "username", "User")
	if err != nil {
		t.Fatalf("got error in GenerateJWT: %s", err)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected StatusOK, got %d", w.Code)
	}
}

func TestGenerateJWTMissingPrivateKeyEnv(t *testing.T) {
	t.Setenv("JWT_PRIVATE_KEY_PATH", "")

	_, err := GenerateJWT(1, "username", "User")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "JWT_PRIVATE_KEY_PATH is not set") {
		t.Fatalf("expected missing env error, got %s", err)
	}
}

func TestParseTokenMissingPublicKeyEnv(t *testing.T) {
	t.Setenv("JWT_PUBLIC_KEY_PATH", "")

	_, err := ParseToken("invalid-token")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "JWT_PUBLIC_KEY_PATH is not set") {
		t.Fatalf("expected missing env error, got %s", err)
	}
}
