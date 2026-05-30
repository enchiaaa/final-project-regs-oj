// 管理員相關的 API 處理函數

package api

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"online-judge/internal/models"
)

// /api/problems
func GetAllProblemsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var problems []models.Problem
		if err := db.Find(&problems).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "無法取得題目列表",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"problems": problems,
		})
	}
}

// /api/problems/:problemId
func GetProblemDetailHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("problemId")
		var problem models.Problem
		if err := db.First(&problem, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "題目不存在",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"problem": problem,
		})
	}
}

func UpdateProblemHandler(c *gin.Context) {
}

func DeleteProblemHandler(c *gin.Context) {
}

func GetProblemTestCasesHandler(c *gin.Context) {
}
