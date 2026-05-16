### 目錄結構  
| 資料夾名稱 | 說明 |   
| - | - |
| /cmd | 程式進入點，啟動 HTTP server 與 Judge 工作程序 |
| /internal/api | API 路由與控制器實作 |
| /internal/models | 資料模型定義（Submission、User、Problem 等）|
| /internal/judge | 判題邏輯與背景工作處理 |
| /internal/docker | 容器操作相關支援程式 |
| /uploads | 上傳資料存放目錄 |
| /solution | 助教的正確解答 |
| /resources | 提供給學生的程式碼框架與必要檔案 |

### 執行方式
```bash
# 啟動背景資料庫容器
docker compose up -d

# 取得相關依賴套件
go mod tidy

# 啟動主程式
go run cmd/main.go
```