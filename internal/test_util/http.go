package testutil

import "github.com/gin-gonic/gin"

func MockAuthMiddleware(userID uint, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("userID", userID)
		c.Set("role", role)
		c.Next()
	}
}
