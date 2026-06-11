// 負責初始化預設的 Role、Permission 與 Role-Permission 對應資料
package database

import (
	"online-judge/internal/rbac"
	"online-judge/internal/models"

	"gorm.io/gorm"
)

func SeedRBAC(db *gorm.DB) error {
	// 列出所有 role
	roleNames := []string{
		rbac.RoleAdmin,
		rbac.RoleUser,
	}

	// 建立所有 Role Model
	roles := make(map[string]models.Role)
	for _, name := range roleNames {
		role := models.Role{Name: name}

		if err := db.FirstOrCreate(&role, models.Role{Name: name}).Error; err != nil {
			return err
		}

		roles[name] = role
	}

	// 建立所有 Permission Model
	permissions := make(map[string]models.Permission)
	for _, name := range rbac.AllPermissions {
		permission := models.Permission{Name: name}

		if err := db.FirstOrCreate(&permission, models.Permission{Name: name}).Error; err != nil {
			return err
		}

		permissions[name] = permission
	}

	// 根據 rbac.RolePermissions 建立 Role 和 Permission 之間的關聯
	for roleName, permissionNames := range rbac.RolePermissions {
		role := roles[roleName]

		rolePermissions := []models.Permission{}
		for _, permissionName := range permissionNames {
			rolePermissions = append(rolePermissions, permissions[permissionName])
		}

		if err := db.Model(&role).Association("Permissions").Replace(rolePermissions); err != nil {
			return err
		}
	}

	return nil
}
