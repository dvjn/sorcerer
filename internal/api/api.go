package api

import (
	_ "embed"
	"net/http"

	"github.com/dvjn/sorcerer/internal/distribution"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type API struct {
	distribution *distribution.Distribution
}

func New(distribution *distribution.Distribution) *API {
	return &API{distribution: distribution}
}

func (a *API) Router() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", index)
	r.Get("/healthz", heartbeat)

	r.Mount("/v2", a.distribution.Router())

	return r
}

//go:embed banner.txt
var sorcererBanner []byte

func index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(sorcererBanner))
	w.WriteHeader(http.StatusOK)
}

func heartbeat(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("."))
	w.WriteHeader(http.StatusOK)
}
