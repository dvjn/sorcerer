package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/dvjn/sorcerer/internal/models"
)

type Storage struct {
	root      string
	uploadsMu sync.RWMutex
	uploads   map[string]*models.UploadInfo
}

const (
	blobsBaseDir     = "blobs"
	manifestsBaseDir = "manifests"
	uploadsBaseDir   = "uploads"
	tagsBaseDir      = "tags"
	referrersBaseDir = "referrers"
)

func NewStorage(root string) (*Storage, error) {
	for _, dir := range []string{
		filepath.Join(root, blobsBaseDir),
		filepath.Join(root, manifestsBaseDir),
		filepath.Join(root, uploadsBaseDir),
		filepath.Join(root, tagsBaseDir),
		filepath.Join(root, referrersBaseDir),
	} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return &Storage{
		root:    root,
		uploads: make(map[string]*models.UploadInfo),
	}, nil
}
