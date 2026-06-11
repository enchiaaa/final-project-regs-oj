// submission api handler

package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

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

		// 解析出 userID
		user := models.User{}
		userID := c.GetUint("userID")
		if err := db.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
			return
		}

		problemCode := c.PostForm("problemCode")
		problem := models.Problem{}

		// 檢查 problemCode 是否存在
		if err := db.Where("problem_code = ?", problemCode).First(&problem).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Problem not found"})
			return
		}

		// 將上傳的檔案儲存到指定路徑 "uploads/{userName}/{problemCode}/{operatorId}.ext"
		ext := filepath.Ext(strings.ToLower(file.Filename))
		dst := filepath.Join("uploads", user.Username, problemCode, operatorId+ext)

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
			OperatorID: operatorId,
			UserID:     user.ID,
			ProblemID:  problem.ID,
			Status:     "Pending",
			SourcePath: dst,
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
		// 解析出 userID
		userID := c.GetUint("userID")
		submissions := []models.Submission{}

		if err := db.Where("user_id = ?", userID).Find(&submissions).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query submissions"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"submissions": submissions})
	}
}

// /api/submissions/:operatorId GET 查詢個人提交結果
func GetSubmissionResultHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		operatorId := c.Param("operatorId")
		submission := models.Submission{}
		if err := db.Preload("Problem").Preload("User").Where("operator_id = ?", operatorId).First(&submission).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Submission not found"})
			return
		}

		// 確認 ownership
		if !CanAccessUserResource(c, submission.UserID){
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"OperatorId":  submission.OperatorID,
			"Username":    submission.User.Username,
			"ProblemCode": submission.Problem.ProblemCode,
			"Status":      submission.Status,
			"Message":     submission.Message,
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

		// 確認 ownership
		if !CanAccessUserResource(c, submission.UserID){
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		c.File(submission.SourcePath)
	}
}

// /api/submissions/:operatorId/logs/:logType GET 回傳指定 log 內容
func GetSubmissionLogHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		operatorId := c.Param("operatorId")
		logType := c.Param("logType")

		submission := models.Submission{}
		if err := db.Where("operator_id = ?", operatorId).First(&submission).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Submission not found"})
			return
		}

		// 確認 ownership
		if !CanAccessUserResource(c, submission.UserID){
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		logPath := ""

		switch logType {
		case "configure":
			logPath = submission.ConfigureLogPath
		case "compile":
			logPath = submission.CompileLogPath
		case "output":
			logPath = submission.OutputLogPath
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid log type"})
			return
		}

		content, err := os.ReadFile(logPath)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "log file not found"})
			return
		}

		c.String(http.StatusOK, string(content))
	}
}
