package api

import (
	_ "embed"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Api struct {
	distribution http.Handler
	auth         http.Handler
}

func New(distribution, auth http.Handler) *Api {
	return &Api{distribution: distribution, auth: auth}
}

func (a *Api) Router() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", a.index)
	r.Get("/healthz", a.heartbeat)

	r.Mount("/v2", a.distribution)
	r.Mount("/auth", a.auth)

	return r
}

//go:embed banner.txt
var sorcererBanner []byte

func (a *Api) index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(sorcererBanner))
	w.WriteHeader(http.StatusOK)
}

func (a *Api) heartbeat(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("."))
	w.WriteHeader(http.StatusOK)
}
