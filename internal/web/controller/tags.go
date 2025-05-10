package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	spec_v1 "github.com/opencontainers/distribution-spec/specs-go/v1"
)

func (c *Controller) ListTags(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repository := chi.URLParam(r, "repository")
	name := owner + "/" + repository

	tags, err := c.store.ListTags(name)
	if err != nil {
		sendError(w, http.StatusNotFound, errNameUnknown, err.Error())
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

	response := spec_v1.TagList{
		Name: name,
		Tags: tags,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
