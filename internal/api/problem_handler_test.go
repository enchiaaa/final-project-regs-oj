package api

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"online-judge/internal/models"
	"online-judge/internal/test_util"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Response struct {
	Message string `json:"message"`
}

const settingsYAML = `
title: Voyage Log
limits:
  totalTime: 1000
  cpuTime: 1000
  memory: 524288
`

const invalidSettingYAML = `
title: Voyage Log
limits: totalTime: 1000`

const problemCode = "114FinalQ006"

func setupProblemTestRouter(db *gorm.DB) *gin.Engine {
	router := gin.New()
	router.GET("/api/problems", GetAllProblemsHandler(db))
	router.GET("/api/problems/:problemId", GetProblemDetailHandler(db))
	router.PUT("/api/problems", UpsertProblemHandler(db))
	router.DELETE("/api/problems/:problemId", DeleteProblemHandler(db))
	router.GET("/api/problems/:problemId/testcases", GetProblemTestCasesHandler(db))
	return router
}

func newProblemUploadRequest(t *testing.T, uploadFilename string, content []byte) *http.Request {
	return newProblemUploadRequestWithCode(t, problemCode, uploadFilename, content)
}

func newProblemUploadRequestWithCode(t *testing.T, code string, uploadFilename string, content []byte) *http.Request {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if err := writer.WriteField("problemCode", code); err != nil {
		t.Fatalf("failed to write problemCode field: %v", err)
	}

	part, err := writer.CreateFormFile("file", uploadFilename)
	if err != nil {
		t.Fatalf("failed to create upload field: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("failed to write upload content: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/api/problems", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func newProblemZipUploadRequest(t *testing.T, filename string) *http.Request {
	t.Helper()

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("failed to read test ZIP %s: %v", filename, err)
	}

	return newProblemUploadRequest(t, filepath.Base(filename), content)
}

func TestGetProblemHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("題目列表正確回傳", func(t *testing.T) {
		// 1. 建立 Test DB
		testDB := testutil.SetupTestDB(t)

		// 2. 寫入資料
		newProblem := models.Problem{
			ProblemCode: problemCode,
			Title:       "Voyage Log",
			LimitTime:   1000,
			ProblemPath: "testfile/problem/114FinalQ006",
		}
		if err := testDB.Create(&newProblem).Error; err != nil {
			t.Error(err)
		}

		// 3. 設定測試 router
		router := setupProblemTestRouter(testDB)

		// 4. Call API
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/problems", nil)
		router.ServeHTTP(w, req)

		// 5. 確認結果
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d\nbody: %s", w.Code, w.Body.String())
		}

		var response ProblemListResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatal(err)
		}

		if response.Problems[0].ProblemCode != "114FinalQ006" {
			t.Fatalf("expected problemCode 114FinalQ006, got %s", response.Problems[0].ProblemCode)
		}

		if response.Problems[0].Title != "Voyage Log" {
			t.Fatalf("expected title 'Voyage Log', got %s", response.Problems[0].Title)
		}
	})
}
func TestGetProblemDetailHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Problem 不存在", func(t *testing.T) {
		// 1. 建立 Test DB
		testDB := testutil.SetupTestDB(t)

		// 2. 設定測試 router
		router := setupProblemTestRouter(testDB)

		// 3. Call API
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/problems/1", nil)
		router.ServeHTTP(w, req)

		// 4. 確認結果
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d\nbody: %s", w.Code, w.Body.String())
		}
	})
	t.Run("Problem 正確回傳", func(t *testing.T) {
		// 1. 建立 Test DB
		testDB := testutil.SetupTestDB(t)

		// 2. 寫入資料
		newProblem := models.Problem{
			ProblemCode: problemCode,
			Title:       "Voyage Log",
			LimitTime:   1000,
			ProblemPath: "testfile/problem/114FinalQ006",
		}
		if err := testDB.Create(&newProblem).Error; err != nil {
			t.Error(err)
		}

		// 3. 設定測試 router
		router := setupProblemTestRouter(testDB)

		// 4. Call API
		w := httptest.NewRecorder()

		url := fmt.Sprintf("/api/problems/%d", newProblem.ID)
		req := httptest.NewRequest(http.MethodGet, url, nil)
		router.ServeHTTP(w, req)

		// 5. 確認結果
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d\nbody: %s", w.Code, w.Body.String())
		}

		var response ProblemDetailResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatal(err)
		}

		if response.ProblemCode != problemCode {
			t.Fatalf("expected problemCode 114FinalQ006, got %s", response.ProblemCode)
		}

		if response.Title != "Voyage Log" {
			t.Fatalf("expected title Voyage Log, got %s", response.Title)
		}

		if response.LimitTime != 1000 {
			t.Fatalf("expected limitTime 1000, got %d", response.LimitTime)
		}
	})
}

func TestUpsertProblemHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("invalid problemCode", func(t *testing.T) {
		invalidCodes := []string{
			"",
			"../problem",
			`..\problem`,
			"problem/code",
			"problem.code",
			"problem code",
		}

		for _, code := range invalidCodes {
			t.Run(code, func(t *testing.T) {
				testDB := testutil.SetupTestDB(t)
				router := setupProblemTestRouter(testDB)

				w := httptest.NewRecorder()
				req := newProblemUploadRequestWithCode(t, code, "problem.zip", nil)
				router.ServeHTTP(w, req)

				if w.Code != http.StatusBadRequest {
					t.Fatalf("expected 400, got %d\nbody: %s", w.Code, w.Body.String())
				}
			})
		}
	})

	t.Run("Insert Problem", func(t *testing.T) {
		// 1. 建立 Test DB
		testDB := testutil.SetupTestDB(t)

		// 2. 設定測試 router
		router := setupProblemTestRouter(testDB)

		// 3. 建立上傳壓縮檔
		files := []File{
			{"template/", ""},
			{"solution/", ""},
			{"spec/", ""},
			{"online-judge/", ""},
			{"settings.yaml", settingsYAML},
		}
		filename := CreateZipFileForTest(t, problemCode, files)

		// 4. Call API 前，確認 Problem 114FinalQ006 不存在
		problem := models.Problem{}
		err := testDB.Where("problem_code = ?", problemCode).First(&problem).Error
		if err == nil {
			t.Fatalf("Problem existed before inserting")
		} else {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				t.Error(err)
			}
		}

		// 5. Call API
		w := httptest.NewRecorder()
		req := newProblemZipUploadRequest(t, filename)
		router.ServeHTTP(w, req)

		// 6. 確認結果
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d\nbody: %s", w.Code, w.Body.String())
		}

		problem = models.Problem{}
		err = testDB.Where("problem_code = ?", problemCode).First(&problem).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				t.Fatalf("Problem didn't exist after inserting")
			} else {
				t.Error(err)
			}
		}
	})

	t.Run("Update Problem", func(t *testing.T) {
		// 1. 建立 Test DB
		testDB := testutil.SetupTestDB(t)

		// 2. 寫入資料
		newProblem := models.Problem{
			ProblemCode: problemCode,
			Title:       "Voyage Log",
			LimitTime:   0,
			ProblemPath: "testfile/problem/114FinalQ006",
		}
		if err := testDB.Create(&newProblem).Error; err != nil {
			t.Error(err)
		}

		// 3. 建立上傳壓縮檔
		files := []File{
			{"template/new.txt", ""},
			{"solution/", ""},
			{"spec/", ""},
			{"online-judge/", ""},
			{"settings.yaml", settingsYAML},
		}
		filename := CreateZipFileForTest(t, problemCode, files)

		// 4. 建立 Problem 資料夾，用 old.txt 標記為舊題目
		problemPath := "testfile/problem/" + problemCode
		err := os.MkdirAll(filepath.Join(problemPath), 0755)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(filepath.Join(problemPath, "old.txt"), []byte("old content"), 0644)
		if err != nil {
			t.Fatal(err)
		}

		// 5. 設定測試 router
		router := setupProblemTestRouter(testDB)

		// 6. Call API
		w := httptest.NewRecorder()
		req := newProblemZipUploadRequest(t, filename)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d\nbody: %s", w.Code, w.Body.String())
		}

		// 7. 確認結果：資料庫以及資料夾更新成功
		problem := models.Problem{}
		err = testDB.Where("problem_code = ?", problemCode).First(&problem).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				t.Fatalf("Problem didn't exist")
			} else {
				t.Error(err)
			}
		}
		if problem.LimitTime == 0 || problem.LimitTime != 1000 {
			t.Fatalf("Failed to update Problem")
		}

		// 確認資料夾更新
		if _, err := os.Stat(filepath.Join(problemPath, "old.txt")); !os.IsNotExist(err) {
			t.Fatalf("old file should be removed after update")
		}
		if _, err := os.Stat(filepath.Join(problemPath, "template/new.txt")); os.IsNotExist(err) {
			t.Fatalf("new file should exist after update")
		}
	})

	t.Run("測試缺少任一資料", func(t *testing.T) {
		// 1. 建立 Test DB
		testDB := testutil.SetupTestDB(t)

		// 2. 設定測試 router
		router := setupProblemTestRouter(testDB)

		// 3. 測試每一種資料不存在的情況
		files := []File{
			{"template/", ""},
			{"solution/", ""},
			{"spec/", ""},
			{"online-judge/", ""},
			{"settings.yaml", settingsYAML},
		}

		for idx := range files {
			// 4. 建立上傳壓縮檔
			tmpFiles := []File{}
			for _idx, file := range files {
				if idx == _idx {
					continue
				}
				tmpFiles = append(tmpFiles, file)
			}
			filename := CreateZipFileForTest(t, problemCode, tmpFiles)

			// 5. Call API
			w := httptest.NewRecorder()
			req := newProblemZipUploadRequest(t, filename)
			router.ServeHTTP(w, req)

			// 6. 確認結果
			if w.Code != http.StatusInternalServerError {
				t.Fatalf("expected 500, got %d\nbody: %s", w.Code, w.Body.String())
			}
		}
	})

	t.Run("測試上傳的 settings.yaml 格式錯誤", func(t *testing.T) {
		// 1. 建立 Test DB
		testDB := testutil.SetupTestDB(t)

		// 2. 設定測試 router
		router := setupProblemTestRouter(testDB)

		// 3. 建立上傳壓縮檔
		files := []File{
			{"template/", ""},
			{"solution/", ""},
			{"spec/", ""},
			{"online-judge/", ""},
			{"settings.yaml", invalidSettingYAML},
		}
		filename := CreateZipFileForTest(t, problemCode, files)

		// 4. Call API
		w := httptest.NewRecorder()
		req := newProblemZipUploadRequest(t, filename)
		router.ServeHTTP(w, req)

		// 5. 確認結果
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d\nbody: %s", w.Code, w.Body.String())
		}
	})

	t.Run("測試非 zip 檔案上傳", func(t *testing.T) {
		// 1. 建立 Test DB
		testDB := testutil.SetupTestDB(t)

		// 2. 設定測試 router
		router := setupProblemTestRouter(testDB)

		// 3. Call API
		w := httptest.NewRecorder()
		req := newProblemUploadRequest(t, "test.txt", nil)
		router.ServeHTTP(w, req)

		// 4. 確認結果
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d\nbody: %s", w.Code, w.Body.String())
		}
	})

	t.Run("測試壞掉的 zip 檔案上傳", func(t *testing.T) {
		// 1. 建立 Test DB
		testDB := testutil.SetupTestDB(t)

		// 2. 設定測試 router
		router := setupProblemTestRouter(testDB)

		// 3. Call API
		w := httptest.NewRecorder()
		req := newProblemUploadRequest(t, "broken.zip", nil)
		router.ServeHTTP(w, req)

		// 4. 確認結果
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d\nbody: %s", w.Code, w.Body.String())
		}
	})

}

func TestDeleteProblemHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Delete Problem", func(t *testing.T) {
		// 1. 建立 Test DB
		testDB := testutil.SetupTestDB(t)

		// 2. 設定測試 router
		router := setupProblemTestRouter(testDB)

		// 3. 寫入資料
		newProblem := models.Problem{
			ProblemCode: problemCode,
			Title:       "Voyage Log",
			LimitTime:   0,
			ProblemPath: "testfile/problem/114FinalQ006",
		}
		if err := testDB.Create(&newProblem).Error; err != nil {
			t.Error(err)
		}

		// 5. Call API 前，確認 Problem 114FinalQ006 存在
		problem := models.Problem{}
		err := testDB.Where("id = ?", newProblem.ID).First(&problem).Error
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				t.Error(err)
			} else {
				t.Fatalf("Problem didn't exist before deleting")
			}
		}

		// 6. Call API
		w := httptest.NewRecorder()
		url := fmt.Sprintf("/api/problems/%d", newProblem.ID)
		req := httptest.NewRequest(http.MethodDelete, url, nil)
		router.ServeHTTP(w, req)

		// 7. 確認結果
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d\nbody: %s", w.Code, w.Body.String())
		}

		problem = models.Problem{}
		err = testDB.Where("id = ?", newProblem.ID).First(&problem).Error
		if err == nil {
			t.Fatalf("Problem still exist after deleting")
		} else {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				t.Error(err)
			}
		}
	})

	t.Run("Problem 不存在", func(t *testing.T) {
		// 1. 建立 Test DB
		testDB := testutil.SetupTestDB(t)

		// 2. 設定測試 router
		router := setupProblemTestRouter(testDB)

		// 3. Call API
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/api/problems/1", nil)
		router.ServeHTTP(w, req)

		// 4. 確認結果
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d\nbody: %s", w.Code, w.Body.String())
		}
	})
}

func TestGetProblemTestCasesHandler(t *testing.T) {
	t.Run("成功下載題目 ZIP", func(t *testing.T) {
		// 1. 建立 Test DB
		testDB := testutil.SetupTestDB(t)

		// 2. 建立 Problem
		problem := testutil.CreateTestProblem(t, testDB, "test-problem")

		// 3. 建立題目資料夾與測試檔案
		problemRoot := problem.ProblemPath
		archivePath := filepath.Join(problemRoot, "testProblem.txt")
		archive, err := os.Create(archivePath)
		if err != nil {
			t.Fatal(err)
		}
		defer archive.Close()

		// 4. 建立 router
		router := gin.New()
		router.GET("/api/problems/:problemId/testcases", GetProblemTestCasesHandler(testDB))

		// 5. 呼叫 API
		w := httptest.NewRecorder()
		url := fmt.Sprintf("/api/problems/%d/testcases", problem.ID)
		req := httptest.NewRequest(http.MethodGet, url, nil)
		router.ServeHTTP(w, req)

		// 6. 確認 HTTP 200
		if w.Code != http.StatusOK {
			t.Fatalf("expected HTTP 200, got %d\nbody: %s", w.Code, w.Body.String())
		}

		// 7. 將 response body 當作 ZIP 開啟
		reader, err := zip.NewReader(
			bytes.NewReader(w.Body.Bytes()),
			int64(w.Body.Len()),
		)
		if err != nil {
			t.Fatalf("response is not a valid ZIP: %v", err)
		}

		// 8. 確認 ZIP 內有預期檔案
		found := false
		for _, file := range reader.File {
			if file.Name == "testProblem.txt" {
				found = true
				break
			}
		}
		if !found {
			t.Fatal("expected testProblem.txt in downloaded ZIP")
		}
	})

	t.Run("題目不存在", func(t *testing.T) {
		// 1. 建立 Test DB
		testDB := testutil.SetupTestDB(t)

		// 2. 建立 router
		router := gin.New()
		router.GET("/api/problems/:problemId/testcases", GetProblemTestCasesHandler(testDB))

		// 5. 呼叫 API
		w := httptest.NewRecorder()
		url := "/api/problems/1/testcases"
		req := httptest.NewRequest(http.MethodGet, url, nil)
		router.ServeHTTP(w, req)

		// 6. 確認 HTTP 200
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected HTTP 404, got %d\nbody: %s", w.Code, w.Body.String())
		}
	})
}

type File struct {
	Name string
	Body string
}

// 將傳入的檔案壓縮成一個 ZIP 檔
func CreateZipFileForTest(t *testing.T, filename string, files []File) string {
	t.Helper()

	archivePath := filepath.Join(t.TempDir(), filename+".zip")
	archive, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()

	// Create a new zip archive.
	zipWriter := zip.NewWriter(archive)

	for _, file := range files {
		f, err := zipWriter.Create(file.Name)
		if err != nil {
			t.Error(err)
		}
		_, err = f.Write([]byte(file.Body))
		if err != nil {
			t.Error(err)
		}
	}

	// Make sure to check the error on Close.
	err = zipWriter.Close()
	if err != nil {
		t.Fatal(err)
	}

	return archivePath
}
