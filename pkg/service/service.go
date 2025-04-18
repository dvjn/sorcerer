package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

type Service struct {
	storage Storage
}

func NewService(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

func (s *Service) ApiVersionCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Service) CheckBlobExists(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repository := chi.URLParam(r, "repository")
	name := owner + "/" + repository
	digest := chi.URLParam(r, "digest")

	exists, size, err := s.storage.HasBlob(name, digest)
	if err != nil {
		sendError(w, errBlobUnknown.WithStatus(http.StatusInternalServerError), err.Error())
		return
	}

	if !exists {
		sendError(w, errBlobUnknown, "Blob not found")
		return
	}

	w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
	w.Header().Set("Docker-Content-Digest", digest)
	w.WriteHeader(http.StatusOK)
}

func (s *Service) GetBlob(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repository := chi.URLParam(r, "repository")
	name := owner + "/" + repository
	digest := chi.URLParam(r, "digest")

	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" {
		fmt.Println("Range request not fully implemented, returning full blob:", rangeHeader)
	}

	blob, size, err := s.storage.GetBlob(name, digest)
	if err != nil {
		sendError(w, errBlobUnknown, err.Error())
		return
	}
	defer blob.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
	w.Header().Set("Docker-Content-Digest", digest)
	w.WriteHeader(http.StatusOK)

	if _, err := io.Copy(w, blob); err != nil {
		fmt.Printf("Error streaming blob: %v\n", err)
	}
}

func (s *Service) DeleteBlob(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repository := chi.URLParam(r, "repository")
	name := owner + "/" + repository
	digest := chi.URLParam(r, "digest")

	err := s.storage.DeleteBlob(name, digest)
	if err != nil {
		sendError(w, errBlobUnknown, err.Error())
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (s *Service) InitiateUpload(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repository := chi.URLParam(r, "repository")
	name := owner + "/" + repository

	if digest := r.URL.Query().Get("digest"); digest != "" {
		err := s.storage.PutBlob(name, digest, r.Body)
		if err != nil {
			sendError(w, errBlobUploadInvalid, err.Error())
			return
		}

		location := fmt.Sprintf("/v2/%s/%s/blobs/%s", owner, repository, digest)
		w.Header().Set("Location", location)
		w.Header().Set("Docker-Content-Digest", digest)
		w.WriteHeader(http.StatusCreated)
		return
	}

	if digest := r.URL.Query().Get("mount"); digest != "" {
		from := r.URL.Query().Get("from")
		if from != "" {
			err := s.storage.MountBlob(from, name, digest)
			if err != nil {
				uploadID, err := s.storage.InitiateUpload(name)
				if err != nil {
					sendError(w, errBlobUploadInvalid, err.Error())
					return
				}

				location := fmt.Sprintf("/v2/%s/%s/blobs/uploads/%s", owner, repository, uploadID)
				w.Header().Set("Location", location)
				w.Header().Set("Range", "0-0")
				w.WriteHeader(http.StatusAccepted)
				return
			}

			location := fmt.Sprintf("/v2/%s/%s/blobs/%s", owner, repository, digest)
			w.Header().Set("Location", location)
			w.Header().Set("Docker-Content-Digest", digest)
			w.WriteHeader(http.StatusCreated)
			return
		}
	}

	uploadID, err := s.storage.InitiateUpload(name)
	if err != nil {
		sendError(w, errBlobUploadInvalid, err.Error())
		return
	}

	location := fmt.Sprintf("/v2/%s/%s/blobs/uploads/%s", owner, repository, uploadID)
	w.Header().Set("Location", location)
	w.Header().Set("Range", "0-0")
	w.Header().Set("Oci-Chunk-Min-Length", "1024")
	w.WriteHeader(http.StatusAccepted)
}

func (s *Service) UploadBlobChunk(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repository := chi.URLParam(r, "repository")
	name := owner + "/" + repository
	reference := chi.URLParam(r, "reference")
	contentRange := r.Header.Get("Content-Range")

	info, err := s.storage.GetUploadInfo(name, reference)
	if err != nil {
		sendError(w, errBlobUploadUnknown, err.Error())
		return
	}

	var start, end int64
	if contentRange != "" {
		_, err := fmt.Sscanf(contentRange, "%d-%d", &start, &end)
		if err != nil {
			sendError(w, errBlobUploadInvalid, "Invalid content range")
			return
		}

		if start != info.Offset {
			errorMsg := ""
			if start < info.Offset {
				errorMsg = fmt.Sprintf("Range start position %d is less than current offset %d", start, info.Offset)
			} else {
				errorMsg = fmt.Sprintf("Range start position %d does not match current offset %d", start, info.Offset)
			}
			sendError(w, errRangeInvalid, errorMsg)
			return
		}
	} else {
		start = info.Offset
		end = start + r.ContentLength - 1
	}

	newOffset, err := s.storage.UploadChunk(name, reference, r.Body, start, end)
	if err != nil {
		if strings.Contains(err.Error(), "invalid range") {
			sendError(w, errRangeInvalid, err.Error())
			return
		} else {
			sendError(w, errBlobUploadInvalid, err.Error())
			return
		}
	}

	location := fmt.Sprintf("/v2/%s/%s/blobs/uploads/%s", owner, repository, reference)
	w.Header().Set("Location", location)
	w.Header().Set("Range", fmt.Sprintf("0-%d", newOffset-1))
	w.WriteHeader(http.StatusAccepted)
}

func (s *Service) CompleteUpload(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repository := chi.URLParam(r, "repository")
	name := owner + "/" + repository
	reference := chi.URLParam(r, "reference")
	digest := r.URL.Query().Get("digest")

	if digest == "" {
		sendError(w, errDigestInvalid, "Digest parameter missing")
		return
	}

	var content io.Reader
	if r.ContentLength > 0 {
		content = r.Body
	}

	err := s.storage.CompleteUpload(name, reference, digest, content)
	if err != nil {
		sendError(w, errBlobUploadInvalid, err.Error())
		return
	}

	location := fmt.Sprintf("/v2/%s/%s/blobs/%s", owner, repository, digest)
	w.Header().Set("Location", location)
	w.Header().Set("Docker-Content-Digest", digest)
	w.WriteHeader(http.StatusCreated)
}

func (s *Service) GetBlobUploadStatus(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repository := chi.URLParam(r, "repository")
	name := owner + "/" + repository
	reference := chi.URLParam(r, "reference")

	info, err := s.storage.GetUploadInfo(name, reference)
	if err != nil {
		sendError(w, errBlobUploadUnknown, err.Error())
		return
	}

	location := fmt.Sprintf("/v2/%s/%s/blobs/uploads/%s", owner, repository, reference)
	w.Header().Set("Location", location)
	w.Header().Set("Range", fmt.Sprintf("0-%d", info.Offset-1))
	w.WriteHeader(http.StatusNoContent)
}

func (s *Service) CheckManifestExists(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repository := chi.URLParam(r, "repository")
	name := owner + "/" + repository
	reference := chi.URLParam(r, "reference")

	var exists bool
	var size int64
	var digest string
	var err error

	if !strings.HasPrefix(reference, "sha256:") {
		exists, _, digest, err = s.storage.HasManifest(name, reference)
		if err != nil {
			sendError(w, errManifestUnknown.WithStatus(http.StatusInternalServerError), err.Error())
			return
		}

		if exists {
			content, _, err := s.storage.GetManifest(name, digest)
			if err != nil {
				sendError(w, errManifestUnknown.WithStatus(http.StatusInternalServerError), err.Error())
				return
			}
			size = int64(len(content))
		}
	} else {
		exists, size, digest, err = s.storage.HasManifest(name, reference)
		if err != nil {
			sendError(w, errManifestUnknown.WithStatus(http.StatusInternalServerError), err.Error())
			return
		}
	}

	if !exists {
		sendError(w, errManifestUnknown, "Manifest not found")
		return
	}

	contentType := "application/vnd.oci.image.manifest.v1+json"
	if strings.HasSuffix(reference, ".json") || strings.HasSuffix(reference, ".pretty") {
		contentType = "application/json"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
	w.Header().Set("Docker-Content-Digest", digest)
	w.WriteHeader(http.StatusOK)
}

func (s *Service) GetManifest(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repository := chi.URLParam(r, "repository")
	name := owner + "/" + repository
	reference := chi.URLParam(r, "reference")

	content, digest, err := s.storage.GetManifest(name, reference)
	if err != nil {
		sendError(w, errManifestUnknown, err.Error())
		return
	}

	mediaType := "application/vnd.oci.image.manifest.v1+json"
	var manifest map[string]any
	if err := json.Unmarshal(content, &manifest); err == nil {
		if mt, ok := manifest["mediaType"].(string); ok && mt != "" {
			mediaType = mt
		}
	}

	w.Header().Set("Content-Type", mediaType)
	w.Header().Set("Docker-Content-Digest", digest)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

func (s *Service) PutManifest(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repository := chi.URLParam(r, "repository")
	name := owner + "/" + repository
	reference := chi.URLParam(r, "reference")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		sendError(w, errManifestInvalid, "Failed to read manifest")
		return
	}

	digest, err := s.storage.PutManifest(name, reference, body)
	if err != nil {
		sendError(w, errManifestInvalid, err.Error())
		return
	}

	var manifest map[string]any
	if err := json.Unmarshal(body, &manifest); err == nil {
		if subject, ok := manifest["subject"].(map[string]any); ok {
			if subjectDigest, ok := subject["digest"].(string); ok {
				if err := s.storage.UpdateReferrers(name, subjectDigest, body); err != nil {
					fmt.Printf("Error updating referrers: %v\n", err)
				}
				w.Header().Set("OCI-Subject", subjectDigest)
			}
		}
	}

	location := fmt.Sprintf("/v2/%s/%s/manifests/%s", owner, repository, reference)
	w.Header().Set("Location", location)
	w.Header().Set("Docker-Content-Digest", digest)
	w.WriteHeader(http.StatusCreated)
}

func (s *Service) DeleteManifest(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repository := chi.URLParam(r, "repository")
	name := owner + "/" + repository
	reference := chi.URLParam(r, "reference")

	if strings.HasPrefix(reference, "sha256:") {
		content, _, err := s.storage.GetManifest(name, reference)
		if err == nil {
			var manifest map[string]any
			if err := json.Unmarshal(content, &manifest); err == nil {
				if subject, ok := manifest["subject"].(map[string]any); ok {
					if subjectDigest, ok := subject["digest"].(string); ok {
						if err := s.storage.RemoveReferrer(name, subjectDigest, reference); err != nil {
							fmt.Printf("Error removing from referrers: %v\n", err)
						}
					}
				}
			}
		}
	}

	err := s.storage.DeleteManifest(name, reference)
	if err != nil {
		sendError(w, errManifestUnknown, err.Error())
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (s *Service) ListTags(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repository := chi.URLParam(r, "repository")
	name := owner + "/" + repository

	tags, err := s.storage.ListTags(name)
	if err != nil {
		sendError(w, errRepositoryUnknown, err.Error())
		return
	}

	n := r.URL.Query().Get("n")
	last := r.URL.Query().Get("last")

	var limit int
	if n != "" {
		fmt.Sscanf(n, "%d", &limit)
	}

	if last != "" {
		filteredTags := []string{}
		pastLast := false
		for _, tag := range tags {
			if pastLast {
				filteredTags = append(filteredTags, tag)
				if limit > 0 && len(filteredTags) >= limit {
					break
				}
			} else if tag == last {
				pastLast = true
			}
		}
		tags = filteredTags
	} else if limit > 0 && len(tags) > limit {
		tags = tags[:limit]
	}

	response := struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}{
		Name: name,
		Tags: tags,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (s *Service) ListReferrers(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repository := chi.URLParam(r, "repository")
	name := owner + "/" + repository
	digest := chi.URLParam(r, "digest")
	artifactType := r.URL.Query().Get("artifactType")

	content, err := s.storage.GetReferrers(name, digest, artifactType)
	if err != nil {
		sendError(w, errManifestUnknown, err.Error())
		return
	}

	if artifactType != "" {
		w.Header().Set("OCI-Filters-Applied", "artifactType")
	}

	w.Header().Set("Content-Type", "application/vnd.oci.image.index.v1+json")
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}
