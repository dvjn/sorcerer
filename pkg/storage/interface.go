package storage

import (
	"io"

	"github.com/dvjn/sorcerer/pkg/models"
)

type Storage interface {
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
	GetUploadInfo(name, id string) (*models.UploadInfo, error)
}
