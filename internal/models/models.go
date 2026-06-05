// 定義 User、Problem、Submission 三個主要的資料模型，使用 GORM 的 struct tags 來指定資料庫欄位的屬性和 JSON 序列化行為。
package models

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
	err = db.AutoMigrate(&User{}, &Problem{}, &Submission{})
	if err != nil {
		panic(fmt.Sprintf("failed to migrate database: %v", err))
	}

	fmt.Println("資料庫連線成功")

	return db
}

type User struct {
	gorm.Model
	Username  string    `gorm:"unique;not null" json:"username"`
	PasswordHash	string	`gorm:"not null" json:"-"`
	Role      string    `gorm:"default:'User'" json:"role"`
}

type Problem struct {
	gorm.Model
	ProblemCode  string		`gorm:"unique;not null" json:"problemCode"`
	LimitTime    int		`json:"limit_time"` // TLE 秒數
	ProblemPath  string		`json:"-"`
}

type Submission struct {
	gorm.Model
	OperatorID	string	`gorm:"unique;not null" json:"operatorId"`

	UserName	string	`json:"username"`
	ProblemCode	string	`json:"problemCode"`

	// 狀態判定：Pending, Configuring, Compiling, Judging, AC, WA, CE, RE, TLE, SE
	Status  string `gorm:"default:'Pending'" json:"status"`
	Message string `json:"message"`

	// 儲存路徑
	SourcePath       string `json:"-"` // 使用者上傳的原始 zip 位置
	WorkspacePath    string `json:"-"` // 解壓縮後的工作目錄
	ConfigureLogPath string `json:"-"` // 評測配置過程的 log 路徑
	CompileLogPath   string `json:"-"` // 編譯過程的 log 路徑
	OutputLogPath    string `json:"-"` // 執行結果的 log 路徑
}
