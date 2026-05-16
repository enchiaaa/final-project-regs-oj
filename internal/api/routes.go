// API routes 定義，將 URL 路徑對應到具體的 Handler 函數

package api

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.Engine, db *gorm.DB, jobQueue chan string) {
	// 使用者相關
	router.POST("/api/users/register", UserRegisterHandler)
	router.POST("/api/users/login", UserLoginHandler)
	router.POST("/api/users/logout", UserLogoutHandler)
	router.GET("/api/users/me", GetUserProfileHandler)
	router.GET("/api/users/{userId}/submissions", GetUserSubmissionsHandler)

	// 題目相關
	router.GET("/api/problems", GetAllProblemsHandler)
	router.GET("/api/problems/{problemId}", GetProblemDetailHandler)
	router.PUT("/api/problems", UpdateProblemHandler)
	router.DELETE("/api/problems/{problemId}", DeleteProblemHandler)
	router.GET("/api/problems/{problemId}/testcases", GetProblemTestCasesHandler)

	// 提交相關
	router.POST("/api/submissions", CreateSubmissionHandler(db, jobQueue))
	router.GET("/api/submissions", GetSubmissionsHandler)
	router.GET("/api/submissions/{operatorId}", GetSubmissionResultHandler)
	router.GET("/api/submissions/{operatorId}/source", GetSubmissionSourceHandler)

	// 統計相關
	router.GET("/api/stats/problems/{problemId}", GetProblemStatsHandler)
	router.GET("/api/stats/users/{userId}", GetUserStatsHandler)
}