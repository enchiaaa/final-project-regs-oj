// 負責執行資料表 migration，建立或更新資料庫 schema
package database

import (
	"online-judge/internal/models"

	"gorm.io/gorm"
)

// 根據模型建立資料表
func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(&models.User{}, &models.Role{}, &models.Permission{}, &models.Problem{}, &models.Submission{})
	return err
}
