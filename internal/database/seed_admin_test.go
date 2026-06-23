package database_test

import (
	"errors"
	"testing"

	"online-judge/internal/database"
	"online-judge/internal/models"
	"online-judge/internal/rbac"
	"online-judge/internal/test_util"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	testAdminUsername = "admin"
	testAdminPassword = "password"
)

func TestSeedAdmin(t *testing.T) {
	t.Run("成功建立 Admin", func(t *testing.T) {
		// 1. 建立 Test DB
		testDB := testutil.SetupTestDB(t)

		// 2. 設定 Admin 帳號與密碼
		t.Setenv("ADMIN_USERNAME", testAdminUsername)
		t.Setenv("ADMIN_PASSWORD", testAdminPassword)

		// 3. 呼叫 SeedAdmin
		if err := database.SeedAdmin(testDB); err != nil {
			t.Fatalf("failed to seed admin: %v", err)
		}

		// 4. 查詢 Admin
		user := models.User{}
		if err := testDB.Preload("Role").Where("username = ?", testAdminUsername).First(&user).Error; err != nil {
			t.Fatalf("failed to find seeded admin: %v", err)
		}

		// 5. 確認 Role 與密碼
		if user.Role.Name != rbac.RoleAdmin {
			t.Fatalf("expected role %s, got %s", rbac.RoleAdmin, user.Role.Name)
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(testAdminPassword)); err != nil {
			t.Fatalf("seeded admin password does not match: %v", err)
		}
	})

	t.Run("重複呼叫不會建立第二筆 Admin", func(t *testing.T) {
		// 1. 建立 Test DB
		testDB := testutil.SetupTestDB(t)

		// 2. 設定 Admin 帳號與密碼
		t.Setenv("ADMIN_USERNAME", testAdminUsername)
		t.Setenv("ADMIN_PASSWORD", testAdminPassword)

		// 3. 重複呼叫 SeedAdmin
		if err := database.SeedAdmin(testDB); err != nil {
			t.Fatalf("first SeedAdmin call failed: %v", err)
		}
		if err := database.SeedAdmin(testDB); err != nil {
			t.Fatalf("second SeedAdmin call failed: %v", err)
		}

		// 4. 確認只有一筆 username
		var count int64
		if err := testDB.Model(&models.User{}).Where("username = ?", testAdminUsername).Count(&count).Error; err != nil {
			t.Fatalf("failed to count admin users: %v", err)
		}
		if count != 1 {
			t.Fatalf("expected 1 admin user, got %d", count)
		}
	})

	t.Run("只設定一個環境變數會失敗", func(t *testing.T) {
		// 1. 建立 Test DB
		testDB := testutil.SetupTestDB(t)

		// 2. 只設定 username
		t.Setenv("ADMIN_USERNAME", testAdminUsername)
		t.Setenv("ADMIN_PASSWORD", "")

		// 3. 預期 SeedAdmin 回傳錯誤
		if err := database.SeedAdmin(testDB); err == nil {
			t.Fatal("expected error when ADMIN_PASSWORD is not set")
		}
	})

	t.Run("兩個環境變數皆未設定時略過建立", func(t *testing.T) {
		// 1. 建立 Test DB
		testDB := testutil.SetupTestDB(t)

		// 2. 清空兩個環境變數
		t.Setenv("ADMIN_USERNAME", "")
		t.Setenv("ADMIN_PASSWORD", "")

		// 3. 呼叫 SeedAdmin
		if err := database.SeedAdmin(testDB); err != nil {
			t.Fatalf("expected no error when admin env is not set, got %v", err)
		}

		// 4. 確認沒有建立 Admin
		user := models.User{}
		err := testDB.Where("username = ?", testAdminUsername).First(&user).Error
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("expected admin not to be created, got error %v", err)
		}
	})
}
