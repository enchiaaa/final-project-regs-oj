// API routes 定義，將 URL 路徑對應到具體的 Handler 函數

package api

import (
	"online-judge/internal/rbac"
	"online-judge/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.Engine, db *gorm.DB, jobQueue chan string) {
	// 使用者相關
	router.POST("/api/users/register", UserRegisterHandler(db))
	router.POST("/api/users/login", UserLoginHandler(db))
	router.POST("/api/users/logout", middleware.AuthMiddleware(), UserLogoutHandler(db))
	router.GET("/api/users/me", middleware.AuthMiddleware(), middleware.RequirePermission(db, rbac.PermissionUserRead), GetUserProfileHandler(db))
	router.GET("/api/users/:userId/submissions", GetUserSubmissionsHandler(db))

	// 題目相關
	router.GET("/api/problems", GetAllProblemsHandler(db))
	router.GET("/api/problems/:problemId", GetProblemDetailHandler(db))
	router.PUT("/api/problems", middleware.AuthMiddleware(), middleware.RequirePermission(db, rbac.PermissionProblemUpsert), UpdateProblemHandler)
	router.DELETE("/api/problems/:problemId", middleware.AuthMiddleware(), middleware.RequirePermission(db, rbac.PermissionProblemDelete), DeleteProblemHandler)
	router.GET("/api/problems/:problemId/testcases", middleware.AuthMiddleware(), middleware.RequirePermission(db, rbac.PermissionProblemTestcasesRead), GetProblemTestCasesHandler)

	// 提交相關
	router.POST("/api/submissions", middleware.AuthMiddleware(), middleware.RequirePermission(db, rbac.PermissionSubmissionCreate), CreateSubmissionHandler(db, jobQueue))
	router.GET("/api/submissions/:operatorId/source", middleware.AuthMiddleware(), middleware.RequirePermission(db, rbac.PermissionSubmissionSourceRead), GetSubmissionSourceHandler(db))
	router.GET("/api/submissions/:operatorId/logs/:logType", middleware.AuthMiddleware(), middleware.RequirePermission(db, rbac.PermissionSubmissionLogRead), GetSubmissionLogHandler(db))
	router.GET("/api/submissions/:operatorId", middleware.AuthMiddleware(), middleware.RequirePermission(db, rbac.PermissionSubmissionRead), GetSubmissionResultHandler(db))
	router.GET("/api/submissions", middleware.AuthMiddleware(), middleware.RequirePermission(db, rbac.PermissionSubmissionRead), GetSubmissionsHandler(db))

	// 統計相關
	router.GET("/api/stats/problems/:problemId", GetProblemStatsHandler)
	router.GET("/api/stats/users/:userId", GetUserStatsHandler)
}
