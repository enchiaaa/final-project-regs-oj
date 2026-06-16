package utils

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

func ZipDir(srcDir string, dstPath string) error {
	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return err
	}

	zipfile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)

	err = filepath.WalkDir(srcDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if path == srcDir {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relPath)

		if entry.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if !entry.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}

			if _, err := io.Copy(writer, file); err != nil {
				file.Close()
				return err
			}

			if err := file.Close(); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		archive.Close()
		return err
	}

	return archive.Close()
}
