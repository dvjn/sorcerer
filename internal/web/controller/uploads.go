package controller

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

func (c *Controller) InitiateUpload(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repository := chi.URLParam(r, "repository")
	name := owner + "/" + repository

	if digest := r.URL.Query().Get("digest"); digest != "" {
		err := c.store.PutBlob(name, digest, r.Body)
		if err != nil {
			sendError(w, http.StatusBadRequest, errBlobUploadInvalid, err.Error())
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
			err := c.store.MountBlob(from, name, digest)
			if err != nil {
				uploadID, err := c.store.InitiateUpload(name)
				if err != nil {
					sendError(w, http.StatusBadRequest, errBlobUploadInvalid, err.Error())
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

	uploadID, err := c.store.InitiateUpload(name)
	if err != nil {
		sendError(w, http.StatusBadRequest, errBlobUploadInvalid, err.Error())
		return
	}

	location := fmt.Sprintf("/v2/%s/%s/blobs/uploads/%s", owner, repository, uploadID)
	w.Header().Set("Location", location)
	w.Header().Set("Range", "0-0")
	w.Header().Set("Oci-Chunk-Min-Length", "1024")
	w.WriteHeader(http.StatusAccepted)
}

func (c *Controller) UploadBlobChunk(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repository := chi.URLParam(r, "repository")
	name := owner + "/" + repository
	reference := chi.URLParam(r, "reference")
	contentRange := r.Header.Get("Content-Range")

	info, err := c.store.GetUploadInfo(name, reference)
	if err != nil {
		sendError(w, http.StatusNotFound, errBlobUploadUnknown, err.Error())
		return
	}

	var start, end int64
	if contentRange != "" {
		_, err := fmt.Sscanf(contentRange, "%d-%d", &start, &end)
		if err != nil {
			sendError(w, http.StatusBadRequest, errBlobUploadInvalid, "Invalid content range")
			return
		}

		if start != info.Offset {
			errorMsg := ""
			if start < info.Offset {
				errorMsg = fmt.Sprintf("Range start position %d is less than current offset %d", start, info.Offset)
			} else {
				errorMsg = fmt.Sprintf("Range start position %d does not match current offset %d", start, info.Offset)
			}
			sendError(w, http.StatusRequestedRangeNotSatisfiable, errRangeInvalid, errorMsg)
			return
		}
	} else {
		start = info.Offset
		end = start + r.ContentLength - 1
	}

	newOffset, err := c.store.UploadChunk(name, reference, r.Body, start, end)
	if err != nil {
		if strings.Contains(err.Error(), "invalid range") {
			sendError(w, http.StatusRequestedRangeNotSatisfiable, errRangeInvalid, err.Error())
			return
		} else {
			sendError(w, http.StatusBadRequest, errBlobUploadInvalid, err.Error())
			return
		}
	}

	location := fmt.Sprintf("/v2/%s/%s/blobs/uploads/%s", owner, repository, reference)
	w.Header().Set("Location", location)
	w.Header().Set("Range", fmt.Sprintf("0-%d", newOffset-1))
	w.WriteHeader(http.StatusAccepted)
}

func (c *Controller) CompleteUpload(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repository := chi.URLParam(r, "repository")
	name := owner + "/" + repository
	reference := chi.URLParam(r, "reference")
	digest := r.URL.Query().Get("digest")

	if digest == "" {
		sendError(w, http.StatusBadRequest, errDigestInvalid, "Digest parameter missing")
		return
	}

	var content io.Reader
	if r.ContentLength > 0 {
		content = r.Body
	}

	err := c.store.CompleteUpload(name, reference, digest, content)
	if err != nil {
		sendError(w, http.StatusBadRequest, errBlobUploadInvalid, err.Error())
		return
	}

	location := fmt.Sprintf("/v2/%s/%s/blobs/%s", owner, repository, digest)
	w.Header().Set("Location", location)
	w.Header().Set("Docker-Content-Digest", digest)
	w.WriteHeader(http.StatusCreated)
}

func (c *Controller) GetBlobUploadStatus(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repository := chi.URLParam(r, "repository")
	name := owner + "/" + repository
	reference := chi.URLParam(r, "reference")

	info, err := c.store.GetUploadInfo(name, reference)
	if err != nil {
		sendError(w, http.StatusNotFound, errBlobUploadUnknown, err.Error())
		return
	}

	location := fmt.Sprintf("/v2/%s/%s/blobs/uploads/%s", owner, repository, reference)
	w.Header().Set("Location", location)
	w.Header().Set("Range", fmt.Sprintf("0-%d", info.Offset-1))
	w.WriteHeader(http.StatusNoContent)
}
