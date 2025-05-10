package distribution

import (
	"net/http"

	"github.com/dvjn/sorcerer/internal/store"
	"github.com/go-chi/chi/v5"
)

type Distribution struct {
	store          store.Store
	authMiddleware func(http.Handler) http.Handler
}

func New(store store.Store, authMiddleware func(http.Handler) http.Handler) *Distribution {
	return &Distribution{store: store, authMiddleware: authMiddleware}
}

func (d *Distribution) Router() *chi.Mux {
	r := chi.NewRouter()
	r.Use(d.authMiddleware)

	r.Get("/", d.apiVersionCheck)

	r.Route("/{owner}/{repository}", func(r chi.Router) {
		r.Route("/blobs", func(r chi.Router) {
			r.Head("/{digest}", d.checkBlobExists)
			r.Get("/{digest}", d.getBlob)
			r.Delete("/{digest}", d.deleteBlob)

			r.Route("/uploads", func(r chi.Router) {
				r.Post("/", d.initiateUpload)
				r.Patch("/{reference}", d.uploadBlobChunk)
				r.Put("/{reference}", d.completeUpload)
				r.Get("/{reference}", d.getBlobUploadStatus)
			})
		})

		r.Route("/manifests", func(r chi.Router) {
			r.Route("/{reference}", func(r chi.Router) {
				r.Head("/", d.checkManifestExists)
				r.Get("/", d.getManifest)
				r.Put("/", d.putManifest)
				r.Delete("/", d.deleteManifest)
			})
		})

		r.Route("/tags", func(r chi.Router) {
			r.Get("/list", d.listTags)
		})

		r.Route("/referrers", func(r chi.Router) {
			r.Get("/{digest}", d.listReferrers)
		})
	})

	return r
}

func (d *Distribution) apiVersionCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
