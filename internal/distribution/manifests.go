package distribution

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

func (d *Distribution) checkManifestExists(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repository := chi.URLParam(r, "repository")
	name := owner + "/" + repository
	reference := chi.URLParam(r, "reference")

	exists, _, digest, err := d.store.HasManifest(name, reference)
	if err != nil {
		sendError(w, http.StatusInternalServerError, errManifestUnknown, err.Error())
		return
	}

	if !exists {
		sendError(w, http.StatusNotFound, errManifestUnknown, "Manifest not found")
		return
	}

	content, _, err := d.store.GetManifest(name, digest)
	if err != nil {
		sendError(w, http.StatusInternalServerError, errManifestUnknown, err.Error())
		return
	}

	size := len(content)

	contentType := "application/vnd.oci.image.manifest.v1+json"
	if strings.HasSuffix(reference, ".json") || strings.HasSuffix(reference, ".pretty") {
		contentType = "application/json"
	}
	var manifest map[string]any
	if err := json.Unmarshal(content, &manifest); err == nil {
		if mt, ok := manifest["mediaType"].(string); ok && mt != "" {
			contentType = mt
		}
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
	w.Header().Set("Docker-Content-Digest", digest)
	w.WriteHeader(http.StatusOK)
}

func (d *Distribution) getManifest(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repository := chi.URLParam(r, "repository")
	name := owner + "/" + repository
	reference := chi.URLParam(r, "reference")

	content, digest, err := d.store.GetManifest(name, reference)
	if err != nil {
		sendError(w, http.StatusNotFound, errManifestUnknown, err.Error())
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

func (d *Distribution) putManifest(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repository := chi.URLParam(r, "repository")
	name := owner + "/" + repository
	reference := chi.URLParam(r, "reference")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		sendError(w, http.StatusBadRequest, errManifestInvalid, "Failed to read manifest")
		return
	}

	digest, err := d.store.PutManifest(name, reference, body)
	if err != nil {
		sendError(w, http.StatusBadRequest, errManifestInvalid, err.Error())
		return
	}

	var manifest map[string]any
	if err := json.Unmarshal(body, &manifest); err == nil {
		if subject, ok := manifest["subject"].(map[string]any); ok {
			if subjectDigest, ok := subject["digest"].(string); ok {
				if err := d.store.UpdateReferrers(name, subjectDigest, body); err != nil {
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

func (d *Distribution) deleteManifest(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repository := chi.URLParam(r, "repository")
	name := owner + "/" + repository
	reference := chi.URLParam(r, "reference")

	if strings.HasPrefix(reference, "sha256:") {
		content, _, err := d.store.GetManifest(name, reference)
		if err == nil {
			var manifest map[string]any
			if err := json.Unmarshal(content, &manifest); err == nil {
				if subject, ok := manifest["subject"].(map[string]any); ok {
					if subjectDigest, ok := subject["digest"].(string); ok {
						if err := d.store.RemoveReferrer(name, subjectDigest, reference); err != nil {
							fmt.Printf("Error removing from referrers: %v\n", err)
						}
					}
				}
			}
		}
	}

	err := d.store.DeleteManifest(name, reference)
	if err != nil {
		sendError(w, http.StatusNotFound, errManifestUnknown, err.Error())
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
