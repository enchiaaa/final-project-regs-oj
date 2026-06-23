package judge

import (
	"path/filepath"
	"testing"
)

func TestCheckTestResults_AllOutputsMatch(t *testing.T) {
	problemRoot := filepath.Join("testdata", "ac_problem")
	actualResultPath := filepath.Join("testdata", "ac_result.xml")

	verdict, err := checkTestResults(actualResultPath, problemRoot)

	if err != nil {
		t.Fatal(err)
	}

	if verdict != "AC" {
		t.Fatalf("expected AC, got %s", verdict)
	}
}

func TestCheckTestResults_OutputMismatch(t *testing.T) {
	problemRoot := filepath.Join("testdata", "ac_problem")
	actualResultPath := filepath.Join("testdata", "wa_result.xml")

	verdict, err := checkTestResults(actualResultPath, problemRoot)
	if err != nil {
		t.Fatal(err)
	}

	if verdict != "WA" {
		t.Fatalf("expected WA, got %s", verdict)
	}
}

func TestCheckTestResults_RuntimeError(t *testing.T) {
	problemRoot := filepath.Join("testdata", "ac_problem")
	actualResultPath := filepath.Join("testdata", "re_result.xml")

	verdict, err := checkTestResults(actualResultPath, problemRoot)
	if err != nil {
		t.Fatal(err)
	}

	if verdict != "RE" {
		t.Fatalf("expected RE, got %s", verdict)
	}
}

func TestCheckTestResults_EmptyTestCases(t *testing.T) {
	problemRoot := filepath.Join("testdata", "ac_problem")
	actualResultPath := filepath.Join("testdata", "empty_result.xml")

	verdict, err := checkTestResults(actualResultPath, problemRoot)
	if err != nil {
		t.Fatal(err)
	}

	if verdict != "RE" {
		t.Fatalf("expected RE, got %s", verdict)
	}
}
