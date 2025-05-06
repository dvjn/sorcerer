package fs_storage

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/dvjn/sorcerer/internal/config"
	"github.com/dvjn/sorcerer/internal/model"
)

type FS struct {
	root      string
	uploadsMu sync.RWMutex
	uploads   map[string]*model.UploadInfo
}

const (
	blobsBaseDir     = "blobs"
	manifestsBaseDir = "manifests"
	uploadsBaseDir   = "uploads"
	tagsBaseDir      = "tags"
	referrersBaseDir = "referrers"
)

func New(c *config.StorageConfig) (*FS, error) {
	for _, dir := range []string{
		filepath.Join(c.Path, blobsBaseDir),
		filepath.Join(c.Path, manifestsBaseDir),
		filepath.Join(c.Path, uploadsBaseDir),
		filepath.Join(c.Path, tagsBaseDir),
		filepath.Join(c.Path, referrersBaseDir),
	} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return &FS{
		root:    c.Path,
		uploads: make(map[string]*model.UploadInfo),
	}, nil
}
