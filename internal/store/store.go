package store

import (
	"io"

	"github.com/dvjn/sorcerer/internal/config"
	"github.com/dvjn/sorcerer/internal/model"
	fs_store "github.com/dvjn/sorcerer/internal/store/fs_store"
)

type Store interface {
	HasBlob(name, digest string) (bool, int64, error)
	GetBlob(name, digest string) (io.ReadCloser, int64, error)
	PutBlob(name, digest string, content io.Reader) error
	DeleteBlob(name, digest string) error
	MountBlob(fromName, toName, digest string) error

	HasManifest(name, reference string) (bool, int64, string, error)
	GetManifest(name, reference string) ([]byte, string, error)
	PutManifest(name, reference string, content []byte) (string, error)
	DeleteManifest(name, reference string) error

	ListTags(name string) ([]string, error)

	GetReferrers(name, digest string, artifactType string) ([]byte, error)
	UpdateReferrers(name, digest string, manifest []byte) error
	RemoveReferrer(name, digest, manifestDigest string) error

	InitiateUpload(name string) (string, error)
	UploadChunk(name, id string, content io.Reader, start int64, end int64) (int64, error)
	CompleteUpload(name, id, digest string, content io.Reader) error
	GetUploadInfo(name, id string) (*model.UploadInfo, error)
}

func New(c *config.StoreConfig) (Store, error) {
	return fs_store.New(c)
}
