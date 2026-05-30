// 定義 User、Problem、Submission 三個主要的資料模型，使用 GORM 的 struct tags 來指定資料庫欄位的屬性和 JSON 序列化行為。
package models

import (
	"fmt"
	"time"

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
	err = db.AutoMigrate(&User{}, &Problem{}, &Submission{})
	if err != nil {
		panic(fmt.Sprintf("failed to migrate database: %v", err))
	}

	fmt.Println("資料庫連線成功")

	return db
}

type User struct {
	gorm.Model
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"unique;not null" json:"username"`
	Password  string    `gorm:"not null" json:"-"`
	Role      string    `gorm:"default:'Guest'" json:"role"` // Guest, User, Admin
	CreatedAt time.Time `json:"createdAt"`
}

type Problem struct {
	gorm.Model
	ID           string    `gorm:"primaryKey" json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	LimitTime    int       `json:"limit_time"` // TLE 秒數
	TestCasePath string    `json:"-"`          // 儲存伺服器上的測資路徑
	CreatedAt    time.Time `json:"createdAt"`
}

type Submission struct {
	gorm.Model
	ID         string `gorm:"primaryKey" json:"id"`
	OperatorID string `json:"operatorId"`
	UserID     uint   `json:"userId"`
	ProblemID  string `json:"problemId"`

	// 狀態判定：Pending, Configuring, Compiling, Judging, AC, WA, CE, RE, TLE, SE
	Status string `gorm:"default:'Pending'" json:"status"`

	// 儲存路徑
	SourcePath       string `json:"-"` // 使用者上傳的原始 zip 位置
	ConfigureLogPath string `json:"-"` // 評測配置過程的 log 路徑
	CompileLogPath   string `json:"-"` // 編譯過程的 log 路徑
	OutputLogPath    string `json:"-"` // 執行結果的 log 路徑

	// 評測數據
	ExitCode    int     `json:"exitCode"`
	ExecuteTime float64 `json:"executeTime"` // 單位：秒，用於 TLE 判定

	SubmittedAt time.Time  `json:"submittedAt"`
	FinishedAt  *time.Time `json:"finishedAt"` // 評測完成時間
}
