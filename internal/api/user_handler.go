// 使用者相關的 API 處理函數
package api

import (
	"errors"
	"net/http"
	"online-judge/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	// "net/http"
)

type UserRegisterRequest struct {
	Username  string	`json:"username" binding:"required"`
	Password  string	`json:"password" binding:"required, min=8"`
}
type UserLoginRequest struct {
	Username  string	`json:"username" binding:"required"`
	Password  string	`json:"password" binding:"required"`
}

func UserRegisterHandler(db *gorm.DB) gin.HandlerFunc{
	return func(c *gin.Context) {
		var req UserRegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
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
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}
		
		newUser := models.User{
			Username: req.Username,
			PasswordHash: string(passwordHash),
			Role: "user",
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

func UserLoginHandler(db *gorm.DB) gin.HandlerFunc{
	return func(c *gin.Context) {
		var req UserLoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 確認 username 是否存在
		var existingUser models.User
		err := db.Where("username = ?", req.Username).First(&existingUser).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "帳號或密碼錯誤",
			})
			return
		}
		if err != nil  {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check username"})
			return
		}

		// 確認密碼是否正確
		if err := bcrypt.CompareHashAndPassword([]byte(existingUser.PasswordHash), []byte(req.Password)); err != nil{
			c.JSON(http.StatusUnauthorized, gin.H{"error": "帳號或密碼錯誤"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "User login successfully!",
			"user":	existingUser.Username,
			"token" : "jwt",
		})
	}
}

func UserLogoutHandler(db *gorm.DB) gin.HandlerFunc{
	return func(c *gin.Context) {}
}

func GetUserProfileHandler(db *gorm.DB) gin.HandlerFunc{
	return func(c *gin.Context) {}
}

func GetUserSubmissionsHandler(db *gorm.DB) gin.HandlerFunc{
	return func(c *gin.Context) {}
}
