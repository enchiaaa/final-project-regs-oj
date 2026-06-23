package testutil

import (
	"online-judge/internal/models"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func CreateTestUser(t *testing.T, db *gorm.DB, username string) models.User {
	t.Helper()

	role := models.Role{}
	if err := db.Where("name = ?", "User").First(&role).Error; err != nil {
		t.Fatalf("failed to create test role: %v", err)
	}
	user := models.User{
		Username:     username,
		PasswordHash: "password",
		RoleID:       role.ID,
	}

	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	return user
}

func CreateTestProblem(t *testing.T, db *gorm.DB, problemCode string) models.Problem {
	t.Helper()

	problem := models.Problem{
		ProblemCode: problemCode,
		Title:       "Test Problem",
		LimitTime:   1000,
		ProblemPath: "test/path",
	}

	if err := db.Create(&problem).Error; err != nil {
		t.Fatalf("failed to create test problem: %v", err)
	}

	return problem
}

func CreateTestSubmission(t *testing.T, db *gorm.DB, userID uint, problemID uint, status string) models.Submission {
	t.Helper()

	submission := models.Submission{
		OperatorID: uuid.NewString(),
		UserID:     userID,
		ProblemID:  problemID,
		Status:     status,
	}

	if err := db.Create(&submission).Error; err != nil {
		t.Fatalf("failed to create test submission: %v", err)
	}

	return submission
}
