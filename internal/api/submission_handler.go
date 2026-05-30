// submission api handler

package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"online-judge/internal/models"
)

// /api/submissions POST 上傳提交，並建立評測任務
func CreateSubmissionHandler(db *gorm.DB, jobQueue chan string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 取得上傳檔案，並檢查是否為 zip 格式
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get uploaded file"})
			return
		}
		if filepath.Ext(strings.ToLower(file.Filename)) != ".zip" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Only zip files are allowed"})
			return
		}

		// 生成唯一的 operatorId
		operatorId := uuid.New().String()

		// TODO: 從 Authorization header 取得 JWT，解析出 userId
		userId := 2
		problemId := c.PostForm("problemId")

		// 檢查 problemId 是否存在
		if db.Where("id = ?", problemId).First(&models.Problem{}).Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Problem not found"})
			return
		}

		// 將上傳的檔案儲存到指定路徑 "uploads/{userId}/{problemId}/{operatorId}.ext"
		ext := filepath.Ext(strings.ToLower(file.Filename))
		dst := filepath.Join("uploads", strconv.Itoa(userId), problemId, operatorId+ext)

		// 0755 是 Linux 的檔案權限，表示擁有者有讀寫執行權限，群組和其他人只有讀取和執行權限
		cmd := os.MkdirAll(filepath.Dir(dst), 0755)
		if cmd != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
			return
		}
		if err := c.SaveUploadedFile(file, dst); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save uploaded file"})
			return
		}

		// 寫入資料庫
		newSubmission := models.Submission{
			ID:          operatorId,
			OperatorID:  operatorId,
			UserID:      uint(userId),
			ProblemID:   problemId,
			Status:      "Pending",
			SourcePath:  dst,
			SubmittedAt: time.Now(),
		}
		if err := db.Create(&newSubmission).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create submission record"})
			return
		}

		jobQueue <- operatorId

		c.JSON(http.StatusOK, gin.H{"OperatorId": operatorId})
	}
}

// /api/submissions GET 查詢個人提交紀錄
func GetSubmissionsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := 2 // TODO: 從 JWT 解析出 userId
		submissions := []models.Submission{}
		if err := db.Where("user_id = ?", userId).Find(&submissions).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query submissions"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"submissions": submissions})
	}
}

// /api/submissions/:operatorId GET 查詢提交結果
func GetSubmissionResultHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		operatorId := c.Param("operatorId")
		submission := models.Submission{}
		if err := db.Where("operator_id = ?", operatorId).First(&submission).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Submission not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"OperatorId":  submission.OperatorID,
			"ProblemId":   submission.ProblemID,
			"Status":      submission.Status,
			"SubmittedAt": submission.SubmittedAt,
			"FinishedAt":  submission.FinishedAt,
		})
	}
}

// /api/submissions/:operatorId/source GET 回傳提交的原始碼（zip）
func GetSubmissionSourceHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		operatorId := c.Param("operatorId")
		submission := models.Submission{}
		if err := db.Where("operator_id = ?", operatorId).First(&submission).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Submission not found"})
			return
		}
		c.File(submission.SourcePath)
	}
}
