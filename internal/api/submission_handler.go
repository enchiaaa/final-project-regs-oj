// submission api handler

package api

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"online-judge/internal/models"
)

// /api/submissions POST 上傳提交，並建立評測任務
func CreateSubmissionHandler(db *gorm.DB, jobQueue chan string) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		// 檢查上傳的檔案，並儲存到指定路徑 "uploads/{userName}/{problemCode}/{operatorId}.zip"
		dst := filepath.Join("uploads", user.Username, problemCode, operatorId + ".zip")
		if statusCode, err := SaveUploadedZipFile(c, "file", dst); err != nil{
			c.JSON(statusCode, gin.H{"error": err.Error()})
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
