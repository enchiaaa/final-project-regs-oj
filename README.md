# REGS Online Judge

REGS Online Judge 是以 Go、Gin、GORM、PostgreSQL 與 Docker 實作的線上評測後端。系統提供使用者認證、RBAC 權限控管、題目管理、非同步評測、併發限制、分段日誌與統計查詢。

## 主要功能

- 使用者註冊、登入與 ES256 JWT 驗證
- Guest、User、Admin 權限控制
- 題目建立、更新、查詢、刪除與測資下載
- ZIP Submission 上傳與 Zip Slip 防護
- CMake Configure、Compile、CTest 自動化評測
- `AC`、`WA`、`CE`、`RE`、`TLE`、`SE` 狀態判定
- 非同步 Job Queue 與 Semaphore 併發控制
- Configure、Compile、Output 分段日誌查詢
- 既有 Submission 重新評測
- 題目與使用者提交統計

## 系統架構

```text
Client
  |
  v
HTTP API（Gin）
  |-- JWT Authentication
  |-- RBAC Permission 與資源 Ownership 檢查
  |-- PostgreSQL（GORM）
  |
  `-- Job Queue（容量 100）
        |
        `-- Semaphore（最多同時執行 2 個 Judge）
              |
              `-- Docker Judge
                    |-- Configure：CMake + Ninja
                    |-- Compile：CMake Build
                    `-- Run：CTest + timeout + --network none
```

Submission 建立後會先以 `Pending` 狀態寫入 PostgreSQL，API 隨即回傳 `operatorId`。背景 Worker 從 Queue 取出任務，依序執行 Configure、Compile 與 Run，最後更新評測狀態並保存日誌。

## 技術環境

- Go 1.26.1
- Gin 1.12
- GORM 1.31
- PostgreSQL 16
- Docker
- CMake、Ninja、CTest
- JWT ES256

評測使用的 Docker image：

```text
yhlib/cs3060701
```

## 專案結構

| 路徑 | 說明 |
| --- | --- |
| `cmd/` | HTTP Server 與 Judge Worker 進入點 |
| `internal/api/` | API Routes、Handlers 與上傳處理 |
| `internal/database/` | 資料庫連線、Migration、RBAC 與 Admin Seed |
| `internal/judge/` | Job Queue、Docker 評測與結果比對 |
| `internal/middleware/` | JWT 驗證與 Permission Middleware |
| `internal/models/` | GORM Models |
| `internal/rbac/` | Role 與 Permission 定義 |
| `internal/utils/` | ZIP 壓縮與安全解壓縮工具 |
| `problem/` | 題目、測資、正解與設定檔 |
| `test_files/` | E2E 測試使用的題目與 Submission ZIP |
| `uploads/` | 使用者上傳的原始 Submission ZIP |
| `tmp/` | 評測 Workspace、日誌與題目暫存資料 |
| `docs/openapi.yaml` | OpenAPI 3.0 API 文件 |
| `docs/ERD.md` | 資料庫 ERD |
| `scripts/e2e.sh` | 端到端測試腳本 |

`uploads/`、`tmp/`、`.env`、JWT 金鑰與 Go 快取皆不應提交至 Git。

## 前置需求

- Go
- Docker Desktop 或 Docker Engine
- Docker Compose
- OpenSSL
- PostgreSQL Client（選用）
- Bash、`curl`、`jq`（執行 E2E 測試時需要）

先確認 Docker 可正常使用：

```bash
docker version
docker compose version
```

> 評測流程會由 Go 程式直接呼叫 `docker` CLI，因此 API Server 所在環境必須能存取 Docker daemon。

## 快速開始

### 1. 啟動 PostgreSQL

```bash
docker compose up -d
```

本機使用 Docker Compose 時，請在 `.env` 設定資料庫連線：

```dotenv
DATABASE_URL=postgres://user:123@localhost:5432/OJ_db?sslmode=disable
```

程式會從 `DATABASE_URL` 讀取連線資訊。若要調整資料庫帳密或連線位置，請同步修改 `docker-compose.yml` 與 `.env`。

### 2. 下載評測 Image

```bash
docker pull yhlib/cs3060701
```

### 3. 產生 JWT 金鑰

在專案根目錄建立 `keys/`：

```bash
mkdir -p keys
openssl genpkey -algorithm EC -pkeyopt ec_paramgen_curve:P-256 -out keys/private.pem
openssl pkey -in keys/private.pem -pubout -out keys/public.pem
```
### 4. 設定環境變數

```bash
cp .env.example .env
```

`.env` 範例：

```dotenv
DATABASE_URL=postgres://user:123@localhost:5432/OJ_db?sslmode=disable

JWT_PRIVATE_KEY_PATH=keys/private.pem
JWT_PUBLIC_KEY_PATH=keys/public.pem

ADMIN_USERNAME=admin
ADMIN_PASSWORD=change-me
```

- JWT 金鑰路徑是相對於程式啟動目錄，請在專案根目錄執行 Server。
- `ADMIN_USERNAME` 與 `ADMIN_PASSWORD` 必須同時設定。
- 若兩者皆未設定，系統會略過預設 Admin 建立。
- 若只設定其中一項，Server 啟動時會回報錯誤。

### 5. 安裝 Dependencies 並啟動

```bash
go mod download
go run ./cmd
```

API 預設位址：

```text
http://localhost:8080
```

啟動時會自動：

- 連線 PostgreSQL
- 執行 GORM `AutoMigrate`
- 建立預設 Role、Permission 與關聯
- 依環境變數建立預設 Admin
- 建立容量為 100 的 Job Queue
- 啟動背景 Judge Worker
- 限制最多同時執行 2 個 Judge

## Role 與權限

| 功能 | Guest | User | Admin |
| --- | :---: | :---: | :---: |
| 註冊、登入 | O | O | O |
| 查看題目、公開提交摘要與統計 | O | O | O |
| 查看自己的資料 | X | O | O |
| 建立及查詢自己的 Submission | X | O | O |
| 下載自己的原始 Submission、查詢日誌、重新評測 | X | O | O |
| 建立、更新、刪除題目 | X | X | O |
| 下載完整題目測資 | X | X | O |
| 存取其他使用者的 Submission 資源 | X | X | O |

需要登入的 API 必須帶入：

```http
Authorization: Bearer <JWT>
```

JWT 有效期限為 24 小時。Logout API 不會建立 Token 黑名單。

## API 一覽

完整 request、response 與錯誤格式請參考 `docs/openapi.yaml`。

| Method | Path | 權限 | 用途 |
| --- | --- | --- | --- |
| `POST` | `/api/users/register` | Guest | 註冊一般使用者 |
| `POST` | `/api/users/login` | Guest | 登入並取得 JWT |
| `POST` | `/api/users/logout` | User | 登出 |
| `GET` | `/api/users/me` | User | 取得目前登入者資料 |
| `GET` | `/api/users/{userId}/submissions` | Guest | 取得指定使用者的題目與狀態摘要 |
| `GET` | `/api/problems` | Guest | 取得題目列表 |
| `GET` | `/api/problems/{problemId}` | Guest | 取得題目詳細資料 |
| `PUT` | `/api/problems` | Admin | 建立或更新題目 |
| `DELETE` | `/api/problems/{problemId}` | Admin | 刪除題目資料 |
| `GET` | `/api/problems/{problemId}/testcases` | Admin | 下載完整題目 ZIP |
| `POST` | `/api/submissions` | User | 上傳 Submission 並建立評測任務 |
| `GET` | `/api/submissions` | User | 查詢自己的 Submission 列表 |
| `GET` | `/api/submissions/{operatorId}` | Owner／Admin | 查詢單筆評測結果 |
| `GET` | `/api/submissions/{operatorId}/source` | Owner／Admin | 下載原始 Submission ZIP |
| `GET` | `/api/submissions/{operatorId}/logs/{logType}` | Owner／Admin | 查詢分段日誌 |
| `POST` | `/api/submissions/{operatorId}/rerun` | Owner／Admin | 重新評測 |
| `GET` | `/api/stats/problems/{problemId}` | Guest | 取得題目統計 |
| `GET` | `/api/stats/users/{userId}` | Guest | 取得使用者統計 |

> `{problemId}` 與 `{userId}` 使用資料庫的數字 ID，不是 `problemCode` 或 `username`。

### 常用 Request 格式

註冊與登入使用 JSON：

```json
{
  "username": "B11215001",
  "password": "password"
}
```

建立或更新題目使用 `multipart/form-data`：

```text
problemCode=<題目代碼>
file=<題目 ZIP>
```

建立 Submission 使用 `multipart/form-data`：

```text
problemCode=<題目代碼>
file=<Submission ZIP>
```

`problemCode` 只允許英文字母、數字、`-` 與 `_`。

## 題目 ZIP 格式

題目 ZIP 解壓縮後，根目錄至少必須包含：

```text
template/
solution/
spec/
online-judge/
settings.yaml
```

請直接壓縮上述內容，不要在 ZIP 外層多包一層題目資料夾。實際評測時還需要：

- `solution/CMakeLists.txt`：供 CMake Configure 使用
- `settings.yaml`：提供題目標題、時間限制、測資替換與 expected 設定
- `online-judge/`：保存 expected 輸出與隱藏測資

`settings.yaml` 範例：

```yaml
title: Example Problem

limits:
  totalTime: 1000
  cpuTime: 1000
  memory: 524288

presets:
  - score: 2
    replace:
      - source: template/entrypoint.cpp
        target: entrypoint.cpp
      - source: online-judge/case1.h
        target: case.h
    expected:
      type: file-content
      value: online-judge/case1
```

`limits.totalTime` 單位為毫秒。`1000` 代表執行時間限制為 1 秒。

## Submission ZIP 格式

Submission ZIP 解壓縮後的根目錄必須包含 `CMakeLists.txt`：

```text
CMakeLists.txt
<學生實作檔案>
```

同樣不要在 ZIP 外層多包一層資料夾，否則系統會因根目錄找不到 `CMakeLists.txt` 而回傳 `SE`。

上傳成功後，原始 ZIP 會保存於：

```text
uploads/<username>/<problemCode>/<operatorId>.zip
```

## Submission 與評測流程

1. User 上傳 ZIP 與 `problemCode`。
2. Server 建立狀態為 `Pending` 的 Submission。
3. Server 回傳 `operatorId`，並將任務放入 Queue。
4. Worker 解壓縮 Submission 並檢查根目錄的 `CMakeLists.txt`。
5. Configure 容器使用題目的 `solution/CMakeLists.txt`，並傳入 `SOURCE_DIR` 與 `PROBLEM_ROOT`。
6. Compile 容器執行 `cmake --build build`。
7. Run 容器使用 `--network none`，並以 `timeout` 執行 CTest。
8. 系統解析 `build/result.xml`，比對 testcase 輸出與題目 expected 檔案。
9. 最終狀態與訊息寫回 PostgreSQL。

狀態流程：

```text
Pending
  -> Configuring
  -> Compiling
  -> Judging
  -> AC / WA / CE / RE / TLE / SE
```

| 狀態 | 意義 |
| --- | --- |
| `Pending` | 等待 Worker 處理 |
| `Configuring` | 正在執行 CMake Configure |
| `Compiling` | 正在編譯 |
| `Judging` | 正在執行 CTest 與結果比對 |
| `AC` | 全部測試通過 |
| `WA` | 至少一筆輸出或測試結果不符合預期 |
| `CE` | 編譯失敗 |
| `RE` | 執行錯誤、結果檔缺失或結果解析失敗 |
| `TLE` | 超過題目時間限制 |
| `SE` | 解壓縮、檔案結構或 Configure 等環境設定錯誤 |

Rerun 會保留原始 ZIP、清除舊 Workspace，並將同一個 `operatorId` 重新加入 Queue。處於 `Pending`、`Configuring`、`Compiling` 或 `Judging` 的 Submission 不可重複 Rerun。

## 評測隔離與日誌

- Run 階段使用 Docker `--network none` 關閉網路。
- Configure、Compile 與 Run 使用獨立的短生命週期容器。
- Run 階段同時使用容器內 `timeout` 與 Go `context` 控制超時。
- 系統限制最多同時執行 2 個評測。

每筆 Submission 會保存：

```text
tmp/upload/<username>/<problemCode>/<operatorId>/logs/configure.log
tmp/upload/<username>/<problemCode>/<operatorId>/logs/compile.log
tmp/upload/<username>/<problemCode>/<operatorId>/logs/output.log
```

查詢 API：

```http
GET /api/submissions/{operatorId}/logs/configure
GET /api/submissions/{operatorId}/logs/compile
GET /api/submissions/{operatorId}/logs/output
```

Judge 內部錯誤會追加至：

```text
tmp/oj/judge_error.log
```

## 測試

### Go 測試

```bash
go test ./...
```

部分 API 測試會使用 PostgreSQL 測試資料庫；執行前請確認測試所需的資料庫環境已就緒。

### E2E 測試

E2E 腳本會驗證登入、RBAC、題目上傳、AC 評測、三段日誌、Rerun、下載與統計 API。

先啟動 PostgreSQL 與 API Server，再於另一個 Bash 終端機執行：

```bash
bash scripts/e2e.sh
```

可透過環境變數覆寫測試設定，例如：

```bash
ADMIN_USERNAME=admin \
ADMIN_PASSWORD=password \
PROBLEM_CODE=114FinalQ001 \
bash scripts/e2e.sh
```

預設 E2E 腳本只執行 AC 流程；WA、CE、SE、RE、TLE 測試案例目前保留在腳本中但預設註解。

## API 與資料庫文件

- OpenAPI 3.0：[`docs/openapi.yaml`](docs/openapi.yaml)
- ERD：[`docs/ERD.md`](docs/ERD.md)