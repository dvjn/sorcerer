package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/dvjn/sorcerer/pkg/models"
	"github.com/dvjn/sorcerer/pkg/storage"
)

type FileSystemStorage struct {
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

func NewFileSystemStorage(root string) (storage.Storage, error) {
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

	return &FileSystemStorage{
		root:    root,
		uploads: make(map[string]*models.UploadInfo),
	}, nil
}
