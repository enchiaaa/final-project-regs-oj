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

type TestSuite struct {
	TestCases []TestCase `xml:"testcase"`
}

type TestCase struct {
	Name      string   `xml:"name,attr"`
	SystemOut string   `xml:"system-out"`
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
func checkTestResults(resultPath string, problemPath string) (bool, error) {
	// 取得 result.xml 裡面的每個 testcase 的輸出結果
	testCases, err := parseTestResult(resultPath)
	if err != nil{
		return false, err
	}
	testResults, err := loadExpectedOutputs(problemPath)
	if err != nil{
		return false, err
	}

	for _, testCase := range testCases{
		if normalizeOutput(testCase.SystemOut) != normalizeOutput(testResults[testCase.Name]){
			return false, nil
		}
	}

	return true, nil
}

// parse result.xml，取得每個 testcase 的輸出結果
func parseTestResult(resultPath string) ([]TestCase, error) {
	// 讀取資料
	resultFile, err := os.ReadFile(resultPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read result file: %w", err)
	}

	// 取得 name 和 SystemOut 的內容
	var suite TestSuite
	if err := xml.Unmarshal(resultFile, &suite); err != nil {
		return nil, fmt.Errorf("failed to parse result file: %w", err)
	}

	return suite.TestCases, nil
}

// 從 problemPath 的 settings.yaml 找到每個 case 的正確輸出結果並回傳
func loadExpectedOutputs(problemPath string) (map[string]string, error) {
	// 讀取 settings.yaml
	settingsPath := filepath.Join(problemPath, "settings.yaml")

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
		expectedPath := filepath.Join(problemPath, preset.Expected.Value)

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