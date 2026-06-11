// 負責建立與管理資料庫連線
package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// 初始化資料庫連線，並根據模型自動建立資料表
func InitDB() *gorm.DB {
	db, err := gorm.Open(postgres.Open("host=localhost user=user password=123 dbname=OJ_db sslmode=disable"), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect database: %v", err))
	}

	// 根據模型建立資料表
	if err := Migrate(db); err != nil {
		panic(fmt.Sprintf("failed to migrate database: %v", err))
	}

	if err := SeedRBAC(db); err != nil {
		panic(fmt.Sprintf("failed to seed RBAC: %v", err))
	}

	fmt.Println("資料庫連線成功")

	return db
}
