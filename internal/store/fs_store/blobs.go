package fs_store

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func (s *FS) blobDir(name string) string {
	return filepath.Join(s.root, blobsBaseDir, name)
}

func (s *FS) blobPath(name, digest string) string {
	return filepath.Join(s.blobDir(name), digest)
}

func (s *FS) HasBlob(name, digest string) (bool, int64, error) {
	path := s.blobPath(name, digest)
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, 0, nil
		}
		return false, 0, err
	}
	return true, info.Size(), nil
}

func (s *FS) GetBlob(name, digest string) (io.ReadCloser, int64, error) {
	path := s.blobPath(name, digest)
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, 0, fmt.Errorf("blob not found")
		}
		return nil, 0, err
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, 0, err
	}

	return file, info.Size(), nil
}

func (s *FS) PutBlob(name, digest string, content io.Reader) error {
	dir := s.blobDir(name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	path := s.blobPath(name, digest)

	tempFile, err := os.CreateTemp(dir, "temp-blob-*")
	if err != nil {
		return err
	}
	tempPath := tempFile.Name()

	defer func() {
		tempFile.Close()
		os.Remove(tempPath)
	}()

	hasher := sha256.New()
	writer := io.MultiWriter(tempFile, hasher)

	if _, err := io.Copy(writer, content); err != nil {
		return err
	}

	actualDigest := "sha256:" + hex.EncodeToString(hasher.Sum(nil))
	if actualDigest != digest {
		return fmt.Errorf("digest mismatch: expected %s, got %s", digest, actualDigest)
	}

	if err := tempFile.Close(); err != nil {
		return err
	}

	if err := os.Rename(tempPath, path); err != nil {
		return err
	}

	return nil
}

func (s *FS) DeleteBlob(name, digest string) error {
	path := s.blobPath(name, digest)
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("blob not found")
		}
		return err
	}
	return nil
}

func (s *FS) MountBlob(fromName, toName, digest string) error {
	sourcePath := s.blobPath(fromName, digest)
	if _, err := os.Stat(sourcePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("source blob not found")
		}
		return err
	}

	destDir := s.blobDir(toName)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	destPath := s.blobPath(toName, digest)

	if err := os.Link(sourcePath, destPath); err != nil {
		sourceFile, err := os.Open(sourcePath)
		if err != nil {
			return err
		}
		defer sourceFile.Close()

		destFile, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer destFile.Close()

		if _, err := io.Copy(destFile, sourceFile); err != nil {
			os.Remove(destPath)
			return err
		}
	}

	return nil
}
