package controller

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (c *Controller) ListReferrers(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repository := chi.URLParam(r, "repository")
	name := owner + "/" + repository
	digest := chi.URLParam(r, "digest")
	artifactType := r.URL.Query().Get("artifactType")

	content, err := c.store.GetReferrers(name, digest, artifactType)
	if err != nil {
		sendError(w, http.StatusNotFound, errManifestUnknown, err.Error())
		return
	}

	if artifactType != "" {
		w.Header().Set("OCI-Filters-Applied", "artifactType")
	}

	w.Header().Set("Content-Type", "application/vnd.oci.image.index.v1+json")
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}
