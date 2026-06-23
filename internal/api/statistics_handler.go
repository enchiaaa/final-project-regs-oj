// statistics api handler

package api

import (
	"net/http"
	"online-judge/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StatusCount struct {
	Status string
	Count  int64
}

func statusRowsToMap(rows []StatusCount) map[string]int64 {
	counts := map[string]int64{}
	for _, row := range rows {
		counts[row.Status] = row.Count
	}
	return counts
}

// /api/stats/problems/:problemId
func GetProblemStatsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 檢查 problem 是否存在
		id := c.Param("problemId")
		var problem models.Problem
		if err := db.First(&problem, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "題目不存在"})
			return
		}

		// 2. 計算所有回傳資料
		// 提交次數
		var totalSubmissions int64
		if err := db.Model(&models.Submission{}).
			Where("problem_id = ?", problem.ID).
			Count(&totalSubmissions).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query problem stats"})
			return
		}

		// 每個 status 統計
		var rows []StatusCount
		if err := db.Model(&models.Submission{}).
			Select("status, COUNT(*) AS count").
			Where("problem_id = ?", problem.ID).
			Group("status").
			Scan(&rows).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query problem stats"})
			return
		}

		// 3. 回傳資料
		c.JSON(http.StatusOK, gin.H{
			"problemId":        problem.ID,
			"problemCode":      problem.ProblemCode,
			"title":            problem.Title,
			"totalSubmissions": totalSubmissions,
			"statusCounts":     statusRowsToMap(rows),
		})
	}
}

// /api/stats/users/:userId
func GetUserStatsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 檢查 user 是否存在
		id := c.Param("userId")

		var user models.User
		if err := db.First(&user, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用戶不存在"})
			return
		}

		// 2. 計算所有要回傳的資料
		// 總提交次數
		var totalSubmissions int64
		if err := db.Model(&models.Submission{}).
			Where("user_id = ?", user.ID).
			Count(&totalSubmissions).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query user stats"})
			return
		}

		// 每個 status 統計
		var rows []StatusCount
		if err := db.Model(&models.Submission{}).
			Select("status, COUNT(*) AS count").
			Where("user_id = ?", user.ID).
			Group("status").
			Scan(&rows).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query user stats"})
			return
		}

		// 3. 回傳資料
		c.JSON(http.StatusOK, gin.H{
			"userId":           user.ID,
			"username":         user.Username,
			"totalSubmissions": totalSubmissions,
			"statusCounts":     statusRowsToMap(rows),
		})
	}
}
