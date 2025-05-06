package router

import (
	"github.com/dvjn/sorcerer/internal/auth"
	"github.com/dvjn/sorcerer/internal/web"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func New(auth auth.Auth, controller web.Controller) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/healthz"))

	r.Route("/v2", func(r chi.Router) {
		r.Get("/", controller.ApiVersionCheck)

		r.Route("/{owner}/{repository}", func(r chi.Router) {
			r.Use(auth.Middleware)

			r.Route("/blobs", func(r chi.Router) {
				r.Head("/{digest}", controller.CheckBlobExists)
				r.Get("/{digest}", controller.GetBlob)
				r.Delete("/{digest}", controller.DeleteBlob)

				r.Route("/uploads", func(r chi.Router) {
					r.Post("/", controller.InitiateUpload)
					r.Patch("/{reference}", controller.UploadBlobChunk)
					r.Put("/{reference}", controller.CompleteUpload)
					r.Get("/{reference}", controller.GetBlobUploadStatus)
				})
			})

			r.Route("/manifests", func(r chi.Router) {
				r.Route("/{reference}", func(r chi.Router) {
					r.Head("/", controller.CheckManifestExists)
					r.Get("/", controller.GetManifest)
					r.Put("/", controller.PutManifest)
					r.Delete("/", controller.DeleteManifest)
				})
			})

			r.Route("/tags", func(r chi.Router) {
				r.Get("/list", controller.ListTags)
			})

			r.Route("/referrers", func(r chi.Router) {
				r.Get("/{digest}", controller.ListReferrers)
			})
		})
	})

	return r
}
