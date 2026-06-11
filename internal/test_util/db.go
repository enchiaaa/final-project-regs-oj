// 執行測試時需要的 test DB

package testutil

import (
	"online-judge/internal/database"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func SetupTestDB(t *testing.T) *gorm.DB {
    t.Helper()	// 用來標註這是一個 helper function

	db, err := gorm.Open(postgres.Open("host=localhost user=user password=123 dbname=OJ_test_db sslmode=disable"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

    if err := database.Migrate(db); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	if err := database.SeedRBAC(db); err != nil {
		t.Fatalf("failed to seed RBAC: %v", err)
	}

    return db
}