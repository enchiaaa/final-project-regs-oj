// 負責 result.xml + expected output 比對
package judge

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type CTestSuiteResult struct {
	TestCases []CTestCaseResult `xml:"testcase"`
}

type CTestCaseResult struct {
	Name      string        `xml:"name,attr"`
	Status    string        `xml:"status,attr"`
	Failure   *CTestFailure `xml:"failure"`
	SystemOut string        `xml:"system-out"`
}
type CTestFailure struct {
	Message string `xml:"message,attr"`
}
type ProblemSettings struct {
	Presets []Preset `yaml:"presets"`
}

type Preset struct {
	Expected Expected `yaml:"expected"`
}

type Expected struct {
	Type  string `yaml:"type"`
	Value string `yaml:"value"`
}

// 確認評測結果
func checkTestResults(actualResultPath string, problemRoot string) (string, error) {
	// 取得 submission 的 result.xml 裡每個 testcase 的實際輸出
	actualTestCases, err := parseActualTestResults(actualResultPath)
	if err != nil {
		return "", err
	}

	// 取得題目提供的每個 testcase 預期輸出
	expectedOutputs, err := loadExpectedOutputs(problemRoot)
	if err != nil {
		return "", err
	}

	for _, actualTestCase := range actualTestCases {
		if isRuntimeError(actualTestCase) {
			return "RE", nil
		}

		expectedOutput, exists := expectedOutputs[actualTestCase.Name]
		if !exists {
			return "", fmt.Errorf("expected output not found for testcase %s", actualTestCase.Name)
		}

		if normalizeOutput(actualTestCase.SystemOut) != normalizeOutput(expectedOutput) {
			return "WA", nil
		}
	}

	if len(actualTestCases) == 0 {
		return "RE", nil
	}

	return "AC", nil
}

// 解析 submission 的 result.xml，取得每個 testcase 的實際輸出
func parseActualTestResults(actualResultPath string) ([]CTestCaseResult, error) {
	// 讀取資料
	resultFile, err := os.ReadFile(actualResultPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read result file: %w", err)
	}

	// 取得 name 和 SystemOut 的內容
	var suite CTestSuiteResult
	if err := xml.Unmarshal(resultFile, &suite); err != nil {
		return nil, fmt.Errorf("failed to parse result file: %w", err)
	}

	return suite.TestCases, nil
}

// 從題目根目錄的 settings.yaml 找到每個 testcase 的預期輸出
func loadExpectedOutputs(problemRoot string) (map[string]string, error) {
	// 讀取 settings.yaml
	settingsPath := filepath.Join(problemRoot, "settings.yaml")

	settingsFile, err := os.ReadFile(settingsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read settings.yaml: %w", err)
	}

	// parse settings.yaml
	var settings ProblemSettings
	if err := yaml.Unmarshal(settingsFile, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse settings.yaml: %w", err)
	}

	// 取得每個 case 的正確輸出結果
	expectedOutputs := make(map[string]string)

	for _, preset := range settings.Presets {
		// 正確輸出的位置
		expectedPath := filepath.Join(problemRoot, preset.Expected.Value)

		// 讀取正確輸出
		content, err := os.ReadFile(expectedPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read expected file %s: %w", expectedPath, err)
		}

		// 儲存到 expectedOutputs
		testCaseName := filepath.Base(preset.Expected.Value)
		expectedOutputs[testCaseName] = string(content)
	}

	return expectedOutputs, nil
}

// 統一字串格式
func normalizeOutput(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.TrimSpace(s)
	return s
}

// 查看 ErrorMessages，確認是否為 RE
func isRuntimeError(testCase CTestCaseResult) bool {
	if testCase.Failure == nil {
		return false
	}

	message := strings.ToLower(testCase.Failure.Message)

	runtimeErrorMessages := []string{
		"subprocess aborted",
		"segfault",
		"segmentation fault",
	}

	for _, runtimeErrorMessage := range runtimeErrorMessages {
		if strings.Contains(message, runtimeErrorMessage) {
			return true
		}
	}

	return false
}
