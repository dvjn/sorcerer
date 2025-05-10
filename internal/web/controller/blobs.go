package controller

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (c *Controller) CheckBlobExists(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repository := chi.URLParam(r, "repository")
	name := owner + "/" + repository
	digest := chi.URLParam(r, "digest")

	exists, size, err := c.store.HasBlob(name, digest)
	if err != nil {
		sendError(w, http.StatusInternalServerError, errBlobUnknown, err.Error())
		return
	}

	if !exists {
		sendError(w, http.StatusNotFound, errBlobUnknown, "Blob not found")
		return
	}

	w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
	w.Header().Set("Docker-Content-Digest", digest)
	w.WriteHeader(http.StatusOK)
}

func (c *Controller) GetBlob(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repository := chi.URLParam(r, "repository")
	name := owner + "/" + repository
	digest := chi.URLParam(r, "digest")

	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" {
		fmt.Println("Range request not fully implemented, returning full blob:", rangeHeader)
	}

	blob, size, err := c.store.GetBlob(name, digest)
	if err != nil {
		sendError(w, http.StatusNotFound, errBlobUnknown, err.Error())
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

func (c *Controller) DeleteBlob(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repository := chi.URLParam(r, "repository")
	name := owner + "/" + repository
	digest := chi.URLParam(r, "digest")

	err := c.store.DeleteBlob(name, digest)
	if err != nil {
		sendError(w, http.StatusNotFound, errBlobUnknown, err.Error())
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
