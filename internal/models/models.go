// 定義資料表
package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username     string `gorm:"unique;not null" json:"username"`
	PasswordHash string `gorm:"not null" json:"-"`

	// 外鍵與關聯的物件
	RoleID uint `gorm:"not null" json:"role_id"`
	Role   Role `json:"role,omitempty"`
}

type Role struct {
	gorm.Model
	Name        string       `gorm:"unique;not null" json:"name"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
}

type Permission struct {
	gorm.Model
	Name  string `gorm:"unique;not null" json:"name"`
	Roles []Role `gorm:"many2many:role_permissions;" json:"roles,omitempty"`
}

type Problem struct {
	gorm.Model
	ProblemCode string `gorm:"unique;not null" json:"problemCode"`
	Title		string `gorm:"title"`
	LimitTime   int    `json:"limit_time"` // TLE 秒數
	ProblemPath string `json:"-"`
}

type Submission struct {
	gorm.Model
	OperatorID string `gorm:"unique;not null" json:"operatorId"`

	// 外鍵與關聯的物件
	UserID uint `gorm:"not null" json:"user_id"`
	User   User `json:"user,omitempty"`

	// 外鍵與關聯的物件
	ProblemID uint    `gorm:"not null" json:"problem_id"`
	Problem   Problem `json:"problem,omitempty"`

	Status  string `gorm:"default:'Pending'" json:"status"`
	Message string `json:"message"`

	// 儲存路徑
	SourcePath       string `json:"-"` // 使用者上傳的原始 zip 位置
	WorkspacePath    string `json:"-"` // 解壓縮後的工作目錄
	ConfigureLogPath string `json:"-"` // 評測配置過程的 log 路徑
	CompileLogPath   string `json:"-"` // 編譯過程的 log 路徑
	OutputLogPath    string `json:"-"` // 執行結果的 log 路徑
}
