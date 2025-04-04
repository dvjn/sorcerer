package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func SetupRouter(handlers *Handlers) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/healthz"))

	r.Route("/v2", func(r chi.Router) {
		r.Get("/", handlers.ApiVersionCheck)

		r.Route("/{name}", func(r chi.Router) {
			r.Route("/blobs", func(r chi.Router) {
				r.Head("/{digest}", handlers.CheckBlobExists)
				r.Get("/{digest}", handlers.GetBlob)
				r.Delete("/{digest}", handlers.DeleteBlob)

				r.Route("/uploads", func(r chi.Router) {
					r.Post("/", handlers.InitiateUpload)
					r.Patch("/{reference}", handlers.UploadBlobChunk)
					r.Put("/{reference}", handlers.CompleteUpload)
					r.Get("/{reference}", handlers.GetBlobUploadStatus)
				})
			})

			r.Route("/manifests", func(r chi.Router) {
				r.Route("/{reference}", func(r chi.Router) {
					r.Head("/", handlers.CheckManifestExists)
					r.Get("/", handlers.GetManifest)
					r.Put("/", handlers.PutManifest)
					r.Delete("/", handlers.DeleteManifest)
				})
			})

			r.Route("/tags", func(r chi.Router) {
				r.Get("/list", handlers.ListTags)
			})

			r.Route("/referrers", func(r chi.Router) {
				r.Get("/{digest}", handlers.ListReferrers)
			})
		})
	})

	return r
}
