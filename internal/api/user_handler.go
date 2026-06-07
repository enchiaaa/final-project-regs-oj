// 使用者相關的 API 處理函數
package api

import (
	"errors"
	"net/http"
	"online-judge/internal/middleware"
	"online-judge/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	// "net/http"
)

type UserRegisterRequest struct {
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"password" form:"password" binding:"required,min=8"`
}
type UserLoginRequest struct {
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}

// /api/users/register POST 註冊
func UserRegisterHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UserRegisterRequest
		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 確認 username 是否已存在
		var existingUser models.User
		err := db.Where("username = ?", req.Username).First(&existingUser).Error
		if err == nil {
			c.JSON(http.StatusConflict, gin.H{
				"error": "username 已存在",
			})
			return
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check username"})
			return
		}

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}

		userRole := models.Role{}
		if err := db.Where("name = ?", "User").First(&userRole).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load default user role"})
			return
		}
		newUser := models.User{
			Username:     req.Username,
			PasswordHash: string(passwordHash),
			RoleID:       userRole.ID,
		}

		if err := db.Create(&newUser).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "User received successfully!",
			"user":    req.Username,
		})
	}
}

// /api/users/login POST 登入
func UserLoginHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UserLoginRequest
		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 確認 username 是否存在
		var existingUser models.User
		err := db.Preload("Role").Where("username = ?", req.Username).First(&existingUser).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "帳號或密碼錯誤",
			})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check username"})
			return
		}

		// 確認密碼是否正確
		if err := bcrypt.CompareHashAndPassword([]byte(existingUser.PasswordHash), []byte(req.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "帳號或密碼錯誤"})
			return
		}

		// 生成 JWT
		jwtToken, err := middleware.GenerateJWT(existingUser.ID, existingUser.Username, existingUser.Role.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "User login successfully!",
			"user":    existingUser.Username,
			"token":   jwtToken,
		})
	}
}

// /api/users/logout POST 登出
func UserLogoutHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {}
}

// /api/users/me GET 取得自己的資料
func GetUserProfileHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")
		username, _ := c.Get("username")
		role, _ := c.Get("role")

		c.JSON(http.StatusOK, gin.H{
			"user_id":  userID,
			"username": username,
			"role":     role,
		})
	}
}

// /api/users/:userId/submissions GET 取得指定使用者的提交紀錄
func GetUserSubmissionsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {}
}
