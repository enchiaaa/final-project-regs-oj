package middleware

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   uint   `json:"userId"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// 生成 JWT 並回傳 JWT
func GenerateJWT(userID uint, username string, role string) (string, error) {
	// 1. 建立 claims
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)

	// 2. 用私鑰簽章
	// 取得 private_key
	pemBytes, err := os.ReadFile("keys/private.pem")
	if err != nil {
		return "", err
	}
	privateKey, err := jwt.ParseECPrivateKeyFromPEM(pemBytes)
	if err != nil {
		return "", err
	}

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}

	// 3. 回傳 token string.
	return tokenString, nil
}

// 驗證 JWT，回傳 Claims 內容
func ParseToken(tokenString string) (*Claims, error) {
	// 取得 public_key
	pemBytes, err := os.ReadFile("keys/public.pem")
	if err != nil {
		return nil, err
	}
	publicKey, err := jwt.ParseECPublicKeyFromPEM(pemBytes)
	if err != nil {
		return nil, err
	}

	claims := &Claims{}

	jwtToken, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodES256 {
			return nil, errors.New("unexpected signing method")
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !jwtToken.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// 驗證 request header，讀取 JWT
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 讀 Authorization header
		authorization := c.GetHeader("Authorization")

		if authorization == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		// 2. 檢查是不是 Bearer 開頭
		if !strings.HasPrefix(authorization, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			return
		}

		// 3. 取出 token string
		tokenString := strings.TrimPrefix(authorization, "Bearer ")

		// 4. 呼叫 ParseToken
		claims, err := ParseToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// 5. 把 claims 放進 c.Set
		c.Set("claims", claims)
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}
