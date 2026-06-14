// 管理員相關的 API 處理函數

package api

import (
	"errors"
	"net/http"
	"online-judge/internal/models"
	"online-judge/internal/utils"
	"os"

	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProblemListItemResponse struct {
	ProblemCode string	`json:"problemCode"`
	Title		string	`json:"title"`
}

type ProblemListResponse struct {
	Problems []ProblemListItemResponse	`json:"problems"`
}

type ProblemDetailResponse struct {
	ProblemCode string	`json:"problemCode"`
	Title		string	`json:"title"`
	LimitTime   int		`json:"limitTime"`
}

// /api/problems
func GetAllProblemsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var problems []models.Problem
		if err := db.Find(&problems).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "無法取得題目列表",
			})
			return
		}

		responses := ProblemListResponse{}

		for _, problem := range problems {
			responses.Problems = append(responses.Problems, ProblemListItemResponse{
				ProblemCode: problem.ProblemCode,
				Title: problem.Title,
			})
		}

		c.JSON(http.StatusOK, responses)
	}
}

// /api/problems/:problemCode
func GetProblemDetailHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("problemCode")
		var problem models.Problem
		if err := db.First(&problem, "problem_code = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "題目不存在",
			})
			return
		}

		c.JSON(http.StatusOK,
			ProblemDetailResponse{
				ProblemCode: problem.ProblemCode,
				Title: problem.Title,
				LimitTime: problem.LimitTime,
			},
		)
	}
}

// /api/problems
func UpsertProblemHandler(db *gorm.DB) gin.HandlerFunc{
	return func(c *gin.Context) {
		// TODO: problemCode 驗證：不能是空字串、只能有英數字 or - or _、不允許 /、\、.、空白
		problemCode := c.PostForm("problemCode")
		operationId := uuid.New().String()

		sourceDst := filepath.Join("tmp", "problem-staging", problemCode, operationId, "source.zip")
		extractedDst := filepath.Join("tmp", "problem-staging", problemCode, operationId, "extracted")
		backupPath := filepath.Join("tmp", "problem-staging", problemCode, operationId, "backup")

		problemExists := false
		hasBackup := false

		// 1. 檢查上傳的檔案，並儲存到指定路徑 "tmp/problem-staging/{problemCode}/{operationId}/source.zip"
		if statusCode, err := SaveUploadedZipFile(c, "file", sourceDst); err != nil{
			c.JSON(statusCode, gin.H{"error": err.Error()})
			return
		}

		// 2. unzip 到暫存位置 "tmp/problem-staging/{problemCode}/{operationId}/extracted/"
		if err := utils.UnzipFile(sourceDst, extractedDst); err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 3. 檢查 file 結構
		necessaryDirs := []string{"template/", "solution/", "spec/", "online-judge/"}
		for _, necessaryDir := range necessaryDirs{
			dirPath := filepath.Join(extractedDst, necessaryDir)
			_, err := os.ReadDir(dirPath)
			if err != nil{
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
		if _, err := os.ReadFile(filepath.Join(extractedDst, "settings.yaml")); err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 4. 讀 setting.yaml 取得題目設定的 limit time
		yamlPath := filepath.Join(extractedDst, "settings.yaml")
		title, totalTime, err := ParseYAML(yamlPath)
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 5. 若題目存在，先備份
		problem := models.Problem{}
		problemPath := ""
		err = db.Where("problem_code = ?", problemCode).First(&problem).Error
		if err != nil{ 
			if errors.Is(err, gorm.ErrRecordNotFound) {
        		// 題目不存在
				problemExists = false
				problemPath = filepath.Join("problem", problemCode)
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query problem"})
				return
			}
		} else {
			// 題目存在
			problemExists = true
			problemPath = problem.ProblemPath
			srcDir := problem.ProblemPath
			dstDir := backupPath
			if err := os.CopyFS(dstDir, os.DirFS(srcDir)); err != nil{
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to copy directory"})
				return
			}
			hasBackup = true
		}

		// 6. 開啟 DB transaction
		// 6.1 Upsert DB Problem
		// 6.2 更新 problem 資料夾內容
		// 6.3 若成功，commit DB transaction；若失敗，rollback
		err = db.Transaction(func(tx *gorm.DB) error {
			if !problemExists{
				// insert
				newProblem := models.Problem{
					ProblemCode: problemCode,
					Title: title,
					LimitTime: totalTime,
					ProblemPath: problemPath,
				}
				if err := tx.Create(&newProblem).Error; err != nil{
					return err
				}
			} else {
				// update
				problem.Title = title
				problem.LimitTime = totalTime
				problem.ProblemPath = problemPath
				if err := tx.Save(&problem).Error; err != nil{
					return err
				}
			}
			
			// 清除 problemPath
			if err := os.RemoveAll(problemPath); err != nil {
				return err
			}

			// 將新 problem 內容複製到 problemPath
			if err := os.CopyFS(problemPath, os.DirFS(extractedDst)); err != nil {
				if hasBackup{
					// 若複製失敗，將備份內容還原
					if err := os.RemoveAll(problemPath); err != nil {
						return err
					}
					if err := os.CopyFS(problemPath, os.DirFS(backupPath)); err != nil {
						return err
					}
				}
				return err
			}

			// return nil will commit the whole transaction
			return nil
		})

		if err != nil{
			c.JSON(http.StatusInternalServerError, "Failed to upsert problem")
			return
		}

		c.JSON(http.StatusOK, "ok")
	}
}

func DeleteProblemHandler(c *gin.Context) {
}

func GetProblemTestCasesHandler(c *gin.Context) {
}
