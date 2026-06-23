package utils

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestUnzipFileRejectsZipSlip(t *testing.T) {
	// 1. 建立測試檔案
	dir := t.TempDir()
	dst := t.TempDir()
	zipPath := filepath.Join(dir, "evil.zip")

	// 2. 建立惡意 ZIP
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}

	zipWriter := zip.NewWriter(zipFile)

	entry, err := zipWriter.Create("../../evil.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("evil content")); err != nil {
		t.Fatal(err)
	}
	if err := zipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := zipFile.Close(); err != nil {
		t.Fatal(err)
	}

	// 3. 呼叫 UnzipFile(zipPath, destination)
	err = UnzipFile(zipPath, dst)

	// 4. 確認有回傳 error
	if err == nil {
		t.Fatal("expected Zip Slip archive to be rejected")
	}

	// 5. 確認檔案沒有被寫到目的資料夾外
	evilPath := filepath.Clean(filepath.Join(dst, "../../evil.txt"))
	if _, err := os.Stat(evilPath); !os.IsNotExist(err) {
		t.Fatal("evil.txt should not be created outside destination")
	}
}
