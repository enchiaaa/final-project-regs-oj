package database

import (
	"errors"
	"fmt"
	"online-judge/internal/models"
	"online-judge/internal/rbac"
	"os"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedAdmin(db *gorm.DB) error {
	username := os.Getenv("ADMIN_USERNAME")
	password := os.Getenv("ADMIN_PASSWORD")

	if username == "" && password == "" {
		return nil
	}

	if username == "" || password == "" {
		return errors.New("ADMIN_USERNAME and ADMIN_PASSWORD must both be set")
	}

	adminRole := models.Role{}
	if err := db.Where("name = ?", rbac.RoleAdmin).First(&adminRole).Error; err != nil {
		return fmt.Errorf("failed to find Admin role: %w", err)
	}

	existingUser := models.User{}
	err := db.Where("username = ?", username).First(&existingUser).Error
	if err == nil {
		if existingUser.RoleID != adminRole.ID {
			return fmt.Errorf("username %q already exists but is not an Admin", username)
		}
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to query admin user: %w", err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	admin := models.User{
		Username:     username,
		PasswordHash: string(passwordHash),
		RoleID:       adminRole.ID,
	}
	if err := db.Create(&admin).Error; err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	return nil
}
