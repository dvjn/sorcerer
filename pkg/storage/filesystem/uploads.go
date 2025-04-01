package filesystem

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/dvjn/sorcerer/pkg/models"
)

func (s *FileSystemStorage) uploadDir(name string) string {
	return filepath.Join(s.root, uploadsBaseDir, name)
}

func (s *FileSystemStorage) uploadPath(name, id string) string {
	return filepath.Join(s.uploadDir(name), id)
}

func (s *FileSystemStorage) InitiateUpload(name string) (string, error) {
	uploadID := fmt.Sprintf("%x", time.Now().UnixNano())

	uploadDir := s.uploadDir(name)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	uploadPath := s.uploadPath(name, uploadID)
	file, err := os.Create(uploadPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	s.uploadsMu.Lock()
	s.uploads[uploadID] = &models.UploadInfo{
		Name:      name,
		ID:        uploadID,
		Path:      uploadPath,
		Size:      0,
		Offset:    0,
		Completed: false,
	}
	s.uploadsMu.Unlock()

	return uploadID, nil
}

func (s *FileSystemStorage) UploadChunk(name, id string, content io.Reader, start int64, end int64) (int64, error) {
	s.uploadsMu.RLock()
	upload, exists := s.uploads[id]
	s.uploadsMu.RUnlock()

	if !exists {
		return 0, fmt.Errorf("upload not found")
	}

	if upload.Completed {
		return 0, fmt.Errorf("upload already completed")
	}

	if start != upload.Offset {
		if start < upload.Offset {
			return 0, fmt.Errorf("invalid range: start position %d has already been uploaded, current offset is %d", start, upload.Offset)
		}
		return 0, fmt.Errorf("invalid range: start position %d does not match current offset %d", start, upload.Offset)
	}

	file, err := os.OpenFile(upload.Path, os.O_WRONLY, 0644)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	if _, err := file.Seek(upload.Offset, io.SeekStart); err != nil {
		return 0, err
	}

	written, err := io.Copy(file, content)
	if err != nil {
		return 0, err
	}

	s.uploadsMu.Lock()
	upload.Offset += written
	upload.Size += written
	s.uploadsMu.Unlock()

	return upload.Offset, nil
}

func (s *FileSystemStorage) CompleteUpload(name, id, digest string, content io.Reader) error {
	s.uploadsMu.RLock()
	upload, exists := s.uploads[id]
	s.uploadsMu.RUnlock()

	if !exists {
		return fmt.Errorf("upload not found")
	}

	if upload.Completed {
		return fmt.Errorf("upload already completed")
	}

	if content != nil {
		file, err := os.OpenFile(upload.Path, os.O_WRONLY, 0644)
		if err != nil {
			return err
		}

		if _, err := file.Seek(upload.Offset, io.SeekStart); err != nil {
			file.Close()
			return err
		}

		written, err := io.Copy(file, content)
		if err != nil {
			file.Close()
			return err
		}

		file.Close()

		s.uploadsMu.Lock()
		upload.Offset += written
		upload.Size += written
		s.uploadsMu.Unlock()
	}

	uploadFile, err := os.Open(upload.Path)
	if err != nil {
		return err
	}
	defer uploadFile.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, uploadFile); err != nil {
		return err
	}

	actualDigest := "sha256:" + hex.EncodeToString(hasher.Sum(nil))
	if actualDigest != digest {
		return fmt.Errorf("digest mismatch: expected %s, got %s", digest, actualDigest)
	}

	blobDir := s.blobDir(name)
	if err := os.MkdirAll(blobDir, 0755); err != nil {
		return err
	}

	blobPath := s.blobPath(name, digest)
	if err := os.Rename(upload.Path, blobPath); err != nil {
		return err
	}

	s.uploadsMu.Lock()
	upload.Completed = true
	s.uploadsMu.Unlock()

	return nil
}

func (s *FileSystemStorage) GetUploadInfo(name, id string) (*models.UploadInfo, error) {
	s.uploadsMu.RLock()
	upload, exists := s.uploads[id]
	s.uploadsMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("upload not found")
	}

	return upload, nil
}
