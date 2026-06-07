// 程式的進入點，負責初始化資料庫連線、啟動 Judge Goroutine，以及註冊 API 路由
package main

import (
	"online-judge/internal/api"
	"online-judge/internal/database"
	"online-judge/internal/judge"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化資料庫連線
	db := database.InitDB()

	// 啟動 Judge Goroutine（從任務隊列拿任務來評測）
	jobQueue := make(chan string, 100)
	go judge.StartWorker(db, jobQueue)

	// 註冊 API 路由
	router := gin.Default()
	api.RegisterRoutes(router, db, jobQueue)
	router.Run(":8080")
}
