// 負責檢查「目前使用者」是否擁有存取該 API 所需的 Permission
package middleware

import (
	"net/http"
	"online-judge/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 確認使用者是否有權限
func RequirePermission(db *gorm.DB, permissionName string) gin.HandlerFunc{
	return func(c *gin.Context) {
		roleName := c.GetString("role")
		role := models.Role{}
		if err :=db.Preload("Permissions").Where("name = ?", roleName).First(&role).Error; err != nil{
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "can't get role in database"})
			return
		}

		// 查 role 是否有 permissionName
		for i := range role.Permissions {
			if role.Permissions[i].Name == permissionName {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	}
}