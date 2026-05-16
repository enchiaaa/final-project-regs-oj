// 專門寫 exec.Command、讀取檔案、比對答案
package judge

import "os/exec"

func runJudgingProcess(submissionId string) {
	// 1. 從資料庫讀取 submission 資訊（包含測資路徑、使用者程式碼路徑等等）
	filename := "path/to/submission/source.zip" // TODO: 從資料庫讀取 submission 資訊，取得使用者上傳的程式碼路徑
	dst := "path/to/unzip/destination" 			// TODO: 定義一個路徑來解壓縮使用者上傳的程式碼
	cmd := exec.Command("unzip", filename, "-d", dst)

	if err := cmd.Run(); err != nil {
		// 更新資料庫中的 submission 狀態為 "SE"（System Error）
		return
	}

	// 2. 編譯程式碼：使用 cmake + ninja 來編譯程式碼，將編譯出來的執行檔放在 `build/` 目錄下
	// 3. 執行程式並抓取輸出：執行編譯出來的執行檔，將結果導向 `user_output.txt`
	// 4. 比對結果：讀取 `user_output.txt` 與題目提供的 `answer.txt`
	// 5. 更新評測結果到資料庫（AC/WA/CE/TLE/RE 等等）
}

