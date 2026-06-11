// 負責初始化預設的 Role、Permission 與 Role-Permission 對應資料
package database

import (
	"online-judge/internal/models"

	"gorm.io/gorm"
)

func SeedRBAC(db *gorm.DB) error {
	// 建立 roles
	roleNames := []string{"Admin", "User"}
	roles := make(map[string]models.Role)

	for _, name := range roleNames {
		role := models.Role{Name: name}

		if err := db.FirstOrCreate(
			&role,
			models.Role{Name: name},
		).Error; err != nil {
			return err
		}

		roles[name] = role
	}

	// 建立 permissions
	permissionNames := []string{
		"problem:upsert",
		"problem:delete",
		"problem:read",
		"submission:self:create",
		"submission:self:read",
		"submission:self:source:read",
		"submission:any:read",
		"submission:any:source:read",
		"user:create",
		"user:self:read",
	}

	permissions := make(map[string]models.Permission)
	for _, name := range permissionNames {
		permission := models.Permission{Name: name}

		if err := db.FirstOrCreate(&permission, models.Permission{Name: name}).Error; err != nil {
			return err
		}

		permissions[name] = permission
	}

	// 建立關聯
	adminRole := roles["Admin"]
	if err := db.Model(&adminRole).Association("Permissions").Replace(
		permissions["submission:self:create"],
		permissions["submission:self:read"],
		permissions["submission:self:source:read"],
		permissions["problem:upsert"],
		permissions["problem:delete"],
		permissions["submission:any:read"],
		permissions["submission:any:source:read"],
		permissions["problem:read"],
		permissions["user:self:read"],
	); err != nil {
		return err
	}

	userRole := roles["User"]
	if err := db.Model(&userRole).Association("Permissions").Replace(
		permissions["submission:self:create"],
		permissions["submission:self:read"],
		permissions["submission:self:source:read"],
		permissions["user:self:read"],
		permissions["problem:read"],
	); err != nil {
		return err
	}

	return nil
}
