// 專門寫 exec.Command、讀取檔案、比對答案
package judge

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"

	"online-judge/internal/models"
)

func runJudgingProcess(jobQueue chan string, db *gorm.DB, submissionId string) {
	// 確認 submissionId 是否存在
	submission := models.Submission{}
	if err := db.Where("id = ?", submissionId).First(&submission).Error; err != nil {
		// 記錄錯誤訊息到資料庫中
		appendJudgeErrorLog(submissionId, "Failed to find submission in database: "+err.Error())
		return
	}

	// 解壓縮上傳的 zip 檔案到 "tmp/oj/{submissionId}/workspace" 目錄下
	// TODO: 檢查專案根目錄是否存在 CMakeLists.txt
	if err := unzipFile(submission.SourcePath, filepath.Join("tmp", "oj", submission.OperatorID, "workspace")); err != nil {
		submission.Status = "SE"
		db.Save(&submission)
		appendJudgeErrorLog(submissionId, "Failed to unzip source file: "+err.Error())
		return
	}

	// 建立評測過程中的 log 檔案（configure.log、compile.log、output.log）
	configureLog, compileLog, outputLog, err := createSubmissionLogs(submissionId)
	if err != nil {
		submission.Status = "SE"
		db.Save(&submission)
		appendJudgeErrorLog(submissionId, "Failed to create submission logs: "+err.Error())
		return
	}

	// 更新 submission 狀態為 "Configuring"，並將 log 檔案路徑寫入資料庫
	submission.Status = "Configuring"
	submission.ConfigureLogPath = configureLog
	submission.CompileLogPath = compileLog
	submission.OutputLogPath = outputLog
	writeToLogFile(configureLog, "Start configuring...\n")
	if err := db.Save(&submission).Error; err != nil {
		appendJudgeErrorLog(submissionId, "Failed to update submission status to Configuring: "+err.Error())
		return
	}
	writeToLogFile(configureLog, "Configuration completed successfully\n")

	// 模擬 compile 過程
	submission.Status = "Compiling"
	writeToLogFile(compileLog, "Start compiling...\n")
	if err := db.Save(&submission).Error; err != nil {
		writeToLogFile(compileLog, "Compilation failed: "+err.Error()+"\n")
		return
	}
	writeToLogFile(compileLog, "Compilation completed successfully\n")

	// 模擬 judging 過程
	submission.Status = "Judging"
	writeToLogFile(outputLog, "Start judging...\n")
	if err := db.Save(&submission).Error; err != nil {
		writeToLogFile(outputLog, "Judging failed: "+err.Error()+"\n")
		return
	}

	// 模擬評測結果為 AC
	now := time.Now()
	submission.Status = "AC"
	submission.FinishedAt = &now
	writeToLogFile(outputLog, "Judging completed: AC\n")
	if err := db.Save(&submission).Error; err != nil {
		writeToLogFile(outputLog, "Failed to update submission status to AC: "+err.Error()+"\n")
		return
	}
}

// 將評測過程中的錯誤訊息記錄到資料庫中
func appendJudgeErrorLog(submissionId string, msg string) {
	// 0755 是 Linux 的檔案權限，表示擁有者有讀寫執行權限，群組和其他人只有讀取和執行權限
	err := os.MkdirAll("tmp/oj", 0755)
	if err != nil {
		log.Printf("Failed to create directory for error logs: %v", err)
		return
	}

	// 錯誤訊息格式為 "[timestamp] submissionId: msg"
	line := fmt.Sprintf("%s: operatorId=%s %s\n", time.Now().Format(time.RFC3339), submissionId, msg)

	// 將錯誤訊息寫入 "tmp/oj/judge_error.log" 檔案中
	// O_APPEND 表示如果檔案已經存在，就在檔案末尾追加內容；O_CREATE 表示如果檔案不存在，就創建一個新檔案；O_WRONLY 表示以寫入模式打開檔案
	// 0644 是檔案權限，表示所有人可讀取，但只有擁有者可以寫入
	f, err := os.OpenFile("tmp/oj/judge_error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open error log file: %v", err)
		return
	}
	defer f.Close()

	if _, err := f.WriteString(line); err != nil {
		log.Printf("Failed to write to error log file: %v", err)
		return
	}
}

// 建立評測過程中的 log 檔案（configure.log、compile.log、output.log）
func createSubmissionLogs(operatorId string) (string, string, string, error) {
	logDir := filepath.Join("tmp", "oj", operatorId, "logs")

	if err := os.MkdirAll(logDir, 0755); err != nil {
		return "", "", "", err
	}

	configureLog := filepath.Join(logDir, "configure.log")
	compileLog := filepath.Join(logDir, "compile.log")
	outputLog := filepath.Join(logDir, "output.log")

	if err := os.WriteFile(configureLog, nil, 0644); err != nil {
		return "", "", "", err
	}
	if err := os.WriteFile(compileLog, nil, 0644); err != nil {
		return "", "", "", err
	}
	if err := os.WriteFile(outputLog, nil, 0644); err != nil {
		return "", "", "", err
	}

	return configureLog, compileLog, outputLog, nil
}

// 將評測過程中的 log 訊息寫入對應的 log 檔案中
func writeToLogFile(logPath string, content string) error {
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		return err
	}
	return nil
}

// 解壓縮 zip 檔案到指定目錄
// 程式碼參考自 https://golang.cafe/blog/golang-unzip-file-example
func unzipFile(src string, dst string) error {
	archive, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer archive.Close() // 記得關閉檔案

	cleanDst := filepath.Clean(dst)

	// 遍歷 zip 檔案中的每個檔案，將它們解壓縮到指定目錄下
	for _, f := range archive.File {
		filePath := filepath.Join(cleanDst, f.Name)
		cleanPath := filepath.Clean(filePath)	// ！防止 zip slip 漏洞，確保解壓縮後的路徑在指定目錄下

		if cleanPath != cleanDst && !strings.HasPrefix(cleanPath, cleanDst+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", f.Name)
		}

		// 如果是目錄，則建立目錄
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(cleanPath, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(cleanPath), 0755); err != nil {
			return err
		}

		// 如果是檔案，則解壓縮到指定目錄
		dstFile, err := os.OpenFile(cleanPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		// 開啟 zip 檔案中的檔案，並將內容寫入剛剛建立的檔案中
		fileInArchive, err := f.Open()
		if err != nil {
			dstFile.Close()
			return err
		}

		// io.Copy 會從 fileInArchive 讀取內容，並寫入 dstFile 中，直到讀取完畢或發生錯誤
		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			dstFile.Close()
			fileInArchive.Close()
			return err
		}

		dstFile.Close()
		fileInArchive.Close()
	}

	return nil
}
