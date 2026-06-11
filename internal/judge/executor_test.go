// executor 的測試 
package judge

import (
	"os"
	"path/filepath"
	"testing"
)

// 測試 checkTestResults：AC 情況
func TestCheckTestResultsAC(t *testing.T) {
	dir := t.TempDir()
	resultPath := filepath.Join(dir, "result.xml")

	xmlContent := `<testsuite failures="0"></testsuite>`

	if err := os.WriteFile(resultPath, []byte(xmlContent), 0644); err != nil {
		t.Fatalf("failed to write result.xml: %v", err)
	}

	ok, err := checkTestResults(resultPath)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !ok {
		t.Fatalf("expected result to pass")
	}
}

// 測試 checkTestResults：WA 情況
func TestCheckTestResultsWA(t *testing.T) {
	dir := t.TempDir()
	resultPath := filepath.Join(dir, "result.xml")

	xmlContent := `<testsuite failures="1"></testsuite>`

	if err := os.WriteFile(resultPath, []byte(xmlContent), 0644); err != nil {
		t.Fatalf("failed to write result.xml: %v", err)
	}

	ok, err := checkTestResults(resultPath)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if ok {
		t.Fatalf("expected result to not pass")
	}
}

// 測試 checkTestResults：XML invalid 情況
func TestCheckTestResultsInvalidXML(t *testing.T) {
	dir := t.TempDir()
	resultPath := filepath.Join(dir, "result.xml")

	xmlContent := `<testsuite failures="0"`

	if err := os.WriteFile(resultPath, []byte(xmlContent), 0644); err != nil {
		t.Fatalf("failed to write result.xml: %v", err)
	}

	ok, err := checkTestResults(resultPath)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if ok {
		t.Fatalf("expected result to not pass")
	}
}
