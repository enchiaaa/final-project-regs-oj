package api

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"io"
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

var settingsYAML = `
title: Voyage Log
limits:
  totalTime: 1000
  cpuTime: 1000
  memory: 524288
`

var invalidSettingYAML = `
title: Voyage Log
limits: totalTime: 1000`

const problemCode = "114FinalQ006"
func TestGetProblemHandler(t *testing.T){
	t.Run("題目列表正確回傳", func(t *testing.T) {
		// 1. 建立 Test DB
		testDB := testutil.SetupTestDB(t)

		// 2. 寫入資料
		newProblem := models.Problem{
			ProblemCode:	problemCode,
			Title:			"Voyage Log",
			LimitTime: 		1000,
			ProblemPath:	"testfile/problem/114FinalQ006",
		}
		if err := testDB.Create(&newProblem).Error; err != nil {
			t.Error(err)
		}

		// 3. 設定測試 router
		gin.SetMode(gin.TestMode)

		router := gin.New()
		router.GET("/api/problems", GetAllProblemsHandler(testDB))

		// 4. 寫上傳內容
		body := &bytes.Buffer{}
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/problems", body)
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
func TestGetProblemDetailHandler(t *testing.T){
	t.Run("Problem 不存在", func(t *testing.T) {
		// 1. 建立 Test DB
		testDB := testutil.SetupTestDB(t)

		// 2. 設定測試 router
		gin.SetMode(gin.TestMode)

		router := gin.New()
		router.GET("/api/problems/:problemCode", GetProblemDetailHandler(testDB))

		// 3. 寫上傳內容
		body := &bytes.Buffer{}
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/problems/114FinalQ006", body)
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
			ProblemCode:	problemCode,
			Title:			"Voyage Log",
			LimitTime: 		1000,
			ProblemPath:	"testfile/problem/114FinalQ006",
		}
		if err := testDB.Create(&newProblem).Error; err != nil {
			t.Error(err)
		}

		// 3. 設定測試 router
		gin.SetMode(gin.TestMode)

		router := gin.New()
		router.GET("/api/problems/:problemCode", GetProblemDetailHandler(testDB))

		// 4. 寫上傳內容
		body := &bytes.Buffer{}
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/problems/114FinalQ006", body)
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
    t.Run("Insert Problem", func(t *testing.T) {
		// 1. 建立 Test DB
		testDB := testutil.SetupTestDB(t)

		// 2. 設定測試 router
		gin.SetMode(gin.TestMode)

		router := gin.New()
		router.PUT("/api/problems", UpsertProblemHandler(testDB))

		// 3. 建立上傳壓縮檔
		filename := problemCode + ".zip"
		var files = []File {
			{"template/", ""}, 
			{"solution/", ""}, 
			{"spec/", ""}, 
			{"online-judge/", ""},
			{"settings.yaml", settingsYAML},
		}
		createZipFile(t, problemCode, files)

		file, err := os.Open(filename)
		if err != nil {
			t.Error(err)
		}
		defer file.Close()

		// 4. 寫上傳內容
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		_ = writer.WriteField("problemCode", problemCode)
		part, err := writer.CreateFormFile("file", filepath.Base(filename))
		if err != nil {
			t.Error(err)
		}
		_, err = io.Copy(part, file)
		if err != nil {
			t.Error(err)
		}
		_ = writer.Close()

		// 5. Call API 前，確認 Problem 114FinalQ006 不存在
		problem := models.Problem{}
		err = testDB.Where("problem_code = ?", problemCode).First(&problem).Error
		if err == nil{
			t.Fatalf("Problem existed before inserting")
		} else { 
			if !errors.Is(err, gorm.ErrRecordNotFound) {
        		t.Error(err)
			}
		}

		// 6. Call API
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/problems", body)
		req.Header.Add("Content-Type", writer.FormDataContentType())
		router.ServeHTTP(w, req)

		// 7. 確認結果
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d\nbody: %s", w.Code, w.Body.String())
		}

		problem = models.Problem{}
		err = testDB.Where("problem_code = ?", problemCode).First(&problem).Error
		if err != nil{ 
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
			ProblemCode:	problemCode,
			Title:			"Voyage Log",
			LimitTime: 0,
			ProblemPath:	"testfile/problem/114FinalQ006",
		}
		if err := testDB.Create(&newProblem).Error; err != nil {
			t.Error(err)
		}

		// 3. 建立上傳壓縮檔
		filename := problemCode + ".zip"
		var files = []File {
			{"template/new.txt", ""}, 
			{"solution/", ""}, 
			{"spec/", ""}, 
			{"online-judge/", ""},
			{"settings.yaml", settingsYAML},
		}
		createZipFile(t, problemCode, files)

		file, err := os.Open(filename)
		if err != nil {
			t.Error(err)
		}
		defer file.Close()

		// 4. 建立 Problem 資料夾，用 old.txt 標記為舊題目
		problemPath := "testfile/problem/" + problemCode
		err = os.MkdirAll(filepath.Join(problemPath), 0755)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(filepath.Join(problemPath, "old.txt"), []byte("old content"), 0644)
		if err != nil {
			t.Fatal(err)
		}

		// 5. 設定測試 router
		gin.SetMode(gin.TestMode)

		router := gin.New()
		router.PUT("/api/problems", UpsertProblemHandler(testDB))

		// 6. 寫上傳內容
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		
		_ = writer.WriteField("problemCode", problemCode)
		part, err := writer.CreateFormFile("file", filepath.Base(filename))
		if err != nil {
			t.Error(err)
		}
		_, err = io.Copy(part, file)
		if err != nil {
			t.Error(err)
		}
		_ = writer.Close()

		// 7. Call API
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/problems", body)
		req.Header.Add("Content-Type", writer.FormDataContentType())
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d\nbody: %s", w.Code, w.Body.String())
		}

		// 8. 確認結果：資料庫以及資料夾更新成功
		problem := models.Problem{}
		err = testDB.Where("problem_code = ?", problemCode).First(&problem).Error
		if err != nil{ 
			if errors.Is(err, gorm.ErrRecordNotFound) {
        		t.Fatalf("Problem didn't exist")
			} else {
        		t.Error(err)
			}
		}
		if problem.LimitTime == 0 || problem.LimitTime != 1000{
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
		gin.SetMode(gin.TestMode)

		router := gin.New()
		router.PUT("/api/problems", UpsertProblemHandler(testDB))

		// 3. 測試每一種資料不存在的情況
		filename := problemCode + ".zip"
		var files = []File {
			{"template/", ""}, 
			{"solution/", ""}, 
			{"spec/", ""}, 
			{"online-judge/", ""},
			{"settings.yaml", settingsYAML},
		}

		for idx, _ := range files{
			// 4. 建立上傳壓縮檔
			tmpFiles := []File{}
			for _idx, file := range files{
				if idx == _idx{
					continue
				}
				tmpFiles = append(tmpFiles, file)
			}
			createZipFile(t, problemCode, tmpFiles)

			file, err := os.Open(filename)
			if err != nil {
				t.Error(err)
			}
			defer file.Close()

			// 5. 寫上傳內容
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			_ = writer.WriteField("problemCode", problemCode)
			part, err := writer.CreateFormFile("file", filepath.Base(filename))
			if err != nil {
				t.Error(err)
			}
			_, err = io.Copy(part, file)
			if err != nil {
				t.Error(err)
			}
			_ = writer.Close()

			// 6. Call API
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("PUT", "/api/problems", body)
			req.Header.Add("Content-Type", writer.FormDataContentType())
			router.ServeHTTP(w, req)

			// 7. 確認結果
			if w.Code != http.StatusInternalServerError {
				t.Fatalf("expected 500, got %d\nbody: %s", w.Code, w.Body.String())
			}
		}
    })

	t.Run("測試上傳的 settings.yaml 格式錯誤", func(t *testing.T) {
		// 1. 建立 Test DB
		testDB := testutil.SetupTestDB(t)

		// 2. 設定測試 router
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.PUT("/api/problems", UpsertProblemHandler(testDB))

		// 3. 建立上傳壓縮檔
		filename := problemCode + ".zip"
		var files = []File {
			{"template/", ""}, 
			{"solution/", ""}, 
			{"spec/", ""}, 
			{"online-judge/", ""},
			{"settings.yaml", invalidSettingYAML},
		}
		createZipFile(t, problemCode, files)

		file, err := os.Open(filename)
		if err != nil {
			t.Error(err)
		}
		defer file.Close()

		// 4. 寫上傳內容
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		_ = writer.WriteField("problemCode", problemCode)
		part, err := writer.CreateFormFile("file", filepath.Base(filename))
		if err != nil {
			t.Error(err)
		}
		_, err = io.Copy(part, file)
		if err != nil {
			t.Error(err)
		}
		_ = writer.Close()

		// 5. Call API
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/problems", body)
		req.Header.Add("Content-Type", writer.FormDataContentType())
		router.ServeHTTP(w, req)

		// 6. 確認結果
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d\nbody: %s", w.Code, w.Body.String())
		}
    })

	t.Run("測試非 zip 檔案上傳", func(t *testing.T){
		// 1. 建立 Test DB
		testDB := testutil.SetupTestDB(t)

		// 2. 設定測試 router
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.PUT("/api/problems", UpsertProblemHandler(testDB))

		// 3. 寫上傳內容
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		_ = writer.WriteField("problemCode", problemCode)

		_, err := writer.CreateFormFile("file", "test.txt")
		if err != nil {
			t.Fatal(err)
		}

		_ = writer.Close()

		// 4. Call API
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/problems", body)
		req.Header.Add("Content-Type", writer.FormDataContentType())
		router.ServeHTTP(w, req)

		// 5. 確認結果
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d\nbody: %s", w.Code, w.Body.String())
		}
	})

	t.Run("測試壞掉的 zip 檔案上傳", func(t *testing.T) {
		// 1. 建立 Test DB
		testDB := testutil.SetupTestDB(t)

		// 2. 設定測試 router
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.PUT("/api/problems", UpsertProblemHandler(testDB))

		// 3. 寫上傳內容
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		_ = writer.WriteField("problemCode", problemCode)

		_, err := writer.CreateFormFile("file", "broken.zip")
		if err != nil {
			t.Fatal(err)
		}

		_ = writer.Close()

		// 4. Call API
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/problems", body)
		req.Header.Add("Content-Type", writer.FormDataContentType())

		router.ServeHTTP(w, req)

		// 5. 確認結果
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d\nbody: %s", w.Code, w.Body.String())
		}
	})

}

type File struct {
	Name string
	Body string
}

// 將傳入的檔案壓縮成一個 ZIP 檔
func createZipFile(t *testing.T, filename string, files []File){
	t.Helper()

	archive, err := os.Create(filename + ".zip")
    if err != nil {
        t.Error(err)
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
		t.Error(err)
	}
}
