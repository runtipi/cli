package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ExtractTarGz(archivePath, destDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open archive file: %v", err)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %v", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	absDestDir, err := filepath.Abs(destDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path of destination directory: %v", err)
	}

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("failed to read tar entry: %v", err)
		}

		target := filepath.Join(destDir, header.Name)
		absTarget, err := filepath.Abs(target)
		if err != nil {
			return fmt.Errorf("failed to get absolute path of target: %v", err)
		}

		if !strings.HasPrefix(absTarget, absDestDir) {
			return fmt.Errorf("path traversal attack detected: %s attempts to escape to %s", header.Name, absTarget)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %v", err)
			}
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("failed to create directory: %v", err)
			}

			if err := os.Symlink(header.Linkname, target); err != nil {
				return fmt.Errorf("failed to create symbolic link: %v", err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("failed to create directory: %v", err)
			}

			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file: %v", err)
			}

			if _, err := io.Copy(f, tarReader); err != nil {
				f.Close()
				return fmt.Errorf("failed to extract file: %v", err)
			}
			f.Close()
		}
	}

	return nil
}
