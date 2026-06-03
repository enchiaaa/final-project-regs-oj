// 專門寫 exec.Command、讀取檔案、比對答案
package judge

import (
	"archive/zip"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"online-judge/internal/models"
)

var errTimeLimitExceeded = errors.New("time limit exceeded")

func runJudgingProcess(db *gorm.DB, submissionId string) {
	// 1. 確認 submissionId 和 problem 是否存在
	submission, err := loadSubmission(db, submissionId)
	if err != nil {
		appendInternalJudgeErrorLog(submissionId, "Failed to find submission in database: "+err.Error())
		return
	}
	problem, err := loadProblem(db, submission.ProblemCode)
	if err != nil {
		appendInternalJudgeErrorLog(submissionId, "Failed to find problem in database: "+err.Error())
		return
	}

	submission.SourcePath = filepath.Join("uploads", submission.UserName, submission.ProblemCode, submission.OperatorID) + ".zip"
	submission.WorkspacePath = filepath.Join("tmp", "upload", submission.UserName, submission.ProblemCode, submission.OperatorID, "workspace")
	logPath := filepath.Join("tmp", "upload", submission.UserName, submission.ProblemCode, submission.OperatorID, "logs")

	// 2. 解壓縮上傳的 zip 檔案到 submission.WorkspacePath 目錄下
	if err := prepareWorkspace(submission); err != nil {
		finishSubmission(db, submission, "SE", "Failed to prepare workspace: "+err.Error())
		appendInternalJudgeErrorLog(submissionId, "Failed to prepare workspace: "+err.Error())
		return
	}

	// 3. 建立評測過程中的 log 檔案（configure.log、compile.log、output.log），並將路徑寫入 submission 中
	if err := prepareLogFiles(submission, logPath); err != nil {
		finishSubmission(db, submission, "SE", "Failed to prepare log files: "+err.Error())
		appendInternalJudgeErrorLog(submissionId, "Failed to prepare log files: "+err.Error())
		return
	}
	if err := db.Save(submission).Error; err != nil {
		finishSubmission(db, submission, "SE", "Failed to save log paths: "+err.Error())
		appendInternalJudgeErrorLog(submissionId, "Failed to save log paths: "+err.Error())
		return
	}

	// 4. 確認解壓縮後的 workspace 目錄下是否有 CMakeLists.txt 檔案
	if err := checkCMakeListsExists(submission.WorkspacePath); err != nil {
		finishSubmission(db, submission, "SE", err.Error())
		return
	}

	// 5. configure 階段
	submission.Status = "Configuring"
	if err := db.Save(submission).Error; err != nil {
		appendInternalJudgeErrorLog(submissionId, "Failed to update submission status to Configuring: "+err.Error())
		finishSubmission(db, submission, "SE", "Failed to update submission status to Configuring: "+err.Error())
		return
	}
	mustWriteLog(submissionId, submission.ConfigureLogPath, "Start configuring\n")
	if err := runDockerConfigure(submission, problem); err != nil {
		finishSubmission(db, submission, "SE", "Configuration failed: "+err.Error())
		return
	}

	// 6. compile 階段
	submission.Status = "Compiling"
	if err := db.Save(submission).Error; err != nil {
		appendInternalJudgeErrorLog(submissionId, "Failed to update submission status to Compiling: "+err.Error())
		finishSubmission(db, submission, "SE", "Failed to update submission status to Compiling: "+err.Error())
		return
	}
	mustWriteLog(submissionId, submission.CompileLogPath, "Start compiling\n")
	if err := runDockerCompile(submission, problem); err != nil {
		finishSubmission(db, submission, "CE", "Compilation failed: "+err.Error())
		return
	}

	// 7. judging 階段
	submission.Status = "Judging"
	if err := db.Save(submission).Error; err != nil {
		appendInternalJudgeErrorLog(submissionId, "Judging failed: "+err.Error())
		finishSubmission(db, submission, "SE", "Judging failed: "+err.Error())
		return
	}
	mustWriteLog(submissionId, submission.OutputLogPath, "Start running\n")

	runErr := runDockerRun(submission, problem)

	resultPath := filepath.Join(submission.WorkspacePath, "build", "result.xml")

	if errors.Is(runErr, errTimeLimitExceeded) {
		finishSubmission(db, submission, "TLE", "Time limit error")
		return
	}

	// 若 result.xml 存在
	if _, resultErr := os.Stat(resultPath); resultErr == nil {
		ok, err := checkTestResults(resultPath)
		if err != nil {
			finishSubmission(db, submission, "RE", "Failed to check test results: "+err.Error())
			return
		}

		if !ok {
			finishSubmission(db, submission, "WA", "Some test cases failed")
			return
		}

		finishSubmission(db, submission, "AC", "Accepted")
		return
	}

	if runErr != nil {
		finishSubmission(db, submission, "RE", "Runtime error: "+runErr.Error())
		return
	}

	finishSubmission(db, submission, "RE", "result.xml not found")
}

// 從資料庫中讀取 submission，如果找不到就回傳錯誤
func loadSubmission(db *gorm.DB, submissionId string) (*models.Submission, error) {
	submission := models.Submission{}
	if err := db.Where("operator_id = ?", submissionId).First(&submission).Error; err != nil {
		return nil, err
	}
	return &submission, nil
}

// 從資料庫中讀取 problem，如果找不到就回傳錯誤
func loadProblem(db *gorm.DB, problemCode string) (*models.Problem, error) {
	problem := models.Problem{}
	if err := db.Where("problem_code = ?", problemCode).First(&problem).Error; err != nil {
		return nil, err
	}
	return &problem, nil
}

// 解壓縮上傳的 zip 檔案到 submission.WorkspacePath 目錄下
func prepareWorkspace(submission *models.Submission) error {
	if err := unzipFile(submission.SourcePath, submission.WorkspacePath); err != nil {
		return fmt.Errorf("failed to unzip file: %v", err)
	}
	return nil
}

// 建立評測過程中的 log 檔案（configure.log、compile.log、output.log）
func prepareLogFiles(submission *models.Submission, logPath string) error {
	logDir := logPath
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	submission.ConfigureLogPath = filepath.Join(logDir, "configure.log")
	submission.CompileLogPath = filepath.Join(logDir, "compile.log")
	submission.OutputLogPath = filepath.Join(logDir, "output.log")

	if err := os.WriteFile(submission.ConfigureLogPath, nil, 0644); err != nil {
		return err
	}
	if err := os.WriteFile(submission.CompileLogPath, nil, 0644); err != nil {
		return err
	}
	if err := os.WriteFile(submission.OutputLogPath, nil, 0644); err != nil {
		return err
	}

	return nil
}

// 確認解壓縮後的 workspace 目錄下是否有 CMakeLists.txt 檔案
func checkCMakeListsExists(workspace string) error {
	if _, err := os.Stat(filepath.Join(workspace, "CMakeLists.txt")); err != nil {
		return fmt.Errorf("CMakeLists.txt not found")
	}
	return nil
}

// 結束 submission 並更新狀態和訊息
func finishSubmission(db *gorm.DB, submission *models.Submission, status string, message string) {
	submission.Status = status
	submission.Message = message
	if err := db.Save(submission).Error; err != nil {
		appendInternalJudgeErrorLog(submission.OperatorID, "Failed to update submission as "+status+": "+err.Error())
	}
}

// 使用 Docker 執行 configure 指令，並將輸出寫入 configure.log 檔案中
func runDockerConfigure(submission *models.Submission, problem *models.Problem) error {
	absWorkspace, err := filepath.Abs(submission.WorkspacePath)
	if err != nil {
		return err
	}
	abProblemPath, err := filepath.Abs(problem.ProblemPath)
	if err != nil {
		return err
	}

	// -v 把資料夾掛載到容器內的 /workspace 和 /problem，讓容器內的程式可以讀取使用者上傳的程式碼和題目資料
	// -w 進入容器後，工作目錄直接設為 /workspace
	cmd := exec.Command(
		"docker", "run", "--rm", "--name", "judge-configure-"+submission.OperatorID,
		"-v", absWorkspace+":/workspace",
		"-v", abProblemPath+":/problem",
		"-w", "/workspace",
		"yhlib/cs3060701",
		"cmake",
		"-S", "/problem/solution",
		"-G", "Ninja",
		"-B", "build",
		"-D", "SOURCE_DIR=/workspace",
		"-D", "PROBLEM_ROOT=/problem",
	)

	// 將 configure 的輸出寫入 configure.log 檔案中
	output, err := cmd.CombinedOutput()
	if writeErr := writeToLogFile(submission.ConfigureLogPath, string(output)); writeErr != nil {
		return fmt.Errorf("failed to write configure log: %w", writeErr)
	}
	if err != nil {
		return fmt.Errorf("docker configure failed: %w", err)
	}

	if err := writeToLogFile(submission.ConfigureLogPath, "Configuration completed successfully\n"); err != nil {
		return fmt.Errorf("failed to write configure log: %w", err)
	}
	return nil
}

// 使用 Docker 執行 compile 指令，並將輸出寫入 compile.log 檔案中
func runDockerCompile(submission *models.Submission, problem *models.Problem) error {
	absWorkspace, err := filepath.Abs(submission.WorkspacePath)
	if err != nil {
		return err
	}
	abProblemPath, err := filepath.Abs(problem.ProblemPath)
	if err != nil {
		return err
	}

	// -v 把資料夾掛載到容器內的 /workspace 和 /problem
	// -w 進入容器後，工作目錄直接設為 /workspace
	cmd := exec.Command(
		"docker", "run", "--rm",
		"-v", absWorkspace+":/workspace",
		"-v", abProblemPath+":/problem",
		"-w", "/workspace",
		"yhlib/cs3060701",
		"cmake", "--build", "build",
	)

	// 將 compile 的輸出寫入 compile.log 檔案中
	output, err := cmd.CombinedOutput()
	if writeErr := writeToLogFile(submission.CompileLogPath, string(output)); writeErr != nil {
		return fmt.Errorf("failed to write compile log: %w", writeErr)
	}
	if err != nil {
		return fmt.Errorf("docker compile failed: %w", err)
	}

	if err := writeToLogFile(submission.CompileLogPath, "Compilation completed successfully\n"); err != nil {
		return fmt.Errorf("failed to write compile log: %w", err)
	}
	return nil
}

// 使用 Docker 執行 run 指令，並將輸出寫入 output.log 檔案中
func runDockerRun(submission *models.Submission, problem *models.Problem) error {
	abwsWorkspace, err := filepath.Abs(submission.WorkspacePath)
	if err != nil {
		return err
	}
	abProblemPath, err := filepath.Abs(problem.ProblemPath)
	if err != nil {
		return err
	}

	containerName := "judge-run-" + submission.OperatorID

	// -v 把資料夾掛載到容器內的 /workspace 和 /problem
	// -w 進入容器後，工作目錄直接設為 /workspace
	cmd := exec.Command(
		"docker", "run", "--name", containerName, "--rm",
		"--network", "none",
		"-v", abwsWorkspace+":/workspace",
		"-v", abProblemPath+":/problem",
		"-w", "/workspace",
		"yhlib/cs3060701",
		"timeout", strconv.Itoa(problem.LimitTime)+"s",
		"ctest", "-Q", "--test-dir", "build", "--output-junit", "result.xml",
	)

	// 將 run 的輸出寫入 output.log 檔案中
	output, err := cmd.CombinedOutput()
	if writeErr := writeToLogFile(submission.OutputLogPath, string(output)); writeErr != nil {
		return fmt.Errorf("failed to write output log: %w", writeErr)
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() == 124 {
			_ = exec.Command("docker", "rm", "-f", containerName).Run()
			return errTimeLimitExceeded
		}
	}

	resultPath := filepath.Join(submission.WorkspacePath, "build", "result.xml")
	if _, resultErr := os.Stat(resultPath); resultErr == nil {
		return nil
	}

	if err != nil {
		return fmt.Errorf("docker run failed without result.xml: %w", err)
	}

	return fmt.Errorf("result.xml not found")
}

// 查看 result.xml 的內容，確認是否有測資沒有通過
func checkTestResults(resultPath string) (bool, error) {
	type TestCaseResult struct {
		Failures int `xml:"failures,attr"`
	}

	resultFile, err := os.ReadFile(resultPath)
	if err != nil {
		return false, fmt.Errorf("failed to read result file: %w", err)
	}
	
	var testResult TestCaseResult
	if err := xml.Unmarshal(resultFile, &testResult); err != nil {
		return false, fmt.Errorf("failed to parse result file: %w", err)
	}

	return testResult.Failures == 0, nil
}

// 將評測過程中的錯誤訊息記錄到系統中，供管理員查看
func appendInternalJudgeErrorLog(submissionId string, msg string) {
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

// 將評測過程中的 log 訊息寫入對應的 log 檔案中
func writeToLogFile(logPath string, content string) error {
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		return err
	}
	return nil
}

// 如果寫入 log 檔案失敗，則將錯誤訊息記錄到系統中，供管理員查看
func mustWriteLog(submissionId string, logPath string, content string) {
	if err := writeToLogFile(logPath, content); err != nil {
		appendInternalJudgeErrorLog(submissionId, "Failed to write log "+logPath+": "+err.Error())
	}
}

// 解壓縮 zip 檔案到指定目錄
// 程式碼參考自 https://golang.cafe/blog/golang-unzip-file-example
func unzipFile(src string, dst string) error {
	archive, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer archive.Close() // 記得關閉檔案

	absDst, err := filepath.Abs(dst)
	if err != nil {
		return err
	}
	cleanDst := filepath.Clean(absDst)

	// 遍歷 zip 檔案中的每個檔案，將它們解壓縮到指定目錄下
	for _, f := range archive.File {
		filePath := filepath.Join(cleanDst, f.Name)
		cleanPath := filepath.Clean(filePath) // ！防止 zip slip 漏洞，確保解壓縮後的路徑在指定目錄下

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
