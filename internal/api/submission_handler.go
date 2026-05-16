// submission api handler

package api

import (
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"online-judge/internal/models"
)

func CreateSubmissionHandler(db	*gorm.DB, jobQueue chan string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 取得上傳檔案
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get uploaded file"})
			return
		}

		// 生成唯一的 operatorId
		operatorId := uuid.New().String()

		// TODO: 從 Authorization header 取得 JWT，解析出 userId
		userId := 1

		// 將上傳的檔案儲存到指定路徑 "uploads/{userId}/{filename}/{operatorId}.ext"
		ext := strings.Split(file.Filename, ".")[1]
		dst := "uploads/" + strconv.Itoa(userId) + "/" + strings.Split(file.Filename, ".")[0]
		cmd := exec.Command("mkdir", "-p", dst)
		if err := cmd.Run(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
			return
		}
		if err := c.SaveUploadedFile(file, dst + "/" + operatorId + "." + ext); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save uploaded file"})
			return
		}
		
		// 寫入資料庫
		newSubmission := models.Submission{
			ID:         operatorId,
			UserID:     uint(userId),
			ProblemID:  c.PostForm("problemId"),
			Status:     "Pending",
			SourcePath: dst,
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

func GetSubmissionsHandler(c *gin.Context) {
}

func GetSubmissionResultHandler(c *gin.Context){
}

func GetSubmissionSourceHandler(c *gin.Context) {
}