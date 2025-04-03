package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/dvjn/sorcerer/pkg/storage"
	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	storage storage.Storage
}

func NewHandlers(storage storage.Storage) *Handlers {
	return &Handlers{
		storage: storage,
	}
}

func (h *Handlers) ApiVersionCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) CheckBlobExists(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	digest := chi.URLParam(r, "digest")

	exists, size, err := h.storage.HasBlob(name, digest)
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

func (h *Handlers) GetBlob(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	digest := chi.URLParam(r, "digest")

	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" {
		fmt.Println("Range request not fully implemented, returning full blob:", rangeHeader)
	}

	blob, size, err := h.storage.GetBlob(name, digest)
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

func (h *Handlers) DeleteBlob(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	digest := chi.URLParam(r, "digest")

	err := h.storage.DeleteBlob(name, digest)
	if err != nil {
		sendError(w, errBlobUnknown, err.Error())
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handlers) InitiateUpload(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	if digest := r.URL.Query().Get("digest"); digest != "" {
		err := h.storage.PutBlob(name, digest, r.Body)
		if err != nil {
			sendError(w, errBlobUploadInvalid, err.Error())
			return
		}

		location := fmt.Sprintf("/v2/%s/blobs/%s", name, digest)
		w.Header().Set("Location", location)
		w.Header().Set("Docker-Content-Digest", digest)
		w.WriteHeader(http.StatusCreated)
		return
	}

	if digest := r.URL.Query().Get("mount"); digest != "" {
		from := r.URL.Query().Get("from")
		if from != "" {
			err := h.storage.MountBlob(from, name, digest)
			if err != nil {
				uploadID, err := h.storage.InitiateUpload(name)
				if err != nil {
					sendError(w, errBlobUploadInvalid, err.Error())
					return
				}

				location := fmt.Sprintf("/v2/%s/blobs/uploads/%s", name, uploadID)
				w.Header().Set("Location", location)
				w.Header().Set("Range", "0-0")
				w.WriteHeader(http.StatusAccepted)
				return
			}

			location := fmt.Sprintf("/v2/%s/blobs/%s", name, digest)
			w.Header().Set("Location", location)
			w.Header().Set("Docker-Content-Digest", digest)
			w.WriteHeader(http.StatusCreated)
			return
		}
	}

	uploadID, err := h.storage.InitiateUpload(name)
	if err != nil {
		sendError(w, errBlobUploadInvalid, err.Error())
		return
	}

	location := fmt.Sprintf("/v2/%s/blobs/uploads/%s", name, uploadID)
	w.Header().Set("Location", location)
	w.Header().Set("Range", "0-0")
	w.Header().Set("Oci-Chunk-Min-Length", "1024")
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handlers) UploadBlobChunk(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	reference := chi.URLParam(r, "reference")
	contentRange := r.Header.Get("Content-Range")

	info, err := h.storage.GetUploadInfo(name, reference)
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

	newOffset, err := h.storage.UploadChunk(name, reference, r.Body, start, end)
	if err != nil {
		if strings.Contains(err.Error(), "invalid range") {
			sendError(w, errRangeInvalid, err.Error())
			return
		} else {
			sendError(w, errBlobUploadInvalid, err.Error())
			return
		}
	}

	location := fmt.Sprintf("/v2/%s/blobs/uploads/%s", name, reference)
	w.Header().Set("Location", location)
	w.Header().Set("Range", fmt.Sprintf("0-%d", newOffset-1))
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handlers) CompleteUpload(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
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

	err := h.storage.CompleteUpload(name, reference, digest, content)
	if err != nil {
		sendError(w, errBlobUploadInvalid, err.Error())
		return
	}

	location := fmt.Sprintf("/v2/%s/blobs/%s", name, digest)
	w.Header().Set("Location", location)
	w.Header().Set("Docker-Content-Digest", digest)
	w.WriteHeader(http.StatusCreated)
}

func (h *Handlers) GetBlobUploadStatus(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	reference := chi.URLParam(r, "reference")

	info, err := h.storage.GetUploadInfo(name, reference)
	if err != nil {
		sendError(w, errBlobUploadUnknown, err.Error())
		return
	}

	location := fmt.Sprintf("/v2/%s/blobs/uploads/%s", name, reference)
	w.Header().Set("Location", location)
	w.Header().Set("Range", fmt.Sprintf("0-%d", info.Offset-1))
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) CheckManifestExists(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	reference := chi.URLParam(r, "reference")

	var exists bool
	var size int64
	var digest string
	var err error

	if !strings.HasPrefix(reference, "sha256:") {
		exists, _, digest, err = h.storage.HasManifest(name, reference)
		if err != nil {
			sendError(w, errManifestUnknown.WithStatus(http.StatusInternalServerError), err.Error())
			return
		}

		if exists {
			content, _, err := h.storage.GetManifest(name, digest)
			if err != nil {
				sendError(w, errManifestUnknown.WithStatus(http.StatusInternalServerError), err.Error())
				return
			}
			size = int64(len(content))
		}
	} else {
		exists, size, digest, err = h.storage.HasManifest(name, reference)
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

func (h *Handlers) GetManifest(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	reference := chi.URLParam(r, "reference")

	content, digest, err := h.storage.GetManifest(name, reference)
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

func (h *Handlers) PutManifest(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	reference := chi.URLParam(r, "reference")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		sendError(w, errManifestInvalid, "Failed to read manifest")
		return
	}

	digest, err := h.storage.PutManifest(name, reference, body)
	if err != nil {
		sendError(w, errManifestInvalid, err.Error())
		return
	}

	var manifest map[string]any
	if err := json.Unmarshal(body, &manifest); err == nil {
		if subject, ok := manifest["subject"].(map[string]any); ok {
			if subjectDigest, ok := subject["digest"].(string); ok {
				if err := h.storage.UpdateReferrers(name, subjectDigest, body); err != nil {
					fmt.Printf("Error updating referrers: %v\n", err)
				}
				w.Header().Set("OCI-Subject", subjectDigest)
			}
		}
	}

	location := fmt.Sprintf("/v2/%s/manifests/%s", name, reference)
	w.Header().Set("Location", location)
	w.Header().Set("Docker-Content-Digest", digest)
	w.WriteHeader(http.StatusCreated)
}

func (h *Handlers) DeleteManifest(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	reference := chi.URLParam(r, "reference")

	if strings.HasPrefix(reference, "sha256:") {
		content, _, err := h.storage.GetManifest(name, reference)
		if err == nil {
			var manifest map[string]any
			if err := json.Unmarshal(content, &manifest); err == nil {
				if subject, ok := manifest["subject"].(map[string]any); ok {
					if subjectDigest, ok := subject["digest"].(string); ok {
						if err := h.storage.RemoveReferrer(name, subjectDigest, reference); err != nil {
							fmt.Printf("Error removing from referrers: %v\n", err)
						}
					}
				}
			}
		}
	}

	err := h.storage.DeleteManifest(name, reference)
	if err != nil {
		sendError(w, errManifestUnknown, err.Error())
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handlers) ListTags(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	tags, err := h.storage.ListTags(name)
	if err != nil {
		sendError(w, errNameUnknown, err.Error())
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

func (h *Handlers) ListReferrers(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	digest := chi.URLParam(r, "digest")
	artifactType := r.URL.Query().Get("artifactType")

	content, err := h.storage.GetReferrers(name, digest, artifactType)
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
