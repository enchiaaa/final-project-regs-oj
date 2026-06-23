#!/usr/bin/env bash
set -euo pipefail

echo "開始 E2E 測試"

# 共用設定
BASE_URL="${BASE_URL:-http://localhost:8080}"

ADMIN_USERNAME="${ADMIN_USERNAME:-admin}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-password}"

USER_USERNAME="${E2E_USER_USERNAME:-e2e-user}"
USER_PASSWORD="${E2E_USER_PASSWORD:-password}"

PROBLEM_CODE="${PROBLEM_CODE:-114FinalQ001}"
PROBLEM_ZIP="${PROBLEM_ZIP:-test_files/114FinalQ001.zip}"

AC_SUBMISSION_ZIP="${AC_SUBMISSION_ZIP:-test_files/114FinalQ001solution(AC).zip}"
WA_SUBMISSION_ZIP="${WA_SUBMISSION_ZIP:-test_files/114FinalQ001solution(WA).zip}"
CE_SUBMISSION_ZIP="${CE_SUBMISSION_ZIP:-test_files/114FinalQ001solution(CE).zip}"
SE_SUBMISSION_ZIP="${SE_SUBMISSION_ZIP:-test_files/114FinalQ001solution(SE).zip}"
RE_SUBMISSION_ZIP="${RE_SUBMISSION_ZIP:-test_files/114FinalQ001solution(RE).zip}"
TLE_SUBMISSION_ZIP="${TLE_SUBMISSION_ZIP:-test_files/114FinalQ001solution(TLE).zip}"

# 設為 1 時，缺少任一非 AC fixture 會讓測試失敗。
REQUIRE_ALL_STATUS_FIXTURES="${REQUIRE_ALL_STATUS_FIXTURES:-0}"

PROCESSING_STATUSES="Pending Configuring Compiling Judging"
MAX_ATTEMPTS="${MAX_ATTEMPTS:-60}"
POLL_INTERVAL_SECONDS="${POLL_INTERVAL_SECONDS:-1}"

TEST_TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TEST_TMP_DIR"' EXIT

command -v curl >/dev/null || {
  echo "curl is required"
  exit 1
}

command -v jq >/dev/null || {
  echo "jq is required"
  exit 1
}

require_file() {
  local path="$1"
  local description="$2"

  if [[ ! -f "$path" ]]; then
    echo "$description not found: $path"
    exit 1
  fi
}

login() {
  local username="$1"
  local password="$2"

  curl --fail --silent \
    -X POST "$BASE_URL/api/users/login" \
    -H "Content-Type: application/json" \
    -d "{
      \"username\":\"$username\",
      \"password\":\"$password\"
    }" |
    jq -r '.token'
}

submit() {
  local submission_zip="$1"

  curl --fail --silent \
    -X POST "$BASE_URL/api/submissions" \
    -H "Authorization: Bearer $USER_TOKEN" \
    -F "problemCode=$PROBLEM_CODE" \
    -F "file=@$submission_zip" |
    jq -r '.OperatorId'
}

wait_for_result() {
  local operator_id="$1"
  local result=""
  local status=""

  for ((attempt = 1; attempt <= MAX_ATTEMPTS; attempt++)); do
    result=$(
      curl --fail --silent \
        "$BASE_URL/api/submissions/$operator_id" \
        -H "Authorization: Bearer $USER_TOKEN"
    )

    status=$(jq -r '.Status' <<<"$result")
    echo "  Attempt $attempt: $status" >&2

    if [[ ! " $PROCESSING_STATUSES " =~ " $status " ]]; then
      printf '%s' "$result"
      return 0
    fi

    sleep "$POLL_INTERVAL_SECONDS"
  done

  echo "Submission $operator_id did not finish after $MAX_ATTEMPTS attempts" >&2
  return 1
}

assert_status() {
  local result="$1"
  local expected_status="$2"
  local operator_id="$3"
  local actual_status
  local message

  actual_status=$(jq -r '.Status' <<<"$result")
  message=$(jq -r '.Message' <<<"$result")

  if [[ "$actual_status" != "$expected_status" ]]; then
    echo "Submission $operator_id: expected $expected_status, got $actual_status: $message"
    exit 1
  fi

  echo "$expected_status flow passed: $operator_id"
}

assert_http_status() {
  local actual_status="$1"
  local expected_status="$2"
  local description="$3"

  if [[ "$actual_status" != "$expected_status" ]]; then
    echo "$description: expected HTTP $expected_status, got $actual_status"
    exit 1
  fi

  echo "$description passed"
}

run_status_case() {
  local expected_status="$1"
  local submission_zip="$2"

  if [[ ! -f "$submission_zip" ]]; then
    if [[ "$REQUIRE_ALL_STATUS_FIXTURES" == "1" ]]; then
      echo "$expected_status fixture not found: $submission_zip"
      exit 1
    fi

    echo "SKIP $expected_status: fixture not found ($submission_zip)"
    return 0
  fi

  echo "測試 $expected_status: $submission_zip"

  local operator_id
  local result
  operator_id=$(submit "$submission_zip")

  if [[ -z "$operator_id" || "$operator_id" == "null" ]]; then
    echo "$expected_status submission did not return OperatorId"
    exit 1
  fi

  echo "Submission created: $operator_id"
  result=$(wait_for_result "$operator_id")
  assert_status "$result" "$expected_status" "$operator_id"
}

check_log() {
  local operator_id="$1"
  local log_type="$2"
  local log_content

  log_content=$(
    curl --fail --silent \
      "$BASE_URL/api/submissions/$operator_id/logs/$log_type" \
      -H "Authorization: Bearer $USER_TOKEN"
  )

  if [[ -z "$log_content" ]]; then
    echo "$log_type log is empty for submission $operator_id"
    exit 1
  fi

  echo "$log_type log query passed"
}

rerun_and_assert_status() {
  local operator_id="$1"
  local expected_status="$2"
  local rerun_response
  local result

  rerun_response=$(
    curl --fail --silent \
      -X POST "$BASE_URL/api/submissions/$operator_id/rerun" \
      -H "Authorization: Bearer $USER_TOKEN"
  )

  if [[ "$(jq -r '.status' <<<"$rerun_response")" != "Pending" ]]; then
    echo "Rerun did not return Pending: $rerun_response"
    exit 1
  fi

  echo "Rerun queued: $operator_id"
  result=$(wait_for_result "$operator_id")
  assert_status "$result" "$expected_status" "$operator_id"
}

# 1. 確認必要檔案與 API Server
require_file "$PROBLEM_ZIP" "Problem ZIP"
require_file "$AC_SUBMISSION_ZIP" "AC Submission ZIP"

curl --fail --silent "$BASE_URL/api/problems" >/dev/null
echo "API server is ready"

# 2. Admin 登入
ADMIN_TOKEN=$(login "$ADMIN_USERNAME" "$ADMIN_PASSWORD")

if [[ -z "$ADMIN_TOKEN" || "$ADMIN_TOKEN" == "null" ]]; then
  echo "Admin login did not return a token"
  exit 1
fi

echo "Admin 登入成功"

# 3. Admin 上傳題目
curl --fail --silent \
  -X PUT "$BASE_URL/api/problems" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -F "problemCode=$PROBLEM_CODE" \
  -F "file=@$PROBLEM_ZIP" >/dev/null

echo "Problem 上傳成功"

# 4. 註冊一般 User；重複執行時允許 409。
REGISTER_RESPONSE="$TEST_TMP_DIR/register-response.json"

REGISTER_STATUS=$(
  curl --silent \
    -o "$REGISTER_RESPONSE" \
    -w "%{http_code}" \
    -X POST "$BASE_URL/api/users/register" \
    -H "Content-Type: application/json" \
    -d "{
      \"username\":\"$USER_USERNAME\",
      \"password\":\"$USER_PASSWORD\"
    }"
)

if [[ "$REGISTER_STATUS" != "201" && "$REGISTER_STATUS" != "409" ]]; then
  cat "$REGISTER_RESPONSE"
  echo "Register failed: HTTP $REGISTER_STATUS"
  exit 1
fi

echo "User 註冊確認完成"

# 5. User 登入
USER_TOKEN=$(login "$USER_USERNAME" "$USER_PASSWORD")

if [[ -z "$USER_TOKEN" || "$USER_TOKEN" == "null" ]]; then
  echo "User login did not return a token"
  exit 1
fi

echo "User 登入成功"

# 6. 驗證認證與基本 RBAC
USER_PROFILE=$(
  curl --fail --silent \
    "$BASE_URL/api/users/me" \
    -H "Authorization: Bearer $USER_TOKEN"
)
USER_ID=$(jq -r '.user_id' <<<"$USER_PROFILE")

if [[ -z "$USER_ID" || "$USER_ID" == "null" ]]; then
  echo "GET /api/users/me did not return user_id: $USER_PROFILE"
  exit 1
fi
if [[ "$(jq -r '.username' <<<"$USER_PROFILE")" != "$USER_USERNAME" ]]; then
  echo "GET /api/users/me returned unexpected username: $USER_PROFILE"
  exit 1
fi
echo "User profile query passed: userId=$USER_ID"

GUEST_SUBMIT_STATUS=$(
  curl --silent \
    -o "$TEST_TMP_DIR/guest-submit-response.json" \
    -w "%{http_code}" \
    -X POST "$BASE_URL/api/submissions"
)
assert_http_status "$GUEST_SUBMIT_STATUS" "401" "Guest submission authorization"

USER_UPSERT_STATUS=$(
  curl --silent \
    -o "$TEST_TMP_DIR/user-upsert-response.json" \
    -w "%{http_code}" \
    -X PUT "$BASE_URL/api/problems" \
    -H "Authorization: Bearer $USER_TOKEN" \
    -F "problemCode=$PROBLEM_CODE" \
    -F "file=@$PROBLEM_ZIP"
)
assert_http_status "$USER_UPSERT_STATUS" "403" "User problem upsert authorization"

# 7. AC、三段 Log 與 Rerun
echo "測試 AC: $AC_SUBMISSION_ZIP"
AC_OPERATOR_ID=$(submit "$AC_SUBMISSION_ZIP")

if [[ -z "$AC_OPERATOR_ID" || "$AC_OPERATOR_ID" == "null" ]]; then
  echo "AC submission did not return OperatorId"
  exit 1
fi

echo "Submission created: $AC_OPERATOR_ID"
AC_RESULT=$(wait_for_result "$AC_OPERATOR_ID")
assert_status "$AC_RESULT" "AC" "$AC_OPERATOR_ID"

check_log "$AC_OPERATOR_ID" "configure"
check_log "$AC_OPERATOR_ID" "compile"
check_log "$AC_OPERATOR_ID" "output"

rerun_and_assert_status "$AC_OPERATOR_ID" "AC"

# 8. 驗證題目、Submission、下載與統計 API
PROBLEMS_RESPONSE=$(curl --fail --silent "$BASE_URL/api/problems")
if ! jq -e --arg code "$PROBLEM_CODE" \
  '.problems | any(.problemCode == $code)' <<<"$PROBLEMS_RESPONSE" >/dev/null; then
  echo "GET /api/problems did not contain $PROBLEM_CODE: $PROBLEMS_RESPONSE"
  exit 1
fi
echo "Problem list query passed"

SUBMISSIONS_RESPONSE=$(
  curl --fail --silent \
    "$BASE_URL/api/submissions" \
    -H "Authorization: Bearer $USER_TOKEN"
)
PROBLEM_ID=$(
  jq -r --arg operator_id "$AC_OPERATOR_ID" \
    '.submissions[] | select(.operatorId == $operator_id) | .problem_id' \
    <<<"$SUBMISSIONS_RESPONSE" |
    head -n 1
)

if [[ -z "$PROBLEM_ID" || "$PROBLEM_ID" == "null" ]]; then
  echo "GET /api/submissions did not contain AC submission: $SUBMISSIONS_RESPONSE"
  exit 1
fi
echo "Personal submission list query passed: problemId=$PROBLEM_ID"

PROBLEM_DETAIL=$(curl --fail --silent "$BASE_URL/api/problems/$PROBLEM_ID")
if [[ "$(jq -r '.problemCode' <<<"$PROBLEM_DETAIL")" != "$PROBLEM_CODE" ]]; then
  echo "GET problem detail returned unexpected data: $PROBLEM_DETAIL"
  exit 1
fi
echo "Problem detail query passed"

PUBLIC_USER_SUBMISSIONS=$(
  curl --fail --silent "$BASE_URL/api/users/$USER_ID/submissions"
)
if ! jq -e --arg code "$PROBLEM_CODE" \
  '.submissions | any(.problemCode == $code)' <<<"$PUBLIC_USER_SUBMISSIONS" >/dev/null; then
  echo "User submissions did not contain $PROBLEM_CODE: $PUBLIC_USER_SUBMISSIONS"
  exit 1
fi
echo "Public user submission query passed"

SOURCE_ZIP="$TEST_TMP_DIR/submission-source.zip"
curl --fail --silent \
  "$BASE_URL/api/submissions/$AC_OPERATOR_ID/source" \
  -H "Authorization: Bearer $USER_TOKEN" \
  -o "$SOURCE_ZIP"
if [[ ! -s "$SOURCE_ZIP" ]]; then
  echo "Submission source download is empty"
  exit 1
fi
echo "Submission source download passed"

TESTCASES_ZIP="$TEST_TMP_DIR/problem-testcases.zip"
curl --fail --silent \
  "$BASE_URL/api/problems/$PROBLEM_ID/testcases" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -o "$TESTCASES_ZIP"
if [[ ! -s "$TESTCASES_ZIP" ]]; then
  echo "Problem testcases download is empty"
  exit 1
fi
echo "Problem testcases download passed"

USER_TESTCASES_STATUS=$(
  curl --silent \
    -o "$TEST_TMP_DIR/user-testcases-response.json" \
    -w "%{http_code}" \
    "$BASE_URL/api/problems/$PROBLEM_ID/testcases" \
    -H "Authorization: Bearer $USER_TOKEN"
)
assert_http_status "$USER_TESTCASES_STATUS" "403" "User testcases authorization"

PROBLEM_STATS=$(curl --fail --silent "$BASE_URL/api/stats/problems/$PROBLEM_ID")
if [[ "$(jq -r '.problemCode' <<<"$PROBLEM_STATS")" != "$PROBLEM_CODE" ]] ||
  [[ "$(jq -r '.totalSubmissions' <<<"$PROBLEM_STATS")" -lt 1 ]]; then
  echo "Problem stats returned unexpected data: $PROBLEM_STATS"
  exit 1
fi
echo "Problem statistics query passed"

USER_STATS=$(curl --fail --silent "$BASE_URL/api/stats/users/$USER_ID")
if [[ "$(jq -r '.username' <<<"$USER_STATS")" != "$USER_USERNAME" ]] ||
  [[ "$(jq -r '.totalSubmissions' <<<"$USER_STATS")" -lt 1 ]]; then
  echo "User stats returned unexpected data: $USER_STATS"
  exit 1
fi
echo "User statistics query passed"

USER_DELETE_STATUS=$(
  curl --silent \
    -o "$TEST_TMP_DIR/user-delete-response.json" \
    -w "%{http_code}" \
    -X DELETE "$BASE_URL/api/problems/$PROBLEM_ID" \
    -H "Authorization: Bearer $USER_TOKEN"
)
assert_http_status "$USER_DELETE_STATUS" "403" "User problem delete authorization"

# 9. 其他最終狀態
run_status_case "WA" "$WA_SUBMISSION_ZIP"
run_status_case "CE" "$CE_SUBMISSION_ZIP"
run_status_case "SE" "$SE_SUBMISSION_ZIP"
run_status_case "RE" "$RE_SUBMISSION_ZIP"
run_status_case "TLE" "$TLE_SUBMISSION_ZIP"

# 10. Logout
LOGOUT_RESPONSE=$(
  curl --fail --silent \
    -X POST "$BASE_URL/api/users/logout" \
    -H "Authorization: Bearer $USER_TOKEN"
)
if [[ "$(jq -r '.message' <<<"$LOGOUT_RESPONSE")" != "Logout successfully" ]]; then
  echo "Logout returned unexpected data: $LOGOUT_RESPONSE"
  exit 1
fi
echo "Logout API passed"

echo "E2E 測試完成"
