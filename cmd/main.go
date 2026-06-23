// 程式的進入點，負責初始化資料庫連線、啟動 Judge Goroutine，以及註冊 API 路由
package main

import (
	"errors"
	"log"
	"online-judge/internal/api"
	"online-judge/internal/database"
	"online-judge/internal/judge"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 載入 .env
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatalf("failed to load .env: %v", err)
	}

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
