package controller

import (
	_ "embed"
	"net/http"

	"github.com/dvjn/sorcerer/internal/store"
)

type Controller struct {
	store store.Store
}

func New(store store.Store) *Controller {
	return &Controller{
		store: store,
	}
}

//go:embed banner.txt
var sorcererBanner []byte

func (c *Controller) Index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(sorcererBanner))
	w.WriteHeader(http.StatusOK)
}

func (c *Controller) Heartbeat(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("."))
	w.WriteHeader(http.StatusOK)
}

func (c *Controller) ApiVersionCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
