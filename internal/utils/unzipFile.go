// unzipFile function

package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// 解壓縮 zip 檔案到指定目錄
// 程式碼參考自 https://golang.cafe/blog/golang-unzip-file-example
func UnzipFile(src string, dst string) error {
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