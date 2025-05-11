package distribution

import (
	"fmt"
	"io"
	"net/http"

	"github.com/dvjn/sorcerer/internal/logger"
	"github.com/go-chi/chi/v5"
)

func (d *Distribution) checkBlobExists(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repository := chi.URLParam(r, "repository")
	name := owner + "/" + repository
	digest := chi.URLParam(r, "digest")

	exists, size, err := d.store.HasBlob(name, digest)
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

func (d *Distribution) getBlob(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repository := chi.URLParam(r, "repository")
	name := owner + "/" + repository
	digest := chi.URLParam(r, "digest")

	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" {
		logger.Get(r.Context()).Warn().Str("range", rangeHeader).Msg("range header for blob not fully implemented")
	}

	blob, size, err := d.store.GetBlob(name, digest)
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
		logger.Get(r.Context()).Error().Err(err).Msg("error streaming blob")
	}
}

func (d *Distribution) deleteBlob(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repository := chi.URLParam(r, "repository")
	name := owner + "/" + repository
	digest := chi.URLParam(r, "digest")

	err := d.store.DeleteBlob(name, digest)
	if err != nil {
		sendError(w, http.StatusNotFound, errBlobUnknown, err.Error())
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
