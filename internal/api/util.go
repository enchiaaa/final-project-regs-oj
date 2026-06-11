package api

import (
	"online-judge/internal/rbac"

	"github.com/gin-gonic/gin"
)

func IsAdmin(c *gin.Context) bool {
	return c.GetString("role") == rbac.RoleAdmin
}

func CanAccessUserResource(c *gin.Context, ownerID uint) bool {
	if IsAdmin(c) {
		return true
	}
	return c.GetUint("userID") == ownerID
}